package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"

	br "github.com/anandasatriaadi/go-idx-scraper/internal/browser"
	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/announcement"
	"github.com/anandasatriaadi/go-idx-scraper/internal/helper"
	"github.com/anandasatriaadi/go-idx-scraper/internal/infra/db/mongo"
	"github.com/anandasatriaadi/go-idx-scraper/internal/infra/idx"
	"go.mongodb.org/mongo-driver/v2/bson"
	mongoDriver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

var nonAlphaRegex = regexp.MustCompile(`[^a-zA-Z]`)

func main() {
	if err := run(); err != nil {
		// Logger might not be initialized yet if run fails early
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

	lr, err := lRepo.FindOne(ctx, bson.M{"scriptName": "idx-announcement"})
	var dateFrom string
	var latestDate *time.Time
	if err != nil {
		if err == mongoDriver.ErrNoDocuments {
			dateFrom = time.Unix(0, 0).Format("20060102")
		} else {
			logger.Error("Failed to find last run", zap.Error(err))
			return err
		}
	} else {
		dateFrom = lr.LastRunAt.Format("20060102")
		if ldate, ok := lr.Metadata["latest_date"].(bson.DateTime); ok {
			tempDate := ldate.Time().Add(-8 * time.Hour)
			latestDate = &tempDate
			logger.Info("Loaded latestDate from last run", zap.Time("latestDate", *latestDate))
		}
	}

	dateTo := time.Now().AddDate(0, 0, 1).Format("20060102")
	logger.Info("Starting data scrape", zap.String("dateFrom", dateFrom), zap.String("dateTo", dateTo))

	// Use the IDXProvider adapter to fetch and parse announcements
	announcements, err := idxProvider.Fetch(ctx, dateFrom, dateTo)
	if err != nil {
		logger.Error("Failed to fetch announcements from IDX", zap.Error(err))
		return err
	}
	logger.Info("Announcements fetched successfully", zap.Int("count", len(announcements)))

	existingDocs, err := aRepo.FindAll(ctx, bson.M{}, options.Find().SetProjection(bson.D{{Key: "_id", Value: 1}}).SetSort(bson.M{"created_date": -1}).SetLimit(500))
	if err != nil {
		logger.Error("Failed to check existing announcements", zap.Error(err))
		return err
	}
	exists := make(map[string]bool)
	for _, doc := range existingDocs {
		exists[doc.ID] = true
	}

	var filtered []*announcement.Announcement
	for _, ann := range announcements {
		if ann.CreatedDate == nil {
			continue
		}
		if (latestDate == nil || ann.CreatedDate.After(*latestDate)) && !exists[ann.ID] {
			logger.Info("New announcement found", zap.String("ID", ann.ID), zap.Time("CreatedDate", *ann.CreatedDate))
			filtered = append(filtered, ann)
		}
	}
	logger.Info("Announcements filtered", zap.Int("new", len(filtered)))

	if len(filtered) == 0 {
		logger.Info("No new announcements to create")
	} else {
		for _, f := range filtered {
			if err := aRepo.Create(ctx, f); err != nil {
				logger.Error("Failed to create announcement", zap.String("ID", f.ID), zap.Error(err))
			}
			if err := annSvc.ProcessFinancialReportAnnouncement(ctx, f); err != nil {
				logger.Error("Failed to process finreport announcement", zap.String("ID", f.ID), zap.Error(err))
			}
		}
		logger.Info("Announcements created", zap.Int("count", len(filtered)))

		latestAnnDocs, err := aRepo.FindAll(ctx, bson.M{}, options.Find().SetSort(bson.M{"created_date": -1}).SetLimit(1))
		if err != nil {
			logger.Error("Failed to get latest announcement", zap.Error(err))
			return err
		}
		if len(latestAnnDocs) > 0 {
			latestDate = latestAnnDocs[0].CreatedDate
		}

		filter := bson.M{"scriptName": "idx-announcement"}
		update := bson.M{"$set": bson.M{
			"scriptName": "idx-announcement",
			"lastRunAt":  time.Now(),
			"metadata":   bson.M{"latest_date": latestDate},
		}}
		opts := options.UpdateOne().SetUpsert(true)
		if err := lRepo.UpdateOne(ctx, filter, update, opts); err != nil {
			logger.Error("Failed to save last run", zap.Error(err))
			return err
		}
		logger.Info("Last run saved")

		excludedTitles := []string{
			"laporanbulananregistrasipemegangefek",
			"penjelasanatasvolatilitastransaksi",
			"penyampaianbuktiiklan",
		}

		var emailFiltered []*announcement.Announcement
		for _, ann := range filtered {
			if ann.JudulPengumuman != nil {
				normalizedTitle := nonAlphaRegex.ReplaceAllString(*ann.JudulPengumuman, "")
				normalizedTitle = strings.TrimSpace(normalizedTitle)
				normalizedTitle = strings.ToLower(normalizedTitle)

				excluded := false
				for _, pattern := range excludedTitles {
					if strings.HasPrefix(normalizedTitle, pattern) {
						excluded = true
						logger.Info("Announcement excluded from email", zap.String("title", *ann.JudulPengumuman))
						break
					}
				}
				if !excluded {
					emailFiltered = append(emailFiltered, ann)
				}
			} else {
				emailFiltered = append(emailFiltered, ann)
			}
		}

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
