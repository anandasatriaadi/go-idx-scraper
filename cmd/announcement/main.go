package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	br "github.com/anandasatriaadi/go-idx-scraper/internal/browser"
	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/announcement"
	"github.com/anandasatriaadi/go-idx-scraper/internal/helper"
	"github.com/anandasatriaadi/go-idx-scraper/internal/infra/db/mongo"
	"github.com/anandasatriaadi/go-idx-scraper/internal/infra/idx"
	"github.com/spf13/viper"
	"github.com/tebeka/selenium"
	"go.mongodb.org/mongo-driver/v2/bson"
	mongoDriver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	var configPath string
	var noHeadless bool
	flag.StringVar(&configPath, "config", "config/config.yml", "Path to configuration file")
	flag.BoolVar(&noHeadless, "no-headless", false, "Disable headless mode for browser")
	flag.Parse()

	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Failed to get home dir: %v", err)
		os.Exit(1)
	}
	logDir := filepath.Join(home, ".idx-scraper", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Failed to create log dir: %v", err)
		os.Exit(1)
	}
	logFile := filepath.Join(logDir, "announce_"+time.Now().Format("20060102_150405")+".log")
	zCfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:       "time",
			LevelKey:      "level",
			NameKey:       "logger",
			CallerKey:     "caller",
			MessageKey:    "msg",
			StacktraceKey: "stacktrace",
			LineEnding:    zapcore.DefaultLineEnding,
			EncodeLevel:   zapcore.LowercaseLevelEncoder,
			EncodeTime: zapcore.TimeEncoder(func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Format("02-01-2006T15:04:05"))
			}),
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{logFile, "stdout"},
		ErrorOutputPaths: []string{logFile, "stderr"},
	}
	logger, err := zCfg.Build()
	if err != nil {
		log.Printf("Failed to create logger: %v", err)
		os.Exit(1)
	}
	defer logger.Sync()

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("Failed to read config", zap.Error(err))
		os.Exit(1)
	}

	if noHeadless {
		cfg.SetHeadless(false)
	}

	browser, err := br.SetupSelenium(cfg)
	if err != nil {
		logger.Error("Failed to setup selenium", zap.Error(err))
		os.Exit(1)
	}
	defer browser.Close()

	ctx := context.Background()
	db, err := mongo.NewClient(logger)
	if err != nil {
		logger.Error("Failed to create database", zap.Error(err))
		os.Exit(1)
	}

	dbInstance := db.Database(viper.GetString("database.db_name"))
	lRepo := mongo.NewSystemRepository(dbInstance)
	aRepo := mongo.NewAnnouncementRepository(dbInstance)

	lr, err := lRepo.FindOne(ctx, bson.M{"scriptName": "idx-announcement"})
	var dateFrom string
	var latestDate *time.Time
	if err != nil {
		if err == mongoDriver.ErrNoDocuments {
			dateFrom = time.Unix(0, 0).Format("20060102")
		} else {
			logger.Error("Failed to find last run", zap.Error(err))
			os.Exit(1)
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
	url := `https://www.idx.co.id/primary/ListedCompany/GetAnnouncement?kodeEmiten=&emitenType=s&indexFrom=0&pageSize=500&dateFrom=` + dateFrom + `&dateTo=` + dateTo + `&lang=id&keyword=`
	logger.Info("Starting data scrape", zap.String("dateFrom", dateFrom), zap.String("dateTo", dateTo))

	if err := browser.Driver.Get(url); err != nil {
		logger.Error("Failed to navigate", zap.Error(err))
		os.Exit(1)
	}
	time.Sleep(1 * time.Second)
	body, err := browser.Driver.FindElement(selenium.ByTagName, "body")
	if err != nil {
		logger.Error("Failed to find body", zap.Error(err))
		os.Exit(1)
	}
	data, err := body.Text()
	if err != nil {
		logger.Error("Failed to get text", zap.Error(err))
		os.Exit(1)
	}
	logger.Info("Data scraped successfully")

	var resp idx.APIResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		logger.Info("Raw data", zap.String("data", data))
		logger.Error("Failed to unmarshal", zap.Error(err))
		os.Exit(1)
	}

	announcements, err := idx.ParseAPIResponse(resp)
	if err != nil {
		logger.Error("Failed to parse API response", zap.Error(err))
		os.Exit(1)
	}
	logger.Info("Announcements parsed", zap.Int("count", len(announcements)))

	existingDocs, err := aRepo.FindAll(ctx, bson.M{}, options.Find().SetProjection(bson.D{{Key: "_id", Value: 1}}).SetSort(bson.M{"created_date": -1}).SetLimit(500))
	if err != nil {
		logger.Error("Failed to check existing announcements", zap.Error(err))
		os.Exit(1)
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
		annStr := ann.CreatedDate.Format("20060102150405")
		annInt, _ := strconv.Atoi(annStr)
		var latestInt int
		if latestDate != nil {
			latestStr := latestDate.Format("20060102150405")
			latestInt, _ = strconv.Atoi(latestStr)
		}
		if (latestDate == nil || annInt > latestInt) && !exists[ann.ID] {
			logger.Info("New announcement found", zap.String("ID", ann.ID), zap.Time("CreatedDate", *ann.CreatedDate))
			filtered = append(filtered, ann)
		}
	}
	logger.Info("Announcements filtered", zap.Int("new", len(filtered)))

	if len(filtered) > 0 {
		for _, f := range filtered {
			if err := aRepo.Create(ctx, f); err != nil {
				logger.Error("Failed to create announcement", zap.String("ID", f.ID), zap.Error(err))
			}
		}
		logger.Info("Announcements created", zap.Int("count", len(filtered)))
	} else {
		logger.Info("No new announcements to create")
		goto finish
	}

	if len(filtered) > 0 {
		latestAnnDocs, err := aRepo.FindAll(ctx, bson.M{}, options.Find().SetSort(bson.M{"created_date": -1}).SetLimit(1))
		if err != nil {
			logger.Error("Failed to get latest announcement", zap.Error(err))
			os.Exit(1)
		}
		if len(latestAnnDocs) > 0 {
			latestDate = latestAnnDocs[0].CreatedDate
		}
	}

	{
		filter := bson.M{"scriptName": "idx-announcement"}
		update := bson.M{"$set": bson.M{
			"scriptName": "idx-announcement",
			"lastRunAt":  time.Now(),
			"metadata":   bson.M{"latest_date": latestDate},
		}}
		opts := options.UpdateOne().SetUpsert(true)
		if err := lRepo.UpdateOne(ctx, filter, update, opts); err != nil {
			logger.Error("Failed to save last run", zap.Error(err))
			os.Exit(1)
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
				re := regexp.MustCompile(`[^a-zA-Z]`)
				normalizedTitle := re.ReplaceAllString(*ann.JudulPengumuman, "")
				normalizedTitle = strings.TrimSpace(normalizedTitle)
				normalizedTitle = strings.ToLower(normalizedTitle)

				excluded := false
				for _, pattern := range excludedTitles {
					if regexp.MustCompile(`^` + pattern).MatchString(normalizedTitle) {
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

finish:
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	logger.Info("Memory usage", zap.Float64("MB", float64(mem.Alloc)/(1024*1024)))
	logger.Info("Process completed successfully")
}
