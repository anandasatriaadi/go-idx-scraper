package calc

import (
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
)

// ComputePiotroskiFScore calculates the 0-9 Piotroski F-Score measuring financial strength.
func ComputePiotroskiFScore(stmt *xbrl.Statement, priorStmt *xbrl.Statement) int {
	c := &stmt.Core
	r := &stmt.ComputedRatios

	effectiveNetIncome := c.NetIncome
	if c.NetIncomeParent != 0 {
		effectiveNetIncome = c.NetIncomeParent
	}

	fScore := 0
	if r.ROA > 0 {
		fScore++
	}
	if c.OperatingCashFlow > 0 {
		fScore++
	}
	if c.OperatingCashFlow > effectiveNetIncome {
		fScore++
	}
	if priorStmt != nil {
		if r.ROA > priorStmt.ComputedRatios.ROA {
			fScore++
		}
		if c.LongTermDebt < priorStmt.Core.LongTermDebt || (c.LongTermDebt == 0 && priorStmt.Core.LongTermDebt == 0) {
			fScore++
		}
		if r.CurrentRatio > priorStmt.ComputedRatios.CurrentRatio {
			fScore++
		}
		if c.SharesOutstanding <= priorStmt.Core.SharesOutstanding {
			fScore++
		}
		if r.GrossMarginPct > priorStmt.ComputedRatios.GrossMarginPct {
			fScore++
		}
		assetTurnoverCurr := 0.0
		if c.TotalAssets > 0 {
			assetTurnoverCurr = c.Revenue / c.TotalAssets
		}
		assetTurnoverPrior := 0.0
		if priorStmt.Core.TotalAssets > 0 {
			assetTurnoverPrior = priorStmt.Core.Revenue / priorStmt.Core.TotalAssets
		}
		if assetTurnoverCurr > assetTurnoverPrior && assetTurnoverCurr > 0 {
			fScore++
		}
	} else {
		// Single period baseline heuristics
		if r.CurrentRatio > 1.2 {
			fScore++
		}
		if c.TotalEquity > 0 && r.DebtToEquity < 1.0 {
			fScore++
		}
		if r.GrossMarginPct > 20.0 {
			fScore++
		}
		fScore += 2 // Neutral prior credit
	}
	if fScore > 9 {
		fScore = 9
	}
	return fScore
}
