package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/stock"
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
	"github.com/anandasatriaadi/go-idx-scraper/internal/helper"
	"github.com/anandasatriaadi/go-idx-scraper/internal/infra/db/mongo"
	"github.com/anandasatriaadi/go-idx-scraper/internal/infra/yahoo"
	"go.uber.org/zap"
)

func main() {
	var (
		configPath    string
		tickerFlag    string
		tickersFlag   string
		stockListPath string
		rangePeriod   string
		delayMs       int
	)

	flag.StringVar(&configPath, "config", "config/config.yml", "Path to config file")
	flag.StringVar(&tickerFlag, "ticker", "", "Single ticker or comma-separated tickers to update (e.g. BBRI or BBRI,BBCA)")
	flag.StringVar(&tickersFlag, "tickers", "", "Alias for -ticker")
	flag.StringVar(&stockListPath, "stock-list", "stock-list.json", "Path to stock-list.json file")
	flag.StringVar(&rangePeriod, "range", "5d", "Yahoo Finance price range period (e.g. 5d, 1mo, 1y, 5y)")
	flag.IntVar(&delayMs, "delay-ms", 100, "Delay in milliseconds between ticker requests")
	flag.Parse()

	logger, err := helper.NewLogger("price_updater")
	if err != nil {
		log.Fatalf("logger init failed: %v", err)
	}
	defer logger.Sync()

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Fatal("loading config", zap.Error(err))
	}

	dbClient, err := mongo.NewClient(logger)
	if err != nil {
		logger.Fatal("mongodb connect", zap.Error(err))
	}
	db := dbClient.Database(cfg.Database.DbName)
	priceRepo := mongo.NewPriceRepository(db)
	xbrlRepo := mongo.NewXBRLRepository(db)
	yahooClient := yahoo.NewClient(yahoo.WithLogger(logger))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var tickers []string
	tickerInput := tickerFlag
	if tickerInput == "" {
		tickerInput = tickersFlag
	}

	if tickerInput != "" {
		rawList := strings.Split(tickerInput, ",")
		for _, t := range rawList {
			clean := strings.ToUpper(strings.TrimSpace(t))
			if clean != "" {
				tickers = append(tickers, clean)
			}
		}
	} else {
		// Load from stock list file
		listPath := stockListPath
		if listPath == "" && cfg.Paths.IssuerList != "" {
			listPath = cfg.Paths.IssuerList
		}
		stocks, err := helper.LoadCurrent(listPath, logger)
		if err != nil {
			logger.Fatal("failed to load stock list", zap.String("path", listPath), zap.Error(err))
		}
		tickers = stocks
	}

	if len(tickers) == 0 {
		logger.Warn("No tickers found to update")
		return
	}

	logger.Info("Starting daily price updater",
		zap.Int("total_tickers", len(tickers)),
		zap.String("range", rangePeriod),
	)

	// Fetch USD/IDR exchange rate as reference
	usdidr, err := yahooClient.FetchUSDIDR(ctx)
	if err != nil {
		logger.Warn("Failed to fetch USD/IDR rate", zap.Error(err))
	} else {
		logger.Info("Current USD/IDR exchange rate", zap.Float64("usd_idr", usdidr))
	}

	successCount := 0
	failCount := 0

	for i, ticker := range tickers {
		select {
		case <-ctx.Done():
			logger.Warn("Price updater interrupted, stopping early", zap.Int("processed", i))
			break
		default:
		}

		cleanTicker := yahoo.CleanTicker(ticker)
		logger.Info("Fetching daily prices", zap.String("ticker", cleanTicker), zap.Int("progress", i+1), zap.Int("total", len(tickers)))

		candles, err := yahooClient.FetchHistoricalPricesWithContext(ctx, cleanTicker, rangePeriod)
		if err != nil {
			logger.Warn("Failed to fetch historical prices", zap.String("ticker", cleanTicker), zap.Error(err))
			failCount++
			continue
		}

		if len(candles) == 0 {
			logger.Warn("No price candles returned", zap.String("ticker", cleanTicker))
			failCount++
			continue
		}

		if err := priceRepo.UpsertCandles(ctx, cleanTicker, candles); err != nil {
			logger.Error("Failed to persist candles to MongoDB", zap.String("ticker", cleanTicker), zap.Error(err))
			failCount++
			continue
		}

		var latestPrice float64
		for j := len(candles) - 1; j >= 0; j-- {
			if candles[j].Close > 0 {
				latestPrice = candles[j].Close
				break
			}
		}

		// Fetch historical candles from MongoDB for rolling valuation bands & momentum timing
		allDbCandles, pErr := priceRepo.GetPrices(ctx, cleanTicker, 1250)
		var fullCandles []stock.PriceCandle
		if pErr == nil && len(allDbCandles) > 0 {
			fullCandles = make([]stock.PriceCandle, len(allDbCandles))
			for k, c := range allDbCandles {
				fullCandles[k] = *c
			}
		} else {
			fullCandles = candles
		}

		// Refresh valuation in latest XBRL statement if present
		stmts, err := xbrlRepo.FindHistoricalByTicker(ctx, cleanTicker, 2)
		if err == nil && len(stmts) > 0 {
			latestStmt := stmts[0]
			var priorStmt *xbrl.Statement
			if len(stmts) > 1 {
				priorStmt = stmts[1]
			}

			if err := xbrl.ComputeValuationAndRatios(latestStmt, priorStmt, latestPrice); err != nil {
				logger.Warn("Failed to recompute valuation with latest price", zap.String("ticker", cleanTicker), zap.Error(err))
			} else {
				if len(fullCandles) > 0 {
					bands := xbrl.ComputeValuationBands(fullCandles, latestStmt.Valuation.NormalizedEPS, latestStmt.Valuation.NormalizedBVPS)
					timing := xbrl.ComputeTimingSignals(fullCandles, bands, latestStmt.Valuation.NormalizedEPS, latestStmt.Valuation.NormalizedBVPS)
					latestStmt.ValuationBands = &bands
					latestStmt.TimingSignal = &timing
					latestStmt.Valuation.ValuationBands = &bands
					latestStmt.Valuation.TimingSignal = &timing
				}

				if err := xbrlRepo.Upsert(ctx, latestStmt); err != nil {
					logger.Warn("Failed to update XBRL statement valuation", zap.String("ticker", cleanTicker), zap.Error(err))
				} else {
					timingScore := 0
					timingStatus := "N/A"
					if latestStmt.TimingSignal != nil {
						timingScore = latestStmt.TimingSignal.Score
						timingStatus = latestStmt.TimingSignal.Status
					}
					logger.Info("Refreshed XBRL valuation multiples & timing signals",
						zap.String("ticker", cleanTicker),
						zap.Float64("price", latestPrice),
						zap.Float64("graham_number", latestStmt.Valuation.GrahamNumber),
						zap.Float64("margin_of_safety_pct", latestStmt.Valuation.MarginOfSafetyPct),
						zap.Float64("pe_ratio", latestStmt.Valuation.PERatio),
						zap.Float64("pb_ratio", latestStmt.Valuation.PBRatio),
						zap.Int("timing_score", timingScore),
						zap.String("timing_status", timingStatus),
					)
				}
			}
		}

		successCount++

		if delayMs > 0 && i < len(tickers)-1 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
	}

	logger.Info("Daily price updater completed",
		zap.Int("total", len(tickers)),
		zap.Int("success", successCount),
		zap.Int("failed", failCount),
	)
	fmt.Printf("Daily price updater finished: %d succeeded, %d failed out of %d tickers.\n", successCount, failCount, len(tickers))
}
