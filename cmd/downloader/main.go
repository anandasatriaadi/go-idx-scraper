package main

import (
	"context"
	"flag"
	"fmt"
	"io"
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
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/finreport"
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
	"github.com/anandasatriaadi/go-idx-scraper/internal/helper"
	"github.com/anandasatriaadi/go-idx-scraper/internal/infra/db/mongo"
	infra "github.com/anandasatriaadi/go-idx-scraper/internal/infra/xbrl"
	"github.com/tebeka/selenium"
	"go.uber.org/zap"
)

const (
	maxServerErrorRetry = 3
	retryDelay          = 5 * time.Second
)

var errorTitles = []string{"404 -", "404 not found", "page not found", "tidak ditemukan", "503", "attention required", "just a moment"}

func main() {
	if err := run(); err != nil {
		log.Printf("Application failed: %v", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		configPath   string
		noHeadless   bool
		tickerFlag   string
		yearsFlag    string
		periodsFlag  string
		fileTypeFlag string
		parseFlag    bool
		cleanFlag    bool
	)

	flag.StringVar(&configPath, "config", "config/config.yml", "Path to configuration file")
	flag.BoolVar(&noHeadless, "no-headless", false, "Disable headless mode for browser")
	flag.StringVar(&tickerFlag, "ticker", "", "Single ticker or comma-separated list (e.g. BBRI or BBRI,TLKM)")
	flag.StringVar(&yearsFlag, "years", "", "Comma-separated years or range (e.g. 2021,2022,2023 or 2021-2025)")
	flag.StringVar(&periodsFlag, "periods", "", "Comma-separated periods (e.g. TW1,TW2,TW3,Audit or I,II,III,Tahunan)")
	flag.StringVar(&fileTypeFlag, "file-type", "instance.zip", "File type to download: instance.zip, inlineXBRL.zip, or .xlsx")
	flag.BoolVar(&parseFlag, "parse", false, "Stream & parse each downloaded XBRL/zip file into MongoDB xbrl_statements")
	flag.BoolVar(&cleanFlag, "clean", false, "Clear download directory before starting")
	flag.Parse()

	logger, err := helper.NewLogger("downloader")
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
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

	if err := os.MkdirAll(cfg.Paths.DownloadDir, 0755); err != nil {
		logger.Error("Failed to create download directory", zap.Error(err))
		return err
	}
	if err := os.MkdirAll(cfg.Paths.CheckDir, 0755); err != nil {
		logger.Error("Failed to create check directory", zap.Error(err))
		return err
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
		logger.Error("Failed to create database client", zap.Error(err))
		return err
	}
	dbInstance := db.Database(cfg.Database.DbName)
	fRepo := mongo.NewFinancialReportRepository(dbInstance)
	fSvc := finreport.NewService(fRepo, logger)
	xbrlRepo := mongo.NewXBRLRepository(dbInstance)

	if cleanFlag {
		logger.Info("Cleaning download directory as requested by -clean")
		clearDownloadDir(cfg, logger)
	}

	// Parse tickers
	var tickers []string
	if tickerFlag != "" {
		for _, t := range strings.Split(tickerFlag, ",") {
			t = strings.ToUpper(strings.TrimSpace(t))
			if t != "" {
				tickers = append(tickers, t)
			}
		}
	} else {
		loaded, err := helper.LoadCurrent(cfg.Paths.IssuerList, logger)
		if err != nil {
			logger.Error("Failed to load issuer list", zap.Error(err))
			return err
		}
		tickers = loaded
	}

	// Parse years
	years, err := parseYears(yearsFlag, cfg.Download.Year)
	if err != nil {
		logger.Error("Failed to parse years", zap.Error(err))
		return err
	}

	// Parse periods
	periods := parsePeriods(periodsFlag, cfg, logger)

	fileType := strings.TrimSpace(fileTypeFlag)
	if fileType == "" {
		fileType = "instance.zip"
	}
	isXlsx := strings.EqualFold(fileType, ".xlsx") || strings.EqualFold(fileType, "xlsx")

	logger.Info("Starting download job",
		zap.Int("tickers_count", len(tickers)),
		zap.Ints("years", years),
		zap.Strings("periods", periods),
		zap.String("file_type", fileType),
		zap.Bool("auto_parse", parseFlag),
	)

	// In default batch xlsx mode (no explicit ticker/years/periods filters), process version updates first
	var updatedStocks []string
	if tickerFlag == "" && yearsFlag == "" && periodsFlag == "" && isXlsx {
		clearDownloadDir(cfg, logger)
		updatedStocks = processVersionUpdates(ctx, fSvc, cfg, browser.Driver, logger)
		if err := helper.MoveFiles(logger, cfg); err != nil {
			logger.Error("Failed to move files", zap.Error(err))
		}
		clearDownloadDir(cfg, logger)
	}

	downloadedCount := 0
	for _, stockName := range tickers {
		select {
		case <-ctx.Done():
			logger.Info("Shutdown requested during stock processing")
			return nil
		default:
		}

		for _, year := range years {
			select {
			case <-ctx.Done():
				logger.Info("Shutdown requested during stock processing")
				return nil
			default:
			}

			for _, period := range periods {
				select {
				case <-ctx.Done():
					logger.Info("Shutdown requested during stock processing")
					return nil
				default:
				}

				ps, modePeriod := finreport.NormalizePeriod(period)

				xlsxName := fmt.Sprintf("FinancialStatement-%d-%s-%s.xlsx", year, ps, stockName)
				instanceZipName := fmt.Sprintf("FinancialStatement-%d-%s-%s-instance.zip", year, modePeriod, stockName)
				inlineZipName := fmt.Sprintf("FinancialStatement-%d-%s-%s-inlineXBRL.zip", year, modePeriod, stockName)

				// Check if already downloaded
				alreadyExists := false
				var existingPath string
				for _, fn := range []string{instanceZipName, inlineZipName, xlsxName} {
					tp := filepath.Join(cfg.Paths.DownloadDir, fn)
					cp := filepath.Join(cfg.Paths.CheckDir, fn)
					if fileExistsAndNotEmpty(tp) {
						alreadyExists = true
						existingPath = tp
						break
					}
					if fileExistsAndNotEmpty(cp) {
						alreadyExists = true
						existingPath = cp
						break
					}
				}

				if alreadyExists && !cleanFlag {
					logger.Info("Filing already downloaded, skipping download", zap.String("file", filepath.Base(existingPath)))
					if parseFlag {
						parseAndUpsertXBRL(ctx, existingPath, xbrlRepo, logger)
					}
					continue
				}

				candidates := []struct {
					url                  string
					expectedDownloadName string
					targetFilename       string
					format               string
				}{
					{
						url:                  fSvc.ConstructXBRLReportURL(year, ps, stockName, "instance.zip"),
						expectedDownloadName: "instance.zip",
						targetFilename:       instanceZipName,
						format:               "instance.zip",
					},
					{
						url:                  fSvc.ConstructXBRLReportURL(year, ps, stockName, "inlineXBRL.zip"),
						expectedDownloadName: "inlineXBRL.zip",
						targetFilename:       inlineZipName,
						format:               "inlineXBRL.zip",
					},
					{
						url:                  fSvc.ConstructReportURL(year, ps, stockName),
						expectedDownloadName: xlsxName,
						targetFilename:       xlsxName,
						format:               ".xlsx",
					},
				}

				downloadSuccess := false
				var savedPath string
				for _, cand := range candidates {
					select {
					case <-ctx.Done():
						return ctx.Err()
					default:
					}

					logger.Info("Attempting filing download",
						zap.String("ticker", stockName),
						zap.Int("year", year),
						zap.String("period", ps),
						zap.String("format", cand.format),
						zap.String("url", cand.url),
					)

					targetPath := filepath.Join(cfg.Paths.DownloadDir, cand.targetFilename)
					_ = os.Remove(filepath.Join(cfg.Paths.DownloadDir, cand.expectedDownloadName))

					if err := downloadFile(cand.url, browser.Driver, logger); err != nil {
						continue
					}

					if err := waitForDownloadedFile(cfg.Paths.DownloadDir, cand.expectedDownloadName, targetPath, 5*time.Second, logger); err == nil {
						logger.Info("Successfully downloaded filing", zap.String("file", cand.targetFilename), zap.String("format", cand.format))
						downloadSuccess = true
						savedPath = targetPath
						downloadedCount++
						break
					}
				}

				if downloadSuccess {
					if parseFlag {
						parseAndUpsertXBRL(ctx, savedPath, xbrlRepo, logger)
					}
				} else {
					logger.Warn("Filing not available in any format on IDX (instance.zip, inlineXBRL.zip, .xlsx)",
						zap.String("ticker", stockName),
						zap.Int("year", year),
						zap.String("period", ps),
					)
				}
			}
		}
	}

	logger.Info("Download process finished", zap.Int("total_downloaded", downloadedCount))

	if tickerFlag == "" && yearsFlag == "" && periodsFlag == "" && isXlsx {
		if err := helper.MoveFiles(logger, cfg); err != nil {
			logger.Error("Failed to move files", zap.Error(err))
		}

		newStocks := helper.FindDownloadedStocks(cfg)
		if len(newStocks) > 0 || len(updatedStocks) > 0 {
			content, err := helper.GenerateNewReportEmail(newStocks, updatedStocks, cfg)
			if err != nil {
				logger.Error("Failed to generate email content", zap.Error(err))
			} else if err := helper.SendMail(content, "", cfg); err != nil {
				logger.Error("Failed to send email", zap.Error(err))
			}
		}
	}

	return nil
}

func parseYears(yearsFlag string, defaultYear string) ([]int, error) {
	target := strings.TrimSpace(yearsFlag)
	if target == "" {
		target = strings.TrimSpace(defaultYear)
	}
	if target == "" {
		return nil, fmt.Errorf("no year specified")
	}

	var years []int
	seen := make(map[int]bool)
	parts := strings.Split(target, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.Contains(p, "-") {
			rangeParts := strings.Split(p, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid year range: %s", p)
			}
			start, err1 := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err1 != nil || err2 != nil {
				return nil, fmt.Errorf("invalid year range numbers: %s", p)
			}
			step := 1
			if start > end {
				step = -1
			}
			for y := start; ; y += step {
				if !seen[y] {
					seen[y] = true
					years = append(years, y)
				}
				if y == end {
					break
				}
			}
		} else {
			y, err := strconv.Atoi(p)
			if err != nil {
				return nil, fmt.Errorf("invalid year: %s", p)
			}
			if !seen[y] {
				seen[y] = true
				years = append(years, y)
			}
		}
	}
	return years, nil
}

func parsePeriods(periodsFlag string, cfg *config.Config, logger *zap.Logger) []string {
	if strings.TrimSpace(periodsFlag) != "" {
		var periods []string
		for _, p := range strings.Split(periodsFlag, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				periods = append(periods, p)
			}
		}
		return periods
	}

	if cfg.Download.Mode == "AUDIT" || strings.EqualFold(cfg.Download.Mode, "audit") {
		return []string{"Tahunan"}
	}
	period, _ := preparePeriodParams(cfg, logger)
	return []string{period}
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

func checkPageTitleForErrors(title, url string, logger *zap.Logger) error {
	titleLower := strings.ToLower(title)
	for _, errorStr := range errorTitles {
		if strings.Contains(titleLower, errorStr) {
			switch errorStr {
			case "404", "document":
				logger.Info("Stock document not found", zap.String("url", url), zap.String("title", titleLower))
				return fmt.Errorf("not found")
			case "503":
				logger.Info("Server error", zap.String("url", url), zap.String("title", titleLower))
				return fmt.Errorf("server error")
			case "attention required", "just a moment":
				logger.Info("Bot detector encountered", zap.String("url", url), zap.String("title", titleLower))
				return fmt.Errorf("bot detector")
			}
		}
	}
	return nil
}

func downloadFile(url string, driver selenium.WebDriver, logger *zap.Logger) error {
	if err := driver.Get(url); err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}

	// Give browser time to process headers and trigger download
	time.Sleep(1 * time.Second)

	title, _ := driver.Title()
	logger.Info("Browser title retrieved", zap.String("title", title))

	if err := checkPageTitleForErrors(title, url, logger); err != nil {
		_ = driver.Get("data:,")
		return err
	}

	return nil
}

