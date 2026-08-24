package xbrl

import (
	"math"
	"testing"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/stock"
)

func TestComputeValuationBands(t *testing.T) {
	// Sample dataset: 5 prices with known mean and sample standard deviation
	// prices: [1000, 1200, 1400, 1600, 1800]
	// eps = 100, bvps = 500
	// PE: [10, 12, 14, 16, 18] -> Mean = 14, Variance = 40/4 = 10, SD = sqrt(10) ≈ 3.16227766
	// PB: [2.0, 2.4, 2.8, 3.2, 3.6] -> Mean = 2.8, SD = sqrt(10)/5 ≈ 0.632455532
	now := time.Now()
	prices := []float64{1000, 1200, 1400, 1600, 1800}
	var candles []stock.PriceCandle
	for i, p := range prices {
		candles = append(candles, stock.PriceCandle{
			Date:   now.AddDate(0, 0, i),
			Close:  p,
			Volume: 10000,
		})
	}

	eps := 100.0
	bvps := 500.0
	bands := ComputeValuationBands(candles, eps, bvps)

	expectedMeanPE := 14.0
	expectedStdDevPE := math.Sqrt(10.0)

	if math.Abs(bands.MeanPE-expectedMeanPE) > 1e-4 {
		t.Errorf("Expected MeanPE %f, got %f", expectedMeanPE, bands.MeanPE)
	}
	if math.Abs(bands.StdDevPE-expectedStdDevPE) > 1e-4 {
		t.Errorf("Expected StdDevPE %f, got %f", expectedStdDevPE, bands.StdDevPE)
	}

	expectedPlus1SD_PE := expectedMeanPE + expectedStdDevPE
	expectedMinus1SD_PE := expectedMeanPE - expectedStdDevPE
	expectedPlus2SD_PE := expectedMeanPE + 2.0*expectedStdDevPE
	expectedMinus2SD_PE := expectedMeanPE - 2.0*expectedStdDevPE

	if math.Abs(bands.Plus1SD_PE-expectedPlus1SD_PE) > 1e-4 {
		t.Errorf("Expected Plus1SD_PE %f, got %f", expectedPlus1SD_PE, bands.Plus1SD_PE)
	}
	if math.Abs(bands.Minus1SD_PE-expectedMinus1SD_PE) > 1e-4 {
		t.Errorf("Expected Minus1SD_PE %f, got %f", expectedMinus1SD_PE, bands.Minus1SD_PE)
	}
	if math.Abs(bands.Plus2SD_PE-expectedPlus2SD_PE) > 1e-4 {
		t.Errorf("Expected Plus2SD_PE %f, got %f", expectedPlus2SD_PE, bands.Plus2SD_PE)
	}
	if math.Abs(bands.Minus2SD_PE-expectedMinus2SD_PE) > 1e-4 {
		t.Errorf("Expected Minus2SD_PE %f, got %f", expectedMinus2SD_PE, bands.Minus2SD_PE)
	}

	// Price bands for PE
	if math.Abs(bands.MeanPrice_PE-(expectedMeanPE*eps)) > 1e-4 {
		t.Errorf("Expected MeanPrice_PE %f, got %f", expectedMeanPE*eps, bands.MeanPrice_PE)
	}
	if math.Abs(bands.Minus1SDPrice_PE-(expectedMinus1SD_PE*eps)) > 1e-4 {
		t.Errorf("Expected Minus1SDPrice_PE %f, got %f", expectedMinus1SD_PE*eps, bands.Minus1SDPrice_PE)
	}
	if math.Abs(bands.Minus2SDPrice_PE-(expectedMinus2SD_PE*eps)) > 1e-4 {
		t.Errorf("Expected Minus2SDPrice_PE %f, got %f", expectedMinus2SD_PE*eps, bands.Minus2SDPrice_PE)
	}

	// Price bands for PB
	expectedMeanPB := 2.8
	expectedStdDevPB := expectedStdDevPE / 5.0
	expectedMinus1SD_PB := expectedMeanPB - expectedStdDevPB
	if math.Abs(bands.MeanPB-expectedMeanPB) > 1e-4 {
		t.Errorf("Expected MeanPB %f, got %f", expectedMeanPB, bands.MeanPB)
	}
	if math.Abs(bands.StdDevPB-expectedStdDevPB) > 1e-4 {
		t.Errorf("Expected StdDevPB %f, got %f", expectedStdDevPB, bands.StdDevPB)
	}
	if math.Abs(bands.Minus1SDPrice_PB-(expectedMinus1SD_PB*bvps)) > 1e-4 {
		t.Errorf("Expected Minus1SDPrice_PB %f, got %f", expectedMinus1SD_PB*bvps, bands.Minus1SDPrice_PB)
	}
}

