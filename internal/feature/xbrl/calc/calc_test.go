package calc_test

import (
	"math"
	"testing"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/stock"
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl/calc"
)

func TestComputeProfitability(t *testing.T) {
	c := xbrl.CoreFinancials{
		Revenue:         1000,
		GrossProfit:     400,
		OperatingIncome: 200,
		NetIncome:       150,
		TotalEquity:     500,
		TotalAssets:     2000,
	}
	var r xbrl.ComputedRatios
	calc.ComputeProfitability(&r, &c)

	if r.GrossMarginPct != 40.0 {
		t.Errorf("expected GrossMarginPct=40.0, got %f", r.GrossMarginPct)
	}
	if r.OperatingMarginPct != 20.0 {
		t.Errorf("expected OperatingMarginPct=20.0, got %f", r.OperatingMarginPct)
	}
	if r.NetMarginPct != 15.0 {
		t.Errorf("expected NetMarginPct=15.0, got %f", r.NetMarginPct)
	}
	if r.ROE != 0.30 {
		t.Errorf("expected ROE=0.30, got %f", r.ROE)
	}
	if r.ROA != 0.075 {
		t.Errorf("expected ROA=0.075, got %f", r.ROA)
	}
}

func TestComputeROIC(t *testing.T) {
	c := xbrl.CoreFinancials{
		OperatingIncome:    100,
		TotalEquity:        500,
		TotalDebt:          200,
		CashAndEquivalents: 100,
	}
	roic := calc.ComputeROIC(&c)
	expected := 0.13
	if math.Abs(roic-expected) > 0.0001 {
		t.Errorf("expected ROIC=%f, got %f", expected, roic)
	}
}

func TestComputeSolvency(t *testing.T) {
	c := xbrl.CoreFinancials{
		CurrentAssets:      600,
		CurrentLiabilities: 300,
		CashAndEquivalents: 150,
		TotalDebt:          400,
		TotalEquity:        800,
		FinanceCosts:       50,
		OperatingIncome:    200,
		FreeCashFlow:       100,
		NetIncome:          100,
	}
	var r xbrl.ComputedRatios
	var v xbrl.ValuationMetrics
	calc.ComputeSolvency(&r, &v, &c)

	if r.CurrentRatio != 2.0 {
		t.Errorf("expected CurrentRatio=2.0, got %f", r.CurrentRatio)
	}
	if v.QuickRatio != 0.5 {
		t.Errorf("expected QuickRatio=0.5, got %f", v.QuickRatio)
	}
	if r.DebtToEquity != 0.5 {
		t.Errorf("expected DebtToEquity=0.5, got %f", r.DebtToEquity)
	}
	if r.NetDebt != 250 {
		t.Errorf("expected NetDebt=250, got %f", r.NetDebt)
	}
	if r.InterestCoverageRatio != 4.0 {
		t.Errorf("expected InterestCoverageRatio=4.0, got %f", r.InterestCoverageRatio)
	}
	if r.FCFConversionPct != 100.0 {
		t.Errorf("expected FCFConversionPct=100.0, got %f", r.FCFConversionPct)
	}
}

func TestComputeAltmanZScore(t *testing.T) {
	c := xbrl.CoreFinancials{
		CurrentAssets:      500,
		CurrentLiabilities: 300,
		TotalAssets:        1000,
		RetainedEarnings:   200,
		OperatingIncome:    150,
		TotalEquity:        600,
		TotalLiabilities:   400,
	}
	zScore := calc.ComputeAltmanZScore(&c)
	expected := 4.547
	if math.Abs(zScore-expected) > 0.001 {
		t.Errorf("expected Altman Z=%f, got %f", expected, zScore)
	}
}

func TestComputePiotroskiFScore(t *testing.T) {
	current := &xbrl.Statement{
		Core: xbrl.CoreFinancials{
			TotalAssets:       1000,
			Revenue:           1200,
			GrossProfit:       400,
			NetIncome:         100,
			OperatingCashFlow: 150,
			LongTermDebt:      100,
			CurrentAssets:     600,
			CurrentLiabilities: 300,
			SharesOutstanding: 1000000,
		},
		ComputedRatios: xbrl.ComputedRatios{
			ROA:            0.10,
			CurrentRatio:   2.0,
			GrossMarginPct: 33.33,
		},
	}
	prior := &xbrl.Statement{
		Core: xbrl.CoreFinancials{
			TotalAssets:       900,
			Revenue:           1000,
			GrossProfit:       300,
			NetIncome:         80,
			OperatingCashFlow: 100,
			LongTermDebt:      120,
			CurrentAssets:     500,
			CurrentLiabilities: 300,
			SharesOutstanding: 1000000,
		},
		ComputedRatios: xbrl.ComputedRatios{
			ROA:            0.088,
			CurrentRatio:   1.67,
			GrossMarginPct: 30.0,
		},
	}

	score := calc.ComputePiotroskiFScore(current, prior)
	if score != 9 {
		t.Errorf("expected Piotroski score=9, got %d", score)
	}
}

