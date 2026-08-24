package browser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

type SeleniumBrowser struct {
	Service *selenium.Service
	Driver  selenium.WebDriver
}

func SetupSelenium(cfg *config.Config) (*SeleniumBrowser, error) {
	// Find an available port
	port, err := getFreePort()
	if err != nil {
		return nil, fmt.Errorf("failed to get free port: %w", err)
	}

	opts := []selenium.ServiceOption{}

	// Use path from config if provided, otherwise fallback to finding it
	chromeDriverPath := cfg.Browser.ChromeDriverPath
	if chromeDriverPath == "" {
		chromeDriverPath = findChromeDriver()
	}

	service, err := selenium.NewChromeDriverService(chromeDriverPath, port, opts...)

	if err != nil {
		return nil, fmt.Errorf("failed to start ChromeDriver service: %w", err)
	}

	caps := selenium.Capabilities{"browserName": "chrome"}
	args := []string{
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"--disable-gpu",
		"--disable-extensions",
		"--remote-debugging-port=9222",
		"--log-level=1",
		"--safebrowsing-disable-download-protection",
		"--safebrowsing-disable-extension-blacklist",
		"--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36",
	}

	if cfg.Browser.Headless {
		args = append(args, "--headless")
	}

	if cfg.Paths.BrowserProfile != "" {
		args = append(args, "--user-data-dir="+cfg.Paths.BrowserProfile)
	}

	chromeCaps := chrome.Capabilities{
		Path: "",
		Args: args,
		Prefs: map[string]interface{}{
			"download.default_directory":                                       cfg.Paths.DownloadDir,
			"download.prompt_for_download":                                     false,
			"download.directory_upgrade":                                       true,
			"safebrowsing.enabled":                                             true,
			"safebrowsing.disable_download_protection":                         true,
			"profile.default_content_settings.popups":                         0,
			"profile.content_settings.exceptions.automatic_downloads.*.setting": 1,
			"credentials_enable_service":                                       false,
			"profile.password_manager_enabled":                                 false,
		},
		ExcludeSwitches: []string{"enable-automation"},
	}

	caps.AddChrome(chromeCaps)

	driver, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		service.Stop()
		return nil, fmt.Errorf("failed to connect to WebDriver: %w", err)
	}

	// Enable headless / remote downloading via CDP
	if cfg.Paths.DownloadDir != "" {
		enableDownloadsInChrome(port, driver.SessionID(), cfg.Paths.DownloadDir)
	}

	return &SeleniumBrowser{
		Service: service,
		Driver:  driver,
	}, nil
}

func enableDownloadsInChrome(port int, sessionID string, downloadDir string) {
	absDir, err := filepath.Abs(downloadDir)
	if err != nil {
		absDir = downloadDir
	}
	commands := []string{"Page.setDownloadBehavior", "Browser.setDownloadBehavior"}
	for _, cmd := range commands {
		payload := map[string]any{
			"cmd": cmd,
			"params": map[string]any{
				"behavior":      "allow",
				"downloadPath":  absDir,
				"eventsEnabled": true,
			},
		}
		body, _ := json.Marshal(payload)
		url := fmt.Sprintf("http://localhost:%d/session/%s/chromium/send_command", port, sessionID)
		resp, err := http.Post(url, "application/json", bytes.NewReader(body))
		if err == nil && resp != nil {
			resp.Body.Close()
		}
	}
}

func (b *SeleniumBrowser) Close() {
	if b.Driver != nil {
		b.Driver.Quit()
	}
	if b.Service != nil {
		b.Service.Stop()
	}
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func findChromeDriver() string {
	// Try 'which chromedriver'
	path, err := exec.LookPath("chromedriver")
	if err == nil {
		return path
	}

	// Common locations
	locations := []string{
		"/usr/local/bin/chromedriver",
		"/usr/bin/chromedriver",
		"/opt/homebrew/bin/chromedriver",
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	return "chromedriver" // Fallback to PATH and hope for the best
}
