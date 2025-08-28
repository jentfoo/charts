package charts

import (
	"strconv"
	"testing"

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
	}
}

func TestDojiPattern(t *testing.T) {
	t.Parallel()

	// Valid doji: open ≈ close
	doji := OHLCData{Open: 100, High: 105, Low: 95, Close: 100.1}
	assert.True(t, DetectDoji(doji, 0.01))

	// Invalid: body too large
	notDoji := OHLCData{Open: 100, High: 105, Low: 95, Close: 103}
	assert.False(t, DetectDoji(notDoji, 0.01))

	// Invalid: invalid OHLC
	invalidOHLC := OHLCData{Open: 100, High: 95, Low: 105, Close: 98}
	assert.False(t, DetectDoji(invalidOHLC, 0.01))
}

func TestHammerPattern(t *testing.T) {
	t.Parallel()

	// Valid hammer: long lower shadow, small body at top
	hammer := OHLCData{Open: 105, High: 107, Low: 95, Close: 106}
	assert.True(t, DetectHammer(hammer, 2.0))

	// Invalid: short lower shadow
	notHammer := OHLCData{Open: 105, High: 107, Low: 104, Close: 106}
	assert.False(t, DetectHammer(notHammer, 2.0))

	// Invalid: long upper shadow
	notHammer2 := OHLCData{Open: 95, High: 107, Low: 94, Close: 96}
	assert.False(t, DetectHammer(notHammer2, 2.0))
}

func TestInvertedHammerPattern(t *testing.T) {
	t.Parallel()

	// Valid inverted hammer: long upper shadow, small body at bottom
	invertedHammer := OHLCData{Open: 95, High: 107, Low: 94, Close: 96}
	assert.True(t, DetectInvertedHammer(invertedHammer, 2.0))

	// Invalid: short upper shadow
	notInvertedHammer := OHLCData{Open: 95, High: 97, Low: 94, Close: 96}
	assert.False(t, DetectInvertedHammer(notInvertedHammer, 2.0))
}

func TestEngulfingPattern(t *testing.T) {
	t.Parallel()

	// Test Bullish Engulfing
	prevBearish := OHLCData{Open: 110, High: 112, Low: 105, Close: 106}
	currentBullish := OHLCData{Open: 104, High: 115, Low: 103, Close: 114}

	bullish, bearish := DetectEngulfing(prevBearish, currentBullish, 0.8)
	assert.True(t, bullish)
	assert.False(t, bearish)

	// Test Bearish Engulfing
	prevBullish := OHLCData{Open: 106, High: 112, Low: 105, Close: 110}
	currentBearish := OHLCData{Open: 114, High: 115, Low: 103, Close: 104}

	bullish, bearish = DetectEngulfing(prevBullish, currentBearish, 0.8)
	assert.False(t, bullish)
	assert.True(t, bearish)

	// Test non-engulfing
	nonEngulfing := OHLCData{Open: 107, High: 109, Low: 106, Close: 108}
	bullish, bearish = DetectEngulfing(prevBullish, nonEngulfing, 0.8)
	assert.False(t, bullish)
	assert.False(t, bearish)
}

func TestHaramiPatterns(t *testing.T) {
	t.Parallel()

	// Test Bullish Harami
	prevCandle := OHLCData{Open: 110, High: 115, Low: 95, Close: 98}      // Large bearish
	currentCandle := OHLCData{Open: 102, High: 106, Low: 100, Close: 104} // Small bullish inside

	bullishHarami, bearishHarami := DetectHarami(prevCandle, currentCandle, 0.3)
	assert.True(t, bullishHarami)
	assert.False(t, bearishHarami)

	// Test Bearish Harami
	prevCandle = OHLCData{Open: 98, High: 115, Low: 95, Close: 110}      // Large bullish
	currentCandle = OHLCData{Open: 106, High: 108, Low: 102, Close: 104} // Small bearish inside

	bullishHarami, bearishHarami = DetectHarami(prevCandle, currentCandle, 0.3)
	assert.False(t, bullishHarami)
	assert.True(t, bearishHarami)

	// Test non-harami (current candle too large)
	currentCandle = OHLCData{Open: 100, High: 112, Low: 96, Close: 108} // Too large
	bullishHarami, bearishHarami = DetectHarami(prevCandle, currentCandle, 0.3)
	assert.False(t, bullishHarami)
	assert.False(t, bearishHarami)
}

func TestShootingStarPattern(t *testing.T) {
	t.Parallel()

	// Valid shooting star: small body at bottom, long upper shadow
	shootingStar := OHLCData{Open: 106, High: 125, Low: 105, Close: 107}
	assert.True(t, DetectShootingStar(shootingStar, 2.0))

	// Invalid: body not near bottom
	notShootingStar := OHLCData{Open: 115, High: 125, Low: 105, Close: 117}
	assert.False(t, DetectShootingStar(notShootingStar, 2.0))

	// Invalid: upper shadow too short
	shortShadow := OHLCData{Open: 106, High: 110, Low: 105, Close: 107}
	assert.False(t, DetectShootingStar(shortShadow, 2.0))
}

func TestGravestoneDojiPattern(t *testing.T) {
	t.Parallel()

	opt := PatternDetectionOption{DojiThreshold: 0.01, ShadowRatio: 2.0}

	// Valid gravestone doji: doji with long upper shadow
	gravestoneDoji := OHLCData{Open: 108, High: 120, Low: 107, Close: 108.1}
	assert.True(t, DetectGravestoneDoji(gravestoneDoji, opt))

	// Invalid: not a doji (body too large)
	notDoji := OHLCData{Open: 108, High: 120, Low: 107, Close: 115}
	assert.False(t, DetectGravestoneDoji(notDoji, opt))

	// Invalid: doji but no long upper shadow
	dojiNoShadow := OHLCData{Open: 108, High: 109, Low: 107, Close: 108.1}
	assert.False(t, DetectGravestoneDoji(dojiNoShadow, opt))
}

func TestDragonflyDojiPattern(t *testing.T) {
	t.Parallel()

	opt := PatternDetectionOption{DojiThreshold: 0.01, ShadowRatio: 2.0}

	// Valid dragonfly doji: doji with long lower shadow
	dragonflyDoji := OHLCData{Open: 109, High: 110, Low: 90, Close: 108.9}
	assert.True(t, DetectDragonflyDoji(dragonflyDoji, opt))

	// Invalid: not a doji
	notDoji := OHLCData{Open: 109, High: 110, Low: 90, Close: 102}
	assert.False(t, DetectDragonflyDoji(notDoji, opt))

	// Invalid: doji but no long lower shadow
	dojiNoShadow := OHLCData{Open: 109, High: 110, Low: 108, Close: 108.9}
	assert.False(t, DetectDragonflyDoji(dojiNoShadow, opt))
}

func TestMorningStarPattern(t *testing.T) {
	t.Parallel()

	opt := PatternDetectionOption{}

	// Valid morning star pattern
	first := OHLCData{Open: 120, High: 125, Low: 105, Close: 108}  // Large bearish
	second := OHLCData{Open: 102, High: 104, Low: 100, Close: 103} // Small body, gap down
	third := OHLCData{Open: 108, High: 125, Low: 106, Close: 122}  // Large bullish, gap up

	assert.True(t, DetectMorningStar(first, second, third, opt))

	// Invalid: first candle not bearish
	invalidFirst := OHLCData{Open: 108, High: 125, Low: 105, Close: 120} // Bullish
	assert.False(t, DetectMorningStar(invalidFirst, second, third, opt))

	// Invalid: no gap down between first and second
	noGapSecond := OHLCData{Open: 109, High: 111, Low: 107, Close: 110} // No gap
	assert.False(t, DetectMorningStar(first, noGapSecond, third, opt))

	// Invalid: third candle not bullish
	invalidThird := OHLCData{Open: 108, High: 110, Low: 105, Close: 107} // Bearish
	assert.False(t, DetectMorningStar(first, second, invalidThird, opt))
}

func TestEveningStarPattern(t *testing.T) {
	t.Parallel()

	opt := PatternDetectionOption{}

	// Valid evening star pattern
	first := OHLCData{Open: 122, High: 140, Low: 120, Close: 138}  // Large bullish
	second := OHLCData{Open: 142, High: 144, Low: 140, Close: 143} // Small body, gap up
	third := OHLCData{Open: 138, High: 140, Low: 115, Close: 118}  // Large bearish, gap down

	assert.True(t, DetectEveningStar(first, second, third, opt))

	// Invalid: first candle not bullish
	invalidFirst := OHLCData{Open: 138, High: 140, Low: 120, Close: 122} // Bearish
	assert.False(t, DetectEveningStar(invalidFirst, second, third, opt))

	// Invalid: no gap up between first and second
	noGapSecond := OHLCData{Open: 136, High: 140, Low: 134, Close: 139} // No gap
	assert.False(t, DetectEveningStar(first, noGapSecond, third, opt))

	// Invalid: third candle not bearish
	invalidThird := OHLCData{Open: 138, High: 145, Low: 135, Close: 142} // Bullish
	assert.False(t, DetectEveningStar(first, second, invalidThird, opt))
}

func newCandlestickWithPatterns(data []OHLCData, options ...PatternDetectionOption) CandlestickSeries {
	// Start with defaults and override with provided options
	detectionOptions := DefaultPatternOptions()
	if len(options) > 0 {
		// Merge provided options with defaults
		opt := options[0]
		if opt.DojiThreshold != 0 {
			detectionOptions.DojiThreshold = opt.DojiThreshold
		}
		if opt.ShadowRatio != 0 {
			detectionOptions.ShadowRatio = opt.ShadowRatio
		}
		if opt.EngulfingMinSize != 0 {
			detectionOptions.EngulfingMinSize = opt.EngulfingMinSize
		}
	}

	series := CandlestickSeries{
		Data: data,
		PatternConfig: &CandlestickPatternConfig{
			ReplaceSeriesLabel: true,
			EnabledPatterns:    PatternsAll().EnabledPatterns,
			DetectionOptions:   detectionOptions,
		},
	}
	return series
}

func TestPatternIntegration(t *testing.T) {
	t.Parallel()

	// Test that all advanced patterns are detected in a comprehensive dataset
	data := makeAdvancedPatternTestData()
	series := CandlestickSeries{Data: data}

	// Use private scan function for testing pattern detection
	patterns := scanCandlestickPatterns(series, PatternDetectionOption{
		DojiThreshold:    0.01,
		ShadowRatio:      2.0,
		EngulfingMinSize: 0.8,
	})

	// Verify we detected some patterns
	assert.NotEmpty(t, patterns)

	// Check for specific pattern types
	patternTypes := make(map[string]int)
	for _, pattern := range patterns {
		patternTypes[pattern.PatternType]++
	}

	// We should have detected various pattern types
	assert.Len(t, patternTypes, 12)

	// Test the convenience function
	seriesWithPatterns := newCandlestickWithPatterns(data, PatternDetectionOption{
		DojiThreshold:    0.01,
		ShadowRatio:      2.0,
		EngulfingMinSize: 0.8,
	})

	// Verify that pattern configuration is properly set
	assert.NotNil(t, seriesWithPatterns.PatternConfig)
	assert.True(t, seriesWithPatterns.PatternConfig.ReplaceSeriesLabel)
	assert.NotEmpty(t, seriesWithPatterns.PatternConfig.EnabledPatterns)

	// The key test: ensure that pattern scanning finds patterns
	assert.Len(t, patterns, 20)
}

func TestMarubozuPatern(t *testing.T) {
	t.Parallel()

	// Bullish Marubozu - no shadows
	bullishMarubozu := OHLCData{Open: 100, High: 120, Low: 100, Close: 120}
	bullish, bearish := DetectMarubozu(bullishMarubozu, 0.01)
	assert.True(t, bullish)
	assert.False(t, bearish)

	// Bearish Marubozu - no shadows
	bearishMarubozu := OHLCData{Open: 120, High: 120, Low: 100, Close: 100}
	bullish, bearish = DetectMarubozu(bearishMarubozu, 0.01)
	assert.False(t, bullish)
	assert.True(t, bearish)

	// Not a marubozu - has significant shadows
	notMarubozu := OHLCData{Open: 105, High: 125, Low: 95, Close: 115}
	bullish, bearish = DetectMarubozu(notMarubozu, 0.01)
	assert.False(t, bullish)
	assert.False(t, bearish)
}

func TestSpinningTopPattern(t *testing.T) {
	t.Parallel()

	// Classic spinning top - small body, long shadows
	spinningTop := OHLCData{Open: 110, High: 125, Low: 95, Close: 112}
	detected := DetectSpinningTop(spinningTop, 0.3)
	assert.True(t, detected)

	// Not spinning top - large body
	largeBody := OHLCData{Open: 100, High: 125, Low: 95, Close: 120}
	detected = DetectSpinningTop(largeBody, 0.3)
	assert.False(t, detected)

	// Not spinning top - shadows too short relative to body
	shortShadows := OHLCData{Open: 110, High: 110.5, Low: 109.5, Close: 111}
	detected = DetectSpinningTop(shortShadows, 0.3)
	assert.False(t, detected)
}

