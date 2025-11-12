package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/chromedp"
)

func main() {
	// Mock config (replace with your actual config loading)
	cfg := config.Config{
		Paths: config.PathConfig{
			Download: "D:\\sync-gitrepo\\go-idx-scraper\\downloads", // Set your download directory
		},
	}

	url := "https://www.idx.co.id/Portals/0/StaticData/ListedCompanies/Corporate_Actions/New_Info_JSX/Jenis_Informasi/01_Laporan_Keuangan/02_Soft_Copy_Laporan_Keuangan//Laporan%20Keuangan%20Tahun%202024/Audit/ASII/FinancialStatement-2024-Tahunan-ASII.xlsx"
	filename := "FinancialStatement-2024-Tahunan-ASII.xlsx" // Expected filename; adjust if dynamic
	filepath := filepath.Join(cfg.Paths.Download, filename)

	// Setup browser (inlined from SetupBrowser function)
	// Define browser options (equivalent to your Chrome capabilities)
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		// chromedp.Flag("headless", true), // Uncomment for headless mode
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("remote-debugging-port", "9222"),
		chromedp.Flag("disable-extensions", true),
		chromedp.UserDataDir("D:\\browser-profile"),
		chromedp.Flag("log-level", "1"),
		chromedp.Flag("safebrowsing-disable-download-protection", true),
		chromedp.Flag("safebrowsing-disable-extension-blacklist", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36"),
		chromedp.Flag("credentials_enable_service", false),
		chromedp.Flag("profile.password_manager_enabled", false),
	)

	// Create allocator context (manages browser process)
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)

	// Create task context (for running actions)
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithDebugf(log.Printf))

	// Combined cancel function to clean up both contexts
	combinedCancel := func() {
		cancel()
		allocCancel()
	}
	defer combinedCancel()

	// Create a temporary HTML page with a download link (data URL to avoid external files)
	html := fmt.Sprintf(`
			<html>
			<body>
				<a id="download-link" href="%s" download="%s">Download File</a>
			</body>
			</html>
		`, url, filename)
	dataURL := "data:text/html;charset=utf-8," + strings.ReplaceAll(html, " ", "%20") // Encode spaces

	// Navigate to the temporary page and click the download link
	err := chromedp.Run(ctx,
		browser.SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllow).
			WithDownloadPath(cfg.Paths.Download),
		chromedp.Navigate(dataURL),
		chromedp.Click("#download-link", chromedp.ByID),
	)
	if err != nil {
		log.Fatalf("Error triggering download: %v", err)
	}

	// Wait for the download to complete (poll the directory for the file)
	timeout := time.After(30 * time.Second) // Adjust timeout as needed
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			log.Fatalf("Download timed out: file not found at %s", filepath)
		case <-ticker.C:
			if _, err := os.Stat(filepath); err == nil {
				// File exists, download complete
				fmt.Printf("Downloaded file to: %s\n", filepath)
				goto done
			}
		}
	}

done:
	fmt.Println("Browser test completed. Downloads will be in:", cfg.Paths.Download)
}
