package xbrl

import (
	"math"
	"sort"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/stock"
)

// ComputeValuationBands calculates historical P/E and P/B rolling series over price candles
// and computes mean, standard deviation, and +/- 1SD, 2SD bands in multiples and IDR prices.
func ComputeValuationBands(candles []stock.PriceCandle, eps float64, bvps float64) ValuationBands {
	var bands ValuationBands
	if len(candles) == 0 {
		return bands
	}

	var peSeries []float64
	var pbSeries []float64

	for _, c := range candles {
		if c.Close <= 0 {
			continue
		}
		if eps > 0 {
			peSeries = append(peSeries, c.Close/eps)
		}
		if bvps > 0 {
			pbSeries = append(pbSeries, c.Close/bvps)
		}
	}

	if len(peSeries) > 0 {
		meanPE, stdDevPE := calcMeanAndStdDev(peSeries)
		bands.MeanPE = meanPE
		bands.StdDevPE = stdDevPE
		bands.Plus2SD_PE = meanPE + 2.0*stdDevPE
		bands.Plus1SD_PE = meanPE + 1.0*stdDevPE
		bands.Minus1SD_PE = meanPE - 1.0*stdDevPE
		bands.Minus2SD_PE = meanPE - 2.0*stdDevPE

		if eps > 0 {
			bands.MeanPrice_PE = meanPE * eps
			bands.Plus2SDPrice_PE = bands.Plus2SD_PE * eps
			bands.Plus1SDPrice_PE = bands.Plus1SD_PE * eps
			bands.Minus1SDPrice_PE = bands.Minus1SD_PE * eps
			bands.Minus2SDPrice_PE = bands.Minus2SD_PE * eps
		}
	}

	if len(pbSeries) > 0 {
		meanPB, stdDevPB := calcMeanAndStdDev(pbSeries)
		bands.MeanPB = meanPB
		bands.StdDevPB = stdDevPB
		bands.Plus2SD_PB = meanPB + 2.0*stdDevPB
		bands.Plus1SD_PB = meanPB + 1.0*stdDevPB
		bands.Minus1SD_PB = meanPB - 1.0*stdDevPB
		bands.Minus2SD_PB = meanPB - 2.0*stdDevPB

		if bvps > 0 {
			bands.MeanPrice_PB = meanPB * bvps
			bands.Plus2SDPrice_PB = bands.Plus2SD_PB * bvps
			bands.Plus1SDPrice_PB = bands.Plus1SD_PB * bvps
			bands.Minus1SDPrice_PB = bands.Minus1SD_PB * bvps
			bands.Minus2SDPrice_PB = bands.Minus2SD_PB * bvps
		}
	}

	return bands
}

