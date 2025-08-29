package charts

import (
	"regexp"
	"strconv"
	"testing"

	"github.com/go-analyze/bulk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data for advanced pattern detection
func makeAdvancedPatternTestData() []OHLCData {
	return []OHLCData{
		// Normal candle for reference
		{Open: 100, High: 110, Low: 95, Close: 105},

		// Harami setup: Large bearish candle
		{Open: 110, High: 115, Low: 95, Close: 98},

		// Bullish Harami: Small bullish candle within previous bearish body
		{Open: 102, High: 106, Low: 100, Close: 104},

		// Shooting Star: Small body at bottom, long upper shadow
		{Open: 106, High: 125, Low: 105, Close: 107},

		// Gravestone Doji: Open ≈ Close with long upper shadow
		{Open: 108, High: 120, Low: 107, Close: 108.1},

		// Dragonfly Doji: Open ≈ Close with long lower shadow
		{Open: 109, High: 110, Low: 90, Close: 108.9},

		// Morning Star setup: Large bearish candle
		{Open: 120, High: 125, Low: 105, Close: 108},

		// Morning Star middle: Small body with gap down
		{Open: 102, High: 104, Low: 100, Close: 103},

		// Morning Star completion: Large bullish candle with gap up
		{Open: 108, High: 125, Low: 106, Close: 122},

		// Evening Star setup: Large bullish candle
		{Open: 122, High: 140, Low: 120, Close: 138},

		// Evening Star middle: Small body with gap up
		{Open: 142, High: 144, Low: 140, Close: 143},

		// Evening Star completion: Large bearish candle with gap down
		{Open: 138, High: 140, Low: 115, Close: 118},

		// Long Legged Doji: Open ≈ Close with exceptionally long shadows on both sides
		{Open: 100, High: 115, Low: 85, Close: 100.2},

		// High Wave: Small body relative to range with very long shadows showing extreme volatility
		{Open: 100, High: 112, Low: 88, Close: 102},

		// Bullish Belt Hold: Opens at/near low, closes at/near high with minimal lower shadow
		{Open: 100, High: 110, Low: 99.8, Close: 109},

		// Bearish Belt Hold: Opens at/near high, closes at/near low with minimal upper shadow
		{Open: 110, High: 110.2, Low: 100, Close: 101},
	}
}

func TestDojiPattern(t *testing.T) {
	t.Parallel()

	// Valid doji: open ≈ close
	doji := OHLCData{Open: 100, High: 105, Low: 95, Close: 100.1}
	data := []OHLCData{doji}
	for _, tt := range []struct {
		name      string
		threshold float64
		expected  bool
	}{
		{"low", 0.009, false},
		{"default", 0.01, true},
		{"high", 0.011, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, detectDojiAt(data, 0, CandlestickPatternConfig{DojiThreshold: tt.threshold}))
		})
	}

	// Invalid: body too large
	notDoji := OHLCData{Open: 100, High: 105, Low: 95, Close: 103}
	data = []OHLCData{notDoji}
	assert.False(t, detectDojiAt(data, 0, CandlestickPatternConfig{DojiThreshold: 0.01}))

	// Invalid: invalid OHLC
	invalidOHLC := OHLCData{Open: 100, High: 95, Low: 105, Close: 98}
	data = []OHLCData{invalidOHLC}
	assert.False(t, detectDojiAt(data, 0, CandlestickPatternConfig{DojiThreshold: 0.01}))
}

func TestHammerPattern(t *testing.T) {
	t.Parallel()

	// Valid hammer: long lower shadow, small body at top
	hammer := OHLCData{Open: 105, High: 107, Low: 95, Close: 106}
	data := []OHLCData{hammer}
	for _, tt := range []struct {
		name     string
		ratio    float64
		expected bool
	}{
		{"low", 1.0, true},
		{"default", 2.0, true},
		{"high", 11.1, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, detectHammerAt(data, 0, CandlestickPatternConfig{ShadowRatio: tt.ratio}))
		})
	}

	// Invalid: short lower shadow
	notHammer := OHLCData{Open: 105, High: 107, Low: 104, Close: 106}
	data = []OHLCData{notHammer}
	assert.False(t, detectHammerAt(data, 0, CandlestickPatternConfig{ShadowRatio: 2.0}))

	// Invalid: long upper shadow
	notHammer2 := OHLCData{Open: 95, High: 107, Low: 94, Close: 96}
	data = []OHLCData{notHammer2}
	assert.False(t, detectHammerAt(data, 0, CandlestickPatternConfig{ShadowRatio: 2.0}))
}

func TestInvertedHammerPattern(t *testing.T) {
	t.Parallel()

	// Valid inverted hammer: long upper shadow, small body at bottom
	invertedHammer := OHLCData{Open: 95, High: 107, Low: 94, Close: 96}
	for _, tt := range []struct {
		name     string
		ratio    float64
		expected bool
	}{
		{"low", 1.0, true},
		{"default", 2.0, true},
		{"high", 11.1, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, detectInvertedHammerAt([]OHLCData{invertedHammer}, 0, CandlestickPatternConfig{ShadowRatio: tt.ratio}))
		})
	}

	// Invalid: short upper shadow
	notInvertedHammer := OHLCData{Open: 95, High: 97, Low: 94, Close: 96}
	assert.False(t, detectInvertedHammerAt([]OHLCData{notInvertedHammer}, 0, CandlestickPatternConfig{ShadowRatio: 2.0}))
}

func TestEngulfingPattern(t *testing.T) {
	t.Parallel()

	// Test Bullish Engulfing
	prevBearish := OHLCData{Open: 110, High: 112, Low: 105, Close: 106}
	currentBullish := OHLCData{Open: 104, High: 115, Low: 103, Close: 114}
	for _, tt := range []struct {
		name     string
		size     float64
		expected bool
	}{
		{"low", 0.5, true},
		{"default", 0.8, true},
		{"high", 2.6, false},
	} {
		t.Run("bullish_"+tt.name, func(t *testing.T) {
			detected := detectBullishEngulfingAt([]OHLCData{prevBearish, currentBullish}, 1, CandlestickPatternConfig{EngulfingMinSize: tt.size})
			assert.Equal(t, tt.expected, detected)
		})
	}
	assert.False(t, detectBearishEngulfingAt([]OHLCData{prevBearish, currentBullish}, 1, CandlestickPatternConfig{EngulfingMinSize: 0.8}))

	// Test Bearish Engulfing
	prevBullish := OHLCData{Open: 106, High: 112, Low: 105, Close: 110}
	currentBearish := OHLCData{Open: 114, High: 115, Low: 103, Close: 104}

	for _, tt := range []struct {
		name     string
		size     float64
		expected bool
	}{
		{"low", 0.5, true},
		{"default", 0.8, true},
		{"high", 2.6, false},
	} {
		t.Run("bearish_"+tt.name, func(t *testing.T) {
			detected := detectBearishEngulfingAt([]OHLCData{prevBullish, currentBearish}, 1, CandlestickPatternConfig{EngulfingMinSize: tt.size})
			assert.Equal(t, tt.expected, detected)
		})
	}
	assert.False(t, detectBullishEngulfingAt([]OHLCData{prevBullish, currentBearish}, 1, CandlestickPatternConfig{EngulfingMinSize: 0.8}))

	// Test non-engulfing
	nonEngulfing := OHLCData{Open: 107, High: 109, Low: 106, Close: 108}
	assert.False(t, detectBullishEngulfingAt([]OHLCData{prevBullish, nonEngulfing}, 1, CandlestickPatternConfig{EngulfingMinSize: 0.8}))
	assert.False(t, detectBearishEngulfingAt([]OHLCData{prevBullish, nonEngulfing}, 1, CandlestickPatternConfig{EngulfingMinSize: 0.8}))
}

func TestHaramiPatterns(t *testing.T) {
	t.Parallel()

	// Test Bullish Harami
	prevCandle := OHLCData{Open: 110, High: 115, Low: 95, Close: 98}      // Large bearish
	currentCandle := OHLCData{Open: 102, High: 106, Low: 100, Close: 104} // Small bullish inside
	for _, tt := range []struct {
		name     string
		size     float64
		expected bool
	}{
		{"low", 1.0 - 0.5, true},
		{"default", 1.0 - 0.3, true},
		{"high", 1.0 - 0.15, false},
	} {
		t.Run("bullish_"+tt.name, func(t *testing.T) {
			detected := detectBullishHaramiAt([]OHLCData{prevCandle, currentCandle}, 1, CandlestickPatternConfig{EngulfingMinSize: tt.size})
			assert.Equal(t, tt.expected, detected)
		})
	}
	assert.False(t, detectBearishHaramiAt([]OHLCData{prevCandle, currentCandle}, 1, CandlestickPatternConfig{EngulfingMinSize: 1.0 - 0.3}))

	// Test Bearish Harami
	prevCandle = OHLCData{Open: 98, High: 115, Low: 95, Close: 110}      // Large bullish
	currentCandle = OHLCData{Open: 106, High: 108, Low: 102, Close: 104} // Small bearish inside

	for _, tt := range []struct {
		name     string
		size     float64
		expected bool
	}{
		{"low", 1.0 - 0.5, true},
		{"default", 1.0 - 0.3, true},
		{"high", 1.0 - 0.15, false},
	} {
		t.Run("bearish_"+tt.name, func(t *testing.T) {
			detected := detectBearishHaramiAt([]OHLCData{prevCandle, currentCandle}, 1, CandlestickPatternConfig{EngulfingMinSize: tt.size})
			assert.Equal(t, tt.expected, detected)
		})
	}
	assert.False(t, detectBullishHaramiAt([]OHLCData{prevCandle, currentCandle}, 1, CandlestickPatternConfig{EngulfingMinSize: 1.0 - 0.3}))

	// Test non-harami (current candle too large)
	currentCandle = OHLCData{Open: 100, High: 112, Low: 96, Close: 108} // Too large
	assert.False(t, detectBullishHaramiAt([]OHLCData{prevCandle, currentCandle}, 1, CandlestickPatternConfig{EngulfingMinSize: 1.0 - 0.3}))
	assert.False(t, detectBearishHaramiAt([]OHLCData{prevCandle, currentCandle}, 1, CandlestickPatternConfig{EngulfingMinSize: 1.0 - 0.3}))
}

func TestShootingStarPattern(t *testing.T) {
	t.Parallel()

	// Valid shooting star: small body at bottom, long upper shadow
	shootingStar := OHLCData{Open: 106, High: 125, Low: 105, Close: 107}
	for _, tt := range []struct {
		name     string
		ratio    float64
		expected bool
	}{
		{"low", 1.0, true},
		{"default", 2.0, true},
		{"high", 18.1, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, detectShootingStarAt([]OHLCData{shootingStar}, 0, CandlestickPatternConfig{ShadowRatio: tt.ratio}))
		})
	}

	// Invalid: body not near bottom
	notShootingStar := OHLCData{Open: 115, High: 125, Low: 105, Close: 117}
	assert.False(t, detectShootingStarAt([]OHLCData{notShootingStar}, 0, CandlestickPatternConfig{ShadowRatio: 2.0}))

	// Invalid: upper shadow too short
	shortShadow := OHLCData{Open: 106, High: 110, Low: 105, Close: 107}
	assert.False(t, detectShootingStarAt([]OHLCData{shortShadow}, 0, CandlestickPatternConfig{ShadowRatio: 2.0}))
}

func TestGravestoneDojiPattern(t *testing.T) {
	t.Parallel()
	gravestoneDoji := OHLCData{Open: 108, High: 120, Low: 107, Close: 108.1}
	for _, tt := range []struct {
		name      string
		threshold float64
		shadow    float64
		expected  bool
	}{
		{"low_threshold", 0.004, 2.0, false},
		{"default", 0.01, 2.0, true},
		{"high_shadow", 0.01, 200, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			opt := CandlestickPatternConfig{DojiThreshold: tt.threshold, ShadowRatio: tt.shadow}
			assert.Equal(t, tt.expected, detectGravestoneDojiAt([]OHLCData{gravestoneDoji}, 0, opt))
		})
	}

	// Invalid: not a doji (body too large)
	notDoji := OHLCData{Open: 108, High: 120, Low: 107, Close: 115}
	assert.False(t, detectGravestoneDojiAt([]OHLCData{notDoji}, 0, CandlestickPatternConfig{DojiThreshold: 0.01, ShadowRatio: 2.0}))

	// Invalid: doji but no long upper shadow
	dojiNoShadow := OHLCData{Open: 108, High: 109, Low: 107, Close: 108.1}
	assert.False(t, detectGravestoneDojiAt([]OHLCData{dojiNoShadow}, 0, CandlestickPatternConfig{DojiThreshold: 0.01, ShadowRatio: 2.0}))
}

func TestDragonflyDojiPattern(t *testing.T) {
	t.Parallel()
	dragonflyDoji := OHLCData{Open: 109, High: 110, Low: 90, Close: 108.9}
	for _, tt := range []struct {
		name      string
		threshold float64
		shadow    float64
		expected  bool
	}{
		{"low_threshold", 0.004, 2.0, false},
		{"default", 0.01, 2.0, true},
		{"high_shadow", 0.01, 200, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			opt := CandlestickPatternConfig{DojiThreshold: tt.threshold, ShadowRatio: tt.shadow}
			assert.Equal(t, tt.expected, detectDragonflyDojiAt([]OHLCData{dragonflyDoji}, 0, opt))
		})
	}

	// Invalid: not a doji
	notDoji := OHLCData{Open: 109, High: 110, Low: 90, Close: 102}
	assert.False(t, detectDragonflyDojiAt([]OHLCData{notDoji}, 0, CandlestickPatternConfig{DojiThreshold: 0.01, ShadowRatio: 2.0}))

	// Invalid: doji but no long lower shadow
	dojiNoShadow := OHLCData{Open: 109, High: 110, Low: 108, Close: 108.9}
	assert.False(t, detectDragonflyDojiAt([]OHLCData{dojiNoShadow}, 0, CandlestickPatternConfig{DojiThreshold: 0.01, ShadowRatio: 2.0}))
}

func TestMorningStarPattern(t *testing.T) {
	t.Parallel()

	opt := CandlestickPatternConfig{}

	// Valid morning star pattern
	first := OHLCData{Open: 120, High: 125, Low: 105, Close: 108}  // Large bearish
	second := OHLCData{Open: 102, High: 104, Low: 100, Close: 103} // Small body, gap down
	third := OHLCData{Open: 108, High: 125, Low: 106, Close: 122}  // Large bullish, gap up

	assert.True(t, detectMorningStarAt([]OHLCData{first, second, third}, 2, opt))

	// Invalid: first candle not bearish
	invalidFirst := OHLCData{Open: 108, High: 125, Low: 105, Close: 120} // Bullish
	assert.False(t, detectMorningStarAt([]OHLCData{invalidFirst, second, third}, 2, opt))

	// Invalid: no gap down between first and second
	noGapSecond := OHLCData{Open: 109, High: 111, Low: 107, Close: 110} // No gap
	assert.False(t, detectMorningStarAt([]OHLCData{first, noGapSecond, third}, 2, opt))

	// Invalid: third candle not bullish
	invalidThird := OHLCData{Open: 108, High: 110, Low: 105, Close: 107} // Bearish
	assert.False(t, detectMorningStarAt([]OHLCData{first, second, invalidThird}, 2, opt))
}

func TestEveningStarPattern(t *testing.T) {
	t.Parallel()

	opt := CandlestickPatternConfig{}

	// Valid evening star pattern
	first := OHLCData{Open: 122, High: 140, Low: 120, Close: 138}  // Large bullish
	second := OHLCData{Open: 142, High: 144, Low: 140, Close: 143} // Small body, gap up
	third := OHLCData{Open: 138, High: 140, Low: 115, Close: 118}  // Large bearish, gap down

	assert.True(t, detectEveningStarAt([]OHLCData{first, second, third}, 2, opt))

	// Invalid: first candle not bullish
	invalidFirst := OHLCData{Open: 138, High: 140, Low: 120, Close: 122} // Bearish
	assert.False(t, detectEveningStarAt([]OHLCData{invalidFirst, second, third}, 2, opt))

	// Invalid: no gap up between first and second
	noGapSecond := OHLCData{Open: 136, High: 140, Low: 134, Close: 139} // No gap
	assert.False(t, detectEveningStarAt([]OHLCData{first, noGapSecond, third}, 2, opt))

	// Invalid: third candle not bearish
	invalidThird := OHLCData{Open: 138, High: 145, Low: 135, Close: 142} // Bullish
	assert.False(t, detectEveningStarAt([]OHLCData{first, second, invalidThird}, 2, opt))
}

func newCandlestickWithPatterns(data []OHLCData, options ...CandlestickPatternConfig) CandlestickSeries {
	// Start with defaults and override with provided options
	config := &CandlestickPatternConfig{
		ReplaceSeriesLabel: true,
		EnabledPatterns:    PatternsAll().EnabledPatterns,
		DojiThreshold:      0.001,
		ShadowTolerance:    0.01,
		BodySizeRatio:      0.3,
		ShadowRatio:        2.0,
		EngulfingMinSize:   0.8,
	}
	if len(options) > 0 {
		// Merge provided options with defaults
		opt := options[0]
		if opt.DojiThreshold > 0 {
			config.DojiThreshold = opt.DojiThreshold
		}
		if opt.ShadowRatio > 0 {
			config.ShadowRatio = opt.ShadowRatio
		}
		if opt.EngulfingMinSize > 0 {
			config.EngulfingMinSize = opt.EngulfingMinSize
		}
		if opt.ShadowTolerance > 0 {
			config.ShadowTolerance = opt.ShadowTolerance
		}
		if opt.BodySizeRatio > 0 {
			config.BodySizeRatio = opt.BodySizeRatio
		}
	}

	return CandlestickSeries{
		Data:          data,
		PatternConfig: config,
	}
}

func TestPatternIntegration(t *testing.T) {
	t.Parallel()

	// Test that all advanced patterns are detected in a comprehensive dataset
	data := makeAdvancedPatternTestData()
	opt := PatternsAll()
	opt.DojiThreshold = 0.01
	opt.ShadowRatio = 2.0
	opt.EngulfingMinSize = 0.8

	// Use private scan function for testing pattern detection
	indexPatterns := scanForCandlestickPatterns(data, *opt)

	assert.Len(t, indexPatterns, 12)

	patternTypes := make(map[string]int)
	for _, patterns := range indexPatterns {
		assert.NotEmpty(t, patterns)
		for _, pattern := range patterns {
			patternTypes[pattern.PatternType]++
		}
	}
	assert.Len(t, patternTypes, 14)

	// Test the convenience function
	seriesWithPatterns := newCandlestickWithPatterns(data, CandlestickPatternConfig{
		DojiThreshold:    0.01,
		ShadowRatio:      2.0,
		EngulfingMinSize: 0.8,
	})

	// Verify that pattern configuration is properly set
	assert.NotNil(t, seriesWithPatterns.PatternConfig)
	assert.True(t, seriesWithPatterns.PatternConfig.ReplaceSeriesLabel)
	assert.NotEmpty(t, seriesWithPatterns.PatternConfig.EnabledPatterns)
}

func TestMarubozuPattern(t *testing.T) {
	t.Parallel()

	// Bullish Marubozu - no shadows
	bullishMarubozu := OHLCData{Open: 100, High: 120, Low: 100, Close: 120}
	for _, tt := range []struct {
		tol      float64
		expected bool
	}{
		{0.005, true},
		{0.01, true},
		{0.02, true},
	} {
		t.Run("bullish_tol_"+strconv.FormatFloat(tt.tol, 'f', 3, 64), func(t *testing.T) {
			detected := detectBullishMarubozuAt([]OHLCData{bullishMarubozu}, 0, CandlestickPatternConfig{ShadowTolerance: tt.tol})
			assert.Equal(t, tt.expected, detected)
		})
	}
	assert.False(t, detectBearishMarubozuAt([]OHLCData{bullishMarubozu}, 0, CandlestickPatternConfig{ShadowTolerance: 0.01}))

	// Bearish Marubozu - no shadows
	bearishMarubozu := OHLCData{Open: 120, High: 120, Low: 100, Close: 100}
	for _, tt := range []struct {
		tol      float64
		expected bool
	}{
		{0.005, true},
		{0.01, true},
		{0.02, true},
	} {
		t.Run("bearish_tol_"+strconv.FormatFloat(tt.tol, 'f', 3, 64), func(t *testing.T) {
			detected := detectBearishMarubozuAt([]OHLCData{bearishMarubozu}, 0, CandlestickPatternConfig{ShadowTolerance: tt.tol})
			assert.Equal(t, tt.expected, detected)
		})
	}
	assert.False(t, detectBullishMarubozuAt([]OHLCData{bearishMarubozu}, 0, CandlestickPatternConfig{ShadowTolerance: 0.01}))

	// Not a marubozu - has significant shadows
	notMarubozu := OHLCData{Open: 105, High: 125, Low: 95, Close: 115}
	assert.False(t, detectBullishMarubozuAt([]OHLCData{notMarubozu}, 0, CandlestickPatternConfig{ShadowTolerance: 0.01}))
	assert.False(t, detectBearishMarubozuAt([]OHLCData{notMarubozu}, 0, CandlestickPatternConfig{ShadowTolerance: 0.01}))
	assert.True(t, detectBullishMarubozuAt([]OHLCData{notMarubozu}, 0, CandlestickPatternConfig{ShadowTolerance: 0.7}))
}

func TestSpinningTopPattern(t *testing.T) {
	t.Parallel()

	// Classic spinning top - small body, long shadows
	spinningTop := OHLCData{Open: 110, High: 125, Low: 95, Close: 112}
	for _, tt := range []struct {
		name     string
		ratio    float64
		expected bool
	}{
		{"low", 0.05, false},
		{"default", 0.3, true},
		{"high", 0.4, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, detectSpinningTopAt([]OHLCData{spinningTop}, 0, CandlestickPatternConfig{BodySizeRatio: tt.ratio}))
		})
	}

	// Not spinning top - large body
	largeBody := OHLCData{Open: 100, High: 125, Low: 95, Close: 120}
	assert.False(t, detectSpinningTopAt([]OHLCData{largeBody}, 0, CandlestickPatternConfig{BodySizeRatio: 0.3}))

	// Not spinning top - shadows too short relative to body
	shortShadows := OHLCData{Open: 110, High: 110.5, Low: 109.5, Close: 111}
	assert.False(t, detectSpinningTopAt([]OHLCData{shortShadows}, 0, CandlestickPatternConfig{BodySizeRatio: 0.3}))
}

