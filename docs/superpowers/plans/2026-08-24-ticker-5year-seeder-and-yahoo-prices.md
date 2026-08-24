# Ticker 5-Year Seeder, Yahoo Historical Prices & Configurable Pipeline Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a configurable data pipeline that can ingest, parse, and evaluate 5 years of historical XBRL financial statements and Yahoo Finance daily price candles for any IDX ticker, store price time-series in MongoDB, expose price history via API, render interactive Price vs. Fair Value charts in the Dark Terminal UI, provide developer tools in `./tools/`, and document everything in an updated `AGENTS.md`.

**Architecture:** 
- Go data collection layer (`internal/infra/yahoo`, `internal/infra/xbrl`, `internal/infra/db/mongo`, `cmd/downloader`, `cmd/xbrl_parser`, `tools/seed_ticker`).
- Nuxt 4 / Nitro API endpoint (`/api/v1/stocks/:ticker/prices`).
- Vue 3 Dark Terminal UI with interactive SVG/canvas financial price chart and valuation bands.
- Developer & Agent guide in `AGENTS.md`.

**Tech Stack:** Go 1.24+, MongoDB v2 Driver, Yahoo Finance API (`{TICKER}.JK`), Nuxt 4 / Vue 3, Makefile.

## Global Constraints

- Use strict `context: "fresh"` for all subagent dispatches (no session forking).
- Never use `context.Background()` in domain logic.
- Wrap errors with `%w`.
- Store daily price history in MongoDB collection `stock_prices` indexed on `{ ticker: 1, date: -1 }`.
- Downloader and parser must accept CLI flags for individual tickers, year ranges, and auto-parse flags.
- Comprehensive documentation in `AGENTS.md`.

---

### Task 1: Yahoo Finance Price Client & Price Repository Adapter

**Files:**
- Create: `internal/infra/yahoo/client.go`
- Create: `internal/infra/yahoo/client_test.go`
- Create: `internal/infra/db/mongo/price_repo.go`
- Create: `internal/infra/db/mongo/price_repo_test.go`
- Create: `cmd/price_updater/main.go`

**Interfaces:**
- `yahoo.FetchHistoricalPrices(ticker string, rangePeriod string) ([]PriceCandle, error)`
- `mongo.PriceRepository`: `UpsertCandles(ctx, ticker, candles)`, `GetPrices(ctx, ticker, limit)`
- `cmd/price_updater`: Daily market close cron CLI command (`0 17 * * 1-5`) updating prices and refreshing valuation multiples.

- [ ] **Step 1: Write test for Yahoo Finance price client**
- [ ] **Step 2: Implement `internal/infra/yahoo/client.go`**
- [ ] **Step 3: Implement `internal/infra/db/mongo/price_repo.go` with `{ ticker: 1, date: -1 }` unique index**
- [ ] **Step 4: Implement `cmd/price_updater/main.go` for daily market close cron execution**
- [ ] **Step 5: Run tests (`go test -v ./internal/infra/yahoo/... ./internal/infra/db/mongo/...`)**
- [ ] **Step 6: Commit Task 1**

```bash
git add internal/infra/yahoo/ internal/infra/db/mongo/ cmd/price_updater/
git commit -m "feat(prices): implement Yahoo Finance price client, MongoDB repository and daily price updater cron CLI"
```

---

### Task 2: Configurable Downloader & Standalone XBRL Ingestion Pipeline

**Files:**
- Modify: `cmd/downloader/main.go`
- Modify: `cmd/xbrl_parser/main.go`

**Interfaces:**
- `cmd/downloader`: Supports `-ticker=BBRI`, `-years=2021,2022,2023,2024,2025`, `-periods=TW1,TW2,TW3,Audit`, `-file-type=instance.zip`, `-parse=true`.
- `cmd/xbrl_parser`: Supports `-ticker=BBRI`, `-clean-db=true`, `-dir=saham`.