// ComputeTimingSignals computes RSI(14), RSI Bullish Divergence, Volume Spread Analysis (RVOL20, CLV, Stopping Volume, VDU5),
// and synthesizes the Smart Timing Score (0-100) with status classification.
func ComputeTimingSignals(candles []stock.PriceCandle, bands ValuationBands, eps float64, bvps float64) TimingSignal {
	var signal TimingSignal
	n := len(candles)
	if n == 0 {
		return signal
	}

	// Sort chronological (oldest to newest)
	sorted := make([]stock.PriceCandle, n)
	copy(sorted, candles)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Date.Before(sorted[j].Date)
	})

	latest := sorted[n-1]
	currentPrice := latest.Close

	// 1. RSI(14) calculation
	rsiSeries := computeRSISeries(sorted, 14)
	currentRSI := 50.0
	if len(rsiSeries) > 0 {
		currentRSI = rsiSeries[len(rsiSeries)-1]
	}
	signal.RSI = math.Round(currentRSI*100) / 100

	// 2. RSI Bullish Divergence Detection
	bullishDivergence := detectRSIBullishDivergence(sorted, rsiSeries)
	signal.RSIBullishDivergence = bullishDivergence

	// 3. Volume Spread Analysis (VSA)
	// 3a. RVOL20 (Relative Volume vs 20-day SMA of Volume)
	vWindow := 20
	if n < vWindow {
		vWindow = n
	}
	var sumVol float64
	for i := n - vWindow; i < n; i++ {
		sumVol += float64(sorted[i].Volume)
	}
	smaVol := sumVol / float64(vWindow)

	rvol := 1.0
	if smaVol > 0 {
		rvol = float64(latest.Volume) / smaVol
	}
	signal.RVOL = math.Round(rvol*100) / 100

	// 3b. CLV (Close Location Value) = ((2 * Close) - High - Low) / (High - Low)
	spread := latest.High - latest.Low
	clv := 0.0
	if spread > 0 {
		clv = ((2.0 * latest.Close) - latest.High - latest.Low) / spread
	}
	if clv > 1.0 {
		clv = 1.0
	} else if clv < -1.0 {
		clv = -1.0
	}
	signal.CLV = math.Round(clv*100) / 100

	// 3c. Stopping Volume: RVOL >= 1.8 AND CLV >= 0.0 on a down day
	isDownDay := false
	if n >= 2 {
		prev := sorted[n-2]
		if latest.Close <= prev.Close || latest.Close <= latest.Open {
			isDownDay = true
		}
	} else if latest.Close <= latest.Open {
		isDownDay = true
	}

	stoppingVolume := isDownDay && (rvol >= 1.8) && (clv >= 0.0)
	signal.StoppingVolume = stoppingVolume

	// 3d. Volume Dry-Up (VDU5): 5-day mean volume / 20-day SMA volume <= 0.50
	vduWindow := 5
	if n < vduWindow {
		vduWindow = n
	}
	var sum5Vol float64
	for i := n - vduWindow; i < n; i++ {
		sum5Vol += float64(sorted[i].Volume)
	}
	mean5Vol := sum5Vol / float64(vduWindow)

	vdu := 1.0
	if smaVol > 0 {
		vdu = mean5Vol / smaVol
	}
	signal.VDU = math.Round(vdu*100) / 100
	volumeDryUp := vdu <= 0.50
	signal.VolumeDryUp = volumeDryUp

	// 4. Smart Timing Score Synthesis (0 - 100)
	score := 0
	var signals []string

	// Factor 1: Valuation Band Discount (Max 35 pts)
	discountZone := "Mean"
	isDeepValue := false
	isValueDiscount := false
	isOverextended := false
	isExtremelyOverextended := false

	// Check against PE bands if valid
	if bands.Minus2SDPrice_PE > 0 && currentPrice <= bands.Minus2SDPrice_PE {
		isDeepValue = true
	} else if bands.Minus1SDPrice_PE > 0 && currentPrice <= bands.Minus1SDPrice_PE {
		isValueDiscount = true
	} else if bands.Plus2SDPrice_PE > 0 && currentPrice > bands.Plus2SDPrice_PE {
		isExtremelyOverextended = true
	} else if bands.Plus1SDPrice_PE > 0 && currentPrice > bands.Plus1SDPrice_PE {
		isOverextended = true
	}

	// Check against PB bands if valid
	if bands.Minus2SDPrice_PB > 0 && currentPrice <= bands.Minus2SDPrice_PB {
		isDeepValue = true
		isValueDiscount = false
	} else if !isDeepValue && bands.Minus1SDPrice_PB > 0 && currentPrice <= bands.Minus1SDPrice_PB {
		isValueDiscount = true
	} else if !isDeepValue && !isValueDiscount && bands.Plus2SDPrice_PB > 0 && currentPrice > bands.Plus2SDPrice_PB {
		isExtremelyOverextended = true
	} else if !isDeepValue && !isValueDiscount && !isExtremelyOverextended && bands.Plus1SDPrice_PB > 0 && currentPrice > bands.Plus1SDPrice_PB {
		isOverextended = true
	}

	if isDeepValue {
		score += 35
		discountZone = "-2SD"
		signals = append(signals, "Valuation: Deep Value (Price <= -2.0 SD Band)")
	} else if isValueDiscount {
		score += 20
		discountZone = "-1SD"
		signals = append(signals, "Valuation: Value Discount (Price <= -1.0 SD Band)")
	} else if isExtremelyOverextended {
		discountZone = "+2SD"
		signals = append(signals, "Valuation: Extremely Overextended (Price > +2.0 SD Band)")
	} else if isOverextended {
		discountZone = "+1SD"
		signals = append(signals, "Valuation: Overextended (Price > +1.0 SD Band)")
	}
	signal.ValuationDiscountZone = discountZone

	// Factor 2: RSI Bullish Divergence / Oversold (Max 25 pts)
	if bullishDivergence {
		score += 25
		signals = append(signals, "Momentum: RSI Bullish Divergence (Oversold Reversal)")
	} else if currentRSI < 35.0 && currentRSI > 0 {
		score += 15
		signals = append(signals, "Momentum: RSI Oversold (< 35)")
	}

	// Factor 3: VSA Stopping Volume / Smart Money Absorption (Max 25 pts)
	if stoppingVolume {
		score += 25
		signals = append(signals, "VSA: Stopping Volume / Smart Money Absorption (High RVOL + High CLV)")
	}

	// Factor 4: Volume Dry-Up / Base Tightening (Max 15 pts)
	if volumeDryUp {
		score += 15
		signals = append(signals, "VSA: Volume Dry-Up / Base Tightening (5d Volume <= 50% 20d SMA)")
	}

	if score > 100 {
		score = 100
	}
	signal.Score = score
	signal.Signals = signals

	// 5. Status Classification
	if score >= 70 {
		signal.Status = "Actionable Buy Setup (Deep Value + Accumulation)"
	} else if score >= 45 {
		signal.Status = "Value Accumulation Zone"
	} else if isOverextended || isExtremelyOverextended {
		signal.Status = "Fair Value / Overextended"
	} else {
		signal.Status = "Neutral / Waiting for Catalyst"
	}

	return signal
}

