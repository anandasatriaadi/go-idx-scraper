package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/helper"
	"github.com/anandasatriaadi/go-idx-scraper/internal/infra/db/mongo"
	"go.uber.org/zap"
)

var defaultCollections = []string{
	"xbrl_statements",
	"stock_prices",
	"financial_reports",
	"news",
	"daily_briefings",
	"announcements",
	"last_runs",
}

func main() {
	if err := run(); err != nil {
		log.Printf("reset_db failed: %v", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		configPath      string
		forceFlag       bool
		shortForceFlag  bool
		collectionsFlag string
	)

	flag.StringVar(&configPath, "config", "config/config.yml", "Path to config file")
	flag.BoolVar(&forceFlag, "force", false, "Force wipe without interactive prompt")
	flag.BoolVar(&shortForceFlag, "f", false, "Alias for -force")
	flag.StringVar(&collectionsFlag, "collections", "", "Comma-separated collections to wipe (default: all scraper collections)")
	flag.Parse()

	force := forceFlag || shortForceFlag
	resolvedConfig := resolveConfigPath(configPath)

	logger, err := helper.NewLogger("reset_db")
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Sync()

	cfg, err := config.Load(resolvedConfig)
	if err != nil {
		logger.Error("Failed to load config", zap.String("config_path", resolvedConfig), zap.Error(err))
		return fmt.Errorf("loading config from %s: %w", resolvedConfig, err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbClient, err := mongo.NewClient(logger)
	if err != nil {
		logger.Error("Failed to connect to MongoDB", zap.Error(err))
		return fmt.Errorf("connecting to mongodb: %w", err)
	}
	db := dbClient.Database(cfg.Database.DbName)

	var targetCollections []string
	if collectionsFlag != "" {
		for _, c := range strings.Split(collectionsFlag, ",") {
			c = strings.TrimSpace(c)
			if c != "" {
				targetCollections = append(targetCollections, c)
			}
		}
	} else {
		targetCollections = append([]string{}, defaultCollections...)
	}

	fmt.Println("==================================================")
	fmt.Println("  IDX Scraper - MongoDB Database Reset Tool       ")
	fmt.Println("==================================================")
	fmt.Printf("Database:    %s\n", cfg.Database.DbName)
	fmt.Printf("Config file: %s\n", resolvedConfig)
	fmt.Printf("Target collections to drop (%d):\n", len(targetCollections))
	for _, col := range targetCollections {
		fmt.Printf("  - %s\n", col)
	}
	fmt.Println("==================================================")

	if !force {
		fmt.Print("\nWARNING: This will permanently DROP all listed collections!\nAre you sure you want to proceed? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading confirmation input: %w", err)
		}
		input = strings.TrimSpace(strings.ToLower(input))
		if input != "y" && input != "yes" {
			fmt.Println("\nOperation aborted by user.")
			logger.Info("Database reset aborted by user")
			return nil
		}
	}

	logger.Info("Starting database wipe",
		zap.String("database", cfg.Database.DbName),
		zap.Strings("collections", targetCollections),
	)

	droppedCount := 0
	for _, colName := range targetCollections {
		select {
		case <-ctx.Done():
			logger.Warn("Reset interrupted", zap.Error(ctx.Err()))
			return ctx.Err()
		default:
		}

		logger.Info("Dropping collection", zap.String("collection", colName))
		if err := db.Collection(colName).Drop(ctx); err != nil {
			logger.Warn("Failed or collection does not exist", zap.String("collection", colName), zap.Error(err))
		} else {
			logger.Info("Successfully dropped collection", zap.String("collection", colName))
			droppedCount++
		}
	}

	fmt.Printf("\nDropped %d collections successfully.\n", droppedCount)
	fmt.Println("Re-initializing repositories and rebuilding compound indexes...")

	logger.Info("Re-creating compound indexes across repositories")

	// Re-ensure indexes across repositories
	_ = mongo.NewXBRLRepository(db)
	_ = mongo.NewPriceRepository(db)
	_ = mongo.NewFinancialReportRepository(db)
	_ = mongo.NewAnnouncementRepository(db)
	_ = mongo.NewNewsRepository(db)
	_ = mongo.NewBriefingRepository(db)
	_ = mongo.NewSystemRepository(db)

	fmt.Println(" Compound indexes re-ensured:")
	fmt.Println("  - xbrl_statements:  { ticker: 1, year: -1, period: -1 } (unique)")
	fmt.Println("                      { valuation.margin_of_safety_pct: -1 }")
	fmt.Println("                      { computed_ratios.roic: -1 }")
	fmt.Println("  - stock_prices:     { ticker: 1, date: -1 } (unique)")
	fmt.Println("  - financial_reports:{ issuer_code: 1, year: 1, period_string: 1 } (unique)")
	fmt.Println("                      { is_latest: 1 }")
	fmt.Println("==================================================")
	fmt.Println(" Database reset completed successfully.")
	fmt.Println("==================================================")

	logger.Info("Database reset and index re-creation finished successfully",
		zap.String("database", cfg.Database.DbName),
		zap.Int("dropped_count", droppedCount),
	)

	return nil
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