- [ ] **Step 1: Update `cmd/downloader/main.go` with configurable CLI flags and auto-parse integration**
- [ ] **Step 2: Update `cmd/xbrl_parser/main.go` with ticker filter and clean-db flags**
- [ ] **Step 3: Test compilation (`go build -o bin/ ./cmd/...`)**
- [ ] **Step 4: Commit Task 2**

```bash
git add cmd/downloader/main.go cmd/xbrl_parser/main.go
git commit -m "feat(cli): enhance downloader and xbrl_parser with configurable ticker, multi-year and auto-parse flags"
```

---

### Task 3: Developer Tools in `./tools/` (`seed_ticker` & `reset_db`)

**Files:**
- Create: `tools/seed_ticker/main.go`
- Create: `tools/reset_db/main.go`

**Interfaces:**
- `tools/seed_ticker`: One-command 5-year historical seeder (XBRL download + stream parse + Yahoo price history + valuation report).
- `tools/reset_db`: Wipes and re-indexes all collections cleanly.

- [ ] **Step 1: Create `tools/reset_db/main.go`**
- [ ] **Step 2: Create `tools/seed_ticker/main.go`**
- [ ] **Step 3: Test tool compilation (`go build -o bin/seed_ticker ./tools/seed_ticker` and `go build -o bin/reset_db ./tools/reset_db`)**
- [ ] **Step 4: Commit Task 3**

```bash
git add tools/
git commit -m "feat(tools): add seed_ticker and reset_db developer tools"
```

---

### Task 4: Nuxt 4 API Endpoint for Historical Stock Prices (`idx-web`)

**Files:**
- Create: `idx-web/src/server/utils/price-repo.ts`
- Create: `idx-web/src/server/api/v1/stocks/[ticker]/prices.get.ts`
- Modify: `idx-web/src/server/utils/types.ts`

**Interfaces:**
- `GET /api/v1/stocks/:ticker/prices?range=5y` returns `{ ticker: string, prices: PriceCandle[] }`.

- [ ] **Step 1: Create `price-repo.ts` and `prices.get.ts`**
- [ ] **Step 2: Verify Nuxt production build (`cd idx-web && npm run build`)**
- [ ] **Step 3: Commit Task 4**

```bash
git add idx-web/src/server/
git commit -m "feat(api): add historical stock price API endpoint for charting"
```

---

### Task 5: Dark Terminal UI — Interactive Price & Valuation Bands Chart

**Files:**
- Create: `idx-web/src/components/PriceValuationChart.vue`
- Modify: `idx-web/src/components/TickerFinancialsModal.vue`

**Interfaces:**
- `PriceValuationChart.vue`: Interactive dark terminal chart plotting historical stock price vs. Benjamin Graham Fair Value with Margin of Safety highlight zone.

- [ ] **Step 1: Create `idx-web/src/components/PriceValuationChart.vue`**
- [ ] **Step 2: Integrate into `TickerFinancialsModal.vue`**
- [ ] **Step 3: Verify Nuxt production build (`cd idx-web && npm run build`)**
- [ ] **Step 4: Commit Task 5**

```bash
git add idx-web/src/components/
git commit -m "feat(web): add interactive Price vs Graham Fair Value chart to Ticker 360 modal"
```

---

### Task 6: Authoritative `AGENTS.md` and Makefile Targets

**Files:**
- Modify: `AGENTS.md`
- Modify: `Makefile`

- [ ] **Step 1: Update `Makefile` with `seed-ticker` and `reset-db` targets**
- [ ] **Step 2: Overhaul `AGENTS.md` with complete architecture, tools, formulas, and workflows**
- [ ] **Step 3: Commit Task 6**

```bash
git add AGENTS.md Makefile
git commit -m "docs: overhaul AGENTS.md with complete architecture manual and add seed-ticker Makefile targets"
```

---

### Task 7: Verification & End-to-End Build

- [ ] **Step 1: Run complete Go test suite (`go test -count=1 ./...`)**
- [ ] **Step 2: Build all Go binaries (`go build -o bin/ ./cmd/... ./tools/...`)**
- [ ] **Step 3: Build Nuxt production server (`cd idx-web && npm run build`)**
- [ ] **Step 4: Verify clean git status**
