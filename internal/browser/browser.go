package browser

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

// SetupBrowser initializes and returns a Selenium WebDriver instance
// The returned service and webdriver should be deferred after calling this function
func SetupBrowser(config config.Config) (*selenium.Service, selenium.WebDriver) {
	service, err := selenium.NewChromeDriverService(config.Paths.ChromeDriver, 4444)
	if err != nil {
		log.Fatalf("Error starting ChromeDriver server: %v", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting user home directory: %v", err)
	}

	caps := selenium.Capabilities{"browserName": "chrome"}
	chromeCaps := chrome.Capabilities{
		Args: []string{
			// "--headless",
			"--no-sandbox",
			"--disable-dev-shm-usage",
			"--disable-gpu",
			"--remote-debugging-port=9222",
			"--disable-extensions",
			fmt.Sprintf("--user-data-dir=%s/idx-fetch-browser-profile", homeDir),
			"--log-level=1",
			"--safebrowsing-disable-download-protection",
			"--safebrowsing-disable-extension-blacklist",
			"user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36",
		},
		Prefs: map[string]interface{}{
			"download.default_directory":       config.Paths.Download,
			"download.prompt_for_download":     false,
			"download.directory_upgrade":       true,
			"download.extensions_to_open":      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			"credentials_enable_service":       false,
			"profile.password_manager_enabled": false,
		},
		ExcludeSwitches: []string{"enable-automation"},
	}
	caps.AddChrome(chromeCaps)

	browser, err := selenium.NewRemote(caps, "")
	if err != nil {
		log.Fatalf("Failed to open browser session: %s", err)
	}
	return service, browser
}

// CheckBrowserTitle checks if a browser title indicates an error for a stock's webpage.
// It returns true only for server errors (503), false otherwise.
// Prints status messages based on the title content.
func CheckBrowserTitle(browserTitle string, stockName string) bool {
	browserTitleLower := strings.ToLower(browserTitle)
	errorTitles := []string{"404", "document", "503", "attention required", "just a moment"}

	for _, errorStr := range errorTitles {
		if strings.Contains(browserTitleLower, errorStr) {
			switch errorStr {
			case "404", "document":
				fmt.Println("NOT FOUND :::", stockName, "-", browserTitleLower)
			case "503":
				fmt.Println("SERVER ERROR :::", stockName, "-", browserTitleLower)
				return true
			case "attention required":
				break
			case "just a moment":
				fmt.Println("BOT DETECTOR :::", stockName, "-", browserTitleLower, "\x1b[0m")
			}
			return false
		}
	}

	fmt.Println("DOWNLOAD :::", stockName, "-", browserTitleLower, "\x1b[0m")
	return false
}
