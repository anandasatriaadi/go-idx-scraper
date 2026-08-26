package xbrl

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	domain "github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
)

func TestParseInstanceXML_Mock(t *testing.T) {
	mockXML := `<?xml version="1.0" encoding="UTF-8"?>
<xbrli:xbrl xmlns:xbrli="http://www.xbrl.org/2003/instance"
            xmlns:idx-dei="http://www.idx.co.id/xbrl/taxonomy/2020-01-01/dei"
            xmlns:idx-cor="http://www.idx.co.id/xbrl/taxonomy/2020-01-01/cor">
  <idx-dei:EntityCode>BBRI</idx-dei:EntityCode>
  <idx-dei:EntityName>PT Bank Rakyat Indonesia (Persero) Tbk</idx-dei:EntityName>
  <idx-dei:Sector>Financials</idx-dei:Sector>
  <idx-dei:Industry>Banking</idx-dei:Industry>
  <idx-dei:DescriptionOfPresentationCurrency>Rupiah / IDR</idx-dei:DescriptionOfPresentationCurrency>
  <idx-dei:CurrentPeriodEndDate>2025-09-30</idx-dei:CurrentPeriodEndDate>
  <idx-dei:PeriodOfFinancialStatementsSubmissions>Kuartal III / Third Quarter</idx-dei:PeriodOfFinancialStatementsSubmissions>
  <idx-cor:Assets contextRef="CurrentYearInstant" decimals="-6" unitRef="IDR">1985000000000000</idx-cor:Assets>
  <idx-cor:CashAndCashEquivalents contextRef="CurrentYearInstant" decimals="-6" unitRef="IDR">250000000000000</idx-cor:CashAndCashEquivalents>
  <idx-cor:Liabilities contextRef="CurrentYearInstant" decimals="-6" unitRef="IDR">1650000000000000</idx-cor:Liabilities>
  <idx-cor:Equity contextRef="CurrentYearInstant" decimals="-6" unitRef="IDR">335000000000000</idx-cor:Equity>
  <idx-cor:SalesAndRevenue contextRef="CurrentYearDuration" decimals="-6" unitRef="IDR">150000000000000</idx-cor:SalesAndRevenue>
  <idx-cor:ProfitLoss contextRef="CurrentYearDuration" decimals="-6" unitRef="IDR">45000000000000</idx-cor:ProfitLoss>
</xbrli:xbrl>`

	stmt, err := ParseInstanceXML(strings.NewReader(mockXML))
	if err != nil {
		t.Fatalf("ParseInstanceXML failed: %v", err)
	}

	if stmt.Ticker != "BBRI" {
		t.Errorf("Expected Ticker BBRI, got %s", stmt.Ticker)
	}
	if stmt.CompanyName != "PT Bank Rakyat Indonesia (Persero) Tbk" {
		t.Errorf("Expected CompanyName match, got %s", stmt.CompanyName)
	}
	if stmt.Metadata.Currency != "IDR" {
		t.Errorf("Expected Currency IDR, got %s", stmt.Metadata.Currency)
	}
	if stmt.Period != "Q3" {
		t.Errorf("Expected Period Q3, got %s", stmt.Period)
	}
	if stmt.Core.TotalAssets != 1985000000000000 {
		t.Errorf("Expected TotalAssets 1985000000000000, got %f", stmt.Core.TotalAssets)
	}
	if stmt.Core.NetIncome != 45000000000000 {
		t.Errorf("Expected NetIncome 45000000000000, got %f", stmt.Core.NetIncome)
	}
}

func TestParseInstanceXML_DSSA(t *testing.T) {
	files, _ := filepath.Glob("../../../saham/*DSSA*.zip")
	var stmts []*domain.Statement
	for _, f := range files {
		stmt, err := ParseAnyFiling(f)
		if err != nil {
			t.Logf("err parsing %s: %v", f, err)
			continue
		}
		stmts = append(stmts, stmt)
	}

	sort.Slice(stmts, func(i, j int) bool {
		if stmts[i].Year != stmts[j].Year {
			return stmts[i].Year < stmts[j].Year
		}
		return stmts[i].PeriodEndDate.Before(stmts[j].PeriodEndDate)
	})

	for i, s := range stmts {
		var prior *domain.Statement
		if i > 0 {
			prior = stmts[i-1]
		}
		_ = domain.ComputeValuationAndRatios(s, prior, 1065.0)
	}

	domain.ApplyStockSplitAdjustment(stmts)

	for _, s := range stmts {
		t.Logf("Period: %d %s (Date: %s) -> OpIncome: %.0f, NetIncome: %.0f, ROIC: %.2f%%, ROE: %.2f%%, EPS: %.2f, BVPS: %.2f, Graham: %.2f, MOS: %.2f%%",
			s.Year, s.Period, s.PeriodEndDate.Format("2006-01-02"),
			s.Core.OperatingIncome, s.Core.NetIncome,
			s.ComputedRatios.ROIC*100, s.ComputedRatios.ROE*100,
			s.Valuation.NormalizedEPS, s.Valuation.NormalizedBVPS,
			s.Valuation.GrahamNumber, s.Valuation.MarginOfSafetyPct,
		)
	}
}