func TestPiercingLinePattern(t *testing.T) {
	t.Parallel()

	// Classic piercing line - bearish then bullish with gap down and close above midpoint
	prev := OHLCData{Open: 120, High: 120, Low: 110, Close: 110}    // Bearish
	current := OHLCData{Open: 108, High: 118, Low: 108, Close: 116} // Bullish, opens below prev low, closes above midpoint (115)
	detected := DetectPiercingLine(prev, current)
	assert.True(t, detected)

	// Not piercing line - current closes below midpoint
	current = OHLCData{Open: 108, High: 114, Low: 108, Close: 112}
	detected = DetectPiercingLine(prev, current)
	assert.False(t, detected)
}

func TestDarkCloudCoverPattern(t *testing.T) {
	t.Parallel()

	// Classic dark cloud cover - bullish then bearish with gap up and close below midpoint
	prev := OHLCData{Open: 110, High: 120, Low: 110, Close: 120}    // Bullish
	current := OHLCData{Open: 122, High: 122, Low: 112, Close: 114} // Bearish, opens above prev high, closes below midpoint (115)
	detected := DetectDarkCloudCover(prev, current)
	assert.True(t, detected)

	// Not dark cloud cover - current closes above midpoint
	current = OHLCData{Open: 122, High: 122, Low: 118, Close: 118}
	detected = DetectDarkCloudCover(prev, current)
	assert.False(t, detected)
}

func TestTweezerPattern(t *testing.T) {
	t.Parallel()

	// Tweezer tops - both highs at 125 (similar resistance)
	prev := OHLCData{Open: 120, High: 125, Low: 118, Close: 124}    // Bullish
	current := OHLCData{Open: 124, High: 125, Low: 119, Close: 121} // Bearish, same high
	detected := DetectTweezerTops(prev, current, 0.005)
	assert.True(t, detected)

	// Tweezer bottoms - both lows at 100 (similar support)
	prev = OHLCData{Open: 105, High: 108, Low: 100, Close: 102}    // Bearish
	current = OHLCData{Open: 102, High: 107, Low: 100, Close: 106} // Bullish, same low
	detected = DetectTweezerBottoms(prev, current, 0.005)
	assert.True(t, detected)
}

func TestThreeWhiteSoldiersPattern(t *testing.T) {
	t.Parallel()

	// Three white soldiers - three consecutive bullish candles
	first := OHLCData{Open: 100, High: 105, Low: 99, Close: 104}
	second := OHLCData{Open: 103, High: 108, Low: 102, Close: 107}
	third := OHLCData{Open: 106, High: 111, Low: 105, Close: 110}
	detected := DetectThreeWhiteSoldiers(first, second, third)
	assert.True(t, detected)

	// Not three white soldiers - third candle closes lower
	third = OHLCData{Open: 106, High: 108, Low: 105, Close: 106}
	detected = DetectThreeWhiteSoldiers(first, second, third)
	assert.False(t, detected)
}

func TestThreeBlackCrowsPattern(t *testing.T) {
	t.Parallel()

	// Classic three black crows - three consecutive bearish candles
	first := OHLCData{Open: 110, High: 111, Low: 105, Close: 106}
	second := OHLCData{Open: 107, High: 108, Low: 102, Close: 103}
	third := OHLCData{Open: 104, High: 105, Low: 99, Close: 100}
	detected := DetectThreeBlackCrows(first, second, third)
	assert.True(t, detected)

	// Not three black crows - third candle closes higher
	third = OHLCData{Open: 104, High: 108, Low: 99, Close: 107}
	detected = DetectThreeBlackCrows(first, second, third)
	assert.False(t, detected)
}

