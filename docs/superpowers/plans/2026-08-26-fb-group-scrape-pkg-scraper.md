# `fb-group-scrape` Reusable `pkg/scraper` Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extract and enhance the core Selenium web scraping functionality and Chrome profile persistence into `/Users/ananda/gitrepo/fb-group-scrape/pkg/scraper` as a reusable, safe Go package.

**Architecture:** A unified `pkg/scraper` package with functional configuration options, automatic ChromeDriver discovery, dynamic TCP port binding, Chrome profile preparation with lockfile cleanup, stealth anti-automation flags, CDP command support, and context-bounded DOM interaction / scrolling utilities.

**Tech Stack:** Go 1.24+, `github.com/tebeka/selenium v0.9.9`, standard library (`net`, `net/http`, `os`, `path/filepath`, `context`, `time`, `sync`).

## Global Constraints

- Package path: `/Users/ananda/gitrepo/fb-group-scrape/pkg/scraper`
- Go module: `github.com/anandasatriaadi/fb-group-scrape`
- Safe Golang principles: maximum 70 lines per function, bounded loops with `ctx.Done()`, error wrapping with `%w`, zero unhandled errors.
- Chrome profile fix: automatic absolute path resolution, directory creation with `0755` permissions, and cleanup of stale singleton locks.

---

### Task 1: Initialize Module & Scaffolding in `fb-group-scrape`

**Files:**
- Create: `/Users/ananda/gitrepo/fb-group-scrape/go.mod`
- Create: `/Users/ananda/gitrepo/fb-group-scrape/.gitignore`

**Interfaces:**
- Produces: Base Go module `github.com/anandasatriaadi/fb-group-scrape` with `github.com/tebeka/selenium v0.9.9` dependency.

- [ ] **Step 1: Initialize git and go.mod in `/Users/ananda/gitrepo/fb-group-scrape`**

```bash
cd /Users/ananda/gitrepo/fb-group-scrape
git init
go mod init github.com/anandasatriaadi/fb-group-scrape
go get github.com/tebeka/selenium@v0.9.9
```

- [ ] **Step 2: Create `.gitignore`**

```bash
cat << 'EOF' > /Users/ananda/gitrepo/fb-group-scrape/.gitignore
# Binaries
bin/
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test artifacts
*.out
*.test
coverage.txt

# Temp & Profile folders
tmp/
saham/
.cache/
profiles/

# OS artifacts
.DS_Store
.vscode/
.idea/
EOF
```

- [ ] **Step 3: Verify module initialization**

Run: `cd /Users/ananda/gitrepo/fb-group-scrape && go mod tidy`
Expected: PASS with `go.mod` and `go.sum` generated.

- [ ] **Step 4: Commit initial repo setup**

```bash
cd /Users/ananda/gitrepo/fb-group-scrape
git add .gitignore go.mod go.sum
git commit -m "chore: initialize repository and go module"
```

---

### Task 2: Implement Configuration & Functional Options (`config.go`)

**Files:**
- Create: `/Users/ananda/gitrepo/fb-group-scrape/pkg/scraper/config.go`
- Test: `/Users/ananda/gitrepo/fb-group-scrape/pkg/scraper/config_test.go`

**Interfaces:**
- Produces: `Config` struct, `DefaultConfig() Config`, `Option` type, and functional options (`WithProfile`, `WithChromeDriver`, `WithHeadless`, `WithDownloadDir`, `WithUserAgent`, `WithPort`, `WithDebugPort`, `WithTimeouts`).

- [ ] **Step 1: Write test for Configuration & Functional Options**

