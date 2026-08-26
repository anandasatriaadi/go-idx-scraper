# Clean DDD Hexagonal Architecture Cleanup Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor the Go codebase into strict Clean DDD Hexagonal architecture by eliminating infrastructure leaks from domain layers, defining explicit domain ports, encapsulating use case orchestration, and cleaning up repository artifacts while preserving 100% MongoDB and `idx-web` compatibility.

**Architecture:** Hexagonal Architecture with DDD Bounded Contexts where domain entities and driven repository ports (`internal/feature/*`) have zero external database/driver imports (`go.mongodb.org`), use cases are encapsulated in application services, MongoDB adapters (`internal/infra/db/mongo/`) implement domain ports, and CLI entry points (`cmd/*`) act purely as driving adapters.

**Tech Stack:** Go 1.24+, MongoDB v2 Driver (`go.mongodb.org/mongo-driver/v2`), zap logger, tebeka/selenium, Nuxt 4 / Nitro server (`idx-web`).

## Global Constraints
- Zero `go.mongodb.org` imports in `internal/feature/*`.
- All entity IDs are standard `string` with `bson:"_id,omitempty" json:"id"`.
- BSON and JSON field tags on domain entities match MongoDB collections and `idx-web` schema exactly.
- All Go tests (`go test ./...`), static analysis (`go vet ./...`), and CLI builds (`make build`) must pass cleanly.
- `idx-web` production build (`npm --prefix idx-web run build`) must compile without errors.

---

### Task 1: Repository Hygiene & Dead Code Cleanup

**Files:**
- Delete: `D:\sync-gitrepo\go-idx-scraper\stock-list-2.json`
- Delete: `internal/infra/external/openrouter/`, `internal/infra/external/firebase/`, `internal/infra/external/email/`, `internal/infra/external/`
- Delete: `tools/mongo_repo/main.go`, `tools/mongo_repo/`
- Modify: `internal/feature/announcement/entity.go:1-15`

**Interfaces:**
- Consumes: N/A
- Produces: Clean workspace free of orphaned artifacts and obsolete generator scripts.

- [ ] **Step 1: Remove untracked file and orphaned directories**

```bash
rm -f "D:\sync-gitrepo\go-idx-scraper\stock-list-2.json"
rm -rf internal/infra/external
rm -rf tools/mongo_repo
rm -f bin/mongo_repo
```

- [ ] **Step 2: Remove `//go:generate` annotation from announcement entity**

In `internal/feature/announcement/entity.go`, remove:
`//go:generate go run ../../../tools/mongo_repo/main.go -type=Announcement -collection=announcements`

- [ ] **Step 3: Run git status and test suite**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 4: Commit cleanup**

```bash
git add -u
git commit -m "chore: remove obsolete mongo_repo tool and empty external stubs"
```

---

### Task 2: Purify Domain Entities & Ports in `system`, `stock`, `finreport`

**Files:**
- Modify: `internal/feature/system/entity.go`
- Modify: `internal/feature/stock/entity.go`
- Modify: `internal/feature/finreport/entity.go`
- Modify: `internal/feature/finreport/service.go`
- Test: `internal/feature/finreport/service_test.go`

**Interfaces:**
- Consumes: Standard Go types (`string`, `time.Time`, `map[string]any`)
- Produces: Pure domain definitions and driven repository interfaces for system, stock, and finreport contexts.

- [ ] **Step 1: Write unit test verifying domain entity purity**

Create or update tests in `internal/feature/finreport/service_test.go` to mock pure repository ports without `options.Lister` or `bson.M`.

