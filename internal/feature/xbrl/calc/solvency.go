package calc

import (
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
)

// ComputeSolvency calculates Leverage, Solvency, and Liquidity ratios.
func ComputeSolvency(r *xbrl.ComputedRatios, v *xbrl.ValuationMetrics, c *xbrl.CoreFinancials) {
	effectiveNetIncome := c.NetIncome
	if c.NetIncomeParent != 0 {
		effectiveNetIncome = c.NetIncomeParent
	}

	if c.TotalEquity > 0 {
		r.DebtToEquity = c.TotalDebt / c.TotalEquity
	}
	r.NetDebt = c.TotalDebt - c.CashAndEquivalents
	if c.FinanceCosts > 0 {
		r.InterestCoverageRatio = c.OperatingIncome / c.FinanceCosts
	}
	if c.CurrentLiabilities > 0 {
		r.CurrentRatio = c.CurrentAssets / c.CurrentLiabilities
		v.QuickRatio = c.CashAndEquivalents / c.CurrentLiabilities
	}
	if effectiveNetIncome > 0 {
		r.FCFConversionPct = (c.FreeCashFlow / effectiveNetIncome) * 100
	}
}
