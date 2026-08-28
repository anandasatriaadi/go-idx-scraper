package calc

import (
	"math"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/stock"
	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/xbrl"
)

// ComputeValuationBands calculates historical P/E and P/B rolling series over price candles
// and computes mean, standard deviation, and +/- 1SD, 2SD bands in multiples and IDR prices.
func ComputeValuationBands(candles []stock.PriceCandle, eps float64, bvps float64) xbrl.ValuationBands {
	var bands xbrl.ValuationBands
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
		meanPE, stdDevPE := CalcMeanAndStdDev(peSeries)
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
		meanPB, stdDevPB := CalcMeanAndStdDev(pbSeries)
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

// CalcMeanAndStdDev calculates sample mean and sample standard deviation (N-1).
func CalcMeanAndStdDev(series []float64) (float64, float64) {
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
