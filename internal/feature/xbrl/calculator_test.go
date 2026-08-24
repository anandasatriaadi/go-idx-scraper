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
