# Engineering & Agent Guide: go-idx-scraper

Comprehensive architectural guide, engineering standards, valuation formulas, and actionable operational reference for human engineers and AI agents working on this repository.

---

## 1. System Architecture Overview

**go-idx-scraper** is a high-performance market intelligence, scraping, and forensic financial valuation terminal for the **Indonesia Stock Exchange (IDX)**.

The system is strictly partitioned into two decoupled tiers:

```
┌────────────────────────────────────────────────────────────────────────┐
│                          go-idx-scraper                                │
│           Go 1.24+ Data Collection, Ingestion & Analysis Layer         │
├────────────────────────────────────────────────────────────────────────┤
│ • Hexagonal DDD Architecture (Ports & Adapters)                        │
│ • Low-Memory Streaming XBRL / XML Financial Statements Parser          │
│ • Forensic Valuation Engine (Piotroski F-Score, Altman Z'', Graham,    │
│   ROIC, Margin of Safety)                                              │
│ • Multi-Source Scraping (Kontan News, IDX Announcements, Reports)      │
│ • Yahoo Finance Historical Daily Price Candles Ingestion               │
│ • MongoDB v2 Driver Persistence Engine                                 │
└──────────────────────────────────┬─────────────────────────────────────┘
                                   │ Shared MongoDB Database
                                   ▼
┌────────────────────────────────────────────────────────────────────────┐
│                              idx-web                                   │
│                Nuxt 4 / Nitro Server API & Dark Terminal UI            │
├────────────────────────────────────────────────────────────────────────┤
│ • Nitro Server Engine exposing REST API (`/api/v1/...`)                │
│ • Vue 3 Dark Terminal UI with Tailwind CSS & Lucide Icons             │
│ • Interactive SVG Price vs. Benjamin Graham Fair Value Chart           │
│ • Piotroski F-Score & Altman Z''-Score Forensic Badges                 │
│ • Multi-Year Financial Statements & Interactive Ratio Explorer         │
│ • Value Investing Screener & 7 AM GMT+8 Daily Market Briefings         │
└────────────────────────────────────────────────────────────────────────┘
```

> **CRITICAL ARCHITECTURAL RULE:**
> The Go codebase is **data collection, parsing, and analysis only**.
> **NEVER add HTTP servers, Chi/Gin routers, or web handlers to Go.**
> All REST API endpoints and web client features live exclusively in `idx-web/`.

---

## 2. Directory Tree & Component Index