```go
package finreport_test

import (
	"context"
	"testing"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/finreport"
)

type mockFinReportRepo struct {
	reports map[string]*finreport.FinancialReport
}

func (m *mockFinReportRepo) Create(ctx context.Context, report *finreport.FinancialReport) error {
	key := report.IssuerCode + "-" + report.PeriodString
	m.reports[key] = report
	return nil
}

func (m *mockFinReportRepo) FindByIssuerAndPeriod(ctx context.Context, issuerCode string, year int, periodString string) (*finreport.FinancialReport, error) {
	key := issuerCode + "-" + periodString
	return m.reports[key], nil
}

func (m *mockFinReportRepo) UpdateIsLatest(ctx context.Context, issuerCode string, year int, periodString string, isLatest bool) error {
	key := issuerCode + "-" + periodString
	if r, ok := m.reports[key]; ok {
		r.IsLatest = isLatest
	}
	return nil
}

func (m *mockFinReportRepo) ListByIssuer(ctx context.Context, issuerCode string, limit int) ([]*finreport.FinancialReport, error) {
	var list []*finreport.FinancialReport
	for _, r := range m.reports {
		if r.IssuerCode == issuerCode {
			list = append(list, r)
		}
	}
	return list, nil
}

func TestParseFinancialStatementFilename(t *testing.T) {
	filename := "FinancialStatement-2024-Tahunan-BBRI.xlsx"
	year, period, issuer, err := finreport.ParseFinancialStatementFilename(filename)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if year != 2024 || period != "Tahunan" || issuer != "BBRI" {
		t.Fatalf("unexpected parse result: year=%d, period=%s, issuer=%s", year, period, issuer)
	}
}
```

- [ ] **Step 2: Purify `internal/feature/system/entity.go`**

```go
package system

import (
	"context"
	"time"
)

type LastRun struct {
	ID         string         `bson:"_id,omitempty" json:"id,omitempty"`
	ScriptName string         `bson:"scriptName" json:"scriptName"`
	LastRunAt  time.Time      `bson:"lastRunAt" json:"lastRunAt"`
	Metadata   map[string]any `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt  time.Time      `bson:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time      `bson:"updatedAt" json:"updatedAt"`
}

type Repository interface {
	GetLastRun(ctx context.Context, scriptName string) (*LastRun, error)
	SaveLastRun(ctx context.Context, lastRun *LastRun) error
}
```

- [ ] **Step 3: Purify `internal/feature/stock/entity.go`**

```go
package stock

import (
	"context"
	"time"
)

type StockData struct {
	Code string `json:"StockCode"`
}

type StockListResponse struct {
	Data []StockData `json:"data"`
}

type PriceCandle struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	Ticker    string    `bson:"ticker" json:"ticker"`
	Date      time.Time `bson:"date" json:"date"`
	Open      float64   `bson:"open" json:"open"`
	High      float64   `bson:"high" json:"high"`
	Low       float64   `bson:"low" json:"low"`
	Close     float64   `bson:"close" json:"close"`
	AdjClose  float64   `bson:"adj_close" json:"adj_close"`
	Volume    int64     `bson:"volume" json:"volume"`
	CreatedAt time.Time `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt time.Time `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}

type PriceRepository interface {
	UpsertCandles(ctx context.Context, ticker string, candles []PriceCandle) error
	GetPrices(ctx context.Context, ticker string, limit int) ([]*PriceCandle, error)
}
```

- [ ] **Step 4: Purify `internal/feature/finreport/entity.go` and `service.go`**

In `internal/feature/finreport/entity.go`:
```go
package finreport

import (
	"context"
	"time"
)

type FinancialReport struct {
	ID             string    `bson:"_id,omitempty" json:"id,omitempty"`
	IssuerCode     string    `bson:"issuer_code" json:"issuer_code"`
	ReportURL      string    `bson:"report_url" json:"report_url"`
	Year           int       `bson:"year" json:"year"`
	PeriodString   string    `bson:"period_string" json:"period_string"`
	AnnouncementID string    `bson:"announcement_id" json:"announcement_id"`
	DownloadedAt   time.Time `bson:"downloaded_at" json:"downloaded_at"`
	IsLatest       bool      `bson:"is_latest" json:"is_latest"`
	CreatedAt      time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at" json:"updated_at"`
}

