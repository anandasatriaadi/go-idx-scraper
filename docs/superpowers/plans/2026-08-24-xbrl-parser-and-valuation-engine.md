# Native Go XBRL Parser, Automated Valuation Engine & Value Screener Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a high-performance native Go XBRL streaming parser that ingests Indonesian Stock Exchange `instance.xbrl` and `instance.zip` financial statements, extracts all 700+ taxonomy facts, normalizes reporting-date foreign exchange rates (USD/IDR), computes deep value-investing metrics (Piotroski F-Score, Altman Z-Score, Owner Earnings, ROIC, Graham Number, DCF Margin of Safety), persists them to MongoDB, and provides a Dark Terminal Value Screener and Ticker 360° Financials UI.

**Architecture:** Hexagonal Domain-Driven Design in Go (`internal/feature/xbrl`, `internal/infra/xbrl`, `internal/infra/db/mongo`). A streaming XML parser (`encoding/xml`) reads raw XBRL instances into structured facts and normalized financial statements. The valuation engine applies quantitative value-investing algorithms and persists to MongoDB (`xbrl_statements`). Nuxt 4 exposes REST endpoints (`/api/v1/stocks/:ticker/financials`, `/api/v1/screener/value`) rendered in the Vue 3 dark terminal dashboard.

**Tech Stack:** Go 1.24+, `encoding/xml`, `archive/zip`, MongoDB v2 Driver (`go.mongodb.org/mongo-driver/v2`), Nuxt 4 / Vue 3, Yahoo Finance API (`USDIDR=X` and `{TICKER}.JK`).

## Global Constraints

- Never use `context.Background()` in domain logic; propagate context from callers.
- Wrap errors with `%w` for error context across layer boundaries.
- Use `zap.Logger` structured logging.
- Store all raw XBRL facts in `facts` map to enable any future formula calculations without re-scraping.
- Target timezone for filing dates: GMT+8 (`time.FixedZone("GMT+8", 8*3600)`).
- Normalize foreign currencies (USD) to IDR using the filing's explicit `ConversionRate` or Yahoo Finance `USDIDR=X` at `period_end_date`.

---

### Task 1: Domain Entities & Repository Ports (Go & TypeScript)

**Files:**
- Create: `internal/feature/xbrl/entity.go`
- Create: `internal/feature/xbrl/entity_test.go`
- Modify: `idx-web/src/server/utils/types.ts`

**Interfaces:**
- Produces: `Statement`, `StatementMetadata`, `CoreFinancials`, `FactValue`, `FactMap`, `ComputedRatios`, `ValuationMetrics`, `Repository`.

- [ ] **Step 1: Write test for XBRL Statement domain entity initialization and properties**

Create `internal/feature/xbrl/entity_test.go`:
```go
package xbrl

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestStatementEntity_Fields(t *testing.T) {
	id := bson.NewObjectID()
	now := time.Now()

	s := &Statement{
		ID:            id,
		Ticker:        "AADI",
		CompanyName:   "PT Adaro Andalan Indonesia Tbk",
		Year:          2026,
		Period:        "Q1",
		PeriodEndDate: now,
		Metadata: StatementMetadata{
			Sector:             "A. Energy",
			Industry:           "A12. Coal",
			Currency:           "USD",
			RoundingMultiplier: 1000,
			AuditStatus:        "Tidak Diaudit / Unaudit",
		},
		Core: CoreFinancials{
			TotalAssets:       5780540000,
			TotalLiabilities:  1999310000,
			TotalEquity:       3781230000,
			Revenue:           1044192000,
			NetIncome:         153768000,
			OperatingCashFlow: 285400000,
			CapEx:             62300000,
			FreeCashFlow:      223100000,
		},
		ComputedRatios: ComputedRatios{
			ROIC:            0.185,
			ROE:             0.198,
			PiotroskiFScore: 8,
			AltmanZScore:    3.45,
		},
		Valuation: ValuationMetrics{
			NormalizedEPS:    790.15,
			NormalizedBVPS:   4624.80,
			GrahamNumber:     9060.50,
			CurrentPrice:     4150.0,
			MarginOfSafetyPct: 54.2,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if s.Ticker != "AADI" {
		t.Errorf("Expected Ticker AADI, got %s", s.Ticker)
	}
	if s.Core.TotalAssets != 5780540000 {
		t.Errorf("Expected TotalAssets 5780540000, got %f", s.Core.TotalAssets)
	}
	if s.ComputedRatios.PiotroskiFScore != 8 {
		t.Errorf("Expected Piotroski F-Score 8, got %d", s.ComputedRatios.PiotroskiFScore)
	}
	if s.Valuation.MarginOfSafetyPct <= 0 {
		t.Errorf("Expected positive MarginOfSafetyPct")
	}
}
```

- [ ] **Step 2: Create `internal/feature/xbrl/entity.go`**

