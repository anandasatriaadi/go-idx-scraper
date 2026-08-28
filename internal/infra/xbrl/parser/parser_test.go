package parser

import (
	"strings"
	"testing"
)

func TestParseInstanceXML_Basic(t *testing.T) {
	sampleXML := `<?xml version="1.0" encoding="utf-8"?>
<xbrli:xbrl xmlns:xbrli="http://www.xbrl.org/2003/instance" xmlns:idx-dei="http://www.idx.co.id/dei" xmlns:idx-cor="http://www.idx.co.id/cor">
	<idx-dei:EntityCode contextRef="CurrentYearDuration">BBRI</idx-dei:EntityCode>
	<idx-dei:EntityName contextRef="CurrentYearDuration">PT Bank Rakyat Indonesia (Persero) Tbk</idx-dei:EntityName>
	<idx-dei:DescriptionOfPresentationCurrency contextRef="CurrentYearDuration">Rupiah</idx-dei:DescriptionOfPresentationCurrency>
	<idx-dei:CurrentPeriodEndDate contextRef="CurrentYearDuration">2024-12-31</idx-dei:CurrentPeriodEndDate>
	<idx-dei:PeriodOfFinancialStatementsSubmissions contextRef="CurrentYearDuration">Kuartal IV</idx-dei:PeriodOfFinancialStatementsSubmissions>
	<idx-cor:Assets contextRef="CurrentYearInstant">1965000000000</idx-cor:Assets>
	<idx-cor:Liabilities contextRef="CurrentYearInstant">1650000000000</idx-cor:Liabilities>
	<idx-cor:Equity contextRef="CurrentYearInstant">315000000000</idx-cor:Equity>
	<idx-cor:SalesAndRevenue contextRef="CurrentYearDuration">200000000000</idx-cor:SalesAndRevenue>
	<idx-cor:CostOfSalesAndRevenue contextRef="CurrentYearDuration">80000000000</idx-cor:CostOfSalesAndRevenue>
	<idx-cor:GrossProfit contextRef="CurrentYearDuration">120000000000</idx-cor:GrossProfit>
	<idx-cor:OperatingIncome contextRef="CurrentYearDuration">75000000000</idx-cor:OperatingIncome>
	<idx-cor:ProfitLossAttributableToOwnersOfParentEntity contextRef="CurrentYearDuration">60000000000</idx-cor:ProfitLossAttributableToOwnersOfParentEntity>
	<idx-cor:NetCashFlowsReceivedFromUsedInOperatingActivities contextRef="CurrentYearDuration">65000000000</idx-cor:NetCashFlowsReceivedFromUsedInOperatingActivities>
	<idx-cor:PaymentsForPropertyPlantEquipment contextRef="CurrentYearDuration">10000000000</idx-cor:PaymentsForPropertyPlantEquipment>
	<idx-cor:NumberOfIssuedAndFullyPaidShares contextRef="CurrentYearInstant">151559000000</idx-cor:NumberOfIssuedAndFullyPaidShares>
</xbrli:xbrl>`

	stmt, err := ParseInstanceXML(strings.NewReader(sampleXML))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stmt.Ticker != "BBRI" {
		t.Errorf("expected ticker BBRI, got %s", stmt.Ticker)
	}
	if stmt.Year != 2024 || stmt.Period != "FY" {
		t.Errorf("expected 2024 FY, got %d %s", stmt.Year, stmt.Period)
	}
	if stmt.Core.TotalAssets != 1965000000000 {
		t.Errorf("expected total assets 1965000000000, got %f", stmt.Core.TotalAssets)
	}
	if stmt.Core.GrossProfit != 120000000000 {
		t.Errorf("expected gross profit 120000000000, got %f", stmt.Core.GrossProfit)
	}
	if stmt.Core.OperatingIncome != 75000000000 {
		t.Errorf("expected operating income 75000000000, got %f", stmt.Core.OperatingIncome)
	}
	if stmt.Core.NetIncomeParent != 60000000000 {
		t.Errorf("expected parent net income 60000000000, got %f", stmt.Core.NetIncomeParent)
	}
	if stmt.Core.FreeCashFlow != 55000000000 {
		t.Errorf("expected free cash flow 55000000000 (65B - 10B), got %f", stmt.Core.FreeCashFlow)
	}
	if stmt.Core.SharesOutstanding != 151559000000 {
		t.Errorf("expected shares 151559000000, got %f", stmt.Core.SharesOutstanding)
	}
}

func TestParseFlexibleDate(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"2024-12-31", 2024},
		{"December 31, 2023", 2023},
		{"30 September 2022", 2022},
		{"31-03-2021", 2021},
	}

	for _, tt := range tests {
		dt, err := ParseFlexibleDate(tt.input)
		if err != nil {
			t.Errorf("failed to parse date %s: %v", tt.input, err)
		} else if dt.Year() != tt.expected {
			t.Errorf("expected year %d for %s, got %d", tt.expected, tt.input, dt.Year())
		}
	}
}