// calcMeanAndStdDev calculates sample mean and sample standard deviation (N-1)
func calcMeanAndStdDev(series []float64) (float64, float64) {
	n := len(series)
	if n == 0 {
		return 0, 0
	}
	if n == 1 {
		return series[0], 0
	}

	var sum float64
	for _, v := range series {
		sum += v
	}
	mean := sum / float64(n)

	var varianceSum float64
	for _, v := range series {
		diff := v - mean
		varianceSum += diff * diff
	}

	variance := varianceSum / float64(n-1)
	stdDev := math.Sqrt(variance)

	return mean, stdDev
}

// computeRSISeries computes Wilder's RSI series for given candle series
func computeRSISeries(candles []stock.PriceCandle, period int) []float64 {
	n := len(candles)
	if n <= 1 {
		return nil
	}

	if n <= period {
		// Not enough bars for full period, compute simple gain/loss over available
		var gainSum, lossSum float64
		for i := 1; i < n; i++ {
			diff := candles[i].Close - candles[i-1].Close
			if diff > 0 {
				gainSum += diff
			} else {
				lossSum -= diff
			}
		}
		if lossSum == 0 {
			if gainSum == 0 {
				return []float64{50.0}
			}
			return []float64{100.0}
		}
		rs := gainSum / lossSum
		rsi := 100.0 - (100.0 / (1.0 + rs))
		return []float64{rsi}
	}

	rsiSeries := make([]float64, 0, n-period)
	var gainSum, lossSum float64

	// First period initial simple average
	for i := 1; i <= period; i++ {
		diff := candles[i].Close - candles[i-1].Close
		if diff > 0 {
			gainSum += diff
		} else {
			lossSum -= diff
		}
	}

	avgGain := gainSum / float64(period)
	avgLoss := lossSum / float64(period)

	var firstRSI float64
	if avgLoss == 0 {
		if avgGain == 0 {
			firstRSI = 50.0
		} else {
			firstRSI = 100.0
		}
	} else {
		rs := avgGain / avgLoss
		firstRSI = 100.0 - (100.0 / (1.0 + rs))
	}
	rsiSeries = append(rsiSeries, firstRSI)

	// Subsequent smoothed Wilder's RSI
	pFloat := float64(period)
	for i := period + 1; i < n; i++ {
		diff := candles[i].Close - candles[i-1].Close
		gain := 0.0
		loss := 0.0
		if diff > 0 {
			gain = diff
		} else {
			loss = -diff
		}

		avgGain = ((avgGain * (pFloat - 1.0)) + gain) / pFloat
		avgLoss = ((avgLoss * (pFloat - 1.0)) + loss) / pFloat

		var rsi float64
		if avgLoss == 0 {
			if avgGain == 0 {
				rsi = 50.0
			} else {
				rsi = 100.0
			}
		} else {
			rs := avgGain / avgLoss
			rsi = 100.0 - (100.0 / (1.0 + rs))
		}
		rsiSeries = append(rsiSeries, rsi)
	}

	return rsiSeries
}

