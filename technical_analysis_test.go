package charts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateSMA(t *testing.T) {
	t.Parallel()

	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	result := CalculateSMA(values, 3)

	require.Len(t, result, 10)

	// First two values should be null
	assert.InDelta(t, GetNullValue(), result[0], 0.001)
	assert.InDelta(t, GetNullValue(), result[1], 0.001)

	// Third value should be (1+2+3)/3 = 2
	assert.InDelta(t, 2.0, result[2], 0.001)

	// Fourth value should be (2+3+4)/3 = 3
	assert.InDelta(t, 3.0, result[3], 0.001)

	// Last value should be (8+9+10)/3 = 9
	assert.InDelta(t, 9.0, result[9], 0.001)
}

func TestCalculateEMA(t *testing.T) {
	t.Parallel()

	values := []float64{1, 2, 3, 4, 5}
	result := CalculateEMA(values, 3)

	require.Len(t, result, 5)

	// First value should equal input
	assert.InDelta(t, 1.0, result[0], 0.001)

	// EMA should be calculated with smoothing factor 2/(3+1) = 0.5
	multiplier := 2.0 / 4.0
	expected := (2.0 * multiplier) + (1.0 * (1 - multiplier))
	assert.InDelta(t, expected, result[1], 0.001)
}

func TestCalculateBollingerBands(t *testing.T) {
	t.Parallel()

	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	bands := CalculateBollingerBands(values, 3, 2.0)

	require.Len(t, bands.Upper, 10)
	require.Len(t, bands.Middle, 10)
	require.Len(t, bands.Lower, 10)

	// First two values should be null
	assert.InDelta(t, GetNullValue(), bands.Upper[0], 0.001)
	assert.InDelta(t, GetNullValue(), bands.Middle[0], 0.001)
	assert.InDelta(t, GetNullValue(), bands.Lower[0], 0.001)

	// Middle band should equal SMA
	sma := CalculateSMA(values, 3)
	for i, middle := range bands.Middle {
		if i < 2 {
			continue // skip null values
		}
		assert.InDelta(t, sma[i], middle, 0.001)
	}

	// Upper should be greater than middle, lower should be less
	for i := 2; i < len(bands.Upper); i++ {
		assert.Greater(t, bands.Upper[i], bands.Middle[i])
		assert.Less(t, bands.Lower[i], bands.Middle[i])
	}
}

func TestCalculateRSI(t *testing.T) {
	t.Parallel()

	// Create test data with known gains/losses
	values := []float64{44, 44.5, 43.8, 44.2, 44.5, 43.9, 44.5, 44.9, 44.5, 44.8}
	result := CalculateRSI(values, 3)

	require.Len(t, result, 10)

	// First three values should be null
	for i := 0; i < 3; i++ {
		assert.InDelta(t, GetNullValue(), result[i], 0.001)
	}

	// RSI values should be between 0 and 100
	for i := 3; i < len(result); i++ {
		assert.GreaterOrEqual(t, result[i], 0.0)
		assert.LessOrEqual(t, result[i], 100.0)
	}
}

func TestPatternDetection(t *testing.T) {
	t.Parallel()

	t.Run("DetectDoji", func(t *testing.T) {
		// Doji: open ≈ close
		doji := OHLCData{Open: 100, High: 105, Low: 95, Close: 100.05}
		assert.True(t, DetectDoji(doji, 0.01)) // 1% threshold

		// Not a doji
		notDoji := OHLCData{Open: 100, High: 105, Low: 95, Close: 103}
		assert.False(t, DetectDoji(notDoji, 0.01))
	})

	t.Run("DetectHammer", func(t *testing.T) {
		// Hammer: long lower shadow, small body at top, small upper shadow
		// Body size = |104.1-104.0| = 0.1, Lower shadow = 104.0-95 = 9.0, Upper shadow = 104.1-104.1 = 0.0
		// Lower shadow (9.0) >= 2.0 * body (0.1) = 0.2 ✓, Upper shadow (0.0) <= body*0.5 = 0.05 ✓
		hammer := OHLCData{Open: 104.1, High: 104.1, Low: 95, Close: 104.0}
		assert.True(t, DetectHammer(hammer, 2.0))

		// Not a hammer (large body)
		notHammer := OHLCData{Open: 100, High: 105, Low: 95, Close: 105}
		assert.False(t, DetectHammer(notHammer, 2.0))
	})

	t.Run("DetectInvertedHammer", func(t *testing.T) {
		// Inverted hammer: long upper shadow, small body at bottom, small lower shadow
		// Body size = |96.1-96.0| = 0.1, Upper shadow = 105-96.1 = 8.9, Lower shadow = 96.0-96.0 = 0.0
		// Upper shadow (8.9) >= 2.0 * body (0.1) = 0.2 ✓, Lower shadow (0.0) <= body*0.5 = 0.05 ✓
		inverted := OHLCData{Open: 96.0, High: 105, Low: 96.0, Close: 96.1}
		assert.True(t, DetectInvertedHammer(inverted, 2.0))

		// Not an inverted hammer
		notInverted := OHLCData{Open: 100, High: 102, Low: 95, Close: 105}
		assert.False(t, DetectInvertedHammer(notInverted, 2.0))
	})

	t.Run("DetectEngulfing", func(t *testing.T) {
		// Bullish engulfing: small bearish followed by large bullish
		prev := OHLCData{Open: 102, High: 103, Low: 101, Close: 101.5}
		current := OHLCData{Open: 100, High: 105, Low: 99, Close: 104}

		bullish, bearish := DetectEngulfing(prev, current, 0.8)
		assert.True(t, bullish)
		assert.False(t, bearish)

		// Bearish engulfing: small bullish followed by large bearish
		prev2 := OHLCData{Open: 101, High: 102, Low: 100, Close: 101.5}
		current2 := OHLCData{Open: 103, High: 105, Low: 99, Close: 100}

		bullish2, bearish2 := DetectEngulfing(prev2, current2, 0.8)
		assert.False(t, bullish2)
		assert.True(t, bearish2)
	})
}