```
go-idx-scraper/
├── cmd/                               # CLI Entry Points
│   ├── announcement/                  # Scrapes IDX official corporate disclosures
│   ├── downloader/                    # Headless Selenium downloader for XBRL/Excel filings
│   ├── issuer/                        # Updates official listed stock issuer registry
│   ├── price_updater/                 # Cron CLI syncing Yahoo daily prices & valuation multiples
│   ├── scraper/                       # Multi-channel news scraper & 7 AM GMT+8 briefing generator
│   └── xbrl_parser/                   # Standalone directory XBRL stream parser & ratio engine
├── internal/
│   ├── browser/                       # Secondary Adapter: Selenium WebDriver & Chrome lifecycle
│   ├── config/                        # YAML configuration loader with macOS/Linux auto-detection
│   ├── feature/                       # Pure DDD Domain Core & Application Use Cases (ZERO infra imports)
│   │   ├── announcement/              # Announcement domain entities, repository port, & sync use case
│   │   ├── common/                    # Shared domain primitives & types (Attachment)
│   │   ├── finreport/                 # Financial report domain models, repository port, & parser use case
│   │   ├── news/                      # News article & briefing domain models, ports, & summary use case
│   │   ├── stock/                     # Stock ticker & price candles domain models and repository port
│   │   ├── system/                    # System maintenance domain model (LastRun) & repository port
│   │   └── xbrl/                      # Statement entity, FactMap, repository port
│   │       └── calc/                  # Single-responsibility valuation & timing formulas (Graham, Piotroski, Altman Z, ROIC, Margins, Solvency, Split, Timing)
│   ├── helper/                        # Logging (zap), Excel parsing, file utils, email
│   └── infra/                         # External Driven Adapters (Driven Ports Implementation)
│       ├── db/mongo/                  # MongoDB v2 driver repositories (price, xbrl, news, system, etc.)
│       ├── idx/                       # IDX disclosure portal HTTP & web adapters
│       ├── scraper/kontan/            # Kontan financial news scraper
│       ├── xbrl/                      # Streaming XML/XBRL parser adapter & Excel parser
│       │   └── parser/                # Single-responsibility statement parsers (Income, Balance, Cash Flow, DEI, Shares, Dates, Zip)
│       └── yahoo/                     # Yahoo Finance API client (`{TICKER}.JK`)
├── tools/                             # Developer CLI Tools
│   ├── reset_db/                      # Database wipe and re-index utility
│   └── seed_ticker/                   # Single-ticker 5-year end-to-end historical data seeder
├── config/                            # Runtime configuration files (config.yml, config-mac.yml)
├── stock-list.json                    # Cached registry of all active IDX stock tickers
├── Makefile                           # Unified automation targets for Go and Web
└── idx-web/                           # Nuxt 4 Dark Terminal Web App & Nitro REST API
    ├── src/
    │   ├── assets/                    # Terminal CSS styles and themes
    │   ├── components/                # Vue 3 Terminal UI Components
    │   │   ├── AnnouncementsView.vue  # Corporate disclosure table with watched badges
    │   │   ├── ArticleModal.vue       # Read modal for news articles & disclosures
    │   │   ├── AuthModal.vue          # Firebase authentication modal
    │   │   ├── BriefingView.vue       # 7 AM GMT+8 Daily Market Intelligence Briefing
    │   │   ├── FinReportsView.vue     # Financial report filings browser
    │   │   ├── Navbar.vue             # Terminal header with status & ticker search
    │   │   ├── NewsTerminalView.vue   # Real-time multi-source financial news stream
    │   │   ├── OverviewView.vue       # Market snapshot dashboard
    │   │   ├── PriceValuationChart.vue# Interactive Price vs Graham Fair Value SVG chart
    │   │   ├── TickerFinancialsModal.vue # Ticker 360 view (Chart + Multi-Year Ratios)
    │   │   └── ValueScreenerView.vue  # Value investing multi-metric screener
    │   ├── composables/               # Vue composables (auth, theme, state)
    │   ├── pages/index.vue            # Single-page Dark Terminal layout
    │   └── server/                    # Nitro API Server
    │       ├── api/v1/                # REST API Endpoints
    │       │   ├── announcements/     # Announcements API
    │       │   ├── briefings/         # Daily briefings API
    │       │   ├── financial-reports/ # Reports API
    │       │   ├── news/              # News API
    │       │   ├── screener/value.get.ts # Quantitative Value Investing Screener
    │       │   ├── stocks/[ticker]/   # Ticker financials & historical price endpoints
    │       │   └── user/watchlist/    # User watchlist persistence
    │       ├── plugins/mongodb.ts     # MongoDB connection plugin
    │       └── utils/                 # Repositories, types, and Firebase admin
    └── nuxt.config.ts                 # Nuxt 4 configuration
```

---

## 3. Core Engineering Rules & Standards

### Must Follow

1. **Context Propagation**
   Pass `ctx context.Context` explicitly through all function calls across domain, service, repository, and scraper boundaries. Never use `context.Background()` inside business logic.
   ```go
   // Good
   func (s *Service) ProcessTicker(ctx context.Context, ticker string) error {
       return s.repo.Upsert(ctx, ticker)
   }
   ```

2. **Wrap Errors with `%w`**
   Add contextual information when propagating errors across architectural layers:
   ```go
   if err != nil {
       return fmt.Errorf("parsing xbrl instance for %s: %w", ticker, err)
   }
   ```

3. **Check Errors Immediately**
   Never ignore returned errors. Handle or propagate immediately after the call.

4. **Structured Logging with `zap.Logger`**
   Inject `*zap.Logger` into structs. Use structured key-value fields. Never use `fmt.Printf` or `log.Printf` in domain, infra, or service layers.
   ```go
   logger.Info("ingesting stock prices",
       zap.String("ticker", ticker),
       zap.Int("candles_count", len(candles)),
       zap.Float64("latest_close", latestPrice),
   )
   ```

5. **Hexagonal Domain Ports & Adapters Purity**
   - Domain entities and driven repository port interfaces live in `internal/feature/<name>/entity.go`.
   - Domain layers must have **ZERO external database or driver imports** (no `go.mongodb.org/mongo-driver/v2`).
   - Driven repository port interfaces must use explicit domain methods (e.g. `FindByTickerAndPeriod`, `GetLastRun`, `UpsertCandles`), never leaky generic signatures like `filter any` or `options.Lister`.
   - Concrete implementations live in `internal/infra/db/mongo/` or `internal/infra/<adapter>/`.

6. **Application Services Encapsulation**
   - Orchestration workflows, scraping pipelines, disclosure filtering, and business coordination live in `internal/feature/<name>/service.go`.
   - CLI entry points in `cmd/*` act purely as driving adapters (composition roots that parse flags, wire dependencies, and call application service methods).

