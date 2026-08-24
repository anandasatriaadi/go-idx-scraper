package xbrl

import (
	"math"
)

// ComputeValuationAndRatios calculates ROIC, Piotroski, Altman Z, Graham Fair Value, and MOS
func ComputeValuationAndRatios(stmt *Statement, priorStmt *Statement, currentStockPrice float64) error {
	c := &stmt.Core
	r := &stmt.ComputedRatios
	v := &stmt.Valuation

	// 1. Profitability & Margins
	if c.Revenue > 0 {
		r.GrossMarginPct = (c.GrossProfit / c.Revenue) * 100
		r.OperatingMarginPct = (c.OperatingIncome / c.Revenue) * 100
		r.NetMarginPct = (c.NetIncome / c.Revenue) * 100
	}
	if c.TotalEquity > 0 {
		r.ROE = c.NetIncome / c.TotalEquity
	}
	if c.TotalAssets > 0 {
		r.ROA = c.NetIncome / c.TotalAssets
	}

	// 2. Return on Invested Capital (ROIC)
	investedCapital := (c.TotalEquity + c.TotalDebt) - c.CashAndEquivalents
	nopat := c.OperatingIncome * (1 - 0.22) // 22% Indonesian standard corporate tax rate
	if investedCapital > 0 {
		r.ROIC = nopat / investedCapital
	}

	// 3. Leverage & Solvency
	if c.TotalEquity > 0 {
		r.DebtToEquity = c.TotalDebt / c.TotalEquity
	}
	r.NetDebt = c.TotalDebt - c.CashAndEquivalents
	if c.FinanceCosts > 0 {
		r.InterestCoverageRatio = c.OperatingIncome / c.FinanceCosts
	}
	if c.CurrentLiabilities > 0 {
		r.CurrentRatio = c.CurrentAssets / c.CurrentLiabilities
	}
	if c.NetIncome > 0 {
		r.FCFConversionPct = (c.FreeCashFlow / c.NetIncome) * 100
	}

	// 4. Emerging Market Altman Z''-Score
	if c.TotalAssets > 0 && c.TotalLiabilities > 0 {
		workingCapital := c.CurrentAssets - c.CurrentLiabilities
		x1 := workingCapital / c.TotalAssets
		x2 := c.RetainedEarnings / c.TotalAssets
		x3 := c.OperatingIncome / c.TotalAssets
		x4 := c.TotalEquity / c.TotalLiabilities
		r.AltmanZScore = (6.56 * x1) + (3.26 * x2) + (6.72 * x3) + (1.05 * x4)
	}

	// 5. Piotroski F-Score (0 to 9)
	fScore := 0
	if r.ROA > 0 {
		fScore++
	}
	if c.OperatingCashFlow > 0 {
		fScore++
	}
	if c.OperatingCashFlow > c.NetIncome {
		fScore++
	}
	if priorStmt != nil {
		if r.ROA > priorStmt.ComputedRatios.ROA {
			fScore++
		}
		if c.LongTermDebt < priorStmt.Core.LongTermDebt {
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
		assetTurnoverCurr := c.Revenue / c.TotalAssets
		assetTurnoverPrior := 0.0
		if priorStmt.Core.TotalAssets > 0 {
			assetTurnoverPrior = priorStmt.Core.Revenue / priorStmt.Core.TotalAssets
		}
		if assetTurnoverCurr > assetTurnoverPrior {
			fScore++
		}
	} else {
		// Single period baseline heuristics
		if r.CurrentRatio > 1.2 {
			fScore++
		}
		if r.DebtToEquity < 1.0 {
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
	r.PiotroskiFScore = fScore

	// 6. Currency Normalization (USD -> IDR) & Satuan / Rounding Scaling
	fxRate := 1.0
	if stmt.Metadata.Currency == "USD" {
		if stmt.Metadata.ConversionRate > 0 {
			fxRate = stmt.Metadata.ConversionRate
		} else {
			fxRate = 16000.0 // Conservative default if unpopulated
		}
	}

	multiplier := stmt.Metadata.RoundingMultiplier
	if multiplier < 1.0 {
		multiplier = 1.0
	}

	// If per-share EPS in the XML is scaled by the rounding multiplier (e.g. 0.000340 instead of 340 IDR)
	if v.NormalizedEPS > 0 && v.NormalizedEPS < 1.0 && multiplier >= 1000 {
		v.NormalizedEPS = v.NormalizedEPS * multiplier
	}

	shares := c.SharesOutstanding
	// If shares was reported in "Ribuan / Jutaan" units (e.g. 12,592 instead of 12,592,000,000)
	if shares > 0 && shares < 1000000 && multiplier >= 1000 {
		shares = shares * multiplier
		c.SharesOutstanding = shares
	}

	if shares <= 1 && v.NormalizedEPS > 0 && c.NetIncome > 0 {
		shares = (c.NetIncome * fxRate) / v.NormalizedEPS
		c.SharesOutstanding = shares
	}
	if shares <= 0 {
		shares = 1.0
	}

	// Normalized per-share values in IDR
	if v.NormalizedEPS == 0 && shares > 1 {
		v.NormalizedEPS = (c.NetIncome * fxRate) / shares
	}
	if shares > 1 {
		v.NormalizedBVPS = (c.TotalEquity * fxRate) / shares
		v.RevenuePerShare = (c.Revenue * fxRate) / shares
		v.CashPerShare = (c.CashAndEquivalents * fxRate) / shares
		v.FreeCashFlowPerShare = (c.FreeCashFlow * fxRate) / shares
	}

	// 7. Benjamin Graham Fair Value Formula
	if v.NormalizedEPS > 0 && v.NormalizedBVPS > 0 {
		v.GrahamNumber = math.Sqrt(22.5 * v.NormalizedEPS * v.NormalizedBVPS)
	}

	// 8. Valuation Multiples & Margin of Safety
	v.CurrentPrice = currentStockPrice
	if currentStockPrice > 0 {
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

	// 9. Quick Ratio (Cash + Trade Receivables / Current Liabilities)
	if c.CurrentLiabilities > 0 {
		v.QuickRatio = c.CashAndEquivalents / c.CurrentLiabilities
	}

	return nil
}
