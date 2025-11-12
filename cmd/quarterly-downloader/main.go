package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/anandasatriaadi/go-idx-scraper/internal/browser"
	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/email"
	"github.com/anandasatriaadi/go-idx-scraper/internal/excel"
	"github.com/anandasatriaadi/go-idx-scraper/internal/stock"
	"github.com/tebeka/selenium"
)

// maxServerErrorRetry defines the maximum number of retries for server errors during stock downloads.
const maxServerErrorRetry = 3

// main is the entry point of the IDX scraper application.
// It initializes the browser, loads configuration and stock list,
// downloads financial statements, processes files, and sends notifications if needed.
func main() {
	configPath := parseArgs()
	cfg := loadConfig(configPath)
	service, seleniumBr := browser.SetupBrowser(*cfg)
	defer service.Stop()
	defer seleniumBr.Quit()
	stockList := loadStocks(cfg)
	period, modeWithPeriod := prepParams(cfg)
	downloadAll(seleniumBr, cfg, stockList, period, modeWithPeriod)
	processFiles(cfg)
	sendIfNeeded(cfg, period)
}

// parseArgs parses the command line arguments and returns the config file path.
// It expects exactly one argument: the path to the config file.
func parseArgs() string {
	if len(os.Args) < 2 {
		log.Fatalf("no config file provided. Usage: %s <config_file>", os.Args[0])
	}
	return os.Args[1]
}

// loadConfig loads the configuration from the specified file path.
// It returns a pointer to the Config struct or fatal logs on error.
func loadConfig(configPath string) *config.Config {
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}
	return cfg
}

// loadStocks loads the current list of stocks from the configured path.
// It returns a slice of stock names.
func loadStocks(cfg *config.Config) []string {
	return stock.LoadCurrent(cfg.Paths.StockList)
}

// prepParams prepares the download parameters based on the config.
// It converts the month period to a string of 'I's and creates a mode with period string.
// Returns the period string and the mode with period string.
func prepParams(cfg *config.Config) (string, string) {
	monthPeriod, err := strconv.Atoi(cfg.Download.MonthPeriod)
	if err != nil {
		monthPeriod = 0
	}
	period := strings.Repeat("I", monthPeriod)
	modeWithPeriod := fmt.Sprintf("%s%d", cfg.Download.Mode, monthPeriod)
	return period, modeWithPeriod
}

// downloadAll attempts to download all stocks with retry logic on server errors.
// It loops up to maxServerErrorRetry times if server errors occur during downloads.
func downloadAll(seleniumBr selenium.WebDriver, cfg *config.Config, stockList []string, period, modeWithPeriod string) {
	serverErrorOccurred := true
	loopCount := 0
	for serverErrorOccurred && loopCount < maxServerErrorRetry {
		loopCount++
		serverErrorOccurred = false
		for _, stockName := range stockList {
			downloadIfNeeded(seleniumBr, cfg, stockName, period, modeWithPeriod, &serverErrorOccurred)
		}
	}
}

// downloadIfNeeded downloads the financial statement for a stock if the file does not already exist.
// It checks for server errors by examining the browser title and updates the error flag accordingly.
func downloadIfNeeded(seleniumBr selenium.WebDriver, cfg *config.Config, stockName, period, modeWithPeriod string, serverErrorOccurred *bool) {
	stockName = strings.TrimSpace(stockName)
	if stockName == "" {
		return
	}
	fileLoc := buildPath(cfg, period, stockName)
	if _, err := os.Stat(fileLoc); os.IsNotExist(err) {
		urlToOpen := buildURL(cfg, modeWithPeriod, period, stockName)
		if err := seleniumBr.Get(urlToOpen); err != nil {
			log.Printf("Failed to load page: %v\n", err)
			return
		}
		if seleniumBrTitle, err := seleniumBr.Title(); err == nil {
			*serverErrorOccurred = browser.CheckBrowserTitle(seleniumBrTitle, stockName)
		}
		if err := seleniumBr.Get("data:,"); err != nil {
			log.Printf("Something happens")
		}
	} else {
		fmt.Printf("SKIPPING ::: %s\n", stockName)
	}
}

// processFiles processes the downloaded Excel files.
// It calls the excel package to handle the files in the download and check paths.
func processFiles(cfg *config.Config) {
	excel.ProcessDownloadedFiles(cfg.Paths.Download, cfg.Paths.Check)
}

// sendIfNeeded checks for downloaded stocks and sends a notification email if any were downloaded.
// It finds downloaded stocks, sorts them, moves files, generates mail content, and sends the email.
func sendIfNeeded(cfg *config.Config, period string) {
	downloadedStocks := email.FindDownloadedStocks(*cfg)
	if len(downloadedStocks) > 0 {
		sort.Strings(downloadedStocks)
		email.MoveFiles(*cfg)
		mailContent := email.GenerateMailContent(downloadedStocks, period, *cfg)
		email.SendMail(mailContent, period, *cfg)
	}
}

// buildPath constructs the file path for a stock's financial statement based on the config and parameters.
// It handles different modes (e.g., AUDIT) and returns the full file path.
func buildPath(cfg *config.Config, period, stockName string) string {
	if cfg.Download.Mode == "AUDIT" {
		return filepath.Join(cfg.Paths.Check, fmt.Sprintf("FinancialStatement-%s-Tahunan-%s.xlsx", cfg.Download.Year, stockName))
	}
	return filepath.Join(cfg.Paths.Check, fmt.Sprintf("FinancialStatement-%s-%s-%s.xlsx", cfg.Download.Year, period, stockName))
}

// buildURL constructs the URL for downloading a stock's financial statement based on the config and parameters.
// It handles different modes (e.g., AUDIT) and returns the full URL.
func buildURL(cfg *config.Config, modeWithPeriod, period, stockName string) string {
	if cfg.Download.Mode == "AUDIT" {
		return fmt.Sprintf("https://www.idx.co.id/Portals/0/StaticData/ListedCompanies/Corporate_Actions/New_Info_JSX/Jenis_Informasi/01_Laporan_Keuangan/02_Soft_Copy_Laporan_Keuangan//Laporan%%20Keuangan%%20Tahun%%20%s/%s/%s/FinancialStatement-%s-Tahunan-%s.xlsx", cfg.Download.Year, "Audit", stockName, cfg.Download.Year, stockName)
	}
	return fmt.Sprintf("https://www.idx.co.id/Portals/0/StaticData/ListedCompanies/Corporate_Actions/New_Info_JSX/Jenis_Informasi/01_Laporan_Keuangan/02_Soft_Copy_Laporan_Keuangan//Laporan%%20Keuangan%%20Tahun%%20%s/%s/%s/FinancialStatement-%s-%s-%s.xlsx", cfg.Download.Year, modeWithPeriod, stockName, cfg.Download.Year, period, stockName)
}