7. **MongoDB Schema Conventions**
   - Use `go.mongodb.org/mongo-driver/v2` in persistence adapters (`internal/infra/db/mongo/`).
   - Use snake_case for `json` and `bson` struct tags.
   - Use standard `string` with `bson:"_id,omitempty" json:"id"` for document IDs in domain entities.
   - Use pointer types (e.g. `*string`, `*float64`) for optional fields.

8. **One Concern per File**
   Keep functions focused and cohesive. Separate calculation logic into single-responsibility files under `internal/feature/xbrl/calc/` (`graham.go`, `piotroski.go`, `altman_z.go`, `roic.go`, `profitability.go`, `solvency.go`, `split.go`, `timing.go`) from domain entities (`entity.go`).

### What to Avoid

- **NO Infrastructure Imports in Domain**: Never import MongoDB driver packages, Selenium, or HTTP libraries into `internal/feature/`.
- **NO Leaky Repository Ports**: Never pass raw `bson.M`, `filter any`, or `options.Lister` into domain port interfaces.
- **NO Business Logic in `cmd/*`**: CLI main files are strictly driving adapters and composition roots.
- **NO HTTP Servers in Go**: All REST APIs live in `idx-web/`.
- **NO Chi/Gin/Fiber Routers** in Go.
- **NO Hardcoded Credentials or Paths**: Use `config/config.yml` or CLI flags.
- **NO Goroutines without Context Cancellation**: Always tie goroutines to `ctx.Done()`.
- **NO Double-Conversion of Foreign Currency**: Never multiply already-converted IDR EPS/BVPS by FX rates again.
- **NO Unadjusted Pre-Split Multi-Year Per-Share Metrics**: Always apply stock-split normalization across historical statement series.
- **NO Silent Zero Operating Income**: Always apply fallback taxonomy mapping (`ProfitLossBeforeIncomeTax + FinanceCosts`).

---

## 4. Makefile Targets & CLI Command Reference

### Makefile Commands