func TestDetectLongLeggedDoji(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ohlc     OHLCData
		expected bool
	}{
		{
			name:     "valid_long_legged_doji",
			ohlc:     OHLCData{Open: 100, High: 115, Low: 85, Close: 100.2},
			expected: true,
		},
		{
			name:     "regular_doji_shadows_too_short",
			ohlc:     OHLCData{Open: 100, High: 102, Low: 98, Close: 100.1},
			expected: false,
		},
		{
			name:     "not_a_doji_body_too_large",
			ohlc:     OHLCData{Open: 100, High: 115, Low: 85, Close: 105},
			expected: false,
		},
		{
			name:     "asymmetric_shadows",
			ohlc:     OHLCData{Open: 100, High: 115, Low: 99, Close: 100.2},
			expected: false,
		},
		{
			name:     "long_upper_shadow_only",
			ohlc:     OHLCData{Open: 100, High: 115, Low: 98, Close: 100.1},
			expected: false,
		},
		{
			name:     "long_lower_shadow_only",
			ohlc:     OHLCData{Open: 100, High: 102, Low: 85, Close: 100.1},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectLongLeggedDojiAt([]OHLCData{tt.ohlc}, 0, CandlestickPatternConfig{DojiThreshold: 0.01, ShadowRatio: 3.0})
			assert.Equal(t, tt.expected, result)
		})
	}

	base := OHLCData{Open: 100, High: 115, Low: 85, Close: 100.2}
	for _, tt := range []struct {
		name      string
		threshold float64
		expected  bool
	}{
		{"low", 0.005, false},
		{"default", 0.01, true},
		{"high", 0.02, true},
	} {
		t.Run("threshold_"+tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, detectLongLeggedDojiAt([]OHLCData{base}, 0, CandlestickPatternConfig{DojiThreshold: tt.threshold, ShadowRatio: 3.0}))
		})
	}
	assert.False(t, detectLongLeggedDojiAt([]OHLCData{base}, 0, CandlestickPatternConfig{DojiThreshold: 0.01, ShadowRatio: 80}))
}

func TestDetectHighWave(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ohlc     OHLCData
		expected bool
	}{
		{
			name:     "valid_high_wave",
			ohlc:     OHLCData{Open: 100, High: 112, Low: 88, Close: 102},
			expected: true,
		},
		{
			name:     "body_too_large",
			ohlc:     OHLCData{Open: 100, High: 112, Low: 88, Close: 108},
			expected: false,
		},
		{
			name:     "shadows_too_short",
			ohlc:     OHLCData{Open: 100, High: 103, Low: 97, Close: 102},
			expected: false,
		},
		{
			name:     "asymmetric_no_lower_shadow",
			ohlc:     OHLCData{Open: 100, High: 115, Low: 100, Close: 102},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectHighWaveAt([]OHLCData{tt.ohlc}, 0, CandlestickPatternConfig{ShadowRatio: 3.0})
			assert.Equal(t, tt.expected, result)
		})
	}

	base := OHLCData{Open: 100, High: 112, Low: 88, Close: 102}
	for _, tt := range []struct {
		name     string
		ratio    float64
		expected bool
	}{
		{"low", 2.0, true},
		{"default", 3.0, true},
		{"high", 12.0, false},
	} {
		t.Run("ratio_"+tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, detectHighWaveAt([]OHLCData{base}, 0, CandlestickPatternConfig{ShadowRatio: tt.ratio}))
		})
	}
}

func TestDetectBeltHold(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ohlc       OHLCData
		expectBull bool
		expectBear bool
	}{
		{
			name:       "bullish_belt_hold",
			ohlc:       OHLCData{Open: 100, High: 110, Low: 99.8, Close: 109},
			expectBull: true,
			expectBear: false,
		},
		{
			name:       "bearish_belt_hold",
			ohlc:       OHLCData{Open: 110, High: 110.2, Low: 100, Close: 101},
			expectBull: false,
			expectBear: true,
		},
		{
			name:       "body_too_small",
			ohlc:       OHLCData{Open: 100, High: 110, Low: 90, Close: 105},
			expectBull: false,
			expectBear: false,
		},
		{
			name:       "too_much_shadow_on_open_side_bullish_attempt",
			ohlc:       OHLCData{Open: 105, High: 110, Low: 90, Close: 109},
			expectBull: false,
			expectBear: false,
		},
		{
			name:       "normal_candle",
			ohlc:       OHLCData{Open: 103, High: 107, Low: 100, Close: 105},
			expectBull: false,
			expectBear: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bull := detectBullishBeltHoldAt([]OHLCData{tt.ohlc}, 0, CandlestickPatternConfig{ShadowTolerance: 0.02})
			bear := detectBearishBeltHoldAt([]OHLCData{tt.ohlc}, 0, CandlestickPatternConfig{ShadowTolerance: 0.02})
			assert.Equal(t, tt.expectBull, bull)
			assert.Equal(t, tt.expectBear, bear)
		})
	}

	bullOHLC := OHLCData{Open: 100, High: 110, Low: 99.8, Close: 109}
	for _, tt := range []struct {
		name     string
		tol      float64
		expected bool
	}{
		{"low", 0.01, false},
		{"default", 0.02, true},
		{"high", 0.03, true},
	} {
		t.Run("bull_tol_"+tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, detectBullishBeltHoldAt([]OHLCData{bullOHLC}, 0, CandlestickPatternConfig{ShadowTolerance: tt.tol}))
		})
	}
	bearOHLC := OHLCData{Open: 110, High: 110.2, Low: 100, Close: 101}
	for _, tt := range []struct {
		name     string
		tol      float64
		expected bool
	}{
		{"low", 0.01, false},
		{"default", 0.02, true},
		{"high", 0.03, true},
	} {
		t.Run("bear_tol_"+tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, detectBearishBeltHoldAt([]OHLCData{bearOHLC}, 0, CandlestickPatternConfig{ShadowTolerance: tt.tol}))
		})
	}
}

func TestPiercingLinePattern(t *testing.T) {
	t.Parallel()

	// Classic piercing line - bearish then bullish with gap down and close above midpoint
	prev := OHLCData{Open: 120, High: 120, Low: 110, Close: 110}    // Bearish
	current := OHLCData{Open: 108, High: 118, Low: 108, Close: 116} // Bullish, opens below prev low, closes above midpoint (115)
	detected := detectPiercingLineAt([]OHLCData{prev, current}, 1, CandlestickPatternConfig{})
	assert.True(t, detected)

	// Not piercing line - current closes below midpoint
	current = OHLCData{Open: 108, High: 114, Low: 108, Close: 112}
	detected = detectPiercingLineAt([]OHLCData{prev, current}, 1, CandlestickPatternConfig{})
	assert.False(t, detected)
}

func TestDarkCloudCoverPattern(t *testing.T) {
	t.Parallel()

	// Classic dark cloud cover - bullish then bearish with gap up and close below midpoint
	prev := OHLCData{Open: 110, High: 120, Low: 110, Close: 120}    // Bullish
	current := OHLCData{Open: 122, High: 122, Low: 112, Close: 114} // Bearish, opens above prev high, closes below midpoint (115)
	detected := detectDarkCloudCoverAt([]OHLCData{prev, current}, 1, CandlestickPatternConfig{})
	assert.True(t, detected)

	// Not dark cloud cover - current closes above midpoint
	current = OHLCData{Open: 122, High: 122, Low: 118, Close: 118}
	detected = detectDarkCloudCoverAt([]OHLCData{prev, current}, 1, CandlestickPatternConfig{})
	assert.False(t, detected)
}

func TestPatternValidation(t *testing.T) {
	t.Parallel()

	// Test with invalid OHLC data
	invalidOHLC := OHLCData{Open: 100, High: 95, Low: 105, Close: 98} // High < Low

	assert.False(t, detectDojiAt([]OHLCData{invalidOHLC}, 0, CandlestickPatternConfig{DojiThreshold: 0.01}))
	assert.False(t, detectHammerAt([]OHLCData{invalidOHLC}, 0, CandlestickPatternConfig{ShadowRatio: 2.0}))
	assert.False(t, detectShootingStarAt([]OHLCData{invalidOHLC}, 0, CandlestickPatternConfig{ShadowRatio: 2.0}))

	// Test three-candle patterns with invalid data
	validOHLC := OHLCData{Open: 100, High: 110, Low: 95, Close: 105}
	opt := CandlestickPatternConfig{}

	assert.False(t, detectMorningStarAt([]OHLCData{invalidOHLC, validOHLC, validOHLC}, 2, opt))
	assert.False(t, detectMorningStarAt([]OHLCData{validOHLC, invalidOHLC, validOHLC}, 2, opt))
	assert.False(t, detectMorningStarAt([]OHLCData{validOHLC, validOHLC, invalidOHLC}, 2, opt))

	assert.False(t, detectEveningStarAt([]OHLCData{invalidOHLC, validOHLC, validOHLC}, 2, opt))
	assert.False(t, detectEveningStarAt([]OHLCData{validOHLC, invalidOHLC, validOHLC}, 2, opt))
	assert.False(t, detectEveningStarAt([]OHLCData{validOHLC, validOHLC, invalidOHLC}, 2, opt))

	// Test two-candle patterns with invalid data
	bullish := detectBullishHaramiAt([]OHLCData{invalidOHLC, validOHLC}, 1, CandlestickPatternConfig{EngulfingMinSize: 1.0 - 0.3})
	bearish := detectBearishHaramiAt([]OHLCData{invalidOHLC, validOHLC}, 1, CandlestickPatternConfig{EngulfingMinSize: 1.0 - 0.3})
	assert.False(t, bullish && bearish)

	bullish = detectBullishHaramiAt([]OHLCData{validOHLC, invalidOHLC}, 1, CandlestickPatternConfig{EngulfingMinSize: 1.0 - 0.3})
	bearish = detectBearishHaramiAt([]OHLCData{validOHLC, invalidOHLC}, 1, CandlestickPatternConfig{EngulfingMinSize: 1.0 - 0.3})
	assert.False(t, bullish && bearish)
}

func TestPatternScanningComprehensive(t *testing.T) {
	t.Parallel()

	// Create data with known patterns
	data := []OHLCData{
		// Index 0: Normal candle
		{Open: 100, High: 110, Low: 95, Close: 105},
		// Index 1: Doji
		{Open: 105, High: 108, Low: 102, Close: 105.05},
		// Index 2: Hammer
		{Open: 108, High: 109, Low: 98, Close: 107},
		// Index 3: Shooting Star
		{Open: 106, High: 125, Low: 105, Close: 107},
		// Index 4: Gravestone Doji
		{Open: 108, High: 120, Low: 107, Close: 108.1},
		// Index 5: Dragonfly Doji
		{Open: 109, High: 110, Low: 90, Close: 108.9},
		// Index 6-8: Morning Star sequence
		{Open: 120, High: 125, Low: 105, Close: 108}, // 6: Large bearish
		{Open: 102, High: 104, Low: 100, Close: 103}, // 7: Small body, gap down
		{Open: 108, High: 125, Low: 106, Close: 122}, // 8: Large bullish, gap up
		// Index 9-11: Evening Star sequence
		{Open: 122, High: 140, Low: 120, Close: 138}, // 9: Large bullish
		{Open: 142, High: 144, Low: 140, Close: 143}, // 10: Small body, gap up
		{Open: 138, High: 140, Low: 115, Close: 118}, // 11: Large bearish, gap down
		// Index 12: Bullish Marubozu (no shadows)
		{Open: 120, High: 135, Low: 120, Close: 135},
		// Index 13: Bearish Marubozu (no shadows)
		{Open: 135, High: 135, Low: 115, Close: 115},
		// Index 14: Spinning Top (small body, long shadows)
		{Open: 118, High: 125, Low: 110, Close: 119},
		// Index 15: Setup for Piercing Line - bearish candle
		{Open: 120, High: 121, Low: 115, Close: 115},
		// Index 16: Piercing Line - bullish candle opening below prev low, closing above midpoint
		{Open: 112, High: 119, Low: 112, Close: 118}, // Opens below 115, closes above midpoint (117.5)
		// Index 17: Setup for Dark Cloud Cover - bullish candle
		{Open: 118, High: 125, Low: 118, Close: 125},
		// Index 18: Dark Cloud Cover - bearish candle opening above prev high, closing below midpoint
		{Open: 127, High: 127, Low: 120, Close: 121}, // Opens above 125, closes below midpoint (121.5)
		// Index 19: Setup for Tweezer Bottom - bearish with low at 100
		{Open: 125, High: 126, Low: 100, Close: 102},
		// Index 20: Tweezer Bottom - bullish with same low at 100
		{Open: 102, High: 108, Low: 100, Close: 107},
		// Index 21-23: Three White Soldiers sequence
		{Open: 110, High: 115, Low: 109, Close: 114}, // 21: First soldier
		{Open: 113, High: 118, Low: 112, Close: 117}, // 22: Second soldier
		{Open: 116, High: 121, Low: 115, Close: 120}, // 23: Third soldier
		// Index 24-26: Three Black Crows sequence
		{Open: 120, High: 121, Low: 115, Close: 116}, // 24: First crow
		{Open: 117, High: 118, Low: 112, Close: 113}, // 25: Second crow
		{Open: 114, High: 115, Low: 108, Close: 109}, // 26: Third crow
	}

	opt := PatternsAll()
	opt.DojiThreshold = 0.01
	opt.ShadowRatio = 2.0
	opt.EngulfingMinSize = 0.8
	indexPatterns := scanForCandlestickPatterns(data, *opt)

	// Verify specific patterns were detected
	patternsByIndex := make(map[int][]string)
	uniquePatterns := make(map[string]bool)
	for index, patterns := range indexPatterns {
		for _, pattern := range patterns {
			patternsByIndex[index] = append(patternsByIndex[index], pattern.PatternType)
			uniquePatterns[pattern.PatternName] = true
		}
	}

	// Check expected patterns
	assert.Len(t, uniquePatterns, 19)
	assert.Contains(t, patternsByIndex[1], CandlestickPatternDoji)
	assert.Contains(t, patternsByIndex[2], CandlestickPatternHammer)
	assert.Contains(t, patternsByIndex[3], CandlestickPatternShootingStar)
	assert.Contains(t, patternsByIndex[4], CandlestickPatternGravestone)
	assert.Contains(t, patternsByIndex[5], CandlestickPatternDragonfly)
	assert.Contains(t, patternsByIndex[8], CandlestickPatternMorningStar)
	assert.Contains(t, patternsByIndex[11], CandlestickPatternEveningStar)
	assert.Contains(t, patternsByIndex[12], CandlestickPatternMarubozuBull)
	assert.Contains(t, patternsByIndex[13], CandlestickPatternMarubozuBear)
	assert.Contains(t, patternsByIndex[14], CandlestickPatternSpinningTop)
	assert.Contains(t, patternsByIndex[16], CandlestickPatternPiercingLine)
	assert.Contains(t, patternsByIndex[18], CandlestickPatternDarkCloudCover)
}