Create `/Users/ananda/gitrepo/fb-group-scrape/pkg/scraper/config_test.go`:
```go
package scraper

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Port != 0 {
		t.Errorf("expected default Port to be 0 (auto), got %d", cfg.Port)
	}
	if cfg.DebugPort != 9222 {
		t.Errorf("expected default DebugPort to be 9222, got %d", cfg.DebugPort)
	}
	if cfg.Headless {
		t.Errorf("expected default Headless to be false, got true")
	}
	if cfg.UserAgent == "" {
		t.Errorf("expected default UserAgent to be non-empty")
	}
	if cfg.PageLoadTimeout != 30*time.Second {
		t.Errorf("expected default PageLoadTimeout to be 30s, got %v", cfg.PageLoadTimeout)
	}
}

func TestFunctionalOptions(t *testing.T) {
	cfg := DefaultConfig()
	opts := []Option{
		WithProfile("/custom/profile"),
		WithChromeDriver("/usr/bin/chromedriver"),
		WithHeadless(true),
		WithDownloadDir("/tmp/downloads"),
		WithUserAgent("CustomAgent/1.0"),
		WithPort(9515),
		WithDebugPort(9333),
		WithTimeouts(45*time.Second, 15*time.Second),
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.ChromeProfilePath != "/custom/profile" {
		t.Errorf("expected ChromeProfilePath '/custom/profile', got %s", cfg.ChromeProfilePath)
	}
	if cfg.ChromeDriverPath != "/usr/bin/chromedriver" {
		t.Errorf("expected ChromeDriverPath '/usr/bin/chromedriver', got %s", cfg.ChromeDriverPath)
	}
	if !cfg.Headless {
		t.Errorf("expected Headless true, got false")
	}
	if cfg.DownloadDir != "/tmp/downloads" {
		t.Errorf("expected DownloadDir '/tmp/downloads', got %s", cfg.DownloadDir)
	}
	if cfg.UserAgent != "CustomAgent/1.0" {
		t.Errorf("expected UserAgent 'CustomAgent/1.0', got %s", cfg.UserAgent)
	}
	if cfg.Port != 9515 {
		t.Errorf("expected Port 9515, got %d", cfg.Port)
	}
	if cfg.DebugPort != 9333 {
		t.Errorf("expected DebugPort 9333, got %d", cfg.DebugPort)
	}
	if cfg.PageLoadTimeout != 45*time.Second {
		t.Errorf("expected PageLoadTimeout 45s, got %v", cfg.PageLoadTimeout)
	}
	if cfg.ImplicitWaitTimeout != 15*time.Second {
		t.Errorf("expected ImplicitWaitTimeout 15s, got %v", cfg.ImplicitWaitTimeout)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/ananda/gitrepo/fb-group-scrape && go test ./pkg/scraper -v`
Expected: FAIL (undefined: DefaultConfig, Option, etc.)

- [ ] **Step 3: Implement `pkg/scraper/config.go`**

Create `/Users/ananda/gitrepo/fb-group-scrape/pkg/scraper/config.go`:
```go
package scraper

import "time"

const (
	DefaultUserAgent           = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36"
	DefaultDebugPort           = 9222
	DefaultPageLoadTimeout     = 30 * time.Second
	DefaultImplicitWaitTimeout = 10 * time.Second
)

// Config holds runtime configuration for the Selenium ChromeDriver instance.
type Config struct {
	ChromeDriverPath    string
	ChromeProfilePath   string
	DownloadDir         string
	UserAgent           string
	Headless            bool
	Port                int
	DebugPort           int
	PageLoadTimeout     time.Duration
	ImplicitWaitTimeout time.Duration
}

// Option modifies a Config struct.
type Option func(*Config)

// DefaultConfig returns default browser configuration.
func DefaultConfig() Config {
	return Config{
		UserAgent:           DefaultUserAgent,
		DebugPort:           DefaultDebugPort,
		PageLoadTimeout:     DefaultPageLoadTimeout,
		ImplicitWaitTimeout: DefaultImplicitWaitTimeout,
	}
}

// WithProfile sets the persistent Chrome user-data-dir.
func WithProfile(path string) Option {
	return func(c *Config) {
		c.ChromeProfilePath = path
	}
}

// WithChromeDriver sets the path to the chromedriver binary.
func WithChromeDriver(path string) Option {
	return func(c *Config) {
		c.ChromeDriverPath = path
	}
}

// WithHeadless enables or disables headless browser execution.
func WithHeadless(headless bool) Option {
	return func(c *Config) {
		c.Headless = headless
	}
}

// WithDownloadDir sets the default download folder.
func WithDownloadDir(path string) Option {
	return func(c *Config) {
		c.DownloadDir = path
	}
}

// WithUserAgent sets a custom User-Agent string.
func WithUserAgent(ua string) Option {
	return func(c *Config) {
		c.UserAgent = ua
	}
}

// WithPort sets an explicit port for ChromeDriver (0 enables auto-allocation).
func WithPort(port int) Option {
	return func(c *Config) {
		c.Port = port
	}
}

// WithDebugPort sets the remote debugging port.
func WithDebugPort(port int) Option {
	return func(c *Config) {
		c.DebugPort = port
	}
}

// WithTimeouts sets page load and implicit element wait timeouts.
func WithTimeouts(pageLoad, implicitWait time.Duration) Option {
	return func(c *Config) {
		c.PageLoadTimeout = pageLoad
		c.ImplicitWaitTimeout = implicitWait
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/ananda/gitrepo/fb-group-scrape && go test ./pkg/scraper -v`
Expected: PASS

