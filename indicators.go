package charts

import (
	"fmt"
	"math"
)

// CalculateSMA calculates Simple Moving Average
func CalculateSMA(values []float64, period int) []float64 {
	if period <= 0 || period > len(values) {
		return nil
	}

	result := make([]float64, len(values))

	// First few values are null until we have enough data
	for i := 0; i < period-1; i++ {
		result[i] = GetNullValue()
	}

	// Calculate SMA for each position
	for i := period - 1; i < len(values); i++ {
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += values[j]
		}
		result[i] = sum / float64(period)
	}

	return result
}

// CalculateEMA calculates Exponential Moving Average
func CalculateEMA(values []float64, period int) []float64 {
	if period <= 0 || len(values) == 0 {
		return nil
	}

	result := make([]float64, len(values))
	multiplier := 2.0 / (float64(period) + 1.0)

	// First value is the same as input
	result[0] = values[0]

	// Calculate EMA for each subsequent value
	for i := 1; i < len(values); i++ {
		result[i] = (values[i] * multiplier) + (result[i-1] * (1 - multiplier))
	}

	return result
}

// BollingerBands represents upper band, middle line (SMA), and lower band
type BollingerBands struct {
	Upper  []float64
	Middle []float64
	Lower  []float64
}

// CalculateBollingerBands calculates Bollinger Bands (SMA ± (stddev * multiplier))
func CalculateBollingerBands(values []float64, period int, multiplier float64) BollingerBands {
	if period <= 0 || period > len(values) {
		return BollingerBands{}
	}

	middle := CalculateSMA(values, period)
	upper := make([]float64, len(values))
	lower := make([]float64, len(values))

	for i := 0; i < len(values); i++ {
		if i < period-1 {
			upper[i] = GetNullValue()
			lower[i] = GetNullValue()
			continue
		}

		// Calculate standard deviation for this period
		variance := 0.0
		mean := middle[i]
		for j := i - period + 1; j <= i; j++ {
			diff := values[j] - mean
			variance += diff * diff
		}
		stddev := math.Sqrt(variance / float64(period))

		upper[i] = mean + (stddev * multiplier)
		lower[i] = mean - (stddev * multiplier)
	}

	return BollingerBands{
		Upper:  upper,
		Middle: middle,
		Lower:  lower,
	}
}

// CalculateRSI calculates Relative Strength Index
func CalculateRSI(values []float64, period int) []float64 {
	if period <= 0 || len(values) < period+1 {
		return nil
	}

	result := make([]float64, len(values))

	// Initialize first values as null
	for i := 0; i < period; i++ {
		result[i] = GetNullValue()
	}

	// Calculate price changes
	gains := make([]float64, len(values)-1)
	losses := make([]float64, len(values)-1)

	for i := 1; i < len(values); i++ {
		change := values[i] - values[i-1]
		if change > 0 {
			gains[i-1] = change
			losses[i-1] = 0
		} else {
			gains[i-1] = 0
			losses[i-1] = -change
		}
	}

	// Calculate initial averages
	avgGain := 0.0
	avgLoss := 0.0
	for i := 0; i < period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	// Calculate RSI
	for i := period; i < len(values); i++ {
		if avgLoss == 0 {
			result[i] = 100
		} else {
			rs := avgGain / avgLoss
			result[i] = 100 - (100 / (1 + rs))
		}

		// Update averages for next iteration
		if i < len(gains) {
			avgGain = ((avgGain * float64(period-1)) + gains[i]) / float64(period)
			avgLoss = ((avgLoss * float64(period-1)) + losses[i]) / float64(period)
		}
	}

	return result
}

// AddSMAToKlines adds Simple Moving Average as a line series to chart data
func AddSMAToKlines(klinesData CandlestickSeries, period int, color Color) LineSeriesList {
	closes := ExtractClosePrices(klinesData)
	smaValues := CalculateSMA(closes, period)

	return NewSeriesListLine([][]float64{smaValues}, LineSeriesOption{
		Names: []string{fmt.Sprintf("SMA(%d)", period)},
		Label: SeriesLabel{
			Show: Ptr(false), // Don't show labels by default for indicators
		},
	})
}

// AddEMAToKlines adds Exponential Moving Average as a line series to chart data
func AddEMAToKlines(klinesData CandlestickSeries, period int, color Color) LineSeriesList {
	closes := ExtractClosePrices(klinesData)
	emaValues := CalculateEMA(closes, period)

	return NewSeriesListLine([][]float64{emaValues}, LineSeriesOption{
		Names: []string{fmt.Sprintf("EMA(%d)", period)},
		Label: SeriesLabel{
			Show: Ptr(false), // Don't show labels by default for indicators
		},
	})
}

// AddBollingerBandsToKlines adds Bollinger Bands as line series
func AddBollingerBandsToKlines(klinesData CandlestickSeries, period int, multiplier float64) LineSeriesList {
	closes := ExtractClosePrices(klinesData)
	bands := CalculateBollingerBands(closes, period, multiplier)

	return NewSeriesListLine([][]float64{bands.Upper, bands.Middle, bands.Lower}, LineSeriesOption{
		Names: []string{
			fmt.Sprintf("BB Upper(%.1f)", multiplier),
			fmt.Sprintf("BB Middle(%d)", period),
			fmt.Sprintf("BB Lower(%.1f)", multiplier),
		},
		Label: SeriesLabel{
			Show: Ptr(false), // Don't show labels by default for indicators
		},
	})
}

// AddRSIToKlines adds RSI as a line series (typically rendered in a separate chart)
func AddRSIToKlines(klinesData CandlestickSeries, period int) LineSeriesList {
	closes := ExtractClosePrices(klinesData)
	rsiValues := CalculateRSI(closes, period)

	return NewSeriesListLine([][]float64{rsiValues}, LineSeriesOption{
		Names: []string{fmt.Sprintf("RSI(%d)", period)},
		Label: SeriesLabel{
			Show: Ptr(false),
		},
	})
}
