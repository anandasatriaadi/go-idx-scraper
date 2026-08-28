package parser

import (
	"math"

	domain "github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
)

// AssignIncomeStatementMetric maps income statement line item taxonomy tags to CoreFinancials & Valuation
func AssignIncomeStatementMetric(stmt *domain.Statement, tag string, val float64) {
	c := &stmt.Core
	switch tag {
	case "SalesAndRevenue", "Revenues":
		c.Revenue = val
	case "InterestIncome", "InterestAndFinanceIncome":
		if c.Revenue == 0 {
			c.Revenue = val
		}
	case "CostOfSalesAndRevenue":
		c.CostOfRevenue = val
	case "InterestExpense":
		if c.CostOfRevenue == 0 {
			c.CostOfRevenue = val
		}
	case "GrossProfit":
		c.GrossProfit = val
	case "NetInterestIncome":
		if c.GrossProfit == 0 {
			c.GrossProfit = val
		}
	case "OperatingIncomeExpense", "OperatingIncome", "OperatingProfitLoss", "ProfitLossFromOperatingActivities":
		c.OperatingIncome = val
	case "ProfitLossBeforeIncomeTax", "ProfitLossBeforeTax":
		if c.OperatingIncome == 0 {
			c.OperatingIncome = val + c.FinanceCosts
		}
	case "FinanceCosts", "InterestAndFinanceCosts":
		c.FinanceCosts = math.Abs(val)
	case "ProfitLoss":
		c.NetIncome = val
	case "ProfitLossAttributableToOwnersOfParentEntity", "ProfitLossAttributableToParentEntity":
		c.NetIncomeParent = val
	case "BasicEarningsLossPerShareFromContinuingOperations", "BasicEarningsLossPerShare",
		"DilutedEarningsLossPerShareFromContinuingOperations", "DilutedEarningsLossPerShare":
		if val > 0 {
			if stmt.Valuation.NormalizedEPS == 0 {
				stmt.Valuation.NormalizedEPS = val
			}
		}
	}
}

// FinalizeIncomeStatement applies cascading fallbacks for NetIncomeParent, OperatingIncome, Revenue, and GrossProfit
func FinalizeIncomeStatement(c *domain.CoreFinancials) {
	if c.NetIncomeParent == 0 && c.NetIncome != 0 {
		c.NetIncomeParent = c.NetIncome
	}
	if c.NetIncome == 0 && c.NetIncomeParent != 0 {
		c.NetIncome = c.NetIncomeParent
	}
	if c.OperatingIncome == 0 && c.NetIncome != 0 {
		c.OperatingIncome = c.NetIncome + c.FinanceCosts
	}
	if c.GrossProfit > 0 && c.CostOfRevenue > 0 && (c.Revenue == 0 || c.Revenue < c.GrossProfit) {
		c.Revenue = c.GrossProfit + c.CostOfRevenue
	}
	if c.GrossProfit == 0 && c.Revenue != 0 && c.CostOfRevenue != 0 {
		c.GrossProfit = c.Revenue - c.CostOfRevenue
	}
}