// TestPatternConfigIntegration tests the new PatternConfig API integration
func TestPatternConfigIntegration(t *testing.T) {
	t.Parallel()

	data := []OHLCData{
		{Open: 100, High: 110, Low: 95, Close: 100.1}, // Doji at index 0
		{Open: 105, High: 115, Low: 100, Close: 112},  // Normal bullish
		{Open: 112, High: 118, Low: 105, Close: 108},  // Bearish
		{Open: 108, High: 125, Low: 107, Close: 109},  // Shooting star at index 3
	}

	tests := []struct {
		name          string
		data          []OHLCData
		config        *CandlestickPatternConfig
		errorExpected bool
	}{
		{"patterns_disabled_nil_config", data, nil, false},
		{"patterns_disabled_empty_list", data, &CandlestickPatternConfig{
			ReplaceSeriesLabel: true,
			EnabledPatterns:    []string{},
			DojiThreshold:      0.001, ShadowTolerance: 0.01, BodySizeRatio: 0.3, ShadowRatio: 2.0, EngulfingMinSize: 0.8,
		}, false},
		{"doji_only_replace_mode", data, &CandlestickPatternConfig{
			ReplaceSeriesLabel: true,
			EnabledPatterns:    []string{CandlestickPatternDoji},
			DojiThreshold:      0.001, ShadowTolerance: 0.01, BodySizeRatio: 0.3, ShadowRatio: 2.0, EngulfingMinSize: 0.8,
		}, false},
		{"multiple_patterns_complement_mode", data, &CandlestickPatternConfig{
			ReplaceSeriesLabel: false,
			EnabledPatterns:    []string{CandlestickPatternDoji, CandlestickPatternShootingStar},
			DojiThreshold:      0.001, ShadowTolerance: 0.01, BodySizeRatio: 0.3, ShadowRatio: 2.0, EngulfingMinSize: 0.8,
		}, false},
		{"all_patterns_enabled", data, PatternsAll(), false},
		{"important_patterns_only", data, PatternsImportant(), false},
		// Edge cases
		{"empty_data", []OHLCData{}, &CandlestickPatternConfig{
			ReplaceSeriesLabel: true,
			EnabledPatterns:    []string{CandlestickPatternDoji},
			DojiThreshold:      0.001, ShadowTolerance: 0.01, BodySizeRatio: 0.3, ShadowRatio: 2.0, EngulfingMinSize: 0.8,
		}, true},
		{"invalid_ohlc_data", []OHLCData{
			{Open: 0, High: 0, Low: 0, Close: 0},
			{Open: 100, High: 90, Low: 110, Close: 105},
		}, &CandlestickPatternConfig{
			ReplaceSeriesLabel: true,
			EnabledPatterns:    []string{CandlestickPatternDoji},
			DojiThreshold:      0.001, ShadowTolerance: 0.01, BodySizeRatio: 0.3, ShadowRatio: 2.0, EngulfingMinSize: 0.8,
		}, false},
		{"nil_enabled_patterns", data, &CandlestickPatternConfig{
			ReplaceSeriesLabel: true,
			EnabledPatterns:    nil,
			DojiThreshold:      0.001, ShadowTolerance: 0.01, BodySizeRatio: 0.3, ShadowRatio: 2.0, EngulfingMinSize: 0.8,
		}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			series := CandlestickSeries{
				Data:          test.data,
				PatternConfig: test.config,
			}

			opt := CandlestickChartOption{
				Padding:    NewBoxEqual(10),
				XAxis:      XAxisOption{Labels: []string{"1", "2", "3", "4"}},
				YAxis:      make([]YAxisOption, 1),
				SeriesList: CandlestickSeriesList{series},
			}

			p := NewPainter(PainterOptions{OutputFormat: ChartOutputSVG, Width: 800, Height: 600})
			err := p.CandlestickChart(opt)
			if test.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestConvenienceFunctions tests preset pattern configuration functions
func TestConvenienceFunctions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   *CandlestickPatternConfig
		validate func(*testing.T, *CandlestickPatternConfig)
	}{
		{
			name:   "patterns_all",
			config: PatternsAll(),
			validate: func(t *testing.T, config *CandlestickPatternConfig) {
				assert.True(t, config.ReplaceSeriesLabel)
				assert.Contains(t, config.EnabledPatterns, CandlestickPatternDoji)
				assert.Contains(t, config.EnabledPatterns, CandlestickPatternHammer)
				assert.GreaterOrEqual(t, len(config.EnabledPatterns), 10)
			},
		},
		{
			name:   "patterns_important",
			config: PatternsImportant(),
			validate: func(t *testing.T, config *CandlestickPatternConfig) {
				assert.True(t, config.ReplaceSeriesLabel)
				assert.Contains(t, config.EnabledPatterns, CandlestickPatternEngulfingBull)
				assert.Contains(t, config.EnabledPatterns, CandlestickPatternHammer)
				assert.LessOrEqual(t, len(config.EnabledPatterns), 8)
			},
		},
		{
			name:   "patterns_bullish",
			config: PatternsBullish(),
			validate: func(t *testing.T, config *CandlestickPatternConfig) {
				assert.True(t, config.ReplaceSeriesLabel)
				assert.Contains(t, config.EnabledPatterns, CandlestickPatternHammer)
				assert.NotContains(t, config.EnabledPatterns, CandlestickPatternShootingStar)
			},
		},
		{
			name: "enable_patterns_custom",
			config: &CandlestickPatternConfig{
				ReplaceSeriesLabel: true,
				EnabledPatterns:    []string{CandlestickPatternDoji, CandlestickPatternHammer},
				DojiThreshold:      0.001, ShadowTolerance: 0.01, BodySizeRatio: 0.3, ShadowRatio: 2.0, EngulfingMinSize: 0.8,
			},
			validate: func(t *testing.T, config *CandlestickPatternConfig) {
				assert.True(t, config.ReplaceSeriesLabel)
				assert.Len(t, config.EnabledPatterns, 2)
				assert.Contains(t, config.EnabledPatterns, CandlestickPatternDoji)
				assert.Contains(t, config.EnabledPatterns, CandlestickPatternHammer)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.NotNil(t, test.config)
			test.validate(t, test.config)
		})
	}
}

func TestCandlestickChartPatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		optGen func() CandlestickChartOption
		svg    string
		pngCRC uint32
	}{
		{
			name: "doji",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 105, High: 107, Low: 103, Close: 105}, // Pure Doji pattern - minimal body and minimal shadows
					{Open: 105, High: 112, Low: 98, Close: 108},  // Normal candle
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
				return CandlestickChartOption{
					XAxis:      XAxisOption{Labels: []string{"1", "2", "3"}},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">112.85</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110.87</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">108.88</text><text x=\"17\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">106.9</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">104.92</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">102.93</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100.95</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">98.97</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">96.98</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 308 569\nL 308 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 549 569\nL 549 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"183\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"665\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path  d=\"M 187 99\nL 187 254\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 187 409\nL 187 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 99\nL 235 99\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 564\nL 235 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 254\nL 283 254\nL 283 409\nL 91 409\nL 91 254\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 428 192\nL 428 254\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 428 254\nL 428 316\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 192\nL 476 192\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 316\nL 476 316\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 332 254\nL 524 254\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 669 37\nL 669 161\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 669 254\nL 669 471\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 37\nL 717 37\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 471\nL 717 471\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 573 161\nL 765 161\nL 765 254\nL 573 254\nL 573 161\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 222\nL 543 222\nL 543 222\nA 4 4 90.00 0 1 547 226\nL 547 278\nL 547 278\nA 4 4 90.00 0 1 543 282\nL 433 282\nL 433 282\nA 4 4 90.00 0 1 429 278\nL 429 226\nL 429 226\nA 4 4 90.00 0 1 433 222\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"471\" y=\"239\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text><text x=\"447\" y=\"252\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">θ Bear Harami</text><text x=\"445\" y=\"265\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"433\" y=\"278\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">‡ Long Legged Doji</text><path  d=\"M 674 142\nL 760 142\nL 760 142\nA 4 4 90.00 0 1 764 146\nL 764 172\nL 764 172\nA 4 4 90.00 0 1 760 176\nL 674 176\nL 674 176\nA 4 4 90.00 0 1 670 172\nL 670 146\nL 670 146\nA 4 4 90.00 0 1 674 142\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"674\" y=\"159\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"680\" y=\"172\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text></svg>",
			pngCRC: 0x1a5504a9,
		},
		{
			name: "hammer",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 108, High: 109, Low: 98, Close: 107},  // Hammer pattern
					{Open: 107, High: 112, Low: 102, Close: 110}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
				return CandlestickChartOption{
					XAxis:      XAxisOption{Labels: []string{"1", "2", "3"}},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">112.85</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110.87</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">108.88</text><text x=\"17\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">106.9</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">104.92</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">102.93</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100.95</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">98.97</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">96.98</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 308 569\nL 308 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 549 569\nL 549 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"183\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"665\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path  d=\"M 187 99\nL 187 254\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 187 409\nL 187 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 99\nL 235 99\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 564\nL 235 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 254\nL 283 254\nL 283 409\nL 91 409\nL 91 254\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 428 130\nL 428 161\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 428 192\nL 428 471\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 380 130\nL 476 130\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 380 471\nL 476 471\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 332 161\nL 524 161\nL 524 192\nL 332 192\nL 332 161\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 669 37\nL 669 99\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 669 192\nL 669 347\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 37\nL 717 37\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 347\nL 717 347\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 573 99\nL 765 99\nL 765 192\nL 573 192\nL 573 99\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 166\nL 519 166\nL 519 166\nA 4 4 90.00 0 1 523 170\nL 523 209\nL 523 209\nA 4 4 90.00 0 1 519 213\nL 433 213\nL 433 213\nA 4 4 90.00 0 1 429 209\nL 429 170\nL 429 170\nA 4 4 90.00 0 1 433 166\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"446\" y=\"183\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><text x=\"433\" y=\"196\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"439\" y=\"209\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text></svg>",
			pngCRC: 0xff3283b4,
		},
		{
			name: "inverted_hammer",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105}, // Normal candle
					{Open: 95, High: 107, Low: 94, Close: 96},   // Inverted hammer
					{Open: 96, High: 102, Low: 91, Close: 98},   // Normal candle
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
				return CandlestickChartOption{
					XAxis:      XAxisOption{Labels: []string{"1", "2", "3"}},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">111</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">108.67</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">106.33</text><text x=\"30\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">104</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">101.67</text><text x=\"17\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">99.33</text><text x=\"39\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">97</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94.67</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">92.33</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 308 569\nL 308 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 549 569\nL 549 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"183\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"665\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path  d=\"M 187 37\nL 187 169\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 187 301\nL 187 433\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 37\nL 235 37\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 433\nL 235 433\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 169\nL 283 169\nL 283 301\nL 91 301\nL 91 169\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 428 116\nL 428 406\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 428 433\nL 428 459\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 116\nL 476 116\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 459\nL 476 459\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 332 406\nL 524 406\nL 524 433\nL 332 433\nL 332 406\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 669 248\nL 669 353\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 669 406\nL 669 538\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 248\nL 717 248\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 538\nL 717 538\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 573 353\nL 765 353\nL 765 406\nL 573 406\nL 573 353\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 374\nL 525 374\nL 525 374\nA 4 4 90.00 0 1 529 378\nL 529 430\nL 529 430\nA 4 4 90.00 0 1 525 434\nL 433 434\nL 433 434\nA 4 4 90.00 0 1 429 430\nL 429 378\nL 429 378\nA 4 4 90.00 0 1 433 374\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"433\" y=\"391\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"437\" y=\"404\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text><text x=\"436\" y=\"417\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"442\" y=\"430\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text><path  d=\"M 674 334\nL 760 334\nL 760 334\nA 4 4 90.00 0 1 764 338\nL 764 364\nL 764 364\nA 4 4 90.00 0 1 760 368\nL 674 368\nL 674 368\nA 4 4 90.00 0 1 670 364\nL 670 338\nL 670 338\nA 4 4 90.00 0 1 674 334\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"674\" y=\"351\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"680\" y=\"364\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text></svg>",
			pngCRC: 0x604d0205,
		},
		{
			name: "shooting_star",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 107, High: 125, Low: 106, Close: 108}, // Shooting star - small body at bottom, long upper shadow
					{Open: 107, High: 112, Low: 102, Close: 109}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
				return CandlestickChartOption{
					XAxis:      XAxisOption{Labels: []string{"1", "2", "3"}},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125.56</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">121.11</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">116.67</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">112.22</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">107.78</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">103.33</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">98.89</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94.44</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 308 569\nL 308 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 549 569\nL 549 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"183\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"665\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path  d=\"M 187 287\nL 187 357\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 187 426\nL 187 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 287\nL 235 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 495\nL 235 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 357\nL 283 357\nL 283 426\nL 91 426\nL 91 357\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 428 80\nL 428 315\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 428 329\nL 428 343\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 80\nL 476 80\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 343\nL 476 343\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 332 315\nL 524 315\nL 524 329\nL 332 329\nL 332 315\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 669 260\nL 669 301\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 669 329\nL 669 398\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 260\nL 717 260\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 398\nL 717 398\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 573 301\nL 765 301\nL 765 329\nL 573 329\nL 573 301\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 283\nL 525 283\nL 525 283\nA 4 4 90.00 0 1 529 287\nL 529 339\nL 529 339\nA 4 4 90.00 0 1 525 343\nL 433 343\nL 433 343\nA 4 4 90.00 0 1 429 339\nL 429 287\nL 429 287\nA 4 4 90.00 0 1 433 283\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"433\" y=\"300\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"437\" y=\"313\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text><text x=\"436\" y=\"326\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"442\" y=\"339\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text><path  d=\"M 674 282\nL 760 282\nL 760 282\nA 4 4 90.00 0 1 764 286\nL 764 312\nL 764 312\nA 4 4 90.00 0 1 760 316\nL 674 316\nL 674 316\nA 4 4 90.00 0 1 670 312\nL 670 286\nL 670 286\nA 4 4 90.00 0 1 674 282\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"674\" y=\"299\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"680\" y=\"312\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text></svg>",
			pngCRC: 0xfc4cd46f,
		},
		{
			name: "gravestone_doji",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 108, High: 125, Low: 108, Close: 108}, // Gravestone doji - minimal body at bottom, long upper shadow only
					{Open: 108, High: 115, Low: 103, Close: 110}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
				return CandlestickChartOption{
					XAxis:      XAxisOption{Labels: []string{"1", "2", "3"}},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125.56</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">121.11</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">116.67</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">112.22</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">107.78</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">103.33</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">98.89</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94.44</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 308 569\nL 308 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 549 569\nL 549 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"183\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"665\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path  d=\"M 187 287\nL 187 357\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 187 426\nL 187 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 287\nL 235 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 495\nL 235 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 357\nL 283 357\nL 283 426\nL 91 426\nL 91 357\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 428 80\nL 428 315\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 80\nL 476 80\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 315\nL 476 315\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 332 315\nL 524 315\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 669 218\nL 669 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 669 315\nL 669 384\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 218\nL 717 218\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 384\nL 717 384\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 573 287\nL 765 287\nL 765 315\nL 573 315\nL 573 287\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 276\nL 525 276\nL 525 276\nA 4 4 90.00 0 1 529 280\nL 529 345\nL 529 345\nA 4 4 90.00 0 1 525 349\nL 433 349\nL 433 349\nA 4 4 90.00 0 1 429 345\nL 429 280\nL 429 280\nA 4 4 90.00 0 1 433 276\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"433\" y=\"293\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"441\" y=\"306\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">† Gravestone</text><text x=\"462\" y=\"319\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text><text x=\"437\" y=\"332\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text><text x=\"436\" y=\"345\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 674 268\nL 760 268\nL 760 268\nA 4 4 90.00 0 1 764 272\nL 764 298\nL 764 298\nA 4 4 90.00 0 1 760 302\nL 674 302\nL 674 302\nA 4 4 90.00 0 1 670 298\nL 670 272\nL 670 272\nA 4 4 90.00 0 1 674 268\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"674\" y=\"285\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"680\" y=\"298\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text></svg>",
			pngCRC: 0xb77cb36,
		},
		{
			name: "dragonfly_doji",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 109, High: 110, Low: 90, Close: 109},  // Dragonfly doji
					{Open: 109, High: 115, Low: 104, Close: 112}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
				return CandlestickChartOption{
					XAxis:      XAxisOption{Labels: []string{"1", "2", "3"}},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">116.25</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">113.33</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110.42</text><text x=\"17\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">107.5</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">104.58</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">101.67</text><text x=\"17\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">98.75</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95.83</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">92.92</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 308 569\nL 308 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 549 569\nL 549 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"183\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"665\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path  d=\"M 187 142\nL 187 248\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 187 353\nL 187 459\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 142\nL 235 142\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 459\nL 235 459\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 248\nL 283 248\nL 283 353\nL 91 353\nL 91 248\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 428 142\nL 428 164\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 428 164\nL 428 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 142\nL 476 142\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 564\nL 476 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 332 164\nL 524 164\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 669 37\nL 669 100\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 669 164\nL 669 269\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 37\nL 717 37\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 269\nL 717 269\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 573 100\nL 765 100\nL 765 164\nL 573 164\nL 573 100\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 132\nL 519 132\nL 519 132\nA 4 4 90.00 0 1 523 136\nL 523 188\nL 523 188\nA 4 4 90.00 0 1 519 192\nL 433 192\nL 433 192\nA 4 4 90.00 0 1 429 188\nL 429 136\nL 429 136\nA 4 4 90.00 0 1 433 132\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"446\" y=\"149\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><text x=\"442\" y=\"162\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ψ Dragonfly</text><text x=\"459\" y=\"175\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text><text x=\"433\" y=\"188\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 674 87\nL 760 87\nL 760 87\nA 4 4 90.00 0 1 764 91\nL 764 104\nL 764 104\nA 4 4 90.00 0 1 760 108\nL 674 108\nL 674 108\nA 4 4 90.00 0 1 670 104\nL 670 91\nL 670 91\nA 4 4 90.00 0 1 674 87\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"674\" y=\"104\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text></svg>",
			pngCRC: 0xb61bd63d,
		},
		{
			name: "bullish_marubozu",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 100, High: 120, Low: 100, Close: 120}, // Bullish marubozu
					{Open: 120, High: 125, Low: 115, Close: 122}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
				return CandlestickChartOption{
					XAxis:      XAxisOption{Labels: []string{"1", "2", "3"}},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125.56</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">121.11</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">116.67</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">112.22</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">107.78</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">103.33</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">98.89</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94.44</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 308 569\nL 308 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 549 569\nL 549 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"183\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"665\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path  d=\"M 187 287\nL 187 357\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 187 426\nL 187 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 287\nL 235 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 495\nL 235 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 357\nL 283 357\nL 283 426\nL 91 426\nL 91 357\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 380 149\nL 476 149\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 426\nL 476 426\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 332 149\nL 524 149\nL 524 426\nL 332 426\nL 332 149\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 669 80\nL 669 121\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 669 149\nL 669 218\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 80\nL 717 80\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 218\nL 717 218\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 573 121\nL 765 121\nL 765 149\nL 573 149\nL 573 121\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 130\nL 525 130\nL 525 130\nA 4 4 90.00 0 1 529 134\nL 529 160\nL 529 160\nA 4 4 90.00 0 1 525 164\nL 433 164\nL 433 164\nA 4 4 90.00 0 1 429 160\nL 429 134\nL 429 134\nA 4 4 90.00 0 1 433 130\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"433\" y=\"147\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">^ Bull Marubozu</text><text x=\"437\" y=\"160\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">[ Bull Belt Hold</text><path  d=\"M 674 102\nL 760 102\nL 760 102\nA 4 4 90.00 0 1 764 106\nL 764 132\nL 764 132\nA 4 4 90.00 0 1 760 136\nL 674 136\nL 674 136\nA 4 4 90.00 0 1 670 132\nL 670 106\nL 670 106\nA 4 4 90.00 0 1 674 102\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"674\" y=\"119\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"680\" y=\"132\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text></svg>",
			pngCRC: 0x72681902,
		},
		{
			name: "bearish_marubozu",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 120, High: 120, Low: 100, Close: 100}, // Bearish marubozu
					{Open: 100, High: 105, Low: 95, Close: 102},  // Normal candle
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
				return CandlestickChartOption{
					XAxis:      XAxisOption{Labels: []string{"1", "2", "3"}},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">121.11</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">117.22</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">113.33</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">109.44</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105.56</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">101.67</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">97.78</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">93.89</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 308 569\nL 308 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 549 569\nL 549 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"183\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"665\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path  d=\"M 187 248\nL 187 327\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 187 406\nL 187 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 248\nL 235 248\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 485\nL 235 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 327\nL 283 327\nL 283 406\nL 91 406\nL 91 327\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 380 90\nL 476 90\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 380 406\nL 476 406\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 332 90\nL 524 90\nL 524 406\nL 332 406\nL 332 90\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 669 327\nL 669 375\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 669 406\nL 669 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 327\nL 717 327\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 485\nL 717 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 573 375\nL 765 375\nL 765 406\nL 573 406\nL 573 375\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 387\nL 530 387\nL 530 387\nA 4 4 90.00 0 1 534 391\nL 534 417\nL 534 417\nA 4 4 90.00 0 1 530 421\nL 433 421\nL 433 421\nA 4 4 90.00 0 1 429 417\nL 429 391\nL 429 391\nA 4 4 90.00 0 1 433 387\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"433\" y=\"404\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">v Bear Marubozu</text><text x=\"437\" y=\"417\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">] Bear Belt Hold</text><path  d=\"M 674 349\nL 760 349\nL 760 349\nA 4 4 90.00 0 1 764 353\nL 764 392\nL 764 392\nA 4 4 90.00 0 1 760 396\nL 674 396\nL 674 396\nA 4 4 90.00 0 1 670 392\nL 670 353\nL 670 353\nA 4 4 90.00 0 1 674 349\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"679\" y=\"366\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ʘ Bull Harami</text><text x=\"674\" y=\"379\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"680\" y=\"392\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text></svg>",
			pngCRC: 0xfc31b740,
		},
		{
			name: "high_wave",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105}, // Normal candle
					{Open: 102, High: 120, Low: 85, Close: 104}, // High wave pattern - small body, very long shadows
					{Open: 104, High: 110, Low: 99, Close: 107}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
				return CandlestickChartOption{
					XAxis:      XAxisOption{Labels: []string{"1", "2", "3"}},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">115</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"18\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><text x=\"18\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><text x=\"18\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">85</text><text x=\"18\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">80</text><path  d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 46 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 46 569\nL 46 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 294 569\nL 294 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 542 569\nL 542 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"166\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"414\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"662\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path  d=\"M 170 195\nL 170 257\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 170 318\nL 170 380\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 121 195\nL 219 195\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 121 380\nL 219 380\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 71 257\nL 269 257\nL 269 318\nL 71 318\nL 71 257\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 418 72\nL 418 269\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 418 294\nL 418 503\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 369 72\nL 467 72\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 369 503\nL 467 503\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 319 269\nL 517 269\nL 517 294\nL 319 294\nL 319 269\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 666 195\nL 666 232\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 666 269\nL 666 331\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 617 195\nL 715 195\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 617 331\nL 715 331\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 567 232\nL 765 232\nL 765 269\nL 567 269\nL 567 232\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 423 250\nL 509 250\nL 509 250\nA 4 4 90.00 0 1 513 254\nL 513 280\nL 513 280\nA 4 4 90.00 0 1 509 284\nL 423 284\nL 423 284\nA 4 4 90.00 0 1 419 280\nL 419 254\nL 419 254\nA 4 4 90.00 0 1 423 250\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"423\" y=\"267\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"429\" y=\"280\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text><path  d=\"M 671 219\nL 757 219\nL 757 219\nA 4 4 90.00 0 1 761 223\nL 761 236\nL 761 236\nA 4 4 90.00 0 1 757 240\nL 671 240\nL 671 240\nA 4 4 90.00 0 1 667 236\nL 667 223\nL 667 223\nA 4 4 90.00 0 1 671 219\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"671\" y=\"236\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text></svg>",
			pngCRC: 0xd1fa1a2c,
		},
		{
			name: "bullish_belt_hold",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 100, High: 120, Low: 100, Close: 120}, // Bullish belt hold - also detects as marubozu (expected)
					{Open: 120, High: 125, Low: 115, Close: 122}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
				return CandlestickChartOption{
					XAxis:      XAxisOption{Labels: []string{"1", "2", "3"}},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125.56</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">121.11</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">116.67</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">112.22</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">107.78</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">103.33</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">98.89</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94.44</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 308 569\nL 308 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 549 569\nL 549 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"183\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"665\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path  d=\"M 187 287\nL 187 357\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 187 426\nL 187 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 287\nL 235 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 495\nL 235 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 357\nL 283 357\nL 283 426\nL 91 426\nL 91 357\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 380 149\nL 476 149\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 426\nL 476 426\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 332 149\nL 524 149\nL 524 426\nL 332 426\nL 332 149\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 669 80\nL 669 121\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 669 149\nL 669 218\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 80\nL 717 80\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 218\nL 717 218\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 573 121\nL 765 121\nL 765 149\nL 573 149\nL 573 121\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 130\nL 525 130\nL 525 130\nA 4 4 90.00 0 1 529 134\nL 529 160\nL 529 160\nA 4 4 90.00 0 1 525 164\nL 433 164\nL 433 164\nA 4 4 90.00 0 1 429 160\nL 429 134\nL 429 134\nA 4 4 90.00 0 1 433 130\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"433\" y=\"147\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">^ Bull Marubozu</text><text x=\"437\" y=\"160\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">[ Bull Belt Hold</text><path  d=\"M 674 102\nL 760 102\nL 760 102\nA 4 4 90.00 0 1 764 106\nL 764 132\nL 764 132\nA 4 4 90.00 0 1 760 136\nL 674 136\nL 674 136\nA 4 4 90.00 0 1 670 132\nL 670 106\nL 670 106\nA 4 4 90.00 0 1 674 102\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"674\" y=\"119\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"680\" y=\"132\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text></svg>",
			pngCRC: 0x72681902,
		},
		{
			name: "bearish_belt_hold",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 120, High: 120, Low: 100, Close: 100}, // Bearish belt hold - also detects as marubozu (expected)
					{Open: 100, High: 105, Low: 95, Close: 102},  // Normal candle - harami overlap is expected
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
				return CandlestickChartOption{
					XAxis:      XAxisOption{Labels: []string{"1", "2", "3"}},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">121.11</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">117.22</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">113.33</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">109.44</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105.56</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">101.67</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">97.78</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">93.89</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 308 569\nL 308 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 549 569\nL 549 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"183\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"665\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path  d=\"M 187 248\nL 187 327\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 187 406\nL 187 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 248\nL 235 248\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 485\nL 235 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 327\nL 283 327\nL 283 406\nL 91 406\nL 91 327\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 380 90\nL 476 90\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 380 406\nL 476 406\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 332 90\nL 524 90\nL 524 406\nL 332 406\nL 332 90\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 669 327\nL 669 375\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 669 406\nL 669 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 327\nL 717 327\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 485\nL 717 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 573 375\nL 765 375\nL 765 406\nL 573 406\nL 573 375\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 387\nL 530 387\nL 530 387\nA 4 4 90.00 0 1 534 391\nL 534 417\nL 534 417\nA 4 4 90.00 0 1 530 421\nL 433 421\nL 433 421\nA 4 4 90.00 0 1 429 417\nL 429 391\nL 429 391\nA 4 4 90.00 0 1 433 387\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"433\" y=\"404\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">v Bear Marubozu</text><text x=\"437\" y=\"417\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">] Bear Belt Hold</text><path  d=\"M 674 349\nL 760 349\nL 760 349\nA 4 4 90.00 0 1 764 353\nL 764 392\nL 764 392\nA 4 4 90.00 0 1 760 396\nL 674 396\nL 674 396\nA 4 4 90.00 0 1 670 392\nL 670 353\nL 670 353\nA 4 4 90.00 0 1 674 349\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"679\" y=\"366\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ʘ Bull Harami</text><text x=\"674\" y=\"379\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"680\" y=\"392\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text></svg>",
			pngCRC: 0xfc31b740,
		},
		{
			name: "spinning_top",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 110, High: 125, Low: 95, Close: 112},  // Spinning top pattern
					{Open: 112, High: 118, Low: 107, Close: 115}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
				return CandlestickChartOption{
					XAxis:      XAxisOption{Labels: []string{"1", "2", "3"}},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125.56</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">121.11</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">116.67</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">112.22</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">107.78</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">103.33</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">98.89</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94.44</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 308 569\nL 308 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 549 569\nL 549 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"183\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"665\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path  d=\"M 187 287\nL 187 357\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 187 426\nL 187 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 287\nL 235 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 495\nL 235 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 357\nL 283 357\nL 283 426\nL 91 426\nL 91 357\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 428 80\nL 428 260\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 428 287\nL 428 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 80\nL 476 80\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 495\nL 476 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 332 260\nL 524 260\nL 524 287\nL 332 287\nL 332 260\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 669 177\nL 669 218\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 669 260\nL 669 329\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 177\nL 717 177\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 329\nL 717 329\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 573 218\nL 765 218\nL 765 260\nL 573 260\nL 573 218\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 241\nL 519 241\nL 519 241\nA 4 4 90.00 0 1 523 245\nL 523 271\nL 523 271\nA 4 4 90.00 0 1 519 275\nL 433 275\nL 433 275\nA 4 4 90.00 0 1 429 271\nL 429 245\nL 429 245\nA 4 4 90.00 0 1 433 241\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"433\" y=\"258\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"439\" y=\"271\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text><path  d=\"M 674 205\nL 760 205\nL 760 205\nA 4 4 90.00 0 1 764 209\nL 764 222\nL 764 222\nA 4 4 90.00 0 1 760 226\nL 674 226\nL 674 226\nA 4 4 90.00 0 1 670 222\nL 670 209\nL 670 209\nA 4 4 90.00 0 1 674 205\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"674\" y=\"222\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text></svg>",
			pngCRC: 0xab11d07d,
		},
		{
			name: "bullish_engulfing",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 110, High: 112, Low: 105, Close: 106}, // Small bearish candle
					{Open: 104, High: 115, Low: 103, Close: 114}, // Bullish engulfing
					{Open: 114, High: 120, Low: 112, Close: 118}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold:    0.01,
					ShadowRatio:      2.0,
					EngulfingMinSize: 0.8,
				})
				return CandlestickChartOption{
					XAxis:      XAxisOption{Labels: []string{"1", "2", "3", "4"}},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">121.11</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">117.22</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">113.33</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">109.44</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105.56</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">101.67</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">97.78</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">93.89</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 247 569\nL 247 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 428 569\nL 428 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 609 569\nL 609 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"153\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"333\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"514\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"695\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><path  d=\"M 157 248\nL 157 327\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 157 406\nL 157 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 121 248\nL 193 248\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 121 485\nL 193 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 85 327\nL 229 327\nL 229 406\nL 85 406\nL 85 327\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 337 216\nL 337 248\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 337 311\nL 337 327\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 301 216\nL 373 216\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 301 327\nL 373 327\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 265 248\nL 409 248\nL 409 311\nL 265 311\nL 265 248\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 518 169\nL 518 185\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 518 343\nL 518 359\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 482 169\nL 554 169\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 482 359\nL 554 359\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 446 185\nL 590 185\nL 590 343\nL 446 343\nL 446 185\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 699 90\nL 699 121\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 699 185\nL 699 216\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 663 90\nL 735 90\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 663 216\nL 735 216\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 627 121\nL 771 121\nL 771 185\nL 627 185\nL 627 121\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 523 172\nL 614 172\nL 614 172\nA 4 4 90.00 0 1 618 176\nL 618 189\nL 618 189\nA 4 4 90.00 0 1 614 193\nL 523 193\nL 523 193\nA 4 4 90.00 0 1 519 189\nL 519 176\nL 519 176\nA 4 4 90.00 0 1 523 172\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"523\" y=\"189\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Λ Bull Engulfing</text></svg>",
			pngCRC: 0x5763c0bd,
		},
		{
			name: "bearish_engulfing",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 106, High: 112, Low: 105, Close: 110}, // Small bullish candle
					{Open: 114, High: 115, Low: 103, Close: 104}, // Bearish engulfing
					{Open: 104, High: 108, Low: 100, Close: 102}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold:    0.01,
					ShadowRatio:      2.0,
					EngulfingMinSize: 0.8,
				})
				return CandlestickChartOption{
					XAxis:      XAxisOption{Labels: []string{"1", "2", "3", "4"}},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">116</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">113.67</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">111.33</text><text x=\"30\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">109</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">106.67</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">104.33</text><text x=\"30\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">102</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">99.67</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">97.33</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 247 569\nL 247 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 428 569\nL 428 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 609 569\nL 609 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"153\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"333\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"514\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"695\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><path  d=\"M 157 169\nL 157 301\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 157 433\nL 157 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 121 169\nL 193 169\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 121 564\nL 193 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 85 301\nL 229 301\nL 229 433\nL 85 433\nL 85 301\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 337 116\nL 337 169\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 337 274\nL 337 301\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 301 116\nL 373 116\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 301 301\nL 373 301\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 265 169\nL 409 169\nL 409 274\nL 265 274\nL 265 169\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 518 37\nL 518 63\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 518 327\nL 518 353\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 482 37\nL 554 37\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 482 353\nL 554 353\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 446 63\nL 590 63\nL 590 327\nL 446 327\nL 446 63\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 699 222\nL 699 327\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 699 380\nL 699 433\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 663 222\nL 735 222\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 663 433\nL 735 433\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 627 327\nL 771 327\nL 771 380\nL 627 380\nL 627 327\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 523 314\nL 618 314\nL 618 314\nA 4 4 90.00 0 1 622 318\nL 622 331\nL 622 331\nA 4 4 90.00 0 1 618 335\nL 523 335\nL 523 335\nA 4 4 90.00 0 1 519 331\nL 519 318\nL 519 318\nA 4 4 90.00 0 1 523 314\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"523\" y=\"331\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">V Bear Engulfing</text><path  d=\"M 704 361\nL 790 361\nL 790 361\nA 4 4 90.00 0 1 794 365\nL 794 391\nL 794 391\nA 4 4 90.00 0 1 790 395\nL 704 395\nL 704 395\nA 4 4 90.00 0 1 700 391\nL 700 365\nL 700 365\nA 4 4 90.00 0 1 704 361\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"704\" y=\"378\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"710\" y=\"391\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text></svg>",
			pngCRC: 0x40825688,
		},
		{
			name: "bullish_harami",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 110, High: 115, Low: 95, Close: 98},   // Large bearish candle
					{Open: 102, High: 106, Low: 100, Close: 104}, // Small bullish candle within previous body (bullish harami)
					{Open: 104, High: 109, Low: 101, Close: 106}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold:    0.01,
					ShadowRatio:      2.0,
					EngulfingMinSize: 0.8,
				})
				return CandlestickChartOption{
					XAxis:      XAxisOption{Labels: []string{"1", "2", "3", "4"}},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">116</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">113.67</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">111.33</text><text x=\"30\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">109</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">106.67</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">104.33</text><text x=\"30\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">102</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">99.67</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">97.33</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 247 569\nL 247 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 428 569\nL 428 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 609 569\nL 609 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"153\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"333\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"514\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"695\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><path  d=\"M 157 169\nL 157 301\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 157 433\nL 157 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 121 169\nL 193 169\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 121 564\nL 193 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 85 301\nL 229 301\nL 229 433\nL 85 433\nL 85 301\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 337 37\nL 337 169\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 337 485\nL 337 564\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 301 37\nL 373 37\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 301 564\nL 373 564\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 265 169\nL 409 169\nL 409 485\nL 265 485\nL 265 169\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 518 274\nL 518 327\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 518 380\nL 518 433\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 482 274\nL 554 274\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 482 433\nL 554 433\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 446 327\nL 590 327\nL 590 380\nL 446 380\nL 446 327\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 699 195\nL 699 274\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 699 327\nL 699 406\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 663 195\nL 735 195\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 663 406\nL 735 406\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 627 274\nL 771 274\nL 771 327\nL 627 327\nL 627 274\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 342 472\nL 437 472\nL 437 472\nA 4 4 90.00 0 1 441 476\nL 441 489\nL 441 489\nA 4 4 90.00 0 1 437 493\nL 342 493\nL 342 493\nA 4 4 90.00 0 1 338 489\nL 338 476\nL 338 476\nA 4 4 90.00 0 1 342 472\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"342\" y=\"489\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">V Bear Engulfing</text><path  d=\"M 523 314\nL 599 314\nL 599 314\nA 4 4 90.00 0 1 603 318\nL 603 331\nL 603 331\nA 4 4 90.00 0 1 599 335\nL 523 335\nL 523 335\nA 4 4 90.00 0 1 519 331\nL 519 318\nL 519 318\nA 4 4 90.00 0 1 523 314\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"523\" y=\"331\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ʘ Bull Harami</text><path  d=\"M 704 255\nL 790 255\nL 790 255\nA 4 4 90.00 0 1 794 259\nL 794 285\nL 794 285\nA 4 4 90.00 0 1 790 289\nL 704 289\nL 704 289\nA 4 4 90.00 0 1 700 285\nL 700 259\nL 700 259\nA 4 4 90.00 0 1 704 255\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"704\" y=\"272\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"710\" y=\"285\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text></svg>",
			pngCRC: 0xbbb404f5,
		},
		{
			name: "bearish_harami",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 98, High: 115, Low: 95, Close: 110},   // Large bullish candle
					{Open: 106, High: 108, Low: 102, Close: 104}, // Small bearish candle within previous body (bearish harami)
					{Open: 104, High: 109, Low: 101, Close: 106}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold:    0.01,
					ShadowRatio:      2.0,
					EngulfingMinSize: 0.8,
				})
				return CandlestickChartOption{
					XAxis:      XAxisOption{Labels: []string{"1", "2", "3", "4"}},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">116</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">113.67</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">111.33</text><text x=\"30\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">109</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">106.67</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">104.33</text><text x=\"30\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">102</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">99.67</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">97.33</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 247 569\nL 247 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 428 569\nL 428 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 609 569\nL 609 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"153\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"333\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"514\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"695\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><path  d=\"M 157 169\nL 157 301\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 157 433\nL 157 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 121 169\nL 193 169\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 121 564\nL 193 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 85 301\nL 229 301\nL 229 433\nL 85 433\nL 85 301\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 337 37\nL 337 169\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 337 485\nL 337 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 301 37\nL 373 37\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 301 564\nL 373 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 265 169\nL 409 169\nL 409 485\nL 265 485\nL 265 169\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 518 222\nL 518 274\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 518 327\nL 518 380\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 482 222\nL 554 222\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 482 380\nL 554 380\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 446 274\nL 590 274\nL 590 327\nL 446 327\nL 446 274\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 699 195\nL 699 274\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 699 327\nL 699 406\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 663 195\nL 735 195\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 663 406\nL 735 406\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 627 274\nL 771 274\nL 771 327\nL 627 327\nL 627 274\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 523 314\nL 605 314\nL 605 314\nA 4 4 90.00 0 1 609 318\nL 609 331\nL 609 331\nA 4 4 90.00 0 1 605 335\nL 523 335\nL 523 335\nA 4 4 90.00 0 1 519 331\nL 519 318\nL 519 318\nA 4 4 90.00 0 1 523 314\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"523\" y=\"331\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">θ Bear Harami</text><path  d=\"M 704 255\nL 790 255\nL 790 255\nA 4 4 90.00 0 1 794 259\nL 794 285\nL 794 285\nA 4 4 90.00 0 1 790 289\nL 704 289\nL 704 289\nA 4 4 90.00 0 1 700 285\nL 700 259\nL 700 259\nA 4 4 90.00 0 1 704 255\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"704\" y=\"272\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"710\" y=\"285\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text></svg>",
			pngCRC: 0x90e73222,
		},
		{
			name: "morning_star",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 120, High: 125, Low: 105, Close: 108}, // Large bearish
					{Open: 102, High: 104, Low: 100, Close: 103}, // Small body, gap down - overlaps are expected
					{Open: 108, High: 125, Low: 106, Close: 122}, // Large bullish, gap up
					{Open: 122, High: 128, Low: 120, Close: 125}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
				return CandlestickChartOption{
					XAxis:      XAxisOption{Labels: []string{"1", "2", "3", "4", "5"}},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">133</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">128.22</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">123.44</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">118.67</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">113.89</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">109.11</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">104.33</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">99.56</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94.78</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 211 569\nL 211 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 356 569\nL 356 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 500 569\nL 500 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 645 569\nL 645 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"135\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"279\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"568\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><text x=\"713\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">5</text><path  d=\"M 139 307\nL 139 371\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 436\nL 139 500\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 111 307\nL 167 307\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 111 500\nL 167 500\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 82 371\nL 196 371\nL 196 436\nL 82 436\nL 82 371\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 283 114\nL 283 178\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 283 333\nL 283 371\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 255 114\nL 311 114\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 255 371\nL 311 371\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 226 178\nL 340 178\nL 340 333\nL 226 333\nL 226 178\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 428 384\nL 428 397\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 428 410\nL 428 436\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 400 384\nL 456 384\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 400 436\nL 456 436\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 371 397\nL 485 397\nL 485 410\nL 371 410\nL 371 397\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 572 114\nL 572 152\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 572 333\nL 572 358\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 544 114\nL 600 114\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 544 358\nL 600 358\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 515 152\nL 629 152\nL 629 333\nL 515 333\nL 515 152\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 717 75\nL 717 114\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 717 152\nL 717 178\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 689 75\nL 745 75\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 689 178\nL 745 178\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 660 114\nL 774 114\nL 774 152\nL 660 152\nL 660 114\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 378\nL 519 378\nL 519 378\nA 4 4 90.00 0 1 523 382\nL 523 408\nL 523 408\nA 4 4 90.00 0 1 519 412\nL 433 412\nL 433 412\nA 4 4 90.00 0 1 429 408\nL 429 382\nL 429 382\nA 4 4 90.00 0 1 433 378\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"433\" y=\"395\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"439\" y=\"408\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text><path  d=\"M 577 139\nL 661 139\nL 661 139\nA 4 4 90.00 0 1 665 143\nL 665 156\nL 665 156\nA 4 4 90.00 0 1 661 160\nL 577 160\nL 577 160\nA 4 4 90.00 0 1 573 156\nL 573 143\nL 573 143\nA 4 4 90.00 0 1 577 139\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"577\" y=\"156\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">* Morning Star</text></svg>",
			pngCRC: 0x90cb3552,
		},
		{
			name: "evening_star",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 122, High: 140, Low: 120, Close: 138}, // Large bullish
					{Open: 142, High: 144, Low: 140, Close: 143}, // Small body, gap up - overlaps are expected
					{Open: 138, High: 140, Low: 115, Close: 118}, // Large bearish, gap down
					{Open: 118, High: 122, Low: 115, Close: 120}, // Normal candle - harami overlap is expected
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
				return CandlestickChartOption{
					XAxis:      XAxisOption{Labels: []string{"1", "2", "3", "4", "5"}},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">149</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">142.44</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">135.89</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">129.33</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">122.78</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">116.22</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">109.67</text><text x=\"9\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">103.11</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">96.56</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 211 569\nL 211 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 356 569\nL 356 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 500 569\nL 500 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 645 569\nL 645 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"135\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"279\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"568\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><text x=\"713\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">5</text><path  d=\"M 139 377\nL 139 424\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 471\nL 139 518\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 111 377\nL 167 377\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 111 518\nL 167 518\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 82 424\nL 196 424\nL 196 471\nL 82 471\nL 82 424\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 283 95\nL 283 114\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 283 264\nL 283 283\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 255 95\nL 311 95\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 255 283\nL 311 283\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 226 114\nL 340 114\nL 340 264\nL 226 264\nL 226 114\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 428 57\nL 428 67\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 428 76\nL 428 95\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 400 57\nL 456 57\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 400 95\nL 456 95\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 371 67\nL 485 67\nL 485 76\nL 371 76\nL 371 67\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 572 95\nL 572 114\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 572 302\nL 572 330\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 544 95\nL 600 95\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 544 330\nL 600 330\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 515 114\nL 629 114\nL 629 302\nL 515 302\nL 515 114\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 717 264\nL 717 283\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 717 302\nL 717 330\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 689 264\nL 745 264\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 689 330\nL 745 330\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 660 283\nL 774 283\nL 774 302\nL 660 302\nL 660 283\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 48\nL 519 48\nL 519 48\nA 4 4 90.00 0 1 523 52\nL 523 78\nL 523 78\nA 4 4 90.00 0 1 519 82\nL 433 82\nL 433 82\nA 4 4 90.00 0 1 429 78\nL 429 52\nL 429 52\nA 4 4 90.00 0 1 433 48\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"433\" y=\"65\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"439\" y=\"78\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text><path  d=\"M 577 289\nL 659 289\nL 659 289\nA 4 4 90.00 0 1 663 293\nL 663 306\nL 663 306\nA 4 4 90.00 0 1 659 310\nL 577 310\nL 577 310\nA 4 4 90.00 0 1 573 306\nL 573 293\nL 573 293\nA 4 4 90.00 0 1 577 289\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"577\" y=\"306\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">⁎ Evening Star</text><path  d=\"M 714 264\nL 800 264\nL 800 264\nA 4 4 90.00 0 1 804 268\nL 804 294\nL 804 294\nA 4 4 90.00 0 1 800 298\nL 714 298\nL 714 298\nA 4 4 90.00 0 1 710 294\nL 710 268\nL 710 268\nA 4 4 90.00 0 1 714 264\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"719\" y=\"281\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ʘ Bull Harami</text><text x=\"714\" y=\"294\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text></svg>",
			pngCRC: 0xa3bc4991,
		},
		{
			name: "piercing_line",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 120, High: 121, Low: 115, Close: 115}, // Bearish candle
					{Open: 112, High: 119, Low: 111, Close: 118}, // Piercing line (opens below prev low, closes above midpoint)
					{Open: 118, High: 125, Low: 116, Close: 122}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
				return CandlestickChartOption{
					XAxis:      XAxisOption{Labels: []string{"1", "2", "3", "4"}},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125.56</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">121.11</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">116.67</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">112.22</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">107.78</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">103.33</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">98.89</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94.44</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 247 569\nL 247 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 428 569\nL 428 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 609 569\nL 609 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"153\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"333\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"514\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"695\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><path  d=\"M 157 287\nL 157 357\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 157 426\nL 157 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 121 287\nL 193 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 121 495\nL 193 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 85 357\nL 229 357\nL 229 426\nL 85 426\nL 85 357\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 337 135\nL 337 149\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 301 135\nL 373 135\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 301 218\nL 373 218\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 265 149\nL 409 149\nL 409 218\nL 265 218\nL 265 149\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 518 163\nL 518 177\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 518 260\nL 518 274\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 482 163\nL 554 163\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 482 274\nL 554 274\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 446 177\nL 590 177\nL 590 260\nL 446 260\nL 446 177\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 699 80\nL 699 121\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 699 177\nL 699 204\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 663 80\nL 735 80\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 663 204\nL 735 204\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 627 121\nL 771 121\nL 771 177\nL 627 177\nL 627 121\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 523 164\nL 604 164\nL 604 164\nA 4 4 90.00 0 1 608 168\nL 608 181\nL 608 181\nA 4 4 90.00 0 1 604 185\nL 523 185\nL 523 185\nA 4 4 90.00 0 1 519 181\nL 519 168\nL 519 168\nA 4 4 90.00 0 1 523 164\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"523\" y=\"181\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">| Piercing Line</text></svg>",
			pngCRC: 0xdde1cf86,
		},
		{
			name: "dark_cloud_cover",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 118, High: 125, Low: 117, Close: 125}, // Bullish candle
					{Open: 127, High: 128, Low: 120, Close: 121}, // Dark cloud cover (opens above prev high, closes below midpoint)
					{Open: 121, High: 124, Low: 118, Close: 120}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
				return CandlestickChartOption{
					XAxis:      XAxisOption{Labels: []string{"1", "2", "3", "4"}},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">133</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">128.22</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">123.44</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">118.67</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">113.89</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">109.11</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">104.33</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">99.56</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94.78</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 247 569\nL 247 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 428 569\nL 428 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 609 569\nL 609 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"153\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"333\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"514\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"695\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><path  d=\"M 157 307\nL 157 371\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 157 436\nL 157 500\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 121 307\nL 193 307\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 121 500\nL 193 500\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 85 371\nL 229 371\nL 229 436\nL 85 436\nL 85 371\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 337 204\nL 337 217\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 301 114\nL 373 114\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 301 217\nL 373 217\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 265 114\nL 409 114\nL 409 204\nL 265 204\nL 265 114\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 518 75\nL 518 88\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 518 165\nL 518 178\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 482 75\nL 554 75\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 482 178\nL 554 178\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 446 88\nL 590 88\nL 590 165\nL 446 165\nL 446 88\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 699 126\nL 699 165\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 699 178\nL 699 204\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 663 126\nL 735 126\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 663 204\nL 735 204\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 627 165\nL 771 165\nL 771 178\nL 627 178\nL 627 165\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 523 152\nL 597 152\nL 597 152\nA 4 4 90.00 0 1 601 156\nL 601 169\nL 601 169\nA 4 4 90.00 0 1 597 173\nL 523 173\nL 523 173\nA 4 4 90.00 0 1 519 169\nL 519 156\nL 519 156\nA 4 4 90.00 0 1 523 152\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"523\" y=\"169\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ξ Dark Cloud</text><path  d=\"M 704 159\nL 790 159\nL 790 159\nA 4 4 90.00 0 1 794 163\nL 794 189\nL 794 189\nA 4 4 90.00 0 1 790 193\nL 704 193\nL 704 193\nA 4 4 90.00 0 1 700 189\nL 700 163\nL 700 163\nA 4 4 90.00 0 1 704 159\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"704\" y=\"176\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"710\" y=\"189\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text></svg>",
			pngCRC: 0xd47c89d4,
		},
		{
			name: "engulfing_and_stars",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 110, High: 112, Low: 105, Close: 106}, // Small bearish candle
					{Open: 104, High: 115, Low: 103, Close: 114}, // Bullish engulfing
					{Open: 120, High: 125, Low: 105, Close: 108}, // Large bearish (morning star setup)
					{Open: 102, High: 104, Low: 100, Close: 103}, // Small body, gap down
					{Open: 108, High: 125, Low: 106, Close: 122}, // Large bullish, gap up (morning star completion)
					{Open: 122, High: 128, Low: 120, Close: 125}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold:    0.01,
					ShadowRatio:      2.0,
					EngulfingMinSize: 0.8,
				})
				return CandlestickChartOption{
					XAxis:      XAxisOption{Show: Ptr(false)},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">133</text><text x=\"9\" y=\"80\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">128.22</text><text x=\"9\" y=\"144\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">123.44</text><text x=\"9\" y=\"208\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">118.67</text><text x=\"9\" y=\"272\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">113.89</text><text x=\"9\" y=\"337\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">109.11</text><text x=\"9\" y=\"401\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">104.33</text><text x=\"17\" y=\"465\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">99.56</text><text x=\"17\" y=\"529\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94.78</text><text x=\"39\" y=\"594\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 74\nL 790 74\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 138\nL 790 138\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 203\nL 790 203\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 267\nL 790 267\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 332\nL 790 332\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 396\nL 790 396\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 461\nL 790 461\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 525\nL 790 525\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 118 321\nL 118 388\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 118 456\nL 118 523\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 98 321\nL 138 321\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 98 523\nL 138 523\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 77 388\nL 159 388\nL 159 456\nL 77 456\nL 77 388\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 221 294\nL 221 321\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 221 375\nL 221 388\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 201 294\nL 241 294\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 201 388\nL 241 388\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 180 321\nL 262 321\nL 262 375\nL 180 375\nL 180 321\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 324 253\nL 324 267\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 324 402\nL 324 415\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 304 253\nL 344 253\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 304 415\nL 344 415\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 283 267\nL 365 267\nL 365 402\nL 283 402\nL 283 267\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 428 118\nL 428 186\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 428 348\nL 428 388\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 408 118\nL 448 118\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 408 388\nL 448 388\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 387 186\nL 469 186\nL 469 348\nL 387 348\nL 387 186\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 531 402\nL 531 415\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 531 429\nL 531 456\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 511 402\nL 551 402\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 511 456\nL 551 456\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 490 415\nL 572 415\nL 572 429\nL 490 429\nL 490 415\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 634 118\nL 634 159\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 634 348\nL 634 375\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 614 118\nL 654 118\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 614 375\nL 654 375\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 593 159\nL 675 159\nL 675 348\nL 593 348\nL 593 159\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 738 78\nL 738 118\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 738 159\nL 738 186\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 718 78\nL 758 78\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 718 186\nL 758 186\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 697 118\nL 779 118\nL 779 159\nL 697 159\nL 697 118\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 329 254\nL 420 254\nL 420 254\nA 4 4 90.00 0 1 424 258\nL 424 271\nL 424 271\nA 4 4 90.00 0 1 420 275\nL 329 275\nL 329 275\nA 4 4 90.00 0 1 325 271\nL 325 258\nL 325 258\nA 4 4 90.00 0 1 329 254\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"329\" y=\"271\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Λ Bull Engulfing</text><path  d=\"M 433 335\nL 507 335\nL 507 335\nA 4 4 90.00 0 1 511 339\nL 511 352\nL 511 352\nA 4 4 90.00 0 1 507 356\nL 433 356\nL 433 356\nA 4 4 90.00 0 1 429 352\nL 429 339\nL 429 339\nA 4 4 90.00 0 1 433 335\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"433\" y=\"352\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ξ Dark Cloud</text><path  d=\"M 536 396\nL 622 396\nL 622 396\nA 4 4 90.00 0 1 626 400\nL 626 426\nL 626 426\nA 4 4 90.00 0 1 622 430\nL 536 430\nL 536 430\nA 4 4 90.00 0 1 532 426\nL 532 400\nL 532 400\nA 4 4 90.00 0 1 536 396\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"536\" y=\"413\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"542\" y=\"426\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text><path  d=\"M 639 146\nL 723 146\nL 723 146\nA 4 4 90.00 0 1 727 150\nL 727 163\nL 727 163\nA 4 4 90.00 0 1 723 167\nL 639 167\nL 639 167\nA 4 4 90.00 0 1 635 163\nL 635 150\nL 635 150\nA 4 4 90.00 0 1 639 146\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"639\" y=\"163\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">* Morning Star</text></svg>",
			pngCRC: 0xa0b878a7,
		},
		{
			name: "combination_mixed",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},     // Normal candle
					{Open: 105, High: 108, Low: 102, Close: 105.05}, // Doji pattern
					{Open: 105, High: 107, Low: 95, Close: 106},     // Hammer pattern
					{Open: 110, High: 125, Low: 95, Close: 112},     // Spinning top pattern
					{Open: 100, High: 120, Low: 100, Close: 120},    // Bullish marubozu pattern
					{Open: 120, High: 120, Low: 100, Close: 100},    // Bearish marubozu pattern
					{Open: 110, High: 112, Low: 105, Close: 106},    // Small bearish candle
					{Open: 104, High: 115, Low: 103, Close: 114},    // Bullish engulfing
					{Open: 106, High: 125, Low: 105, Close: 107},    // Shooting star pattern
					{Open: 109, High: 110, Low: 90, Close: 108.9},   // Dragonfly doji pattern
					{Open: 108, High: 120, Low: 107, Close: 108.1},  // Gravestone doji pattern
					{Open: 108, High: 115, Low: 103, Close: 110},    // Normal candle
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold:    0.01,
					ShadowRatio:      2.0,
					EngulfingMinSize: 0.8,
				})
				series.PatternConfig.EnabledPatterns = bulk.SliceFilterInPlace(func(pattern string) bool {
					// remove a few high volume patterns
					if pattern == CandlestickPatternDoji {
						return false
					} else if pattern == CandlestickPatternSpinningTop {
						return false
					} else if pattern == CandlestickPatternHighWave {
						return false
					}
					return true
				}, series.PatternConfig.EnabledPatterns)
				return CandlestickChartOption{
					XAxis:      XAxisOption{Show: Ptr(false)},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">126.75</text><text x=\"9\" y=\"80\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">122.67</text><text x=\"9\" y=\"144\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">118.58</text><text x=\"17\" y=\"208\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">114.5</text><text x=\"9\" y=\"272\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110.42</text><text x=\"9\" y=\"337\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">106.33</text><text x=\"9\" y=\"401\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">102.25</text><text x=\"17\" y=\"465\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">98.17</text><text x=\"17\" y=\"529\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94.08</text><text x=\"39\" y=\"594\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 74\nL 790 74\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 138\nL 790 138\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 203\nL 790 203\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 267\nL 790 267\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 332\nL 790 332\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 396\nL 790 396\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 461\nL 790 461\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 525\nL 790 525\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 97 275\nL 97 354\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 97 433\nL 97 512\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 85 275\nL 109 275\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 85 512\nL 109 512\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 73 354\nL 121 354\nL 121 433\nL 73 433\nL 73 354\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 157 306\nL 157 353\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 157 354\nL 157 401\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 145 306\nL 169 306\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 145 401\nL 169 401\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 133 353\nL 181 353\nL 181 354\nL 133 354\nL 133 353\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 217 322\nL 217 338\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 217 354\nL 217 512\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 205 322\nL 229 322\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 205 512\nL 229 512\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 193 338\nL 241 338\nL 241 354\nL 193 354\nL 193 338\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 277 38\nL 277 243\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 277 275\nL 277 512\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 265 38\nL 289 38\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 265 512\nL 289 512\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 253 243\nL 301 243\nL 301 275\nL 253 275\nL 253 243\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 326 117\nL 350 117\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 326 433\nL 350 433\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 314 117\nL 362 117\nL 362 433\nL 314 433\nL 314 117\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 386 117\nL 410 117\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 386 433\nL 410 433\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 374 117\nL 422 117\nL 422 433\nL 374 433\nL 374 117\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 458 243\nL 458 275\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 458 338\nL 458 354\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 446 243\nL 470 243\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 446 354\nL 470 354\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 434 275\nL 482 275\nL 482 338\nL 434 338\nL 434 275\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 518 196\nL 518 212\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 518 370\nL 518 385\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 506 196\nL 530 196\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 506 385\nL 530 385\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 494 212\nL 542 212\nL 542 370\nL 494 370\nL 494 212\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 579 38\nL 579 322\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 579 338\nL 579 354\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 567 38\nL 591 38\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 567 354\nL 591 354\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 555 322\nL 603 322\nL 603 338\nL 555 338\nL 555 322\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 639 275\nL 639 291\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 639 292\nL 639 590\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 627 275\nL 651 275\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 627 590\nL 651 590\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 615 291\nL 663 291\nL 663 292\nL 615 292\nL 615 291\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 699 117\nL 699 305\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 699 306\nL 699 322\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 687 117\nL 711 117\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 687 322\nL 711 322\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 675 305\nL 723 305\nL 723 306\nL 675 306\nL 675 305\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 759 196\nL 759 275\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 759 306\nL 759 385\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 747 196\nL 771 196\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 747 385\nL 771 385\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 735 275\nL 783 275\nL 783 306\nL 735 306\nL 735 275\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 162 340\nL 272 340\nL 272 340\nA 4 4 90.00 0 1 276 344\nL 276 357\nL 276 357\nA 4 4 90.00 0 1 272 361\nL 162 361\nL 162 361\nA 4 4 90.00 0 1 158 357\nL 158 344\nL 158 344\nA 4 4 90.00 0 1 162 340\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"162\" y=\"357\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">‡ Long Legged Doji</text><path  d=\"M 222 325\nL 282 325\nL 282 325\nA 4 4 90.00 0 1 286 329\nL 286 342\nL 286 342\nA 4 4 90.00 0 1 282 346\nL 222 346\nL 222 346\nA 4 4 90.00 0 1 218 342\nL 218 329\nL 218 329\nA 4 4 90.00 0 1 222 325\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"222\" y=\"342\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><path  d=\"M 343 98\nL 435 98\nL 435 98\nA 4 4 90.00 0 1 439 102\nL 439 128\nL 439 128\nA 4 4 90.00 0 1 435 132\nL 343 132\nL 343 132\nA 4 4 90.00 0 1 339 128\nL 339 102\nL 339 102\nA 4 4 90.00 0 1 343 98\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"343\" y=\"115\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">^ Bull Marubozu</text><text x=\"347\" y=\"128\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">[ Bull Belt Hold</text><path  d=\"M 403 414\nL 500 414\nL 500 414\nA 4 4 90.00 0 1 504 418\nL 504 444\nL 504 444\nA 4 4 90.00 0 1 500 448\nL 403 448\nL 403 448\nA 4 4 90.00 0 1 399 444\nL 399 418\nL 399 418\nA 4 4 90.00 0 1 403 414\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"403\" y=\"431\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">v Bear Marubozu</text><text x=\"407\" y=\"444\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">] Bear Belt Hold</text><path  d=\"M 523 199\nL 614 199\nL 614 199\nA 4 4 90.00 0 1 618 203\nL 618 216\nL 618 216\nA 4 4 90.00 0 1 614 220\nL 523 220\nL 523 220\nA 4 4 90.00 0 1 519 216\nL 519 203\nL 519 203\nA 4 4 90.00 0 1 523 199\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"523\" y=\"216\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Λ Bull Engulfing</text><path  d=\"M 584 303\nL 676 303\nL 676 303\nA 4 4 90.00 0 1 680 307\nL 680 333\nL 680 333\nA 4 4 90.00 0 1 676 337\nL 584 337\nL 584 337\nA 4 4 90.00 0 1 580 333\nL 580 307\nL 580 307\nA 4 4 90.00 0 1 584 303\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"584\" y=\"320\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"588\" y=\"333\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text><path  d=\"M 644 273\nL 712 273\nL 712 273\nA 4 4 90.00 0 1 716 277\nL 716 303\nL 716 303\nA 4 4 90.00 0 1 712 307\nL 644 307\nL 644 307\nA 4 4 90.00 0 1 640 303\nL 640 277\nL 640 277\nA 4 4 90.00 0 1 644 273\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"648\" y=\"290\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><text x=\"644\" y=\"303\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ψ Dragonfly</text><path  d=\"M 704 279\nL 796 279\nL 796 279\nA 4 4 90.00 0 1 800 283\nL 800 322\nL 800 322\nA 4 4 90.00 0 1 796 326\nL 704 326\nL 704 326\nA 4 4 90.00 0 1 700 322\nL 700 283\nL 700 283\nA 4 4 90.00 0 1 704 279\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"704\" y=\"296\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"712\" y=\"309\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">† Gravestone</text><text x=\"708\" y=\"322\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text></svg>",
			pngCRC: 0x53654a72,
		},
		{
			name: "combination_three_candle_patterns",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105}, // Normal candle
					// Morning star sequence
					{Open: 120, High: 125, Low: 105, Close: 108}, // Large bearish
					{Open: 102, High: 104, Low: 100, Close: 103}, // Small body, gap down
					{Open: 108, High: 125, Low: 106, Close: 122}, // Large bullish, gap up
					// Three white soldiers sequence
					{Open: 110, High: 115, Low: 109, Close: 114}, // First soldier
					{Open: 113, High: 118, Low: 112, Close: 117}, // Second soldier
					{Open: 116, High: 121, Low: 115, Close: 120}, // Third soldier
					// Evening star sequence
					{Open: 122, High: 140, Low: 120, Close: 138}, // Large bullish
					{Open: 142, High: 144, Low: 140, Close: 143}, // Small body, gap up
					{Open: 138, High: 140, Low: 115, Close: 118}, // Large bearish, gap down
					// Three black crows sequence
					{Open: 120, High: 121, Low: 115, Close: 116}, // Second crow
					{Open: 117, High: 118, Low: 112, Close: 113}, // Third crow
					{Open: 113, High: 132, Low: 106, Close: 128}, // Normal candle
				}
				series := CandlestickSeries{
					Data: data,
					PatternConfig: &CandlestickPatternConfig{
						ReplaceSeriesLabel: true,
						EnabledPatterns: []string{
							CandlestickPatternMorningStar,
							CandlestickPatternEveningStar,
						},
						DojiThreshold: 0.01,
						ShadowRatio:   2.0,
					},
				}
				return CandlestickChartOption{
					XAxis:      XAxisOption{Show: Ptr(false)},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">149</text><text x=\"9\" y=\"80\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">142.44</text><text x=\"9\" y=\"144\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">135.89</text><text x=\"9\" y=\"208\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">129.33</text><text x=\"9\" y=\"272\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">122.78</text><text x=\"9\" y=\"337\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">116.22</text><text x=\"9\" y=\"401\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">109.67</text><text x=\"9\" y=\"465\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">103.11</text><text x=\"17\" y=\"529\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">96.56</text><text x=\"39\" y=\"594\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 74\nL 790 74\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 138\nL 790 138\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 203\nL 790 203\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 267\nL 790 267\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 332\nL 790 332\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 396\nL 790 396\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 461\nL 790 461\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 525\nL 790 525\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 94 394\nL 94 443\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 94 492\nL 94 541\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 83 394\nL 105 394\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 83 541\nL 105 541\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 72 443\nL 116 443\nL 116 492\nL 72 492\nL 72 443\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 150 246\nL 150 296\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 150 414\nL 150 443\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 139 246\nL 161 246\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 139 443\nL 161 443\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 128 296\nL 172 296\nL 172 414\nL 128 414\nL 128 296\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 205 453\nL 205 463\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 205 473\nL 205 492\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 194 453\nL 216 453\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 194 492\nL 216 492\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 183 463\nL 227 463\nL 227 473\nL 183 473\nL 183 463\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 261 246\nL 261 276\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 261 414\nL 261 433\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 250 246\nL 272 246\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 250 433\nL 272 433\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 239 276\nL 283 276\nL 283 414\nL 239 414\nL 239 276\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 317 345\nL 317 355\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 317 394\nL 317 404\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 306 345\nL 328 345\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 306 404\nL 328 404\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 295 355\nL 339 355\nL 339 394\nL 295 394\nL 295 355\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 372 315\nL 372 325\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 372 364\nL 372 374\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 361 315\nL 383 315\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 361 374\nL 383 374\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 350 325\nL 394 325\nL 394 364\nL 350 364\nL 350 325\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 428 286\nL 428 296\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 428 335\nL 428 345\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 417 286\nL 439 286\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 417 345\nL 439 345\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 406 296\nL 450 296\nL 450 335\nL 406 335\nL 406 296\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 483 99\nL 483 119\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 483 276\nL 483 296\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 472 99\nL 494 99\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 472 296\nL 494 296\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 461 119\nL 505 119\nL 505 276\nL 461 276\nL 461 119\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 539 60\nL 539 69\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 539 79\nL 539 99\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 528 60\nL 550 60\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 528 99\nL 550 99\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 517 69\nL 561 69\nL 561 79\nL 517 79\nL 517 69\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 595 99\nL 595 119\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 595 315\nL 595 345\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 584 99\nL 606 99\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 584 345\nL 606 345\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 573 119\nL 617 119\nL 617 315\nL 573 315\nL 573 119\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 650 286\nL 650 296\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 650 335\nL 650 345\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 639 286\nL 661 286\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 639 345\nL 661 345\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 628 296\nL 672 296\nL 672 335\nL 628 335\nL 628 296\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 706 315\nL 706 325\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 706 364\nL 706 374\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 695 315\nL 717 315\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 695 374\nL 717 374\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 684 325\nL 728 325\nL 728 364\nL 684 364\nL 684 325\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 762 178\nL 762 217\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 762 364\nL 762 433\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 751 178\nL 773 178\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 751 433\nL 773 433\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 740 217\nL 784 217\nL 784 364\nL 740 364\nL 740 217\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 266 263\nL 350 263\nL 350 263\nA 4 4 90.00 0 1 354 267\nL 354 280\nL 354 280\nA 4 4 90.00 0 1 350 284\nL 266 284\nL 266 284\nA 4 4 90.00 0 1 262 280\nL 262 267\nL 262 267\nA 4 4 90.00 0 1 266 263\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"266\" y=\"280\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">* Morning Star</text><path  d=\"M 600 302\nL 682 302\nL 682 302\nA 4 4 90.00 0 1 686 306\nL 686 319\nL 686 319\nA 4 4 90.00 0 1 682 323\nL 600 323\nL 600 323\nA 4 4 90.00 0 1 596 319\nL 596 306\nL 596 306\nA 4 4 90.00 0 1 600 302\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"600\" y=\"319\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">⁎ Evening Star</text></svg>",
			pngCRC: 0x560307c7,
		},
		{
			name: "bullish_patterns",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 110, High: 112, Low: 105, Close: 106}, // Small bearish candle
					{Open: 104, High: 115, Low: 103, Close: 114}, // Bullish engulfing
					{Open: 108, High: 109, Low: 98, Close: 107},  // Hammer pattern
					{Open: 100, High: 120, Low: 100, Close: 120}, // Bullish belt hold / marubozu
					{Open: 120, High: 140, Low: 118, Close: 138}, // Large bullish
					{Open: 110, High: 119, Low: 110, Close: 118}, // Piercing line
					{Open: 118, High: 125, Low: 115, Close: 122}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold:    0.01,
					ShadowRatio:      2.0,
					EngulfingMinSize: 0.8,
				})
				series.PatternConfig = PatternsBullish()
				return CandlestickChartOption{
					XAxis:      XAxisOption{Show: Ptr(false)},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">144</text><text x=\"9\" y=\"80\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">138</text><text x=\"9\" y=\"144\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">132</text><text x=\"9\" y=\"208\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">126</text><text x=\"9\" y=\"272\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"9\" y=\"337\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">114</text><text x=\"9\" y=\"401\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">108</text><text x=\"9\" y=\"465\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">102</text><text x=\"18\" y=\"529\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">96</text><text x=\"18\" y=\"594\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 74\nL 790 74\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 138\nL 790 138\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 203\nL 790 203\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 267\nL 790 267\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 332\nL 790 332\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 396\nL 790 396\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 461\nL 790 461\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 525\nL 790 525\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 92 376\nL 92 429\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 92 483\nL 92 537\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 74 376\nL 110 376\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 74 537\nL 110 537\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 55 429\nL 129 429\nL 129 483\nL 55 483\nL 55 429\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 185 354\nL 185 376\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 185 419\nL 185 429\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 167 354\nL 203 354\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 167 429\nL 203 429\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 148 376\nL 222 376\nL 222 419\nL 148 419\nL 148 376\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 278 322\nL 278 333\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 278 440\nL 278 451\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 260 322\nL 296 322\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 260 451\nL 296 451\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 241 333\nL 315 333\nL 315 440\nL 241 440\nL 241 333\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 371 386\nL 371 397\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 371 408\nL 371 505\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 353 386\nL 389 386\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 353 505\nL 389 505\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 334 397\nL 408 397\nL 408 408\nL 334 408\nL 334 397\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 446 268\nL 482 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 446 483\nL 482 483\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 427 268\nL 501 268\nL 501 483\nL 427 483\nL 427 268\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 557 53\nL 557 75\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 557 268\nL 557 290\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 539 53\nL 575 53\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 539 290\nL 575 290\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 520 75\nL 594 75\nL 594 268\nL 520 268\nL 520 75\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 650 279\nL 650 290\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 632 279\nL 668 279\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 632 376\nL 668 376\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 613 290\nL 687 290\nL 687 376\nL 613 376\nL 613 290\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 743 215\nL 743 247\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 743 290\nL 743 322\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 725 215\nL 761 215\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 725 322\nL 761 322\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 706 247\nL 780 247\nL 780 290\nL 706 290\nL 706 247\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 283 320\nL 374 320\nL 374 320\nA 4 4 90.00 0 1 378 324\nL 378 337\nL 378 337\nA 4 4 90.00 0 1 374 341\nL 283 341\nL 283 341\nA 4 4 90.00 0 1 279 337\nL 279 324\nL 279 324\nA 4 4 90.00 0 1 283 320\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"283\" y=\"337\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Λ Bull Engulfing</text><path  d=\"M 376 395\nL 436 395\nL 436 395\nA 4 4 90.00 0 1 440 399\nL 440 412\nL 440 412\nA 4 4 90.00 0 1 436 416\nL 376 416\nL 376 416\nA 4 4 90.00 0 1 372 412\nL 372 399\nL 372 399\nA 4 4 90.00 0 1 376 395\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"376\" y=\"412\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><path  d=\"M 469 242\nL 561 242\nL 561 242\nA 4 4 90.00 0 1 565 246\nL 565 285\nL 565 285\nA 4 4 90.00 0 1 561 289\nL 469 289\nL 469 289\nA 4 4 90.00 0 1 465 285\nL 465 246\nL 465 246\nA 4 4 90.00 0 1 469 242\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"469\" y=\"259\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">^ Bull Marubozu</text><text x=\"469\" y=\"272\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Λ Bull Engulfing</text><text x=\"473\" y=\"285\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">[ Bull Belt Hold</text><path  d=\"M 655 277\nL 739 277\nL 739 277\nA 4 4 90.00 0 1 743 281\nL 743 294\nL 743 294\nA 4 4 90.00 0 1 739 298\nL 655 298\nL 655 298\nA 4 4 90.00 0 1 651 294\nL 651 281\nL 651 281\nA 4 4 90.00 0 1 655 277\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"655\" y=\"294\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">[ Bull Belt Hold</text></svg>",
			pngCRC: 0x314bb52e,
		},
		{
			name: "bearish_patterns",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 106, High: 112, Low: 105, Close: 110}, // Small bullish candle
					{Open: 114, High: 115, Low: 103, Close: 104}, // Bearish engulfing
					{Open: 106, High: 125, Low: 105, Close: 107}, // Shooting star pattern
					{Open: 120, High: 120, Low: 100, Close: 100}, // Bearish belt hold / marubozu
					{Open: 118, High: 125, Low: 117, Close: 125}, // Bullish candle
					{Open: 127, High: 128, Low: 120, Close: 121}, // Dark cloud cover
					{Open: 121, High: 124, Low: 118, Close: 120}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold:    0.01,
					ShadowRatio:      2.0,
					EngulfingMinSize: 0.8,
				})
				series.PatternConfig = PatternsBearish()
				return CandlestickChartOption{
					XAxis:      XAxisOption{Show: Ptr(false)},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">133</text><text x=\"9\" y=\"80\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">128.22</text><text x=\"9\" y=\"144\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">123.44</text><text x=\"9\" y=\"208\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">118.67</text><text x=\"9\" y=\"272\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">113.89</text><text x=\"9\" y=\"337\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">109.11</text><text x=\"9\" y=\"401\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">104.33</text><text x=\"17\" y=\"465\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">99.56</text><text x=\"17\" y=\"529\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94.78</text><text x=\"39\" y=\"594\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 74\nL 790 74\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 138\nL 790 138\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 203\nL 790 203\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 267\nL 790 267\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 332\nL 790 332\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 396\nL 790 396\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 461\nL 790 461\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 525\nL 790 525\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 112 321\nL 112 388\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 112 456\nL 112 523\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 94 321\nL 130 321\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 94 523\nL 130 523\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 76 388\nL 148 388\nL 148 456\nL 76 456\nL 76 388\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 202 294\nL 202 321\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 202 375\nL 202 388\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 184 294\nL 220 294\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 184 388\nL 220 388\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 166 321\nL 238 321\nL 238 375\nL 166 375\nL 166 321\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 292 253\nL 292 267\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 292 402\nL 292 415\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 274 253\nL 310 253\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 274 415\nL 310 415\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 256 267\nL 328 267\nL 328 402\nL 256 402\nL 256 267\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 383 118\nL 383 361\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 383 375\nL 383 388\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 365 118\nL 401 118\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 365 388\nL 401 388\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 347 361\nL 419 361\nL 419 375\nL 347 375\nL 347 361\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 455 186\nL 491 186\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 455 456\nL 491 456\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 437 186\nL 509 186\nL 509 456\nL 437 456\nL 437 186\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 563 213\nL 563 226\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 545 118\nL 581 118\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 545 226\nL 581 226\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 527 118\nL 599 118\nL 599 213\nL 527 213\nL 527 118\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 654 78\nL 654 91\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 654 172\nL 654 186\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 636 78\nL 672 78\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 636 186\nL 672 186\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 618 91\nL 690 91\nL 690 172\nL 618 172\nL 618 91\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 744 132\nL 744 172\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 744 186\nL 744 213\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 726 132\nL 762 132\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 726 213\nL 762 213\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 708 172\nL 780 172\nL 780 186\nL 708 186\nL 708 172\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 297 389\nL 392 389\nL 392 389\nA 4 4 90.00 0 1 396 393\nL 396 406\nL 396 406\nA 4 4 90.00 0 1 392 410\nL 297 410\nL 297 410\nA 4 4 90.00 0 1 293 406\nL 293 393\nL 293 393\nA 4 4 90.00 0 1 297 389\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"297\" y=\"406\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">V Bear Engulfing</text><path  d=\"M 388 348\nL 480 348\nL 480 348\nA 4 4 90.00 0 1 484 352\nL 484 365\nL 484 365\nA 4 4 90.00 0 1 480 369\nL 388 369\nL 388 369\nA 4 4 90.00 0 1 384 365\nL 384 352\nL 384 352\nA 4 4 90.00 0 1 388 348\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"388\" y=\"365\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><path  d=\"M 478 430\nL 575 430\nL 575 430\nA 4 4 90.00 0 1 579 434\nL 579 473\nL 579 473\nA 4 4 90.00 0 1 575 477\nL 478 477\nL 478 477\nA 4 4 90.00 0 1 474 473\nL 474 434\nL 474 434\nA 4 4 90.00 0 1 478 430\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"478\" y=\"447\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">v Bear Marubozu</text><text x=\"479\" y=\"460\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">V Bear Engulfing</text><text x=\"482\" y=\"473\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">] Bear Belt Hold</text><path  d=\"M 659 159\nL 733 159\nL 733 159\nA 4 4 90.00 0 1 737 163\nL 737 176\nL 737 176\nA 4 4 90.00 0 1 733 180\nL 659 180\nL 659 180\nA 4 4 90.00 0 1 655 176\nL 655 163\nL 655 163\nA 4 4 90.00 0 1 659 159\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"659\" y=\"176\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ξ Dark Cloud</text></svg>",
			pngCRC: 0x16eb782a,
		},
		{
			name: "reversal_patterns",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 120, High: 121, Low: 115, Close: 115}, // Bearish candle
					{Open: 112, High: 119, Low: 112, Close: 118}, // Piercing line (bullish reversal)
					{Open: 118, High: 125, Low: 118, Close: 125}, // Bullish candle
					{Open: 127, High: 127, Low: 120, Close: 121}, // Dark cloud cover (bearish reversal)
					{Open: 125, High: 126, Low: 100, Close: 102}, // Bearish with low at 100
					{Open: 102, High: 108, Low: 100, Close: 107}, // Tweezer bottom (bullish reversal)
					{Open: 107, High: 112, Low: 102, Close: 110}, // Normal candle
					// Additional reversal patterns
					{Open: 115, High: 117, Low: 95, Close: 114},    // Hammer pattern (bullish reversal)
					{Open: 112, High: 130, Low: 111, Close: 113},   // Shooting star pattern (bearish reversal)
					{Open: 108, High: 110, Low: 85, Close: 108.1},  // Dragonfly doji (bullish reversal)
					{Open: 105, High: 125, Low: 104, Close: 105.1}, // Gravestone doji (bearish reversal)
					{Open: 130, High: 135, Low: 110, Close: 115},   // Large bearish for engulfing setup
					{Open: 110, High: 140, Low: 108, Close: 138},   // Bullish engulfing (reversal)
					{Open: 140, High: 145, Low: 105, Close: 110},   // Bearish engulfing (reversal)
					// Three candle reversal patterns
					{Open: 125, High: 130, Low: 105, Close: 110}, // Large bearish for morning star
					{Open: 105, High: 108, Low: 102, Close: 106}, // Small body (morning star middle)
					{Open: 110, High: 135, Low: 108, Close: 130}, // Large bullish (morning star completion)
					{Open: 115, High: 140, Low: 113, Close: 135}, // Large bullish for evening star
					{Open: 138, High: 145, Low: 136, Close: 140}, // Small body (evening star middle)
					{Open: 135, High: 136, Low: 110, Close: 115}, // Large bearish (evening star completion)
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
				series.PatternConfig = PatternsReversal()
				return CandlestickChartOption{
					XAxis:      XAxisOption{Show: Ptr(false)},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">152</text><text x=\"9\" y=\"80\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">144</text><text x=\"9\" y=\"144\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">136</text><text x=\"9\" y=\"208\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">128</text><text x=\"9\" y=\"272\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"9\" y=\"337\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">112</text><text x=\"9\" y=\"401\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">104</text><text x=\"18\" y=\"465\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">96</text><text x=\"18\" y=\"529\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">88</text><text x=\"18\" y=\"594\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">80</text><path  d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 74\nL 790 74\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 138\nL 790 138\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 203\nL 790 203\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 267\nL 790 267\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 332\nL 790 332\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 396\nL 790 396\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 461\nL 790 461\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 525\nL 790 525\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 349\nL 63 389\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 63 429\nL 63 470\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 56 349\nL 70 349\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 56 470\nL 70 470\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 49 389\nL 77 389\nL 77 429\nL 49 429\nL 49 389\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 98 260\nL 98 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 91 260\nL 105 260\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 91 309\nL 105 309\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 84 268\nL 112 268\nL 112 309\nL 84 309\nL 84 268\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 134 276\nL 134 284\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 127 276\nL 141 276\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 127 333\nL 141 333\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 120 284\nL 148 284\nL 148 333\nL 120 333\nL 120 284\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 162 228\nL 176 228\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 162 284\nL 176 284\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 155 228\nL 183 228\nL 183 284\nL 155 284\nL 155 228\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 205 260\nL 205 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 198 212\nL 212 212\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 198 268\nL 212 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 191 212\nL 219 212\nL 219 260\nL 191 260\nL 191 212\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 240 220\nL 240 228\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 240 413\nL 240 429\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 233 220\nL 247 220\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 233 429\nL 247 429\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 226 228\nL 254 228\nL 254 413\nL 226 413\nL 226 228\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 276 365\nL 276 373\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 276 413\nL 276 429\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 269 365\nL 283 365\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 269 429\nL 283 429\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 262 373\nL 290 373\nL 290 413\nL 262 413\nL 262 373\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 311 333\nL 311 349\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 311 373\nL 311 413\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 304 333\nL 318 333\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 304 413\nL 318 413\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 297 349\nL 325 349\nL 325 373\nL 297 373\nL 297 349\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 346 292\nL 346 309\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 346 317\nL 346 470\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 339 292\nL 353 292\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 339 470\nL 353 470\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 332 309\nL 360 309\nL 360 317\nL 332 317\nL 332 309\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 382 188\nL 382 325\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 382 333\nL 382 341\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 375 188\nL 389 188\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 375 341\nL 389 341\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 368 325\nL 396 325\nL 396 333\nL 368 333\nL 368 325\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 417 349\nL 417 364\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 417 365\nL 417 550\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 410 349\nL 424 349\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 410 550\nL 424 550\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 403 364\nL 431 364\nL 431 365\nL 403 365\nL 403 364\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 453 228\nL 453 388\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 453 389\nL 453 397\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 446 228\nL 460 228\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 446 397\nL 460 397\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 439 388\nL 467 388\nL 467 389\nL 439 389\nL 439 388\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 488 147\nL 488 188\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 488 309\nL 488 349\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 481 147\nL 495 147\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 481 349\nL 495 349\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 474 188\nL 502 188\nL 502 309\nL 474 309\nL 474 188\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 524 107\nL 524 123\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 524 349\nL 524 365\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 517 107\nL 531 107\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 517 365\nL 531 365\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 510 123\nL 538 123\nL 538 349\nL 510 349\nL 510 123\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 559 67\nL 559 107\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 559 349\nL 559 389\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 552 67\nL 566 67\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 552 389\nL 566 389\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 545 107\nL 573 107\nL 573 349\nL 545 349\nL 545 107\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 594 188\nL 594 228\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 594 349\nL 594 389\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 587 188\nL 601 188\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 587 389\nL 601 389\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 580 228\nL 608 228\nL 608 349\nL 580 349\nL 580 228\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 630 365\nL 630 381\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 630 389\nL 630 413\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 623 365\nL 637 365\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 623 413\nL 637 413\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 616 381\nL 644 381\nL 644 389\nL 616 389\nL 616 381\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 665 147\nL 665 188\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 665 349\nL 665 365\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 658 147\nL 672 147\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 658 365\nL 672 365\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 651 188\nL 679 188\nL 679 349\nL 651 349\nL 651 188\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 701 107\nL 701 147\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 701 309\nL 701 325\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 694 107\nL 708 107\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 694 325\nL 708 325\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 687 147\nL 715 147\nL 715 309\nL 687 309\nL 687 147\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 736 67\nL 736 107\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 736 123\nL 736 139\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 729 67\nL 743 67\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 729 139\nL 743 139\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 722 107\nL 750 107\nL 750 123\nL 722 123\nL 722 107\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 772 139\nL 772 147\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 772 309\nL 772 349\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 765 139\nL 779 139\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 765 349\nL 779 349\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 758 147\nL 786 147\nL 786 309\nL 758 309\nL 758 147\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 139 271\nL 220 271\nL 220 271\nA 4 4 90.00 0 1 224 275\nL 224 288\nL 224 288\nA 4 4 90.00 0 1 220 292\nL 139 292\nL 139 292\nA 4 4 90.00 0 1 135 288\nL 135 275\nL 135 275\nA 4 4 90.00 0 1 139 271\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"139\" y=\"288\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">| Piercing Line</text><path  d=\"M 210 247\nL 284 247\nL 284 247\nA 4 4 90.00 0 1 288 251\nL 288 264\nL 288 264\nA 4 4 90.00 0 1 284 268\nL 210 268\nL 210 268\nA 4 4 90.00 0 1 206 264\nL 206 251\nL 206 251\nA 4 4 90.00 0 1 210 247\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"210\" y=\"264\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ξ Dark Cloud</text><path  d=\"M 351 304\nL 411 304\nL 411 304\nA 4 4 90.00 0 1 415 308\nL 415 321\nL 415 321\nA 4 4 90.00 0 1 411 325\nL 351 325\nL 351 325\nA 4 4 90.00 0 1 347 321\nL 347 308\nL 347 308\nA 4 4 90.00 0 1 351 304\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"351\" y=\"321\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><path  d=\"M 387 312\nL 479 312\nL 479 312\nA 4 4 90.00 0 1 483 316\nL 483 329\nL 483 329\nA 4 4 90.00 0 1 479 333\nL 387 333\nL 387 333\nA 4 4 90.00 0 1 383 329\nL 383 316\nL 383 316\nA 4 4 90.00 0 1 387 312\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"387\" y=\"329\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><path  d=\"M 422 351\nL 482 351\nL 482 351\nA 4 4 90.00 0 1 486 355\nL 486 368\nL 486 368\nA 4 4 90.00 0 1 482 372\nL 422 372\nL 422 372\nA 4 4 90.00 0 1 418 368\nL 418 355\nL 418 355\nA 4 4 90.00 0 1 422 351\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"422\" y=\"368\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><path  d=\"M 458 375\nL 550 375\nL 550 375\nA 4 4 90.00 0 1 554 379\nL 554 392\nL 554 392\nA 4 4 90.00 0 1 550 396\nL 458 396\nL 458 396\nA 4 4 90.00 0 1 454 392\nL 454 379\nL 454 379\nA 4 4 90.00 0 1 458 375\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"458\" y=\"392\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><path  d=\"M 529 110\nL 620 110\nL 620 110\nA 4 4 90.00 0 1 624 114\nL 624 127\nL 624 127\nA 4 4 90.00 0 1 620 131\nL 529 131\nL 529 131\nA 4 4 90.00 0 1 525 127\nL 525 114\nL 525 114\nA 4 4 90.00 0 1 529 110\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"529\" y=\"127\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Λ Bull Engulfing</text><path  d=\"M 670 175\nL 754 175\nL 754 175\nA 4 4 90.00 0 1 758 179\nL 758 192\nL 758 192\nA 4 4 90.00 0 1 754 196\nL 670 196\nL 670 196\nA 4 4 90.00 0 1 666 192\nL 666 179\nL 666 179\nA 4 4 90.00 0 1 670 175\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"670\" y=\"192\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">* Morning Star</text><path  d=\"M 718 296\nL 800 296\nL 800 296\nA 4 4 90.00 0 1 804 300\nL 804 313\nL 804 313\nA 4 4 90.00 0 1 800 317\nL 718 317\nL 718 317\nA 4 4 90.00 0 1 714 313\nL 714 300\nL 714 300\nA 4 4 90.00 0 1 718 296\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"718\" y=\"313\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">⁎ Evening Star</text></svg>",
			pngCRC: 0x54baf0d0,
		},
		{
			name: "indecision_patterns",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},    // Normal candle
					{Open: 110, High: 111, Low: 109, Close: 110.1}, // Doji pattern - very small body
					{Open: 105, High: 108, Low: 102, Close: 106},   // Spinning top - small body with shadows
					{Open: 108, High: 118, Low: 98, Close: 107},    // High wave - very long shadows, small body
					{Open: 120, High: 130, Low: 115, Close: 118},   // Large bearish candle for harami setup
					{Open: 119, High: 122, Low: 117, Close: 120},   // Harami pattern - small body inside previous
					{Open: 115, High: 116, Low: 114, Close: 115.1}, // Another doji
					{Open: 112, High: 125, Low: 100, Close: 113},   // Long-legged doji setup
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold:    0.01,
					ShadowRatio:      2.0,
					EngulfingMinSize: 0.8,
				})
				series.PatternConfig = PatternsIndecision()
				return CandlestickChartOption{
					XAxis:      XAxisOption{Show: Ptr(false)},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">135</text><text x=\"9\" y=\"80\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"144\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125</text><text x=\"9\" y=\"208\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"9\" y=\"272\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">115</text><text x=\"9\" y=\"337\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"9\" y=\"401\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105</text><text x=\"9\" y=\"465\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100</text><text x=\"18\" y=\"529\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><text x=\"18\" y=\"594\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 74\nL 790 74\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 138\nL 790 138\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 203\nL 790 203\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 267\nL 790 267\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 332\nL 790 332\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 396\nL 790 396\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 461\nL 790 461\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 525\nL 790 525\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 92 333\nL 92 397\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 92 462\nL 92 526\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 74 333\nL 110 333\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 74 526\nL 110 526\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 55 397\nL 129 397\nL 129 462\nL 55 462\nL 55 397\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 185 320\nL 185 331\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 185 333\nL 185 346\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 167 320\nL 203 320\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 167 346\nL 203 346\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 148 331\nL 222 331\nL 222 333\nL 148 333\nL 148 331\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 278 358\nL 278 384\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 278 397\nL 278 436\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 260 358\nL 296 358\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 260 436\nL 296 436\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 241 384\nL 315 384\nL 315 397\nL 241 397\nL 241 384\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 371 230\nL 371 358\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 371 371\nL 371 487\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 353 230\nL 389 230\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 353 487\nL 389 487\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 334 358\nL 408 358\nL 408 371\nL 334 371\nL 334 358\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 464 75\nL 464 204\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 464 230\nL 464 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 446 75\nL 482 75\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 446 268\nL 482 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 427 204\nL 501 204\nL 501 230\nL 427 230\nL 427 204\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 557 178\nL 557 204\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 557 217\nL 557 242\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 539 178\nL 575 178\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 539 242\nL 575 242\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 520 204\nL 594 204\nL 594 217\nL 520 217\nL 520 204\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 650 255\nL 650 267\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 650 268\nL 650 281\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 632 255\nL 668 255\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 632 281\nL 668 281\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 613 267\nL 687 267\nL 687 268\nL 613 268\nL 613 267\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 743 139\nL 743 294\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 743 307\nL 743 462\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 725 139\nL 761 139\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 725 462\nL 761 462\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 706 294\nL 780 294\nL 780 307\nL 706 307\nL 706 294\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 190 312\nL 276 312\nL 276 312\nA 4 4 90.00 0 1 280 316\nL 280 342\nL 280 342\nA 4 4 90.00 0 1 276 346\nL 190 346\nL 190 346\nA 4 4 90.00 0 1 186 342\nL 186 316\nL 186 316\nA 4 4 90.00 0 1 190 312\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"196\" y=\"329\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text><text x=\"190\" y=\"342\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 283 365\nL 369 365\nL 369 365\nA 4 4 90.00 0 1 373 369\nL 373 395\nL 373 395\nA 4 4 90.00 0 1 369 399\nL 283 399\nL 283 399\nA 4 4 90.00 0 1 279 395\nL 279 369\nL 279 369\nA 4 4 90.00 0 1 283 365\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"289\" y=\"382\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text><text x=\"283\" y=\"395\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 376 352\nL 462 352\nL 462 352\nA 4 4 90.00 0 1 466 356\nL 466 382\nL 466 382\nA 4 4 90.00 0 1 462 386\nL 376 386\nL 376 386\nA 4 4 90.00 0 1 372 382\nL 372 356\nL 372 356\nA 4 4 90.00 0 1 376 352\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"382\" y=\"369\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text><text x=\"376\" y=\"382\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 469 211\nL 555 211\nL 555 211\nA 4 4 90.00 0 1 559 215\nL 559 241\nL 559 241\nA 4 4 90.00 0 1 555 245\nL 469 245\nL 469 245\nA 4 4 90.00 0 1 465 241\nL 465 215\nL 465 215\nA 4 4 90.00 0 1 469 211\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"475\" y=\"228\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text><text x=\"469\" y=\"241\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 562 185\nL 648 185\nL 648 185\nA 4 4 90.00 0 1 652 189\nL 652 215\nL 652 215\nA 4 4 90.00 0 1 648 219\nL 562 219\nL 562 219\nA 4 4 90.00 0 1 558 215\nL 558 189\nL 558 189\nA 4 4 90.00 0 1 562 185\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"568\" y=\"202\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text><text x=\"562\" y=\"215\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 655 248\nL 741 248\nL 741 248\nA 4 4 90.00 0 1 745 252\nL 745 278\nL 745 278\nA 4 4 90.00 0 1 741 282\nL 655 282\nL 655 282\nA 4 4 90.00 0 1 651 278\nL 651 252\nL 651 252\nA 4 4 90.00 0 1 655 248\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"661\" y=\"265\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text><text x=\"655\" y=\"278\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 714 275\nL 800 275\nL 800 275\nA 4 4 90.00 0 1 804 279\nL 804 305\nL 804 305\nA 4 4 90.00 0 1 800 309\nL 714 309\nL 714 309\nA 4 4 90.00 0 1 710 305\nL 710 279\nL 710 279\nA 4 4 90.00 0 1 714 275\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"720\" y=\"292\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text><text x=\"714\" y=\"305\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text></svg>",
			pngCRC: 0xb9d45195,
		},
		{
			name: "trend_patterns",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 110, High: 120, Low: 100, Close: 120}, // Marubozu bullish - trend continuation
					{Open: 125, High: 125, Low: 115, Close: 115}, // Marubozu bearish - trend continuation
					{Open: 120, High: 130, Low: 115, Close: 125}, // Large bullish for belt hold setup
					{Open: 120, High: 140, Low: 120, Close: 140}, // Belt hold bullish - trend continuation
					{Open: 135, High: 135, Low: 115, Close: 115}, // Belt hold bearish - trend continuation
					{Open: 118, High: 125, Low: 117, Close: 122}, // Normal candle
					{Open: 122, High: 130, Low: 120, Close: 128}, // Trend continuation candle
				}
				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold:    0.01,
					ShadowRatio:      2.0,
					EngulfingMinSize: 0.8,
				})
				series.PatternConfig = PatternsTrend()
				return CandlestickChartOption{
					XAxis:      XAxisOption{Show: Ptr(false)},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">144</text><text x=\"9\" y=\"80\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">138</text><text x=\"9\" y=\"144\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">132</text><text x=\"9\" y=\"208\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">126</text><text x=\"9\" y=\"272\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">120</text><text x=\"9\" y=\"337\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">114</text><text x=\"9\" y=\"401\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">108</text><text x=\"9\" y=\"465\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">102</text><text x=\"18\" y=\"529\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">96</text><text x=\"18\" y=\"594\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 74\nL 790 74\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 138\nL 790 138\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 203\nL 790 203\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 267\nL 790 267\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 332\nL 790 332\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 396\nL 790 396\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 461\nL 790 461\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 525\nL 790 525\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 92 376\nL 92 429\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 92 483\nL 92 537\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 74 376\nL 110 376\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 74 537\nL 110 537\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 55 429\nL 129 429\nL 129 483\nL 55 483\nL 55 429\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 185 376\nL 185 483\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 167 268\nL 203 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 167 483\nL 203 483\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 148 268\nL 222 268\nL 222 376\nL 148 376\nL 148 268\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 260 215\nL 296 215\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 260 322\nL 296 322\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 241 215\nL 315 215\nL 315 322\nL 241 322\nL 241 215\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 371 161\nL 371 215\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 371 268\nL 371 322\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 353 161\nL 389 161\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 353 322\nL 389 322\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 334 215\nL 408 215\nL 408 268\nL 334 268\nL 334 215\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 446 53\nL 482 53\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 446 268\nL 482 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 427 53\nL 501 53\nL 501 268\nL 427 268\nL 427 53\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 539 107\nL 575 107\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 539 322\nL 575 322\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 520 107\nL 594 107\nL 594 322\nL 520 322\nL 520 107\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 650 215\nL 650 247\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 650 290\nL 650 300\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 632 215\nL 668 215\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 632 300\nL 668 300\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 613 247\nL 687 247\nL 687 290\nL 613 290\nL 613 247\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 743 161\nL 743 182\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 743 247\nL 743 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 725 161\nL 761 161\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 725 268\nL 761 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 706 182\nL 780 182\nL 780 247\nL 706 247\nL 706 182\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 283 303\nL 380 303\nL 380 303\nA 4 4 90.00 0 1 384 307\nL 384 333\nL 384 333\nA 4 4 90.00 0 1 380 337\nL 283 337\nL 283 337\nA 4 4 90.00 0 1 279 333\nL 279 307\nL 279 307\nA 4 4 90.00 0 1 283 303\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"287\" y=\"320\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">] Bear Belt Hold</text><text x=\"283\" y=\"333\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">v Bear Marubozu</text><path  d=\"M 469 34\nL 561 34\nL 561 34\nA 4 4 90.00 0 1 565 38\nL 565 64\nL 565 64\nA 4 4 90.00 0 1 561 68\nL 469 68\nL 469 68\nA 4 4 90.00 0 1 465 64\nL 465 38\nL 465 38\nA 4 4 90.00 0 1 469 34\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"473\" y=\"51\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">[ Bull Belt Hold</text><text x=\"469\" y=\"64\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">^ Bull Marubozu</text><path  d=\"M 562 303\nL 659 303\nL 659 303\nA 4 4 90.00 0 1 663 307\nL 663 333\nL 663 333\nA 4 4 90.00 0 1 659 337\nL 562 337\nL 562 337\nA 4 4 90.00 0 1 558 333\nL 558 307\nL 558 307\nA 4 4 90.00 0 1 562 303\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"566\" y=\"320\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">] Bear Belt Hold</text><text x=\"562\" y=\"333\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">v Bear Marubozu</text><path  d=\"M 655 234\nL 731 234\nL 731 234\nA 4 4 90.00 0 1 735 238\nL 735 251\nL 735 251\nA 4 4 90.00 0 1 731 255\nL 655 255\nL 655 255\nA 4 4 90.00 0 1 651 251\nL 651 238\nL 651 238\nA 4 4 90.00 0 1 655 234\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"655\" y=\"251\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ʘ Bull Harami</text></svg>",
			pngCRC: 0xabb45fa1,
		},
		{
			name: "all_patterns_showcase",
			optGen: func() CandlestickChartOption {
				// Comprehensive dataset showcasing all supported candlestick patterns
				data := []OHLCData{
					// 0: Setup - Normal candle
					{Open: 100, High: 110, Low: 95, Close: 105},
					// 1: Regular candle (reduce spinning top frequency)
					{Open: 105, High: 108, Low: 102, Close: 107},
					// 2: Hammer pattern
					{Open: 108, High: 109, Low: 98, Close: 107},
					// 3: Regular candle (was inverted hammer, reduce shooting star frequency)
					{Open: 95, High: 102, Low: 94, Close: 100},
					// 4: Regular candle (was shooting star, reduce frequency)
					{Open: 106, High: 115, Low: 105, Close: 112},
					// 5: Gravestone Doji pattern
					{Open: 108, High: 120, Low: 107, Close: 108.1},
					// 6: Hammer-like pattern (preserve dragonfly, reduce doji frequency)
					{Open: 109, High: 111, Low: 90, Close: 108},
					// 7: Bullish Marubozu pattern
					{Open: 100, High: 120, Low: 100, Close: 120},
					// 8: Bearish Marubozu pattern
					{Open: 120, High: 120, Low: 100, Close: 100},
					// 9: Regular candle (break harami pattern, reduce spinning top)
					{Open: 110, High: 120, Low: 107, Close: 118},
					// Setup for two-candle patterns - Large bearish candle
					{Open: 130, High: 135, Low: 110, Close: 115},
					// 11: Bullish Engulfing pattern
					{Open: 110, High: 140, Low: 108, Close: 138},
					// Setup for Bearish Engulfing - Large bullish candle
					{Open: 110, High: 140, Low: 108, Close: 138},
					// 13: Bearish Engulfing pattern (fixed to properly engulf)
					{Open: 140, High: 142, Low: 105, Close: 107},
					// Setup for Harami - Large bearish candle
					{Open: 130, High: 135, Low: 100, Close: 105},
					// 15: Regular candle (break harami by extending body)
					{Open: 110, High: 125, Low: 95, Close: 120},
					// Setup for Bearish Harami - Large bullish candle
					{Open: 100, High: 135, Low: 98, Close: 130},
					// 17: Bearish Harami pattern
					{Open: 125, High: 128, Low: 120, Close: 122},
					// Setup for Piercing Line - Bearish candle
					{Open: 120, High: 125, Low: 110, Close: 112},
					// 19: Piercing Line pattern
					{Open: 108, High: 125, Low: 107, Close: 118},
					// Setup for Dark Cloud Cover - Bullish candle
					{Open: 110, High: 125, Low: 108, Close: 123},
					// 21: Dark Cloud Cover pattern (fixed to gap up and close below midpoint)
					{Open: 128, High: 130, Low: 112, Close: 115},
					// Setup for Tweezer Top - Two candles with same high
					{Open: 110, High: 130, Low: 108, Close: 125},
					// 23: Tweezer Top pattern
					{Open: 123, High: 130, Low: 115, Close: 118},
					// Setup for Tweezer Bottom - Two candles with same low
					{Open: 120, High: 125, Low: 100, Close: 105},
					// 25: Tweezer Bottom pattern
					{Open: 108, High: 115, Low: 100, Close: 112},
					// Setup for Morning Star - Large bearish candle
					{Open: 130, High: 135, Low: 110, Close: 115},
					// 27: Morning Star middle - Small body with gap down (reduce spinning top)
					{Open: 108, High: 112, Low: 107, Close: 110},
					// 28: Morning Star completion - Large bullish candle
					{Open: 115, High: 140, Low: 113, Close: 135},
					// Setup for Evening Star - Large bullish candle
					{Open: 110, High: 140, Low: 108, Close: 135},
					// 30: Evening Star middle - Small body with proper gap up (fixed)
					{Open: 137, High: 145, Low: 136, Close: 140},
					// 31: Evening Star completion - Large bearish candle (fixed)
					{Open: 135, High: 136, Low: 115, Close: 120},
					// Setup for Three White Soldiers - Start with bearish sentiment
					{Open: 120, High: 125, Low: 110, Close: 115},
					// 33: Three White Soldiers - First soldier
					{Open: 118, High: 128, Low: 116, Close: 125},
					// 34: Three White Soldiers - Second soldier
					{Open: 127, High: 135, Low: 125, Close: 132},
					// 35: Three White Soldiers - Third soldier
					{Open: 134, High: 142, Low: 132, Close: 140},
					// Setup for Three Black Crows - Start with bullish sentiment
					{Open: 130, High: 145, Low: 128, Close: 142},
					// 37: Three Black Crows - First crow (fixed to open within previous body)
					{Open: 138, High: 140, Low: 128, Close: 132},
					// 38: Three Black Crows - Second crow (fixed to open within previous body)
					{Open: 130, High: 132, Low: 120, Close: 125},
					// 39: Three Black Crows - Third crow (fixed to open within previous body)
					{Open: 124, High: 127, Low: 115, Close: 118},
					// 40: Regular candle (reduce spinning top frequency)
					{Open: 115, High: 120, Low: 114, Close: 118},
					// 41: Regular candle (was spinning top, reduce frequency)
					{Open: 118, High: 125, Low: 115, Close: 122},
					// 42: Setup for Shooting Star - rising trend
					{Open: 120, High: 125, Low: 118, Close: 124},
					// 43: Shooting Star pattern - long upper shadow, small body near low
					{Open: 123, High: 140, Low: 122, Close: 125},
					// 44: Setup for Gravestone Doji - uptrend
					{Open: 125, High: 130, Low: 123, Close: 128},
					// 45: Gravestone Doji pattern - doji with long upper shadow
					{Open: 128, High: 145, Low: 127, Close: 128.05},
					// 46: Setup for Dragonfly Doji - downtrend
					{Open: 128, High: 130, Low: 125, Close: 126},
					// 47: Dragonfly Doji pattern - doji with long lower shadow
					{Open: 125, High: 126, Low: 110, Close: 125.05},
					// 48: Setup for Tweezer Bottom - bearish candle
					{Open: 125, High: 127, Low: 115, Close: 118},
					// 49: Tweezer Bottom pattern - same low as previous, bullish reversal
					{Open: 120, High: 125, Low: 115, Close: 123},
					// 50: Setup for Three Black Crows - high bullish candle
					{Open: 120, High: 135, Low: 118, Close: 133},
					// 51: Three Black Crows - First crow (bearish, substantial body)
					{Open: 132, High: 133, Low: 125, Close: 126},
					// 52: Three Black Crows - Second crow (bearish, opens within prev body, closes lower)
					{Open: 130, High: 131, Low: 121, Close: 122},
					// 53: Three Black Crows - Third crow (bearish, opens within prev body, closes lower)
					{Open: 125, High: 126, Low: 115, Close: 116},
					// 54: Long-Legged Doji pattern - very long shadows on both sides, small body
					{Open: 118, High: 135, Low: 95, Close: 118.1},
				}

				series := newCandlestickWithPatterns(data, CandlestickPatternConfig{
					DojiThreshold:    0.01,
					ShadowRatio:      2.0,
					EngulfingMinSize: 0.8,
				})

				return CandlestickChartOption{
					XAxis:      XAxisOption{Show: Ptr(false)},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">153</text><text x=\"9\" y=\"80\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">146</text><text x=\"9\" y=\"144\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">139</text><text x=\"9\" y=\"208\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">132</text><text x=\"9\" y=\"272\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125</text><text x=\"9\" y=\"337\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">118</text><text x=\"9\" y=\"401\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">111</text><text x=\"9\" y=\"465\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">104</text><text x=\"18\" y=\"529\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">97</text><text x=\"18\" y=\"594\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 74\nL 790 74\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 138\nL 790 138\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 203\nL 790 203\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 267\nL 790 267\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 332\nL 790 332\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 396\nL 790 396\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 461\nL 790 461\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 525\nL 790 525\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 52 406\nL 52 452\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 52 498\nL 52 544\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 50 406\nL 54 406\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 50 544\nL 54 544\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 47 452\nL 57 452\nL 57 498\nL 47 498\nL 47 452\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 66 425\nL 66 434\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 66 452\nL 66 480\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 64 425\nL 68 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 64 480\nL 68 480\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 61 434\nL 71 434\nL 71 452\nL 61 452\nL 61 434\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 79 416\nL 79 425\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 79 434\nL 79 517\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 77 416\nL 81 416\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 77 517\nL 81 517\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 74 425\nL 84 425\nL 84 434\nL 74 434\nL 74 425\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 93 480\nL 93 498\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 93 544\nL 93 554\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 480\nL 95 480\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 554\nL 95 554\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 88 498\nL 98 498\nL 98 544\nL 88 544\nL 88 498\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 106 360\nL 106 388\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 106 443\nL 106 452\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 104 360\nL 108 360\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 104 452\nL 108 452\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 101 388\nL 111 388\nL 111 443\nL 101 443\nL 101 388\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 120 314\nL 120 424\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 120 425\nL 120 434\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 118 314\nL 122 314\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 118 434\nL 122 434\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 115 424\nL 125 424\nL 125 425\nL 115 425\nL 115 424\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 133 397\nL 133 416\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 133 425\nL 133 590\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 131 397\nL 135 397\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 131 590\nL 135 590\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 128 416\nL 138 416\nL 138 425\nL 128 425\nL 128 416\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 145 314\nL 149 314\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 145 498\nL 149 498\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 142 314\nL 152 314\nL 152 498\nL 142 498\nL 142 314\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 158 314\nL 162 314\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 158 498\nL 162 498\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 155 314\nL 165 314\nL 165 498\nL 155 498\nL 155 314\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 174 314\nL 174 333\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 174 406\nL 174 434\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 172 314\nL 176 314\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 172 434\nL 176 434\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 169 333\nL 179 333\nL 179 406\nL 169 406\nL 169 333\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 187 176\nL 187 222\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 187 360\nL 187 406\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 185 176\nL 189 176\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 185 406\nL 189 406\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 182 222\nL 192 222\nL 192 360\nL 182 360\nL 182 222\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 201 130\nL 201 149\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 201 406\nL 201 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 199 130\nL 203 130\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 199 425\nL 203 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 196 149\nL 206 149\nL 206 406\nL 196 406\nL 196 149\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 214 130\nL 214 149\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 214 406\nL 214 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 212 130\nL 216 130\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 212 425\nL 216 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 209 149\nL 219 149\nL 219 406\nL 209 406\nL 209 149\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 228 112\nL 228 130\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 228 434\nL 228 452\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 226 112\nL 230 112\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 226 452\nL 230 452\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 223 130\nL 233 130\nL 233 434\nL 223 434\nL 223 130\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 241 176\nL 241 222\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 241 452\nL 241 498\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 239 176\nL 243 176\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 239 498\nL 243 498\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 236 222\nL 246 222\nL 246 452\nL 236 452\nL 236 222\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 255 268\nL 255 314\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 255 406\nL 255 544\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 253 268\nL 257 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 253 544\nL 257 544\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 250 314\nL 260 314\nL 260 406\nL 250 406\nL 250 314\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 268 176\nL 268 222\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 268 498\nL 268 517\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 266 176\nL 270 176\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 266 517\nL 270 517\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 263 222\nL 273 222\nL 273 498\nL 263 498\nL 263 222\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 282 241\nL 282 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 282 296\nL 282 314\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 280 241\nL 284 241\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 280 314\nL 284 314\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 277 268\nL 287 268\nL 287 296\nL 277 296\nL 277 268\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 296 268\nL 296 314\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 296 388\nL 296 406\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 294 268\nL 298 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 294 406\nL 298 406\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 291 314\nL 301 314\nL 301 388\nL 291 388\nL 291 314\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 309 268\nL 309 333\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 309 425\nL 309 434\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 307 268\nL 311 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 307 434\nL 311 434\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 304 333\nL 314 333\nL 314 425\nL 304 425\nL 304 333\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 323 268\nL 323 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 323 406\nL 323 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 321 268\nL 325 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 321 425\nL 325 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 318 287\nL 328 287\nL 328 406\nL 318 406\nL 318 287\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 336 222\nL 336 241\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 336 360\nL 336 388\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 334 222\nL 338 222\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 334 388\nL 338 388\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 331 241\nL 341 241\nL 341 360\nL 331 360\nL 331 241\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 350 222\nL 350 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 350 406\nL 350 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 348 222\nL 352 222\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 348 425\nL 352 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 345 268\nL 355 268\nL 355 406\nL 345 406\nL 345 268\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 363 222\nL 363 287\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 363 333\nL 363 360\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 361 222\nL 365 222\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 361 360\nL 365 360\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 358 287\nL 368 287\nL 368 333\nL 358 333\nL 358 287\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 377 268\nL 377 314\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 377 452\nL 377 498\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 375 268\nL 379 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 375 498\nL 379 498\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 372 314\nL 382 314\nL 382 452\nL 372 452\nL 372 314\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 390 360\nL 390 388\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 390 425\nL 390 498\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 388 360\nL 392 360\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 388 498\nL 392 498\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 385 388\nL 395 388\nL 395 425\nL 385 425\nL 385 388\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 404 176\nL 404 222\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 404 360\nL 404 406\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 402 176\nL 406 176\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 402 406\nL 406 406\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 399 222\nL 409 222\nL 409 360\nL 399 360\nL 399 222\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 417 388\nL 417 406\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 417 425\nL 417 434\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 415 388\nL 419 388\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 415 434\nL 419 434\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 412 406\nL 422 406\nL 422 425\nL 412 425\nL 412 406\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 431 130\nL 431 176\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 431 360\nL 431 379\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 429 130\nL 433 130\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 429 379\nL 433 379\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 426 176\nL 436 176\nL 436 360\nL 426 360\nL 426 176\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 444 130\nL 444 176\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 444 406\nL 444 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 442 130\nL 446 130\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 442 425\nL 446 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 439 176\nL 449 176\nL 449 406\nL 439 406\nL 439 176\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 458 84\nL 458 130\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 458 158\nL 458 167\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 456 84\nL 460 84\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 456 167\nL 460 167\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 453 130\nL 463 130\nL 463 158\nL 453 158\nL 453 130\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 471 167\nL 471 176\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 471 314\nL 471 360\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 469 167\nL 473 167\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 469 360\nL 473 360\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 466 176\nL 476 176\nL 476 314\nL 466 314\nL 466 176\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 485 268\nL 485 314\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 485 360\nL 485 406\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 483 268\nL 487 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 483 406\nL 487 406\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 480 314\nL 490 314\nL 490 360\nL 480 360\nL 480 314\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 498 241\nL 498 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 498 333\nL 498 351\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 496 241\nL 500 241\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 496 351\nL 500 351\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 493 268\nL 503 268\nL 503 333\nL 493 333\nL 493 268\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 512 176\nL 512 204\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 512 250\nL 512 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 510 176\nL 514 176\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 510 268\nL 514 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 507 204\nL 517 204\nL 517 250\nL 507 250\nL 507 204\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 525 112\nL 525 130\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 525 185\nL 525 204\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 523 112\nL 527 112\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 523 204\nL 527 204\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 520 130\nL 530 130\nL 530 185\nL 520 185\nL 520 130\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 539 84\nL 539 112\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 539 222\nL 539 241\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 537 84\nL 541 84\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 537 241\nL 541 241\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 534 112\nL 544 112\nL 544 222\nL 534 222\nL 534 112\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 553 130\nL 553 149\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 553 204\nL 553 241\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 551 130\nL 555 130\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 551 241\nL 555 241\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 548 149\nL 558 149\nL 558 204\nL 548 204\nL 548 149\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 566 204\nL 566 222\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 566 268\nL 566 314\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 564 204\nL 568 204\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 564 314\nL 568 314\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 561 222\nL 571 222\nL 571 268\nL 561 268\nL 561 222\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 580 250\nL 580 277\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 580 333\nL 580 360\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 578 250\nL 582 250\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 578 360\nL 582 360\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 575 277\nL 585 277\nL 585 333\nL 575 333\nL 575 277\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 593 314\nL 593 333\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 593 360\nL 593 370\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 591 314\nL 595 314\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 591 370\nL 595 370\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 588 333\nL 598 333\nL 598 360\nL 588 360\nL 588 333\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 607 268\nL 607 296\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 607 333\nL 607 360\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 605 268\nL 609 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 605 360\nL 609 360\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 602 296\nL 612 296\nL 612 333\nL 602 333\nL 602 296\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 620 268\nL 620 277\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 620 314\nL 620 333\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 618 268\nL 622 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 618 333\nL 622 333\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 615 277\nL 625 277\nL 625 314\nL 615 314\nL 615 277\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 634 130\nL 634 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 634 287\nL 634 296\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 632 130\nL 636 130\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 632 296\nL 636 296\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 629 268\nL 639 268\nL 639 287\nL 629 287\nL 629 268\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 647 222\nL 647 241\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 647 268\nL 647 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 645 222\nL 649 222\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 645 287\nL 649 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 642 241\nL 652 241\nL 652 268\nL 642 268\nL 642 241\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 661 84\nL 661 240\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 661 241\nL 661 250\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 659 84\nL 663 84\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 659 250\nL 663 250\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 656 240\nL 666 240\nL 666 241\nL 656 241\nL 656 240\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 674 222\nL 674 241\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 674 259\nL 674 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 672 222\nL 676 222\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 672 268\nL 676 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 669 241\nL 679 241\nL 679 259\nL 669 259\nL 669 241\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 688 259\nL 688 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 688 268\nL 688 406\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 686 259\nL 690 259\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 686 406\nL 690 406\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 683 268\nL 693 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 701 250\nL 701 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 701 333\nL 701 360\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 699 250\nL 703 250\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 699 360\nL 703 360\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 696 268\nL 706 268\nL 706 333\nL 696 333\nL 696 268\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 715 268\nL 715 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 715 314\nL 715 360\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 713 268\nL 717 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 713 360\nL 717 360\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 710 287\nL 720 287\nL 720 314\nL 710 314\nL 710 287\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 728 176\nL 728 195\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 728 314\nL 728 333\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 726 176\nL 730 176\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 726 333\nL 730 333\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 723 195\nL 733 195\nL 733 314\nL 723 314\nL 723 195\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 742 195\nL 742 204\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 742 259\nL 742 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 740 195\nL 744 195\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 740 268\nL 744 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 737 204\nL 747 204\nL 747 259\nL 737 259\nL 737 204\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 755 213\nL 755 222\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 755 296\nL 755 305\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 753 213\nL 757 213\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 753 305\nL 757 305\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 750 222\nL 760 222\nL 760 296\nL 750 296\nL 750 222\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 769 259\nL 769 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 769 351\nL 769 360\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 767 259\nL 771 259\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 767 360\nL 771 360\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 764 268\nL 774 268\nL 774 351\nL 764 351\nL 764 268\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 783 176\nL 783 332\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 783 333\nL 783 544\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 781 176\nL 785 176\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 781 544\nL 785 544\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 778 332\nL 788 332\nL 788 333\nL 778 333\nL 778 332\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 84 408\nL 170 408\nL 170 408\nA 4 4 90.00 0 1 174 412\nL 174 451\nL 174 451\nA 4 4 90.00 0 1 170 455\nL 84 455\nL 84 455\nA 4 4 90.00 0 1 80 451\nL 80 412\nL 80 412\nA 4 4 90.00 0 1 84 408\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"97\" y=\"425\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><text x=\"84\" y=\"438\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"90\" y=\"451\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text><path  d=\"M 125 379\nL 217 379\nL 217 379\nA 4 4 90.00 0 1 221 383\nL 221 461\nL 221 461\nA 4 4 90.00 0 1 217 465\nL 125 465\nL 125 465\nA 4 4 90.00 0 1 121 461\nL 121 383\nL 121 383\nA 4 4 90.00 0 1 125 379\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"125\" y=\"396\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"133\" y=\"409\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">† Gravestone</text><text x=\"154\" y=\"422\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text><text x=\"129\" y=\"435\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text><text x=\"128\" y=\"448\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"134\" y=\"461\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text><path  d=\"M 138 399\nL 224 399\nL 224 399\nA 4 4 90.00 0 1 228 403\nL 228 442\nL 228 442\nA 4 4 90.00 0 1 224 446\nL 138 446\nL 138 446\nA 4 4 90.00 0 1 134 442\nL 134 403\nL 134 403\nA 4 4 90.00 0 1 138 399\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"151\" y=\"416\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><text x=\"138\" y=\"429\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"144\" y=\"442\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text><path  d=\"M 152 288\nL 244 288\nL 244 288\nA 4 4 90.00 0 1 248 292\nL 248 331\nL 248 331\nA 4 4 90.00 0 1 244 335\nL 152 335\nL 152 335\nA 4 4 90.00 0 1 148 331\nL 148 292\nL 148 292\nA 4 4 90.00 0 1 152 288\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"152\" y=\"305\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Λ Bull Engulfing</text><text x=\"152\" y=\"318\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">^ Bull Marubozu</text><text x=\"156\" y=\"331\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">[ Bull Belt Hold</text><path  d=\"M 165 479\nL 262 479\nL 262 479\nA 4 4 90.00 0 1 266 483\nL 266 509\nL 266 509\nA 4 4 90.00 0 1 262 513\nL 165 513\nL 165 513\nA 4 4 90.00 0 1 161 509\nL 161 483\nL 161 483\nA 4 4 90.00 0 1 165 479\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"165\" y=\"496\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">v Bear Marubozu</text><text x=\"169\" y=\"509\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">] Bear Belt Hold</text><path  d=\"M 206 136\nL 297 136\nL 297 136\nA 4 4 90.00 0 1 301 140\nL 301 153\nL 301 153\nA 4 4 90.00 0 1 297 157\nL 206 157\nL 206 157\nA 4 4 90.00 0 1 202 153\nL 202 140\nL 202 140\nA 4 4 90.00 0 1 206 136\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"206\" y=\"153\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Λ Bull Engulfing</text><path  d=\"M 233 421\nL 328 421\nL 328 421\nA 4 4 90.00 0 1 332 425\nL 332 438\nL 332 438\nA 4 4 90.00 0 1 328 442\nL 233 442\nL 233 442\nA 4 4 90.00 0 1 229 438\nL 229 425\nL 229 425\nA 4 4 90.00 0 1 233 421\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"233\" y=\"438\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">V Bear Engulfing</text><path  d=\"M 287 283\nL 369 283\nL 369 283\nA 4 4 90.00 0 1 373 287\nL 373 300\nL 373 300\nA 4 4 90.00 0 1 369 304\nL 287 304\nL 287 304\nA 4 4 90.00 0 1 283 300\nL 283 287\nL 283 287\nA 4 4 90.00 0 1 287 283\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"287\" y=\"300\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">θ Bear Harami</text><path  d=\"M 314 320\nL 395 320\nL 395 320\nA 4 4 90.00 0 1 399 324\nL 399 337\nL 399 337\nA 4 4 90.00 0 1 395 341\nL 314 341\nL 314 341\nA 4 4 90.00 0 1 310 337\nL 310 324\nL 310 324\nA 4 4 90.00 0 1 314 320\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"314\" y=\"337\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">| Piercing Line</text><path  d=\"M 341 347\nL 415 347\nL 415 347\nA 4 4 90.00 0 1 419 351\nL 419 364\nL 419 364\nA 4 4 90.00 0 1 415 368\nL 341 368\nL 341 368\nA 4 4 90.00 0 1 337 364\nL 337 351\nL 337 351\nA 4 4 90.00 0 1 341 347\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"341\" y=\"364\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ξ Dark Cloud</text><path  d=\"M 355 255\nL 436 255\nL 436 255\nA 4 4 90.00 0 1 440 259\nL 440 272\nL 440 272\nA 4 4 90.00 0 1 436 276\nL 355 276\nL 355 276\nA 4 4 90.00 0 1 351 272\nL 351 259\nL 351 259\nA 4 4 90.00 0 1 355 255\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"355\" y=\"272\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">| Piercing Line</text><path  d=\"M 436 163\nL 520 163\nL 520 163\nA 4 4 90.00 0 1 524 167\nL 524 180\nL 524 180\nA 4 4 90.00 0 1 520 184\nL 436 184\nL 436 184\nA 4 4 90.00 0 1 432 180\nL 432 167\nL 432 167\nA 4 4 90.00 0 1 436 163\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"436\" y=\"180\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">* Morning Star</text><path  d=\"M 476 301\nL 558 301\nL 558 301\nA 4 4 90.00 0 1 562 305\nL 562 318\nL 562 318\nA 4 4 90.00 0 1 558 322\nL 476 322\nL 476 322\nA 4 4 90.00 0 1 472 318\nL 472 305\nL 472 305\nA 4 4 90.00 0 1 476 301\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"476\" y=\"318\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">⁎ Evening Star</text><path  d=\"M 639 249\nL 731 249\nL 731 249\nA 4 4 90.00 0 1 735 253\nL 735 279\nL 735 279\nA 4 4 90.00 0 1 731 283\nL 639 283\nL 639 283\nA 4 4 90.00 0 1 635 279\nL 635 253\nL 635 253\nA 4 4 90.00 0 1 639 249\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"639\" y=\"266\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"643\" y=\"279\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text><path  d=\"M 666 195\nL 758 195\nL 758 195\nA 4 4 90.00 0 1 762 199\nL 762 277\nL 762 277\nA 4 4 90.00 0 1 758 281\nL 666 281\nL 666 281\nA 4 4 90.00 0 1 662 277\nL 662 199\nL 662 199\nA 4 4 90.00 0 1 666 195\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"666\" y=\"212\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"674\" y=\"225\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">† Gravestone</text><text x=\"695\" y=\"238\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text><text x=\"670\" y=\"251\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text><text x=\"669\" y=\"264\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"675\" y=\"277\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text><path  d=\"M 693 229\nL 779 229\nL 779 229\nA 4 4 90.00 0 1 783 233\nL 783 298\nL 783 298\nA 4 4 90.00 0 1 779 302\nL 693 302\nL 693 302\nA 4 4 90.00 0 1 689 298\nL 689 233\nL 689 233\nA 4 4 90.00 0 1 693 229\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"706\" y=\"246\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><text x=\"702\" y=\"259\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ψ Dragonfly</text><text x=\"719\" y=\"272\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text><text x=\"693\" y=\"285\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"699\" y=\"298\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text><path  d=\"M 690 293\nL 800 293\nL 800 293\nA 4 4 90.00 0 1 804 297\nL 804 362\nL 804 362\nA 4 4 90.00 0 1 800 366\nL 690 366\nL 690 366\nA 4 4 90.00 0 1 686 362\nL 686 297\nL 686 297\nA 4 4 90.00 0 1 690 293\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"728\" y=\"310\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text><text x=\"707\" y=\"323\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ʘ Bull Harami</text><text x=\"702\" y=\"336\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"690\" y=\"349\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">‡ Long Legged Doji</text><text x=\"708\" y=\"362\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">~ High Wave</text></svg>",
			pngCRC: 0x980d0993,
		},
	}

	for i, tc := range tests {
		t.Run(strconv.Itoa(i)+"-"+tc.name, func(t *testing.T) {
			p := NewPainter(PainterOptions{
				OutputFormat: ChartOutputSVG,
				Width:        800,
				Height:       600,
			})
			r := NewPainter(PainterOptions{
				OutputFormat: ChartOutputPNG,
				Width:        800,
				Height:       600,
			})

			opt := tc.optGen()
			opt.Theme = GetTheme(ThemeVividLight)

			validateCandlestickChartRender(t, p, r, opt, tc.svg, tc.pngCRC)
		})
	}
}