```go
package xbrl

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type FactValue struct {
	Value    float64 `bson:"value" json:"value"`
	Unit     string  `bson:"unit,omitempty" json:"unit,omitempty"`
	Decimals int     `bson:"decimals,omitempty" json:"decimals,omitempty"`
}

type FactMap map[string]map[string]FactValue // [Tag][ContextRef] -> FactValue

type StatementMetadata struct {
	Sector             string  `bson:"sector" json:"sector"`
	Subsector          string  `bson:"subsector" json:"subsector"`
	Industry           string  `bson:"industry" json:"industry"`
	Subindustry        string  `bson:"subindustry" json:"subindustry"`
	AccountingStandard string  `bson:"accounting_standard" json:"accounting_standard"`
	Currency           string  `bson:"currency" json:"currency"`
	RoundingLevel      string  `bson:"rounding_level" json:"rounding_level"`
	RoundingMultiplier float64 `bson:"rounding_multiplier" json:"rounding_multiplier"`
	AuditStatus        string  `bson:"audit_status" json:"audit_status"`
	AuditorOpinion     string  `bson:"auditor_opinion,omitempty" json:"auditor_opinion,omitempty"`
	AuditorName        string  `bson:"auditor_name,omitempty" json:"auditor_name,omitempty"`
	SourceFile         string  `bson:"source_file,omitempty" json:"source_file,omitempty"`
	ConversionRate     float64 `bson:"conversion_rate,omitempty" json:"conversion_rate,omitempty"`
}

type CoreFinancials struct {
	SharesOutstanding  float64 `bson:"shares_outstanding" json:"shares_outstanding"`
	TotalAssets        float64 `bson:"total_assets" json:"total_assets"`
	CashAndEquivalents float64 `bson:"cash_and_equivalents" json:"cash_and_equivalents"`
	CurrentAssets      float64 `bson:"current_assets" json:"current_assets"`
	TotalLiabilities   float64 `bson:"total_liabilities" json:"total_liabilities"`
	CurrentLiabilities float64 `bson:"current_liabilities" json:"current_liabilities"`
	ShortTermDebt      float64 `bson:"short_term_debt" json:"short_term_debt"`
	LongTermDebt       float64 `bson:"long_term_debt" json:"long_term_debt"`
	TotalDebt          float64 `bson:"total_debt" json:"total_debt"`
	TotalEquity        float64 `bson:"total_equity" json:"total_equity"`
	RetainedEarnings   float64 `bson:"retained_earnings" json:"retained_earnings"`

	Revenue            float64 `bson:"revenue" json:"revenue"`
	CostOfRevenue      float64 `bson:"cost_of_revenue" json:"cost_of_revenue"`
	GrossProfit        float64 `bson:"gross_profit" json:"gross_profit"`
	OperatingIncome    float64 `bson:"operating_income" json:"operating_income"`
	FinanceCosts       float64 `bson:"finance_costs" json:"finance_costs"`
	NetIncome          float64 `bson:"net_income" json:"net_income"`
	NetIncomeParent    float64 `bson:"net_income_parent" json:"net_income_parent"`

	OperatingCashFlow  float64 `bson:"operating_cash_flow" json:"operating_cash_flow"`
	CapEx              float64 `bson:"capex" json:"capex"`
	FreeCashFlow       float64 `bson:"free_cash_flow" json:"free_cash_flow"`
	DividendsPaid      float64 `bson:"dividends_paid" json:"dividends_paid"`
}

type ComputedRatios struct {
	ROIC                  float64 `bson:"roic" json:"roic"`
	ROE                   float64 `bson:"roe" json:"roe"`
	ROA                   float64 `bson:"roa" json:"roa"`
	GrossMarginPct        float64 `bson:"gross_margin_pct" json:"gross_margin_pct"`
	OperatingMarginPct    float64 `bson:"operating_margin_pct" json:"operating_margin_pct"`
	NetMarginPct          float64 `bson:"net_margin_pct" json:"net_margin_pct"`
	DebtToEquity          float64 `bson:"debt_to_equity" json:"debt_to_equity"`
	NetDebt               float64 `bson:"net_debt" json:"net_debt"`
	InterestCoverageRatio float64 `bson:"interest_coverage_ratio" json:"interest_coverage_ratio"`
	CurrentRatio          float64 `bson:"current_ratio" json:"current_ratio"`
	FCFConversionPct      float64 `bson:"fcf_conversion_pct" json:"fcf_conversion_pct"`
	PiotroskiFScore       int     `bson:"piotroski_f_score" json:"piotroski_f_score"`
	AltmanZScore          float64 `bson:"altman_z_score" json:"altman_z_score"`
}

type ValuationMetrics struct {
	NormalizedEPS     float64 `bson:"normalized_eps" json:"normalized_eps"`
	NormalizedBVPS    float64 `bson:"normalized_bvps" json:"normalized_bvps"`
	GrahamNumber      float64 `bson:"graham_number" json:"graham_number"`
	DCFFairValue      float64 `bson:"dcf_fair_value" json:"dcf_fair_value"`
	CurrentPrice      float64 `bson:"current_price" json:"current_price"`
	MarginOfSafetyPct float64 `bson:"margin_of_safety_pct" json:"margin_of_safety_pct"`
	PERatio           float64 `bson:"pe_ratio" json:"pe_ratio"`
	PBRatio           float64 `bson:"pb_ratio" json:"pb_ratio"`
}

type Statement struct {
	ID             bson.ObjectID     `bson:"_id,omitempty" json:"id"`
	Ticker         string            `bson:"ticker" json:"ticker"`
	CompanyName    string            `bson:"company_name" json:"company_name"`
	Year           int               `bson:"year" json:"year"`
	Period         string            `bson:"period" json:"period"`
	PeriodType     string            `bson:"period_type" json:"period_type"`
	PeriodEndDate  time.Time         `bson:"period_end_date" json:"period_end_date"`
	StatementType  string            `bson:"statement_type" json:"statement_type"`
	Metadata       StatementMetadata `bson:"metadata" json:"metadata"`
	Core           CoreFinancials    `bson:"core" json:"core"`
	Facts          FactMap           `bson:"facts" json:"facts"`
	ComputedRatios ComputedRatios    `bson:"computed_ratios" json:"computed_ratios"`
	Valuation      ValuationMetrics  `bson:"valuation" json:"valuation"`
	CreatedAt      time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time         `bson:"updated_at" json:"updated_at"`
}

type Repository interface {
	Upsert(ctx context.Context, s *Statement) error
	FindByTickerAndPeriod(ctx context.Context, ticker string, year int, period string) (*Statement, error)
	FindHistoricalByTicker(ctx context.Context, ticker string, limit int) ([]*Statement, error)
	FindAllWithFilter(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*Statement, error)
}
```

