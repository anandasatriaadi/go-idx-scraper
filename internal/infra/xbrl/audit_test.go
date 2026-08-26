package xbrl

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestParseAllSampleFilings(t *testing.T) {
	files, err := filepath.Glob("../../../saham/*-Audit-*-instance.zip")
	if err != nil || len(files) == 0 {
		t.Skip("No audit filings found in saham/")
	}

	t.Logf("Found %d audit filings to audit", len(files))

	for _, file := range files {
		stmt, err := ParseAnyFiling(file)
		if err != nil {
			t.Errorf("[%s] Failed to parse: %v", filepath.Base(file), err)
			continue
		}

		c := stmt.Core
		m := stmt.Metadata

		// Audit 1: Revenue vs Net Income Check for non-banks
		isBank := strings.Contains(m.Sector, "Financial") || strings.Contains(m.Industry, "Bank")
		if !isBank {
			if c.Revenue > 0 && c.NetIncomeParent > 0 && c.Revenue < c.NetIncomeParent {
				t.Errorf("[%s] ANOMALY: Revenue (%f) is SMALLER than Net Income Parent (%f)", stmt.Ticker, c.Revenue, c.NetIncomeParent)
			}
			if c.GrossProfit > 0 && c.Revenue > 0 && c.Revenue < c.GrossProfit {
				t.Errorf("[%s] ANOMALY: Revenue (%f) is SMALLER than Gross Profit (%f)", stmt.Ticker, c.Revenue, c.GrossProfit)
			}
		}

		// Audit 2: CapEx should be positive magnitude
		if c.CapEx < 0 {
			t.Errorf("[%s] ANOMALY: CapEx is negative (%f)", stmt.Ticker, c.CapEx)
		}

		// Audit 3: Total Assets should be positive
		if c.TotalAssets <= 0 {
			t.Errorf("[%s] ANOMALY: TotalAssets is non-positive (%f)", stmt.Ticker, c.TotalAssets)
		}

		// Audit 4: Total Equity should be present
		if c.TotalEquity == 0 {
			t.Errorf("[%s] ANOMALY: TotalEquity is zero", stmt.Ticker)
		}

		t.Logf("[%s %d %s] Cur: %s | Rev: %.2e | GP: %.2e | NI: %.2e | Cash: %.2e | Debt: %.2e | CapEx: %.2e | FCF: %.2e",
			stmt.Ticker, stmt.Year, stmt.Period, m.Currency, c.Revenue, c.GrossProfit, c.NetIncomeParent, c.CashAndEquivalents, c.TotalDebt, c.CapEx, c.FreeCashFlow)
	}
}