func TestGrahamAndMultiples(t *testing.T) {
	v := xbrl.ValuationMetrics{
		NormalizedEPS:  100,
		NormalizedBVPS: 500,
	}
	calc.ComputeGrahamFairValue(&v)
	if math.Abs(v.GrahamNumber-1060.66) > 0.01 {
		t.Errorf("expected Graham Number=1060.66, got %f", v.GrahamNumber)
	}

	c := xbrl.CoreFinancials{
		OperatingIncome:    200000,
		TotalDebt:          50000,
		CashAndEquivalents: 20000,
	}
	currentPrice := 800.0
	calc.ComputeValuationMultiples(&v, &c, currentPrice, 10000, 1.0)

	if v.PERatio != 8.0 {
		t.Errorf("expected PERatio=8.0, got %f", v.PERatio)
	}
	if v.PBRatio != 1.6 {
		t.Errorf("expected PBRatio=1.6, got %f", v.PBRatio)
	}
	if math.Abs(v.MarginOfSafetyPct-24.57) > 0.1 {
		t.Errorf("expected MOS=24.57, got %f", v.MarginOfSafetyPct)
	}
}

func TestCalculator_ForensicsAndValuation(t *testing.T) {
	curr := &xbrl.Statement{
		Ticker: "AADI",
		Metadata: xbrl.StatementMetadata{
			Currency:       "USD",
			ConversionRate: 15400.0,
		},
		Core: xbrl.CoreFinancials{
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

	err := calc.ComputeValuationAndRatios(curr, nil, 4150.0)
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

func TestCalculator_ZeroDebtAndParentNetIncome(t *testing.T) {
	prior := &xbrl.Statement{
		Ticker: "DEBTFREE",
		Core: xbrl.CoreFinancials{
			SharesOutstanding: 1000000,
			TotalAssets:       1000000000,
			LongTermDebt:      0,
			Revenue:           500000000,
		},
		ComputedRatios: xbrl.ComputedRatios{
			ROA:            0.05,
			CurrentRatio:   1.5,
			GrossMarginPct: 25.0,
		},
	}

	curr := &xbrl.Statement{
		Ticker: "DEBTFREE",
		Metadata: xbrl.StatementMetadata{
			Currency: "IDR",
		},
		Core: xbrl.CoreFinancials{
			SharesOutstanding:  1000000,
			TotalAssets:        1200000000,
			CashAndEquivalents: 300000000,
			CurrentAssets:      600000000,
			TotalLiabilities:   200000000,
			CurrentLiabilities: 200000000,
			LongTermDebt:       0,
			TotalDebt:          0,
			TotalEquity:        1000000000,
			RetainedEarnings:   800000000,
			Revenue:            700000000,
			GrossProfit:        350000000,
			OperatingIncome:    200000000,
			NetIncome:          180000000,
			NetIncomeParent:    170000000,
			OperatingCashFlow:  220000000,
			CapEx:              50000000,
			FreeCashFlow:       170000000,
		},
	}

	err := calc.ComputeValuationAndRatios(curr, prior, 1500.0)
	if err != nil {
		t.Fatalf("ComputeValuationAndRatios failed: %v", err)
	}

	expectedEPS := 170000000.0 / 1000000.0
	if curr.Valuation.NormalizedEPS != expectedEPS {
		t.Errorf("Expected NormalizedEPS %f from NetIncomeParent, got %f", expectedEPS, curr.Valuation.NormalizedEPS)
	}

	if curr.ComputedRatios.PiotroskiFScore != 9 {
		t.Errorf("Expected perfect Piotroski F-Score 9 for pristine debt-free compounder, got %d", curr.ComputedRatios.PiotroskiFScore)
	}
}

func TestCalculator_USDEPSNormalization(t *testing.T) {
	curr := &xbrl.Statement{
		Ticker: "USDFILE",
		Metadata: xbrl.StatementMetadata{
			Currency:       "USD",
			ConversionRate: 16000.0,
		},
		Core: xbrl.CoreFinancials{
			SharesOutstanding:  2000000000,
			TotalAssets:        4000000000,
			CashAndEquivalents: 500000000,
			CurrentAssets:      1000000000,
			TotalLiabilities:   1000000000,
			CurrentLiabilities: 500000000,
			TotalEquity:        3000000000,
			Revenue:            2000000000,
			GrossProfit:        600000000,
			OperatingIncome:    400000000,
			NetIncome:          300000000,
			OperatingCashFlow:  350000000,
			CapEx:              100000000,
			FreeCashFlow:       250000000,
		},
		Valuation: xbrl.ValuationMetrics{
			NormalizedEPS: 0.15,
		},
	}

	err := calc.ComputeValuationAndRatios(curr, nil, 3000.0)
	if err != nil {
		t.Fatalf("ComputeValuationAndRatios failed: %v", err)
	}

	expectedEPS := 2400.0
	if curr.Valuation.NormalizedEPS != expectedEPS {
		t.Errorf("Expected NormalizedEPS in IDR to be %f, got %f", expectedEPS, curr.Valuation.NormalizedEPS)
	}
	expectedBVPS := 24000.0
	if curr.Valuation.NormalizedBVPS != expectedBVPS {
		t.Errorf("Expected NormalizedBVPS in IDR to be %f, got %f", expectedBVPS, curr.Valuation.NormalizedBVPS)
	}
	expectedPE := 3000.0 / 2400.0
	if curr.Valuation.PERatio != expectedPE {
		t.Errorf("Expected PE ratio %f, got %f", expectedPE, curr.Valuation.PERatio)
	}
}

func TestApplyStockSplitAdjustment(t *testing.T) {
	stmt2021 := &xbrl.Statement{
		Core: xbrl.CoreFinancials{
			SharesOutstanding: 1000000,
			NetIncome:         100000000,
			TotalEquity:       500000000,
		},
	}
	stmt2024 := &xbrl.Statement{
		Core: xbrl.CoreFinancials{
			SharesOutstanding: 10000000, // 10:1 split
			NetIncome:         200000000,
			TotalEquity:       800000000,
		},
	}

	calc.ApplyStockSplitAdjustment([]*xbrl.Statement{stmt2021, stmt2024})

	if stmt2021.Valuation.NormalizedEPS != 10.0 {
		t.Errorf("expected split-adjusted EPS=10.0, got %f", stmt2021.Valuation.NormalizedEPS)
	}
}

func TestApplyStockSplitAdjustment_DSSA_10to1(t *testing.T) {
	preSplit := &xbrl.Statement{
		Ticker: "DSSA",
		Year:   2022,
		Period: "FY",
		Metadata: xbrl.StatementMetadata{
			Currency:       "USD",
			ConversionRate: 15500.0,
		},
		Core: xbrl.CoreFinancials{
			SharesOutstanding: 770552320,
			TotalEquity:       2000000000,
			NetIncome:         600000000,
		},
	}
	_ = calc.ComputeValuationAndRatios(preSplit, nil, 40000)

	postSplit := &xbrl.Statement{
		Ticker: "DSSA",
		Year:   2024,
		Period: "FY",
		Metadata: xbrl.StatementMetadata{
			Currency:       "USD",
			ConversionRate: 16000.0,
		},
		Core: xbrl.CoreFinancials{
			SharesOutstanding: 7705523200,
			TotalEquity:       2200000000,
			NetIncome:         500000000,
		},
	}
	_ = calc.ComputeValuationAndRatios(postSplit, nil, 4000)

	statements := []*xbrl.Statement{preSplit, postSplit}
	calc.ApplyStockSplitAdjustment(statements)

	expectedAdjEPS := (600000000.0 * 15500.0) / 7705523200.0
	if math.Abs(preSplit.Valuation.NormalizedEPS-expectedAdjEPS) > 1.0 {
		t.Errorf("Expected split-adjusted EPS %f, got %f", expectedAdjEPS, preSplit.Valuation.NormalizedEPS)
	}

	expectedAdjBVPS := (2000000000.0 * 15500.0) / 7705523200.0
	if math.Abs(preSplit.Valuation.NormalizedBVPS-expectedAdjBVPS) > 1.0 {
		t.Errorf("Expected split-adjusted BVPS %f, got %f", expectedAdjBVPS, preSplit.Valuation.NormalizedBVPS)
	}

	expectedAdjGraham := math.Sqrt(22.5 * expectedAdjEPS * expectedAdjBVPS)
	if math.Abs(preSplit.Valuation.GrahamNumber-expectedAdjGraham) > 5.0 {
		t.Errorf("Expected split-adjusted Graham Number %f, got %f", expectedAdjGraham, preSplit.Valuation.GrahamNumber)
	}
}

func TestComputeValuationBandsAndTiming(t *testing.T) {
	candles := []stock.PriceCandle{
		{Date: time.Now().AddDate(0, 0, -3), Close: 1000, High: 1020, Low: 990, Volume: 10000},
		{Date: time.Now().AddDate(0, 0, -2), Close: 1050, High: 1060, Low: 1000, Volume: 12000},
		{Date: time.Now().AddDate(0, 0, -1), Close: 1100, High: 1110, Low: 1040, Volume: 15000},
	}

	bands := calc.ComputeValuationBands(candles, 100.0, 500.0)
	if bands.MeanPE <= 0 || bands.MeanPB <= 0 {
		t.Errorf("expected positive MeanPE and MeanPB, got %f and %f", bands.MeanPE, bands.MeanPB)
	}

	signal := calc.ComputeTimingSignals(candles, bands, 100.0, 500.0)
	if signal.Status == "" {
		t.Errorf("expected non-empty timing status")
	}
}
