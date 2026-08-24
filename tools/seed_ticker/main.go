package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	br "github.com/anandasatriaadi/go-idx-scraper/internal/browser"
	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/finreport"
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
	"github.com/anandasatriaadi/go-idx-scraper/internal/helper"
	"github.com/anandasatriaadi/go-idx-scraper/internal/infra/db/mongo"
	infra "github.com/anandasatriaadi/go-idx-scraper/internal/infra/xbrl"
	"github.com/anandasatriaadi/go-idx-scraper/internal/infra/yahoo"
	"github.com/tebeka/selenium"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

var errorTitles = []string{"404 -", "404 not found", "page not found", "tidak ditemukan", "503", "attention required", "just a moment"}

func main() {
	if err := run(); err != nil {
		log.Printf("seed_ticker failed: %v", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		tickerFlag       string
		yearsFlag        string
		periodsFlag      string
		configPath       string
		skipDownloadFlag bool
		noHeadlessFlag   bool
		cleanDBFlag      bool
		fileTypeFlag     string
	)

	flag.StringVar(&tickerFlag, "ticker", "", "Target stock ticker symbol (e.g. BBRI, BBCA, TLKM) [REQUIRED]")
	flag.StringVar(&yearsFlag, "years", "5", "Historical years: count (e.g. 5), range (e.g. 2021-2025), or list (e.g. 2021,2022,2023,2024,2025)")
	flag.StringVar(&periodsFlag, "periods", "TW1,TW2,TW3,Audit", "Comma-separated filing periods (e.g. TW1,TW2,TW3,Audit or I,II,III,Tahunan)")
	flag.StringVar(&configPath, "config", "config/config.yml", "Path to config YAML file")
	flag.BoolVar(&skipDownloadFlag, "skip-download", false, "Skip Selenium XBRL downloading and use existing files")
	flag.BoolVar(&noHeadlessFlag, "no-headless", false, "Disable headless browser mode")
	flag.BoolVar(&cleanDBFlag, "clean-db", false, "Clean prior statements and prices for this ticker before seeding")
	flag.StringVar(&fileTypeFlag, "file-type", "instance.zip", "Filing file type to download (e.g. instance.zip)")
	flag.Parse()

	cleanTicker := yahoo.CleanTicker(tickerFlag)
	if cleanTicker == "" {
		fmt.Println("Error: -ticker is required (e.g. -ticker=BBRI)")
		flag.Usage()
		return fmt.Errorf("ticker flag is required")
	}

	resolvedConfig := resolveConfigPath(configPath)

	logger, err := helper.NewLogger("seed_ticker")
	if err != nil {
		return fmt.Errorf("initializing logger: %w", err)
	}
	defer logger.Sync()

	cfg, err := config.Load(resolvedConfig)
	if err != nil {
		logger.Error("Failed to load config", zap.String("config_path", resolvedConfig), zap.Error(err))
		return fmt.Errorf("loading config from %s: %w", resolvedConfig, err)
	}

	if noHeadlessFlag {
		cfg.SetHeadless(false)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbClient, err := mongo.NewClient(logger)
	if err != nil {
		logger.Error("Failed to connect to MongoDB", zap.Error(err))
		return fmt.Errorf("connecting to mongodb: %w", err)
	}
	db := dbClient.Database(cfg.Database.DbName)

	xbrlRepo := mongo.NewXBRLRepository(db)
	priceRepo := mongo.NewPriceRepository(db)
	fRepo := mongo.NewFinancialReportRepository(db)
	fSvc := finreport.NewService(fRepo, logger)
	yahooClient := yahoo.NewClient(yahoo.WithLogger(logger))

	years, err := parseYears(yearsFlag, cfg.Download.Year)
	if err != nil {
		logger.Error("Failed to parse years", zap.Error(err))
		return fmt.Errorf("parsing years: %w", err)
	}
	sort.Ints(years)

	periods := parsePeriods(periodsFlag)

	fmt.Println("================================================================================")
	fmt.Printf("  IDX 5-YEAR HISTORICAL SEEDER & FORENSIC VALUATION PIPELINE: %s\n", cleanTicker)
	fmt.Println("================================================================================")
	fmt.Printf(" Ticker:         %s (%s.JK)\n", cleanTicker, cleanTicker)
	fmt.Printf(" Target Years:   %v\n", years)
	fmt.Printf(" Target Periods: %v\n", periods)
	fmt.Printf(" Database:       %s\n", cfg.Database.DbName)
	fmt.Printf(" Config:         %s\n", resolvedConfig)
	fmt.Printf(" Skip Download:  %t\n", skipDownloadFlag)
	fmt.Printf(" Clean DB:       %t\n", cleanDBFlag)
	fmt.Println("================================================================================")

	// Step 1: Optional Clean DB
	if cleanDBFlag {
		fmt.Printf("\n[Step 1/5] Cleaning prior MongoDB data for ticker %s...\n", cleanTicker)
		logger.Info("Cleaning prior records for ticker", zap.String("ticker", cleanTicker))
		resStmt, err1 := db.Collection("xbrl_statements").DeleteMany(ctx, bson.M{"ticker": cleanTicker})
		if err1 != nil {
			logger.Warn("Failed to delete prior xbrl statements", zap.Error(err1))
		} else {
			logger.Info("Deleted prior xbrl statements", zap.Int64("count", resStmt.DeletedCount))
		}
		resPrices, err2 := db.Collection("stock_prices").DeleteMany(ctx, bson.M{"ticker": cleanTicker})
		if err2 != nil {
			logger.Warn("Failed to delete prior stock prices", zap.Error(err2))
		} else {
			logger.Info("Deleted prior stock prices", zap.Int64("count", resPrices.DeletedCount))
		}
		fmt.Println("  Cleaned prior XBRL statements and stock prices from database.")
	}

	// Step 2: Download XBRL Filings via Selenium
	if !skipDownloadFlag {
		fmt.Printf("\n[Step 2/5] Downloading %d years of historical XBRL filings via Selenium...\n", len(years))
		if err := os.MkdirAll(cfg.Paths.DownloadDir, 0755); err != nil {
			return fmt.Errorf("creating download directory: %w", err)
		}
		if err := os.MkdirAll(cfg.Paths.CheckDir, 0755); err != nil {
			return fmt.Errorf("creating check directory: %w", err)
		}

		browser, err := br.SetupSelenium(cfg)
		if err != nil {
			logger.Error("Failed to initialize Selenium", zap.Error(err))
			return fmt.Errorf("initializing selenium: %w", err)
		}
		defer browser.Close()

		fileType := strings.TrimSpace(fileTypeFlag)
		if fileType == "" {
			fileType = "instance.zip"
		}

		for _, year := range years {
			for _, period := range periods {
				select {
				case <-ctx.Done():
					logger.Warn("Seeding interrupted during download")
					return ctx.Err()
				default:
				}

				ps, modePeriod := finreport.NormalizePeriod(period)

				// Skip periods not yet released on IDX for current calendar year
				if !finreport.IsPeriodReleasedOnIDX(year, period, time.Now()) {
					logger.Debug("Period not yet released for current calendar year, skipping",
						zap.String("ticker", cleanTicker),
						zap.Int("year", year),
						zap.String("period", ps),
					)
					continue
				}

				xlsxName := fmt.Sprintf("FinancialStatement-%d-%s-%s.xlsx", year, ps, cleanTicker)
				instanceZipName := fmt.Sprintf("FinancialStatement-%d-%s-%s-instance.zip", year, modePeriod, cleanTicker)
				inlineZipName := fmt.Sprintf("FinancialStatement-%d-%s-%s-inlineXBRL.zip", year, modePeriod, cleanTicker)

				// Check if any filing is already present for this year & period
				if !cleanDBFlag {
					if exists, existingFile := findExistingFiling([]string{cfg.Paths.DownloadDir, cfg.Paths.CheckDir, "saham"}, year, period, cleanTicker); exists {
						logger.Info("Filing already present, skipping download",
							zap.String("ticker", cleanTicker),
							zap.Int("year", year),
							zap.String("period", ps),
							zap.String("file", filepath.Base(existingFile)),
						)
						continue
					}
				}

				candidates := []struct {
					url                  string
					expectedDownloadName string
					targetFilename       string
					format               string
				}{
					{
						url:                  fSvc.ConstructXBRLReportURL(year, ps, cleanTicker, "instance.zip"),
						expectedDownloadName: "instance.zip",
						targetFilename:       instanceZipName,
						format:               "instance.zip",
					},
					{
						url:                  fSvc.ConstructXBRLReportURL(year, ps, cleanTicker, "inlineXBRL.zip"),
						expectedDownloadName: "inlineXBRL.zip",
						targetFilename:       inlineZipName,
						format:               "inlineXBRL.zip",
					},
					{
						url:                  fSvc.ConstructReportURL(year, ps, cleanTicker),
						expectedDownloadName: xlsxName,
						targetFilename:       xlsxName,
						format:               ".xlsx",
					},
				}

				downloadSuccess := false
				for _, cand := range candidates {
					select {
					case <-ctx.Done():
						return ctx.Err()
					default:
					}

					logger.Info("Attempting filing download",
						zap.String("ticker", cleanTicker),
						zap.Int("year", year),
						zap.String("period", ps),
						zap.String("format", cand.format),
						zap.String("url", cand.url),
					)

					targetPath := filepath.Join(cfg.Paths.DownloadDir, cand.targetFilename)
					_ = os.Remove(filepath.Join(cfg.Paths.DownloadDir, cand.expectedDownloadName))

					if err := downloadFile(cand.url, browser.Driver, logger); err != nil {
						continue
					}

					if err := waitForDownloadedFile(cfg.Paths.DownloadDir, cand.expectedDownloadName, targetPath, 5*time.Second, logger); err == nil {
						logger.Info("Successfully downloaded filing", zap.String("file", cand.targetFilename), zap.String("format", cand.format))
						downloadSuccess = true
						break
					}
				}

				if !downloadSuccess {
					logger.Warn("Filing not available in any format on IDX (instance.zip, inlineXBRL.zip, .xlsx)",
						zap.String("ticker", cleanTicker),
						zap.Int("year", year),
						zap.String("period", ps),
					)
				}
			}
		}
	} else {
		fmt.Println("\n[Step 2/5] Skipped download step (-skip-download enabled).")
	}

	// Step 3: Stream Parse Downloaded XBRL Files into MongoDB
	fmt.Printf("\n[Step 3/5] Streaming & parsing historical XBRL files into MongoDB...\n")
	parsedFiles := parseAndIngestTickerFiles(ctx, cleanTicker, cfg, xbrlRepo, logger)
	fmt.Printf("  Parsed & ingested %d XBRL statements for %s.\n", parsedFiles, cleanTicker)

	// Step 4: Ingest Yahoo Finance 5-Year Historical Daily Candles
	fmt.Printf("\n[Step 4/5] Ingesting 5-year daily price history from Yahoo Finance (%s.JK)...\n", cleanTicker)
	candles, err := yahooClient.FetchHistoricalPricesWithContext(ctx, cleanTicker, "5y")
	var latestPrice float64
	if err != nil {
		logger.Warn("Failed to fetch historical prices from Yahoo Finance", zap.String("ticker", cleanTicker), zap.Error(err))
		fmt.Printf("  Warning: Failed to fetch Yahoo price history: %v\n", err)
	} else if len(candles) > 0 {
		if err := priceRepo.UpsertCandles(ctx, cleanTicker, candles); err != nil {
			logger.Error("Failed to persist price candles to MongoDB", zap.String("ticker", cleanTicker), zap.Error(err))
			fmt.Printf("  Warning: Failed to persist price candles: %v\n", err)
		} else {
			for i := len(candles) - 1; i >= 0; i-- {
				if candles[i].Close > 0 {
					latestPrice = candles[i].Close
					break
				}
			}
			latestCandle := candles[len(candles)-1]
			logger.Info("Ingested historical price candles",
				zap.String("ticker", cleanTicker),
				zap.Int("candles_count", len(candles)),
				zap.Float64("latest_price", latestPrice),
				zap.Time("latest_date", latestCandle.Date),
			)
			fmt.Printf("  Successfully stored %d daily price candles. Latest price: IDR %.2f (%s)\n",
				len(candles), latestPrice, latestCandle.Date.Format("2006-01-02"))
		}
	}

	if latestPrice == 0 {
		if p, pErr := yahooClient.FetchLatestPrice(ctx, cleanTicker); pErr == nil && p > 0 {
			latestPrice = p
			fmt.Printf("  Retrieved latest market price: IDR %.2f\n", latestPrice)
		}
	}

	// Fetch current USD/IDR exchange rate
	usdidr, _ := yahooClient.FetchUSDIDR(ctx)

	// Step 5: Compute Multi-Year Forensic Valuation & Re-index
	fmt.Printf("\n[Step 5/5] Computing multi-year forensic valuations (Piotroski, Altman Z, ROIC, Graham FV)...\n")
	statements, err := xbrlRepo.FindHistoricalByTicker(ctx, cleanTicker, 100)
	if err != nil {
		logger.Error("Failed to retrieve historical statements for valuation", zap.Error(err))
		return fmt.Errorf("retrieving historical statements: %w", err)
	}

	if len(statements) == 0 {
		fmt.Printf("  No XBRL statements found in database for ticker %s. Cannot compute valuation.\n", cleanTicker)
		return nil
	}

	// Sort chronologically (oldest to newest) to pass prior statement YoY for Piotroski F-score
	sort.Slice(statements, func(i, j int) bool {
		if statements[i].Year != statements[j].Year {
			return statements[i].Year < statements[j].Year
		}
		return periodRank(statements[i].Period) < periodRank(statements[j].Period)
	})

	for i, stmt := range statements {
		var priorStmt *xbrl.Statement
		if i > 0 {
			priorStmt = statements[i-1]
		}

		if stmt.Metadata.Currency == "USD" && stmt.Metadata.ConversionRate == 0 && usdidr > 0 {
			stmt.Metadata.ConversionRate = usdidr
		}

		if err := xbrl.ComputeValuationAndRatios(stmt, priorStmt, latestPrice); err != nil {
			logger.Warn("Valuation calculation error", zap.String("ticker", cleanTicker), zap.Int("year", stmt.Year), zap.Error(err))
		}

		if err := xbrlRepo.Upsert(ctx, stmt); err != nil {
			logger.Error("Failed to persist updated valuation", zap.String("ticker", cleanTicker), zap.Int("year", stmt.Year), zap.Error(err))
		}
	}
	fmt.Printf("  Recomputed valuation multiples and forensic metrics across %d filing periods.\n", len(statements))

	// Step 6: Print Formatted 5-Year Forensic Valuation Report Table
	printValuationReport(cleanTicker, statements, latestPrice, usdidr)

	logger.Info("5-year seeding and valuation completed successfully",
		zap.String("ticker", cleanTicker),
		zap.Int("statements_count", len(statements)),
		zap.Float64("latest_price", latestPrice),
	)

	return nil
}

func periodRank(period string) int {
	p := strings.ToUpper(strings.TrimSpace(period))
	switch p {
	case "I", "TW1", "Q1":
		return 1
	case "II", "TW2", "Q2":
		return 2
	case "III", "TW3", "Q3":
		return 3
	case "IV", "TW4", "Q4":
		return 4
	case "TAHUNAN", "AUDIT", "FY":
		return 5
	default:
		return 0
	}
}

func parseYears(yearsFlag string, defaultYear string) ([]int, error) {
	target := strings.TrimSpace(yearsFlag)
	if target == "" {
		target = strings.TrimSpace(defaultYear)
	}
	if target == "" {
		target = "5"
	}

	// Single count (e.g. "5" or "10") -> generate last N years up to current year
	if count, err := strconv.Atoi(target); err == nil && count <= 20 {
		currentYear := time.Now().Year()
		var years []int
		for y := currentYear - count + 1; y <= currentYear; y++ {
			years = append(years, y)
		}
		return years, nil
	}

	var years []int
	seen := make(map[int]bool)
	parts := strings.Split(target, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.Contains(p, "-") {
			rangeParts := strings.Split(p, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid year range: %s", p)
			}
			start, err1 := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err1 != nil || err2 != nil {
				return nil, fmt.Errorf("invalid year range numbers: %s", p)
			}
			step := 1
			if start > end {
				step = -1
			}
			for y := start; ; y += step {
				if !seen[y] {
					seen[y] = true
					years = append(years, y)
				}
				if y == end {
					break
				}
			}
		} else {
			y, err := strconv.Atoi(p)
			if err != nil {
				return nil, fmt.Errorf("invalid year: %s", p)
			}
			if !seen[y] {
				seen[y] = true
				years = append(years, y)
			}
		}
	}
	return years, nil
}

func parsePeriods(periodsFlag string) []string {
	if strings.TrimSpace(periodsFlag) == "" {
		return []string{"TW1", "TW2", "TW3", "Audit"}
	}
	var periods []string
	for _, p := range strings.Split(periodsFlag, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			periods = append(periods, p)
		}
	}
	return periods
}

func parseAndIngestTickerFiles(ctx context.Context, ticker string, cfg *config.Config, repo xbrl.Repository, logger *zap.Logger) int {
	cleanTicker := strings.ToUpper(strings.TrimSpace(ticker))
	searchDirs := []string{cfg.Paths.DownloadDir, cfg.Paths.CheckDir, "saham"}
	seenPaths := make(map[string]bool)
	parsedCount := 0

	for _, dir := range searchDirs {
		if dir == "" {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			fullPath := filepath.Join(dir, entry.Name())
			if seenPaths[fullPath] {
				continue
			}

			nameLower := strings.ToLower(entry.Name())
			nameUpper := strings.ToUpper(entry.Name())

			if !strings.HasSuffix(nameLower, ".zip") && !strings.HasSuffix(nameLower, ".xml") && !strings.HasSuffix(nameLower, ".xbrl") && !strings.HasSuffix(nameLower, ".xlsx") {
				continue
			}

			// If filename contains specific ticker, must match cleanTicker
			if !strings.Contains(nameUpper, cleanTicker) && !strings.HasPrefix(nameLower, "instance") && !strings.HasPrefix(nameLower, "inline") {
				continue
			}

			stmt, pErr := infra.ParseAnyFiling(fullPath)
			if pErr != nil {
				continue
			}

			if strings.ToUpper(strings.TrimSpace(stmt.Ticker)) != cleanTicker {
				continue
			}

			seenPaths[fullPath] = true
			_ = xbrl.ComputeValuationAndRatios(stmt, nil, 0)

			if err := repo.Upsert(ctx, stmt); err != nil {
				logger.Error("Failed to upsert parsed statement", zap.String("ticker", cleanTicker), zap.Error(err))
			} else {
				logger.Info("Upserted XBRL filing",
					zap.String("ticker", cleanTicker),
					zap.Int("year", stmt.Year),
					zap.String("period", stmt.Period),
					zap.String("file", entry.Name()),
				)
				parsedCount++
			}
		}
	}

	return parsedCount
}

func printValuationReport(ticker string, statements []*xbrl.Statement, currentPrice float64, usdidr float64) {
	if len(statements) == 0 {
		return
	}

	// Clone and sort descending (newest at top) for report table
	descStmts := make([]*xbrl.Statement, len(statements))
	copy(descStmts, statements)
	sort.Slice(descStmts, func(i, j int) bool {
		if descStmts[i].Year != descStmts[j].Year {
			return descStmts[i].Year > descStmts[j].Year
		}
		return periodRank(descStmts[i].Period) > periodRank(descStmts[j].Period)
	})

	latest := descStmts[0]
	compName := latest.CompanyName
	if compName == "" {
		compName = ticker
	}

	fmt.Println("\n========================================================================================================================")
	fmt.Printf("                              HISTORICAL FORENSIC VALUATION REPORT: %s (%s)\n", ticker, compName)
	fmt.Println("========================================================================================================================")
	fmt.Printf(" Market Price:  IDR %.2f  |  Sector: %s  |  Industry: %s\n", currentPrice, latest.Metadata.Sector, latest.Metadata.Industry)
	if usdidr > 0 {
		fmt.Printf(" USD/IDR Rate:  IDR %.2f  |  Auditor: %s (%s)\n", usdidr, latest.Metadata.AuditorName, latest.Metadata.AuditStatus)
	}
	fmt.Println("------------------------------------------------------------------------------------------------------------------------")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Period\tRevenue\tNet Income\tROIC\tROE\tCurrent\tD/E\tAltman Z\tPiotroski\tEPS (IDR)\tBVPS (IDR)\tGraham FV\tMOS %")
	fmt.Fprintln(w, "------\t-------\t----------\t----\t---\t-------\t---\t--------\t---------\t---------\t----------\t---------\t-----")

	for _, s := range descStmts {
		periodLabel := fmt.Sprintf("%d %s", s.Year, s.Period)
		revStr := formatIDR(s.Core.Revenue)
		netStr := formatIDR(s.Core.NetIncome)
		roicStr := formatPct(s.ComputedRatios.ROIC * 100)
		roeStr := formatPct(s.ComputedRatios.ROE * 100)
		currRatioStr := fmt.Sprintf("%.2f", s.ComputedRatios.CurrentRatio)
		deStr := fmt.Sprintf("%.2f", s.ComputedRatios.DebtToEquity)
		altmanStr := fmt.Sprintf("%.2f", s.ComputedRatios.AltmanZScore)
		piotroskiStr := fmt.Sprintf("%d/9", s.ComputedRatios.PiotroskiFScore)
		epsStr := fmt.Sprintf("%.2f", s.Valuation.NormalizedEPS)
		bvpsStr := fmt.Sprintf("%.2f", s.Valuation.NormalizedBVPS)
		grahamStr := fmt.Sprintf("%.2f", s.Valuation.GrahamNumber)
		mosStr := formatSignedPct(s.Valuation.MarginOfSafetyPct)

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			periodLabel, revStr, netStr, roicStr, roeStr, currRatioStr, deStr,
			altmanStr, piotroskiStr, epsStr, bvpsStr, grahamStr, mosStr,
		)
	}
	w.Flush()

	fmt.Println("------------------------------------------------------------------------------------------------------------------------")
	fmt.Println("                                           EXECUTIVE VALUATION VERDICT                                                  ")
	fmt.Println("------------------------------------------------------------------------------------------------------------------------")

	// Piotroski Assessment
	fScore := latest.ComputedRatios.PiotroskiFScore
	var fVerdict string
	if fScore >= 8 {
		fVerdict = "Strong Financial Health (Value Investment Grade)"
	} else if fScore >= 5 {
		fVerdict = "Moderate / Stable Financial Operations"
	} else {
		fVerdict = "Weak Operational Performance / Potential Value Trap"
	}
	fmt.Printf(" * Piotroski F-Score:   %d/9 -> %s\n", fScore, fVerdict)

	// Altman Z Assessment
	zScore := latest.ComputedRatios.AltmanZScore
	var zVerdict string
	if zScore > 2.60 {
		zVerdict = "Safe Zone (Low Probability of Financial Distress)"
	} else if zScore >= 1.10 {
		zVerdict = "Grey Zone (Moderate Credit / Solvency Risk)"
	} else {
		zVerdict = "Distress Zone (Elevated Insolvency Risk)"
	}
	fmt.Printf(" * Altman Z''-Score:    %.2f -> %s\n", zScore, zVerdict)

	// Graham Fair Value & MOS
	grahamVal := latest.Valuation.GrahamNumber
	mos := latest.Valuation.MarginOfSafetyPct
	var mosVerdict string
	if currentPrice > 0 && grahamVal > 0 {
		if mos > 0 {
			mosVerdict = fmt.Sprintf("UNDERVALUED by %.2f%% relative to Benjamin Graham Fair Value", mos)
		} else {
			mosVerdict = fmt.Sprintf("OVERVALUED by %.2f%% relative to Benjamin Graham Fair Value", -mos)
		}
	} else {
		mosVerdict = "Insufficient data to compute market price discount"
	}
	fmt.Printf(" * Graham Fair Value:   IDR %.2f vs Current Market Price: IDR %.2f\n", grahamVal, currentPrice)
	fmt.Printf(" * Margin of Safety:    %s\n", mosVerdict)

	// Valuation Multiples
	pe := latest.Valuation.PERatio
	pb := latest.Valuation.PBRatio
	peStr := "N/A"
	pbStr := "N/A"
	if pe > 0 {
		peStr = fmt.Sprintf("%.2fx", pe)
	}
	if pb > 0 {
		pbStr = fmt.Sprintf("%.2fx", pb)
	}
	fmt.Printf(" * Valuation Multiples: P/E: %s | P/B: %s | ROIC: %.2f%% | ROE: %.2f%%\n",
		peStr, pbStr, latest.ComputedRatios.ROIC*100, latest.ComputedRatios.ROE*100)

	fmt.Println("========================================================================================================================")
}

