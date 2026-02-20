package browser

import (
	"fmt"
	"net"
	"os"
	"os/exec"

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
	chromeCaps := chrome.Capabilities{
		Path: "",
		Args: []string{
			"--no-sandbox",
			"--disable-dev-shm-usage",
			"--disable-gpu",
			"--disable-extensions",
			"--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36",
		},
	}

	if cfg.Browser.Headless {
		chromeCaps.Args = append(chromeCaps.Args, "--headless")
	}

	if cfg.Paths.BrowserProfile != "" {
		chromeCaps.Args = append(chromeCaps.Args, "--user-data-dir="+cfg.Paths.BrowserProfile)
	}

	caps.AddChrome(chromeCaps)

	driver, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		service.Stop()
		return nil, fmt.Errorf("failed to connect to WebDriver: %w", err)
	}

	return &SeleniumBrowser{
		Service: service,
		Driver:  driver,
	}, nil
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
