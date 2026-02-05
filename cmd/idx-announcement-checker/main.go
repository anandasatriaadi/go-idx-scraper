package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/browser"
	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/db"
	"github.com/anandasatriaadi/go-idx-scraper/internal/domain/announcement"
	"github.com/anandasatriaadi/go-idx-scraper/internal/helper"
	"github.com/anandasatriaadi/go-idx-scraper/internal/infrastructure/idx"
	"github.com/anandasatriaadi/go-idx-scraper/internal/infrastructure/persistence/mongo"
	"github.com/chromedp/chromedp"
	"go.mongodb.org/mongo-driver/v2/bson"
	mongoDriver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// --- Main Logic ---

func main() {
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
		OutputPaths:      []string{logFile},
		ErrorOutputPaths: []string{logFile},
	}
	logger, err := zCfg.Build()
	if err != nil {
		log.Printf("Failed to create logger: %v", err)
		os.Exit(1)
	}
	defer logger.Sync()
	if len(os.Args) < 2 {
		logger.Error("no config file provided", zap.String("usage", os.Args[0]+" <config_file>"))
		os.Exit(1)
	}
	configPath := os.Args[1]
	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("Failed to read config", zap.Error(err))
		os.Exit(1)
	}

	ctx, cancel := browser.SetupChromeDp(cfg)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic occurred", zap.Any("panic", r))
			cancel()
			os.Exit(1)
		}
		cancel()
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Info("Received interrupt signal, shutting down gracefully...")
		cancel()
	}()

	database, err := db.New(logger)
	if err != nil {
		logger.Error("Failed to create database", zap.Error(err))
		cancel()
		os.Exit(1)
	}

	dbInstance := database.GetDatabase("idx")
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
			cancel()
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

	var data string
	// get page data as string
	dateTo := time.Now().AddDate(0, 0, 1).Format("20060102")
	url := `https://www.idx.co.id/primary/ListedCompany/GetAnnouncement?kodeEmiten=&emitenType=s&indexFrom=0&pageSize=500&dateFrom=` + dateFrom + `&dateTo=` + dateTo + `&lang=id&keyword=`
	logger.Info("Starting data scrape", zap.String("dateFrom", dateFrom), zap.String("dateTo", dateTo))
	if err := chromedp.Run(ctx, getPageData(url, &data)); err != nil {
		logger.Error("Failed to run chromedp", zap.Error(err))
		cancel()
		os.Exit(1)
	}
	logger.Info("Data scraped successfully")
	// Optional: save debug data
	// if err := os.WriteFile("data.json", []byte(data), 0o644); err != nil { ... }

	var resp idx.APIResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		logger.Error("Failed to unmarshal", zap.Error(err))
		cancel()
		os.Exit(1)
	}

	announcements, err := idx.ParseAPIResponse(resp)
	if err != nil {
		logger.Error("Failed to parse API response", zap.Error(err))
		cancel()
		os.Exit(1)
	}
	logger.Info("Announcements parsed", zap.Int("count", len(announcements)))

	// Check existing using FindAll with projection
	// Note: Repo FindAll signature is strict. We might need to implement FindIDs or similar if performance matters.
	// For now, fetching full docs limit 500 is okay as per original code logic.
	existingDocs, err := aRepo.FindAll(ctx, bson.M{}, options.Find().SetProjection(bson.D{{Key: "_id", Value: 1}}).SetSort(bson.M{"created_date": -1}).SetLimit(500))
	if err != nil {
		logger.Error("Failed to check existing announcements", zap.Error(err))
		cancel()
		os.Exit(1)
	}
	exists := make(map[string]bool)
	for _, doc := range existingDocs {
		exists[doc.ID] = true
	}
	logger.Info("Existing announcements checked", zap.Int("existing", len(exists)))

	// Filter announcements based on latestDate and existence in DB
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
		goto done
	}

	// Update latestID if we have new data
	if len(filtered) > 0 {
		// Find latest from DB again to be sure
		latestAnnDocs, err := aRepo.FindAll(ctx, bson.M{}, options.Find().SetSort(bson.M{"created_date": -1}).SetLimit(1))
		if err != nil {
			logger.Error("Failed to get latest announcement", zap.Error(err))
			cancel()
			os.Exit(1)
		}
		if len(latestAnnDocs) > 0 {
			latestDate = latestAnnDocs[0].CreatedDate
		}
	}

	{
		// Save last run
		filter := bson.M{"scriptName": "idx-announcement"}
		update := bson.M{"$set": bson.M{
			"scriptName": "idx-announcement",
			"lastRunAt":  time.Now(),
			"metadata":   bson.M{"latest_date": latestDate},
		}}
		opts := options.UpdateOne().SetUpsert(true)
		if err := lRepo.UpdateOne(ctx, filter, update, opts); err != nil {
			logger.Error("Failed to save last run", zap.Error(err))
			cancel()
			os.Exit(1)
		}
		logger.Info("Last run saved")

		// Filter announcements to exclude specific titles
		excludedTitles := []string{
			"laporanbulananregistrasipemegangefek",
			"penjelasanatasvolatilitastransaksi",
			"penyampaianbuktiiklan",
		}

		var emailFiltered []*announcement.Announcement
		for _, ann := range filtered {
			if ann.JudulPengumuman != nil {
				// Remove non-alphabetic characters and whitespace
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
			content, err := helper.GenerateAnnouncementEmail(emailFiltered)
			if err != nil {
				logger.Error("Failed to generate email content", zap.Error(err))
			} else {
				logger.Info("Email content generated")
			}
			if err := helper.SendAnnouncementMail(content, cfg); err != nil {
				logger.Error("Failed to send email", zap.Error(err))
			} else {
				logger.Info("Email sent successfully")
			}
		} else {
			logger.Info("No announcements to send after filtering excluded titles")
		}
	}

done:
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	logger.Info("Memory usage", zap.Float64("MB", float64(m.Alloc)/(1024*1024)))

	logger.Info("Process completed successfully")
}

// getPageData navigates to the URL and retrieves the page data as a string.
//
// Assumes the page content is JSON text in the body.
func getPageData(urlstr string, res *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.Evaluate(`document.body.innerText`, res),
	}
}
