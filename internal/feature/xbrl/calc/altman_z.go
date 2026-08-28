package calc

import (
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
)

// ComputeAltmanZScore calculates Emerging Market Altman Z''-Score: 6.56*X1 + 3.26*X2 + 6.72*X3 + 1.05*X4.
func ComputeAltmanZScore(c *xbrl.CoreFinancials) float64 {
	if c.TotalAssets > 0 && c.TotalLiabilities > 0 {
		workingCapital := c.CurrentAssets - c.CurrentLiabilities
		x1 := workingCapital / c.TotalAssets
		x2 := c.RetainedEarnings / c.TotalAssets
		x3 := c.OperatingIncome / c.TotalAssets
		x4 := c.TotalEquity / c.TotalLiabilities
		return (6.56 * x1) + (3.26 * x2) + (6.72 * x3) + (1.05 * x4)
	}
	return 0.0
}
