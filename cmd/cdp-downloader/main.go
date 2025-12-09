package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	br "github.com/anandasatriaadi/go-idx-scraper/internal/browser"
	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/helper"
	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/chromedp"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("no config file provided. Usage: %s <config_file>", os.Args[0])
	}
	configPath := os.Args[1]
	cfg := loadConfig(configPath)

	// Setup chromedp context
	ctx, cancel := br.SetupChromeDp(cfg)
	defer cancel()

	// Load stocks list and prep
	stockList := helper.LoadCurrent(cfg.Paths.StockList)
	period, modePeriod := prepParams(cfg)

	for _, sName := range stockList {
		fmt.Printf("Processing stock: %s\n", sName)

		var fn string
		if cfg.Download.Mode == "AUDIT" {
			fn = fmt.Sprintf("FinancialStatement-%s-Tahunan-%s.xlsx", cfg.Download.Year, sName)
		} else {
			fn = fmt.Sprintf("FinancialStatement-%s-%s-%s.xlsx", cfg.Download.Year, period, sName)
		}

		if cfg.Download.Mode == "AUDIT" {
			modePeriod = "Audit"
		}

		url := fmt.Sprintf("https://www.idx.co.id/Portals/0/StaticData/ListedCompanies/Corporate_Actions/New_Info_JSX/Jenis_Informasi/01_Laporan_Keuangan/02_Soft_Copy_Laporan_Keuangan//Laporan%%20Keuangan%%20Tahun%%20%s/%s/%s/%s", cfg.Download.Year, modePeriod, sName, fn)
		fp := filepath.Join(cfg.Paths.DownloadDir, fn)

		// Create a temporary HTML page with a download link (data URL to avoid external files)
		html := fmt.Sprintf(`
			<html>
			<body>
				<a id="download-link" href="%s" download="%s">Download File</a>
			</body>
			</html>
		`, url, fn)
		dataURL := "data:text/html;charset=utf-8," + strings.ReplaceAll(html, " ", "%20")

		var title string
		// Navigate to the temporary page and click the download link
		err := chromedp.Run(ctx,
			browser.SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllow).
				WithDownloadPath(cfg.Paths.DownloadDir),
			chromedp.Navigate(dataURL),
			chromedp.Click("#download-link", chromedp.ByID),
			chromedp.Title(&title),
		)
		if err != nil {
			log.Fatalf("Error triggering download: %v", err)
		}
		fmt.Printf("Browser Title: %s\n", title)

		// Wait for the download to complete
		timeout := time.After(30 * time.Second)
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				fmt.Printf("Download timed out: file not found at %s", fp)
			case <-ticker.C:
				err := checkTitle(title, sName)
				if err != nil {
					goto done
				}

				if _, err := os.Stat(fp); err == nil {
					goto done
				}
			}
		}

	done:
	}
}

func loadConfig(configPath string) *config.Config {
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}
	return cfg
}

func prepParams(cfg *config.Config) (period string, modePeriod string) {
	monthPeriod, err := strconv.Atoi(cfg.Download.MonthPeriod)
	if err != nil {
		monthPeriod = 0
	}
	period = strings.Repeat("I", monthPeriod)
	modePeriod = fmt.Sprintf("%s%d", cfg.Download.Mode, monthPeriod)
	return period, modePeriod
}

func checkTitle(title string, sName string) error {
	titleLower := strings.ToLower(title)
	errorTitles := []string{"404", "document", "503", "attention required", "just a moment"}

	for _, errorStr := range errorTitles {
		if strings.Contains(titleLower, errorStr) {
			switch errorStr {
			case "404", "document":
				fmt.Println("NOT FOUND :::", sName, "-", titleLower)
				return fmt.Errorf("not found")
			case "503":
				fmt.Println("SERVER ERROR :::", sName, "-", titleLower)
				return fmt.Errorf("server error")
			case "attention required":
			case "just a moment":
				fmt.Println("BOT DETECTOR :::", sName, "-", titleLower, "\x1b[0m")
				return fmt.Errorf("bot detector")
			}
			return nil
		}
	}

	fmt.Println("DOWNLOAD :::", sName, "-", titleLower, "\x1b[0m")
	return nil
}
