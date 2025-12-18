package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/browser"
	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/db"
	"github.com/anandasatriaadi/go-idx-scraper/internal/db/model"
	"github.com/anandasatriaadi/go-idx-scraper/internal/helper"
	"github.com/anandasatriaadi/go-idx-scraper/internal/types"
	"github.com/chromedp/chromedp"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
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

	lRepo := model.NewLastRunRepository(database.GetDatabase("idx"))
	lr, err := lRepo.FindOne(ctx, bson.M{"scriptName": "idx-announcement"})
	var dateFrom string
	var latestID string
	if err != nil {
		if err == mongo.ErrNoDocuments {
			dateFrom = time.Unix(0, 0).Format("20060102")
			latestID = ""
		} else {
			logger.Error("Failed to find last run", zap.Error(err))
			cancel()
			os.Exit(1)
		}
	} else {
		dateFrom = lr.LastRunAt.Format("20060102")
		if lid, ok := lr.Metadata["latestId"].(string); ok {
			latestID = lid
		} else {
			latestID = ""
		}
	}

	var data string
	// get page data as string
	dateTo := time.Now().AddDate(0, 0, 1).Format("20060102")
	url := `https://www.idx.co.id/primary/ListedCompany/GetAnnouncement?kodeEmiten=&emitenType=s&indexFrom=0&pageSize=200&dateFrom=` + dateFrom + `&dateTo=` + dateTo + `&lang=id&keyword=`
	if err := chromedp.Run(ctx, getPageData(url, &data)); err != nil {
		logger.Error("Failed to run chromedp", zap.Error(err))
		cancel()
		os.Exit(1)
	}
	if err := os.WriteFile("data.json", []byte(data), 0o644); err != nil {
		logger.Error("Failed to write file", zap.Error(err))
		cancel()
		os.Exit(1)
	}

	var resp types.APIResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		logger.Error("Failed to unmarshal", zap.Error(err))
		cancel()
		os.Exit(1)
	}

	announcements, err := model.ParseAPIResponse(resp)
	if err != nil {
		logger.Error("Failed to parse API response", zap.Error(err))
		cancel()
		os.Exit(1)
	}

	// Filter announcements based on latestID
	var filtered []*model.Announcement
	for _, ann := range announcements {
		if ann.ID > latestID {
			filtered = append(filtered, ann)
		}
	}

	aRepo := model.NewAnnouncementRepository(database.GetDatabase("idx"))
	if len(filtered) > 0 {
		if _, err := aRepo.CreateMany(ctx, filtered); err != nil {
			logger.Error("Failed to create announcements", zap.Error(err))
			cancel()
			os.Exit(1)
		}
		logger.Info("Announcements created", zap.Int("count", len(filtered)))
	} else {
		logger.Info("No new announcements to create")
		goto done
	}

	// Update latestID if we have new data
	if len(announcements) > 0 {
		latestID = announcements[0].ID
	}

	{
		// Save last run
		filter := bson.M{"scriptName": "idx-announcement"}
		update := bson.M{"$set": bson.M{
			"scriptName": "idx-announcement",
			"lastRunAt":  time.Now(),
			"metadata":   bson.M{"latestId": latestID},
		}}
		opts := options.UpdateOne().SetUpsert(true)
		if _, err := lRepo.UpdateOne(ctx, filter, update, opts); err != nil {
			logger.Error("Failed to save last run", zap.Error(err))
			cancel()
			os.Exit(1)
		}

		content, err := helper.GenerateAnnouncementEmail(announcements)
		if err != nil {
			logger.Error("Failed to generate email content", zap.Error(err))
		}
		if err := helper.SendAnnouncementMail(content, cfg); err != nil {
			logger.Error("Failed to send email", zap.Error(err))
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
