package calc

import (
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
)

// NormalizeCurrencyAndShares normalizes USD currency to IDR via exchange rate and handles rounding multiplier scaling.
// Returns (fxRate, effectiveShares).
func NormalizeCurrencyAndShares(stmt *xbrl.Statement, priorStmt *xbrl.Statement) (float64, float64) {
	c := &stmt.Core
	v := &stmt.Valuation

	effectiveNetIncome := c.NetIncome
	if c.NetIncomeParent != 0 {
		effectiveNetIncome = c.NetIncomeParent
	}

	fxRate := 1.0
	if stmt.Metadata.Currency == "USD" {
		if stmt.Metadata.ConversionRate >= 1000 {
			fxRate = stmt.Metadata.ConversionRate
		} else if stmt.Metadata.ConversionRate > 0 && stmt.Metadata.ConversionRate < 1000 {
			fxRate = stmt.Metadata.ConversionRate * 1000
		} else {
			fxRate = 16000.0 // Conservative default if unpopulated
		}
	}

	multiplier := stmt.Metadata.RoundingMultiplier
	if multiplier < 1.0 {
		multiplier = 1.0
	}

	// If per-share EPS in the XML is scaled by the rounding multiplier (e.g. 0.000340 instead of 340 IDR)
	if v.NormalizedEPS > 0 && v.NormalizedEPS < 1.0 && multiplier >= 1000 && stmt.Metadata.Currency != "USD" {
		v.NormalizedEPS = v.NormalizedEPS * multiplier
	}

	shares := c.SharesOutstanding
	// If shares was reported in "Ribuan / Jutaan" units (e.g. 12,592 instead of 12,592,000,000)
	if shares > 0 && shares < 1000000 && multiplier >= 1000 {
		shares = shares * multiplier
		c.SharesOutstanding = shares
	}

	if shares <= 1 && v.NormalizedEPS > 0 {
		if c.NetIncome > 0 {
			shares = c.NetIncome / v.NormalizedEPS
		} else if effectiveNetIncome > 0 {
			shares = effectiveNetIncome / v.NormalizedEPS
		}
		c.SharesOutstanding = shares
	}
	if shares <= 1 && priorStmt != nil && priorStmt.Core.SharesOutstanding > 1000 {
		shares = priorStmt.Core.SharesOutstanding
		c.SharesOutstanding = shares
	}
	if shares <= 0 {
		shares = 1.0
	}

	if shares > 1000 {
		if effectiveNetIncome != 0 {
			v.NormalizedEPS = (effectiveNetIncome * fxRate) / shares
		} else if stmt.Metadata.Currency == "USD" && v.NormalizedEPS > 0 {
			v.NormalizedEPS = v.NormalizedEPS * fxRate
		}
		v.NormalizedBVPS = (c.TotalEquity * fxRate) / shares
		v.RevenuePerShare = (c.Revenue * fxRate) / shares
		v.CashPerShare = (c.CashAndEquivalents * fxRate) / shares
		v.FreeCashFlowPerShare = (c.FreeCashFlow * fxRate) / shares
	} else {
		v.NormalizedEPS = 0
		v.NormalizedBVPS = 0
	}

	return fxRate, shares
}
