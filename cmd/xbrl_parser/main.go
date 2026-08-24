package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	domain "github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
	"github.com/anandasatriaadi/go-idx-scraper/internal/helper"
	"github.com/anandasatriaadi/go-idx-scraper/internal/infra/db/mongo"
	infra "github.com/anandasatriaadi/go-idx-scraper/internal/infra/xbrl"
	"go.uber.org/zap"
)

func main() {
	var configPath string
	var targetDir string
	var tickerFlag string
	var cleanDB bool

	flag.StringVar(&configPath, "config", "config/config.yml", "Path to config file")
	flag.StringVar(&targetDir, "dir", "saham", "Directory containing XBRL zip or xml files")
	flag.StringVar(&tickerFlag, "ticker", "", "Filter filings by single ticker or comma-separated list (e.g. BBRI,TLKM)")
	flag.BoolVar(&cleanDB, "clean-db", false, "Drop xbrl_statements collection before ingestion")
	flag.Parse()

	logger, err := helper.NewLogger("xbrl_parser")
	if err != nil {
		log.Fatalf("logger init: %v", err)
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

	ctx := context.Background()

	if cleanDB {
		logger.Info("Dropping xbrl_statements collection as requested by -clean-db")
		if err := db.Collection("xbrl_statements").Drop(ctx); err != nil {
			logger.Warn("Failed to drop xbrl_statements collection", zap.Error(err))
		} else {
			logger.Info("xbrl_statements collection dropped successfully")
		}
	}

	repo := mongo.NewXBRLRepository(db)

	tickerFilter := make(map[string]bool)
	if tickerFlag != "" {
		for _, t := range strings.Split(tickerFlag, ",") {
			t = strings.ToUpper(strings.TrimSpace(t))
			if t != "" {
				tickerFilter[t] = true
			}
		}
		logger.Info("Filtering by ticker(s)", zap.Any("tickers", tickerFilter))
	}

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		logger.Warn("Target directory does not exist", zap.String("dir", targetDir))
		fmt.Printf("Directory '%s' does not exist. 0 XBRL statements parsed.\n", targetDir)
		return
	}

	files, err := os.ReadDir(targetDir)
	if err != nil {
		logger.Fatal("reading dir", zap.String("dir", targetDir), zap.Error(err))
	}

	parsedCount := 0
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		fileNameLower := strings.ToLower(f.Name())
		if strings.HasSuffix(fileNameLower, ".zip") || strings.HasSuffix(fileNameLower, ".xbrl") || strings.HasSuffix(fileNameLower, ".xml") || strings.HasSuffix(fileNameLower, ".xlsx") {
			// Pre-filter: if filename explicitly contains a ticker not in tickerFilter, skip
			if len(tickerFilter) > 0 {
				matched := false
				fileNameUpper := strings.ToUpper(f.Name())
				for t := range tickerFilter {
					if strings.Contains(fileNameUpper, t) {
						matched = true
						break
					}
				}
				isGeneric := strings.HasPrefix(fileNameLower, "instance") || strings.HasPrefix(fileNameLower, "inline")
				if !matched && !isGeneric {
					continue
				}
			}

			fullPath := filepath.Join(targetDir, f.Name())
			logger.Info("Processing filing", zap.String("file", f.Name()))

			stmt, err := infra.ParseAnyFiling(fullPath)
			if err != nil {
				logger.Warn("Failed to parse filing", zap.String("file", f.Name()), zap.Error(err))
				continue
			}

			// Post-parse filter by ticker if specified
			if len(tickerFilter) > 0 && !tickerFilter[strings.ToUpper(stmt.Ticker)] {
				logger.Debug("Skipping statement due to ticker filter", zap.String("ticker", stmt.Ticker), zap.String("file", f.Name()))
				continue
			}

			if err := domain.ComputeValuationAndRatios(stmt, nil, 0); err != nil {
				logger.Warn("Valuation calculation failed", zap.String("ticker", stmt.Ticker), zap.Error(err))
			}

			if err := repo.Upsert(ctx, stmt); err != nil {
				logger.Error("Failed to upsert to MongoDB", zap.String("ticker", stmt.Ticker), zap.Error(err))
			} else {
				logger.Info("Successfully parsed & saved XBRL statement", zap.String("ticker", stmt.Ticker), zap.Int("year", stmt.Year), zap.String("period", stmt.Period))
				parsedCount++
			}
		}
	}

	logger.Info("XBRL parsing completed", zap.Int("total_parsed", parsedCount))
	fmt.Printf("Parsed %d XBRL statements successfully.\n", parsedCount)
}