func TestScanCandlestickPatterns(t *testing.T) {
	t.Parallel()

	data := []OHLCData{
		// Normal candle
		{Open: 100, High: 110, Low: 95, Close: 105},
		// Doji
		{Open: 105, High: 108, Low: 102, Close: 105.01},
		// Hammer
		{Open: 108, High: 109, Low: 98, Close: 107},
		// Setup for engulfing
		{Open: 107, High: 108, Low: 103, Close: 104},
		// Bullish engulfing
		{Open: 102, High: 115, Low: 101, Close: 113},
	}

	series := CandlestickSeries{Data: data}
	patterns := ScanCandlestickPatterns(series)

	// Should detect at least the doji, hammer, and engulfing patterns
	assert.NotEmpty(t, patterns)

	// Check that pattern marks have required fields
	for _, pattern := range patterns {
		assert.Equal(t, SeriesMarkTypePattern, pattern.Type)
		assert.NotNil(t, pattern.Index)
		assert.NotEmpty(t, pattern.PatternType)
		assert.NotNil(t, pattern.Value)
	}
}

func TestAggregateCandlestick(t *testing.T) {
	t.Parallel()

	data := []OHLCData{
		{Open: 100, High: 110, Low: 95, Close: 105},  // Period 1
		{Open: 105, High: 115, Low: 100, Close: 112}, // Period 1
		{Open: 112, High: 118, Low: 108, Close: 115}, // Period 2
		{Open: 115, High: 120, Low: 110, Close: 118}, // Period 2
		{Open: 118, High: 125, Low: 115, Close: 122}, // Incomplete period
	}

	series := CandlestickSeries{Data: data, Name: "Test"}
	aggregated := AggregateCandlestick(series, 2)

	// Should have 3 periods (2 complete + 1 incomplete)
	assert.Len(t, aggregated.Data, 3)
	assert.Equal(t, "Test (Aggregated)", aggregated.Name)

	// Test first aggregated period
	first := aggregated.Data[0]
	assert.InDelta(t, 100.0, first.Open, 0.001)  // First open
	assert.InDelta(t, 112.0, first.Close, 0.001) // Last close
	assert.InDelta(t, 115.0, first.High, 0.001)  // Max high
	assert.InDelta(t, 95.0, first.Low, 0.001)    // Min low

	// Test second aggregated period
	second := aggregated.Data[1]
	assert.InDelta(t, 112.0, second.Open, 0.001)  // First open
	assert.InDelta(t, 118.0, second.Close, 0.001) // Last close
	assert.InDelta(t, 120.0, second.High, 0.001)  // Max high
	assert.InDelta(t, 108.0, second.Low, 0.001)   // Min low

	// Test incomplete period (single candle)
	third := aggregated.Data[2]
	assert.Equal(t, data[4], third) // Should be unchanged
}