- [ ] **Step 5: Commit `config.go`**

```bash
cd /Users/ananda/gitrepo/fb-group-scrape
git add pkg/scraper/config.go pkg/scraper/config_test.go
git commit -m "feat(scraper): add Config and functional options"
```

---

### Task 3: Implement Driver Discovery, Port Allocation, Chrome Profile Fix & Capabilities (`driver.go`)

**Files:**
- Create: `/Users/ananda/gitrepo/fb-group-scrape/pkg/scraper/driver.go`
- Test: `/Users/ananda/gitrepo/fb-group-scrape/pkg/scraper/driver_test.go`

**Interfaces:**
- Produces: `getFreePort() (int, error)`, `findChromeDriver(customPath string) string`, `prepareProfileDir(profilePath string) (string, error)`, `buildChromeCapabilities(cfg Config, absProfile, absDownload string) (selenium.Capabilities, error)`, `sendCDPCommand(port int, sessionID, cmd string, params map[string]any) error`.

- [ ] **Step 1: Write tests for driver discovery, port allocation, and profile cleanup**

Create `/Users/ananda/gitrepo/fb-group-scrape/pkg/scraper/driver_test.go`:
```go
package scraper

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetFreePort(t *testing.T) {
	port, err := getFreePort()
	if err != nil {
		t.Fatalf("getFreePort failed: %v", err)
	}
	if port <= 0 || port > 65535 {
		t.Errorf("invalid port returned: %d", port)
	}
}

func TestFindChromeDriver(t *testing.T) {
	// 1. Explicit path
	explicit := "/tmp/mock-chromedriver"
	if got := findChromeDriver(explicit); got != explicit {
		t.Errorf("expected explicit path %s, got %s", explicit, got)
	}

	// 2. Fallback search
	found := findChromeDriver("")
	if found == "" {
		t.Errorf("expected non-empty chromedriver path")
	}
}

func TestPrepareProfileDir(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "test-scraper-profile")
	defer os.RemoveAll(tmpDir)

	// Create dummy lock files
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("failed to create tmp dir: %v", err)
	}
	lockFile := filepath.Join(tmpDir, "SingletonLock")
	if err := os.WriteFile(lockFile, []byte("dummy"), 0644); err != nil {
		t.Fatalf("failed to create lock file: %v", err)
	}

	absPath, err := prepareProfileDir(tmpDir)
	if err != nil {
		t.Fatalf("prepareProfileDir failed: %v", err)
	}
	if absPath == "" {
		t.Errorf("expected valid absolute path")
	}

	// Ensure lock file was removed
	if _, err := os.Stat(lockFile); !os.IsNotExist(err) {
		t.Errorf("expected SingletonLock to be cleaned up, but it still exists")
	}
}

func TestBuildChromeCapabilities(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Headless = true
	caps, err := buildChromeCapabilities(cfg, "/tmp/profile", "/tmp/download")
	if err != nil {
		t.Fatalf("buildChromeCapabilities failed: %v", err)
	}
	if caps["browserName"] != "chrome" {
		t.Errorf("expected browserName chrome, got %v", caps["browserName"])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/ananda/gitrepo/fb-group-scrape && go test ./pkg/scraper -v`