- [ ] **Step 3: Update `idx-web/src/server/utils/types.ts` with XBRL interfaces**

Add TypeScript definitions in `idx-web/src/server/utils/types.ts`:
```typescript
export interface FactValue {
  value: number;
  unit?: string;
  decimals?: number;
}

export type FactMap = Record<string, Record<string, FactValue>>;

export interface StatementMetadata {
  sector: string;
  subsector?: string;
  industry: string;
  subindustry?: string;
  accounting_standard: string;
  currency: string;
  rounding_level: string;
  rounding_multiplier: number;
  audit_status: string;
  auditor_opinion?: string;
  conversion_rate?: number;
}

export interface CoreFinancials {
  shares_outstanding: number;
  total_assets: number;
  cash_and_equivalents: number;
  current_assets: number;
  total_liabilities: number;
  current_liabilities: number;
  short_term_debt: number;
  long_term_debt: number;
  total_debt: number;
  total_equity: number;
  retained_earnings: number;
  revenue: number;
  cost_of_revenue: number;
  gross_profit: number;
  operating_income: number;
  finance_costs: number;
  net_income: number;
  net_income_parent: number;
  operating_cash_flow: number;
  capex: number;
  free_cash_flow: number;
  dividends_paid: number;
}

export interface ComputedRatios {
  roic: number;
  roe: number;
  roa: number;
  gross_margin_pct: number;
  operating_margin_pct: number;
  net_margin_pct: number;
  debt_to_equity: number;
  net_debt: number;
  interest_coverage_ratio: number;
  current_ratio: number;
  fcf_conversion_pct: number;
  piotroski_f_score: number;
  altman_z_score: number;
}

export interface ValuationMetrics {
  normalized_eps: number;
  normalized_bvps: number;
  graham_number: number;
  dcf_fair_value: number;
  current_price: number;
  margin_of_safety_pct: number;
  pe_ratio: number;
  pb_ratio: number;
}

export interface XBRLStatement {
  _id?: string;
  id: string;
  ticker: string;
  company_name: string;
  year: number;
  period: string;
  period_type: string;
  period_end_date: string;
  statement_type: string;
  metadata: StatementMetadata;
  core: CoreFinancials;
  facts?: FactMap;
  computed_ratios: ComputedRatios;
  valuation: ValuationMetrics;
  created_at?: string;
  updated_at?: string;
}
```

- [ ] **Step 4: Run unit test to verify domain entities**

Run: `go test -v ./internal/feature/xbrl/...`  
Expected: PASS

- [ ] **Step 5: Commit Task 1**

```bash
git add internal/feature/xbrl/ idx-web/src/server/utils/types.ts
git commit -m "feat(xbrl): define XBRL domain entities, fact maps and repository ports"
```

---

### Task 2: Native Go XBRL XML Streaming Parser (`internal/infra/xbrl`)

**Files:**
- Create: `internal/infra/xbrl/parser.go`
- Create: `internal/infra/xbrl/parser_test.go`

**Interfaces:**
- Produces: `ParseInstanceXML(r io.Reader) (*xbrl.Statement, error)` and `ParseInstanceZip(zipPath string) (*xbrl.Statement, error)`.

- [ ] **Step 1: Write test for `ParseInstanceXML` using sample `instance.xbrl`**

Create `internal/infra/xbrl/parser_test.go`:
```go
package xbrl

import (
	"os"
	"testing"
)

func TestParseInstanceXML_AADI(t *testing.T) {
	// Sample file from extracted instance.zip
	filePath := "/tmp/idx-samples/instance/instance.xbrl"
	f, err := os.Open(filePath)
	if err != nil {
		t.Skipf("Skipping live file test: %v", err)
	}
	defer f.Close()

	stmt, err := ParseInstanceXML(f)
	if err != nil {
		t.Fatalf("ParseInstanceXML failed: %v", err)
	}

	if stmt.Ticker != "AADI" {
		t.Errorf("Expected Ticker AADI, got %s", stmt.Ticker)
	}
	if stmt.CompanyName != "PT Adaro Andalan Indonesia Tbk" {
		t.Errorf("Expected CompanyName PT Adaro Andalan Indonesia Tbk, got %s", stmt.CompanyName)
	}
	if stmt.Metadata.Currency != "USD" {
		t.Errorf("Expected Currency USD, got %s", stmt.Metadata.Currency)
	}
	if stmt.Core.TotalAssets != 5780540000 {
		t.Errorf("Expected TotalAssets 5780540000, got %f", stmt.Core.TotalAssets)
	}
	if stmt.Core.CashAndEquivalents != 914431000 {
		t.Errorf("Expected CashAndEquivalents 914431000, got %f", stmt.Core.CashAndEquivalents)
	}
	if stmt.Core.Revenue != 1044192000 {
		t.Errorf("Expected Revenue 1044192000, got %f", stmt.Core.Revenue)
	}
	if len(stmt.Facts) < 50 {
		t.Errorf("Expected at least 50 extracted raw facts, got %d", len(stmt.Facts))
	}
}
```

- [ ] **Step 2: Create `internal/infra/xbrl/parser.go` with streaming XML reader**

