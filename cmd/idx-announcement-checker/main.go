package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/browser"
	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/db"
	"github.com/anandasatriaadi/go-idx-scraper/internal/db/model"
	"github.com/anandasatriaadi/go-idx-scraper/internal/types"
	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
)

// --- Main Logic ---

func main() {
	logger, err := zap.NewProduction()
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
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Info("Received interrupt signal, shutting down gracefully...")
		cancel()
	}()

	var data string
	// get page data as string
	dateTo := time.Now().AddDate(0, 0, 1).Format("20060102")
	url := `https://www.idx.co.id/primary/ListedCompany/GetAnnouncement?kodeEmiten=&emitenType=s&indexFrom=0&pageSize=100&dateFrom=19010101&dateTo=` + dateTo + `&lang=id&keyword=`
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

	database, err := db.New(logger)
	if err != nil {
		logger.Error("Failed to create database", zap.Error(err))
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

	aRepo := model.NewAnnouncementRepository(database.GetDatabase("idx"))
	if _, err := aRepo.CreateMany(ctx, announcements); err != nil {
		logger.Error("Failed to create announcements", zap.Error(err))
		cancel()
		os.Exit(1)
	}
	logger.Info("Announcements created", zap.Int("count", len(announcements)))

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
