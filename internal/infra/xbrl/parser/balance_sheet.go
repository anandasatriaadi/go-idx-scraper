package parser

import (
	"math"

	domain "github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
)

// AssignBalanceSheetMetric maps balance sheet line item taxonomy tags to CoreFinancials
func AssignBalanceSheetMetric(c *domain.CoreFinancials, tag string, val float64) {
	switch tag {
	case "Assets", "TotalAssets":
		c.TotalAssets = val
	case "CashAndCashEquivalents", "CashAndBankBalances", "CashAndCashEquivalentsAtEndOfPeriod", "CashAndCashEquivalentsAtEndOfYear":
		if c.CashAndEquivalents == 0 {
			c.CashAndEquivalents = val
		}
	case "CurrentAssets", "TotalCurrentAssets":
		c.CurrentAssets = val
	case "Liabilities", "TotalLiabilities":
		c.TotalLiabilities = val
	case "CurrentLiabilities", "TotalCurrentLiabilities":
		c.CurrentLiabilities = val
	case "ShortTermBankLoans", "ShortTermLoans", "ShortTermBorrowings", "CurrentMaturitiesOfLongTermBankLoans",
		"CurrentMaturitiesOfLongTermLoans", "CurrentMaturitiesOfBondsPayable", "CurrentMaturitiesOfSukuk",
		"CurrentMaturitiesOfLeaseLiabilities", "CurrentLeaseLiabilities", "ShortTermBondsPayable",
		"ShortTermSukuk", "ShortTermNotesPayable":
		c.ShortTermDebt += math.Abs(val)
	case "LongTermBankLoans", "LongTermLoans", "LongTermBorrowings", "BondsPayable", "LongTermBondsPayable",
		"Sukuk", "LongTermSukuk", "NonCurrentLeaseLiabilities", "LongTermLeaseLiabilities",
		"LongTermNotesPayable", "SubordinatedLoans", "SubordinatedBonds":
		c.LongTermDebt += math.Abs(val)
	case "Equity", "TotalEquity", "EquityAttributableToOwnersOfParentEntity":
		if c.TotalEquity == 0 || tag == "Equity" || tag == "TotalEquity" {
			c.TotalEquity = val
		}
	case "RetainedEarningsUnappropriated", "UnappropriatedRetainedEarnings", "RetainedEarnings":
		if c.RetainedEarnings == 0 {
			c.RetainedEarnings = val
		}
	}
}

// FinalizeBalanceSheet calculates derived balance sheet totals like TotalDebt and WorkingCapital
func FinalizeBalanceSheet(c *domain.CoreFinancials) {
	if c.TotalDebt == 0 {
		c.TotalDebt = c.ShortTermDebt + c.LongTermDebt
	}
	if c.WorkingCapital == 0 && c.CurrentAssets != 0 {
		c.WorkingCapital = c.CurrentAssets - c.CurrentLiabilities
	}
}