type Repository interface {
	Create(ctx context.Context, report *FinancialReport) error
	FindByIssuerAndPeriod(ctx context.Context, issuerCode string, year int, periodString string) (*FinancialReport, error)
	UpdateIsLatest(ctx context.Context, issuerCode string, year int, periodString string, isLatest bool) error
	ListByIssuer(ctx context.Context, issuerCode string, limit int) ([]*FinancialReport, error)
}
```

In `internal/feature/finreport/service.go`:
Remove `go.mongodb.org` imports. Keep helper functions `ParseFinancialStatementFilename` and provide pure service methods.

- [ ] **Step 5: Run tests for `finreport`, `system`, and `stock`**

Run: `go test ./internal/feature/finreport ./internal/feature/system ./internal/feature/stock`
Expected: PASS

- [ ] **Step 6: Commit purified domain models**

```bash
git add internal/feature/system/ internal/feature/stock/ internal/feature/finreport/
git commit -m "refactor(domain): purify system, stock, and finreport domain entities and ports"
```

---

### Task 3: Purify Domain Entities & Ports in `xbrl`

**Files:**
- Modify: `internal/feature/xbrl/entity.go`
- Modify: `internal/feature/xbrl/entity_test.go`
- Test: `internal/feature/xbrl/calculator_test.go`, `internal/feature/xbrl/timing_test.go`

**Interfaces:**
- Consumes: `xbrl.Statement`, `xbrl.FactMap`
- Produces: Decoupled XBRL entity and `xbrl.Repository` port with zero external mongo imports.

- [ ] **Step 1: Update `internal/feature/xbrl/entity_test.go`**

Ensure `entity_test.go` does not import `go.mongodb.org/mongo-driver/v2/bson`.

- [ ] **Step 2: Update `internal/feature/xbrl/entity.go`**

Remove `import ("go.mongodb.org/mongo-driver/v2/bson"; "go.mongodb.org/mongo-driver/v2/mongo/options")`.
Update `Statement.ID` to `string`.
Update `Repository` port to:
```go
type Repository interface {
	Upsert(ctx context.Context, s *Statement) error
	FindByTickerAndPeriod(ctx context.Context, ticker string, year int, period string) (*Statement, error)
	FindHistoricalByTicker(ctx context.Context, ticker string, limit int) ([]*Statement, error)
	FindLatestByTicker(ctx context.Context, ticker string) (*Statement, error)
}
```

- [ ] **Step 3: Run all `xbrl` domain tests**

Run: `go test -v ./internal/feature/xbrl/...`
Expected: PASS

- [ ] **Step 4: Commit purified XBRL domain**

```bash
git add internal/feature/xbrl/
git commit -m "refactor(xbrl): remove mongo driver dependencies from domain entity and repository port"
```

---

### Task 4: Purify Domain Entities, Ports & Services in `news`

**Files:**
- Modify: `internal/feature/news/entity.go`
- Modify: `internal/feature/news/service.go`
- Modify: `internal/feature/news/batch.go`
- Modify: `internal/feature/news/service_test.go`
- Modify: `internal/feature/news/batch_test.go`

**Interfaces:**
- Consumes: `news.News`, `news.Briefing`, `news.NewsSummaryUpdate`
- Produces: Decoupled News domain model, ports, and AI summarization / briefing services.

- [ ] **Step 1: Write test for News domain ports and services**

Update `internal/feature/news/service_test.go` to mock pure `news.Repository` and `news.BriefingRepository`.

- [ ] **Step 2: Update `internal/feature/news/entity.go`**

Remove `go.mongodb.org` imports.
Update `News.ID` and `Briefing.ID` to `string`.
Add `NewsSummaryUpdate` struct:
```go
type NewsSummaryUpdate struct {
	Summary            string
	ValueScore         int
	ImpactDirection    string
	InvestmentTakeaway string
	Tickers            []string
	Sector             string
	Subsector          string
	Industry           string
	IsIndustryWide     bool
	Status             string
}
```
Update ports:
```go
type Repository interface {
	Create(ctx context.Context, news *News) error
	FindByID(ctx context.Context, id string) (*News, error)
	UpdateSummary(ctx context.Context, id string, summary *NewsSummaryUpdate) error
	ExistsByLink(ctx context.Context, link string) (bool, error)
	FindPendingSummary(ctx context.Context, limit int) ([]*News, error)
	FindRecent(ctx context.Context, limit int) ([]*News, error)
}

type BriefingRepository interface {
	Create(ctx context.Context, b *Briefing) error
	FindByDate(ctx context.Context, date time.Time) (*Briefing, error)
	FindLatest(ctx context.Context) (*Briefing, error)
	FindRecent(ctx context.Context, limit int) ([]*Briefing, error)
}

