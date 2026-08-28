package calc

import (
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
)

// ComputeROIC calculates Return on Invested Capital: NOPAT / Invested Capital (with 22% Indonesian standard corporate tax rate).
func ComputeROIC(c *xbrl.CoreFinancials) float64 {
	investedCapital := (c.TotalEquity + c.TotalDebt) - c.CashAndEquivalents
	nopat := c.OperatingIncome * (1 - 0.22) // 22% Indonesian standard corporate tax rate
	if investedCapital > 0 {
		return nopat / investedCapital
	}
	return 0.0
}
