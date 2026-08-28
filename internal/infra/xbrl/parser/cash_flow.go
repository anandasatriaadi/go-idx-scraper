package parser

import (
	"math"

	domain "github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
)

// AssignCashFlowMetric maps cash flow statement line item taxonomy tags to CoreFinancials
func AssignCashFlowMetric(c *domain.CoreFinancials, tag string, val float64) {
	switch tag {
	case "NetCashFlowsReceivedFromUsedInOperatingActivities", "NetCashFlowsFromUsedInOperatingActivities",
		"CashGeneratedFromUsedInOperations":
		if c.OperatingCashFlow == 0 {
			c.OperatingCashFlow = val
		}
	case "NetCashFlowsReceivedFromUsedInInvestingActivities", "NetCashFlowsFromUsedInInvestingActivities":
		c.InvestingCashFlow = val
	case "NetCashFlowsReceivedFromUsedInFinancingActivities", "NetCashFlowsFromUsedInFinancingActivities":
		c.FinancingCashFlow = val
	case "PaymentsForPropertyPlantEquipment", "PaymentsForAcquisitionOfPropertyPlantAndEquipment",
		"PaymentsForAcquisitionOfPropertyAndEquipment", "PurchaseOfPropertyPlantAndEquipment",
		"AdditionInPropertyPlantAndEquipment", "PaymentsForAcquisitionOfFixedAssets",
		"PaymentsForAcquisitionOfOilAndGasProperties", "PaymentsForAcquisitionOfMiningProperties",
		"PaymentsForAcquisitionOfBearerPlants", "PaymentsForAcquisitionOfBiologicalAssets",
		"PaymentsForExplorationAndEvaluationAssets":
		c.CapEx += math.Abs(val)
	case "DistributionsOfCashDividends", "DividendsPaidFromFinancingActivities", "PaymentsOfCashDividends",
		"PaymentOfDividends", "DividendsPaid", "DividendsPaidToOwnersOfParentEntity":
		c.DividendsPaid += math.Abs(val)
	}
}

// FinalizeCashFlow derives FreeCashFlow (CFO - CapEx) and EBITDA estimates
func FinalizeCashFlow(c *domain.CoreFinancials) {
	if c.FreeCashFlow == 0 && c.OperatingCashFlow != 0 {
		c.FreeCashFlow = c.OperatingCashFlow - c.CapEx
	}
	if c.EBITDA == 0 && c.OperatingIncome != 0 {
		c.EBITDA = c.OperatingIncome + (c.CapEx * 0.7) // Estimate D&A if not explicitly separated
	}
}