Expected: FAIL (undefined functions)

- [ ] **Step 3: Implement `pkg/scraper/driver.go`**

Create `/Users/ananda/gitrepo/fb-group-scrape/pkg/scraper/driver.go`:
```go
package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, fmt.Errorf("resolving tcp address: %w", err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, fmt.Errorf("binding tcp port: %w", err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func findChromeDriver(customPath string) string {
	if customPath != "" {
		return customPath
	}
	if path, err := exec.LookPath("chromedriver"); err == nil {
		return path
	}
	standardLocations := []string{
		"/opt/homebrew/bin/chromedriver",
		"/usr/local/bin/chromedriver",
		"/usr/bin/chromedriver",
	}
	for _, loc := range standardLocations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}
	return "chromedriver"
}

func prepareProfileDir(profilePath string) (string, error) {
	if profilePath == "" {
		return "", nil
	}
	absPath, err := filepath.Abs(profilePath)
	if err != nil {
		return "", fmt.Errorf("resolving absolute profile path: %w", err)
	}
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return "", fmt.Errorf("creating profile directory: %w", err)
	}
	cleanStaleLocks(absPath)
	return absPath, nil
}

func cleanStaleLocks(profileDir string) {
	lockFiles := []string{"SingletonLock", "SingletonCookie", "SingletonSocket"}
	for _, f := range lockFiles {
		p := filepath.Join(profileDir, f)
		if _, err := os.Lstat(p); err == nil {
			_ = os.Remove(p)
		}
	}
}

func buildChromeCapabilities(cfg Config, absProfile, absDownload string) (selenium.Capabilities, error) {
	caps := selenium.Capabilities{"browserName": "chrome"}
	args := []string{
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"--disable-gpu",
		"--disable-extensions",
		fmt.Sprintf("--remote-debugging-port=%d", cfg.DebugPort),
		"--log-level=1",
		"--safebrowsing-disable-download-protection",
		"--safebrowsing-disable-extension-blacklist",
		"--user-agent=" + cfg.UserAgent,
	}

	if cfg.Headless {
		args = append(args, "--headless")
	}
	if absProfile != "" {
		args = append(args, "--user-data-dir="+absProfile)
	}

	prefs := map[string]any{
		"safebrowsing.enabled":                                              true,
		"safebrowsing.disable_download_protection":                          true,
		"profile.default_content_settings.popups":                          0,
		"profile.content_settings.exceptions.automatic_downloads.*.setting": 1,
		"credentials_enable_service":                                        false,
		"profile.password_manager_enabled":                                  false,
	}

	if absDownload != "" {
		prefs["download.default_directory"] = absDownload
		prefs["download.prompt_for_download"] = false
		prefs["download.directory_upgrade"] = true
	}

	chromeCaps := chrome.Capabilities{
		Args:            args,
		Prefs:           prefs,
		ExcludeSwitches: []string{"enable-automation"},
	}

	caps.AddChrome(chromeCaps)
	return caps, nil
}

func sendCDPCommand(port int, sessionID, cmd string, params map[string]any) error {
	payload := map[string]any{
		"cmd":    cmd,
		"params": params,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling CDP command: %w", err)
	}
	url := fmt.Sprintf("http://localhost:%d/session/%s/chromium/send_command", port, sessionID)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("sending CDP command: %w", err)
	}
	if resp != nil {
		_ = resp.Body.Close()
	}
	return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/ananda/gitrepo/fb-group-scrape && go test ./pkg/scraper -v`
Expected: PASS

- [ ] **Step 5: Commit `driver.go`**