func formatIDR(val float64) string {
	abs := math.Abs(val)
	if abs == 0 {
		return "-"
	}
	if abs >= 1e12 {
		return fmt.Sprintf("%.2f T", val/1e12)
	}
	if abs >= 1e9 {
		return fmt.Sprintf("%.2f B", val/1e9)
	}
	if abs >= 1e6 {
		return fmt.Sprintf("%.2f M", val/1e6)
	}
	return fmt.Sprintf("%.0f", val)
}

func formatPct(val float64) string {
	if val == 0 {
		return "-"
	}
	return fmt.Sprintf("%.2f%%", val)
}

func formatSignedPct(val float64) string {
	if val == 0 {
		return "-"
	}
	if val > 0 {
		return fmt.Sprintf("+%.2f%%", val)
	}
	return fmt.Sprintf("%.2f%%", val)
}

func resolveConfigPath(customPath string) string {
	if customPath != "config/config.yml" && customPath != "" {
		return customPath
	}
	if runtime.GOOS == "darwin" {
		if _, err := os.Stat("config/config-mac.yml"); err == nil {
			return "config/config-mac.yml"
		}
	}
	if _, err := os.Stat("config/config.yml"); err == nil {
		return "config/config.yml"
	}
	if _, err := os.Stat("config/config-mac.yml"); err == nil {
		return "config/config-mac.yml"
	}
	return customPath
}

