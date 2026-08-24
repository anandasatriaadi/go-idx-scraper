# Valuation Mean-Reversion Bands & Smart VSA Timing Engine Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement an institutional-grade quantitative timing engine replacing generic moving averages with Historical Valuation Mean-Reversion Bands (P/E & P/B $\pm 1\sigma, \pm 2\sigma$), RSI(14) Bullish Divergence detection, and Volume Spread Analysis (VSA) stopping volume / volume dry-up indicators to generate high-conviction "When to Buy" signals on undervalued IDX stocks.

**Architecture:** 
- In Go (`internal/feature/xbrl/timing.go`): Quantitative algorithms calculate rolling P/E and P/B Standard Deviation Bands ($\text{Mean}, \pm 1\sigma, \pm 2\sigma$), RSI Bullish Divergence, Relative Volume ($RVOL_{20}$), Close Location Value ($CLV$), Volume Dry-Up ($VDU_5$), and synthesize a Smart Timing Score (0 to 100).
- In MongoDB: Persists valuation bands and timing signals in `xbrl_statements` and `stock_prices`.
- In Nuxt 4 / Vue 3: Renders dynamic P/E & P/B valuation band channels, VSA volume footprints, and actionable buy signal badges on the Dark Terminal UI.

**Tech Stack:** Go 1.24+, Nuxt 4 / Vue 3 / SVG Charting, MongoDB v2 Driver, Yahoo Finance daily candles.

## Global Constraints

- Use strict `context: "fresh"` for all subagent dispatches (no session forking).
- Never use `context.Background()` in domain logic; propagate context from callers.
- Wrap errors with `%w`.
- Mathematical rigor: Standard deviation bands must use rolling sample statistics ($\mu \pm 1\sigma, \pm 2\sigma$).
- Smart Timing Score is an integer from 0 to 100 with clear qualitative status tiers.

---

### Task 1: Quantitative Indicators & Valuation Bands Engine (Go Feature)

**Files:**
- Create: `internal/feature/xbrl/timing.go`
- Create: `internal/feature/xbrl/timing_test.go`
- Modify: `internal/feature/xbrl/entity.go`

**Interfaces:**
- `ComputeValuationBands(candles []domain.PriceCandle, eps float64, bvps float64) ValuationBands`
- `ComputeTimingSignals(candles []domain.PriceCandle, bands ValuationBands, eps float64, bvps float64) TimingSignal`

- [ ] **Step 1: Write comprehensive unit tests for Valuation Bands, RSI Divergence, and VSA**
- [ ] **Step 2: Implement `internal/feature/xbrl/timing.go`**
- [ ] **Step 3: Update `internal/feature/xbrl/entity.go` with `ValuationBands` and `TimingSignal` structs**
- [ ] **Step 4: Run unit tests (`go test -v ./internal/feature/xbrl/...`)**
- [ ] **Step 5: Commit Task 1**

```bash
git add internal/feature/xbrl/
git commit -m "feat(timing): implement historical valuation bands and VSA momentum timing engine"
```

---

### Task 2: Valuation Bands & Timing Signal Persistence in Pipeline

**Files:**
- Modify: `cmd/price_updater/main.go`
- Modify: `tools/seed_ticker/main.go`

**Interfaces:**
- Updates `cmd/price_updater` and `tools/seed_ticker` to compute valuation bands and timing signals upon candle ingestion and store in `xbrl_statements`.

- [ ] **Step 1: Update `cmd/price_updater/main.go` to compute timing signals and valuation bands**
- [ ] **Step 2: Update `tools/seed_ticker/main.go` to compute and display timing signals in the CLI report**
- [ ] **Step 3: Test compilation (`make build`)**
- [ ] **Step 4: Commit Task 2**

```bash
git add cmd/price_updater/main.go tools/seed_ticker/main.go
git commit -m "feat(pipeline): integrate valuation bands and timing signals into price updater and seeder"
```

---

### Task 3: Nuxt 4 API Endpoint Updates for Valuation Bands & Timing

**Files:**
- Modify: `idx-web/src/server/utils/types.ts`
- Modify: `idx-web/src/server/utils/xbrl-repo.ts`
- Modify: `idx-web/src/server/api/v1/screener/value.get.ts`

**Interfaces:**
- `GET /api/v1/screener/value?min_timing_score=70` filters stocks with active institutional buy accumulation.
- Exposes `valuation_bands` and `timing_signal` in statement endpoints.

- [ ] **Step 1: Update TypeScript interfaces in `idx-web/src/server/utils/types.ts`**
- [ ] **Step 2: Update `xbrl-repo.ts` and `screener/value.get.ts` to support timing score filters**
- [ ] **Step 3: Verify Nuxt build (`cd idx-web && npm run build`)**
- [ ] **Step 4: Commit Task 3**

```bash
git add idx-web/src/server/
git commit -m "feat(api): expose valuation bands and timing signals in screener and financials endpoints"
```

---

### Task 4: Dark Terminal UI — Valuation Bands Chart & Smart Buy Badges

**Files:**
- Modify: `idx-web/src/components/PriceValuationChart.vue`
- Modify: `idx-web/src/components/ValueScreenerView.vue`
- Modify: `idx-web/src/components/TickerFinancialsModal.vue`

**Interfaces:**
- `PriceValuationChart.vue`: Renders P/E and P/B Standard Deviation Bands ($\pm 1\sigma, \pm 2\sigma$) as dynamic channel overlays with interactive toggles.
- Displays `TimingSignal` badge (e.g. `🟢 P/E -2SD Deep Value + RSI Divergence (Score: 85/100)`).
- `ValueScreenerView.vue`: Adds Timing Score column and 1-click preset `⚡ Actionable Buy Setup (Score >= 70)`.

- [ ] **Step 1: Update `PriceValuationChart.vue` with Valuation Bands overlay and Timing Signal badge**
- [ ] **Step 2: Update `ValueScreenerView.vue` with Timing Score column and filter**
- [ ] **Step 3: Update `TickerFinancialsModal.vue` to display Timing Score summary**
- [ ] **Step 4: Verify Nuxt production build (`cd idx-web && npm run build`)**
- [ ] **Step 5: Commit Task 4**

```bash
git add idx-web/src/components/
git commit -m "feat(web): render historical valuation bands and smart timing signal badges in UI"
```

---

### Task 5: Verification, Full Build & Final Code Review

- [ ] **Step 1: Run complete Go test suite (`go test -count=1 ./...`)**
- [ ] **Step 2: Build all Go binaries (`go build -o bin/ ./cmd/... ./tools/...`)**
- [ ] **Step 3: Run Nuxt production build (`cd idx-web && npm run build`)**
- [ ] **Step 4: Verify clean git status**