func TestTechnicalAnalysisConvenienceFunctions(t *testing.T) {
	t.Parallel()

	ohlcData := []OHLCData{
		{Open: 100, High: 110, Low: 95, Close: 105},
		{Open: 105, High: 115, Low: 100, Close: 112},
		{Open: 112, High: 118, Low: 108, Close: 115},
		{Open: 115, High: 120, Low: 110, Close: 118},
		{Open: 118, High: 125, Low: 115, Close: 122},
	}

	series := CandlestickSeries{Data: ohlcData}

	t.Run("AddSMAToKlines", func(t *testing.T) {
		smaLines := AddSMAToKlines(series, 3, Color{})
		assert.Len(t, smaLines, 1)
		assert.Equal(t, "SMA(3)", smaLines[0].Name)
		assert.Len(t, smaLines[0].Values, 5)
	})

	t.Run("AddEMAToKlines", func(t *testing.T) {
		emaLines := AddEMAToKlines(series, 3, Color{})
		assert.Len(t, emaLines, 1)
		assert.Equal(t, "EMA(3)", emaLines[0].Name)
		assert.Len(t, emaLines[0].Values, 5)
	})

	t.Run("AddBollingerBandsToKlines", func(t *testing.T) {
		bbLines := AddBollingerBandsToKlines(series, 3, 2.0)
		assert.Len(t, bbLines, 3) // Upper, middle, lower
		assert.Equal(t, "BB Upper(2.0)", bbLines[0].Name)
		assert.Equal(t, "BB Middle(3)", bbLines[1].Name)
		assert.Equal(t, "BB Lower(2.0)", bbLines[2].Name)
	})

	t.Run("AddRSIToKlines", func(t *testing.T) {
		rsiLines := AddRSIToKlines(series, 3)
		assert.Len(t, rsiLines, 1)
		assert.Equal(t, "RSI(3)", rsiLines[0].Name)
		assert.Len(t, rsiLines[0].Values, 5)
	})
}

func TestNewCandlestickWithPatterns(t *testing.T) {
	t.Parallel()

	data := []OHLCData{
		{Open: 100, High: 110, Low: 95, Close: 105},
		{Open: 105, High: 108, Low: 102, Close: 105.01},   // Doji - small body
		{Open: 104.1, High: 104.1, Low: 95, Close: 104.0}, // Hammer - long lower shadow, small body
	}

	// Test with custom options more likely to detect patterns
	options := PatternDetectionOption{
		DojiThreshold:    0.01, // 1% threshold - more lenient
		ShadowRatio:      2.0,  // Standard ratio
		EngulfingMinSize: 0.5,  // Smaller minimum size
	}
	series := NewCandlestickWithPatterns(data, options)
	assert.NotNil(t, series.MarkPoint.Points) // Should at least initialize the slice

	// Should detect at least the doji and hammer patterns
	assert.NotEmpty(t, series.MarkPoint.Points, "Should have detected patterns with lenient thresholds")
}

func TestValidateOHLCData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ohlc     OHLCData
		expected bool
	}{
		{
			name:     "valid_bullish",
			ohlc:     OHLCData{Open: 100, High: 110, Low: 95, Close: 105},
			expected: true,
		},
		{
			name:     "valid_bearish",
			ohlc:     OHLCData{Open: 110, High: 115, Low: 100, Close: 105},
			expected: true,
		},
		{
			name:     "valid_doji",
			ohlc:     OHLCData{Open: 100, High: 105, Low: 95, Close: 100},
			expected: true,
		},
		{
			name:     "invalid_high_too_low",
			ohlc:     OHLCData{Open: 100, High: 98, Low: 95, Close: 105},
			expected: false,
		},
		{
			name:     "invalid_low_too_high",
			ohlc:     OHLCData{Open: 100, High: 110, Low: 102, Close: 105},
			expected: false,
		},
		{
			name:     "invalid_null_values",
			ohlc:     OHLCData{Open: GetNullValue(), High: 110, Low: 95, Close: 105},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateOHLCData(tt.ohlc)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests for performance
func BenchmarkCalculateSMA(b *testing.B) {
	values := make([]float64, 1000)
	for i := range values {
		values[i] = float64(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateSMA(values, 20)
	}
}

func BenchmarkCalculateEMA(b *testing.B) {
	values := make([]float64, 1000)
	for i := range values {
		values[i] = float64(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateEMA(values, 20)
	}
}

func BenchmarkScanCandlestickPatterns(b *testing.B) {
	data := make([]OHLCData, 1000)
	for i := range data {
		base := float64(100 + i)
		data[i] = OHLCData{
			Open:  base,
			High:  base + 5,
			Low:   base - 3,
			Close: base + 2,
		}
	}

	series := CandlestickSeries{Data: data}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ScanCandlestickPatterns(series)
	}
}