func checkPageTitleForErrors(title, url string, logger *zap.Logger) error {
	titleLower := strings.ToLower(title)
	for _, errorStr := range errorTitles {
		if strings.Contains(titleLower, errorStr) {
			switch errorStr {
			case "404", "document":
				logger.Info("Stock document not found", zap.String("url", url), zap.String("title", titleLower))
				return fmt.Errorf("not found")
			case "503":
				logger.Info("Server error", zap.String("url", url), zap.String("title", titleLower))
				return fmt.Errorf("server error")
			case "attention required", "just a moment":
				logger.Info("Bot detector encountered", zap.String("url", url), zap.String("title", titleLower))
				return fmt.Errorf("bot detector")
			}
		}
	}
	return nil
}

func downloadFile(url string, driver selenium.WebDriver, logger *zap.Logger) error {
	if err := driver.Get(url); err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}

	time.Sleep(1 * time.Second)

	title, _ := driver.Title()
	logger.Info("Browser title retrieved", zap.String("title", title))

	if err := checkPageTitleForErrors(title, url, logger); err != nil {
		_ = driver.Get("data:,")
		return err
	}

	return nil
}

func waitForDownloadedFile(downloadDir string, expectedName string, targetPath string, timeout time.Duration, logger *zap.Logger) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		time.Sleep(300 * time.Millisecond)

		crdownloadFiles, _ := filepath.Glob(filepath.Join(downloadDir, "*.crdownload"))
		if len(crdownloadFiles) > 0 {
			continue
		}

		if info, err := os.Stat(targetPath); err == nil && info.Size() > 0 {
			return nil
		}

		expectedPath := filepath.Join(downloadDir, expectedName)
		if info, err := os.Stat(expectedPath); err == nil && info.Size() > 0 {
			if expectedPath != targetPath {
				if err := os.Rename(expectedPath, targetPath); err != nil {
					if err := copyAndRemove(expectedPath, targetPath); err != nil {
						return fmt.Errorf("moving downloaded file %s to %s: %w", expectedPath, targetPath, err)
					}
				}
			}
			return nil
		}

		entries, err := os.ReadDir(downloadDir)
		if err == nil {
			baseExpected := strings.TrimSuffix(expectedName, filepath.Ext(expectedName))
			extExpected := filepath.Ext(expectedName)
			for _, entry := range entries {
				if entry.IsDir() || strings.HasSuffix(entry.Name(), ".crdownload") {
					continue
				}
				fullEntryPath := filepath.Join(downloadDir, entry.Name())
				if fullEntryPath == targetPath {
					if info, err := os.Stat(targetPath); err == nil && info.Size() > 0 {
						return nil
					}
				}
				if strings.HasPrefix(entry.Name(), baseExpected) && strings.HasSuffix(entry.Name(), extExpected) {
					if info, err := os.Stat(fullEntryPath); err == nil && info.Size() > 0 {
						if err := os.Rename(fullEntryPath, targetPath); err != nil {
							if err := copyAndRemove(fullEntryPath, targetPath); err != nil {
								return fmt.Errorf("moving variant file %s to %s: %w", fullEntryPath, targetPath, err)
							}
						}
						return nil
					}
				}
			}
		}
	}
	return fmt.Errorf("timeout waiting for %s to download", expectedName)
}

