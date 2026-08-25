package xbrl

import (
	"testing"
)

func TestCalculator_ForensicsAndValuation(t *testing.T) {
	curr := &Statement{
		Ticker: "AADI",
		Metadata: StatementMetadata{
			Currency:       "USD",
			ConversionRate: 15400.0,
		},
		Core: CoreFinancials{
			SharesOutstanding:  3000000000,
			TotalAssets:        5780540000,
			CashAndEquivalents: 914431000,
			CurrentAssets:      1780200000,
			TotalLiabilities:   1999310000,
			CurrentLiabilities: 650200000,
			TotalDebt:          570000000,
			TotalEquity:        3781230000,
			RetainedEarnings:   2450000000,
			Revenue:            1044192000,
			GrossProfit:        257553000,
			OperatingIncome:    210400000,
			FinanceCosts:       15200000,
			NetIncome:          153768000,
			OperatingCashFlow:  285400000,
			CapEx:              62300000,
			FreeCashFlow:       223100000,
		},
	}

	err := ComputeValuationAndRatios(curr, nil, 4150.0)
	if err != nil {
		t.Fatalf("ComputeValuationAndRatios failed: %v", err)
	}

	if curr.ComputedRatios.ROIC <= 0 {
		t.Errorf("Expected positive ROIC, got %f", curr.ComputedRatios.ROIC)
	}
	if curr.ComputedRatios.AltmanZScore <= 0 {
		t.Errorf("Expected positive Altman Z-Score, got %f", curr.ComputedRatios.AltmanZScore)
	}
	if curr.Valuation.NormalizedEPS <= 0 {
		t.Errorf("Expected positive Normalized EPS, got %f", curr.Valuation.NormalizedEPS)
	}
	if curr.Valuation.GrahamNumber <= 0 {
		t.Errorf("Expected positive Graham Number, got %f", curr.Valuation.GrahamNumber)
	}
	if curr.Valuation.MarginOfSafetyPct <= 0 {
		t.Errorf("Expected positive Margin of Safety, got %f", curr.Valuation.MarginOfSafetyPct)
	}
}

func TestCalculator_WithPriorPeriod(t *testing.T) {
	prior := &Statement{
		Ticker: "AADI",
		Core: CoreFinancials{
			SharesOutstanding: 3000000000,
			TotalAssets:       5000000000,
			LongTermDebt:      600000000,
			Revenue:           900000000,
		},
		ComputedRatios: ComputedRatios{
			ROA:            0.08,
			CurrentRatio:   2.0,
			GrossMarginPct: 20.0,
		},
	}

	curr := &Statement{
		Ticker: "AADI",
		Metadata: StatementMetadata{
			Currency:       "USD",
			ConversionRate: 15400.0,
		},
		Core: CoreFinancials{
			SharesOutstanding:  3000000000,
			TotalAssets:        5780540000,
			CashAndEquivalents: 914431000,
			CurrentAssets:      1780200000,
			TotalLiabilities:   1999310000,
			CurrentLiabilities: 650200000,
			LongTermDebt:       450000000,
			TotalDebt:          570000000,
			TotalEquity:        3781230000,
			RetainedEarnings:   2450000000,
			Revenue:            1044192000,
			GrossProfit:        257553000,
			OperatingIncome:    210400000,
			FinanceCosts:       15200000,
			NetIncome:          153768000,
			OperatingCashFlow:  285400000,
			CapEx:              62300000,
			FreeCashFlow:       223100000,
		},
	}

	err := ComputeValuationAndRatios(curr, prior, 4150.0)
	if err != nil {
		t.Fatalf("ComputeValuationAndRatios with prior failed: %v", err)
	}

	if curr.ComputedRatios.PiotroskiFScore < 5 {
		t.Errorf("Expected strong Piotroski F-Score >= 5 with improving YoY fundamentals, got %d", curr.ComputedRatios.PiotroskiFScore)
	}
}

