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
	flag.StringVar(&configPath, "config", "config/config.yml", "Path to config file")
	flag.StringVar(&targetDir, "dir", "saham", "Directory containing XBRL zip or xml files")
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
	repo := mongo.NewXBRLRepository(db)

	ctx := context.Background()

	files, err := os.ReadDir(targetDir)
	if err != nil {
		logger.Fatal("reading dir", zap.Error(err))
	}

	parsedCount := 0
	for _, f := range files {
		if strings.HasSuffix(strings.ToLower(f.Name()), ".zip") || strings.HasSuffix(strings.ToLower(f.Name()), ".xbrl") {
			fullPath := filepath.Join(targetDir, f.Name())
			logger.Info("Processing filing", zap.String("file", f.Name()))

			var stmt *domain.Statement
			if strings.HasSuffix(strings.ToLower(f.Name()), ".zip") {
				stmt, err = infra.ParseInstanceZip(fullPath)
			} else {
				fileHandle, oErr := os.Open(fullPath)
				if oErr == nil {
					stmt, err = infra.ParseInstanceXML(fileHandle)
					fileHandle.Close()
				} else {
					err = oErr
				}
			}

			if err != nil {
				logger.Warn("Failed to parse XBRL filing", zap.String("file", f.Name()), zap.Error(err))
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
