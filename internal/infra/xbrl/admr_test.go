package xbrl

import (
	"os"
	"testing"

	domain "github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
)

func TestParseADMR_Diagnostic(t *testing.T) {
	for _, ticker := range []string{"ADMR", "ADRO", "INKP"} {
		for _, yr := range []string{"2024", "2025"} {
			path := "../../../saham/FinancialStatement-" + yr + "-Audit-" + ticker + "-instance.zip"
			if _, err := os.Stat(path); err != nil {
				continue
			}

			stmt, err := ParseAnyFiling(path)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			_ = domain.ComputeValuationAndRatios(stmt, nil, 1500)

			t.Logf("=== %s %s Audit ===", ticker, yr)
			t.Logf("  Currency: %s | FX: %f | Multiplier: %f", stmt.Metadata.Currency, stmt.Metadata.ConversionRate, stmt.Metadata.RoundingMultiplier)
			t.Logf("  NetIncomeParent: %.2f | Equity: %.2f | Shares: %.2f", stmt.Core.NetIncomeParent, stmt.Core.TotalEquity, stmt.Core.SharesOutstanding)
			t.Logf("  NormalizedEPS: %.2f IDR | NormalizedBVPS: %.2f IDR | Graham: %.2f IDR | MOS: %.2f%%", stmt.Valuation.NormalizedEPS, stmt.Valuation.NormalizedBVPS, stmt.Valuation.GrahamNumber, stmt.Valuation.MarginOfSafetyPct)
		}
	}
}
