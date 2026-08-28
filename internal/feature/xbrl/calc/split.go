package calc

import (
	"math"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
)

// ApplyStockSplitAdjustment detects significant changes in shares outstanding across historical statements
// (e.g. 1:2, 1:5, 1:10 stock splits or reverse splits) and normalizes historical per-share metrics
// (EPS, BVPS, Graham Number, RevenuePerShare, CashPerShare, FCFPerShare) to the latest share basis
// to ensure historical valuation aligns with split-adjusted market price time series.
func ApplyStockSplitAdjustment(statements []*xbrl.Statement) {
	if len(statements) < 2 {
		return
	}

	// Find latest statement with valid shares
	var latestStmt *xbrl.Statement
	for i := len(statements) - 1; i >= 0; i-- {
		if statements[i].Core.SharesOutstanding > 1000 {
			latestStmt = statements[i]
			break
		}
	}
	if latestStmt == nil {
		return
	}

	targetShares := latestStmt.Core.SharesOutstanding

	for _, stmt := range statements {
		if stmt == latestStmt {
			continue
		}
		currShares := stmt.Core.SharesOutstanding
		if currShares <= 1 || targetShares <= 1 {
			continue
		}

		ratio := targetShares / currShares
		// If ratio indicates a stock split (>= 1.8x, like 2:1, 5:1, 10:1) or reverse split (<= 0.55x)
		if ratio >= 1.8 || ratio <= 0.55 {
			fxRate := 1.0
			if stmt.Metadata.Currency == "USD" {
				if stmt.Metadata.ConversionRate >= 1000 {
					fxRate = stmt.Metadata.ConversionRate
				} else if stmt.Metadata.ConversionRate > 0 && stmt.Metadata.ConversionRate < 1000 {
					fxRate = stmt.Metadata.ConversionRate * 1000
				} else {
					fxRate = 16000.0
				}
			}

			effectiveNetIncome := stmt.Core.NetIncome
			if stmt.Core.NetIncomeParent != 0 {
				effectiveNetIncome = stmt.Core.NetIncomeParent
			}

			v := &stmt.Valuation
			c := &stmt.Core

			// Recompute per-share metrics using targetShares
			v.NormalizedEPS = (effectiveNetIncome * fxRate) / targetShares
			v.NormalizedBVPS = (c.TotalEquity * fxRate) / targetShares
			v.RevenuePerShare = (c.Revenue * fxRate) / targetShares
			v.CashPerShare = (c.CashAndEquivalents * fxRate) / targetShares
			v.FreeCashFlowPerShare = (c.FreeCashFlow * fxRate) / targetShares

			if v.NormalizedEPS > 0 && v.NormalizedBVPS > 0 {
				v.GrahamNumber = math.Sqrt(22.5 * v.NormalizedEPS * v.NormalizedBVPS)
			} else {
				v.GrahamNumber = 0
			}

			if v.CurrentPrice > 0 {
				if v.NormalizedEPS > 0 {
					v.PERatio = v.CurrentPrice / v.NormalizedEPS
					v.EarningsYieldPct = (v.NormalizedEPS / v.CurrentPrice) * 100
				}
				if v.NormalizedBVPS > 0 {
					v.PBRatio = v.CurrentPrice / v.NormalizedBVPS
				}
				if v.GrahamNumber > 0 {
					v.MarginOfSafetyPct = ((v.GrahamNumber - v.CurrentPrice) / v.GrahamNumber) * 100
				}
			}
		}
	}
}