```go
package xbrl

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	domain "github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
)

// ParseInstanceZip opens an instance.zip file and parses the instance.xbrl inside
func ParseInstanceZip(zipPath string) (*domain.Statement, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("opening zip file: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if strings.HasSuffix(strings.ToLower(f.Name), ".xbrl") || strings.HasSuffix(strings.ToLower(f.Name), ".xml") {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("reading file in zip: %w", err)
			}
			defer rc.Close()
			stmt, err := ParseInstanceXML(rc)
			if err != nil {
				return nil, err
			}
			stmt.Metadata.SourceFile = filepath.Base(zipPath)
			return stmt, nil
		}
	}
	return nil, fmt.Errorf("no .xbrl or .xml instance file found in zip archive: %s", zipPath)
}

// ParseInstanceXML parses an XBRL instance stream into a domain Statement
func ParseInstanceXML(r io.Reader) (*domain.Statement, error) {
	decoder := xml.NewDecoder(r)

	stmt := &domain.Statement{
		Metadata: domain.StatementMetadata{
			RoundingMultiplier: 1.0,
		},
		Facts: make(domain.FactMap),
	}

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("xml decoding token: %w", err)
		}

		switch se := token.(type) {
		case xml.StartElement:
			local := se.Name.Local
			space := se.Name.Space

			// Extract DEI Metadata
			if strings.Contains(space, "/dei") {
				var textVal string
				if err := decoder.DecodeElement(&textVal, &se); err == nil {
					assignDEIMetadata(stmt, local, strings.TrimSpace(textVal))
				}
				continue
			}

			// Extract Core Financial Facts
			if strings.Contains(space, "/cor") {
				var contextRef, unitRef, decimalsStr, scaleStr string
				var isNil bool

				for _, attr := range se.Attr {
					switch attr.Name.Local {
					case "contextRef":
						contextRef = attr.Value
					case "unitRef":
						unitRef = attr.Value
					case "decimals":
						decimalsStr = attr.Value
					case "scale":
						scaleStr = attr.Value
					case "nil":
						isNil = (attr.Value == "true")
					}
				}

				var rawVal string
				if err := decoder.DecodeElement(&rawVal, &se); err == nil {
					rawVal = strings.TrimSpace(rawVal)
					if rawVal != "" && !isNil {
						numVal, err := parseNumericValue(rawVal)
						if err == nil {
							dec, _ := strconv.Atoi(decimalsStr)
							if stmt.Facts[local] == nil {
								stmt.Facts[local] = make(map[string]domain.FactValue)
							}
							stmt.Facts[local][contextRef] = domain.FactValue{
								Value:    numVal,
								Unit:     unitRef,
								Decimals: dec,
							}

							// Populate core financials for primary contexts
							if contextRef == "CurrentYearInstant" || contextRef == "CurrentYearDuration" {
								assignCoreMetric(&stmt.Core, local, numVal)
							}
						}
					}
				}
			}
		}
	}

	// Post-processing: Calculate derived metrics
	finalizeCoreFinancials(stmt)

	return stmt, nil
}

func parseNumericValue(s string) (float64, error) {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, " ", "")
	return strconv.ParseFloat(s, 64)
}

func assignDEIMetadata(s *domain.Statement, tag, val string) {
	switch tag {
	case "EntityCode":
		s.Ticker = val
	case "EntityName":
		s.CompanyName = val
	case "Sector":
		s.Metadata.Sector = val
	case "Subsector":
		s.Metadata.Subsector = val
	case "Industry":
		s.Metadata.Industry = val
	case "Subindustry":
		s.Metadata.Subindustry = val
	case "DescriptionOfPresentationCurrency":
		if strings.Contains(val, "USD") || strings.Contains(val, "Dollar") {
			s.Metadata.Currency = "USD"
		} else {
			s.Metadata.Currency = "IDR"
		}
	case "LevelOfRoundingUsedInFinancialStatements":
		s.Metadata.RoundingLevel = val
		if strings.Contains(val, "Thousand") || strings.Contains(val, "Ribuan") {
			s.Metadata.RoundingMultiplier = 1000
		} else if strings.Contains(val, "Million") || strings.Contains(val, "Jutaan") {
			s.Metadata.RoundingMultiplier = 1000000
		} else if strings.Contains(val, "Billion") || strings.Contains(val, "Miliaran") {
			s.Metadata.RoundingMultiplier = 1000000000
		}
	case "ConversionRateAtReportingDateIfPresentationCurrencyIsOtherThanRupiah":
		rate, _ := parseNumericValue(val)
		s.Metadata.ConversionRate = rate
	case "CurrentPeriodEndDate":
		t, err := time.Parse("2006-01-02", val)
		if err == nil {
			s.PeriodEndDate = t
			s.Year = t.Year()
		}
	case "PeriodOfFinancialStatementsSubmissions":
		s.PeriodType = val
		if strings.Contains(val, "Kuartal I") || strings.Contains(val, "First") {
			s.Period = "Q1"
		} else if strings.Contains(val, "Kuartal II") || strings.Contains(val, "Second") {
			s.Period = "Q2"
		} else if strings.Contains(val, "Kuartal III") || strings.Contains(val, "Third") {
			s.Period = "Q3"
		} else {
			s.Period = "FY"
		}
	case "TypeOfReportOnFinancialStatements":
		s.Metadata.AuditStatus = val
	case "TypeOfAuditorsOpinion":
		s.Metadata.AuditorOpinion = val
	}
}

func assignCoreMetric(c *domain.CoreFinancials, tag string, val float64) {
	switch tag {
	case "Assets":
		c.TotalAssets = val
	case "CashAndCashEquivalents":
		c.CashAndEquivalents = val
	case "CurrentAssets":
		c.CurrentAssets = val
	case "Liabilities":
		c.TotalLiabilities = val
	case "CurrentLiabilities":
		c.CurrentLiabilities = val
	case "ShortTermBankLoans":
		c.ShortTermDebt = val
	case "LongTermBankLoans":
		c.LongTermDebt = val
	case "Equity", "TotalEquity":
		c.TotalEquity = val
	case "RetainedEarningsUnappropriated":
		c.RetainedEarnings = val
	case "SalesAndRevenue", "Revenues":
		c.Revenue = val
	case "CostOfSalesAndRevenue":
		c.CostOfRevenue = val
	case "GrossProfit":
		c.GrossProfit = val
	case "OperatingIncomeExpense":
		c.OperatingIncome = val
	case "FinanceCosts":
		c.FinanceCosts = val
	case "ProfitLoss":
		c.NetIncome = val
	case "ProfitLossAttributableToOwnersOfParentEntity":
		c.NetIncomeParent = val
	case "NetCashFlowsFromUsedInOperatingActivities":
		c.OperatingCashFlow = val
	case "PaymentsForPropertyPlantEquipment":
		c.CapEx = val
	case "WeightedAverageShares", "NumberOfIssuedAndFullyPaidShares":
		c.SharesOutstanding = val
	}
}

func finalizeCoreFinancials(s *domain.Statement) {
	c := &s.Core
	if c.TotalDebt == 0 {
		c.TotalDebt = c.ShortTermDebt + c.LongTermDebt
	}
	if c.FreeCashFlow == 0 && c.OperatingCashFlow != 0 {
		c.FreeCashFlow = c.OperatingCashFlow - c.CapEx
	}
}
```

