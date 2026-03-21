package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	br "github.com/anandasatriaadi/go-idx-scraper/internal/browser"
	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/news"
	"github.com/anandasatriaadi/go-idx-scraper/internal/helper"
	newsRepo "github.com/anandasatriaadi/go-idx-scraper/internal/infra/db/mongo"
	"github.com/anandasatriaadi/go-idx-scraper/internal/infra/scraper/kontan"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Printf("Application failed: %v", err)
		os.Exit(1)
	}
}

func run() error {
	var configPath string
	var noHeadless bool
	var scrapeStartDate string
	var scrapeEndDate string

	flag.StringVar(&configPath, "config", "config/config.yml", "Path to configuration file")
	flag.BoolVar(&noHeadless, "no-headless", false, "Disable headless mode for browser")
	flag.StringVar(&scrapeStartDate, "start-date", "", "Start date (YYYY-MM-DD), default today")
	flag.StringVar(&scrapeEndDate, "end-date", "", "End date (YYYY-MM-DD), default today")
	flag.Parse()

	logger, err := helper.NewLogger("scraper")
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Sync()

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("Failed to load config", zap.Error(err))
		return err
	}

	if noHeadless {
		cfg.SetHeadless(false)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
		cancel()
	}()

	dbClient, err := newsRepo.NewClient(logger)
	if err != nil {
		logger.Error("Failed to connect to MongoDB", zap.Error(err))
		return err
	}

	repo := newsRepo.NewNewsRepository(dbClient.Database(viper.GetString("database.db_name")))
	service := news.NewService(repo, logger, cfg)

	browser, err := br.SetupSelenium(cfg)
	if err != nil {
		logger.Error("Failed to setup selenium", zap.Error(err))
		return err
	}
	defer browser.Close()

	scraper := kontan.NewScraper(logger, kontan.NewDefaultBrowser(browser.Driver))

	now := time.Now()
	var startDate, endDate time.Time

	if scrapeStartDate == "" {
		startDate = now
	} else {
		startDate, err = time.Parse("2006-01-02", scrapeStartDate)
		if err != nil {
			logger.Error("Failed to parse start-date", zap.Error(err))
			return err
		}
	}

	if scrapeEndDate == "" {
		endDate = startDate
	} else {
		endDate, err = time.Parse("2006-01-02", scrapeEndDate)
		if err != nil {
			logger.Error("Failed to parse end-date", zap.Error(err))
			return err
		}
	}

	var ids []bson.ObjectID
	err = scraper.Scrape(ctx, startDate, endDate, func(n *news.News) error {
		if err := service.Create(ctx, n); err != nil {
			return err
		}
		ids = append(ids, n.ID)
		return nil
	})

	if err != nil {
		logger.Error("Scraping finished with error", zap.Error(err))
	}

	if len(ids) > 0 {
		if err := service.Summarize(ctx, ids); err != nil {
			logger.Error("Failed to summarize news", zap.Error(err))
		}
	}
	return nil
}
