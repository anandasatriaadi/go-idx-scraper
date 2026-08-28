# go-idx-scraper

> High-performance market intelligence, streaming XBRL financial parser, and forensic valuation terminal for the **Indonesia Stock Exchange (IDX)**.

---

## 1. System Architecture

`go-idx-scraper` follows a strict **Clean Architecture, Domain-Driven Design (DDD), and Hexagonal Architecture (Ports and Adapters)** pattern.

```
┌────────────────────────────────────────────────────────────────────────┐
│                          go-idx-scraper                                │
│           Go 1.24+ Data Collection, Ingestion & Analysis Layer         │
├────────────────────────────────────────────────────────────────────────┤
│ • Hexagonal DDD Architecture (Zero infra imports in Domain)            │
│ • Low-Memory Streaming XBRL / XML Financial Statements Parser          │
│ • Forensic Valuation Engine (Piotroski F-Score, Altman Z'', Graham,    │
│   ROIC, Margin of Safety, Technical Timing Signals)                    │
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

---

## 2. Directory Structure

```
go-idx-scraper/
├── cmd/                               # Driving Adapters (CLI Entry Points)
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
```

---

## 3. Quick Start & Makefile Commands

### Installation & Prerequisites
- **Go 1.24+**
- **Node.js 20+ & npm**
- **MongoDB 7.0+** running on `localhost:27017`
- **Google Chrome & ChromeDriver** (for Selenium-based downloaders)

### Common Commands

```bash
# Build all 8 Go binaries into bin/
make build

# 5-Year Historical Seeder (XBRL statements + Yahoo prices + Graham/Piotroski valuation)
make seed-ticker TICKER=BBRI YEARS=5

# Sync daily price candles from Yahoo Finance
make update-prices TICKER=BBRI RANGE=5d

# Run 7 AM GMT+8 Multi-channel news scraper & daily market briefing
make briefing

# Parse XBRL filings in saham/ into MongoDB
make parse-xbrl FLAGS="-ticker=BBRI"

# Run tests and static analysis
make test
make vet

# Run Nuxt 4 Dark Terminal Web App
make web-install
make web
```

---

## 4. Engineering Standards & Architectural Integrity

- **Pure Domain Core**: `internal/feature/*` has **zero external imports** of MongoDB drivers or HTTP libraries.
- **Explicit Driven Ports**: Repository interfaces define domain-typed methods (`FindByTickerAndPeriod`, `GetLastRun`, `UpsertCandles`), without leaky generic options.
- **Driving Adapters in CLI**: Entry points in `cmd/*` act purely as composition roots that parse flags, wire dependencies, and call application use case services.
- **Separation of Concerns**: REST APIs and web client components live exclusively in `idx-web/`.

For complete valuation formulas, financial integrity rules, and database schema documentation, see [AGENTS.md](AGENTS.md).