- [ ] **Step 3: Run parser tests**

Run: `go test -v ./internal/infra/xbrl/...`  
Expected: PASS

- [ ] **Step 4: Commit Task 2**

```bash
git add internal/infra/xbrl/
git commit -m "feat(xbrl): implement native Go XBRL streaming parser for instance XML and ZIP archives"
```

---

### Task 3: Forensic Valuation & Currency Engine (`internal/feature/xbrl`)

**Files:**
- Create: `internal/feature/xbrl/calculator.go`
- Create: `internal/feature/xbrl/calculator_test.go`

**Interfaces:**
- Produces: `ComputeValuationAndRatios(stmt *Statement, priorStmt *Statement, currentStockPrice float64) error`.

- [ ] **Step 1: Write unit tests for Piotroski F-Score, Altman Z, ROIC, and Graham Number**

Create `internal/feature/xbrl/calculator_test.go`:
```go
package xbrl

import (
	"testing"
)

func TestCalculator_ForensicsAndValuation(t *testing.T) {
	curr := &Statement{
		Ticker: "AADI",
		Metadata: StatementMetadata{
			Currency:       "USD",
			ConversionRate: 15400.0,
		},
		Core: CoreFinancials{
			SharesOutstanding:  3000000000,
			TotalAssets:        5780540000,
			CashAndEquivalents: 914431000,
			CurrentAssets:      1780200000,
			TotalLiabilities:   1999310000,
			CurrentLiabilities: 650200000,
			TotalDebt:          570000000,
			TotalEquity:        3781230000,
			RetainedEarnings:   2450000000,
			Revenue:            1044192000,
			GrossProfit:        257553000,
			OperatingIncome:    210400000,
			FinanceCosts:       15200000,
			NetIncome:          153768000,
			OperatingCashFlow:  285400000,
			CapEx:              62300000,
			FreeCashFlow:       223100000,
		},
	}

	err := ComputeValuationAndRatios(curr, nil, 4150.0)
	if err != nil {
		t.Fatalf("ComputeValuationAndRatios failed: %v", err)
	}

	if curr.ComputedRatios.ROIC <= 0 {
		t.Errorf("Expected positive ROIC, got %f", curr.ComputedRatios.ROIC)
	}
	if curr.ComputedRatios.AltmanZScore <= 0 {
		t.Errorf("Expected positive Altman Z-Score, got %f", curr.ComputedRatios.AltmanZScore)
	}
	if curr.Valuation.NormalizedEPS <= 0 {
		t.Errorf("Expected positive Normalized EPS, got %f", curr.Valuation.NormalizedEPS)
	}
	if curr.Valuation.GrahamNumber <= 0 {
		t.Errorf("Expected positive Graham Number, got %f", curr.Valuation.GrahamNumber)
	}
	if curr.Valuation.MarginOfSafetyPct <= 0 {
		t.Errorf("Expected positive Margin of Safety, got %f", curr.Valuation.MarginOfSafetyPct)
	}
}
```

- [ ] **Step 2: Implement `internal/feature/xbrl/calculator.go`**

