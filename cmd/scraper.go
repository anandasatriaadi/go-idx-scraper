package cmd

import (
	"context"
	"log"
	"time"

	br "github.com/anandasatriaadi/go-idx-scraper/internal/browser"
	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/domain/news"
	newsRepo "github.com/anandasatriaadi/go-idx-scraper/internal/infra/db/mongo"
	"github.com/anandasatriaadi/go-idx-scraper/internal/infra/scraper/kontan"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

var (
	scrapeStartDate string
	scrapeEndDate   string
)

var ScrapeKontanCmd = &cobra.Command{
	Use:   "scrape-kontan",
	Short: "Scrape Kontan news",
	Run: func(cmd *cobra.Command, args []string) {
		runScraper()
	},
}

func init() {
	ScrapeKontanCmd.Flags().StringVar(&scrapeStartDate, "start-date", "", "Start date (YYYY-MM-DD), default today")
	ScrapeKontanCmd.Flags().StringVar(&scrapeEndDate, "end-date", "", "End date (YYYY-MM-DD), default today")
}

func runScraper() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("Failed to load config", zap.Error(err))
		return
	}

	// Connect to MongoDB
	ctx := context.Background()
	dbClient, err := newsRepo.NewClient(logger)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}

	// Infrastructure: Repositories
	repo := newsRepo.NewNewsRepository(dbClient.Database(viper.GetString("database.db_names")))

	// Application: Services
	service := news.NewService(repo, logger, cfg)

	// Infrastructure: Scraper
	scraper := kontan.NewScraper(logger)

	// Parse dates
	now := time.Now()
	var startDate, endDate time.Time

	if scrapeStartDate == "" {
		startDate = now
	} else {
		startDate, err = time.Parse("2006-01-02", scrapeStartDate)
		if err != nil {
			logger.Fatal("Failed to parse start-date", zap.Error(err))
		}
	}

	if scrapeEndDate == "" {
		endDate = startDate
	} else {
		endDate, err = time.Parse("2006-01-02", scrapeEndDate)
		if err != nil {
			logger.Fatal("Failed to parse end-date", zap.Error(err))
		}
	}

	// Setup Chromedp context
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