```bash
cd /Users/ananda/gitrepo/fb-group-scrape
git add pkg/scraper/driver.go pkg/scraper/driver_test.go
git commit -m "feat(scraper): implement driver discovery, profile preparation, and capabilities"
```

---

### Task 4: Implement Core Scraper Struct & Lifecycle (`scraper.go`)

**Files:**
- Create: `/Users/ananda/gitrepo/fb-group-scrape/pkg/scraper/scraper.go`

**Interfaces:**
- Produces: `Scraper` struct, `New(cfg Config, opts ...Option) (*Scraper, error)`, `Close() error`, `Driver() selenium.WebDriver`, `Service() *selenium.Service`, `Port() int`, `Config() Config`, `SendCDPCommand(ctx context.Context, cmd string, params map[string]any) error`.

- [ ] **Step 1: Implement `pkg/scraper/scraper.go`**

Create `/Users/ananda/gitrepo/fb-group-scrape/pkg/scraper/scraper.go`:
```go
package scraper

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/tebeka/selenium"
)

// Scraper manages the lifecycle and high-level operations of a Selenium Chrome browser.
type Scraper struct {
	service *selenium.Service
	driver  selenium.WebDriver
	port    int
	cfg     Config
}

// New creates and initializes a new Scraper instance.
func New(cfg Config, opts ...Option) (*Scraper, error) {
	for _, opt := range opts {
		opt(&cfg)
	}

	port := cfg.Port
	if port == 0 {
		freePort, err := getFreePort()
		if err != nil {
			return nil, fmt.Errorf("getting free port: %w", err)
		}
		port = freePort
	}

	chromeDriverPath := findChromeDriver(cfg.ChromeDriverPath)
	service, err := selenium.NewChromeDriverService(chromeDriverPath, port)
	if err != nil {
		return nil, fmt.Errorf("starting ChromeDriver service (%s on port %d): %w", chromeDriverPath, port, err)
	}

	absProfile, err := prepareProfileDir(cfg.ChromeProfilePath)
	if err != nil {
		_ = service.Stop()
		return nil, fmt.Errorf("preparing chrome profile: %w", err)
	}

	var absDownload string
	if cfg.DownloadDir != "" {
		if abs, err := filepath.Abs(cfg.DownloadDir); err == nil {
			absDownload = abs
		} else {
			absDownload = cfg.DownloadDir
		}
	}

	caps, err := buildChromeCapabilities(cfg, absProfile, absDownload)
	if err != nil {
		_ = service.Stop()
		return nil, fmt.Errorf("building chrome capabilities: %w", err)
	}

	driver, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		_ = service.Stop()
		return nil, fmt.Errorf("connecting to WebDriver remote: %w", err)
	}

	if cfg.PageLoadTimeout > 0 {
		_ = driver.SetPageLoadTimeout(cfg.PageLoadTimeout)
	}
	if cfg.ImplicitWaitTimeout > 0 {
		_ = driver.SetImplicitWaitTimeout(cfg.ImplicitWaitTimeout)
	}

	s := &Scraper{
		service: service,
		driver:  driver,
		port:    port,
		cfg:     cfg,
	}

	if absDownload != "" {
		_ = s.enableCDPDownloads(absDownload)
	}

	return s, nil
}

func (s *Scraper) enableCDPDownloads(downloadDir string) error {
	commands := []string{"Page.setDownloadBehavior", "Browser.setDownloadBehavior"}
	params := map[string]any{
		"behavior":      "allow",
		"downloadPath":  downloadDir,
		"eventsEnabled": true,
	}
	var errs []error
	for _, cmd := range commands {
		if err := sendCDPCommand(s.port, s.driver.SessionID(), cmd, params); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Close gracefully closes the WebDriver session and terminates the ChromeDriver service.
func (s *Scraper) Close() error {
	var errs []error
	if s.driver != nil {
		if err := s.driver.Quit(); err != nil {
			errs = append(errs, fmt.Errorf("closing webdriver: %w", err))
		}
	}
	if s.service != nil {
		if err := s.service.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("stopping service: %w", err))
		}
	}
	return errors.Join(errs...)
}

// Driver returns the underlying Selenium WebDriver instance.
func (s *Scraper) Driver() selenium.WebDriver {
	return s.driver
}

// Service returns the underlying ChromeDriver service.
func (s *Scraper) Service() *selenium.Service {
	return s.service
}

// Port returns the ChromeDriver TCP port.
func (s *Scraper) Port() int {
	return s.port
}

// Config returns the active Scraper configuration.
func (s *Scraper) Config() Config {
	return s.cfg
}

// SendCDPCommand dispatches a custom Chrome DevTools Protocol command.
func (s *Scraper) SendCDPCommand(ctx context.Context, cmd string, params map[string]any) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return sendCDPCommand(s.port, s.driver.SessionID(), cmd, params)
}
```

