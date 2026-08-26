# Design Document: Reusable Selenium & Chrome Profile Scraper Package for `fb-group-scrape`

**Date:** 2026-08-26  
**Target Repository:** `/Users/ananda/gitrepo/fb-group-scrape`  
**Go Module:** `github.com/anandasatriaadi/fb-group-scrape`  
**Package:** `pkg/scraper`

---

## 1. Overview & Objectives

Extract and enhance the core Selenium web scraping engine and Chrome profile persistence mechanism from `go-idx-scraper` into a standalone, reusable Go package located at `/Users/ananda/gitrepo/fb-group-scrape/pkg/scraper`.

### Key Requirements:
1. **Chrome Profile Fix & Persistence**: Correct `--user-data-dir` configuration with absolute path resolution, directory auto-creation, and stale singleton lock cleanup to prevent crash loops and preserve logged-in sessions (e.g. Facebook credentials).
2. **Anti-Automation & Stealth Flags**: Automatic removal of `enable-automation` switches, disabling automation banners, suppressing password manager popups, configuring modern desktop user-agents, and enabling download behaviors via Chrome DevTools Protocol (CDP).
3. **Dynamic Port & Driver Discovery**: Auto-detection of `chromedriver` binaries across standard macOS/Linux paths (`/opt/homebrew/bin`, `/usr/local/bin`, `/usr/bin`, `$PATH`) and dynamic ephemeral TCP port assignment.
4. **Scraping Actions with Context**: High-level scraping utilities (`Navigate`, `WaitForElement`, `WaitForElements`, `Click`, `SendKeys`, `ScrollToBottom`, `ScrollBy`, `GetPageSource`, `ExecuteScript`, `TakeScreenshot`, `GetCookies`, `AddCookie`, `SendCDPCommand`) with strict `context.Context` cancellation support.
5. **Safe Golang Compliance**: Maximum 70 lines per function, zero unhandled errors, bounded loops and timeouts, zero third-party dependencies outside `github.com/tebeka/selenium`.

---

## 2. Directory Layout & Module Structure

```
/Users/ananda/gitrepo/fb-group-scrape/
├── go.mod                     # Module: github.com/anandasatriaadi/fb-group-scrape (Go 1.24+)
├── go.sum
└── pkg/
    └── scraper/
        ├── config.go          # Configuration struct & functional options
        ├── driver.go          # ChromeDriver discovery, port allocation, capabilities & CDP setup
        ├── scraper.go         # Core Scraper struct, lifecycle (New, Close, Driver accessor)
        ├── actions.go         # High-level DOM interaction, navigation & scrolling methods
        └── scraper_test.go    # Unit tests for discovery, config options, and port allocation
```

---

## 3. Configuration & Options (`config.go`)

### Data Structures

```go
package scraper

import "time"

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

type Option func(*Config)
```

### Functional Options
- `WithProfile(path string) Option`: Sets user data directory and normalizes to absolute path.
- `WithChromeDriver(path string) Option`: Specifies explicit `chromedriver` binary path.
- `WithHeadless(headless bool) Option`: Toggles headless mode (default: false for interactive or true for automation).
- `WithDownloadDir(path string) Option`: Sets download destination and configures CDP download behavior.
- `WithUserAgent(ua string) Option`: Overrides default desktop User-Agent string.
- `WithPort(port int) Option`: Specifies fixed ChromeDriver port (0 = auto-assign ephemeral port).
- `WithDebugPort(port int) Option`: Configures remote debugging port.
- `WithTimeouts(pageLoad, implicitWait time.Duration) Option`: Sets default WebDriver timeouts.

---

## 4. Driver Initialization & Profile Fix (`driver.go`)

### 1. Ephemeral Port Allocation
- `getFreePort() (int, error)`: Binds to TCP `:0` on `localhost` and reads the assigned port to avoid conflicts.

### 2. ChromeDriver Discovery
- `findChromeDriver(customPath string) string`: Searches `customPath`, `/opt/homebrew/bin/chromedriver`, `/usr/local/bin/chromedriver`, `/usr/bin/chromedriver`, and `exec.LookPath("chromedriver")`.

### 3. Chrome Profile Fix & Lock Cleanup
- `prepareProfileDir(profilePath string) (string, error)`:
  1. Resolves path to absolute path using `filepath.Abs`.
  2. Ensures directory exists (`os.MkdirAll(absPath, 0755)`).
  3. Removes stale lock files (`SingletonLock`, `SingletonCookie`, `SingletonSocket`) if Chrome process previously terminated abnormally.

