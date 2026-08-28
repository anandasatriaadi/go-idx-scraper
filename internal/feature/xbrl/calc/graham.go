package calc

import (
	"math"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
)

// ComputeGrahamFairValue calculates Benjamin Graham Fair Value: sqrt(22.5 * EPS * BVPS).
func ComputeGrahamFairValue(v *xbrl.ValuationMetrics) {
	if v.NormalizedEPS > 0 && v.NormalizedBVPS > 0 {
		v.GrahamNumber = math.Sqrt(22.5 * v.NormalizedEPS * v.NormalizedBVPS)
	} else {
		v.GrahamNumber = 0
	}
}

// ComputeValuationMultiples calculates MarketCap, EnterpriseValue, P/E, P/B, P/S, P/FCF, EV/EBIT, EV/EBITDA, Earnings Yield %, and Margin of Safety %.
func ComputeValuationMultiples(v *xbrl.ValuationMetrics, c *xbrl.CoreFinancials, currentStockPrice float64, shares float64, fxRate float64) {
	v.CurrentPrice = currentStockPrice
	if currentStockPrice <= 0 {
		return
	}

	if shares > 1 {
		v.MarketCap = currentStockPrice * shares
		v.EnterpriseValue = v.MarketCap + (c.TotalDebt * fxRate) - (c.CashAndEquivalents * fxRate)
	}
	if v.NormalizedEPS > 0 {
		v.PERatio = currentStockPrice / v.NormalizedEPS
		v.EarningsYieldPct = (v.NormalizedEPS / currentStockPrice) * 100
	}
	if v.NormalizedBVPS > 0 {
		v.PBRatio = currentStockPrice / v.NormalizedBVPS
	}
	if v.RevenuePerShare > 0 {
		v.PSRatio = currentStockPrice / v.RevenuePerShare
	}
	if v.FreeCashFlowPerShare > 0 {
		v.PFCFRatio = currentStockPrice / v.FreeCashFlowPerShare
	}
	if c.OperatingIncome > 0 && v.EnterpriseValue > 0 {
		v.EVToEBIT = v.EnterpriseValue / (c.OperatingIncome * fxRate)
	}
	if c.EBITDA > 0 && v.EnterpriseValue > 0 {
		v.EVToEBITDA = v.EnterpriseValue / (c.EBITDA * fxRate)
	}
	if v.GrahamNumber > 0 {
		v.MarginOfSafetyPct = ((v.GrahamNumber - currentStockPrice) / v.GrahamNumber) * 100
	}
}
