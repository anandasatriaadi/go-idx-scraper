# Design Specification: Clean DDD Hexagonal Refactoring & Cleanup

## 1. Overview & Objectives

This specification outlines the architectural refactoring of **go-idx-scraper** to achieve strict compliance with **Clean Architecture, Domain-Driven Design (DDD), and Hexagonal Architecture (Ports and Adapters)** principles.

### Key Objectives
1. **Zero Infrastructure Dependencies in Domain**: Purge all `go.mongodb.org/mongo-driver/v2` packages (`bson.ObjectID`, `bson.M`, `options.Lister`) from `internal/feature/*`.
2. **Explicit Domain Driven Ports**: Replace leaky generic repository signatures (`filter any, opts ...options.Lister`) with explicit, strongly-typed domain methods.
3. **Application Use Case Encapsulation**: Move business orchestration and query logic out of driving CLI adapters (`cmd/*`) into dedicated application services (`internal/feature/*/service.go`).
4. **Adapter Purity**: Encapsulate all database specifics (filters, projections, sorting, indexes) inside `internal/infra/db/mongo/`.
5. **Zero Web / REST API Regression**: Maintain 100% database schema compatibility with `idx-web`.
6. **Codebase Hygiene**: Prune untracked artifacts, empty directories, and obsolete boilerplate generators.

---

## 2. Target Architecture & Package Layout

```
go-idx-scraper/
├── cmd/                               # Driving Adapters (CLI Entry Points)
│   ├── announcement/                  # Disclosures scraper CLI
│   ├── downloader/                    # Headless filing downloader CLI
│   ├── issuer/                        # Stock registry updater CLI
│   ├── price_updater/                 # Daily Yahoo price & valuation cron CLI
│   ├── scraper/                       # News scraper & briefing generator CLI
│   └── xbrl_parser/                   # XBRL streaming parser CLI
├── internal/
│   ├── browser/                       # Secondary Adapter: Selenium WebDriver & Chrome
│   ├── config/                        # Cross-cutting: Configuration loader & validation
│   ├── helper/                        # Cross-cutting: Logger, Excel, and file utilities
│   ├── feature/                       # Core Business Domains & Application Services (ZERO infra imports)
│   │   ├── announcement/              # Announcement entity, repository port, & sync use case
│   │   ├── common/                    # Shared domain primitives (Attachment)
│   │   ├── finreport/                 # Financial report entity, repository port, & parser use case
│   │   ├── news/                      # News & Briefing entities, repository ports, & summary use case
│   │   ├── stock/                     # Stock & PriceCandle entities, repository ports
│   │   ├── system/                    # System maintenance entity (LastRun) & repository port
│   │   └── xbrl/                      # Statement entity, FactMap, Valuation & Timing Calculators, repository port
│   └── infra/                         # Driven Adapters (External Implementations)
│       ├── db/mongo/                  # MongoDB persistence adapters implementing feature repository ports
│       ├── idx/                       # IDX disclosure HTTP adapter implementing announcement.IDXDataProvider
│       ├── scraper/kontan/            # Kontan scraper adapter implementing news.Scraper
│       ├── xbrl/                      # Streaming XML & Excel parser adapters
│       └── yahoo/                     # Yahoo Finance market data adapter
├── tools/                             # Developer Utilities
│   ├── reset_db/                      # Database wipe & re-index utility
│   └── seed_ticker/                   # Single-ticker 5-year historical seeder
├── config/                            # Runtime YAML configuration
├── stock-list.json                    # Active IDX stock tickers registry
└── Makefile                           # Unified automation targets
```

---

## 3. Domain Model & Port Specifications

### 3.1 `internal/feature/xbrl`
- **Models**:
  - `Statement`: `ID` typed as `string` (`bson:"_id,omitempty" json:"id"`).
  - Retain `StatementMetadata`, `CoreFinancials`, `ComputedRatios`, `ValuationMetrics`, `ValuationBands`, `TimingSignal`, `FactMap`, `FactValue`.
  - Zero imports from `go.mongodb.org/mongo-driver/v2`.