func TestCandlestickPatternConfigMergePatterns(t *testing.T) {
	t.Parallel()

	t.Run("merge_two_configs", func(t *testing.T) {
		config1 := &CandlestickPatternConfig{
			ReplaceSeriesLabel: true,
			EnabledPatterns:    []string{CandlestickPatternDoji, CandlestickPatternHammer},
			DojiThreshold:      0.01,
		}
		config2 := &CandlestickPatternConfig{
			ReplaceSeriesLabel: false,
			EnabledPatterns:    []string{CandlestickPatternShootingStar, CandlestickPatternDoji}, // Doji is duplicate
			DojiThreshold:      0.02,
		}

		merged := config1.MergePatterns(config2)

		// Should preserve config1's settings
		assert.True(t, merged.ReplaceSeriesLabel)
		assert.InDelta(t, 0.01, merged.DojiThreshold, 0)

		// Should have union of patterns without duplicates, preserving order
		assert.Len(t, merged.EnabledPatterns, 3)
		assert.Equal(t, CandlestickPatternDoji, merged.EnabledPatterns[0])
		assert.Equal(t, CandlestickPatternHammer, merged.EnabledPatterns[1])
		assert.Equal(t, CandlestickPatternShootingStar, merged.EnabledPatterns[2])
	})

	t.Run("merge_with_nil", func(t *testing.T) {
		config := &CandlestickPatternConfig{
			ReplaceSeriesLabel: true,
			EnabledPatterns:    []string{CandlestickPatternDoji, CandlestickPatternHammer},
		}

		// Merge nil with config
		var nilConfig *CandlestickPatternConfig
		merged1 := nilConfig.MergePatterns(config)
		assert.NotNil(t, merged1)
		assert.True(t, merged1.ReplaceSeriesLabel)
		assert.Len(t, merged1.EnabledPatterns, 2)

		// Merge config with nil
		merged2 := config.MergePatterns(nil)
		assert.NotNil(t, merged2)
		assert.True(t, merged2.ReplaceSeriesLabel)
		assert.Len(t, merged2.EnabledPatterns, 2)

		// Merge nil with nil
		merged3 := nilConfig.MergePatterns(nil)
		assert.Nil(t, merged3)
	})

	t.Run("merge_identical_patterns", func(t *testing.T) {
		config1 := &CandlestickPatternConfig{
			EnabledPatterns: []string{CandlestickPatternDoji, CandlestickPatternHammer, CandlestickPatternShootingStar},
		}
		config2 := &CandlestickPatternConfig{
			EnabledPatterns: []string{CandlestickPatternDoji, CandlestickPatternHammer, CandlestickPatternShootingStar},
		}

		merged := config1.MergePatterns(config2)
		assert.Len(t, merged.EnabledPatterns, 3) // No duplicates
		assert.Equal(t, CandlestickPatternDoji, merged.EnabledPatterns[0])
		assert.Equal(t, CandlestickPatternHammer, merged.EnabledPatterns[1])
		assert.Equal(t, CandlestickPatternShootingStar, merged.EnabledPatterns[2])
	})

	t.Run("merge_empty_patterns", func(t *testing.T) {
		config1 := &CandlestickPatternConfig{
			ReplaceSeriesLabel: true,
			EnabledPatterns:    []string{},
		}
		config2 := &CandlestickPatternConfig{
			EnabledPatterns: []string{CandlestickPatternDoji, CandlestickPatternHammer},
		}

		merged := config1.MergePatterns(config2)
		assert.True(t, merged.ReplaceSeriesLabel)
		assert.Len(t, merged.EnabledPatterns, 2)
		assert.Equal(t, CandlestickPatternDoji, merged.EnabledPatterns[0])
		assert.Equal(t, CandlestickPatternHammer, merged.EnabledPatterns[1])
	})

	t.Run("merge_predefined_configs", func(t *testing.T) {
		important := PatternsImportant()
		indecision := PatternsIndecision()

		merged := important.MergePatterns(indecision)

		// Should have all patterns from both configs
		assert.Greater(t, len(merged.EnabledPatterns), len(important.EnabledPatterns))
		assert.Greater(t, len(merged.EnabledPatterns), len(indecision.EnabledPatterns))

		// Should preserve important config's settings
		assert.Equal(t, important.ReplaceSeriesLabel, merged.ReplaceSeriesLabel)

		// Should contain patterns from both
		assert.Contains(t, merged.EnabledPatterns, CandlestickPatternEngulfingBull)  // From important
		assert.Contains(t, merged.EnabledPatterns, CandlestickPatternLongLeggedDoji) // From indecision
	})
}