```go
package xbrl

import (
	"math"
)

// ComputeValuationAndRatios calculates ROIC, Piotroski, Altman Z, Graham Fair Value, and MOS
func ComputeValuationAndRatios(stmt *Statement, priorStmt *Statement, currentStockPrice float64) error {
	c := &stmt.Core
	r := &stmt.ComputedRatios
	v := &stmt.Valuation

	// 1. Profitability & Margins
	if c.Revenue > 0 {
		r.GrossMarginPct = (c.GrossProfit / c.Revenue) * 100
		r.OperatingMarginPct = (c.OperatingIncome / c.Revenue) * 100
		r.NetMarginPct = (c.NetIncome / c.Revenue) * 100
	}
	if c.TotalEquity > 0 {
		r.ROE = c.NetIncome / c.TotalEquity
	}
	if c.TotalAssets > 0 {
		r.ROA = c.NetIncome / c.TotalAssets
	}

	// 2. Return on Invested Capital (ROIC)
	investedCapital := (c.TotalEquity + c.TotalDebt) - c.CashAndEquivalents
	nopat := c.OperatingIncome * (1 - 0.22) // 22% Indonesian standard corporate tax rate
	if investedCapital > 0 {
		r.ROIC = nopat / investedCapital
	}

	// 3. Leverage & Solvency
	if c.TotalEquity > 0 {
		r.DebtToEquity = c.TotalDebt / c.TotalEquity
	}
	r.NetDebt = c.TotalDebt - c.CashAndEquivalents
	if c.FinanceCosts > 0 {
		r.InterestCoverageRatio = c.OperatingIncome / c.FinanceCosts
	}
	if c.CurrentLiabilities > 0 {
		r.CurrentRatio = c.CurrentAssets / c.CurrentLiabilities
	}
	if c.NetIncome > 0 {
		r.FCFConversionPct = (c.FreeCashFlow / c.NetIncome) * 100
	}

	// 4. Emerging Market Altman Z''-Score
	if c.TotalAssets > 0 && c.TotalLiabilities > 0 {
		workingCapital := c.CurrentAssets - c.CurrentLiabilities
		x1 := workingCapital / c.TotalAssets
		x2 := c.RetainedEarnings / c.TotalAssets
		x3 := c.OperatingIncome / c.TotalAssets
		x4 := c.TotalEquity / c.TotalLiabilities
		r.AltmanZScore = (6.56 * x1) + (3.26 * x2) + (6.72 * x3) + (1.05 * x4)
	}

	// 5. Piotroski F-Score (0 to 9)
	fScore := 0
	if r.ROA > 0 {
		fScore++
	}
	if c.OperatingCashFlow > 0 {
		fScore++
	}
	if c.OperatingCashFlow > c.NetIncome {
		fScore++
	}
	if priorStmt != nil {
		if r.ROA > priorStmt.ComputedRatios.ROA {
			fScore++
		}
		if c.LongTermDebt < priorStmt.Core.LongTermDebt {
			fScore++
		}
		if r.CurrentRatio > priorStmt.ComputedRatios.CurrentRatio {
			fScore++
		}
		if c.SharesOutstanding <= priorStmt.Core.SharesOutstanding {
			fScore++
		}
		if r.GrossMarginPct > priorStmt.ComputedRatios.GrossMarginPct {
			fScore++
		}
		assetTurnoverCurr := c.Revenue / c.TotalAssets
		assetTurnoverPrior := priorStmt.Core.Revenue / priorStmt.Core.TotalAssets
		if assetTurnoverCurr > assetTurnoverPrior {
			fScore++
		}
	} else {
		// Single period baseline heuristics
		if r.CurrentRatio > 1.2 {
			fScore++
		}
		if r.DebtToEquity < 1.0 {
			fScore++
		}
		if r.GrossMarginPct > 20.0 {
			fScore++
		}
		fScore += 2 // Neutral prior credit
	}
	if fScore > 9 {
		fScore = 9
	}
	r.PiotroskiFScore = fScore

	// 6. Currency Normalization (USD -> IDR)
	fxRate := 1.0
	if stmt.Metadata.Currency == "USD" {
		if stmt.Metadata.ConversionRate > 0 {
			fxRate = stmt.Metadata.ConversionRate
		} else {
			fxRate = 16000.0 // Conservative default if unpopulated
		}
	}

	shares := c.SharesOutstanding
	if shares <= 0 {
		shares = 1.0
	}

	// Normalized per-share values in IDR
	v.NormalizedEPS = (c.NetIncome * fxRate) / shares
	v.NormalizedBVPS = (c.TotalEquity * fxRate) / shares

	// 7. Benjamin Graham Fair Value Formula
	if v.NormalizedEPS > 0 && v.NormalizedBVPS > 0 {
		v.GrahamNumber = math.Sqrt(22.5 * v.NormalizedEPS * v.NormalizedBVPS)
	}

	// 8. Valuation Multiples & Margin of Safety
	v.CurrentPrice = currentStockPrice
	if currentStockPrice > 0 {
		if v.NormalizedEPS > 0 {
			v.PERatio = currentStockPrice / v.NormalizedEPS
		}
		if v.NormalizedBVPS > 0 {
			v.PBRatio = currentStockPrice / v.NormalizedBVPS
		}
		if v.GrahamNumber > 0 {
			v.MarginOfSafetyPct = ((v.GrahamNumber - currentStockPrice) / v.GrahamNumber) * 100
		}
	}

	return nil
}
```

- [ ] **Step 3: Run calculator unit test**

Run: `go test -v ./internal/feature/xbrl/...`  
Expected: PASS

- [ ] **Step 4: Commit Task 3**

```bash
git add internal/feature/xbrl/
git commit -m "feat(xbrl): implement forensic metrics calculator (ROIC, Piotroski, Altman Z, Graham Fair Value)"
```

---

### Task 4: MongoDB Repository Adapter & Parser CLI (`internal/infra/db/mongo`, `cmd/xbrl_parser`)

**Files:**
- Create: `internal/infra/db/mongo/xbrl_repo.go`
- Create: `internal/infra/db/mongo/xbrl_repo_test.go`
- Create: `cmd/xbrl_parser/main.go`

**Interfaces:**
- Produces: `XBRLRepository` implementation and standalone `xbrl_parser` CLI tool.

- [ ] **Step 1: Create `internal/infra/db/mongo/xbrl_repo.go`**

