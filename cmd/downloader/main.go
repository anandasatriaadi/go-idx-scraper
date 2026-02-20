package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	br "github.com/anandasatriaadi/go-idx-scraper/internal/browser"
	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/helper"
	"github.com/tebeka/selenium"
	"go.uber.org/zap"
)

const (
	DownloadTimeout = 30 * time.Second
	PollInterval    = 500 * time.Millisecond
)

var errorTitles = []string{"404", "document", "503", "attention required", "just a moment"}

func main() {
	var configPath string
	var noHeadless bool
	flag.StringVar(&configPath, "config", "config/config.yml", "Path to configuration file")
	flag.BoolVar(&noHeadless, "no-headless", false, "Disable headless mode for browser")
	flag.Parse()

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("Failed to read config", zap.Error(err))
		return
	}

	if noHeadless {
		cfg.SetHeadless(false)
	}

	browser, err := br.SetupSelenium(cfg)
	if err != nil {
		logger.Error("Failed to setup selenium", zap.Error(err))
		return
	}
	defer browser.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
		cancel()
	}()

	issuerList, err := helper.LoadCurrent(cfg.Paths.IssuerList, logger)
	if err != nil {
		logger.Error("Failed to load issuer list", zap.Error(err))
		return
	}
	processStocks(issuerList, cfg, ctx, browser.Driver, logger)
}

func preparePeriodParams(cfg *config.Config, logger *zap.Logger) (period string, modePeriod string) {
	monthPeriod, err := strconv.Atoi(cfg.Download.MonthPeriod)
	if err != nil {
		logger.Warn("Invalid MonthPeriod, using 0", zap.String("value", cfg.Download.MonthPeriod), zap.Error(err))
		monthPeriod = 0
	}
	period = strings.Repeat("I", monthPeriod)
	modePeriod = fmt.Sprintf("%s%d", cfg.Download.Mode, monthPeriod)
	return period, modePeriod
}

func checkPageTitleForErrors(title, stockName string, logger *zap.Logger) error {
	titleLower := strings.ToLower(title)
	for _, errorStr := range errorTitles {
		if strings.Contains(titleLower, errorStr) {
			switch errorStr {
			case "404", "document":
				logger.Info("Stock not found", zap.String("stock", stockName), zap.String("title", titleLower))
				return fmt.Errorf("not found")
			case "503":
				logger.Info("Server error", zap.String("stock", stockName), zap.String("title", titleLower))
				return fmt.Errorf("server error")
			case "attention required", "just a moment":
				logger.Info("Bot detector", zap.String("stock", stockName), zap.String("title", titleLower))
				return fmt.Errorf("bot detector")
			}
		}
	}
	return nil
}

func downloadFile(url, filePath string, cfg *config.Config, ctx context.Context, driver selenium.WebDriver, logger *zap.Logger) error {
	if err := driver.Get(url); err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}

	title, _ := driver.Title()
	logger.Info("Browser title retrieved", zap.String("title", title))

	if err := checkPageTitleForErrors(title, filepath.Base(filePath), logger); err != nil {
		return err
	}

	timeout := time.After(DownloadTimeout)
	ticker := time.NewTicker(PollInterval)
	defer ticker.Stop()

waitLoop:
	for {
		select {
		case <-ctx.Done():
			logger.Info("Shutdown requested during download wait")
			os.Remove(filePath)
			return ctx.Err()
		case <-timeout:
			logger.Warn("Download timed out", zap.String("file", filePath))
			return fmt.Errorf("download timeout")
		case <-ticker.C:
			if _, err := os.Stat(filePath); err == nil {
				logger.Info("Download completed", zap.String("file", filePath))
				break waitLoop
			}
		}
	}
	return nil
}

func processStocks(issuerList []string, cfg *config.Config, ctx context.Context, driver selenium.WebDriver, logger *zap.Logger) {
	period, modePeriod := preparePeriodParams(cfg, logger)

	var downloadedStocks []string
	for _, stockName := range issuerList {
		select {
		case <-ctx.Done():
			logger.Info("Shutdown requested during stock processing")
			return
		default:
		}

		logger.Info("Processing stock", zap.String("stock", stockName))

		var filename string
		if cfg.Download.Mode == "AUDIT" {
			filename = fmt.Sprintf("FinancialStatement-%s-Tahunan-%s.xlsx", cfg.Download.Year, stockName)
			modePeriod = "Audit"
		} else {
			filename = fmt.Sprintf("FinancialStatement-%s-%s-%s.xlsx", cfg.Download.Year, period, stockName)
		}

		checkPath := filepath.Join(cfg.Paths.CheckDir, filename)
		if _, err := os.Stat(checkPath); err == nil {
			logger.Info("Skip - File Exists", zap.String("stock", stockName), zap.String("file", filename))
			continue
		}

		url := fmt.Sprintf("https://www.idx.co.id/Portals/0/StaticData/ListedCompanies/Corporate_Actions/New_Info_JSX/Jenis_Informasi/01_Laporan_Keuangan/02_Soft_Copy_Laporan_Keuangan//Laporan%%20Keuangan%%20Tahun%%20%s/%s/%s/%s", cfg.Download.Year, modePeriod, stockName, filename)
		filePath := filepath.Join(cfg.Paths.DownloadDir, filename)

		if err := downloadFile(url, filePath, cfg, ctx, driver, logger); err != nil {
			logger.Error("Failed to download file", zap.String("stock", stockName), zap.Error(err))
			continue
		}
		downloadedStocks = append(downloadedStocks, stockName)
	}

	if len(downloadedStocks) == 0 {
		logger.Info("No new files downloaded")
		return
	}

	err := helper.MoveFiles(logger, cfg)
	if err != nil {
		logger.Error("Failed to move files", zap.Error(err))
	}

	content := helper.GenerateNewReportEmail(downloadedStocks, period, cfg)
	if err := helper.SendMail(content, period, cfg); err != nil {
		logger.Error("Failed to send email", zap.Error(err))
	}
}
