package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	newsApp "github.com/anandasatriaadi/go-idx-scraper/internal/application/news"
	br "github.com/anandasatriaadi/go-idx-scraper/internal/browser"
	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/db"
	"github.com/anandasatriaadi/go-idx-scraper/internal/domain/news"
	newsRepo "github.com/anandasatriaadi/go-idx-scraper/internal/infrastructure/persistence/mongo"
	"github.com/anandasatriaadi/go-idx-scraper/internal/infrastructure/scraper/kontan"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

func main() {
	logger, err := initializeLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	cfg, err := loadConfig()
	if err != nil {
		logger.Error("Failed to load config", zap.Error(err))
		return
	}

	// Connect to MongoDB
	ctx := context.Background()
	dbClient, err := db.New(logger)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}

	// Infrastructure: Repositories
	repo := newsRepo.NewNewsRepository(dbClient.GetDatabase("idx"))

	// Application: Services
	service := newsApp.NewService(repo, logger, cfg)

	// Infrastructure: Scraper
	scraper := kontan.NewScraper(logger)

	// Parse start and end dates from command line arguments
	now := time.Now()
	var startDate, endDate time.Time
	if len(os.Args) == 2 {
		startDate = now
		endDate = now
	} else if len(os.Args) == 3 {
		startDate, err = time.Parse("2006-01-02", os.Args[2])
		if err != nil {
			logger.Fatal("Failed to parse startDate", zap.Error(err))
		}
		endDate = startDate
	} else if len(os.Args) == 4 {
		startDate, err = time.Parse("2006-01-02", os.Args[2])
		if err != nil {
			logger.Fatal("Failed to parse startDate", zap.Error(err))
		}
		endDate, err = time.Parse("2006-01-02", os.Args[3])
		if err != nil {
			logger.Fatal("Failed to parse endDate", zap.Error(err))
		}
	} else {
		logger.Fatal("Invalid number of arguments. Usage: <program> <config_file> [startDate] [endDate]")
	}

	// Setup Chromedp context (This should technically be inside the scraper, but since SetupChromeDp is a global helper, we can pass context to scraper if needed, or scraper manages it.
	// Looking at my scraper impl, it accepts context in Scrape.
	// But `br.SetupChromeDp` creates a context with options.
	// Ideally, Scraper should accept a ChromeContext or create one.
	// My current Scraper.Scrape accepts `ctx`. I will pass the chromeCtx there.
	chromeCtx, cancel := br.SetupChromeDp(cfg)
	defer cancel()

	var ids []bson.ObjectID

	err = scraper.Scrape(chromeCtx, startDate, endDate, func(n *news.News) error {
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
}

// initializeLogger sets up the zap logger
func initializeLogger() (*zap.Logger, error) {
	return zap.NewDevelopment()
}

// loadConfig loads the configuration from the provided file
func loadConfig() (*config.Config, error) {
	if len(os.Args) < 2 {
		return nil, fmt.Errorf("no config file provided, usage: %s <config_file>", os.Args[0])
	}
	configPath := os.Args[1]
	return config.Load(configPath)
}