func TestComputeValuationBands_EmptyAndEdgeCases(t *testing.T) {
	// Empty slice
	bandsEmpty := ComputeValuationBands(nil, 100, 500)
	if bandsEmpty.MeanPE != 0 || bandsEmpty.StdDevPE != 0 {
		t.Errorf("Expected zeros for empty candles, got %v", bandsEmpty)
	}

	// Single candle
	singleCandle := []stock.PriceCandle{
		{Date: time.Now(), Close: 1500, Volume: 1000},
	}
	bandsSingle := ComputeValuationBands(singleCandle, 100, 500)
	if bandsSingle.MeanPE != 15.0 || bandsSingle.StdDevPE != 0.0 {
		t.Errorf("Expected MeanPE 15.0 and StdDev 0.0, got MeanPE %f, StdDev %f", bandsSingle.MeanPE, bandsSingle.StdDevPE)
	}
	if bandsSingle.MeanPrice_PE != 1500.0 {
		t.Errorf("Expected MeanPrice_PE 1500.0, got %f", bandsSingle.MeanPrice_PE)
	}

	// Negative or zero EPS/BVPS
	bandsNoEPS := ComputeValuationBands(singleCandle, 0, 0)
	if bandsNoEPS.MeanPE != 0 || bandsNoEPS.MeanPB != 0 {
		t.Errorf("Expected zeros when EPS and BVPS are zero, got %v", bandsNoEPS)
	}
}

func TestComputeTimingSignals_RSI_Oversold(t *testing.T) {
	// 25 candles with steady price decrease to trigger oversold RSI (< 35)
	now := time.Now()
	var candles []stock.PriceCandle
	price := 5000.0
	for i := 0; i < 25; i++ {
		price -= 100.0
		candles = append(candles, stock.PriceCandle{
			Date:   now.AddDate(0, 0, i),
			Open:   price + 20,
			High:   price + 30,
			Low:    price - 20,
			Close:  price,
			Volume: 50000,
		})
	}

	bands := ValuationBands{
		MeanPrice_PE:     4000,
		Minus1SDPrice_PE: 3200,
		Minus2SDPrice_PE: 2800,
	}

	signal := ComputeTimingSignals(candles, bands, 200, 1000)

	if signal.RSI >= 35.0 {
		t.Errorf("Expected RSI < 35.0 (oversold), got %f", signal.RSI)
	}
	if signal.Score < 15 {
		t.Errorf("Expected Score >= 15 for oversold RSI, got %d", signal.Score)
	}
}

func TestComputeTimingSignals_RSI_BullishDivergence(t *testing.T) {
	// Create a sequence that triggers regular bullish divergence:
	// 1. Initial 15 days: Sharp dump from 3000 to 1800 (RSI reaches deep oversold < 30)
	// 2. Next 10 days: Rebound from 1800 to 2400 (RSI recovers to ~50)
	// 3. Next 12 days: Slow drift down to 1700 (new lower price low < 1800), but RSI stays at ~33 (higher than first trough < 30)
	now := time.Now()
	var candles []stock.PriceCandle

	// Phase 1: Sharp dump (days 0-14)
	p := 3000.0
	for i := 0; i < 15; i++ {
		p -= 80.0
		candles = append(candles, stock.PriceCandle{
			Date:   now.AddDate(0, 0, len(candles)),
			Open:   p + 10,
			High:   p + 15,
			Low:    p - 10,
			Close:  p,
			Volume: 100000,
		})
	}

	// Phase 2: Strong rebound (days 15-24)
	for i := 0; i < 10; i++ {
		p += 60.0
		candles = append(candles, stock.PriceCandle{
			Date:   now.AddDate(0, 0, len(candles)),
			Open:   p - 10,
			High:   p + 15,
			Low:    p - 15,
			Close:  p,
			Volume: 60000,
		})
	}

	// Phase 3: Slower drift to lower price (days 25-40)
	// Drops gradually with pullbacks to 1750 (< 1800)
	priceDeltas := []float64{-80, +30, -90, +30, -100, +35, -100, +35, -110, +40, -110, +40, -120, +40, -120, -60}
	for _, delta := range priceDeltas {
		p += delta
		candles = append(candles, stock.PriceCandle{
			Date:   now.AddDate(0, 0, len(candles)),
			Open:   p + 5,
			High:   p + 10,
			Low:    p - 10,
			Close:  p,
			Volume: 30000,
		})
	}

	bands := ValuationBands{
		MeanPrice_PE:     2500,
		Minus1SDPrice_PE: 2000,
		Minus2SDPrice_PE: 1600,
	}

	signal := ComputeTimingSignals(candles, bands, 150, 800)

	if !signal.RSIBullishDivergence {
		t.Errorf("Expected RSI Bullish Divergence to be detected, got false (RSI: %f)", signal.RSI)
	}
	if signal.Score < 25 {
		t.Errorf("Expected Score >= 25 with Bullish Divergence, got %d", signal.Score)
	}
}

