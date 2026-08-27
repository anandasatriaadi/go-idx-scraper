# Single Responsibility Calculation & Parser Subpackages Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor financial calculations into `internal/feature/xbrl/calc/` and XBRL streaming parsers into `internal/infra/xbrl/parser/` with dedicated single-responsibility files for each financial formula and statement line-item category.

**Architecture:** Hexagonal DDD subpackages where `internal/feature/xbrl/calc/` houses pure domain calculation services (`graham.go`, `piotroski.go`, `altman_z.go`, `roic.go`, `profitability.go`, `solvency.go`, `currency.go`, `split.go`, `valuation_bands.go`, `timing.go`, `engine.go`) and `internal/infra/xbrl/parser/` houses low-memory streaming XML adapters (`income_statement.go`, `balance_sheet.go`, `cash_flow.go`, `dei.go`, `shares.go`, `dates.go`, `zip.go`, `parser.go`).

**Tech Stack:** Go 1.24+, `encoding/xml`, `archive/zip`, zap logger, MongoDB v2 driver, Nuxt 4 web app.

## Global Constraints
- `internal/feature/xbrl/calc/` has ZERO imports of `go.mongodb.org` or any external infrastructure adapter.
- Each formula (Graham, Piotroski, Altman Z, ROIC, Margins, Solvency, FX, Split, Valuation Bands, Timing) lives in its own `.go` file.
- Each parser component (DEI, Balance Sheet, Income Statement, Cash Flow, Shares, Dates, ZIP) lives in its own `.go` file.
- All Go test suites (`go test -v -race ./...`), static analysis (`go vet ./...`), CLI builds (`make build`), and web production builds (`npm --prefix idx-web run build`) must pass without errors.

---

### Task 1: Implement Domain Calculation Subpackage (`internal/feature/xbrl/calc/`)

**Files:**
- Create: `internal/feature/xbrl/calc/profitability.go`
- Create: `internal/feature/xbrl/calc/roic.go`
- Create: `internal/feature/xbrl/calc/solvency.go`
- Create: `internal/feature/xbrl/calc/altman_z.go`
- Create: `internal/feature/xbrl/calc/piotroski.go`
- Create: `internal/feature/xbrl/calc/currency.go`
- Create: `internal/feature/xbrl/calc/graham.go`
- Create: `internal/feature/xbrl/calc/split.go`
- Create: `internal/feature/xbrl/calc/valuation_bands.go`
- Create: `internal/feature/xbrl/calc/timing.go`
- Create: `internal/feature/xbrl/calc/engine.go`
- Create: `internal/feature/xbrl/calc/calc_test.go`
- Modify: `internal/feature/xbrl/calculator.go`
- Modify: `internal/feature/xbrl/timing.go`
- Modify: `internal/feature/xbrl/calculator_test.go`

**Interfaces:**
- Consumes: `xbrl.Statement`, `xbrl.CoreFinancials`, `xbrl.ComputedRatios`, `xbrl.ValuationMetrics`, `stock.PriceCandle`
- Produces: `calc.ComputeValuationAndRatios`, `calc.ComputeGrahamFairValue`, `calc.ComputePiotroskiFScore`, `calc.ComputeAltmanZScore`, `calc.ComputeROIC`, `calc.ApplyStockSplitAdjustment`, `calc.ComputeValuationBands`, `calc.ComputeTimingSignals`.

- [ ] **Step 1: Create focused formula files in `internal/feature/xbrl/calc/`**

Create each formula in its own file:
- `profitability.go`: `ComputeProfitability(r *xbrl.ComputedRatios, c *xbrl.CoreFinancials)`
- `roic.go`: `ComputeROIC(c *xbrl.CoreFinancials) float64`
- `solvency.go`: `ComputeSolvency(r *xbrl.ComputedRatios, v *xbrl.ValuationMetrics, c *xbrl.CoreFinancials)`
- `altman_z.go`: `ComputeAltmanZScore(c *xbrl.CoreFinancials) float64`
- `piotroski.go`: `ComputePiotroskiFScore(stmt *xbrl.Statement, priorStmt *xbrl.Statement) int`
- `currency.go`: `NormalizeCurrencyAndShares(stmt *xbrl.Statement, priorStmt *xbrl.Statement) (float64, float64)`
- `graham.go`: `ComputeGrahamFairValue(v *xbrl.ValuationMetrics)` and `ComputeValuationMultiples(v *xbrl.ValuationMetrics, c *xbrl.CoreFinancials, currentPrice, shares, fxRate float64)`
- `split.go`: `ApplyStockSplitAdjustment(statements []*xbrl.Statement)`
- `valuation_bands.go`: `ComputeValuationBands(historicalStmts []*xbrl.Statement) *xbrl.ValuationBands`
- `timing.go`: `ComputeTimingSignals(candles []stock.PriceCandle, latestStmt *xbrl.Statement) *xbrl.TimingSignal`
- `engine.go`: `ComputeValuationAndRatios(stmt *xbrl.Statement, priorStmt *xbrl.Statement, currentStockPrice float64) error`

- [ ] **Step 2: Create unit tests in `internal/feature/xbrl/calc/calc_test.go`**

Test each formula in isolation: Graham, Piotroski, Altman Z, ROIC, Solvency, Margins, Currency normalization, Stock split adjustment, Valuation bands, Timing signals.

- [ ] **Step 3: Update `internal/feature/xbrl/calculator.go` and `timing.go` to delegate to `calc`**

Forward calls to `calc.ComputeValuationAndRatios`, `calc.ApplyStockSplitAdjustment`, `calc.ComputeValuationBands`, `calc.ComputeTimingSignals` to preserve backward compatibility.

- [ ] **Step 4: Run tests in `internal/feature/xbrl/...`**

