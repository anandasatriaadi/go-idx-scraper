package calc

import (
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
)

// ComputeValuationAndRatios calculates ROIC, Piotroski, Altman Z, Graham Fair Value, and MOS.
func ComputeValuationAndRatios(stmt *xbrl.Statement, priorStmt *xbrl.Statement, currentStockPrice float64) error {
	c := &stmt.Core
	r := &stmt.ComputedRatios
	v := &stmt.Valuation

	// 1. Profitability & Margins
	ComputeProfitability(r, c)

	// 2. Return on Invested Capital (ROIC)
	r.ROIC = ComputeROIC(c)

	// 3. Leverage & Solvency
	ComputeSolvency(r, v, c)

	// 4. Emerging Market Altman Z''-Score
	r.AltmanZScore = ComputeAltmanZScore(c)

	// 5. Piotroski F-Score (0 to 9)
	r.PiotroskiFScore = ComputePiotroskiFScore(stmt, priorStmt)

	// 6. Currency Normalization (USD -> IDR) & Satuan / Rounding Scaling
	fxRate, shares := NormalizeCurrencyAndShares(stmt, priorStmt)

	// 7. Benjamin Graham Fair Value Formula
	if shares > 1000 {
		ComputeGrahamFairValue(v)
	}

	// 8. Valuation Multiples & Margin of Safety
	ComputeValuationMultiples(v, c, currentStockPrice, shares, fxRate)

	return nil
}