func copyAndRemove(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	if err = out.Sync(); err != nil {
		return err
	}
	_ = in.Close()
	return os.Remove(src)
}

func fileExistsAndNotEmpty(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir() && info.Size() > 0
}

func findExistingFiling(dirs []string, year int, period, ticker string) (bool, string) {
	cleanTicker := strings.ToUpper(strings.TrimSpace(ticker))
	ps, modePeriod := finreport.NormalizePeriod(period)

	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			nameUpper := strings.ToUpper(entry.Name())
			if !strings.Contains(nameUpper, cleanTicker) {
				continue
			}
			if !strings.Contains(nameUpper, fmt.Sprintf("-%d-", year)) && !strings.Contains(nameUpper, fmt.Sprintf("%d", year)) {
				continue
			}
			matchPeriod := strings.Contains(nameUpper, fmt.Sprintf("-%s-", ps)) ||
				strings.Contains(nameUpper, fmt.Sprintf("-%s-", modePeriod)) ||
				strings.Contains(nameUpper, fmt.Sprintf("-%s.", ps)) ||
				strings.Contains(nameUpper, fmt.Sprintf("-%s.", modePeriod)) ||
				strings.Contains(nameUpper, fmt.Sprintf("-%s-", strings.ToUpper(period)))
			if matchPeriod {
				fullPath := filepath.Join(dir, entry.Name())
				if info, err := os.Stat(fullPath); err == nil && info.Size() > 0 {
					return true, fullPath
				}
			}
		}
	}
	return false, ""
}