func waitForDownloadedFile(downloadDir string, expectedName string, targetPath string, timeout time.Duration, logger *zap.Logger) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		time.Sleep(300 * time.Millisecond)

		// Wait while any Chrome download is in progress
		crdownloadFiles, _ := filepath.Glob(filepath.Join(downloadDir, "*.crdownload"))
		if len(crdownloadFiles) > 0 {
			continue
		}

		// Check if targetPath already exists
		if info, err := os.Stat(targetPath); err == nil && info.Size() > 0 {
			return nil
		}

		// Check if expectedName exists in download directory
		expectedPath := filepath.Join(downloadDir, expectedName)
		if info, err := os.Stat(expectedPath); err == nil && info.Size() > 0 {
			if expectedPath != targetPath {
				if err := os.Rename(expectedPath, targetPath); err != nil {
					if err := copyAndRemove(expectedPath, targetPath); err != nil {
						return fmt.Errorf("moving downloaded file %s to %s: %w", expectedPath, targetPath, err)
					}
				}
			}
			return nil
		}

		// Look for numbered variant (e.g., instance (1).zip)
		entries, err := os.ReadDir(downloadDir)
		if err == nil {
			baseExpected := strings.TrimSuffix(expectedName, filepath.Ext(expectedName))
			extExpected := filepath.Ext(expectedName)
			for _, entry := range entries {
				if entry.IsDir() || strings.HasSuffix(entry.Name(), ".crdownload") {
					continue
				}
				fullEntryPath := filepath.Join(downloadDir, entry.Name())
				if fullEntryPath == targetPath {
					if info, err := os.Stat(targetPath); err == nil && info.Size() > 0 {
						return nil
					}
				}
				if strings.HasPrefix(entry.Name(), baseExpected) && strings.HasSuffix(entry.Name(), extExpected) {
					if info, err := os.Stat(fullEntryPath); err == nil && info.Size() > 0 {
						if err := os.Rename(fullEntryPath, targetPath); err != nil {
							if err := copyAndRemove(fullEntryPath, targetPath); err != nil {
								return fmt.Errorf("moving variant file %s to %s: %w", fullEntryPath, targetPath, err)
							}
						}
						return nil
					}
				}
			}
		}
	}
	return fmt.Errorf("timeout waiting for %s to download", expectedName)
}