type Scraper interface {
	Scrape(ctx context.Context, startDate, endDate time.Time, onNewsFound func(*News) error) error
}
```

- [ ] **Step 3: Update `internal/feature/news/service.go` & `batch.go`**

Replace any `bson.M` or mongo driver calls with `repo.UpdateSummary` and standard domain types.

- [ ] **Step 4: Run tests in `internal/feature/news`**

Run: `go test -v ./internal/feature/news/...`
Expected: PASS

- [ ] **Step 5: Commit purified News domain**

```bash
git add internal/feature/news/
git commit -m "refactor(news): decouple news domain and briefing services from mongo driver"
```

---

### Task 5: Purify Domain Entities, Ports & Services in `announcement`

**Files:**
- Modify: `internal/feature/announcement/entity.go`
- Modify: `internal/feature/announcement/service.go`
- Modify: `internal/feature/announcement/service_test.go`

**Interfaces:**
- Consumes: `announcement.Announcement`, `announcement.Repository`, `announcement.IDXDataProvider`, `finreport.Repository`, `system.Repository`
- Produces: Pure announcement domain model and comprehensive disclosure sync & email filtering use case service.

- [ ] **Step 1: Write test for Announcement use case service**

In `internal/feature/announcement/service_test.go`, test `ProcessFinancialReportAnnouncement`, `FilterDisclosuresForEmail`, and `SyncDisclosures`.

- [ ] **Step 2: Update `internal/feature/announcement/entity.go`**

Remove `go.mongodb.org` imports.
Define:
```go
type Repository interface {
	Create(ctx context.Context, announcement *Announcement) error
	FindByID(ctx context.Context, id string) (*Announcement, error)
	Exists(ctx context.Context, id string) (bool, error)
	FindRecent(ctx context.Context, limit int) ([]*Announcement, error)
	GetLatestCreatedDate(ctx context.Context) (*time.Time, error)
	FindExistingIDs(ctx context.Context, limit int) (map[string]bool, error)
}