func TestPatternValidation(t *testing.T) {
	t.Parallel()

	// Test with invalid OHLC data
	invalidOHLC := OHLCData{Open: 100, High: 95, Low: 105, Close: 98} // High < Low

	assert.False(t, DetectDoji(invalidOHLC, 0.01), "Should not detect doji with invalid OHLC")
	assert.False(t, DetectHammer(invalidOHLC, 2.0), "Should not detect hammer with invalid OHLC")
	assert.False(t, DetectShootingStar(invalidOHLC, 2.0), "Should not detect shooting star with invalid OHLC")

	// Test three-candle patterns with invalid data
	validOHLC := OHLCData{Open: 100, High: 110, Low: 95, Close: 105}
	opt := PatternDetectionOption{}

	assert.False(t, DetectMorningStar(invalidOHLC, validOHLC, validOHLC, opt), "Should not detect morning star with invalid first candle")
	assert.False(t, DetectMorningStar(validOHLC, invalidOHLC, validOHLC, opt), "Should not detect morning star with invalid second candle")
	assert.False(t, DetectMorningStar(validOHLC, validOHLC, invalidOHLC, opt), "Should not detect morning star with invalid third candle")

	assert.False(t, DetectEveningStar(invalidOHLC, validOHLC, validOHLC, opt), "Should not detect evening star with invalid first candle")
	assert.False(t, DetectEveningStar(validOHLC, invalidOHLC, validOHLC, opt), "Should not detect evening star with invalid second candle")
	assert.False(t, DetectEveningStar(validOHLC, validOHLC, invalidOHLC, opt), "Should not detect evening star with invalid third candle")

	// Test two-candle patterns with invalid data
	bullish, bearish := DetectHarami(invalidOHLC, validOHLC, 0.3)
	assert.False(t, bullish && bearish, "Should not detect harami with invalid first candle")

	bullish, bearish = DetectHarami(validOHLC, invalidOHLC, 0.3)
	assert.False(t, bullish && bearish, "Should not detect harami with invalid second candle")
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

	series := CandlestickSeries{Data: data}
	patterns := scanCandlestickPatterns(series, PatternDetectionOption{
		DojiThreshold:    0.01,
		ShadowRatio:      2.0,
		EngulfingMinSize: 0.8,
	})

	// Verify specific patterns were detected
	patternsByIndex := make(map[int][]string)
	for _, pattern := range patterns {
		patternsByIndex[pattern.Index] = append(patternsByIndex[pattern.Index], pattern.PatternType)
	}

	// Check expected patterns
	assert.Contains(t, patternsByIndex[1], PatternDoji)
	assert.Contains(t, patternsByIndex[2], PatternHammer)
	assert.Contains(t, patternsByIndex[3], PatternShootingStar)
	assert.Contains(t, patternsByIndex[4], PatternGravestone)
	assert.Contains(t, patternsByIndex[5], PatternDragonfly)
	assert.Contains(t, patternsByIndex[8], PatternMorningStar)
	assert.Contains(t, patternsByIndex[11], PatternEveningStar)
	assert.Contains(t, patternsByIndex[12], PatternMarubozuBull)
	assert.Contains(t, patternsByIndex[13], PatternMarubozuBear)
	assert.Contains(t, patternsByIndex[14], PatternSpinningTop)
	assert.Contains(t, patternsByIndex[16], PatternPiercingLine)
	assert.Contains(t, patternsByIndex[18], PatternDarkCloudCover)
	assert.Contains(t, patternsByIndex[20], PatternTweezerBottom)
	assert.Contains(t, patternsByIndex[23], PatternThreeWhiteSoldiers)
	assert.Contains(t, patternsByIndex[26], PatternThreeBlackCrows)

	// Verify we found multiple different pattern types
	uniquePatterns := make(map[string]bool)
	for _, pattern := range patterns {
		uniquePatterns[pattern.PatternType] = true
	}
	assert.Len(t, uniquePatterns, 19)
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
			DetectionOptions:   DefaultPatternOptions(),
		}, false},
		{"doji_only_replace_mode", data, &CandlestickPatternConfig{
			ReplaceSeriesLabel: true,
			EnabledPatterns:    []string{PatternDoji},
			DetectionOptions:   DefaultPatternOptions(),
		}, false},
		{"multiple_patterns_complement_mode", data, &CandlestickPatternConfig{
			ReplaceSeriesLabel: false,
			EnabledPatterns:    []string{PatternDoji, PatternShootingStar},
			DetectionOptions:   DefaultPatternOptions(),
		}, false},
		{"all_patterns_enabled", data, PatternsAll(), false},
		{"important_patterns_only", data, PatternsImportant(), false},
		// Edge cases
		{"empty_data", []OHLCData{}, &CandlestickPatternConfig{
			ReplaceSeriesLabel: true,
			EnabledPatterns:    []string{PatternDoji},
			DetectionOptions:   DefaultPatternOptions(),
		}, true},
		{"invalid_ohlc_data", []OHLCData{
			{Open: 0, High: 0, Low: 0, Close: 0},
			{Open: 100, High: 90, Low: 110, Close: 105},
		}, &CandlestickPatternConfig{
			ReplaceSeriesLabel: true,
			EnabledPatterns:    []string{PatternDoji},
			DetectionOptions:   DefaultPatternOptions(),
		}, false},
		{"nil_enabled_patterns", data, &CandlestickPatternConfig{
			ReplaceSeriesLabel: true,
			EnabledPatterns:    nil,
			DetectionOptions:   DefaultPatternOptions(),
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
				assert.Contains(t, config.EnabledPatterns, PatternDoji)
				assert.Contains(t, config.EnabledPatterns, PatternHammer)
				assert.GreaterOrEqual(t, len(config.EnabledPatterns), 10)
			},
		},
		{
			name:   "patterns_important",
			config: PatternsImportant(),
			validate: func(t *testing.T, config *CandlestickPatternConfig) {
				assert.True(t, config.ReplaceSeriesLabel)
				assert.Contains(t, config.EnabledPatterns, PatternEngulfingBull)
				assert.Contains(t, config.EnabledPatterns, PatternHammer)
				assert.LessOrEqual(t, len(config.EnabledPatterns), 8)
			},
		},
		{
			name:   "patterns_bullish",
			config: PatternsBullish(),
			validate: func(t *testing.T, config *CandlestickPatternConfig) {
				assert.True(t, config.ReplaceSeriesLabel)
				assert.Contains(t, config.EnabledPatterns, PatternHammer)
				assert.NotContains(t, config.EnabledPatterns, PatternShootingStar)
			},
		},
		{
			name:   "enable_patterns_custom",
			config: EnablePatterns(PatternDoji, PatternHammer),
			validate: func(t *testing.T, config *CandlestickPatternConfig) {
				assert.True(t, config.ReplaceSeriesLabel)
				assert.Len(t, config.EnabledPatterns, 2)
				assert.Contains(t, config.EnabledPatterns, PatternDoji)
				assert.Contains(t, config.EnabledPatterns, PatternHammer)
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
					{Open: 100, High: 110, Low: 95, Close: 105},     // Normal candle
					{Open: 105, High: 108, Low: 102, Close: 105.05}, // Doji pattern
					{Open: 105, High: 112, Low: 98, Close: 108},     // Normal candle
				}
				series := newCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">112.85</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110.87</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">108.88</text><text x=\"17\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">106.9</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">104.92</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">102.93</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100.95</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">98.97</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">96.98</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 308 569\nL 308 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 549 569\nL 549 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"183\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"665\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path  d=\"M 187 99\nL 187 254\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 187 409\nL 187 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 99\nL 235 99\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 564\nL 235 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 254\nL 283 254\nL 283 409\nL 91 409\nL 91 254\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 428 161\nL 428 253\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 428 254\nL 428 347\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 161\nL 476 161\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 347\nL 476 347\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 332 253\nL 524 253\nL 524 254\nL 332 254\nL 332 253\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 669 37\nL 669 161\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 669 254\nL 669 471\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 37\nL 717 37\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 471\nL 717 471\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 573 161\nL 765 161\nL 765 254\nL 573 254\nL 573 161\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 234\nL 519 234\nL 519 234\nA 4 4 90.00 0 1 523 238\nL 523 264\nL 523 264\nA 4 4 90.00 0 1 519 268\nL 433 268\nL 433 268\nA 4 4 90.00 0 1 429 264\nL 429 238\nL 429 238\nA 4 4 90.00 0 1 433 234\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"459\" y=\"251\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text><text x=\"433\" y=\"264\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 674 148\nL 760 148\nL 760 148\nA 4 4 90.00 0 1 764 152\nL 764 165\nL 764 165\nA 4 4 90.00 0 1 760 169\nL 674 169\nL 674 169\nA 4 4 90.00 0 1 670 165\nL 670 152\nL 670 152\nA 4 4 90.00 0 1 674 148\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"674\" y=\"165\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text></svg>",
			pngCRC: 0x60bae992,
		},
		{
			name: "hammer",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 108, High: 109, Low: 98, Close: 107},  // Hammer pattern
					{Open: 107, High: 112, Low: 102, Close: 110}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">112.85</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110.87</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">108.88</text><text x=\"17\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">106.9</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">104.92</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">102.93</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100.95</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">98.97</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">96.98</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 308 569\nL 308 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 549 569\nL 549 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"183\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"665\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path  d=\"M 187 99\nL 187 254\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 187 409\nL 187 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 99\nL 235 99\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 564\nL 235 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 254\nL 283 254\nL 283 409\nL 91 409\nL 91 254\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 428 130\nL 428 161\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 428 192\nL 428 471\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 380 130\nL 476 130\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 380 471\nL 476 471\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 332 161\nL 524 161\nL 524 192\nL 332 192\nL 332 161\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 669 37\nL 669 99\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 669 192\nL 669 347\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 37\nL 717 37\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 347\nL 717 347\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 573 99\nL 765 99\nL 765 192\nL 573 192\nL 573 99\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 173\nL 519 173\nL 519 173\nA 4 4 90.00 0 1 523 177\nL 523 203\nL 523 203\nA 4 4 90.00 0 1 519 207\nL 433 207\nL 433 207\nA 4 4 90.00 0 1 429 203\nL 429 177\nL 429 177\nA 4 4 90.00 0 1 433 173\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"446\" y=\"190\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><text x=\"433\" y=\"203\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text></svg>",
			pngCRC: 0x1c947c4b,
		},
		{
			name: "inverted_hammer",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105}, // Normal candle
					{Open: 95, High: 107, Low: 94, Close: 96},   // Inverted hammer pattern
					{Open: 96, High: 102, Low: 91, Close: 98},   // Normal candle
				}
				series := newCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">111</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">108.67</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">106.33</text><text x=\"30\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">104</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">101.67</text><text x=\"17\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">99.33</text><text x=\"39\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">97</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94.67</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">92.33</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 308 569\nL 308 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 549 569\nL 549 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"183\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"665\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path  d=\"M 187 37\nL 187 169\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 187 301\nL 187 433\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 37\nL 235 37\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 433\nL 235 433\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 169\nL 283 169\nL 283 301\nL 91 301\nL 91 169\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 428 116\nL 428 406\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 428 433\nL 428 459\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 116\nL 476 116\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 459\nL 476 459\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 332 406\nL 524 406\nL 524 433\nL 332 433\nL 332 406\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 669 248\nL 669 353\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 669 406\nL 669 538\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 248\nL 717 248\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 538\nL 717 538\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 573 353\nL 765 353\nL 765 406\nL 573 406\nL 573 353\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 380\nL 525 380\nL 525 380\nA 4 4 90.00 0 1 529 384\nL 529 423\nL 529 423\nA 4 4 90.00 0 1 525 427\nL 433 427\nL 433 427\nA 4 4 90.00 0 1 429 423\nL 429 384\nL 429 384\nA 4 4 90.00 0 1 433 380\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"437\" y=\"397\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text><text x=\"433\" y=\"410\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"436\" y=\"423\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 674 340\nL 760 340\nL 760 340\nA 4 4 90.00 0 1 764 344\nL 764 357\nL 764 357\nA 4 4 90.00 0 1 760 361\nL 674 361\nL 674 361\nA 4 4 90.00 0 1 670 357\nL 670 344\nL 670 344\nA 4 4 90.00 0 1 674 340\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"674\" y=\"357\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text></svg>",
			pngCRC: 0x81a74b66,
		},
		{
			name: "shooting_star",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 106, High: 125, Low: 105, Close: 107}, // Shooting star pattern
					{Open: 107, High: 112, Low: 102, Close: 109}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125.56</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">121.11</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">116.67</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">112.22</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">107.78</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">103.33</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">98.89</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94.44</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 308 569\nL 308 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 549 569\nL 549 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"183\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"665\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path  d=\"M 187 287\nL 187 357\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 187 426\nL 187 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 287\nL 235 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 495\nL 235 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 357\nL 283 357\nL 283 426\nL 91 426\nL 91 357\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 428 80\nL 428 329\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 428 343\nL 428 357\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 80\nL 476 80\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 357\nL 476 357\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 332 329\nL 524 329\nL 524 343\nL 332 343\nL 332 329\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 669 260\nL 669 301\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 669 329\nL 669 398\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 260\nL 717 260\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 398\nL 717 398\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 573 301\nL 765 301\nL 765 329\nL 573 329\nL 573 301\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 303\nL 525 303\nL 525 303\nA 4 4 90.00 0 1 529 307\nL 529 346\nL 529 346\nA 4 4 90.00 0 1 525 350\nL 433 350\nL 433 350\nA 4 4 90.00 0 1 429 346\nL 429 307\nL 429 307\nA 4 4 90.00 0 1 433 303\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"437\" y=\"320\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text><text x=\"433\" y=\"333\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"436\" y=\"346\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 674 288\nL 760 288\nL 760 288\nA 4 4 90.00 0 1 764 292\nL 764 305\nL 764 305\nA 4 4 90.00 0 1 760 309\nL 674 309\nL 674 309\nA 4 4 90.00 0 1 670 305\nL 670 292\nL 670 292\nA 4 4 90.00 0 1 674 288\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"674\" y=\"305\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text></svg>",
			pngCRC: 0x9fdbbed1,
		},
		{
			name: "gravestone_doji",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},    // Normal candle
					{Open: 108, High: 120, Low: 107, Close: 108.1}, // Gravestone doji pattern
					{Open: 108, High: 115, Low: 103, Close: 110},   // Normal candle
				}
				series := newCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">121.11</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">117.22</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">113.33</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">109.44</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105.56</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">101.67</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">97.78</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">93.89</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 308 569\nL 308 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 549 569\nL 549 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"183\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"665\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path  d=\"M 187 248\nL 187 327\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 187 406\nL 187 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 248\nL 235 248\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 485\nL 235 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 327\nL 283 327\nL 283 406\nL 91 406\nL 91 327\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 428 90\nL 428 278\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 428 280\nL 428 295\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 90\nL 476 90\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 295\nL 476 295\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 332 278\nL 524 278\nL 524 280\nL 332 280\nL 332 278\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 669 169\nL 669 248\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 669 280\nL 669 359\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 169\nL 717 169\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 359\nL 717 359\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 573 248\nL 765 248\nL 765 280\nL 573 280\nL 573 248\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 239\nL 525 239\nL 525 239\nA 4 4 90.00 0 1 529 243\nL 529 308\nL 529 308\nA 4 4 90.00 0 1 525 312\nL 433 312\nL 433 312\nA 4 4 90.00 0 1 429 308\nL 429 243\nL 429 243\nA 4 4 90.00 0 1 433 239\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"462\" y=\"256\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text><text x=\"437\" y=\"269\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text><text x=\"433\" y=\"282\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"441\" y=\"295\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">† Gravestone</text><text x=\"436\" y=\"308\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 674 235\nL 760 235\nL 760 235\nA 4 4 90.00 0 1 764 239\nL 764 252\nL 764 252\nA 4 4 90.00 0 1 760 256\nL 674 256\nL 674 256\nA 4 4 90.00 0 1 670 252\nL 670 239\nL 670 239\nA 4 4 90.00 0 1 674 235\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"674\" y=\"252\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text></svg>",
			pngCRC: 0x79b87d2d,
		},
		{
			name: "dragonfly_doji",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},   // Normal candle
					{Open: 109, High: 110, Low: 90, Close: 108.9}, // Dragonfly doji pattern
					{Open: 109, High: 115, Low: 104, Close: 112},  // Normal candle
				}
				series := newCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">116.25</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">113.33</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110.42</text><text x=\"17\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">107.5</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">104.58</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">101.67</text><text x=\"17\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">98.75</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95.83</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">92.92</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 308 569\nL 308 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 549 569\nL 549 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"183\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"665\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path  d=\"M 187 142\nL 187 248\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 187 353\nL 187 459\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 142\nL 235 142\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 459\nL 235 459\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 248\nL 283 248\nL 283 353\nL 91 353\nL 91 248\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 428 142\nL 428 164\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 428 166\nL 428 564\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 380 142\nL 476 142\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 380 564\nL 476 564\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 332 164\nL 524 164\nL 524 166\nL 332 166\nL 332 164\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 669 37\nL 669 100\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 669 164\nL 669 269\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 37\nL 717 37\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 269\nL 717 269\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 573 100\nL 765 100\nL 765 164\nL 573 164\nL 573 100\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 127\nL 519 127\nL 519 127\nA 4 4 90.00 0 1 523 131\nL 523 196\nL 523 196\nA 4 4 90.00 0 1 519 200\nL 433 200\nL 433 200\nA 4 4 90.00 0 1 429 196\nL 429 131\nL 429 131\nA 4 4 90.00 0 1 433 127\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"459\" y=\"144\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text><text x=\"446\" y=\"157\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><text x=\"442\" y=\"170\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ψ Dragonfly</text><text x=\"433\" y=\"183\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"434\" y=\"196\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">‖ Tweezer Top</text><path  d=\"M 674 87\nL 760 87\nL 760 87\nA 4 4 90.00 0 1 764 91\nL 764 104\nL 764 104\nA 4 4 90.00 0 1 760 108\nL 674 108\nL 674 108\nA 4 4 90.00 0 1 670 104\nL 670 91\nL 670 91\nA 4 4 90.00 0 1 674 87\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"674\" y=\"104\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text></svg>",
			pngCRC: 0x2f0d45cc,
		},
		{
			name: "bullish_marubozu",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 100, High: 120, Low: 100, Close: 120}, // Bullish marubozu pattern
					{Open: 120, High: 125, Low: 115, Close: 122}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125.56</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">121.11</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">116.67</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">112.22</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">107.78</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">103.33</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">98.89</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94.44</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 308 569\nL 308 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 549 569\nL 549 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"183\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"665\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path  d=\"M 187 287\nL 187 357\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 187 426\nL 187 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 287\nL 235 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 495\nL 235 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 357\nL 283 357\nL 283 426\nL 91 426\nL 91 357\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 380 149\nL 476 149\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 426\nL 476 426\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 332 149\nL 524 149\nL 524 426\nL 332 426\nL 332 149\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 669 80\nL 669 121\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 669 149\nL 669 218\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 80\nL 717 80\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 218\nL 717 218\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 573 121\nL 765 121\nL 765 149\nL 573 149\nL 573 121\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 136\nL 525 136\nL 525 136\nA 4 4 90.00 0 1 529 140\nL 529 153\nL 529 153\nA 4 4 90.00 0 1 525 157\nL 433 157\nL 433 157\nA 4 4 90.00 0 1 429 153\nL 429 140\nL 429 140\nA 4 4 90.00 0 1 433 136\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"433\" y=\"153\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">^ Bull Marubozu</text><path  d=\"M 674 108\nL 760 108\nL 760 108\nA 4 4 90.00 0 1 764 112\nL 764 125\nL 764 125\nA 4 4 90.00 0 1 760 129\nL 674 129\nL 674 129\nA 4 4 90.00 0 1 670 125\nL 670 112\nL 670 112\nA 4 4 90.00 0 1 674 108\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"674\" y=\"125\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text></svg>",
			pngCRC: 0x203eb157,
		},
		{
			name: "bearish_marubozu",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 120, High: 120, Low: 100, Close: 100}, // Bearish marubozu pattern
					{Open: 100, High: 105, Low: 95, Close: 102},  // Normal candle
				}
				series := newCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">121.11</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">117.22</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">113.33</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">109.44</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105.56</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">101.67</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">97.78</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">93.89</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 308 569\nL 308 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 549 569\nL 549 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"183\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"665\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path  d=\"M 187 248\nL 187 327\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 187 406\nL 187 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 248\nL 235 248\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 485\nL 235 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 327\nL 283 327\nL 283 406\nL 91 406\nL 91 327\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 380 90\nL 476 90\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 380 406\nL 476 406\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 332 90\nL 524 90\nL 524 406\nL 332 406\nL 332 90\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 669 327\nL 669 375\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 669 406\nL 669 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 327\nL 717 327\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 485\nL 717 485\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 573 375\nL 765 375\nL 765 406\nL 573 406\nL 573 375\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 393\nL 530 393\nL 530 393\nA 4 4 90.00 0 1 534 397\nL 534 410\nL 534 410\nA 4 4 90.00 0 1 530 414\nL 433 414\nL 433 414\nA 4 4 90.00 0 1 429 410\nL 429 397\nL 429 397\nA 4 4 90.00 0 1 433 393\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"433\" y=\"410\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">v Bear Marubozu</text><path  d=\"M 674 356\nL 760 356\nL 760 356\nA 4 4 90.00 0 1 764 360\nL 764 386\nL 764 386\nA 4 4 90.00 0 1 760 390\nL 674 390\nL 674 390\nA 4 4 90.00 0 1 670 386\nL 670 360\nL 670 360\nA 4 4 90.00 0 1 674 356\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"674\" y=\"373\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"679\" y=\"386\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ʘ Bull Harami</text></svg>",
			pngCRC: 0x8d50df32,
		},
		{
			name: "spinning_top",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 110, High: 125, Low: 95, Close: 112},  // Spinning top pattern
					{Open: 112, High: 118, Low: 107, Close: 115}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125.56</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">121.11</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">116.67</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">112.22</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">107.78</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">103.33</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">98.89</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94.44</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 308 569\nL 308 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 549 569\nL 549 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"183\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"665\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><path  d=\"M 187 287\nL 187 357\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 187 426\nL 187 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 287\nL 235 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 495\nL 235 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 357\nL 283 357\nL 283 426\nL 91 426\nL 91 357\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 428 80\nL 428 260\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 428 287\nL 428 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 80\nL 476 80\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 380 495\nL 476 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 332 260\nL 524 260\nL 524 287\nL 332 287\nL 332 260\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 669 177\nL 669 218\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 669 260\nL 669 329\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 177\nL 717 177\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 621 329\nL 717 329\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 573 218\nL 765 218\nL 765 260\nL 573 260\nL 573 218\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 247\nL 519 247\nL 519 247\nA 4 4 90.00 0 1 523 251\nL 523 264\nL 523 264\nA 4 4 90.00 0 1 519 268\nL 433 268\nL 433 268\nA 4 4 90.00 0 1 429 264\nL 429 251\nL 429 251\nA 4 4 90.00 0 1 433 247\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"433\" y=\"264\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 674 205\nL 760 205\nL 760 205\nA 4 4 90.00 0 1 764 209\nL 764 222\nL 764 222\nA 4 4 90.00 0 1 760 226\nL 674 226\nL 674 226\nA 4 4 90.00 0 1 670 222\nL 670 209\nL 670 209\nA 4 4 90.00 0 1 674 205\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"674\" y=\"222\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text></svg>",
			pngCRC: 0x55e33229,
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
				series := newCandlestickWithPatterns(data, PatternDetectionOption{
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
				series := newCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">116</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">113.67</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">111.33</text><text x=\"30\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">109</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">106.67</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">104.33</text><text x=\"30\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">102</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">99.67</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">97.33</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 247 569\nL 247 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 428 569\nL 428 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 609 569\nL 609 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"153\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"333\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"514\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"695\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><path  d=\"M 157 169\nL 157 301\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 157 433\nL 157 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 121 169\nL 193 169\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 121 564\nL 193 564\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 85 301\nL 229 301\nL 229 433\nL 85 433\nL 85 301\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 337 116\nL 337 169\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 337 274\nL 337 301\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 301 116\nL 373 116\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 301 301\nL 373 301\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 265 169\nL 409 169\nL 409 274\nL 265 274\nL 265 169\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 518 37\nL 518 63\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 518 327\nL 518 353\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 482 37\nL 554 37\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 482 353\nL 554 353\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 446 63\nL 590 63\nL 590 327\nL 446 327\nL 446 63\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 699 222\nL 699 327\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 699 380\nL 699 433\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 663 222\nL 735 222\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 663 433\nL 735 433\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 627 327\nL 771 327\nL 771 380\nL 627 380\nL 627 327\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 523 314\nL 618 314\nL 618 314\nA 4 4 90.00 0 1 622 318\nL 622 331\nL 622 331\nA 4 4 90.00 0 1 618 335\nL 523 335\nL 523 335\nA 4 4 90.00 0 1 519 331\nL 519 318\nL 519 318\nA 4 4 90.00 0 1 523 314\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"523\" y=\"331\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">V Bear Engulfing</text><path  d=\"M 704 367\nL 790 367\nL 790 367\nA 4 4 90.00 0 1 794 371\nL 794 384\nL 794 384\nA 4 4 90.00 0 1 790 388\nL 704 388\nL 704 388\nA 4 4 90.00 0 1 700 384\nL 700 371\nL 700 371\nA 4 4 90.00 0 1 704 367\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"704\" y=\"384\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text></svg>",
			pngCRC: 0x32999c81,
		},
		{
			name: "morning_star",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 120, High: 125, Low: 105, Close: 108}, // Large bearish
					{Open: 102, High: 104, Low: 100, Close: 103}, // Small body, gap down
					{Open: 108, High: 125, Low: 106, Close: 122}, // Large bullish, gap up
					{Open: 122, High: 128, Low: 120, Close: 125}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">133</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">128.22</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">123.44</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">118.67</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">113.89</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">109.11</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">104.33</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">99.56</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94.78</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 211 569\nL 211 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 356 569\nL 356 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 500 569\nL 500 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 645 569\nL 645 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"135\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"279\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"568\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><text x=\"713\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">5</text><path  d=\"M 139 307\nL 139 371\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 436\nL 139 500\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 111 307\nL 167 307\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 111 500\nL 167 500\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 82 371\nL 196 371\nL 196 436\nL 82 436\nL 82 371\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 283 114\nL 283 178\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 283 333\nL 283 371\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 255 114\nL 311 114\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 255 371\nL 311 371\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 226 178\nL 340 178\nL 340 333\nL 226 333\nL 226 178\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 428 384\nL 428 397\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 428 410\nL 428 436\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 400 384\nL 456 384\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 400 436\nL 456 436\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 371 397\nL 485 397\nL 485 410\nL 371 410\nL 371 397\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 572 114\nL 572 152\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 572 333\nL 572 358\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 544 114\nL 600 114\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 544 358\nL 600 358\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 515 152\nL 629 152\nL 629 333\nL 515 333\nL 515 152\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 717 75\nL 717 114\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 717 152\nL 717 178\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 689 75\nL 745 75\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 689 178\nL 745 178\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 660 114\nL 774 114\nL 774 152\nL 660 152\nL 660 114\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 384\nL 519 384\nL 519 384\nA 4 4 90.00 0 1 523 388\nL 523 401\nL 523 401\nA 4 4 90.00 0 1 519 405\nL 433 405\nL 433 405\nA 4 4 90.00 0 1 429 401\nL 429 388\nL 429 388\nA 4 4 90.00 0 1 433 384\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"433\" y=\"401\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 577 139\nL 661 139\nL 661 139\nA 4 4 90.00 0 1 665 143\nL 665 156\nL 665 156\nA 4 4 90.00 0 1 661 160\nL 577 160\nL 577 160\nA 4 4 90.00 0 1 573 156\nL 573 143\nL 573 143\nA 4 4 90.00 0 1 577 139\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"577\" y=\"156\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">* Morning Star</text><path  d=\"M 701 101\nL 800 101\nL 800 101\nA 4 4 90.00 0 1 804 105\nL 804 118\nL 804 118\nA 4 4 90.00 0 1 800 122\nL 701 122\nL 701 122\nA 4 4 90.00 0 1 697 118\nL 697 105\nL 697 105\nA 4 4 90.00 0 1 701 101\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"701\" y=\"118\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ш Three Soldiers</text></svg>",
			pngCRC: 0x1769c10d,
		},
		{
			name: "evening_star",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 122, High: 140, Low: 120, Close: 138}, // Large bullish
					{Open: 142, High: 144, Low: 140, Close: 143}, // Small body, gap up
					{Open: 138, High: 140, Low: 115, Close: 118}, // Large bearish, gap down
					{Open: 118, High: 122, Low: 115, Close: 120}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">149</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">142.44</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">135.89</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">129.33</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">122.78</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">116.22</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">109.67</text><text x=\"9\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">103.11</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">96.56</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 211 569\nL 211 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 356 569\nL 356 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 500 569\nL 500 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 645 569\nL 645 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"135\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"279\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"568\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><text x=\"713\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">5</text><path  d=\"M 139 377\nL 139 424\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 471\nL 139 518\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 111 377\nL 167 377\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 111 518\nL 167 518\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 82 424\nL 196 424\nL 196 471\nL 82 471\nL 82 424\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 283 95\nL 283 114\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 283 264\nL 283 283\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 255 95\nL 311 95\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 255 283\nL 311 283\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 226 114\nL 340 114\nL 340 264\nL 226 264\nL 226 114\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 428 57\nL 428 67\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 428 76\nL 428 95\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 400 57\nL 456 57\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 400 95\nL 456 95\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 371 67\nL 485 67\nL 485 76\nL 371 76\nL 371 67\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 572 95\nL 572 114\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 572 302\nL 572 330\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 544 95\nL 600 95\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 544 330\nL 600 330\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 515 114\nL 629 114\nL 629 302\nL 515 302\nL 515 114\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 717 264\nL 717 283\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 717 302\nL 717 330\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 689 264\nL 745 264\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 689 330\nL 745 330\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 660 283\nL 774 283\nL 774 302\nL 660 302\nL 660 283\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 433 54\nL 519 54\nL 519 54\nA 4 4 90.00 0 1 523 58\nL 523 71\nL 523 71\nA 4 4 90.00 0 1 519 75\nL 433 75\nL 433 75\nA 4 4 90.00 0 1 429 71\nL 429 58\nL 429 58\nA 4 4 90.00 0 1 433 54\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"433\" y=\"71\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 577 289\nL 659 289\nL 659 289\nA 4 4 90.00 0 1 663 293\nL 663 306\nL 663 306\nA 4 4 90.00 0 1 659 310\nL 577 310\nL 577 310\nA 4 4 90.00 0 1 573 306\nL 573 293\nL 573 293\nA 4 4 90.00 0 1 577 289\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"577\" y=\"306\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">⁎ Evening Star</text><path  d=\"M 697 257\nL 800 257\nL 800 257\nA 4 4 90.00 0 1 804 261\nL 804 300\nL 804 300\nA 4 4 90.00 0 1 800 304\nL 697 304\nL 697 304\nA 4 4 90.00 0 1 693 300\nL 693 261\nL 693 261\nA 4 4 90.00 0 1 697 257\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"705\" y=\"274\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><text x=\"710\" y=\"287\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ʘ Bull Harami</text><text x=\"697\" y=\"300\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ǁ Tweezer Bottom</text></svg>",
			pngCRC: 0x6f01a83c,
		},
		{
			name: "piercing_line",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 120, High: 121, Low: 115, Close: 115}, // Bearish candle
					{Open: 112, High: 119, Low: 112, Close: 118}, // Piercing line (opens below prev low, closes above midpoint)
					{Open: 118, High: 125, Low: 116, Close: 122}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125.56</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">121.11</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">116.67</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">112.22</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">107.78</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">103.33</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">98.89</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94.44</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 247 569\nL 247 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 428 569\nL 428 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 609 569\nL 609 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"153\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"333\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"514\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"695\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><path  d=\"M 157 287\nL 157 357\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 157 426\nL 157 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 121 287\nL 193 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 121 495\nL 193 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 85 357\nL 229 357\nL 229 426\nL 85 426\nL 85 357\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 337 135\nL 337 149\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 301 135\nL 373 135\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 301 218\nL 373 218\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 265 149\nL 409 149\nL 409 218\nL 265 218\nL 265 149\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 518 163\nL 518 177\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 482 163\nL 554 163\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 482 260\nL 554 260\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 446 177\nL 590 177\nL 590 260\nL 446 260\nL 446 177\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 699 80\nL 699 121\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 699 177\nL 699 204\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 663 80\nL 735 80\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 663 204\nL 735 204\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 627 121\nL 771 121\nL 771 177\nL 627 177\nL 627 121\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 523 164\nL 604 164\nL 604 164\nA 4 4 90.00 0 1 608 168\nL 608 181\nL 608 181\nA 4 4 90.00 0 1 604 185\nL 523 185\nL 523 185\nA 4 4 90.00 0 1 519 181\nL 519 168\nL 519 168\nA 4 4 90.00 0 1 523 164\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"523\" y=\"181\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">| Piercing Line</text></svg>",
			pngCRC: 0x924811cf,
		},
		{
			name: "dark_cloud_cover",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 118, High: 125, Low: 118, Close: 125}, // Bullish candle
					{Open: 127, High: 127, Low: 120, Close: 121}, // Dark cloud cover (opens above prev high, closes below midpoint)
					{Open: 121, High: 124, Low: 118, Close: 120}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">132</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">127.33</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">122.67</text><text x=\"30\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">118</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">113.33</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">108.67</text><text x=\"30\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">104</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">99.33</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94.67</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 247 569\nL 247 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 428 569\nL 428 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 609 569\nL 609 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"153\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"333\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"514\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"695\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><path  d=\"M 157 301\nL 157 367\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 157 433\nL 157 499\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 121 301\nL 193 301\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 121 499\nL 193 499\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 85 367\nL 229 367\nL 229 433\nL 85 433\nL 85 367\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 301 103\nL 373 103\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 301 195\nL 373 195\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 265 103\nL 409 103\nL 409 195\nL 265 195\nL 265 103\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 518 156\nL 518 169\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 482 76\nL 554 76\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 482 169\nL 554 169\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 446 76\nL 590 76\nL 590 156\nL 446 156\nL 446 76\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 699 116\nL 699 156\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 699 169\nL 699 195\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 663 116\nL 735 116\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 663 195\nL 735 195\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 627 156\nL 771 156\nL 771 169\nL 627 169\nL 627 156\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 342 90\nL 434 90\nL 434 90\nA 4 4 90.00 0 1 438 94\nL 438 107\nL 438 107\nA 4 4 90.00 0 1 434 111\nL 342 111\nL 342 111\nA 4 4 90.00 0 1 338 107\nL 338 94\nL 338 94\nA 4 4 90.00 0 1 342 90\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"342\" y=\"107\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">^ Bull Marubozu</text><path  d=\"M 523 143\nL 597 143\nL 597 143\nA 4 4 90.00 0 1 601 147\nL 601 160\nL 601 160\nA 4 4 90.00 0 1 597 164\nL 523 164\nL 523 164\nA 4 4 90.00 0 1 519 160\nL 519 147\nL 519 147\nA 4 4 90.00 0 1 523 143\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"523\" y=\"160\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ξ Dark Cloud</text><path  d=\"M 704 156\nL 790 156\nL 790 156\nA 4 4 90.00 0 1 794 160\nL 794 173\nL 794 173\nA 4 4 90.00 0 1 790 177\nL 704 177\nL 704 177\nA 4 4 90.00 0 1 700 173\nL 700 160\nL 700 160\nA 4 4 90.00 0 1 704 156\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"704\" y=\"173\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text></svg>",
			pngCRC: 0xeb647ac8,
		},
		{
			name: "three_white_soldiers",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 110, High: 115, Low: 109, Close: 114}, // First soldier
					{Open: 113, High: 118, Low: 112, Close: 117}, // Second soldier
					{Open: 116, High: 121, Low: 115, Close: 120}, // Third soldier
					{Open: 120, High: 125, Low: 118, Close: 123}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">130</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125.56</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">121.11</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">116.67</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">112.22</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">107.78</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">103.33</text><text x=\"17\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">98.89</text><text x=\"17\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94.44</text><text x=\"39\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 67 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 67 569\nL 67 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 211 569\nL 211 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 356 569\nL 356 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 500 569\nL 500 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 645 569\nL 645 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"135\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"279\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"424\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"568\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><text x=\"713\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">5</text><path  d=\"M 139 287\nL 139 357\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 426\nL 139 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 111 287\nL 167 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 111 495\nL 167 495\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 82 357\nL 196 357\nL 196 426\nL 82 426\nL 82 357\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 283 218\nL 283 232\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 283 287\nL 283 301\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 255 218\nL 311 218\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 255 301\nL 311 301\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 226 232\nL 340 232\nL 340 287\nL 226 287\nL 226 232\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 428 177\nL 428 191\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 428 246\nL 428 260\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 400 177\nL 456 177\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 400 260\nL 456 260\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 371 191\nL 485 191\nL 485 246\nL 371 246\nL 371 191\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 572 135\nL 572 149\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 572 204\nL 572 218\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 544 135\nL 600 135\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 544 218\nL 600 218\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 515 149\nL 629 149\nL 629 204\nL 515 204\nL 515 149\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 717 80\nL 717 107\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 717 149\nL 717 177\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 689 80\nL 745 80\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 689 177\nL 745 177\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 660 107\nL 774 107\nL 774 149\nL 660 149\nL 660 107\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 577 136\nL 676 136\nL 676 136\nA 4 4 90.00 0 1 680 140\nL 680 153\nL 680 153\nA 4 4 90.00 0 1 676 157\nL 577 157\nL 577 157\nA 4 4 90.00 0 1 573 153\nL 573 140\nL 573 140\nA 4 4 90.00 0 1 577 136\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"577\" y=\"153\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ш Three Soldiers</text></svg>",
			pngCRC: 0xd4ba85a0,
		},
		{
			name: "three_black_crows",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 120, High: 121, Low: 115, Close: 116}, // First crow
					{Open: 117, High: 118, Low: 112, Close: 113}, // Second crow
					{Open: 114, High: 115, Low: 108, Close: 109}, // Third crow
					{Open: 109, High: 112, Low: 106, Close: 108}, // Normal candle
				}
				series := newCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">126</text><text x=\"9\" y=\"77\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">122</text><text x=\"9\" y=\"138\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">118</text><text x=\"9\" y=\"200\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">114</text><text x=\"9\" y=\"261\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110</text><text x=\"9\" y=\"322\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">106</text><text x=\"9\" y=\"384\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">102</text><text x=\"18\" y=\"445\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">98</text><text x=\"18\" y=\"506\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94</text><text x=\"18\" y=\"568\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 71\nL 790 71\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 133\nL 790 133\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 194\nL 790 194\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 256\nL 790 256\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 317\nL 790 317\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 379\nL 790 379\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 440\nL 790 440\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 502\nL 790 502\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 46 564\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 46 569\nL 46 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 194 569\nL 194 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 343 569\nL 343 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 492 569\nL 492 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 641 569\nL 641 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 790 569\nL 790 564\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"116\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">1</text><text x=\"264\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">2</text><text x=\"413\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">3</text><text x=\"562\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">4</text><text x=\"711\" y=\"590\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">5</text><path  d=\"M 120 257\nL 120 334\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 120 411\nL 120 488\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 257\nL 149 257\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 91 488\nL 149 488\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 61 334\nL 179 334\nL 179 411\nL 61 411\nL 61 334\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 268 87\nL 268 103\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 268 164\nL 268 180\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 239 87\nL 297 87\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 239 180\nL 297 180\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 209 103\nL 327 103\nL 327 164\nL 209 164\nL 209 103\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 417 134\nL 417 149\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 417 211\nL 417 226\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 388 134\nL 446 134\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 388 226\nL 446 226\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 358 149\nL 476 149\nL 476 211\nL 358 211\nL 358 149\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 566 180\nL 566 195\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 566 272\nL 566 287\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 537 180\nL 595 180\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 537 287\nL 595 287\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 507 195\nL 625 195\nL 625 272\nL 507 272\nL 507 195\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 715 226\nL 715 272\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 715 287\nL 715 318\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 686 226\nL 744 226\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 686 318\nL 744 318\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 656 272\nL 774 272\nL 774 287\nL 656 287\nL 656 272\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 571 259\nL 658 259\nL 658 259\nA 4 4 90.00 0 1 662 263\nL 662 276\nL 662 276\nA 4 4 90.00 0 1 658 280\nL 571 280\nL 571 280\nA 4 4 90.00 0 1 567 276\nL 567 263\nL 567 263\nA 4 4 90.00 0 1 571 259\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"571\" y=\"276\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ω Three Crows</text><path  d=\"M 714 274\nL 800 274\nL 800 274\nA 4 4 90.00 0 1 804 278\nL 804 291\nL 804 291\nA 4 4 90.00 0 1 800 295\nL 714 295\nL 714 295\nA 4 4 90.00 0 1 710 291\nL 710 278\nL 710 278\nA 4 4 90.00 0 1 714 274\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"714\" y=\"291\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text></svg>",
			pngCRC: 0x815ce1e6,
		},
		{
			name: "doji_and_hammers",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},     // Normal candle
					{Open: 105, High: 108, Low: 102, Close: 105.05}, // Doji pattern
					{Open: 105, High: 107, Low: 95, Close: 106},     // Hammer pattern
					{Open: 106, High: 118, Low: 105, Close: 107},    // Shooting star pattern
					{Open: 107, High: 115, Low: 102, Close: 112},    // Normal candle
				}
				series := newCandlestickWithPatterns(data, PatternDetectionOption{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
				return CandlestickChartOption{
					XAxis:      XAxisOption{Show: Ptr(false)},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">119.15</text><text x=\"9\" y=\"80\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">116.47</text><text x=\"9\" y=\"144\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">113.78</text><text x=\"17\" y=\"208\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">111.1</text><text x=\"9\" y=\"272\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">108.42</text><text x=\"9\" y=\"337\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">105.73</text><text x=\"9\" y=\"401\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">103.05</text><text x=\"9\" y=\"465\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">100.37</text><text x=\"17\" y=\"529\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">97.68</text><text x=\"39\" y=\"594\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">95</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 74\nL 790 74\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 138\nL 790 138\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 203\nL 790 203\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 267\nL 790 267\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 332\nL 790 332\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 396\nL 790 396\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 461\nL 790 461\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 525\nL 790 525\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 139 230\nL 139 350\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 139 470\nL 139 590\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 111 230\nL 167 230\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 111 590\nL 167 590\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 82 350\nL 196 350\nL 196 470\nL 82 470\nL 82 350\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 283 278\nL 283 349\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 283 350\nL 283 422\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 255 278\nL 311 278\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 255 422\nL 311 422\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 226 349\nL 340 349\nL 340 350\nL 226 350\nL 226 349\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 428 302\nL 428 326\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 428 350\nL 428 590\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 400 302\nL 456 302\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 400 590\nL 456 590\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 371 326\nL 485 326\nL 485 350\nL 371 350\nL 371 326\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 572 38\nL 572 302\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 572 326\nL 572 350\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 544 38\nL 600 38\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 544 350\nL 600 350\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 515 302\nL 629 302\nL 629 326\nL 515 326\nL 515 302\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 717 110\nL 717 182\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 717 302\nL 717 422\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 689 110\nL 745 110\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 689 422\nL 745 422\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 660 182\nL 774 182\nL 774 302\nL 660 302\nL 660 182\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 288 330\nL 374 330\nL 374 330\nA 4 4 90.00 0 1 378 334\nL 378 360\nL 378 360\nA 4 4 90.00 0 1 374 364\nL 288 364\nL 288 364\nA 4 4 90.00 0 1 284 360\nL 284 334\nL 284 334\nA 4 4 90.00 0 1 288 330\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"314\" y=\"347\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text><text x=\"288\" y=\"360\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 433 307\nL 519 307\nL 519 307\nA 4 4 90.00 0 1 523 311\nL 523 337\nL 523 337\nA 4 4 90.00 0 1 519 341\nL 433 341\nL 433 341\nA 4 4 90.00 0 1 429 337\nL 429 311\nL 429 311\nA 4 4 90.00 0 1 433 307\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"446\" y=\"324\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><text x=\"433\" y=\"337\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 577 276\nL 669 276\nL 669 276\nA 4 4 90.00 0 1 673 280\nL 673 319\nL 673 319\nA 4 4 90.00 0 1 669 323\nL 577 323\nL 577 323\nA 4 4 90.00 0 1 573 319\nL 573 280\nL 573 280\nA 4 4 90.00 0 1 577 276\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"581\" y=\"293\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text><text x=\"577\" y=\"306\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"580\" y=\"319\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text></svg>",
			pngCRC: 0xbc076b66,
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
				series := newCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">133</text><text x=\"9\" y=\"80\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">128.22</text><text x=\"9\" y=\"144\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">123.44</text><text x=\"9\" y=\"208\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">118.67</text><text x=\"9\" y=\"272\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">113.89</text><text x=\"9\" y=\"337\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">109.11</text><text x=\"9\" y=\"401\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">104.33</text><text x=\"17\" y=\"465\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">99.56</text><text x=\"17\" y=\"529\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94.78</text><text x=\"39\" y=\"594\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 74\nL 790 74\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 138\nL 790 138\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 203\nL 790 203\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 267\nL 790 267\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 332\nL 790 332\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 396\nL 790 396\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 461\nL 790 461\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 525\nL 790 525\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 118 321\nL 118 388\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 118 456\nL 118 523\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 98 321\nL 138 321\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 98 523\nL 138 523\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 77 388\nL 159 388\nL 159 456\nL 77 456\nL 77 388\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 221 294\nL 221 321\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 221 375\nL 221 388\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 201 294\nL 241 294\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 201 388\nL 241 388\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 180 321\nL 262 321\nL 262 375\nL 180 375\nL 180 321\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 324 253\nL 324 267\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 324 402\nL 324 415\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 304 253\nL 344 253\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 304 415\nL 344 415\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 283 267\nL 365 267\nL 365 402\nL 283 402\nL 283 267\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 428 118\nL 428 186\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 428 348\nL 428 388\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 408 118\nL 448 118\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 408 388\nL 448 388\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 387 186\nL 469 186\nL 469 348\nL 387 348\nL 387 186\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 531 402\nL 531 415\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 531 429\nL 531 456\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 511 402\nL 551 402\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 511 456\nL 551 456\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 490 415\nL 572 415\nL 572 429\nL 490 429\nL 490 415\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 634 118\nL 634 159\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 634 348\nL 634 375\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 614 118\nL 654 118\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 614 375\nL 654 375\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 593 159\nL 675 159\nL 675 348\nL 593 348\nL 593 159\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 738 78\nL 738 118\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 738 159\nL 738 186\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 718 78\nL 758 78\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 718 186\nL 758 186\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 697 118\nL 779 118\nL 779 159\nL 697 159\nL 697 118\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 329 254\nL 420 254\nL 420 254\nA 4 4 90.00 0 1 424 258\nL 424 271\nL 424 271\nA 4 4 90.00 0 1 420 275\nL 329 275\nL 329 275\nA 4 4 90.00 0 1 325 271\nL 325 258\nL 325 258\nA 4 4 90.00 0 1 329 254\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"329\" y=\"271\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Λ Bull Engulfing</text><path  d=\"M 433 335\nL 507 335\nL 507 335\nA 4 4 90.00 0 1 511 339\nL 511 352\nL 511 352\nA 4 4 90.00 0 1 507 356\nL 433 356\nL 433 356\nA 4 4 90.00 0 1 429 352\nL 429 339\nL 429 339\nA 4 4 90.00 0 1 433 335\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"433\" y=\"352\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ξ Dark Cloud</text><path  d=\"M 536 402\nL 622 402\nL 622 402\nA 4 4 90.00 0 1 626 406\nL 626 419\nL 626 419\nA 4 4 90.00 0 1 622 423\nL 536 423\nL 536 423\nA 4 4 90.00 0 1 532 419\nL 532 406\nL 532 406\nA 4 4 90.00 0 1 536 402\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"536\" y=\"419\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 639 146\nL 723 146\nL 723 146\nA 4 4 90.00 0 1 727 150\nL 727 163\nL 727 163\nA 4 4 90.00 0 1 723 167\nL 639 167\nL 639 167\nA 4 4 90.00 0 1 635 163\nL 635 150\nL 635 150\nA 4 4 90.00 0 1 639 146\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"639\" y=\"163\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">* Morning Star</text><path  d=\"M 701 105\nL 800 105\nL 800 105\nA 4 4 90.00 0 1 804 109\nL 804 122\nL 804 122\nA 4 4 90.00 0 1 800 126\nL 701 126\nL 701 126\nA 4 4 90.00 0 1 697 122\nL 697 109\nL 697 109\nA 4 4 90.00 0 1 701 105\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"701\" y=\"122\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ш Three Soldiers</text></svg>",
			pngCRC: 0x5908621e,
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
				series := newCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">126.75</text><text x=\"9\" y=\"80\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">122.67</text><text x=\"9\" y=\"144\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">118.58</text><text x=\"17\" y=\"208\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">114.5</text><text x=\"9\" y=\"272\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">110.42</text><text x=\"9\" y=\"337\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">106.33</text><text x=\"9\" y=\"401\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">102.25</text><text x=\"17\" y=\"465\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">98.17</text><text x=\"17\" y=\"529\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94.08</text><text x=\"39\" y=\"594\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 74\nL 790 74\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 138\nL 790 138\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 203\nL 790 203\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 267\nL 790 267\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 332\nL 790 332\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 396\nL 790 396\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 461\nL 790 461\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 525\nL 790 525\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 97 275\nL 97 354\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 97 433\nL 97 512\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 85 275\nL 109 275\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 85 512\nL 109 512\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 73 354\nL 121 354\nL 121 433\nL 73 433\nL 73 354\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 157 306\nL 157 353\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 157 354\nL 157 401\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 145 306\nL 169 306\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 145 401\nL 169 401\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 133 353\nL 181 353\nL 181 354\nL 133 354\nL 133 353\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 217 322\nL 217 338\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 217 354\nL 217 512\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 205 322\nL 229 322\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 205 512\nL 229 512\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 193 338\nL 241 338\nL 241 354\nL 193 354\nL 193 338\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 277 38\nL 277 243\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 277 275\nL 277 512\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 265 38\nL 289 38\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 265 512\nL 289 512\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 253 243\nL 301 243\nL 301 275\nL 253 275\nL 253 243\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 326 117\nL 350 117\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 326 433\nL 350 433\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 314 117\nL 362 117\nL 362 433\nL 314 433\nL 314 117\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 386 117\nL 410 117\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 386 433\nL 410 433\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 374 117\nL 422 117\nL 422 433\nL 374 433\nL 374 117\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 458 243\nL 458 275\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 458 338\nL 458 354\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 446 243\nL 470 243\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 446 354\nL 470 354\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 434 275\nL 482 275\nL 482 338\nL 434 338\nL 434 275\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 518 196\nL 518 212\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 518 370\nL 518 385\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 506 196\nL 530 196\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 506 385\nL 530 385\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 494 212\nL 542 212\nL 542 370\nL 494 370\nL 494 212\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 579 38\nL 579 322\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 579 338\nL 579 354\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 567 38\nL 591 38\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 567 354\nL 591 354\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 555 322\nL 603 322\nL 603 338\nL 555 338\nL 555 322\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 639 275\nL 639 291\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 639 292\nL 639 590\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 627 275\nL 651 275\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 627 590\nL 651 590\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 615 291\nL 663 291\nL 663 292\nL 615 292\nL 615 291\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 699 117\nL 699 305\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 699 306\nL 699 322\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 687 117\nL 711 117\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 687 322\nL 711 322\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 675 305\nL 723 305\nL 723 306\nL 675 306\nL 675 305\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 759 196\nL 759 275\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 759 306\nL 759 385\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 747 196\nL 771 196\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 747 385\nL 771 385\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 735 275\nL 783 275\nL 783 306\nL 735 306\nL 735 275\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 162 334\nL 248 334\nL 248 334\nA 4 4 90.00 0 1 252 338\nL 252 364\nL 252 364\nA 4 4 90.00 0 1 248 368\nL 162 368\nL 162 368\nA 4 4 90.00 0 1 158 364\nL 158 338\nL 158 338\nA 4 4 90.00 0 1 162 334\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"188\" y=\"351\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text><text x=\"162\" y=\"364\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 222 319\nL 308 319\nL 308 319\nA 4 4 90.00 0 1 312 323\nL 312 349\nL 312 349\nA 4 4 90.00 0 1 308 353\nL 222 353\nL 222 353\nA 4 4 90.00 0 1 218 349\nL 218 323\nL 218 323\nA 4 4 90.00 0 1 222 319\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"235\" y=\"336\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><text x=\"222\" y=\"349\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 282 230\nL 368 230\nL 368 230\nA 4 4 90.00 0 1 372 234\nL 372 247\nL 372 247\nA 4 4 90.00 0 1 368 251\nL 282 251\nL 282 251\nA 4 4 90.00 0 1 278 247\nL 278 234\nL 278 234\nA 4 4 90.00 0 1 282 230\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"282\" y=\"247\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 343 104\nL 435 104\nL 435 104\nA 4 4 90.00 0 1 439 108\nL 439 121\nL 439 121\nA 4 4 90.00 0 1 435 125\nL 343 125\nL 343 125\nA 4 4 90.00 0 1 339 121\nL 339 108\nL 339 108\nA 4 4 90.00 0 1 343 104\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"343\" y=\"121\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">^ Bull Marubozu</text><path  d=\"M 403 414\nL 500 414\nL 500 414\nA 4 4 90.00 0 1 504 418\nL 504 444\nL 504 444\nA 4 4 90.00 0 1 500 448\nL 403 448\nL 403 448\nA 4 4 90.00 0 1 399 444\nL 399 418\nL 399 418\nA 4 4 90.00 0 1 403 414\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"403\" y=\"431\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">v Bear Marubozu</text><text x=\"409\" y=\"444\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">‖ Tweezer Top</text><path  d=\"M 523 199\nL 614 199\nL 614 199\nA 4 4 90.00 0 1 618 203\nL 618 216\nL 618 216\nA 4 4 90.00 0 1 614 220\nL 523 220\nL 523 220\nA 4 4 90.00 0 1 519 216\nL 519 203\nL 519 203\nA 4 4 90.00 0 1 523 199\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"523\" y=\"216\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Λ Bull Engulfing</text><path  d=\"M 584 296\nL 676 296\nL 676 296\nA 4 4 90.00 0 1 680 300\nL 680 339\nL 680 339\nA 4 4 90.00 0 1 676 343\nL 584 343\nL 584 343\nA 4 4 90.00 0 1 580 339\nL 580 300\nL 580 300\nA 4 4 90.00 0 1 584 296\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"588\" y=\"313\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text><text x=\"584\" y=\"326\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"587\" y=\"339\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 644 260\nL 730 260\nL 730 260\nA 4 4 90.00 0 1 734 264\nL 734 316\nL 734 316\nA 4 4 90.00 0 1 730 320\nL 644 320\nL 644 320\nA 4 4 90.00 0 1 640 316\nL 640 264\nL 640 264\nA 4 4 90.00 0 1 644 260\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"670\" y=\"277\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text><text x=\"657\" y=\"290\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><text x=\"653\" y=\"303\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ψ Dragonfly</text><text x=\"644\" y=\"316\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 704 266\nL 796 266\nL 796 266\nA 4 4 90.00 0 1 800 270\nL 800 335\nL 800 335\nA 4 4 90.00 0 1 796 339\nL 704 339\nL 704 339\nA 4 4 90.00 0 1 700 335\nL 700 270\nL 700 270\nA 4 4 90.00 0 1 704 266\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"733\" y=\"283\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text><text x=\"708\" y=\"296\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text><text x=\"704\" y=\"309\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"712\" y=\"322\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">† Gravestone</text><text x=\"707\" y=\"335\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 714 262\nL 800 262\nL 800 262\nA 4 4 90.00 0 1 804 266\nL 804 279\nL 804 279\nA 4 4 90.00 0 1 800 283\nL 714 283\nL 714 283\nA 4 4 90.00 0 1 710 279\nL 710 266\nL 710 266\nA 4 4 90.00 0 1 714 262\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"714\" y=\"279\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text></svg>",
			pngCRC: 0xa2133f4,
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
							PatternMorningStar,
							PatternEveningStar,
							PatternThreeWhiteSoldiers,
							PatternThreeBlackCrows,
						},
						DetectionOptions: PatternDetectionOption{
							DojiThreshold: 0.01,
							ShadowRatio:   2.0,
						},
					},
				}
				return CandlestickChartOption{
					XAxis:      XAxisOption{Show: Ptr(false)},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">149</text><text x=\"9\" y=\"80\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">142.44</text><text x=\"9\" y=\"144\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">135.89</text><text x=\"9\" y=\"208\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">129.33</text><text x=\"9\" y=\"272\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">122.78</text><text x=\"9\" y=\"337\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">116.22</text><text x=\"9\" y=\"401\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">109.67</text><text x=\"9\" y=\"465\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">103.11</text><text x=\"17\" y=\"529\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">96.56</text><text x=\"39\" y=\"594\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 74\nL 790 74\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 138\nL 790 138\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 203\nL 790 203\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 267\nL 790 267\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 332\nL 790 332\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 396\nL 790 396\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 461\nL 790 461\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 525\nL 790 525\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 94 394\nL 94 443\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 94 492\nL 94 541\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 83 394\nL 105 394\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 83 541\nL 105 541\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 72 443\nL 116 443\nL 116 492\nL 72 492\nL 72 443\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 150 246\nL 150 296\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 150 414\nL 150 443\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 139 246\nL 161 246\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 139 443\nL 161 443\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 128 296\nL 172 296\nL 172 414\nL 128 414\nL 128 296\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 205 453\nL 205 463\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 205 473\nL 205 492\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 194 453\nL 216 453\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 194 492\nL 216 492\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 183 463\nL 227 463\nL 227 473\nL 183 473\nL 183 463\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 261 246\nL 261 276\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 261 414\nL 261 433\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 250 246\nL 272 246\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 250 433\nL 272 433\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 239 276\nL 283 276\nL 283 414\nL 239 414\nL 239 276\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 317 345\nL 317 355\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 317 394\nL 317 404\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 306 345\nL 328 345\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 306 404\nL 328 404\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 295 355\nL 339 355\nL 339 394\nL 295 394\nL 295 355\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 372 315\nL 372 325\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 372 364\nL 372 374\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 361 315\nL 383 315\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 361 374\nL 383 374\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 350 325\nL 394 325\nL 394 364\nL 350 364\nL 350 325\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 428 286\nL 428 296\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 428 335\nL 428 345\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 417 286\nL 439 286\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 417 345\nL 439 345\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 406 296\nL 450 296\nL 450 335\nL 406 335\nL 406 296\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 483 99\nL 483 119\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 483 276\nL 483 296\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 472 99\nL 494 99\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 472 296\nL 494 296\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 461 119\nL 505 119\nL 505 276\nL 461 276\nL 461 119\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 539 60\nL 539 69\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 539 79\nL 539 99\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 528 60\nL 550 60\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 528 99\nL 550 99\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 517 69\nL 561 69\nL 561 79\nL 517 79\nL 517 69\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 595 99\nL 595 119\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 595 315\nL 595 345\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 584 99\nL 606 99\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 584 345\nL 606 345\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 573 119\nL 617 119\nL 617 315\nL 573 315\nL 573 119\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 650 286\nL 650 296\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 650 335\nL 650 345\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 639 286\nL 661 286\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 639 345\nL 661 345\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 628 296\nL 672 296\nL 672 335\nL 628 335\nL 628 296\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 706 315\nL 706 325\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 706 364\nL 706 374\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 695 315\nL 717 315\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 695 374\nL 717 374\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 684 325\nL 728 325\nL 728 364\nL 684 364\nL 684 325\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 762 178\nL 762 217\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 762 364\nL 762 433\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 751 178\nL 773 178\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 751 433\nL 773 433\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 740 217\nL 784 217\nL 784 364\nL 740 364\nL 740 217\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 266 263\nL 350 263\nL 350 263\nA 4 4 90.00 0 1 354 267\nL 354 280\nL 354 280\nA 4 4 90.00 0 1 350 284\nL 266 284\nL 266 284\nA 4 4 90.00 0 1 262 280\nL 262 267\nL 262 267\nA 4 4 90.00 0 1 266 263\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"266\" y=\"280\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">* Morning Star</text><path  d=\"M 433 283\nL 532 283\nL 532 283\nA 4 4 90.00 0 1 536 287\nL 536 300\nL 536 300\nA 4 4 90.00 0 1 532 304\nL 433 304\nL 433 304\nA 4 4 90.00 0 1 429 300\nL 429 287\nL 429 287\nA 4 4 90.00 0 1 433 283\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"433\" y=\"300\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ш Three Soldiers</text><path  d=\"M 488 106\nL 587 106\nL 587 106\nA 4 4 90.00 0 1 591 110\nL 591 123\nL 591 123\nA 4 4 90.00 0 1 587 127\nL 488 127\nL 488 127\nA 4 4 90.00 0 1 484 123\nL 484 110\nL 484 110\nA 4 4 90.00 0 1 488 106\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"488\" y=\"123\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ш Three Soldiers</text><path  d=\"M 544 56\nL 643 56\nL 643 56\nA 4 4 90.00 0 1 647 60\nL 647 73\nL 647 73\nA 4 4 90.00 0 1 643 77\nL 544 77\nL 544 77\nA 4 4 90.00 0 1 540 73\nL 540 60\nL 540 60\nA 4 4 90.00 0 1 544 56\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"544\" y=\"73\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ш Three Soldiers</text><path  d=\"M 600 302\nL 682 302\nL 682 302\nA 4 4 90.00 0 1 686 306\nL 686 319\nL 686 319\nA 4 4 90.00 0 1 682 323\nL 600 323\nL 600 323\nA 4 4 90.00 0 1 596 319\nL 596 306\nL 596 306\nA 4 4 90.00 0 1 600 302\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"600\" y=\"319\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">⁎ Evening Star</text><path  d=\"M 711 351\nL 798 351\nL 798 351\nA 4 4 90.00 0 1 802 355\nL 802 368\nL 802 368\nA 4 4 90.00 0 1 798 372\nL 711 372\nL 711 372\nA 4 4 90.00 0 1 707 368\nL 707 355\nL 707 355\nA 4 4 90.00 0 1 711 351\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"711\" y=\"368\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ω Three Crows</text></svg>",
			pngCRC: 0x164bc641,
		},
		{
			name: "combination_reversal_patterns",
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
				}
				series := newCandlestickWithPatterns(data, PatternDetectionOption{
					DojiThreshold: 0.01,
					ShadowRatio:   2.0,
				})
				return CandlestickChartOption{
					XAxis:      XAxisOption{Show: Ptr(false)},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
					Padding:    NewBoxEqual(10),
				}
			},
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"30\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">132</text><text x=\"9\" y=\"80\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">127.33</text><text x=\"9\" y=\"144\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">122.67</text><text x=\"30\" y=\"208\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">118</text><text x=\"9\" y=\"272\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">113.33</text><text x=\"9\" y=\"337\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">108.67</text><text x=\"30\" y=\"401\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">104</text><text x=\"17\" y=\"465\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">99.33</text><text x=\"17\" y=\"529\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">94.67</text><text x=\"39\" y=\"594\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 63 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 74\nL 790 74\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 138\nL 790 138\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 203\nL 790 203\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 267\nL 790 267\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 332\nL 790 332\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 396\nL 790 396\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 461\nL 790 461\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 63 525\nL 790 525\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 112 314\nL 112 383\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 112 452\nL 112 521\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 94 314\nL 130 314\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 94 521\nL 130 521\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 76 383\nL 148 383\nL 148 452\nL 76 452\nL 76 383\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 202 162\nL 202 176\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 184 162\nL 220 162\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 184 245\nL 220 245\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 166 176\nL 238 176\nL 238 245\nL 166 245\nL 166 176\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 292 190\nL 292 204\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 274 190\nL 310 190\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 274 287\nL 310 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 256 204\nL 328 204\nL 328 287\nL 256 287\nL 256 204\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 365 107\nL 401 107\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 365 204\nL 401 204\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 347 107\nL 419 107\nL 419 204\nL 347 204\nL 347 107\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 473 162\nL 473 176\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 455 80\nL 491 80\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 455 176\nL 491 176\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 437 80\nL 509 80\nL 509 162\nL 437 162\nL 437 80\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 563 93\nL 563 107\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 563 425\nL 563 452\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 545 93\nL 581 93\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 545 452\nL 581 452\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 527 107\nL 599 107\nL 599 425\nL 527 425\nL 527 107\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 654 342\nL 654 356\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 654 425\nL 654 452\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 636 342\nL 672 342\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 636 452\nL 672 452\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 618 356\nL 690 356\nL 690 425\nL 618 425\nL 618 356\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 744 287\nL 744 314\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 744 356\nL 744 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 726 287\nL 762 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 726 425\nL 762 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 708 314\nL 780 314\nL 780 356\nL 708 356\nL 708 314\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 297 191\nL 378 191\nL 378 191\nA 4 4 90.00 0 1 382 195\nL 382 208\nL 382 208\nA 4 4 90.00 0 1 378 212\nL 297 212\nL 297 212\nA 4 4 90.00 0 1 293 208\nL 293 195\nL 293 195\nA 4 4 90.00 0 1 297 191\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"297\" y=\"208\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">| Piercing Line</text><path  d=\"M 388 94\nL 480 94\nL 480 94\nA 4 4 90.00 0 1 484 98\nL 484 111\nL 484 111\nA 4 4 90.00 0 1 480 115\nL 388 115\nL 388 115\nA 4 4 90.00 0 1 384 111\nL 384 98\nL 384 98\nA 4 4 90.00 0 1 388 94\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"388\" y=\"111\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">^ Bull Marubozu</text><path  d=\"M 478 149\nL 552 149\nL 552 149\nA 4 4 90.00 0 1 556 153\nL 556 166\nL 556 166\nA 4 4 90.00 0 1 552 170\nL 478 170\nL 478 170\nA 4 4 90.00 0 1 474 166\nL 474 153\nL 474 153\nA 4 4 90.00 0 1 478 149\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"478\" y=\"166\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ξ Dark Cloud</text><path  d=\"M 659 337\nL 762 337\nL 762 337\nA 4 4 90.00 0 1 766 341\nL 766 367\nL 766 367\nA 4 4 90.00 0 1 762 371\nL 659 371\nL 659 371\nA 4 4 90.00 0 1 655 367\nL 655 341\nL 655 341\nA 4 4 90.00 0 1 659 337\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"672\" y=\"354\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ʘ Bull Harami</text><text x=\"659\" y=\"367\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ǁ Tweezer Bottom</text></svg>",
			pngCRC: 0xaa863548,
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
				}

				series := newCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 800 600\"><path  d=\"M 0 0\nL 800 0\nL 800 600\nL 0 600\nL 0 0\" style=\"stroke:none;fill:white\"/><text x=\"9\" y=\"16\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">153</text><text x=\"9\" y=\"80\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">146</text><text x=\"9\" y=\"144\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">139</text><text x=\"9\" y=\"208\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">132</text><text x=\"9\" y=\"272\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">125</text><text x=\"9\" y=\"337\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">118</text><text x=\"9\" y=\"401\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">111</text><text x=\"9\" y=\"465\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">104</text><text x=\"18\" y=\"529\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">97</text><text x=\"18\" y=\"594\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">90</text><path  d=\"M 42 10\nL 790 10\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 74\nL 790 74\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 138\nL 790 138\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 203\nL 790 203\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 267\nL 790 267\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 332\nL 790 332\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 396\nL 790 396\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 461\nL 790 461\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 42 525\nL 790 525\" style=\"stroke-width:1;stroke:rgb(224,230,242);fill:none\"/><path  d=\"M 52 406\nL 52 452\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 52 498\nL 52 544\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 50 406\nL 54 406\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 50 544\nL 54 544\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 47 452\nL 57 452\nL 57 498\nL 47 498\nL 47 452\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 66 425\nL 66 434\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 66 452\nL 66 480\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 64 425\nL 68 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 64 480\nL 68 480\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 61 434\nL 71 434\nL 71 452\nL 61 452\nL 61 434\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 80 416\nL 80 425\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 80 434\nL 80 517\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 78 416\nL 82 416\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 78 517\nL 82 517\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 75 425\nL 85 425\nL 85 434\nL 75 434\nL 75 425\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 94 480\nL 94 498\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 94 544\nL 94 554\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 92 480\nL 96 480\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 92 554\nL 96 554\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 89 498\nL 99 498\nL 99 544\nL 89 544\nL 89 498\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 107 360\nL 107 388\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 107 443\nL 107 452\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 105 360\nL 109 360\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 105 452\nL 109 452\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 102 388\nL 112 388\nL 112 443\nL 102 443\nL 102 388\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 121 314\nL 121 424\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 121 425\nL 121 434\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 119 314\nL 123 314\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 119 434\nL 123 434\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 116 424\nL 126 424\nL 126 425\nL 116 425\nL 116 424\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 135 397\nL 135 416\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 135 425\nL 135 590\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 133 397\nL 137 397\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 133 590\nL 137 590\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 130 416\nL 140 416\nL 140 425\nL 130 425\nL 130 416\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 147 314\nL 151 314\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 147 498\nL 151 498\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 144 314\nL 154 314\nL 154 498\nL 144 498\nL 144 314\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 161 314\nL 165 314\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 161 498\nL 165 498\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 158 314\nL 168 314\nL 168 498\nL 158 498\nL 158 314\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 176 314\nL 176 333\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 176 406\nL 176 434\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 174 314\nL 178 314\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 174 434\nL 178 434\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 171 333\nL 181 333\nL 181 406\nL 171 406\nL 171 333\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 190 176\nL 190 222\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 190 360\nL 190 406\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 188 176\nL 192 176\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 188 406\nL 192 406\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 185 222\nL 195 222\nL 195 360\nL 185 360\nL 185 222\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 204 130\nL 204 149\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 204 406\nL 204 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 202 130\nL 206 130\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 202 425\nL 206 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 199 149\nL 209 149\nL 209 406\nL 199 406\nL 199 149\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 218 130\nL 218 149\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 218 406\nL 218 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 216 130\nL 220 130\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 216 425\nL 220 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 213 149\nL 223 149\nL 223 406\nL 213 406\nL 213 149\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 231 112\nL 231 130\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 231 434\nL 231 452\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 229 112\nL 233 112\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 229 452\nL 233 452\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 226 130\nL 236 130\nL 236 434\nL 226 434\nL 226 130\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 245 176\nL 245 222\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 245 452\nL 245 498\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 243 176\nL 247 176\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 243 498\nL 247 498\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 240 222\nL 250 222\nL 250 452\nL 240 452\nL 240 222\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 259 268\nL 259 314\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 259 406\nL 259 544\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 257 268\nL 261 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 257 544\nL 261 544\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 254 314\nL 264 314\nL 264 406\nL 254 406\nL 254 314\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 273 176\nL 273 222\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 273 498\nL 273 517\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 271 176\nL 275 176\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 271 517\nL 275 517\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 268 222\nL 278 222\nL 278 498\nL 268 498\nL 268 222\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 287 241\nL 287 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 287 296\nL 287 314\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 285 241\nL 289 241\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 285 314\nL 289 314\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 282 268\nL 292 268\nL 292 296\nL 282 296\nL 282 268\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 300 268\nL 300 314\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 300 388\nL 300 406\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 298 268\nL 302 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 298 406\nL 302 406\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 295 314\nL 305 314\nL 305 388\nL 295 388\nL 295 314\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 314 268\nL 314 333\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 314 425\nL 314 434\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 312 268\nL 316 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 312 434\nL 316 434\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 309 333\nL 319 333\nL 319 425\nL 309 425\nL 309 333\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 328 268\nL 328 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 328 406\nL 328 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 326 268\nL 330 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 326 425\nL 330 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 323 287\nL 333 287\nL 333 406\nL 323 406\nL 323 287\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 342 222\nL 342 241\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 342 360\nL 342 388\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 340 222\nL 344 222\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 340 388\nL 344 388\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 337 241\nL 347 241\nL 347 360\nL 337 360\nL 337 241\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 355 222\nL 355 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 355 406\nL 355 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 353 222\nL 357 222\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 353 425\nL 357 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 350 268\nL 360 268\nL 360 406\nL 350 406\nL 350 268\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 369 222\nL 369 287\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 369 333\nL 369 360\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 367 222\nL 371 222\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 367 360\nL 371 360\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 364 287\nL 374 287\nL 374 333\nL 364 333\nL 364 287\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 383 268\nL 383 314\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 383 452\nL 383 498\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 381 268\nL 385 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 381 498\nL 385 498\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 378 314\nL 388 314\nL 388 452\nL 378 452\nL 378 314\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 397 360\nL 397 388\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 397 425\nL 397 498\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 395 360\nL 399 360\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 395 498\nL 399 498\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 392 388\nL 402 388\nL 402 425\nL 392 425\nL 392 388\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 411 176\nL 411 222\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 411 360\nL 411 406\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 409 176\nL 413 176\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 409 406\nL 413 406\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 406 222\nL 416 222\nL 416 360\nL 406 360\nL 406 222\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 424 388\nL 424 406\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 424 425\nL 424 434\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 422 388\nL 426 388\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 422 434\nL 426 434\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 419 406\nL 429 406\nL 429 425\nL 419 425\nL 419 406\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 438 130\nL 438 176\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 438 360\nL 438 379\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 436 130\nL 440 130\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 436 379\nL 440 379\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 433 176\nL 443 176\nL 443 360\nL 433 360\nL 433 176\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 452 130\nL 452 176\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 452 406\nL 452 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 450 130\nL 454 130\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 450 425\nL 454 425\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 447 176\nL 457 176\nL 457 406\nL 447 406\nL 447 176\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 466 84\nL 466 130\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 466 158\nL 466 167\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 464 84\nL 468 84\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 464 167\nL 468 167\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 461 130\nL 471 130\nL 471 158\nL 461 158\nL 461 130\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 479 167\nL 479 176\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 479 314\nL 479 360\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 477 167\nL 481 167\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 477 360\nL 481 360\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 474 176\nL 484 176\nL 484 314\nL 474 314\nL 474 176\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 493 268\nL 493 314\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 493 360\nL 493 406\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 491 268\nL 495 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 491 406\nL 495 406\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 488 314\nL 498 314\nL 498 360\nL 488 360\nL 488 314\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 507 241\nL 507 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 507 333\nL 507 351\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 505 241\nL 509 241\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 505 351\nL 509 351\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 502 268\nL 512 268\nL 512 333\nL 502 333\nL 502 268\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 521 176\nL 521 204\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 521 250\nL 521 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 519 176\nL 523 176\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 519 268\nL 523 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 516 204\nL 526 204\nL 526 250\nL 516 250\nL 516 204\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 535 112\nL 535 130\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 535 185\nL 535 204\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 533 112\nL 537 112\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 533 204\nL 537 204\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 530 130\nL 540 130\nL 540 185\nL 530 185\nL 530 130\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 548 84\nL 548 112\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 548 222\nL 548 241\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 546 84\nL 550 84\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 546 241\nL 550 241\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 543 112\nL 553 112\nL 553 222\nL 543 222\nL 543 112\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 562 130\nL 562 149\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 562 204\nL 562 241\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 560 130\nL 564 130\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 560 241\nL 564 241\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 557 149\nL 567 149\nL 567 204\nL 557 204\nL 557 149\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 576 204\nL 576 222\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 576 268\nL 576 314\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 574 204\nL 578 204\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 574 314\nL 578 314\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 571 222\nL 581 222\nL 581 268\nL 571 268\nL 571 222\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 590 250\nL 590 277\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 590 333\nL 590 360\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 588 250\nL 592 250\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 588 360\nL 592 360\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 585 277\nL 595 277\nL 595 333\nL 585 333\nL 585 277\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 603 314\nL 603 333\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 603 360\nL 603 370\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 601 314\nL 605 314\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 601 370\nL 605 370\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 598 333\nL 608 333\nL 608 360\nL 598 360\nL 598 333\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 617 268\nL 617 296\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 617 333\nL 617 360\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 615 268\nL 619 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 615 360\nL 619 360\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 612 296\nL 622 296\nL 622 333\nL 612 333\nL 612 296\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 631 268\nL 631 277\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 631 314\nL 631 333\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 629 268\nL 633 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 629 333\nL 633 333\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 626 277\nL 636 277\nL 636 314\nL 626 314\nL 626 277\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 645 130\nL 645 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 645 287\nL 645 296\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 643 130\nL 647 130\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 643 296\nL 647 296\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 640 268\nL 650 268\nL 650 287\nL 640 287\nL 640 268\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 659 222\nL 659 241\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 659 268\nL 659 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 657 222\nL 661 222\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 657 287\nL 661 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 654 241\nL 664 241\nL 664 268\nL 654 268\nL 654 241\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 672 84\nL 672 240\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 672 241\nL 672 250\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 670 84\nL 674 84\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 670 250\nL 674 250\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 667 240\nL 677 240\nL 677 241\nL 667 241\nL 667 240\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 686 222\nL 686 241\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 686 259\nL 686 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 684 222\nL 688 222\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 684 268\nL 688 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 681 241\nL 691 241\nL 691 259\nL 681 259\nL 681 241\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 700 259\nL 700 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 700 268\nL 700 406\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 698 259\nL 702 259\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 698 406\nL 702 406\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 695 268\nL 705 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 714 250\nL 714 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 714 333\nL 714 360\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 712 250\nL 716 250\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 712 360\nL 716 360\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 709 268\nL 719 268\nL 719 333\nL 709 333\nL 709 268\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 727 268\nL 727 287\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 727 314\nL 727 360\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 725 268\nL 729 268\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 725 360\nL 729 360\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 722 287\nL 732 287\nL 732 314\nL 722 314\nL 722 287\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 741 176\nL 741 195\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 741 314\nL 741 333\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 739 176\nL 743 176\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 739 333\nL 743 333\" style=\"stroke-width:1;stroke:rgb(34,197,94);fill:none\"/><path  d=\"M 736 195\nL 746 195\nL 746 314\nL 736 314\nL 736 195\" style=\"stroke:none;fill:rgb(34,197,94)\"/><path  d=\"M 755 195\nL 755 204\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 755 259\nL 755 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 753 195\nL 757 195\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 753 268\nL 757 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 750 204\nL 760 204\nL 760 259\nL 750 259\nL 750 204\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 769 213\nL 769 222\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 769 296\nL 769 305\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 767 213\nL 771 213\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 767 305\nL 771 305\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 764 222\nL 774 222\nL 774 296\nL 764 296\nL 764 222\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 783 259\nL 783 268\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 783 351\nL 783 360\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 781 259\nL 785 259\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 781 360\nL 785 360\" style=\"stroke-width:1;stroke:rgb(239,68,68);fill:none\"/><path  d=\"M 778 268\nL 788 268\nL 788 351\nL 778 351\nL 778 268\" style=\"stroke:none;fill:rgb(239,68,68)\"/><path  d=\"M 85 415\nL 171 415\nL 171 415\nA 4 4 90.00 0 1 175 419\nL 175 445\nL 175 445\nA 4 4 90.00 0 1 171 449\nL 85 449\nL 85 449\nA 4 4 90.00 0 1 81 445\nL 81 419\nL 81 419\nA 4 4 90.00 0 1 85 415\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"98\" y=\"432\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><text x=\"85\" y=\"445\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 126 385\nL 218 385\nL 218 385\nA 4 4 90.00 0 1 222 389\nL 222 454\nL 222 454\nA 4 4 90.00 0 1 218 458\nL 126 458\nL 126 458\nA 4 4 90.00 0 1 122 454\nL 122 389\nL 122 389\nA 4 4 90.00 0 1 126 385\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"155\" y=\"402\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text><text x=\"130\" y=\"415\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text><text x=\"126\" y=\"428\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"134\" y=\"441\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">† Gravestone</text><text x=\"129\" y=\"454\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 140 406\nL 226 406\nL 226 406\nA 4 4 90.00 0 1 230 410\nL 230 436\nL 230 436\nA 4 4 90.00 0 1 226 440\nL 140 440\nL 140 440\nA 4 4 90.00 0 1 136 436\nL 136 410\nL 136 410\nA 4 4 90.00 0 1 140 406\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"153\" y=\"423\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><text x=\"140\" y=\"436\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 154 295\nL 246 295\nL 246 295\nA 4 4 90.00 0 1 250 299\nL 250 325\nL 250 325\nA 4 4 90.00 0 1 246 329\nL 154 329\nL 154 329\nA 4 4 90.00 0 1 150 325\nL 150 299\nL 150 299\nA 4 4 90.00 0 1 154 295\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"154\" y=\"312\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">^ Bull Marubozu</text><text x=\"154\" y=\"325\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Λ Bull Engulfing</text><path  d=\"M 168 479\nL 265 479\nL 265 479\nA 4 4 90.00 0 1 269 483\nL 269 509\nL 269 509\nA 4 4 90.00 0 1 265 513\nL 168 513\nL 168 513\nA 4 4 90.00 0 1 164 509\nL 164 483\nL 164 483\nA 4 4 90.00 0 1 168 479\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"168\" y=\"496\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">v Bear Marubozu</text><text x=\"174\" y=\"509\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">‖ Tweezer Top</text><path  d=\"M 209 136\nL 300 136\nL 300 136\nA 4 4 90.00 0 1 304 140\nL 304 153\nL 304 153\nA 4 4 90.00 0 1 300 157\nL 209 157\nL 209 157\nA 4 4 90.00 0 1 205 153\nL 205 140\nL 205 140\nA 4 4 90.00 0 1 209 136\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"209\" y=\"153\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Λ Bull Engulfing</text><path  d=\"M 236 421\nL 331 421\nL 331 421\nA 4 4 90.00 0 1 335 425\nL 335 438\nL 335 438\nA 4 4 90.00 0 1 331 442\nL 236 442\nL 236 442\nA 4 4 90.00 0 1 232 438\nL 232 425\nL 232 425\nA 4 4 90.00 0 1 236 421\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"236\" y=\"438\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">V Bear Engulfing</text><path  d=\"M 292 283\nL 374 283\nL 374 283\nA 4 4 90.00 0 1 378 287\nL 378 300\nL 378 300\nA 4 4 90.00 0 1 374 304\nL 292 304\nL 292 304\nA 4 4 90.00 0 1 288 300\nL 288 287\nL 288 287\nA 4 4 90.00 0 1 292 283\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"292\" y=\"300\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">θ Bear Harami</text><path  d=\"M 319 320\nL 400 320\nL 400 320\nA 4 4 90.00 0 1 404 324\nL 404 337\nL 404 337\nA 4 4 90.00 0 1 400 341\nL 319 341\nL 319 341\nA 4 4 90.00 0 1 315 337\nL 315 324\nL 315 324\nA 4 4 90.00 0 1 319 320\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"319\" y=\"337\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">| Piercing Line</text><path  d=\"M 347 347\nL 421 347\nL 421 347\nA 4 4 90.00 0 1 425 351\nL 425 364\nL 425 364\nA 4 4 90.00 0 1 421 368\nL 347 368\nL 347 368\nA 4 4 90.00 0 1 343 364\nL 343 351\nL 343 351\nA 4 4 90.00 0 1 347 347\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"347\" y=\"364\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ξ Dark Cloud</text><path  d=\"M 360 255\nL 441 255\nL 441 255\nA 4 4 90.00 0 1 445 259\nL 445 272\nL 445 272\nA 4 4 90.00 0 1 441 276\nL 360 276\nL 360 276\nA 4 4 90.00 0 1 356 272\nL 356 259\nL 356 259\nA 4 4 90.00 0 1 360 255\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"360\" y=\"272\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">| Piercing Line</text><path  d=\"M 374 320\nL 458 320\nL 458 320\nA 4 4 90.00 0 1 462 324\nL 462 337\nL 462 337\nA 4 4 90.00 0 1 458 341\nL 374 341\nL 374 341\nA 4 4 90.00 0 1 370 337\nL 370 324\nL 370 324\nA 4 4 90.00 0 1 374 320\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"374\" y=\"337\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">‖ Tweezer Top</text><path  d=\"M 402 369\nL 505 369\nL 505 369\nA 4 4 90.00 0 1 509 373\nL 509 399\nL 509 399\nA 4 4 90.00 0 1 505 403\nL 402 403\nL 402 403\nA 4 4 90.00 0 1 398 399\nL 398 373\nL 398 373\nA 4 4 90.00 0 1 402 369\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"415\" y=\"386\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ʘ Bull Harami</text><text x=\"402\" y=\"399\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ǁ Tweezer Bottom</text><path  d=\"M 443 163\nL 527 163\nL 527 163\nA 4 4 90.00 0 1 531 167\nL 531 180\nL 531 180\nA 4 4 90.00 0 1 527 184\nL 443 184\nL 443 184\nA 4 4 90.00 0 1 439 180\nL 439 167\nL 439 167\nA 4 4 90.00 0 1 443 163\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"443\" y=\"180\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">* Morning Star</text><path  d=\"M 484 301\nL 566 301\nL 566 301\nA 4 4 90.00 0 1 570 305\nL 570 318\nL 570 318\nA 4 4 90.00 0 1 566 322\nL 484 322\nL 484 322\nA 4 4 90.00 0 1 480 318\nL 480 305\nL 480 305\nA 4 4 90.00 0 1 484 301\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"484\" y=\"318\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">⁎ Evening Star</text><path  d=\"M 540 117\nL 639 117\nL 639 117\nA 4 4 90.00 0 1 643 121\nL 643 134\nL 643 134\nA 4 4 90.00 0 1 639 138\nL 540 138\nL 540 138\nA 4 4 90.00 0 1 536 134\nL 536 121\nL 536 121\nA 4 4 90.00 0 1 540 117\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"540\" y=\"134\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ш Three Soldiers</text><path  d=\"M 650 249\nL 742 249\nL 742 249\nA 4 4 90.00 0 1 746 253\nL 746 279\nL 746 279\nA 4 4 90.00 0 1 742 283\nL 650 283\nL 650 283\nA 4 4 90.00 0 1 646 279\nL 646 253\nL 646 253\nA 4 4 90.00 0 1 650 249\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"654\" y=\"266\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text><text x=\"650\" y=\"279\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><path  d=\"M 677 201\nL 769 201\nL 769 201\nA 4 4 90.00 0 1 773 205\nL 773 270\nL 773 270\nA 4 4 90.00 0 1 769 274\nL 677 274\nL 677 274\nA 4 4 90.00 0 1 673 270\nL 673 205\nL 673 205\nA 4 4 90.00 0 1 677 201\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"706\" y=\"218\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text><text x=\"681\" y=\"231\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Ʇ Inv. Hammer</text><text x=\"677\" y=\"244\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">※ Shooting Star</text><text x=\"685\" y=\"257\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">† Gravestone</text><text x=\"680\" y=\"270\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 705 236\nL 791 236\nL 791 236\nA 4 4 90.00 0 1 795 240\nL 795 292\nL 795 292\nA 4 4 90.00 0 1 791 296\nL 705 296\nL 705 296\nA 4 4 90.00 0 1 701 292\nL 701 240\nL 701 240\nA 4 4 90.00 0 1 705 236\nZ\" style=\"stroke-width:1.2;stroke:rgb(200,200,200);fill:rgba(255,255,255,0.7)\"/><text x=\"731\" y=\"253\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">± Doji</text><text x=\"718\" y=\"266\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">Γ Hammer</text><text x=\"714\" y=\"279\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ψ Dragonfly</text><text x=\"705\" y=\"292\" style=\"stroke:none;fill:rgb(128,128,128);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">◌ Spinning Top</text><path  d=\"M 697 274\nL 800 274\nL 800 274\nA 4 4 90.00 0 1 804 278\nL 804 291\nL 804 291\nA 4 4 90.00 0 1 800 295\nL 697 295\nL 697 295\nA 4 4 90.00 0 1 693 291\nL 693 278\nL 693 278\nA 4 4 90.00 0 1 697 274\nZ\" style=\"stroke-width:1.2;stroke:rgb(34,197,94);fill:rgba(255,255,255,0.7)\"/><text x=\"697\" y=\"291\" style=\"stroke:none;fill:rgb(12,75,35);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ǁ Tweezer Bottom</text><path  d=\"M 713 338\nL 800 338\nL 800 338\nA 4 4 90.00 0 1 804 342\nL 804 355\nL 804 355\nA 4 4 90.00 0 1 800 359\nL 713 359\nL 713 359\nA 4 4 90.00 0 1 709 355\nL 709 342\nL 709 342\nA 4 4 90.00 0 1 713 338\nZ\" style=\"stroke-width:1.2;stroke:rgb(239,68,68);fill:rgba(255,255,255,0.7)\"/><text x=\"713\" y=\"355\" style=\"stroke:none;fill:rgb(151,12,12);font-size:12.8px;font-family:'Roboto Medium',sans-serif\">ω Three Crows</text></svg>",
			pngCRC: 0xdb21a56a,
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