// detectRSIBullishDivergence identifies if price makes a lower low over the last 20-40 days
// while RSI makes a higher low from an oversold level (< 35).
func detectRSIBullishDivergence(candles []stock.PriceCandle, rsiSeries []float64) bool {
	n := len(candles)
	rsiLen := len(rsiSeries)
	if n < 20 || rsiLen == 0 {
		return false
	}

	// rsiSeries index i corresponds to candles[i + (n - rsiLen)]
	offset := n - rsiLen

	lookback := 40
	if lookback > n {
		lookback = n
	}
	startCandleIdx := n - lookback
	if startCandleIdx < offset {
		startCandleIdx = offset
	}

	type Trough struct {
		CandleIdx int
		Price     float64
		RSI       float64
	}

	var troughs []Trough

	// 1. Find local price troughs in the lookback window
	for cIdx := startCandleIdx; cIdx < n-1; cIdx++ {
		if cIdx <= 0 {
			continue
		}
		rIdx := cIdx - offset
		if rIdx < 0 || rIdx >= rsiLen {
			continue
		}

		isTrough := false
		if cIdx == startCandleIdx {
			if candles[cIdx].Close <= candles[cIdx+1].Close {
				isTrough = true
			}
		} else {
			if candles[cIdx].Close <= candles[cIdx-1].Close && candles[cIdx].Close <= candles[cIdx+1].Close {
				isTrough = true
			}
		}

		if isTrough {
			troughs = append(troughs, Trough{
				CandleIdx: cIdx,
				Price:     candles[cIdx].Close,
				RSI:       rsiSeries[rIdx],
			})
		}
	}

	// 2. Also check the absolute RSI minimum in the earlier half as a key swing trough
	minRSI := 100.0
	minRSIIdx := -1
	earlyEndIdx := n - 10
	for cIdx := startCandleIdx; cIdx < earlyEndIdx; cIdx++ {
		rIdx := cIdx - offset
		if rIdx >= 0 && rIdx < rsiLen {
			if rsiSeries[rIdx] < minRSI {
				minRSI = rsiSeries[rIdx]
				minRSIIdx = cIdx
			}
		}
	}
	if minRSIIdx >= 0 && minRSI < 35.0 {
		found := false
		for _, t := range troughs {
			if t.CandleIdx == minRSIIdx {
				found = true
				break
			}
		}
		if !found {
			troughs = append(troughs, Trough{
				CandleIdx: minRSIIdx,
				Price:     candles[minRSIIdx].Close,
				RSI:       minRSI,
			})
		}
	}

	// 3. Current bar as recent candidate
	latestCIdx := n - 1
	latestRIdx := rsiLen - 1
	currentCandidate := Trough{
		CandleIdx: latestCIdx,
		Price:     candles[latestCIdx].Close,
		RSI:       rsiSeries[latestRIdx],
	}

	allRecent := append(troughs, currentCandidate)

	// 4. Compare all pairs of troughs
	for _, t1 := range troughs {
		// Prior trough must have reached oversold (< 35)
		if t1.RSI >= 35.0 {
			continue
		}

		for _, t2 := range allRecent {
			// t2 must be at least 5 bars after t1
			if t2.CandleIdx-t1.CandleIdx < 5 {
				continue
			}
			// t2 must be within the recent 15 bars
			if n-1-t2.CandleIdx > 15 {
				continue
			}

			// Bullish Divergence: Lower price low, Higher RSI low
			if t2.Price < t1.Price && t2.RSI > t1.RSI {
				return true
			}
		}
	}

	return false
}