func TestComputeTimingSignals_VSA_StoppingVolume(t *testing.T) {
	// 20 baseline days with volume 10,000
	now := time.Now()
	var candles []stock.PriceCandle
	for i := 0; i < 20; i++ {
		candles = append(candles, stock.PriceCandle{
			Date:   now.AddDate(0, 0, i),
			Open:   2000,
			High:   2050,
			Low:    1950,
			Close:  2000,
			Volume: 10000,
		})
	}

	// Day 21: Down day with high volume absorption (RVOL = 2.5 >= 1.8, CLV = (2*1960 - 2000 - 1900)/100 = 20/100 = +0.20 >= 0)
	candles = append(candles, stock.PriceCandle{
		Date:   now.AddDate(0, 0, 20),
		Open:   2000,
		High:   2000,
		Low:    1900,
		Close:  1960,
		Volume: 25000,
	})

	bands := ValuationBands{}
	signal := ComputeTimingSignals(candles, bands, 0, 0)

	if !signal.StoppingVolume {
		t.Errorf("Expected StoppingVolume to be true, got false (RVOL: %f, CLV: %f)", signal.RVOL, signal.CLV)
	}
	if signal.RVOL < 1.8 {
		t.Errorf("Expected RVOL >= 1.8, got %f", signal.RVOL)
	}
	if signal.CLV < 0.0 {
		t.Errorf("Expected CLV >= 0.0, got %f", signal.CLV)
	}
	if signal.Score < 25 {
		t.Errorf("Expected Score >= 25 for Stopping Volume, got %d", signal.Score)
	}
}

func TestComputeTimingSignals_VSA_VolumeDryUp(t *testing.T) {
	// 20 candles: Days 0-14 volume = 20,000; Days 15-19 volume = 5,000
	// 20d SMA volume = (15*20000 + 5*5000)/20 = 16,250
	// 5d Mean volume = 5,000 -> VDU = 5000/16250 ≈ 0.31 <= 0.50
	now := time.Now()
	var candles []stock.PriceCandle
	for i := 0; i < 15; i++ {
		candles = append(candles, stock.PriceCandle{
			Date:   now.AddDate(0, 0, i),
			Open:   2000,
			High:   2050,
			Low:    1950,
			Close:  2000,
			Volume: 20000,
		})
	}
	for i := 15; i < 20; i++ {
		candles = append(candles, stock.PriceCandle{
			Date:   now.AddDate(0, 0, i),
			Open:   2000,
			High:   2050,
			Low:    1950,
			Close:  2000,
			Volume: 5000,
		})
	}

	bands := ValuationBands{}
	signal := ComputeTimingSignals(candles, bands, 0, 0)

	if !signal.VolumeDryUp {
		t.Errorf("Expected VolumeDryUp to be true, got false (VDU: %f)", signal.VDU)
	}
	if signal.VDU > 0.50 {
		t.Errorf("Expected VDU <= 0.50, got %f", signal.VDU)
	}
	if signal.Score < 15 {
		t.Errorf("Expected Score >= 15 for Volume Dry-Up, got %d", signal.Score)
	}
}

