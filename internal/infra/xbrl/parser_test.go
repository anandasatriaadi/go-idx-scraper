package xbrl

import (
	"os"
	"strings"
	"testing"
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