| Target | Description | Example / Parameters |
|--------|-------------|----------------------|
| `make build` | Builds all 8 Go binaries into `bin/` | `make build` |
| `make seed-ticker` | 5-Year Seeder (XBRL + Prices + Valuation) | `make seed-ticker TICKER=BBRI YEARS=5` |
| `make reset-db` | Drops and re-indexes MongoDB collections | `make reset-db FLAGS="-force"` |
| `make update-prices` | Syncs daily Yahoo price candles into DB | `make update-prices TICKER=BBRI RANGE=5d` |
| `make briefing` | Scrapes news & generates 7 AM briefing | `make briefing` |
| `make scrape-news` | Alias for `make briefing` | `make scrape-news` |
| `make parse-xbrl` | Parses XBRL filings & computes ratios | `make parse-xbrl FLAGS="-ticker=BBRI"` |
| `make downloader` | Downloads XBRL/Excel filings via Selenium | `make downloader FLAGS="-ticker=BBRI -parse"` |
| `make announcement` | Scrapes official IDX disclosures | `make announcement` |
| `make issuer` | Updates active listed stock tickers list | `make issuer` |
| `make web` | Runs Nuxt 4 Dark Terminal in dev mode | `make web` (http://localhost:3000) |
| `make web-build` | Builds Nuxt 4 Web UI for production | `make web-build` |
| `make web-preview` | Previews built Nuxt 4 production server | `make web-preview` |
| `make web-install` | Installs npm dependencies in `idx-web/` | `make web-install` |
| `make test` | Runs all Go test suites | `make test` |
| `make vet` | Runs Go static analysis | `make vet` |
| `make clean` | Removes `bin/`, `.nuxt`, `.output` | `make clean` |

---

### Detailed CLI Binary Reference

#### 1. `tools/seed_ticker` — 5-Year Single Ticker Seeder
```bash
go run tools/seed_ticker/main.go [flags]
```
- `-ticker`: Target stock symbol (e.g. `BBRI`, `BBCA`, `TLKM`) [REQUIRED].
- `-years`: Year count (`5`), range (`2021-2025`), or list (`2021,2022,2023,2024,2025`) (default: `5`).
- `-periods`: Filing periods: `TW1,TW2,TW3,Audit` or `I,II,III,Tahunan` (default: `TW1,TW2,TW3,Audit`).
- `-skip-download`: Skip Selenium download and parse existing filings in `saham/`.
- `-clean-db`: Wipe prior statements and prices for this ticker before seeding.
- `-file-type`: Filing type to download (default: `instance.zip`).
- `-no-headless`: Run Chrome browser with visible GUI for debugging.
- `-config`: Path to config YAML file (default: `config/config.yml`).

#### 2. `tools/reset_db` — MongoDB Reset & Re-Indexing
```bash
go run tools/reset_db/main.go [flags]
```
- `-force` / `-f`: Force wipe without interactive terminal confirmation prompt.
- `-collections`: Comma-separated list of specific collections to wipe (default: all scraper collections).
- `-config`: Path to config file.

#### 3. `cmd/price_updater` — Yahoo Finance Price Sync Cron
```bash
go run cmd/price_updater/main.go [flags]
```
- `-ticker` / `-tickers`: Single ticker or comma-separated list (e.g. `BBRI,BBCA`). If omitted, updates all active tickers in `stock-list.json`.
- `-range`: Price history range: `5d`, `1mo`, `1y`, `5y`, `max` (default: `5d`).
- `-delay-ms`: Inter-request delay in milliseconds to avoid rate-limiting (default: `100`).
- `-stock-list`: Path to `stock-list.json` file.
- `-config`: Path to config file.

#### 4. `cmd/downloader` — Financial Report Downloader
```bash
go run cmd/downloader/main.go [flags]
```
- `-ticker`: Specific ticker or comma-separated list (e.g. `BBRI,TLKM`).
- `-years`: Comma-separated years or range (e.g. `2021-2025`).
- `-periods`: Filing periods (`TW1,TW2,TW3,Audit`).
- `-file-type`: `instance.zip`, `inlineXBRL.zip`, or `.xlsx`.
- `-parse`: Automatically stream-parse downloaded XBRL files into MongoDB upon download.
- `-clean`: Clear download directory before running.
- `-no-headless`: Disable headless Chrome mode.

#### 5. `cmd/xbrl_parser` — Standalone XBRL Ingestion Pipeline
```bash
go run cmd/xbrl_parser/main.go [flags]
```
- `-dir`: Directory containing `.zip` or `.xml` filings (default: `saham`).
- `-ticker`: Filter ingestion to specific ticker(s).
- `-clean-db`: Drop `xbrl_statements` collection before ingesting.

#### 6. `cmd/scraper` — Multi-Channel Scraper & Market Briefing
```bash
go run cmd/scraper/main.go [flags]
```
- `-start-date`: Scrape start date (`YYYY-MM-DD`, default yesterday GMT+8).
- `-end-date`: Scrape end date (`YYYY-MM-DD`, default today GMT+8).
- `-skip-briefing`: Scrape news articles without generating daily briefing.
- `-no-headless`: Disable headless Chrome mode.

---

## 5. Single-Ticker 5-Year Seeding Workflow

The single-ticker 5-year seeding tool (`make seed-ticker TICKER=BBRI YEARS=5`) orchestrates an end-to-end ingestion and valuation workflow:

```
                  ┌─────────────────────────────────────┐
                  │  make seed-ticker TICKER=BBRI       │
                  └──────────────────┬──────────────────┘
                                     │
                 1. Browser Automation (Selenium)
                                     ▼
        ┌────────────────────────────────────────────────────────┐
        │ Queries IDX Disclosure Portal for BBRI (2021 - 2025)   │
        │ Iterates TW1, TW2, TW3, Audit filing periods           │
        │ Downloads instance.zip packages into saham/            │
        └────────────────────────────┬───────────────────────────┘
                                     │
                 2. Streaming XML & Fact Extraction
                                     ▼
        ┌────────────────────────────────────────────────────────┐
        │ Unzips instance.xbrl stream in-memory                  │
        │ xml.NewDecoder extracts DEI metadata & /cor facts      │
        │ Maps multi-period ContextRefs into domain.FactMap      │
        └────────────────────────────┬───────────────────────────┘
                                     │
                 3. Forensic Ratios & Valuation Computation
                                     ▼
        ┌────────────────────────────────────────────────────────┐
        │ Computes Piotroski F-Score (0-9) & Altman Z''-Score    │
        │ Calculates ROIC, Gross/Operating/Net Margins, ROE/ROA  │
        │ Normalizes USD statements to IDR via filing FX rate    │
        │ Upserts into MongoDB xbrl_statements                   │
        └────────────────────────────┬───────────────────────────┘
                                     │
                 4. Yahoo Finance Price History Sync
                                     ▼
        ┌────────────────────────────────────────────────────────┐
        │ Queries Yahoo Finance API (BBRI.JK) for 5-year daily   │
        │ candles; upserts into MongoDB stock_prices             │
        │ Recalculates Benjamin Graham Fair Value & MOS %        │
        └────────────────────────────┬───────────────────────────┘
                                     │
                 5. Formatted Terminal Report Output
                                     ▼
        ┌────────────────────────────────────────────────────────┐
        │ Prints Tabular Valuation Matrix & Multi-Year Financials│
        │ Data instantly available in Nuxt 4 Terminal UI & Chart │
        └────────────────────────────────────────────────────────┘
```

---

## 6. XBRL Streaming Parser, Fact Maps & Taxonomy

### Low-Memory Streaming Architecture

IDX financial statement instance files can exceed several megabytes. The parser in `internal/infra/xbrl/parser/` uses `xml.NewDecoder` to stream XML tokens sequentially with $O(1)$ memory overhead across modular statement parsers (`income_statement.go`, `balance_sheet.go`, `cash_flow.go`, `dei.go`, `shares.go`, `dates.go`, `zip.go`):

```go
decoder := xml.NewDecoder(reader)
for {
    token, err := decoder.Token()
    if err == io.EOF { break }
    // Process StartElement tokens matching DEI and /cor taxonomies
}
```

### DEI Metadata Extraction (`/dei`)

The Document and Entity Information (DEI) taxonomy defines entity attributes:
- `EntityRegistrantName`: Full registered corporate name.
- `DocumentPeriodEndDate`: Filing period end date (`YYYY-MM-DD`).
- `CurrentFiscalYearEndDate`: Fiscal year end date (`--MM-DD`).
- `RoundingMultiplier`: Scale factor (e.g. `1`, `1000`, `1000000`).
- `Currency`: Statement reporting currency (`IDR`, `USD`).
- `ConversionRate`: Exchange rate to IDR for foreign currency filings.
- `EntitySharesOutstanding`: Total listed shares.

### Multi-Period Fact Maps (`domain.FactMap`)

Facts are stored in a two-dimensional map indexed by metric concept and `contextRef`:

```go
type FactMap map[string]map[string]FactValue

type FactValue struct {
    Value    float64 `json:"value" bson:"value"`
    Unit     string  `json:"unit" bson:"unit"`
    Decimals int     `json:"decimals" bson:"decimals"`
}
```

Standard IDX `contextRef` patterns:
- `CurrentYearInstant`: Balance sheet items at the end of the current period.
- `CurrentYearDuration`: Income statement and cash flow items for the current period.
- `PriorYearInstant` / `PriorYearDuration`: Comparative previous period line items.
- `RestatedInstant` / `RestatedDuration`: Restated comparative periods.

---

## 7. Forensic Valuation & Quantitative Formulas

All quantitative financial ratios and forensic scores are implemented as single-responsibility modules in `internal/feature/xbrl/calc/` (`graham.go`, `piotroski.go`, `altman_z.go`, `roic.go`, `profitability.go`, `solvency.go`, `currency.go`, `split.go`, `valuation_bands.go`, `timing.go`).

### 1. Piotroski F-Score (0 to 9 Integer Score)

Measures financial strength across 9 discrete binary criteria. A score of 8-9 indicates high financial strength; 0-2 indicates financial weakness.

| Category | # | Metric / Criterion | Formula / Logic | Points |
|----------|---|--------------------|-----------------|--------|
| **Profitability** | 1 | Positive ROA | `ROA > 0` | +1 |
| | 2 | Positive Operating Cash Flow | `CFO > 0` | +1 |
| | 3 | Earnings Quality | `CFO > Net Income` (Accruals check) | +1 |
| | 4 | Improving ROA | `ROA_current > ROA_prior` | +1 |
| **Leverage & Liquidity** | 5 | Lower Long-Term Debt | `LongTermDebt_current < LongTermDebt_prior` | +1 |
| | 6 | Higher Current Ratio | `CurrentRatio_current > CurrentRatio_prior` | +1 |
| | 7 | No Share Dilution | `Shares_current <= Shares_prior` | +1 |
| **Operating Efficiency** | 8 | Higher Gross Margin | `GrossMarginPct_current > GrossMarginPct_prior` | +1 |
| | 9 | Higher Asset Turnover | `AssetTurnover_current > AssetTurnover_prior` | +1 |

---

### 2. Emerging Market Altman Z''-Score

Evaluates corporate bankruptcy risk for emerging markets and non-manufacturing companies:

$$\text{Altman } Z''\text{-Score} = 6.56 X_1 + 3.26 X_2 + 6.72 X_3 + 1.05 X_4$$

Where:
- $X_1 = \frac{\text{Working Capital}}{\text{Total Assets}} = \frac{\text{Current Assets} - \text{Current Liabilities}}{\text{Total Assets}}$
- $X_2 = \frac{\text{Retained Earnings}}{\text{Total Assets}}$
- $X_3 = \frac{\text{Operating Income (EBIT)}}{\text{Total Assets}}$
- $X_4 = \frac{\text{Total Equity (Book Value)}}{\text{Total Liabilities}}$

**Risk Zones:**
- **Safe Zone:** $Z'' > 2.60$ (Low probability of insolvency)
- **Grey Zone:** $1.10 \le Z'' \le 2.60$ (Moderate risk / indeterminate)
- **Distress Zone:** $Z'' < 1.10$ (High insolvency risk)

---

### 3. Return on Invested Capital (ROIC)

Measures how efficiently a company allocates capital to generate profits:

$$\text{ROIC} = \frac{\text{NOPAT}}{\text{Invested Capital}}$$

Where:
- $\text{NOPAT} = \text{Operating Income} \times (1 - t)$ (using standard Indonesian corporate tax rate $t = 22\%$)
- $\text{Invested Capital} = (\text{Total Equity} + \text{Total Debt}) - \text{Cash \& Cash Equivalents}$

---

### 4. Benjamin Graham Fair Value (Graham Number)

Calculates the maximum fair price a defensive investor should pay:

$$\text{Graham Number} = \sqrt{22.5 \times \text{Normalized EPS} \times \text{Normalized BVPS}}$$

Where:
- $\text{Normalized EPS} = \frac{\text{Net Income} \times \text{FX Rate}}{\text{Shares Outstanding}}$ (in IDR)
- $\text{Normalized BVPS} = \frac{\text{Total Equity} \times \text{FX Rate}}{\text{Shares Outstanding}}$ (in IDR)
- $22.5 = 15 \times 1.5$ represents Graham's rule of thumb (Max P/E of 15 and Max P/B of 1.5).

---

### 5. Margin of Safety (MOS %)

$$\text{Margin of Safety \%} = \frac{\text{Graham Number} - \text{Current Stock Price}}{\text{Graham Number}} \times 100$$

- **Positive MOS %:** Stock trades at a discount to intrinsic Graham Fair Value.
- **Negative MOS %:** Stock trades at a premium above Graham Fair Value.

---

### 6. Additional Solvency & Profitability Metrics

- **Gross Margin:** $(\text{Gross Profit} / \text{Revenue}) \times 100$
- **Operating Margin:** $(\text{Operating Income} / \text{Revenue}) \times 100$
- **Net Margin:** $(\text{Net Income} / \text{Revenue}) \times 100$
- **Return on Equity (ROE):** $\text{Net Income} / \text{Total Equity}$
- **Return on Assets (ROA):** $\text{Net Income} / \text{Total Assets}$
- **Current Ratio:** $\text{Current Assets} / \text{Current Liabilities}$
- **Debt-to-Equity (D/E):** $\text{Total Debt} / \text{Total Equity}$
- **Net Debt:** $\text{Total Debt} - \text{Cash \& Equivalents}$
- **Interest Coverage:** $\text{Operating Income} / \text{Finance Costs}$
- **FCF Conversion %:** $(\text{Free Cash Flow} / \text{Net Income}) \times 100$
- **P/E Ratio:** $\text{Current Price} / \text{Normalized EPS}$
- **P/B Ratio:** $\text{Current Price} / \text{Normalized BVPS}$

---

## 8. Automated Cron Schedules & Pipelines

### 1. Daily 7:00 AM GMT+8 Market Briefing Pipeline
- **Schedule:** `0 7 * * *` (Daily 07:00 WITA / 06:00 WIB / 23:00 UTC).
- **CLI Command:** `make briefing` / `go run cmd/scraper/main.go`.
- **Actions:**
  1. Scrapes financial news headlines and articles from Kontan.
  2. Aggregates official disclosures from IDX disclosures portal.
  3. Synthesizes macro commentary and market highlights.
  4. Stores compiled briefing document into MongoDB `daily_briefings` collection.
  5. Displays prominently on the Dark Terminal web homepage.

### 2. Daily 5:00 PM GMT+8 Market Close Price Updater
- **Schedule:** `0 17 * * 1-5` (Monday–Friday 17:00 WITA / 16:00 WIB post-market close).
- **CLI Command:** `make update-prices` / `go run cmd/price_updater/main.go`.
- **Actions:**
  1. Fetches daily OHLCV price candles for listed tickers from Yahoo Finance (`{TICKER}.JK`).
  2. Upserts latest daily candles into `stock_prices` collection indexed on `{ ticker: 1, date: -1 }`.
  3. Updates the latest market price in `xbrl_statements`.
  4. Recomputes live valuation multiples: P/E, P/B, and Graham Margin of Safety %.

---

## 9. MongoDB Schema & Indexing Standards

All collections use compound indexes with unique constraints where appropriate to prevent data duplication across repeated ingestion runs.

### Collections & Indexes

1. **`stock_prices`**
   - **Key Fields:** `ticker` (string), `date` (time.Time), `open` (float64), `high` (float64), `low` (float64), `close` (float64), `volume` (int64), `adjusted_close` (float64).
   - **Unique Index:** `{ ticker: 1, date: -1 }` (Unique).

2. **`xbrl_statements`**
   - **Key Fields:** `metadata` (ticker, year, period, entity_name, currency), `core` (financial statement line items), `computed_ratios` (Piotroski, Altman Z, ROIC, Margins), `valuation` (Graham Number, MOS %, P/E, P/B), `facts` (raw FactMap).
   - **Unique Index:** `{ "metadata.ticker": 1, "metadata.year": -1, "metadata.period": 1 }` (Unique).

3. **`daily_briefings`**
   - **Key Fields:** `date` (string `YYYY-MM-DD`), `title` (string), `summary` (string), `highlights` (array), `sentiment` (string).
   - **Index:** `{ date: -1 }` (Unique).

4. **`financial_reports`**
   - **Key Fields:** `symbol` (string), `year` (string), `period` (string), `file_path` (string), `file_type` (string).
   - **Index:** `{ symbol: 1, year: -1, period: 1 }`.

5. **`announcements`**
   - **Key Fields:** `id` (string), `ticker` (string), `title` (string), `published_at` (time.Time), `url` (string).
   - **Index:** `{ id: 1 }` (Unique), `{ published_at: -1 }`.

6. **`news`**
   - **Key Fields:** `id` (string), `ticker` (string), `title` (string), `published_at` (time.Time), `url` (string), `summary` (string).
   - **Index:** `{ url: 1 }` (Unique), `{ published_at: -1 }`, `{ ticker: 1 }`.

7. **`last_runs`**
   - **Key Fields:** `feature` (string), `last_run` (time.Time), `status` (string).
   - **Index:** `{ feature: 1 }` (Unique).

---

## 10. `idx-web` (Nuxt 4 Dark Terminal UI & REST API)

`idx-web/` is a modern, responsive Dark Terminal web application powered by **Nuxt 4**, **Nitro server engine**, and **Tailwind CSS**.

### REST API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/stocks/:ticker/financials` | `GET` | Returns multi-year historical XBRL statements, ratios, and Graham valuation |
| `/api/v1/stocks/:ticker/prices?range=5y` | `GET` | Returns daily price candles for interactive charting (`5d`, `1mo`, `1y`, `5y`, `max`) |
| `/api/v1/screener/value` | `GET` | Quantitative screener filtering by Piotroski, Altman Z, ROIC, and Graham MOS |
| `/api/v1/briefings/latest` | `GET` | Latest 7 AM GMT+8 daily market intelligence briefing |
| `/api/v1/news` | `GET` | Paginated financial news stream with ticker filters |
| `/api/v1/announcements` | `GET` | Official IDX corporate disclosures and corporate actions |
| `/api/v1/user/watchlist` | `GET` / `PUT` | User watchlist persistence (requires Firebase auth token) |

### Terminal UI Components (`idx-web/src/components/`)

- **`PriceValuationChart.vue`**: Custom interactive SVG financial price chart plotting daily price candles against the Benjamin Graham Fair Value band, with a color-coded Margin of Safety discount/premium area.
- **`TickerFinancialsModal.vue`**: Complete Ticker 360 view with integrated price vs. fair value chart, Piotroski 9-point score checklist, Altman Z''-score safety breakdown, and multi-year interactive financial statements table.
- **`ValueScreenerView.vue`**: Multi-factor value screener filterable by minimum Piotroski score, Altman Z zone, ROIC %, and Margin of Safety %.
- **`BriefingView.vue`**: Daily market overview displaying macro indicators, sector sentiment, and key events.
- **`NewsTerminalView.vue` & `AnnouncementsView.vue`**: Real-time terminal feeds for news and corporate disclosures.

### Running & Building the Web UI

```bash
# Install dependencies
npm --prefix idx-web install

# Run development server (localhost:3000)
make web

# Build production server
make web-build

# Preview production build
make web-preview
```

---

## 11. Environment Configuration

### Go Configuration (`config/config.yml` or `config/config-mac.yml`)

```yaml
database:
  db_name: idx_scraper
  mongo_uri: mongodb://localhost:27017

paths:
  download_dir: saham
  stock_list: stock-list.json

browser:
  headless: true
  chromedriver_path: /usr/local/bin/chromedriver
  port: 9515

scraper:
  delay_ms: 200
  timeout_sec: 30
```

### Web Environment (`idx-web/.env`)

```bash
MONGODB_URI=mongodb://localhost:27017
MONGODB_DB_NAME=idx_scraper
FIREBASE_PROJECT_ID=
FIREBASE_CREDENTIALS_PATH=
```

---

## 12. Financial Valuation, Foreign Currency & Stock Split Integrity Rules

To guarantee mathematical correctness and prevent data distortions, all AI agents and human engineers must adhere strictly to these quantitative rules:

### Rule 1: Single-Pass Foreign Currency (USD) Normalization
1. When parsing USD-denominated filings (`s.Metadata.Currency == "USD"`), extract `ConversionRate`. If written in Indonesian dot notation (e.g. `16.680` parsed as `16.68`), multiply by $1,000$ (`if rate > 0 && rate < 1000 { rate *= 1000 }`).
2. `NormalizedEPS` and `NormalizedBVPS` must be converted to IDR **exactly once**:
   $$\text{NormalizedEPS}_{\text{IDR}} = \text{NormalizedEPS}_{\text{USD}} \times \text{ConversionRate}$$
   $$\text{NormalizedBVPS}_{\text{IDR}} = \frac{\text{Total Equity} \times \text{ConversionRate}}{\text{Shares Outstanding}}$$
3. Never multiply an already IDR-normalized per-share metric by the exchange rate a second time.
4. The Benjamin Graham Number $\sqrt{22.5 \times \text{EPS}_{\text{IDR}} \times \text{BVPS}_{\text{IDR}}}$ must always evaluate in nominal IDR space to directly match IDX market stock prices.

### Rule 2: Stock Split Adjustment Across Multi-Year Time Horizons
1. Yahoo Finance daily price time series are retroactively **split-adjusted**.
2. When ingesting multi-year historical statements across a corporate stock split (e.g. `DSSA` 1:10 split in 2024, `BBCA` 1:5 split in 2021, `BBRI` 1:5 split in 2017), the engine must call `ApplyStockSplitAdjustment(statements)`:
   $$\text{Split Ratio}_t = \frac{\text{Shares Outstanding}_{\text{latest}}}{\text{Shares Outstanding}_t}$$
   $$\text{Adjusted EPS}_t = \frac{\text{Net Income}_t \times \text{FX}_t}{\text{Shares Outstanding}_{\text{latest}}}$$
   $$\text{Adjusted BVPS}_t = \frac{\text{Total Equity}_t \times \text{FX}_t}{\text{Shares Outstanding}_{\text{latest}}}$$
   $$\text{Adjusted Graham Number}_t = \sqrt{22.5 \times \text{Adjusted EPS}_t \times \text{Adjusted BVPS}_t}$$
3. Financial statement totals (Revenue, Net Income, Total Equity, Free Cash Flow) and core percentage ratios (**ROIC, ROE, Gross Margin, Operating Margin, Piotroski F-Score, Altman Z''-Score**) are **immutable** and must **never** be divided by split ratios.

### Rule 3: Operating Income (EBIT) Taxonomy & Sector Fallback Guarantees
1. Certain issuers (e.g. mining, energy, conglomerates) do not use standard `idx-cor:OperatingIncome` tags.
2. In `assignCoreMetric` and `finalizeCoreFinancials`, always apply cascading fallbacks:
   $$\text{Operating Income (EBIT)} = \text{OperatingIncome} \lor (\text{ProfitLossBeforeIncomeTax} + \text{FinanceCosts}) \lor (\text{NetIncome} + \text{FinanceCosts})$$
3. Never allow `OperatingIncome` or `ROIC` to default silently to $0.00$ when Gross Profit or Pre-Tax Income is positive.

### Rule 4: Parent Entity Net Income Prioritization
1. Always prioritize `ProfitLossAttributableToOwnersOfParentEntity` (`NetIncomeParent`) over total consolidated `ProfitLoss` when computing EPS, ROE, and equity valuation to prevent distorting shareholder equity in conglomerates with heavy non-controlling interests.

### Rule 5: Zero-Debt & Positive Equity Piotroski Scoring
1. In Piotroski Criterion 5 (Long-Term Debt), zero-debt balance sheets ($0 \le 0$) must be awarded the +1 point for pristine solvency.
2. Never award baseline debt credit to insolvent firms with negative total equity.

### Rule 6: Chronological Sequence & Period Sorting
1. MongoDB repositories and Nitro API handlers must sort statements by `{ year: -1, period_end_date: -1 }` (not alphabetical `period: -1`) to ensure audited annual statements (`12-31`) sort before interim quarters (`09-30`, `06-30`, `03-31`).

---

## 13. Commit & Contribution Guidelines

- Use conventional imperative commit messages:
  - `feat(prices): ...`
  - `feat(xbrl): ...`
  - `feat(tools): ...`
  - `feat(web): ...`
  - `docs: ...`
  - `fix: ...`
- Keep commits focused and atomic.
- Verify all Go test suites (`make test`) and static analysis (`make vet`) pass before submitting.
- Verify all Go binaries compile cleanly via `make build`.
- Verify Nuxt 4 production build passes via `make web-build`.
- Run regression unit tests covering multi-currency and stock-split tickers (`TestParseInstanceXML_DSSA`, `TestApplyStockSplitAdjustment_DSSA_10to1`, `TestCalculator_USDEPSNormalization`, `TestCalculator_ZeroDebtAndParentNetIncome`).