func TestComputeTimingSignals_ScoreTiersAndStatus(t *testing.T) {
	now := time.Now()

	// 1. Actionable Buy Setup (Score >= 70)
	// Deep Value (<= -2SD: 35 pts) + Stopping Volume (25 pts) + Volume Dry-Up (15 pts) = 75 pts
	var candlesBuySetup []stock.PriceCandle
	for i := 0; i < 15; i++ {
		candlesBuySetup = append(candlesBuySetup, stock.PriceCandle{
			Date:   now.AddDate(0, 0, i),
			Open:   1200,
			High:   1220,
			Low:    1180,
			Close:  1200,
			Volume: 20000,
		})
	}
	for i := 15; i < 19; i++ {
		candlesBuySetup = append(candlesBuySetup, stock.PriceCandle{
			Date:   now.AddDate(0, 0, i),
			Open:   1000,
			High:   1020,
			Low:    980,
			Close:  1000,
			Volume: 2000,
		})
	}
	// Last bar: Down day with high volume spike and upper half close (CLV >= 0)
	// 5d volume average = (2000*4 + 18000)/5 = 26000/5 = 5200
	// 20d SMA volume = (15*20000 + 4*2000 + 18000)/20 = 326000/20 = 16300
	// RVOL = 18000 / 16300 = 1.10 -> let's increase last bar volume to 35000 -> RVOL = 35000/17150 = 2.04
	// 5d avg = (8000 + 35000)/5 = 8600 / 17150 = 0.501 -> let's make prior 4 bars 1000 vol each:
	// Prior 4 bars = 1000 each -> (4000 + 32000)/5 = 7200; 20d SMA = (15*20000 + 4*1000 + 32000)/20 = 336000/20 = 16800.
	// VDU = 7200 / 16800 = 0.428 <= 0.50 (VDU true!)
	// RVOL = 32000 / 16800 = 1.90 >= 1.8 (Stopping volume true!)
	candlesBuySetup = make([]stock.PriceCandle, 0)
	for i := 0; i < 15; i++ {
		candlesBuySetup = append(candlesBuySetup, stock.PriceCandle{
			Date:   now.AddDate(0, 0, i),
			Open:   1200,
			High:   1220,
			Low:    1180,
			Close:  1200,
			Volume: 20000,
		})
	}
	for i := 15; i < 19; i++ {
		candlesBuySetup = append(candlesBuySetup, stock.PriceCandle{
			Date:   now.AddDate(0, 0, i),
			Open:   1000,
			High:   1020,
			Low:    980,
			Close:  1000,
			Volume: 1000,
		})
	}
	candlesBuySetup = append(candlesBuySetup, stock.PriceCandle{
		Date:   now.AddDate(0, 0, 19),
		Open:   1000,
		High:   1000,
		Low:    900,
		Close:  960, // Down day (Open 1000, Close 960), CLV = (2*960 - 1000 - 900)/100 = 20/100 = +0.20
		Volume: 32000,
	})

	bandsDeepVal := ValuationBands{
		Minus2SDPrice_PE: 1000, // Current price 960 <= 1000 (-2SD: +35 pts)
		Minus1SDPrice_PE: 1100,
		MeanPrice_PE:     1300,
	}

	sigBuy := ComputeTimingSignals(candlesBuySetup, bandsDeepVal, 100, 500)
	if sigBuy.Score < 70 {
		t.Errorf("Expected Score >= 70, got %d (Signals: %v)", sigBuy.Score, sigBuy.Signals)
	}
	expectedStatusBuy := "Actionable Buy Setup (Deep Value + Accumulation)"
	if sigBuy.Status != expectedStatusBuy {
		t.Errorf("Expected status '%s', got '%s'", expectedStatusBuy, sigBuy.Status)
	}

	// 2. Value Accumulation Zone (Score 45 - 69)
	// Value Discount (-1SD: +20) + VDU (+15) + RSI Oversold (+15) = 50 pts
	bandsValueDiscount := ValuationBands{
		Minus2SDPrice_PE: 900,
		Minus1SDPrice_PE: 1050, // Current price 960 <= 1050 (-1SD: +20 pts)
		MeanPrice_PE:     1300,
	}
	sigAccum := ComputeTimingSignals(candlesBuySetup, bandsValueDiscount, 100, 500)
	if sigAccum.Score < 45 || sigAccum.Score >= 70 {
		t.Errorf("Expected Score between 45 and 69, got %d", sigAccum.Score)
	}
	if sigAccum.Status != "Value Accumulation Zone" {
		t.Errorf("Expected status 'Value Accumulation Zone', got '%s'", sigAccum.Status)
	}

	// 3. Overextended (> +1SD Band)
	var candlesOverextended []stock.PriceCandle
	for i := 0; i < 20; i++ {
		candlesOverextended = append(candlesOverextended, stock.PriceCandle{
			Date:   now.AddDate(0, 0, i),
			Open:   2000,
			High:   2050,
			Low:    1950,
			Close:  2000,
			Volume: 10000,
		})
	}
	bandsOverextended := ValuationBands{
		MeanPrice_PE:    1500,
		Plus1SDPrice_PE: 1800, // Current price 2000 > 1800 (+1SD)
		Plus2SDPrice_PE: 2200,
	}
	sigOver := ComputeTimingSignals(candlesOverextended, bandsOverextended, 100, 500)
	if sigOver.Status != "Fair Value / Overextended" {
		t.Errorf("Expected status 'Fair Value / Overextended', got '%s'", sigOver.Status)
	}

	// 4. Neutral
	bandsNeutral := ValuationBands{
		MeanPrice_PE:     2000,
		Minus1SDPrice_PE: 1800,
		Plus1SDPrice_PE:  2200,
	}
	sigNeutral := ComputeTimingSignals(candlesOverextended, bandsNeutral, 100, 500)
	if sigNeutral.Status != "Neutral / Waiting for Catalyst" {
		t.Errorf("Expected status 'Neutral / Waiting for Catalyst', got '%s'", sigNeutral.Status)
	}
}