- [ ] **Step 2: Run test suite to verify compilation and tests pass**

Run: `cd /Users/ananda/gitrepo/fb-group-scrape && go test ./pkg/scraper -v`
Expected: PASS

- [ ] **Step 3: Commit `scraper.go`**

```bash
cd /Users/ananda/gitrepo/fb-group-scrape
git add pkg/scraper/scraper.go
git commit -m "feat(scraper): implement Scraper struct, lifecycle, and CDP command dispatcher"
```

---

### Task 5: Implement High-Level Scraping & Navigation Actions (`actions.go`)

**Files:**
- Create: `/Users/ananda/gitrepo/fb-group-scrape/pkg/scraper/actions.go`

**Interfaces:**
- Produces: `Navigate`, `GetPageSource`, `GetCurrentURL`, `GetTitle`, `WaitForElement`, `WaitForElements`, `Click`, `SendKeys`, `GetText`, `ScrollToBottom`, `ScrollBy`, `ExecuteScript`, `TakeScreenshot`, `GetCookies`, `AddCookie`.

- [ ] **Step 1: Implement `pkg/scraper/actions.go`**

Create `/Users/ananda/gitrepo/fb-group-scrape/pkg/scraper/actions.go`:
```go
package scraper

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tebeka/selenium"
)

// Navigate opens the target URL.
func (s *Scraper) Navigate(ctx context.Context, url string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := s.driver.Get(url); err != nil {
		return fmt.Errorf("navigating to %s: %w", url, err)
	}
	return nil
}

// GetPageSource retrieves the current DOM HTML.
func (s *Scraper) GetPageSource(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	source, err := s.driver.PageSource()
	if err != nil {
		return "", fmt.Errorf("retrieving page source: %w", err)
	}
	return source, nil
}

// GetCurrentURL returns the active browser URL.
func (s *Scraper) GetCurrentURL(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	u, err := s.driver.CurrentURL()
	if err != nil {
		return "", fmt.Errorf("getting current url: %w", err)
	}
	return u, nil
}

// GetTitle returns the current page title.
func (s *Scraper) GetTitle(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	title, err := s.driver.Title()
	if err != nil {
		return "", fmt.Errorf("getting title: %w", err)
	}
	return title, nil
}

// WaitForElement polls until an element matching the selector is found or timeout expires.
func (s *Scraper) WaitForElement(ctx context.Context, by, selector string, timeout time.Duration) (selenium.WebElement, error) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			elem, err := s.driver.FindElement(by, selector)
			if err == nil && elem != nil {
				return elem, nil
			}
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("timeout waiting for element %s (%s)", selector, by)
			}
		}
	}
}

// WaitForElements polls until at least one matching element is found or timeout expires.
func (s *Scraper) WaitForElements(ctx context.Context, by, selector string, timeout time.Duration) ([]selenium.WebElement, error) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			elems, err := s.driver.FindElements(by, selector)
			if err == nil && len(elems) > 0 {
				return elems, nil
			}
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("timeout waiting for elements %s (%s)", selector, by)
			}
		}
	}
}

// Click waits for an element and executes a click on it.
func (s *Scraper) Click(ctx context.Context, by, selector string, timeout time.Duration) error {
	elem, err := s.WaitForElement(ctx, by, selector, timeout)
	if err != nil {
		return fmt.Errorf("finding element to click: %w", err)
	}
	if err := elem.Click(); err != nil {
		return fmt.Errorf("clicking element %s: %w", selector, err)
	}
	return nil
}

// SendKeys waits for an input element and types the specified text into it.
func (s *Scraper) SendKeys(ctx context.Context, by, selector, text string, timeout time.Duration) error {
	elem, err := s.WaitForElement(ctx, by, selector, timeout)
	if err != nil {
		return fmt.Errorf("finding element to send keys: %w", err)
	}
	if err := elem.SendKeys(text); err != nil {
		return fmt.Errorf("sending keys to %s: %w", selector, err)
	}
	return nil
}

// GetText waits for an element and extracts its inner text.
func (s *Scraper) GetText(ctx context.Context, by, selector string, timeout time.Duration) (string, error) {
	elem, err := s.WaitForElement(ctx, by, selector, timeout)
	if err != nil {
		return "", fmt.Errorf("finding element for text: %w", err)
	}
	text, err := elem.Text()
	if err != nil {
		return "", fmt.Errorf("getting text from %s: %w", selector, err)
	}
	return text, nil
}

// ScrollToBottom repeatedly scrolls down the window to trigger dynamic loading (e.g. Facebook feed).
func (s *Scraper) ScrollToBottom(ctx context.Context, pause time.Duration, maxScrolls int) error {
	for i := 0; i < maxScrolls; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		_, err := s.driver.ExecuteScript("window.scrollTo(0, document.body.scrollHeight);", nil)
		if err != nil {
			return fmt.Errorf("executing scroll script on iteration %d: %w", i, err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pause):
		}
	}
	return nil
}

// ScrollBy scrolls the browser viewport by the specified delta x and y.
func (s *Scraper) ScrollBy(ctx context.Context, x, y int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	script := fmt.Sprintf("window.scrollBy(%d, %d);", x, y)
	if _, err := s.driver.ExecuteScript(script, nil); err != nil {
		return fmt.Errorf("executing scrollBy(%d, %d): %w", x, y, err)
	}
	return nil
}

// ExecuteScript executes arbitrary JavaScript code in the browser context.
func (s *Scraper) ExecuteScript(ctx context.Context, script string, args ...any) (any, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	res, err := s.driver.ExecuteScript(script, args)
	if err != nil {
		return nil, fmt.Errorf("executing script: %w", err)
	}
	return res, nil
}

// TakeScreenshot captures the active page and writes PNG data to outputPath.
func (s *Scraper) TakeScreenshot(ctx context.Context, outputPath string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	data, err := s.driver.Screenshot()
	if err != nil {
		return fmt.Errorf("capturing screenshot: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("creating screenshot dir: %w", err)
	}
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("saving screenshot file: %w", err)
	}
	return nil
}

// GetCookies returns the browser session cookies.
func (s *Scraper) GetCookies(ctx context.Context) ([]selenium.Cookie, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	cookies, err := s.driver.GetCookies()
	if err != nil {
		return nil, fmt.Errorf("getting cookies: %w", err)
	}
	return cookies, nil
}

// AddCookie sets a cookie in the active browser session.
func (s *Scraper) AddCookie(ctx context.Context, cookie *selenium.Cookie) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if cookie == nil {
		return fmt.Errorf("cookie cannot be nil")
	}
	if err := s.driver.AddCookie(cookie); err != nil {
		return fmt.Errorf("adding cookie: %w", err)
	}
	return nil
}
```