// extract dashed trend line x coordinates from svg
func extractDashedPathXCoords(svg string) [][]int {
	rePath := regexp.MustCompile(`<path[^>]*stroke-dasharray[^>]*d="([^"]+)"`)
	paths := rePath.FindAllStringSubmatch(svg, -1)
	coordRe := regexp.MustCompile(`[ML] ([0-9]+) [0-9]+`)
	result := make([][]int, len(paths))
	for i, p := range paths {
		coords := coordRe.FindAllStringSubmatch(p[1], -1)
		xs := make([]int, len(coords))
		for j, c := range coords {
			x, _ := strconv.Atoi(c[1])
			xs[j] = x
		}
		result[i] = xs
	}
	return result
}

// compute expected center positions (absolute) for series
func computeCenters(r *defaultRenderResult, opt CandlestickChartOption, seriesIndex int) []int {
	width := r.seriesPainter.Width()
	seriesCount := opt.SeriesList.len()
	maxDataCount := getSeriesMaxDataCount(opt.SeriesList)
	candleWidthRatio := opt.CandleWidth
	if candleWidthRatio <= 0 {
		candleWidthRatio = 0.8
	}
	candleWidth := int(float64(width) * candleWidthRatio / float64(maxDataCount))
	if candleWidth < 1 {
		candleWidth = 1
	}
	candleWidthPerSeries := candleWidth / seriesCount
	if candleWidthPerSeries < 1 {
		candleWidthPerSeries = 1
	}
	divideValues := r.xaxisRange.autoDivide()
	centers := make([]int, len(opt.SeriesList.getSeries(seriesIndex).(*CandlestickSeries).Data))
	for j := range centers {
		if j >= len(divideValues) {
			continue
		}
		var sectionWidth int
		if j < len(divideValues)-1 {
			sectionWidth = divideValues[j+1] - divideValues[j]
		} else if j > 0 {
			sectionWidth = divideValues[j] - divideValues[j-1]
		} else {
			sectionWidth = width / maxDataCount
		}
		var groupMargin, candleMargin, cWidth int
		if seriesCount == 1 {
			cWidth = candleWidthPerSeries
		} else {
			var candleMarginFloat *float64
			if opt.CandleMargin != nil {
				marginPixels := float64(sectionWidth) * (*opt.CandleMargin)
				candleMarginFloat = &marginPixels
			}
			groupMargin, candleMargin, cWidth = calculateCandleMarginsAndSize(seriesCount, sectionWidth, candleWidthPerSeries, candleMarginFloat)
		}
		var center int
		if seriesCount == 1 {
			center = divideValues[j] + sectionWidth/2
		} else {
			x := divideValues[j] + groupMargin + seriesIndex*(cWidth+candleMargin)
			center = x + cWidth/2
		}
		centers[j] = center + r.seriesPainter.box.Left
	}
	return centers
}