```go
package mongo

import (
	"context"
	"log"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type XBRLRepository struct {
	collection *mongo.Collection
}

func NewXBRLRepository(db *mongo.Database) xbrl.Repository {
	repo := &XBRLRepository{
		collection: db.Collection("xbrl_statements"),
	}
	if err := repo.ensureIndexes(context.Background()); err != nil {
		log.Printf("warn: failed to ensure xbrl indexes: %v", err)
	}
	return repo
}

func (r *XBRLRepository) ensureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "ticker", Value: 1},
				{Key: "year", Value: -1},
				{Key: "period", Value: -1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "valuation.margin_of_safety_pct", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "computed_ratios.roic", Value: -1},
			},
		},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func (r *XBRLRepository) Upsert(ctx context.Context, s *xbrl.Statement) error {
	now := time.Now()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	s.UpdatedAt = now

	filter := bson.M{
		"ticker": s.Ticker,
		"year":   s.Year,
		"period": s.Period,
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, bson.M{"$set": s}, opts)
	return err
}

func (r *XBRLRepository) FindByTickerAndPeriod(ctx context.Context, ticker string, year int, period string) (*xbrl.Statement, error) {
	filter := bson.M{
		"ticker": ticker,
		"year":   year,
		"period": period,
	}
	var stmt xbrl.Statement
	err := r.collection.FindOne(ctx, filter).Decode(&stmt)
	if err != nil {
		return nil, err
	}
	return &stmt, nil
}

func (r *XBRLRepository) FindHistoricalByTicker(ctx context.Context, ticker string, limit int) ([]*xbrl.Statement, error) {
	filter := bson.M{"ticker": ticker}
	opts := options.Find().SetSort(bson.D{{Key: "year", Value: -1}, {Key: "period", Value: -1}}).SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*xbrl.Statement
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *XBRLRepository) FindAllWithFilter(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*xbrl.Statement, error) {
	cursor, err := r.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*xbrl.Statement
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}
```

- [ ] **Step 2: Create `cmd/xbrl_parser/main.go`**

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	domain "github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
	"github.com/anandasatriaadi/go-idx-scraper/internal/helper"
	"github.com/anandasatriaadi/go-idx-scraper/internal/infra/db/mongo"
	infra "github.com/anandasatriaadi/go-idx-scraper/internal/infra/xbrl"
	"go.uber.org/zap"
)

func main() {
	var configPath string
	var targetDir string
	flag.StringVar(&configPath, "config", "config/config.yml", "Path to config file")
	flag.StringVar(&targetDir, "dir", "saham", "Directory containing XBRL zip or xml files")
	flag.Parse()

	logger, err := helper.NewLogger("xbrl_parser")
	if err != nil {
		log.Fatalf("logger init: %v", err)
	}
	defer logger.Sync()

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Fatal("loading config", zap.Error(err))
	}

	dbClient, err := mongo.NewClient(logger)
	if err != nil {
		logger.Fatal("mongodb connect", zap.Error(err))
	}
	db := dbClient.Database(cfg.Database.DbName)
	repo := mongo.NewXBRLRepository(db)

	ctx := context.Background()

	files, err := os.ReadDir(targetDir)
	if err != nil {
		logger.Fatal("reading dir", zap.Error(err))
	}

	parsedCount := 0
	for _, f := range files {
		if strings.HasSuffix(strings.ToLower(f.Name()), ".zip") || strings.HasSuffix(strings.ToLower(f.Name()), ".xbrl") {
			fullPath := filepath.Join(targetDir, f.Name())
			logger.Info("Processing filing", zap.String("file", f.Name()))

			var stmt *domain.Statement
			if strings.HasSuffix(strings.ToLower(f.Name()), ".zip") {
				stmt, err = infra.ParseInstanceZip(fullPath)
			} else {
				fileHandle, oErr := os.Open(fullPath)
				if oErr == nil {
					stmt, err = infra.ParseInstanceXML(fileHandle)
					fileHandle.Close()
				} else {
					err = oErr
				}
			}

			if err != nil {
				logger.Warn("Failed to parse XBRL filing", zap.String("file", f.Name()), zap.Error(err))
				continue
			}

			if err := domain.ComputeValuationAndRatios(stmt, nil, 0); err != nil {
				logger.Warn("Valuation calculation failed", zap.String("ticker", stmt.Ticker), zap.Error(err))
			}

			if err := repo.Upsert(ctx, stmt); err != nil {
				logger.Error("Failed to upsert to MongoDB", zap.String("ticker", stmt.Ticker), zap.Error(err))
			} else {
				logger.Info("Successfully parsed & saved XBRL statement", zap.String("ticker", stmt.Ticker), zap.Int("year", stmt.Year), zap.String("period", stmt.Period))
				parsedCount++
			}
		}
	}

	logger.Info("XBRL parsing completed", zap.Int("total_parsed", parsedCount))
	fmt.Printf("Parsed %d XBRL statements successfully.\n", parsedCount)
}
```

- [ ] **Step 3: Run compilation and test**

Run: `go build -o bin/xbrl_parser ./cmd/xbrl_parser`  
Run: `go test -v ./internal/infra/db/mongo/...`  
Expected: PASS

- [ ] **Step 4: Commit Task 4**

```bash
git add internal/infra/db/mongo/xbrl_repo.go cmd/xbrl_parser/
git commit -m "feat(xbrl): implement MongoDB XBRL repository adapter and standalone parser CLI"
```

---

### Task 5: Downloader XBRL Support (`internal/feature/finreport`, `cmd/downloader`)

**Files:**
- Modify: `internal/feature/finreport/service.go`
- Modify: `cmd/downloader/main.go`

**Interfaces:**
- Produces: `ConstructXBRLReportURL(year, periodString, issuerCode, fileType)` downloading `instance.zip` / `inlineXBRL.zip` directly.

- [ ] **Step 1: Add `ConstructXBRLReportURL` in `internal/feature/finreport/service.go`**

```go
func (s *Service) ConstructXBRLReportURL(year int, periodString, issuerCode string, fileType string) string {
	var modePeriod string
	if periodString == "Tahunan" {
		modePeriod = "Audit"
	} else {
		modePeriod = "TW" + romanToNumeral(periodString)
	}
	if fileType == "" {
		fileType = "instance.zip"
	}
	url := fmt.Sprintf("https://www.idx.co.id/Portals/0/StaticData/ListedCompanies/Corporate_Actions/New_Info_JSX/Jenis_Informasi/01_Laporan_Keuangan/02_Soft_Copy_Laporan_Keuangan//Laporan%%20Keuangan%%20Tahun%%20%d/%s/%s/%s",
		year, modePeriod, issuerCode, fileType)
	return url
}
```

- [ ] **Step 2: Commit Task 5**

```bash
git add internal/feature/finreport/service.go
git commit -m "feat(downloader): support XBRL instance.zip and inlineXBRL.zip URL construction"
```

---

### Task 6: Nuxt 4 API Server & Screener Endpoints (`idx-web`)

**Files:**
- Create: `idx-web/src/server/utils/xbrl-repo.ts`
- Create: `idx-web/src/server/api/v1/stocks/[ticker]/financials.get.ts`
- Create: `idx-web/src/server/api/v1/screener/value.get.ts`

- [ ] **Step 1: Create `idx-web/src/server/utils/xbrl-repo.ts`**

```typescript
import { Db } from 'mongodb'
import { XBRLStatement } from './types'