func copyAndRemove(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	if err = out.Sync(); err != nil {
		return err
	}
	_ = in.Close()
	return os.Remove(src)
}

func clearDownloadDir(cfg *config.Config, logger *zap.Logger) {
	entries, err := os.ReadDir(cfg.Paths.DownloadDir)
	if err != nil {
		logger.Warn("Failed to read download dir", zap.Error(err))
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			if err := os.Remove(filepath.Join(cfg.Paths.DownloadDir, entry.Name())); err != nil {
				logger.Warn("Failed to remove file", zap.String("file", entry.Name()), zap.Error(err))
			}
		}
	}
}

func processVersionUpdates(ctx context.Context, fSvc *finreport.Service, cfg *config.Config, driver selenium.WebDriver, logger *zap.Logger) []string {
	reports, err := fSvc.FindAllNotLatest(ctx)
	if err != nil {
		logger.Error("Failed to find reports needing update", zap.Error(err))
		return nil
	}

	var updatedStocks []string
	for _, r := range reports {
		select {
		case <-ctx.Done():
			logger.Info("Shutdown requested during version updates")
			return updatedStocks
		default:
		}

		filename := fmt.Sprintf("FinancialStatement-%d-%s-%s.xlsx", r.Year, r.PeriodString, r.IssuerCode)
		filePath := filepath.Join(cfg.Paths.DownloadDir, filename)

		logger.Info("Processing updated report", zap.String("issuer", r.IssuerCode), zap.Int("year", r.Year), zap.String("period", r.PeriodString))

		url := fSvc.ConstructReportURL(r.Year, r.PeriodString, r.IssuerCode)
		if err := downloadFile(url, driver, logger); err != nil {
			logger.Warn("Failed to download updated report", zap.String("issuer", r.IssuerCode), zap.Error(err))
			continue
		}

		_ = waitForDownloadedFile(cfg.Paths.DownloadDir, filename, filePath, 15*time.Second, logger)

		if err := fSvc.MarkAsDownloaded(ctx, r.ID, url); err != nil {
			logger.Error("Failed to mark as downloaded", zap.String("issuer", r.IssuerCode), zap.Error(err))
		}

		updatedStocks = append(updatedStocks, r.IssuerCode)
	}

	return updatedStocks
}

func parseAndUpsertXBRL(ctx context.Context, filePath string, repo xbrl.Repository, logger *zap.Logger) {
	stmt, err := infra.ParseAnyFiling(filePath)
	if err != nil {
		logger.Warn("Failed to parse financial statement file", zap.String("path", filePath), zap.Error(err))
		return
	}

	if err := xbrl.ComputeValuationAndRatios(stmt, nil, 0); err != nil {
		logger.Warn("Valuation calculation failed", zap.String("ticker", stmt.Ticker), zap.Error(err))
	}

	if err := repo.Upsert(ctx, stmt); err != nil {
		logger.Error("Failed to upsert statement into MongoDB", zap.String("ticker", stmt.Ticker), zap.Error(err))
	} else {
		logger.Info("Successfully parsed & ingested statement into MongoDB",
			zap.String("ticker", stmt.Ticker),
			zap.Int("year", stmt.Year),
			zap.String("period", stmt.Period),
		)
	}
}

func fileExistsAndNotEmpty(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir() && info.Size() > 0
}