- **Driven Port (`Repository`)**:
  ```go
  type Repository interface {
      Upsert(ctx context.Context, s *Statement) error
      FindByTickerAndPeriod(ctx context.Context, ticker string, year int, period string) (*Statement, error)
      FindHistoricalByTicker(ctx context.Context, ticker string, limit int) ([]*Statement, error)
      FindLatestByTicker(ctx context.Context, ticker string) (*Statement, error)
  }
  ```

### 3.2 `internal/feature/announcement`
- **Models**:
  - `Announcement`: `ID` typed as `string` (`bson:"_id,omitempty" json:"id"`).
  - Zero imports from `go.mongodb.org/mongo-driver/v2`.
- **Driven Port (`Repository`)**:
  ```go
  type Repository interface {
      Create(ctx context.Context, announcement *Announcement) error
      FindByID(ctx context.Context, id string) (*Announcement, error)
      Exists(ctx context.Context, id string) (bool, error)
      FindRecent(ctx context.Context, limit int) ([]*Announcement, error)
      GetLatestCreatedDate(ctx context.Context) (*time.Time, error)
      FindExistingIDs(ctx context.Context, limit int) (map[string]bool, error)
  }
  ```
- **Driven Port (`IDXDataProvider`)**:
  ```go
  type IDXDataProvider interface {
      Fetch(ctx context.Context, dateFrom, dateTo string) ([]*Announcement, error)
  }
  ```

### 3.3 `internal/feature/finreport`
- **Models**:
  - `FinancialReport`: `ID` typed as `string` (`bson:"_id,omitempty" json:"id"`).
  - Zero imports from `go.mongodb.org/mongo-driver/v2`.
- **Driven Port (`Repository`)**:
  ```go
  type Repository interface {
      Create(ctx context.Context, report *FinancialReport) error
      FindByIssuerAndPeriod(ctx context.Context, issuerCode string, year int, periodString string) (*FinancialReport, error)
      UpdateIsLatest(ctx context.Context, issuerCode string, year int, periodString string, isLatest bool) error
      ListByIssuer(ctx context.Context, issuerCode string, limit int) ([]*FinancialReport, error)
  }
  ```

### 3.4 `internal/feature/news`
- **Models**:
  - `News`: `ID` typed as `string` (`bson:"_id,omitempty" json:"id"`).
  - `Briefing`: `ID` typed as `string` (`bson:"_id,omitempty" json:"id"`).
  - `NewsSummaryUpdate`: DTO containing summary fields.
  - Zero imports from `go.mongodb.org/mongo-driver/v2`.
- **Driven Ports**:
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

### 3.5 `internal/feature/stock`
- **Models**:
  - `PriceCandle`: `ID` typed as `string` (`bson:"_id,omitempty" json:"id,omitempty"`).
  - `StockData`, `StockListResponse`.
  - Zero imports from `go.mongodb.org/mongo-driver/v2`.
- **Driven Port (`PriceRepository`)**:
  ```go
  type PriceRepository interface {
      UpsertCandles(ctx context.Context, ticker string, candles []PriceCandle) error
      GetPrices(ctx context.Context, ticker string, limit int) ([]*PriceCandle, error)
  }
  ```

### 3.6 `internal/feature/system`
- **Models**:
  - `LastRun`: `ID` typed as `string`, `Metadata` typed as `map[string]any`.
  - Zero imports from `go.mongodb.org/mongo-driver/v2`.
- **Driven Port (`Repository`)**:
  ```go
  type Repository interface {
      GetLastRun(ctx context.Context, scriptName string) (*LastRun, error)
      SaveLastRun(ctx context.Context, lastRun *LastRun) error
  }
  ```

---

## 4. Application Use Case Layer (`internal/feature/*/service.go`)

Application services encapsulate orchestration and domain workflow:

1. **`announcement.Service`**:
   - `SyncDisclosures(ctx, startDate, endDate, idxProvider)`: Coordinates fetching from provider, filtering new announcements, persistence in repo, financial report linking, and updating last run state.
   - `ProcessFinancialReportAnnouncement(ctx, a)`: Resolves financial report attachments and marks latest reports.
   - `FilterDisclosuresForEmail(announcements)`: Extracts noise filtering logic.