### 4. Capabilities & Anti-Automation Switches
- **Arguments**:
  - `--no-sandbox`
  - `--disable-dev-shm-usage`
  - `--disable-gpu`
  - `--disable-extensions`
  - `--remote-debugging-port=<debugPort>`
  - `--log-level=1`
  - `--safebrowsing-disable-download-protection`
  - `--safebrowsing-disable-extension-blacklist`
  - `--user-agent=<userAgent>`
  - `--headless` (when `Headless == true`)
  - `--user-data-dir=<profilePath>` (when `ChromeProfilePath != ""`)
- **Exclude Switches**: `["enable-automation"]`
- **Preferences**:
  - `credentials_enable_service: false`
  - `profile.password_manager_enabled: false`
  - `profile.default_content_settings.popups: 0`
  - `profile.content_settings.exceptions.automatic_downloads.*.setting: 1`
  - `safebrowsing.enabled: true`
  - `safebrowsing.disable_download_protection: true`
  - `download.default_directory: <absDownloadDir>`
  - `download.prompt_for_download: false`
  - `download.directory_upgrade: true`

### 5. CDP Command Sender
- `sendCDPCommand(port int, sessionID string, cmd string, params map[string]any) error`: Dispatches HTTP POST to `http://localhost:<port>/session/<session_id>/chromium/send_command`.

---

## 5. Scraper Lifecycle & Actions (`scraper.go`, `actions.go`)

### Struct Definition
```go
type Scraper struct {
	service *selenium.Service
	driver  selenium.WebDriver
	port    int
	cfg     Config
}
```

### Methods
- `New(cfg Config, opts ...Option) (*Scraper, error)`: Orchestrates port discovery, driver service start, capabilities setup, remote session creation, timeout configuration, and CDP initialization.
- `Close() error`: Gracefully closes WebDriver (`driver.Quit()`) and stops ChromeDriver service (`service.Stop()`), joining any errors.
- `Driver() selenium.WebDriver`: Returns underlying Selenium WebDriver.
- `Service() *selenium.Service`: Returns underlying ChromeDriver service.
- `Port() int`: Returns active ChromeDriver port.
- `Config() Config`: Returns active configuration.

### Action Methods
1. `Navigate(ctx context.Context, url string) error`
2. `GetPageSource(ctx context.Context) (string, error)`
3. `GetCurrentURL(ctx context.Context) (string, error)`
4. `GetTitle(ctx context.Context) (string, error)`
5. `WaitForElement(ctx context.Context, by, selector string, timeout time.Duration) (selenium.WebElement, error)`
6. `WaitForElements(ctx context.Context, by, selector string, timeout time.Duration) ([]selenium.WebElement, error)`
7. `Click(ctx context.Context, by, selector string, timeout time.Duration) error`
8. `SendKeys(ctx context.Context, by, selector, text string, timeout time.Duration) error`
9. `GetText(ctx context.Context, by, selector string, timeout time.Duration) (string, error)`
10. `ScrollToBottom(ctx context.Context, pause time.Duration, maxScrolls int) error`
11. `ScrollBy(ctx context.Context, x, y int) error`
12. `ExecuteScript(ctx context.Context, script string, args ...any) (any, error)`
13. `TakeScreenshot(ctx context.Context, outputPath string) error`
14. `GetCookies(ctx context.Context) ([]selenium.Cookie, error)`
15. `AddCookie(ctx context.Context, cookie *selenium.Cookie) error`
16. `SendCDPCommand(ctx context.Context, cmd string, params map[string]any) error`

---

## 6. Testing & Verification Plan

1. **Unit Tests (`pkg/scraper/scraper_test.go`)**:
   - `TestDefaultConfig`: Verify default configuration values and functional options mutation.
   - `TestGetFreePort`: Verify dynamic TCP port allocation works and releases port.
   - `TestFindChromeDriver`: Verify auto-discovery logic finds valid path or returns sensible fallback.
   - `TestPrepareProfileDir`: Verify directory creation, absolute path resolution, and lock file cleanup.
2. **Integration Verification (`go test -v ./...` & `go vet ./...`)**:
   - Run in `/Users/ananda/gitrepo/fb-group-scrape`.
   - Verify zero compilation errors, zero vet warnings, and test coverage.