export class XBRLStatementRepository {
  private collection: any

  constructor(db: Db) {
    this.collection = db.collection('xbrl_statements')
  }

  async findByTicker(ticker: string, limit: number = 8): Promise<XBRLStatement[]> {
    const docs = await this.collection
      .find({ ticker: ticker.toUpperCase().trim() })
      .sort({ year: -1, period: -1 })
      .limit(limit)
      .toArray()
    return docs.map((d: any) => ({ ...d, id: d._id.toString() }))
  }

  async findValueScreener(filters: {
    minRoic?: number
    minFScore?: number
    minMosPct?: number
    maxDebtEquity?: number
    sector?: string
    limit?: number
    skip?: number
  }): Promise<{ statements: XBRLStatement[]; total: number }> {
    const query: any = {}
    if (filters.minRoic !== undefined) query['computed_ratios.roic'] = { $gte: filters.minRoic }
    if (filters.minFScore !== undefined) query['computed_ratios.piotroski_f_score'] = { $gte: filters.minFScore }
    if (filters.minMosPct !== undefined) query['valuation.margin_of_safety_pct'] = { $gte: filters.minMosPct }
    if (filters.maxDebtEquity !== undefined) query['computed_ratios.debt_to_equity'] = { $lte: filters.maxDebtEquity }
    if (filters.sector) query['metadata.sector'] = filters.sector

    const limit = filters.limit || 50
    const skip = filters.skip || 0

    const [statements, total] = await Promise.all([
      this.collection.find(query).sort({ 'valuation.margin_of_safety_pct': -1 }).skip(skip).limit(limit).toArray(),
      this.collection.countDocuments(query)
    ])

    return {
      statements: statements.map((d: any) => ({ ...d, id: d._id.toString() })),
      total
    }
  }
}
```

- [ ] **Step 2: Create API Endpoints**

`idx-web/src/server/api/v1/stocks/[ticker]/financials.get.ts`:
```typescript
import { defineEventHandler } from 'h3'
import { useDb } from '../../../../plugins/mongodb'
import { XBRLStatementRepository } from '../../../../utils/xbrl-repo'

export default defineEventHandler(async (event) => {
  const ticker = event.context.params?.ticker
  if (!ticker) {
    throw createError({ statusCode: 400, statusMessage: 'Ticker is required' })
  }
  const db = useDb()
  const repo = new XBRLStatementRepository(db)
  return await repo.findByTicker(ticker)
})
```

`idx-web/src/server/api/v1/screener/value.get.ts`:
```typescript
import { defineEventHandler, getQuery } from 'h3'
import { useDb } from '../../../plugins/mongodb'
import { XBRLStatementRepository } from '../../../utils/xbrl-repo'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const db = useDb()
  const repo = new XBRLStatementRepository(db)

  return await repo.findValueScreener({
    minRoic: query.min_roic ? parseFloat(String(query.min_roic)) : undefined,
    minFScore: query.min_f_score ? parseInt(String(query.min_f_score), 10) : undefined,
    minMosPct: query.min_mos ? parseFloat(String(query.min_mos)) : undefined,
    maxDebtEquity: query.max_de ? parseFloat(String(query.max_de)) : undefined,
    sector: query.sector ? String(query.sector) : undefined,
    limit: query.limit ? parseInt(String(query.limit), 10) : 50,
    skip: query.skip ? parseInt(String(query.skip), 10) : 0
  })
})
```

- [ ] **Step 3: Commit Task 6**

```bash
git add idx-web/src/server/
git commit -m "feat(api): add XBRL financial statement and value screener API routes"
```

---

### Task 7: Dark Terminal UI — Value Screener & Ticker 360° View

**Files:**
- Create: `idx-web/src/components/ValueScreenerView.vue`
- Create: `idx-web/src/components/TickerFinancialsModal.vue`
- Modify: `idx-web/src/components/Navbar.vue`
- Modify: `idx-web/src/pages/index.vue`

- [ ] **Step 1: Create `idx-web/src/components/ValueScreenerView.vue`**
- [ ] **Step 2: Create `idx-web/src/components/TickerFinancialsModal.vue`**
- [ ] **Step 3: Add `Value Screener` tab to `Navbar.vue` and `index.vue`**
- [ ] **Step 4: Verify build with `cd idx-web && npm run build`**
- [ ] **Step 5: Commit Task 7**

---

### Task 8: End-to-End Build & Verification

- [ ] **Step 1: Run complete Go test suite (`go test -count=1 ./...`)**
- [ ] **Step 2: Run Nuxt production build (`cd idx-web && npm run build`)**
- [ ] **Step 3: Build all binaries (`go build -o bin/ ./cmd/...`)**
- [ ] **Step 4: Verify clean git status**
