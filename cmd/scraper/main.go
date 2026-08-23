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
	var skipBriefing bool

	flag.StringVar(&configPath, "config", "config/config.yml", "Path to configuration file")
	flag.BoolVar(&noHeadless, "no-headless", false, "Disable headless mode for browser")
	flag.StringVar(&scrapeStartDate, "start-date", "", "Start date (YYYY-MM-DD), default yesterday GMT+8")
	flag.StringVar(&scrapeEndDate, "end-date", "", "End date (YYYY-MM-DD), default today GMT+8")
	flag.BoolVar(&skipBriefing, "skip-briefing", false, "Skip generating daily briefing")
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

	db := dbClient.Database(cfg.Database.DbName)
	repo := newsRepo.NewNewsRepository(db)
	briefingRepo := newsRepo.NewBriefingRepository(db)
	service := news.NewService(repo, logger, cfg)

	browser, err := br.SetupSelenium(cfg)
	if err != nil {
		logger.Error("Failed to setup selenium", zap.Error(err))
		return err
	}
	defer browser.Close()

	scraper := kontan.NewScraper(logger, kontan.NewDefaultBrowser(browser.Driver))

	// GMT+8 (WITA / Singapore / Perth timezone)
	locGMT8 := time.FixedZone("GMT+8", 8*3600)
	nowGMT8 := time.Now().In(locGMT8)

	var startDate, endDate time.Time

	if scrapeStartDate == "" {
		startDate = nowGMT8.AddDate(0, 0, -1) // Yesterday
	} else {
		startDate, err = time.ParseInLocation("2006-01-02", scrapeStartDate, locGMT8)
		if err != nil {
			logger.Error("Failed to parse start-date", zap.Error(err))
			return err
		}
	}

	if scrapeEndDate == "" {
		endDate = nowGMT8 // Today
	} else {
		endDate, err = time.ParseInLocation("2006-01-02", scrapeEndDate, locGMT8)
		if err != nil {
			logger.Error("Failed to parse end-date", zap.Error(err))
			return err
		}
	}

	logger.Info("Starting scrape window in GMT+8", zap.Time("start", startDate), zap.Time("end", endDate))

	var ids []bson.ObjectID
	err = scraper.Scrape(ctx, startDate, endDate, func(n *news.News) error {
		// Idempotency: skip if already in MongoDB
		exists, err := repo.ExistsByLink(ctx, n.Link)
		if err != nil {
			logger.Warn("Failed to check if news exists by link", zap.String("link", n.Link), zap.Error(err))
		} else if exists {
			logger.Debug("News article already exists in DB, skipping", zap.String("link", n.Link))
			return nil
		}

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
		logger.Info("Summarizing newly fetched articles", zap.Int("count", len(ids)))
		if err := service.Summarize(ctx, ids); err != nil {
			logger.Error("Failed to summarize news", zap.Error(err))
		}
	}

	// Generate Daily Briefing and send email
	if !skipBriefing {
		briefing, err := service.GenerateDailyBriefing(ctx, nowGMT8, briefingRepo)
		if err != nil {
			logger.Error("Failed to generate Daily Briefing", zap.Error(err))
		} else if briefing != nil {
			logger.Info("Sending Daily Market Briefing email", zap.String("title", briefing.Title))
			if err := helper.SendMail(briefing.RawMarkdown, briefing.Title, cfg); err != nil {
				logger.Error("Failed to send Daily Briefing email", zap.Error(err))
			} else {
				logger.Info("Daily Briefing email successfully dispatched")
			}
		}
	}

	return nil
}