Run: `go test -v ./internal/feature/xbrl/...`
Expected: PASS

- [ ] **Step 5: Commit calculation subpackage**

```bash
git add internal/feature/xbrl/
git commit -m "refactor(xbrl): decompose calculations into single-responsibility files under calc subpackage"
```

---

### Task 2: Implement Infrastructure Parser Subpackage (`internal/infra/xbrl/parser/`)

**Files:**
- Create: `internal/infra/xbrl/parser/dates.go`
- Create: `internal/infra/xbrl/parser/dei.go`
- Create: `internal/infra/xbrl/parser/balance_sheet.go`
- Create: `internal/infra/xbrl/parser/income_statement.go`
- Create: `internal/infra/xbrl/parser/cash_flow.go`
- Create: `internal/infra/xbrl/parser/shares.go`
- Create: `internal/infra/xbrl/parser/zip.go`
- Create: `internal/infra/xbrl/parser/parser.go`
- Create: `internal/infra/xbrl/parser/parser_test.go`
- Modify: `internal/infra/xbrl/parser.go`
- Modify: `internal/infra/xbrl/parser_test.go`
- Modify: `internal/infra/xbrl/admr_test.go`
- Modify: `internal/infra/xbrl/audit_test.go`

**Interfaces:**
- Consumes: XML reader `io.Reader`, ZIP file path `string`
- Produces: `parser.ParseInstanceXML(r io.Reader) (*xbrl.Statement, error)`, `parser.ParseInstanceZip(zipPath string) (*xbrl.Statement, error)`.

- [ ] **Step 1: Create focused statement parser files in `internal/infra/xbrl/parser/`**

- `dates.go`: `ParseFlexibleDate(val string) (time.Time, error)`
- `dei.go`: `AssignDEIMetadata(s *xbrl.Statement, tag, val string)`
- `balance_sheet.go`: `AssignBalanceSheetMetric(c *xbrl.CoreFinancials, tag string, val float64)` and `FinalizeBalanceSheet(c *xbrl.CoreFinancials)`
- `income_statement.go`: `AssignIncomeStatementMetric(stmt *xbrl.Statement, tag string, val float64)` and `FinalizeIncomeStatement(c *xbrl.CoreFinancials)`
- `cash_flow.go`: `AssignCashFlowMetric(c *xbrl.CoreFinancials, tag string, val float64)` and `FinalizeCashFlow(c *xbrl.CoreFinancials)`
- `shares.go`: `AssignSharesMetric(c *xbrl.CoreFinancials, tag string, val float64)` and `FinalizeSharesOutstanding(s *xbrl.Statement)`
- `zip.go`: `ParseInstanceZip(zipPath string) (*xbrl.Statement, error)`
- `parser.go`: `ParseInstanceXML(r io.Reader) (*xbrl.Statement, error)`

- [ ] **Step 2: Update `internal/infra/xbrl/parser.go` to delegate to `parser/`**

Keep `xbrl.ParseInstanceZip` and `xbrl.ParseInstanceXML` as thin forwards to `parser.ParseInstanceZip` and `parser.ParseInstanceXML`.

- [ ] **Step 3: Run all parser unit and regression tests**

Run: `go test -v ./internal/infra/xbrl/...`
Expected: PASS (all sample filings, ADMR, PGAS, BBRI, ASII tests pass)

- [ ] **Step 4: Commit parser subpackage**

```bash
git add internal/infra/xbrl/
git commit -m "refactor(xbrl): decompose streaming parser into single-responsibility files under parser subpackage"
```

---

### Task 3: Integration Audit & Callers Refactoring

**Files:**
- Modify: `cmd/downloader/main.go`
- Modify: `cmd/xbrl_parser/main.go`
- Modify: `cmd/price_updater/main.go`
- Modify: `tools/seed_ticker/main.go`

**Interfaces:**
- Consumes: `xbrl/calc` and `xbrl/parser` subpackages
- Produces: Verified CLI commands and developer tools directly importing clean subpackages.

- [ ] **Step 1: Check and update imports in CLI commands and tools**

Verify that `cmd/xbrl_parser`, `cmd/downloader`, `cmd/price_updater`, `tools/seed_ticker` use the new clean subpackages or top-level package APIs cleanly.

- [ ] **Step 2: Compile all commands and tools**

Run: `go build ./cmd/... ./tools/...`
Expected: PASS

- [ ] **Step 3: Commit integration updates if any**

```bash
git add cmd/ tools/
git commit -m "refactor(cmd): align CLI tools with calculation and parser subpackages"
```

---

### Task 4: End-to-End Verification, Documentation & Final Build

**Files:**
- Modify: `AGENTS.md`
- Modify: `README.md`

**Interfaces:**
- Consumes: Complete project
- Produces: Verified production builds, zero warnings, and updated documentation.

- [ ] **Step 1: Run all Go tests with race detector**

Run: `go test -v -race ./...`
Expected: PASS

- [ ] **Step 2: Run Go static analysis**

Run: `go vet ./...`
Expected: 0 warnings or diagnostics

- [ ] **Step 3: Build all Go binaries**

Run: `make build`
Expected: All 8 binaries compiled into `bin/`

- [ ] **Step 4: Build Nuxt 4 Web UI**

Run: `npm --prefix idx-web run build` / `make web-build`
Expected: Production server and client bundles built cleanly

- [ ] **Step 5: Update `AGENTS.md` and `README.md`**

Document the new single-responsibility file structure for `internal/feature/xbrl/calc/` and `internal/infra/xbrl/parser/`.

- [ ] **Step 6: Final Commit**

```bash
git add AGENTS.md README.md
git commit -m "docs: update guide with single-responsibility calculation and parser subpackages"
```