2. **`news.Service`**:
   - `ScrapeNews(ctx, scraper, startDate, endDate)`: Scrapes articles, deduplicates by link, and stores pending news.
   - `SummarizePendingNews(ctx, limit)`: Runs AI summarization prompts and updates news items via `UpdateSummary`.
   - `GenerateDailyBriefing(ctx, date)`: Gathers summarized news, runs daily briefing synthesis prompt, and saves briefing.
3. **`finreport.Service`**:
   - `ParseFinancialStatementFilename(filename)`: Pure parsing domain service.
   - `ProcessFinancialReports(ctx, ...)`: Coordinates downloading and XBRL/Excel parsing.

---

## 5. Infrastructure Persistence Layer (`internal/infra/db/mongo/`)

All MongoDB-specific mechanics (`bson.M`, `bson.D`, `options.Find()`, `mongo.Collection`, `mongo.ErrNoDocuments`) are encapsulated in repository implementations:

- `XBRLRepository`: Implements `xbrl.Repository`. Handles indexes on `metadata.ticker, metadata.year, metadata.period`.
- `AnnouncementRepository`: Implements `announcement.Repository`. Handles indexes on `_id` and `created_date`.
- `FinancialReportRepository`: Implements `finreport.Repository`. Handles queries for issuer and period.
- `NewsRepository`: Implements `news.Repository`. Handles URL deduplication and pending summary queries.
- `BriefingRepository`: Implements `news.BriefingRepository`. Handles date-based lookup and latest retrieval.
- `PriceRepository`: Implements `stock.PriceRepository`. Handles compound index `ticker: 1, date: -1` and bulk upserts.
- `SystemRepository`: Implements `system.Repository`. Handles `last_runs` upserts with timestamp tracking.

---

## 6. Driving Adapters (CLI Commands in `cmd/*`)

Refactor all CLI entry points to act strictly as driving adapters:
- `cmd/announcement/main.go`: Parse flags $\rightarrow$ load config $\rightarrow$ init logger/DB/browser $\rightarrow$ call `announcementService.SyncDisclosures()`.
- `cmd/scraper/main.go`: Parse flags $\rightarrow$ load config $\rightarrow$ init logger/DB $\rightarrow$ call `newsService.ScrapeAndIngest()` $\rightarrow$ call `newsService.GenerateDailyBriefing()`.
- `cmd/price_updater/main.go`: Parse flags $\rightarrow$ load config $\rightarrow$ init logger/DB/yahoo $\rightarrow$ fetch candles $\rightarrow$ `priceRepo.UpsertCandles()`.
- `cmd/downloader/main.go`: Parse flags $\rightarrow$ load config $\rightarrow$ init logger/browser $\rightarrow$ orchestrate file downloads & parsing.
- `cmd/xbrl_parser/main.go`: Parse flags $\rightarrow$ load config $\rightarrow$ init logger/DB $\rightarrow$ parse directory into statements $\rightarrow$ `xbrlRepo.Upsert()`.
- `cmd/issuer/main.go`: Parse flags $\rightarrow$ load config $\rightarrow$ fetch listed tickers $\rightarrow$ update `stock-list.json`.

---

## 7. Cleanup & Dead Code Removal

1. **Delete Untracked File**: `D:\sync-gitrepo\go-idx-scraper\stock-list-2.json`.
2. **Prune Empty Directories**: `internal/infra/external/openrouter`, `internal/infra/external/firebase`, `internal/infra/external/email`.
3. **Remove Obsolete Generator**: Delete `tools/mongo_repo/` and delete all `//go:generate` tags referencing it.
4. **Update `AGENTS.md`**: Update architectural diagrams, package trees, and rules to match clean ports and adapters.

---

## 8. Testing & Verification Plan

1. **Unit & Package Tests**: Run `go test -v -race ./...` across all packages (`internal/feature/*`, `internal/infra/*`, `internal/helper/*`, `tools/*`).
2. **Static Analysis**: Run `go vet ./...` and verify zero diagnostics.
3. **Binary Compilation**: Run `make build` and confirm all 6 CLI binaries compile cleanly into `bin/`.
4. **Web UI Compatibility**: Run `npm --prefix idx-web run build` / `make web-build` and verify zero breakages in Nitro server and Vue terminal UI.