type IDXDataProvider interface {
	Fetch(ctx context.Context, dateFrom, dateTo string) ([]*Announcement, error)
}
```

- [ ] **Step 3: Update `internal/feature/announcement/service.go`**

Implement complete use cases in `announcement.Service`:
- `ProcessFinancialReportAnnouncement(ctx, a)` using `finreport.Repository.FindByIssuerAndPeriod` and `UpdateIsLatest`.
- `FilterDisclosuresForEmail(announcements)`: Extracts noise filtering regex and pattern exclusions from `cmd/`.
- `SyncDisclosures(ctx, dateFrom, dateTo, provider)`: Coordinates fetch, deduplication against `repo.FindExistingIDs()`, persistence, and financial report linking.

- [ ] **Step 4: Run tests in `internal/feature/announcement`**

Run: `go test -v ./internal/feature/announcement/...`
Expected: PASS

- [ ] **Step 5: Commit announcement domain**

```bash
git add internal/feature/announcement/
git commit -m "refactor(announcement): encapsulate announcement service use cases and purify ports"
```

---

### Task 6: Implement Driven MongoDB Persistence Adapters

**Files:**
- Modify: `internal/infra/db/mongo/system_repo.go`
- Modify: `internal/infra/db/mongo/price_repo.go`
- Modify: `internal/infra/db/mongo/financial_report_repo.go`
- Modify: `internal/infra/db/mongo/xbrl_repo.go`
- Modify: `internal/infra/db/mongo/news_repo.go`
- Modify: `internal/infra/db/mongo/briefing_repo.go`
- Modify: `internal/infra/db/mongo/announcement_repo.go`
- Modify: `internal/infra/db/mongo/price_repo_test.go`
- Modify: `internal/infra/db/mongo/xbrl_repo_test.go`
- Modify: `internal/infra/db/mongo/briefing_repo_test.go`

**Interfaces:**
- Consumes: Pure domain ports (`xbrl.Repository`, `announcement.Repository`, `finreport.Repository`, `news.Repository`, `news.BriefingRepository`, `stock.PriceRepository`, `system.Repository`)
- Produces: Concrete MongoDB driven adapters implementing all domain ports.

- [ ] **Step 1: Implement `internal/infra/db/mongo/system_repo.go`**

Implement `GetLastRun(ctx, scriptName)` and `SaveLastRun(ctx, lastRun)` using `bson.M{"scriptName": scriptName}` and upsert option.

- [ ] **Step 2: Implement `internal/infra/db/mongo/price_repo.go`**

Implement `UpsertCandles(ctx, ticker, candles)` and `GetPrices(ctx, ticker, limit)`.

- [ ] **Step 3: Implement `internal/infra/db/mongo/financial_report_repo.go`**

Implement `Create`, `FindByIssuerAndPeriod`, `UpdateIsLatest`, `ListByIssuer`.

- [ ] **Step 4: Implement `internal/infra/db/mongo/xbrl_repo.go`**

Implement `Upsert`, `FindByTickerAndPeriod`, `FindHistoricalByTicker`, `FindLatestByTicker`.

- [ ] **Step 5: Implement `internal/infra/db/mongo/news_repo.go` & `briefing_repo.go`**

Implement `Create`, `FindByID`, `UpdateSummary`, `ExistsByLink`, `FindPendingSummary`, `FindRecent` in `NewsRepository`.
Implement `Create`, `FindByDate`, `FindLatest`, `FindRecent` in `BriefingRepository`.

- [ ] **Step 6: Implement `internal/infra/db/mongo/announcement_repo.go`**

Implement `Create`, `FindByID`, `Exists`, `FindRecent`, `GetLatestCreatedDate`, `FindExistingIDs`.

- [ ] **Step 7: Update and run tests in `internal/infra/db/mongo`**

Run: `go test -v ./internal/infra/db/mongo/...`
Expected: PASS

- [ ] **Step 8: Commit MongoDB repository adapters**

```bash
git add internal/infra/db/mongo/
git commit -m "feat(mongo): implement pure domain repository ports in mongodb adapters"
```

---

### Task 7: Refactor Driving CLI Adapters (`cmd/*` & `tools/*`)

**Files:**
- Modify: `cmd/announcement/main.go`
- Modify: `cmd/scraper/main.go`
- Modify: `cmd/price_updater/main.go`
- Modify: `cmd/downloader/main.go`
- Modify: `cmd/xbrl_parser/main.go`
- Modify: `cmd/issuer/main.go`
- Modify: `tools/seed_ticker/main.go`
- Modify: `tools/reset_db/main.go`

**Interfaces:**
- Consumes: Application Services and Infra Adapters
- Produces: Clean CLI composition roots with zero raw MongoDB queries or inline business logic leaks.

- [ ] **Step 1: Refactor `cmd/announcement/main.go`**

Use `announcement.Service` and `system.Repository` with pure domain calls. Remove inline `bson.M` queries and filter loops.

- [ ] **Step 2: Refactor `cmd/scraper/main.go`**

Wire `news.Service` with `kontan.Scraper`, `NewsRepository`, `BriefingRepository`. Call application use case methods.

- [ ] **Step 3: Refactor `cmd/price_updater/main.go`**

Wire `yahoo.Client`, `priceRepo`, and `xbrlRepo`. Trigger daily price updates and valuation sync.

- [ ] **Step 4: Refactor `cmd/downloader/main.go` & `cmd/xbrl_parser/main.go`**

Ensure `downloader` and `xbrl_parser` use pure domain repositories and clean helper methods.

- [ ] **Step 5: Verify `tools/seed_ticker/main.go` & `tools/reset_db/main.go`**

Ensure tools build cleanly and interact with pure domain types.

- [ ] **Step 6: Test all CLI commands compile**

Run: `go build ./cmd/... ./tools/...`
Expected: PASS

- [ ] **Step 7: Commit CLI refactoring**

```bash
git add cmd/ tools/
git commit -m "refactor(cmd): clean driving adapters to use application services and pure repository ports"
```

---

### Task 8: End-to-End Verification, Documentation & Final Build

**Files:**
- Modify: `AGENTS.md`
- Modify: `README.md`

**Interfaces:**
- Consumes: Entire system
- Produces: Verified production builds, documentation, and zero technical debt.

- [ ] **Step 1: Run all Go test suites**

Run: `go test -v -race ./...`
Expected: PASS (all tests pass)

- [ ] **Step 2: Run Go static analysis**

Run: `go vet ./...`
Expected: Zero warnings/errors

- [ ] **Step 3: Build all Go binaries**

Run: `make build`
Expected: All binaries built into `bin/` successfully

- [ ] **Step 4: Verify Web App Build**

Run: `make web-build` or `npm --prefix idx-web run build`
Expected: Nitro server and Vue components build cleanly with zero type/BSON errors

- [ ] **Step 5: Update `AGENTS.md`**

Update `AGENTS.md` to reflect the purified Clean DDD Hexagonal boundaries, domain port methods, and repository responsibilities.

- [ ] **Step 6: Final Commit & Summary**

```bash
git add AGENTS.md README.md
git commit -m "docs: update architecture guide and documentation for clean DDD hexagonal compliance"
```
