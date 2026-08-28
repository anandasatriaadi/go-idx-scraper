package calc

import (
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
)

// ComputeProfitability calculates Gross Margin %, Operating Margin %, Net Margin %, ROE, and ROA.
func ComputeProfitability(r *xbrl.ComputedRatios, c *xbrl.CoreFinancials) {
	effectiveNetIncome := c.NetIncome
	if c.NetIncomeParent != 0 {
		effectiveNetIncome = c.NetIncomeParent
	}

	if c.Revenue > 0 {
		r.GrossMarginPct = (c.GrossProfit / c.Revenue) * 100
		r.OperatingMarginPct = (c.OperatingIncome / c.Revenue) * 100
		r.NetMarginPct = (effectiveNetIncome / c.Revenue) * 100
	}
	if c.TotalEquity > 0 {
		r.ROE = effectiveNetIncome / c.TotalEquity
	}
	if c.TotalAssets > 0 {
		r.ROA = effectiveNetIncome / c.TotalAssets
	}
}