func TestParseInstanceXML_PGAS(t *testing.T) {
	files, _ := filepath.Glob("../../../saham/*PGAS*.zip")
	var stmts []*domain.Statement
	for _, f := range files {
		stmt, err := ParseAnyFiling(f)
		if err != nil {
			t.Logf("err parsing %s: %v", f, err)
			continue
		}
		stmts = append(stmts, stmt)
	}

	sort.Slice(stmts, func(i, j int) bool {
		if stmts[i].Year != stmts[j].Year {
			return stmts[i].Year < stmts[j].Year
		}
		return stmts[i].PeriodEndDate.Before(stmts[j].PeriodEndDate)
	})

	for i, s := range stmts {
		var prior *domain.Statement
		if i > 0 {
			prior = stmts[i-1]
		}
		_ = domain.ComputeValuationAndRatios(s, prior, 1500.0) // ~1500 IDR market price for PGAS
	}

	domain.ApplyStockSplitAdjustment(stmts)

	for _, s := range stmts {
		t.Logf("=== PGAS %d %s (Date: %s) === Shares: %.0f, EPS: %.2f, BVPS: %.2f, Graham: %.2f, ROIC: %.2f%%, ROE: %.2f%%, PE: %.2fx, MOS: %.2f%%",
			s.Year, s.Period, s.PeriodEndDate.Format("2006-01-02"),
			s.Core.SharesOutstanding, s.Valuation.NormalizedEPS, s.Valuation.NormalizedBVPS,
			s.Valuation.GrahamNumber, s.ComputedRatios.ROIC*100, s.ComputedRatios.ROE*100,
			s.Valuation.PERatio, s.Valuation.MarginOfSafetyPct,
		)
	}
}

func TestParseInstanceXML_BBRI(t *testing.T) {
	files, _ := filepath.Glob("../../../saham/*BBRI*.zip")
	var stmts []*domain.Statement
	for _, f := range files {
		stmt, err := ParseAnyFiling(f)
		if err != nil {
			t.Logf("err parsing %s: %v", f, err)
			continue
		}
		stmts = append(stmts, stmt)
	}

	sort.Slice(stmts, func(i, j int) bool {
		if stmts[i].Year != stmts[j].Year {
			return stmts[i].Year < stmts[j].Year
		}
		return stmts[i].PeriodEndDate.Before(stmts[j].PeriodEndDate)
	})

	for i, s := range stmts {
		var prior *domain.Statement
		if i > 0 {
			prior = stmts[i-1]
		}
		_ = domain.ComputeValuationAndRatios(s, prior, 4800.0) // ~4800 IDR market price for BBRI
	}

	for _, s := range stmts {
		t.Logf("=== BBRI %d %s (Date: %s) === Sector: %s, CurrentRatio: %.2f, DE: %.2f, AltmanZ: %.2f, Piotroski: %d/9, ROIC: %.2f%%, ROE: %.2f%%, EPS: %.2f, BVPS: %.2f, Graham: %.2f, MOS: %.2f%%",
			s.Year, s.Period, s.PeriodEndDate.Format("2006-01-02"), s.Metadata.Sector,
			s.ComputedRatios.CurrentRatio, s.ComputedRatios.DebtToEquity, s.ComputedRatios.AltmanZScore,
			s.ComputedRatios.PiotroskiFScore, s.ComputedRatios.ROIC*100, s.ComputedRatios.ROE*100,
			s.Valuation.NormalizedEPS, s.Valuation.NormalizedBVPS,
			s.Valuation.GrahamNumber, s.Valuation.MarginOfSafetyPct,
		)
	}
}

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

func TestParseASII_RealFiling(t *testing.T) {
	path := "../../../saham/FinancialStatement-2024-Audit-ASII-instance.zip"
	if _, err := os.Stat(path); err != nil {
		t.Skipf("File not found: %v", err)
	}

	stmt, err := ParseAnyFiling(path)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	t.Logf("ASII 2024 Audit:")
	t.Logf("  Revenue: %f", stmt.Core.Revenue)
	t.Logf("  CostOfRevenue: %f", stmt.Core.CostOfRevenue)
	t.Logf("  GrossProfit: %f", stmt.Core.GrossProfit)
	t.Logf("  OperatingIncome: %f", stmt.Core.OperatingIncome)
	t.Logf("  NetIncome: %f", stmt.Core.NetIncome)
	t.Logf("  NetIncomeParent: %f", stmt.Core.NetIncomeParent)
	t.Logf("  SharesOutstanding: %f", stmt.Core.SharesOutstanding)
	t.Logf("  RoundingMultiplier: %f", stmt.Metadata.RoundingMultiplier)

	// List facts related to revenue or sales
	for k, v := range stmt.Facts {
		if strings.Contains(strings.ToLower(k), "revenue") || strings.Contains(strings.ToLower(k), "sales") || strings.Contains(strings.ToLower(k), "profit") {
			t.Logf("Fact: %s => %+v", k, v)
		}
	}
}

func TestParseAnyFiling_ExcelSample(t *testing.T) {
	samplePath := "saham/FinancialStatement-2025-III-AALI.xlsx"
	if _, err := os.Stat("../../../" + samplePath); err != nil {
		t.Skipf("Skipping excel test: %v", err)
	}

	stmt, err := ParseAnyFiling("../../../" + samplePath)
	if err != nil {
		t.Fatalf("ParseAnyFiling failed: %v", err)
	}

	if stmt.Ticker != "AALI" {
		t.Errorf("Expected Ticker AALI, got %s", stmt.Ticker)
	}
	if stmt.Year != 2025 {
		t.Errorf("Expected Year 2025, got %d", stmt.Year)
	}
	if stmt.Core.TotalAssets <= 0 {
		t.Errorf("Expected positive TotalAssets, got %f", stmt.Core.TotalAssets)
	}
}
