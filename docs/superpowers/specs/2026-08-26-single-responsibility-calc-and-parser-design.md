# Design Specification: Single Responsibility Calculation & Parser Subpackages

## 1. Overview & Objectives

This specification defines the structural decomposition of financial calculations and XBRL streaming parsers in **go-idx-scraper** into single-responsibility files and dedicated subpackages.

### Key Objectives
1. **Single Responsibility Principle (SRP)**: Each mathematical valuation formula (Graham, Piotroski, Altman Z, ROIC, Margins, Solvency, FX normalization, Stock split adjustment, Timing) and each financial statement parser (DEI, Balance Sheet, Income Statement, Cash Flow, Shares, Dates, ZIP) lives in its own dedicated, focused `.go` file.
2. **Dedicated Domain Calculation Subpackage (`internal/feature/xbrl/calc/`)**: Encapsulate all quantitative valuation and financial health calculations in a pure domain calculation package with zero external dependencies.
3. **Dedicated Infrastructure Parser Subpackage (`internal/infra/xbrl/parser/`)**: Encapsulate low-memory XML streaming parsing into a specialized adapter package.
4. **Backward Compatibility**: Maintain exact mathematical results, struct schemas, and interfaces across the entire pipeline (`cmd/*`, `tools/*`, `idx-web`).

---

## 2. Package & File Layout

```
internal/
├── feature/
│   └── xbrl/                          # Core Domain Entities & Repository Port
│       ├── entity.go                  # Statement, CoreFinancials, ComputedRatios, Repository Port
│       ├── entity_test.go             # Entity tests
│       └── calc/                      # Pure Domain Calculation Subpackage
│           ├── engine.go              # ComputeValuationAndRatios orchestrator
│           ├── graham.go              # Graham Fair Value, MOS %, Valuation Multiples (P/E, P/B, P/S, EV/EBIT, EV/EBITDA)
│           ├── piotroski.go           # Piotroski F-Score (9 criteria + baseline fallback)
│           ├── altman_z.go            # Emerging Market Altman Z''-Score
│           ├── roic.go                # Return on Invested Capital (NOPAT / Invested Capital)
│           ├── profitability.go       # Gross Margin %, Operating Margin %, Net Margin %, ROE, ROA
│           ├── solvency.go            # Current Ratio, Quick Ratio, Debt-to-Equity, Net Debt, FCF Conversion %
│           ├── currency.go            # USD/IDR FX Conversion & Rounding Multiplier Normalization
│           ├── split.go               # Stock Split Adjustment across historical time series (ApplyStockSplitAdjustment)
│           ├── valuation_bands.go     # PE & PB Standard Deviation Valuation Bands
│           ├── timing.go              # VSA Stopping Volume, Volume Dry-Up, RSI Divergence
│           ├── engine_test.go         # Full engine integration test
│           ├── graham_test.go         # Graham & Valuation Multiples unit tests
│           ├── piotroski_test.go      # Piotroski F-Score unit tests
│           ├── altman_z_test.go       # Altman Z''-Score unit tests
│           ├── roic_test.go           # ROIC unit tests
│           ├── split_test.go          # Stock split normalization unit tests
│           └── timing_test.go         # Timing & VSA signals unit tests
│
└── infra/
    └── xbrl/
        ├── excel_parser.go            # Excel workbook statement parser adapter
        ├── excel_parser_test.go       # Excel parser tests
        └── parser/                    # Streaming XBRL XML Parser Subpackage
            ├── parser.go              # Streaming xml.NewDecoder orchestrator (ParseInstanceXML)
            ├── zip.go                 # Zip archive unpacker & inline XBRL HTML merger (ParseInstanceZip)
            ├── dei.go                 # Document & Entity Information (/dei) metadata extractor
            ├── income_statement.go    # Revenue, Cost of Revenue, Gross Profit, Operating Income, Net Income taxonomy
            ├── balance_sheet.go       # Assets, Liabilities, Short/Long Debt, Equity, Working Capital taxonomy
            ├── cash_flow.go           # Operating, Investing, Financing Cash Flows, CapEx, Dividends taxonomy
            ├── shares.go              # Shares outstanding fallback taxonomy extractor
            ├── dates.go               # Flexible IDX filing date format parser
            ├── parser_test.go         # XML streaming parser unit & regression tests
            ├── audit_test.go          # Multi-filing parsing audit tests
            └── admr_test.go           # ADMR/ADRO multi-currency regression tests
```

---

## 3. Calculation Subpackage (`internal/feature/xbrl/calc/`)

Each file has a single responsibility:

1. **`graham.go`**:
   - `ComputeGrahamFairValue(v *xbrl.ValuationMetrics)`: Evaluates $\sqrt{22.5 \times \text{EPS} \times \text{BVPS}}$.
   - `ComputeValuationMultiples(v *xbrl.ValuationMetrics, c *xbrl.CoreFinancials, currentPrice float64, shares float64, fxRate float64)`: Evaluates P/E, P/B, P/S, P/FCF, EV/EBIT, EV/EBITDA, Earnings Yield %, and Margin of Safety %.
2. **`piotroski.go`**:
   - `ComputePiotroskiFScore(stmt *xbrl.Statement, priorStmt *xbrl.Statement)`: Calculates the 9 discrete binary criteria (ROA, CFO, Accruals, $\Delta\text{ROA}$, $\Delta\text{Debt}$, $\Delta\text{CurrentRatio}$, $\Delta\text{Shares}$, $\Delta\text{GrossMargin}$, $\Delta\text{AssetTurnover}$) and single-period heuristics.