- [ ] **Step 2: Run test suite to verify everything compiles cleanly**

Run: `cd /Users/ananda/gitrepo/fb-group-scrape && go test ./pkg/scraper -v`
Expected: PASS

- [ ] **Step 3: Commit `actions.go`**

```bash
cd /Users/ananda/gitrepo/fb-group-scrape
git add pkg/scraper/actions.go
git commit -m "feat(scraper): implement high-level actions, scrolling, and cookie management"
```

---

### Task 6: Add Example Runner & End-to-End Verification

**Files:**
- Create: `/Users/ananda/gitrepo/fb-group-scrape/cmd/example/main.go`
- Create: `/Users/ananda/gitrepo/fb-group-scrape/README.md`

**Interfaces:**
- Produces: Working CLI example verifying package usage and README documentation.

- [ ] **Step 1: Create `cmd/example/main.go`**

Create `/Users/ananda/gitrepo/fb-group-scrape/cmd/example/main.go`:
```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anandasatriaadi/fb-group-scrape/pkg/scraper"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received termination signal, shutting down...")
		cancel()
	}()

	cfg := scraper.DefaultConfig()
	cfg.Headless = true
	cfg.ChromeProfilePath = "./tmp/chrome-profile"

	s, err := scraper.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize scraper: %v", err)
	}
	defer func() {
		if err := s.Close(); err != nil {
			log.Printf("Error closing scraper: %v", err)
		}
	}()

	log.Printf("Scraper started on port %d with profile %s", s.Port(), cfg.ChromeProfilePath)

	if err := s.Navigate(ctx, "https://example.com"); err != nil {
		log.Fatalf("Failed to navigate: %v", err)
	}

	title, err := s.GetTitle(ctx)
	if err != nil {
		log.Fatalf("Failed to get title: %v", err)
	}
	fmt.Printf("Page Title: %s\n", title)

	h1Text, err := s.GetText(ctx, "tag name", "h1", 5*time.Second)
	if err != nil {
		log.Printf("Could not get h1: %v", err)
	} else {
		fmt.Printf("H1 Header: %s\n", h1Text)
	}
}
```

