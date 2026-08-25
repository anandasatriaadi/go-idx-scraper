# System Design: `chromedriver-mcp` (Go Headless Browser MCP Server)

## 1. Overview
`chromedriver-mcp` is a lightweight, standalone Model Context Protocol (MCP) server written in Go. It provides AI agents (such as Pi, Claude Desktop, Cursor, and custom agent harnesses) with direct browser automation, web scraping, and Chrome DevTools Protocol (CDP) download capabilities via standard JSON-RPC 2.0 over `stdio`.

**Target Location**: `/Users/ananda/gitrepo/chromedriver-mcp`

---

## 2. Core Architectural Principles
1. **Direct Chrome DevTools Protocol (CDP)**: Controls Chrome/Chromium via WebSocket using `github.com/chromedp/chromedp`, eliminating the need for external `chromedriver` binaries or port collisions.
2. **Session Lifecycle & Resource Management**: Reuses a single headless browser instance across multi-step tool calls (e.g. navigate -> fill form -> click -> download), while strictly emphasizing `browser_close` / process teardown to prevent zombie Chrome processes or locked ports.
3. **CDP Native Downloads**: Directly configures `Browser.setDownloadBehavior` / `Page.setDownloadBehavior` to enable automated downloads to custom directories without UI popups.
4. **Standard JSON-RPC 2.0 / MCP Compliance**: Implements `initialize`, `tools/list`, and `tools/call` methods over `stdio` according to the Model Context Protocol specification.

---

## 3. Repository Structure

```
/Users/ananda/gitrepo/chromedriver-mcp/
├── cmd/
│   └── chromedriver-mcp/
│       └── main.go              # CLI entry point (MCP stdio loop)
├── internal/
│   ├── browser/
│   │   ├── browser.go           # Browser lifecycle & Chrome process supervisor
│   │   ├── actions.go           # Navigation, clicking, typing, screenshots, JS eval
│   │   └── download.go          # CDP download handling and file completion polling
│   └── mcp/
│       ├── server.go            # JSON-RPC 2.0 stdio server loop
│       ├── protocol.go          # MCP standard request/response types
│       └── tools.go             # Tool definitions, schemas, and dispatchers
├── go.mod
├── go.sum
├── Makefile                     # Build, test, and binary installation
└── README.md                    # Integration guide for Pi, Claude Desktop, Cursor
```

---

## 4. MCP Tools Specification

### 1. `browser_navigate`
- **Description**: Navigates to the specified URL in the headless browser and waits for the page to load.
- **Parameters**:
  - `url` *(string, required)*: The HTTP/HTTPS URL to open.
  - `wait_until` *(string, optional, default: `"network_idle"`)*: `"dom_content_loaded"` or `"network_idle"`.
  - `timeout_ms` *(integer, optional, default: `30000`)*: Maximum navigation timeout in milliseconds.
- **Output**: JSON containing `status`, `title`, `url`, and HTTP status code.

### 2. `browser_get_content`
- **Description**: Extracts visible text, structured markdown, or raw HTML from the current page or a scoped CSS selector.
- **Parameters**:
  - `selector` *(string, optional, default: `"body"`)*: CSS selector to scope extraction.
  - `format` *(string, optional, default: `"text"`)*: `"text"`, `"markdown"`, or `"html"`.
- **Output**: String containing page content formatted according to requested type.

### 3. `browser_click`
- **Description**: Clicks on an element identified by CSS selector or XPath.
- **Parameters**:
  - `selector` *(string, required)*: Target CSS selector or XPath.
  - `wait_timeout_ms` *(integer, optional, default: `5000`)*: Time to wait for element visibility.
- **Output**: JSON confirmation with final URL if navigation occurred.

### 4. `browser_type`
- **Description**: Types text into an input field or textarea.
- **Parameters**:
  - `selector` *(string, required)*: Target input element selector.
  - `text` *(string, required)*: Text string to input.
  - `clear` *(boolean, optional, default: `true`)*: Whether to clear existing content before typing.
  - `press_enter` *(boolean, optional, default: `false`)*: Whether to send the Enter key after typing.
- **Output**: JSON confirmation of typing operation.

### 5. `browser_screenshot`
- **Description**: Captures a screenshot of the current page (viewport or full scrollable height).
- **Parameters**:
  - `full_page` *(boolean, optional, default: `false`)*: Set true for full scrollable page snapshot.
  - `output_path` *(string, optional)*: Local filesystem destination for the PNG file.
- **Output**: JSON containing `output_path` or base64 image data.

### 6. `browser_evaluate`
- **Description**: Executes arbitrary JavaScript expression inside the active page context and returns the evaluated JSON result.
- **Parameters**:
  - `expression` *(string, required)*: JavaScript snippet to evaluate.
- **Output**: Evaluated result formatted as JSON string.

### 7. `browser_download`
- **Description**: Triggers a file download via element click or direct URL navigation, configuring CDP download behavior and awaiting file completion on disk.
- **Parameters**:
  - `download_dir` *(string, required)*: Destination folder path on local filesystem.
  - `click_selector` *(string, optional)*: Element selector to click that initiates download.
  - `url` *(string, optional)*: Direct file download URL if not triggered by click.
  - `timeout_ms` *(integer, optional, default: `60000`)*: Maximum time in ms to wait for download to finish.
- **Output**: JSON containing `downloaded_file_path`, `file_name`, and `file_size_bytes`.

### 8. `browser_close`
- **Description**: **CRITICAL**: Closes the active browser session and terminates the underlying Chrome process. Callers MUST invoke `browser_close` when finishing browser tasks to release ports and prevent subsequent browser launch failures.
- **Parameters**: (none)
- **Output**: JSON confirming browser shutdown and resource release.

---

## 5. Error Handling & Reliability
- **Chrome Binary Auto-Detection**: Searches common paths (`/Applications/Google Chrome.app/Contents/MacOS/Google Chrome`, `/usr/bin/google-chrome`, `/usr/bin/chromium`, `C:\Program Files\Google\Chrome\Application\chrome.exe`) and `$PATH`.
- **Process Supervisor & Signal Handling**: Traps `SIGINT`, `SIGTERM`, and `SIGHUP` to ensure any running Chrome processes are cleanly killed on exit.
- **Timeouts**: Every CDP operation executes within a bounded `context.WithTimeout` to avoid hanging indefinitely on unresponsive web servers.

---

## 6. Testing & Validation Strategy
1. **Unit Tests**: Test JSON-RPC MCP request/response serialization and tool routing.
2. **Integration Tests (`browser_test.go`)**:
   - Spawns local `httptest.Server` serving test HTML files.
   - Tests navigation, clicking, text input, JS evaluation, screenshot generation, and CDP file download to temp directory.
   - Tests `browser_close` ensuring clean shutdown and process termination.