3. **`altman_z.go`**:
   - `ComputeAltmanZScore(c *xbrl.CoreFinancials)`: Calculates Emerging Market Altman $Z'' = 6.56 X_1 + 3.26 X_2 + 6.72 X_3 + 1.05 X_4$.
4. **`roic.go`**:
   - `ComputeROIC(c *xbrl.CoreFinancials)`: Calculates $\text{NOPAT} / \text{Invested Capital}$ with $22\%$ Indonesian corporate tax rate.
5. **`profitability.go`**:
   - `ComputeProfitability(r *xbrl.ComputedRatios, c *xbrl.CoreFinancials)`: Calculates Gross Margin %, Operating Margin %, Net Margin %, ROE, ROA with parent net income prioritization.
6. **`solvency.go`**:
   - `ComputeSolvency(r *xbrl.ComputedRatios, v *xbrl.ValuationMetrics, c *xbrl.CoreFinancials)`: Calculates Current Ratio, Quick Ratio, Debt-to-Equity, Net Debt, Interest Coverage, and FCF Conversion %.
7. **`currency.go`**:
   - `NormalizeCurrencyAndPerShare(stmt *xbrl.Statement, priorStmt *xbrl.Statement)`: Normalizes USD currency to IDR via exchange rate and handles rounding multiplier scaling.
8. **`split.go`**:
   - `ApplyStockSplitAdjustment(statements []*xbrl.Statement)`: Detects split ratios ($\ge 1.8\times$ or $\le 0.55\times$) and normalizes historical per-share metrics to the latest share basis.
9. **`valuation_bands.go`**:
   - `ComputeValuationBands(historicalStmts []*xbrl.Statement)`: Computes rolling mean and $\pm 1\text{SD}, \pm 2\text{SD}$ PE and PB bands.
10. **`timing.go`**:
    - `ComputeTimingSignals(candles []stock.PriceCandle, latestStmt *xbrl.Statement)`: Evaluates VSA Stopping Volume, Volume Dry-Up (VDU), RSI Bullish Divergence, and overall timing score.
11. **`engine.go`**:
    - `ComputeValuationAndRatios(stmt *xbrl.Statement, priorStmt *xbrl.Statement, currentStockPrice float64) error`: High-level entry point orchestrating all domain calculation functions in sequential order.

---

## 4. Parser Subpackage (`internal/infra/xbrl/parser/`)

Each file has a single responsibility:

1. **`dei.go`**:
   - `AssignDEIMetadata(s *xbrl.Statement, tag, val string)`: Extracts `/dei` metadata (Ticker, Company Name, Currency, Conversion Rate, Reporting Period End Date, Fiscal Year, Rounding Multiplier, Audit Status, Auditor Opinion).
2. **`balance_sheet.go`**:
   - `AssignBalanceSheetMetric(c *xbrl.CoreFinancials, tag string, val float64)`: Maps Assets, Current Assets, Cash & Equivalents, Liabilities, Current Liabilities, Short-Term Debt, Long-Term Debt, Equity, Retained Earnings.
   - `FinalizeBalanceSheet(c *xbrl.CoreFinancials)`: Derives Total Debt and Working Capital.
3. **`income_statement.go`**:
   - `AssignIncomeStatementMetric(stmt *xbrl.Statement, tag string, val float64)`: Maps Revenue, Cost of Revenue, Gross Profit, Operating Income/EBIT, Pre-tax Income, Finance Costs, Net Income, Net Income Attributable to Parent, EPS.
   - `FinalizeIncomeStatement(c *xbrl.CoreFinancials)`: Fallbacks for Net Income Parent, Operating Income (EBIT), and Gross Profit.
4. **`cash_flow.go`**:
   - `AssignCashFlowMetric(c *xbrl.CoreFinancials, tag string, val float64)`: Maps Operating Cash Flow, Investing Cash Flow, Financing Cash Flow, CapEx, Dividends Paid.
   - `FinalizeCashFlow(c *xbrl.CoreFinancials)`: Derives Free Cash Flow ($\text{CFO} - \text{CapEx}$) and EBITDA estimate.
5. **`shares.go`**:
   - `AssignSharesMetric(c *xbrl.CoreFinancials, tag string, val float64)`: Maps shares outstanding tags.
   - `FinalizeSharesOutstanding(s *xbrl.Statement)`: Cascading search across raw FactMap tags for shares outstanding $> 1000$.
6. **`dates.go`**:
   - `ParseFlexibleDate(val string) (time.Time, error)`: Parses IDX date formats.
7. **`zip.go`**:
   - `ParseInstanceZip(zipPath string) (*xbrl.Statement, error)`: Traverses ZIP archive for `.xbrl`, `.xml`, or inline XBRL `.html` files, merges statements, and finalizes core financials.
8. **`parser.go`**:
   - `ParseInstanceXML(r io.Reader) (*xbrl.Statement, error)`: Streaming XML token parser mapping DEI and `/cor` facts to `FactMap` and delegating to statement mappers.

---

## 5. Migration & Integration Strategy

1. Create `internal/feature/xbrl/calc/` and move/refactor `calculator.go` and `timing.go` into dedicated files.
2. In `internal/feature/xbrl/`, expose forwarding aliases or direct package use `calc.ComputeValuationAndRatios` and `calc.ApplyStockSplitAdjustment` so existing callers have a smooth transition.
3. Create `internal/infra/xbrl/parser/` and decompose `internal/infra/xbrl/parser.go` into dedicated files.
4. In `internal/infra/xbrl/`, expose `ParseInstanceZip`, `ParseInstanceXML`, `ParseExcelStatement` delegating to `parser/` and `excel_parser.go`.
5. Update all callers in `cmd/`, `tools/`, and test suites.
6. Verify all test suites pass with `-race`, `go vet` is clean, and `make build` compiles all binaries.