func TestCandlestickTrendLineAlignmentSingleSeries(t *testing.T) {
	p := NewPainter(PainterOptions{OutputFormat: ChartOutputSVG, Width: 600, Height: 400})
	data := []OHLCData{{Open: 100, High: 110, Low: 90, Close: 105}, {Open: 105, High: 115, Low: 95, Close: 108}, {Open: 108, High: 118, Low: 100, Close: 112}}
	opt := CandlestickChartOption{
		Theme:   GetDefaultTheme(),
		Padding: NewBoxEqual(0),
		XAxis:   XAxisOption{Labels: []string{"A", "B", "C"}, Show: Ptr(false)},
		YAxis:   make([]YAxisOption, 1),
		SeriesList: CandlestickSeriesList{{
			Data:           data,
			CloseTrendLine: []SeriesTrendLine{{Type: SeriesTrendTypeLinear, DashedLine: Ptr(true)}},
		}},
		ShowWicks: Ptr(false),
	}

	renderResult, err := defaultRender(p, defaultRenderOption{
		theme:          opt.Theme,
		padding:        opt.Padding,
		seriesList:     &opt.SeriesList,
		xAxis:          &opt.XAxis,
		yAxis:          opt.YAxis,
		title:          opt.Title,
		legend:         &opt.Legend,
		valueFormatter: opt.ValueFormatter,
	})
	require.NoError(t, err)

	_, err = newCandlestickChart(p, opt).renderChart(renderResult)
	require.NoError(t, err)
	svgBytes, err := p.Bytes()
	require.NoError(t, err)

	paths := extractDashedPathXCoords(string(svgBytes))
	require.Len(t, paths, 1)
	expected := computeCenters(renderResult, opt, 0)
	assert.Equal(t, expected, paths[0])
}