- [ ] **Step 2: Create `README.md` in `fb-group-scrape`**

Create `/Users/ananda/gitrepo/fb-group-scrape/README.md`:
```markdown
# fb-group-scrape

High-performance Facebook web scraper and data collection engine with persistent Chrome profiles and automated session recovery.

## Package: `pkg/scraper`

The `pkg/scraper` package provides a standalone, production-ready Selenium Chrome automation engine:
- **Persistent Chrome Profiles**: `--user-data-dir` normalization and automated stale singleton lock cleanup.
- **Anti-Automation Capabilities**: Excludes `enable-automation` switch, suppresses test banners and password popups.
- **Dynamic Port Allocation**: Ephemeral TCP port binding to prevent port conflicts.
- **ChromeDriver Discovery**: Auto-detects `chromedriver` binaries across standard macOS/Linux paths.
- **High-Level Actions**: Context-aware helpers for navigation, polling, clicking, typing, scrolling, screenshots, cookies, and Chrome DevTools Protocol (CDP) commands.

## Quick Start

```go
package main

import (
    "context"
    "time"
    "github.com/anandasatriaadi/fb-group-scrape/pkg/scraper"
)

func main() {
    ctx := context.Background()

    // Initialize scraper with persistent profile
    s, err := scraper.New(
        scraper.DefaultConfig(),
        scraper.WithProfile("./profiles/fb-session"),
        scraper.WithHeadless(false),
    )
    if err != nil {
        panic(err)
    }
    defer s.Close()

    // Navigate to Facebook Group
    _ = s.Navigate(ctx, "https://www.facebook.com/groups/...")

    // Scroll dynamic feed
    _ = s.ScrollToBottom(ctx, 2*time.Second, 10)
}
```

## Running Verification

```bash
go test -v ./...
go vet ./...
```
```

- [ ] **Step 3: Run full verification suite**

Run: `cd /Users/ananda/gitrepo/fb-group-scrape && go vet ./... && go test -v ./...`
Expected: PASS with 0 warnings, 0 errors.

- [ ] **Step 4: Commit example and documentation**

```bash
cd /Users/ananda/gitrepo/fb-group-scrape
git add cmd/example/main.go README.md
git commit -m "feat: add example runner and documentation"
```

---
