package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	br "github.com/anandasatriaadi/go-idx-scraper/internal/browser"
	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/announcement"
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/system"
	"github.com/anandasatriaadi/go-idx-scraper/internal/helper"
	"github.com/anandasatriaadi/go-idx-scraper/internal/infra/db/mongo"
	"github.com/anandasatriaadi/go-idx-scraper/internal/infra/idx"
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
	flag.StringVar(&configPath, "config", "config/config.yml", "Path to configuration file")
	flag.BoolVar(&noHeadless, "no-headless", false, "Disable headless mode for browser")
	flag.Parse()

	logger, err := helper.NewLogger("announcement")
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	defer logger.Sync()

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("Failed to read config", zap.Error(err))
		return err
	}

	if noHeadless {
		cfg.SetHeadless(false)
	}

	browser, err := br.SetupSelenium(cfg)
	if err != nil {
		logger.Error("Failed to setup selenium", zap.Error(err))
		return err
	}
	defer browser.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
		cancel()
	}()

	db, err := mongo.NewClient(logger)
	if err != nil {
		logger.Error("Failed to create database", zap.Error(err))
		return err
	}

	dbInstance := db.Database(cfg.Database.DbName)
	lRepo := mongo.NewSystemRepository(dbInstance)
	aRepo := mongo.NewAnnouncementRepository(dbInstance)
	fRepo := mongo.NewFinancialReportRepository(dbInstance)

	annSvc := announcement.NewService(aRepo, fRepo, logger)

	// Initialize the IDXProvider adapter (External Service)
	idxProvider := idx.NewIDXProvider(logger, browser.Driver)

	lr, err := lRepo.GetLastRun(ctx, "idx-announcement")
	var dateFrom string
	var latestDate *time.Time
	if err != nil {
		logger.Error("Failed to find last run", zap.Error(err))
		return err
	}
	if lr == nil {
		dateFrom = time.Unix(0, 0).Format("20060102")
	} else {
		dateFrom = lr.LastRunAt.Format("20060102")
		if lr.Metadata != nil {
			if ldate, ok := lr.Metadata["latest_date"].(time.Time); ok {
				tempDate := ldate.Add(-8 * time.Hour)
				latestDate = &tempDate
				logger.Info("Loaded latestDate from last run", zap.Time("latestDate", *latestDate))
			}
		}
	}

	dateTo := time.Now().AddDate(0, 0, 1).Format("20060102")
	logger.Info("Starting data scrape", zap.String("dateFrom", dateFrom), zap.String("dateTo", dateTo))

	// Sync disclosures through the application service use case
	filtered, err := annSvc.SyncDisclosures(ctx, dateFrom, dateTo, idxProvider, latestDate)
	if err != nil {
		logger.Error("Failed to sync disclosures", zap.Error(err))
		return err
	}

	if len(filtered) == 0 {
		logger.Info("No new announcements to create")
	} else {
		latestDate, err = annSvc.GetLatestCreatedDate(ctx)
		if err != nil {
			logger.Error("Failed to get latest announcement date", zap.Error(err))
			return err
		}

		lastRun := &system.LastRun{
			ScriptName: "idx-announcement",
			LastRunAt:  time.Now(),
			Metadata:   map[string]any{"latest_date": latestDate},
		}
		if err := lRepo.SaveLastRun(ctx, lastRun); err != nil {
			logger.Error("Failed to save last run", zap.Error(err))
			return err
		}
		logger.Info("Last run saved")

		emailFiltered := annSvc.FilterDisclosuresForEmail(filtered)

		if len(emailFiltered) > 0 {
			const batchSize = 50
			for i := 0; i < len(emailFiltered); i += batchSize {
				end := i + batchSize
				if end > len(emailFiltered) {
					end = len(emailFiltered)
				}
				batch := emailFiltered[i:end]

				logger.Info("Sending email batch", zap.Int("batch_index", i/batchSize+1), zap.Int("batch_size", len(batch)), zap.Int("total", len(emailFiltered)))

				content, err := helper.GenerateAnnouncementEmail(batch)
				if err != nil {
					logger.Error("Failed to generate email content", zap.Error(err))
					continue
				}

				if err := helper.SendAnnouncementMail(content, cfg); err != nil {
					logger.Error("Failed to send email batch", zap.Int("batch_index", i/batchSize+1), zap.Error(err))
				} else {
					logger.Info("Email batch sent successfully", zap.Int("batch_index", i/batchSize+1))
				}

				if end < len(emailFiltered) {
					time.Sleep(2 * time.Second)
				}
			}
		} else {
			logger.Info("No announcements to send after filtering excluded titles")
		}
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	logger.Info("Memory usage", zap.Float64("MB", float64(mem.Alloc)/(1024*1024)))
	logger.Info("Process completed successfully")
	return nil
}