func TestCalculator_ZeroDebtAndParentNetIncome(t *testing.T) {
	prior := &Statement{
		Ticker: "DEBTFREE",
		Core: CoreFinancials{
			SharesOutstanding: 1000000,
			TotalAssets:       1000000000,
			LongTermDebt:      0, // Zero debt
			Revenue:           500000000,
		},
		ComputedRatios: ComputedRatios{
			ROA:            0.05,
			CurrentRatio:   1.5,
			GrossMarginPct: 25.0,
		},
	}

	curr := &Statement{
		Ticker: "DEBTFREE",
		Metadata: StatementMetadata{
			Currency: "IDR",
		},
		Core: CoreFinancials{
			SharesOutstanding:  1000000,
			TotalAssets:        1200000000,
			CashAndEquivalents: 300000000,
			CurrentAssets:      600000000,
			TotalLiabilities:   200000000,
			CurrentLiabilities: 200000000,
			LongTermDebt:       0, // Stays zero debt -> should get +1 point in Piotroski
			TotalDebt:          0,
			TotalEquity:        1000000000,
			RetainedEarnings:   800000000,
			Revenue:            700000000,
			GrossProfit:        350000000,
			OperatingIncome:    200000000,
			NetIncome:          180000000,
			NetIncomeParent:    170000000, // Attributable to parent owners
			OperatingCashFlow:  220000000,
			CapEx:              50000000,
			FreeCashFlow:       170000000,
		},
	}

	err := ComputeValuationAndRatios(curr, prior, 1500.0)
	if err != nil {
		t.Fatalf("ComputeValuationAndRatios failed: %v", err)
	}

	// Verify NetIncomeParent is used for NormalizedEPS
	expectedEPS := 170000000.0 / 1000000.0 // 170
	if curr.Valuation.NormalizedEPS != expectedEPS {
		t.Errorf("Expected NormalizedEPS %f from NetIncomeParent, got %f", expectedEPS, curr.Valuation.NormalizedEPS)
	}

	// Piotroski: ROA>0 (+1), CFO>0 (+1), CFO>NetIncome (+1), ROA up (+1), Zero Debt maintained (+1), CurrentRatio up (+1), No dilution (+1), Gross Margin up (+1), Asset Turnover up (+1) = 9
	if curr.ComputedRatios.PiotroskiFScore != 9 {
		t.Errorf("Expected perfect Piotroski F-Score 9 for pristine debt-free compounder, got %d", curr.ComputedRatios.PiotroskiFScore)
	}
}

func TestCalculator_USDEPSNormalization(t *testing.T) {
	curr := &Statement{
		Ticker: "USDFILE",
		Metadata: StatementMetadata{
			Currency:       "USD",
			ConversionRate: 16000.0,
		},
		Core: CoreFinancials{
			SharesOutstanding:  2000000000,
			TotalAssets:        4000000000, // 4B USD
			CashAndEquivalents: 500000000,
			CurrentAssets:      1000000000,
			TotalLiabilities:   1000000000,
			CurrentLiabilities: 500000000,
			TotalEquity:        3000000000, // 3B USD
			Revenue:            2000000000,
			GrossProfit:        600000000,
			OperatingIncome:    400000000,
			NetIncome:          300000000, // 300M USD
			OperatingCashFlow:  350000000,
			CapEx:              100000000,
			FreeCashFlow:       250000000,
		},
		Valuation: ValuationMetrics{
			NormalizedEPS: 0.15, // Explicit USD EPS (300M / 2B = 0.15 USD)
		},
	}

	err := ComputeValuationAndRatios(curr, nil, 3000.0) // 3000 IDR stock price
	if err != nil {
		t.Fatalf("ComputeValuationAndRatios failed: %v", err)
	}

	// NormalizedEPS in IDR should be 0.15 * 16000 = 2400 IDR
	expectedEPS := 2400.0
	if curr.Valuation.NormalizedEPS != expectedEPS {
		t.Errorf("Expected NormalizedEPS in IDR to be %f, got %f", expectedEPS, curr.Valuation.NormalizedEPS)
	}

	// BVPS in IDR should be (3B * 16000) / 2B = 24000 IDR
	expectedBVPS := 24000.0
	if curr.Valuation.NormalizedBVPS != expectedBVPS {
		t.Errorf("Expected NormalizedBVPS in IDR to be %f, got %f", expectedBVPS, curr.Valuation.NormalizedBVPS)
	}

	// PE Ratio should be 3000 IDR / 2400 IDR = 1.25
	expectedPE := 3000.0 / 2400.0
	if curr.Valuation.PERatio != expectedPE {
		t.Errorf("Expected PE ratio %f, got %f", expectedPE, curr.Valuation.PERatio)
	}
}