func TestCandlestickTrendLineAlignmentMultiSeries(t *testing.T) {
	p := NewPainter(PainterOptions{OutputFormat: ChartOutputSVG, Width: 600, Height: 400})
	data := []OHLCData{{Open: 100, High: 110, Low: 90, Close: 105}, {Open: 105, High: 115, Low: 95, Close: 108}, {Open: 108, High: 118, Low: 100, Close: 112}}
	opt := CandlestickChartOption{
		Theme:   GetDefaultTheme(),
		Padding: NewBoxEqual(0),
		XAxis:   XAxisOption{Labels: []string{"A", "B", "C"}, Show: Ptr(false)},
		YAxis:   make([]YAxisOption, 1),
		SeriesList: CandlestickSeriesList{
			{Data: data, CloseTrendLine: []SeriesTrendLine{{Type: SeriesTrendTypeLinear, DashedLine: Ptr(true)}}},
			{Data: data, CloseTrendLine: []SeriesTrendLine{{Type: SeriesTrendTypeLinear, DashedLine: Ptr(true)}}},
		},
		ShowWicks: Ptr(false),
	}

	renderResult, err := defaultRender(p, defaultRenderOption{
		theme:          opt.Theme,
		padding:        opt.Padding,
		seriesList:     &opt.SeriesList,
		xAxis:          &opt.XAxis,
		yAxis:          opt.YAxis,
		title:          opt.Title,
		legend:         &opt.Legend,
		valueFormatter: opt.ValueFormatter,
	})
	require.NoError(t, err)

	_, err = newCandlestickChart(p, opt).renderChart(renderResult)
	require.NoError(t, err)
	svgBytes, err := p.Bytes()
	require.NoError(t, err)

	paths := extractDashedPathXCoords(string(svgBytes))
	require.Len(t, paths, 2)
	expected0 := computeCenters(renderResult, opt, 0)
	expected1 := computeCenters(renderResult, opt, 1)
	assert.Equal(t, expected0, paths[0])
	assert.Equal(t, expected1, paths[1])
}
