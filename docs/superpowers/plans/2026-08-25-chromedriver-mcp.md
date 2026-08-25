# `chromedriver-mcp` (Go Headless Browser MCP Server) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a standalone, high-performance Model Context Protocol (MCP) server in Go located at `/Users/ananda/gitrepo/chromedriver-mcp` that exposes headless Chrome DevTools Protocol (CDP) automation, page scraping, and file downloading capabilities over stdio JSON-RPC 2.0 for AI agents.

**Architecture:** 
- `internal/browser/`: Manages headless Chrome lifecycle via `chromedp/chromedp`, handling navigation, DOM interactions, evaluation, full-page screenshots, and CDP downloads (`Browser.setDownloadBehavior`).
- `internal/mcp/`: Implements JSON-RPC 2.0 stdio loop and MCP tool schemas (`browser_navigate`, `browser_get_content`, `browser_click`, `browser_type`, `browser_screenshot`, `browser_evaluate`, `browser_download`, `browser_close`).
- `cmd/chromedriver-mcp/`: CLI binary entry point with OS signal traps (`SIGINT`, `SIGTERM`) for resource cleanup.

**Tech Stack:** Go 1.24+, `github.com/chromedp/chromedp`, `github.com/chromedp/cdproto`, JSON-RPC 2.0 stdio MCP.

## Global Constraints

- Target directory: `/Users/ananda/gitrepo/chromedriver-mcp`.
- Standalone Go module: `github.com/anandasatriaadi/chromedriver-mcp`.
- Transport: Standard JSON-RPC 2.0 over `stdio`.
- Context propagation: Every CDP action must be bounded by explicit timeouts.
- Tool schemas must explicitly emphasize `browser_close` to prevent zombie Chrome processes.
- All tests must pass with `go test -count=1 ./...`.

---

### Task 1: Project Scaffolding & Core Chrome CDP Engine (`internal/browser/`)

**Files:**
- Create: `/Users/ananda/gitrepo/chromedriver-mcp/go.mod`
- Create: `/Users/ananda/gitrepo/chromedriver-mcp/internal/browser/browser.go`
- Create: `/Users/ananda/gitrepo/chromedriver-mcp/internal/browser/actions.go`
- Create: `/Users/ananda/gitrepo/chromedriver-mcp/internal/browser/download.go`
- Create: `/Users/ananda/gitrepo/chromedriver-mcp/internal/browser/browser_test.go`

**Interfaces:**
- `Manager`:
  - `NewManager(headless bool) *Manager`
  - `EnsureBrowser(ctx context.Context) error`
  - `Close() error`
  - `Navigate(ctx context.Context, url string, waitUntil string, timeoutMs int) (*NavResult, error)`
  - `GetContent(ctx context.Context, selector string, format string) (string, error)`
  - `Click(ctx context.Context, selector string, waitTimeoutMs int) (*ClickResult, error)`
  - `Type(ctx context.Context, selector string, text string, clear bool, pressEnter bool) error`
  - `Screenshot(ctx context.Context, fullPage bool, outputPath string) ([]byte, string, error)`
  - `Evaluate(ctx context.Context, expression string) (any, error)`
  - `Download(ctx context.Context, clickSelector string, directURL string, downloadDir string, timeoutMs int) (*DownloadResult, error)`

- [ ] **Step 1: Initialize git repository and Go module at `/Users/ananda/gitrepo/chromedriver-mcp`**
- [ ] **Step 2: Implement `internal/browser/browser.go` with Chrome auto-discovery and context supervisor**
- [ ] **Step 3: Implement `internal/browser/actions.go` (Navigate, GetContent, Click, Type, Screenshot, Evaluate)**
- [ ] **Step 4: Implement `internal/browser/download.go` with CDP download behavior and polling**
- [ ] **Step 5: Write comprehensive integration tests in `internal/browser/browser_test.go` using `httptest.Server`**
- [ ] **Step 6: Run browser tests (`go test -v -count=1 ./internal/browser/...`)**
- [ ] **Step 7: Commit Task 1 in the new repository**

---

### Task 2: JSON-RPC 2.0 & MCP Protocol Engine (`internal/mcp/`)

**Files:**
- Create: `/Users/ananda/gitrepo/chromedriver-mcp/internal/mcp/protocol.go`
- Create: `/Users/ananda/gitrepo/chromedriver-mcp/internal/mcp/tools.go`
- Create: `/Users/ananda/gitrepo/chromedriver-mcp/internal/mcp/server.go`
- Create: `/Users/ananda/gitrepo/chromedriver-mcp/internal/mcp/server_test.go`

**Interfaces:**
- `Server`:
  - `NewServer(browserManager *browser.Manager) *Server`
  - `Serve(in io.Reader, out io.Writer) error`
  - `HandleMessage(req []byte) ([]byte, error)`

- [ ] **Step 1: Define JSON-RPC 2.0 and MCP data structures in `internal/mcp/protocol.go`**
- [ ] **Step 2: Define MCP Tool schemas and parameter validators in `internal/mcp/tools.go`**
- [ ] **Step 3: Implement the stdio server loop in `internal/mcp/server.go` with tool dispatching**
- [ ] **Step 4: Write unit tests in `internal/mcp/server_test.go` verifying initialize, tools/list, and tools/call JSON-RPC methods**
- [ ] **Step 5: Run tests (`go test -v -count=1 ./internal/mcp/...`)**
- [ ] **Step 6: Commit Task 2 in the new repository**

---

### Task 3: CLI Entry Point, Build System & Documentation

**Files:**
- Create: `/Users/ananda/gitrepo/chromedriver-mcp/cmd/chromedriver-mcp/main.go`
- Create: `/Users/ananda/gitrepo/chromedriver-mcp/Makefile`
- Create: `/Users/ananda/gitrepo/chromedriver-mcp/README.md`
- Create: `/Users/ananda/gitrepo/chromedriver-mcp/.gitignore`

**Interfaces:**
- `main()`: Sets up signal trap (`SIGINT`, `SIGTERM`), instantiates `browser.Manager` and `mcp.Server`, runs stdio loop, and cleans up on exit.

- [ ] **Step 1: Implement `cmd/chromedriver-mcp/main.go` with OS signal handling and `--headless` CLI flag**
- [ ] **Step 2: Create `Makefile` with `build`, `test`, `install`, and `clean` targets**
- [ ] **Step 3: Create comprehensive `README.md` with integration configs for Pi, Claude Desktop, and Cursor**
- [ ] **Step 4: Build binary (`make build`) and run end-to-end stdio verification test**
- [ ] **Step 5: Commit Task 3 in the new repository**

---

### Task 4: Verification, Full Build & Final Code Review

- [ ] **Step 1: Run complete Go test suite (`go test -count=1 -v ./...`)**
- [ ] **Step 2: Compile binary to `bin/chromedriver-mcp`**
- [ ] **Step 3: Verify clean git status and tag v1.0.0**
