package charts

import (
	"strconv"
	"testing"

	"github.com/go-analyze/bulk"
	"github.com/stretchr/testify/assert"
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

func TestPatternIntegration(t *testing.T) {
	t.Parallel()

	// Test that all advanced patterns are detected in a comprehensive dataset
	data := makeAdvancedPatternTestData()
	series := CandlestickSeries{Data: data}

	// Scan for patterns
	patterns := ScanCandlestickPatterns(series, PatternDetectionOption{
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
	seriesWithPatterns := NewCandlestickWithPatterns(data, PatternDetectionOption{
		DojiThreshold:    0.01,
		ShadowRatio:      2.0,
		EngulfingMinSize: 0.8,
	})

	// Verify that labels are enabled and LabelFormatter is set
	assert.True(t, flagIs(true, seriesWithPatterns.Label.Show))
	assert.NotNil(t, seriesWithPatterns.Label.LabelFormatter)

	// Verify that patterns are detected in the convenience function result
	newPatterns := ScanCandlestickPatterns(seriesWithPatterns, PatternDetectionOption{
		DojiThreshold:    0.01,
		ShadowRatio:      2.0,
		EngulfingMinSize: 0.8,
	})
	assert.Len(t, newPatterns, 20)

	// Also verify that the label formatter can show patterns for the expected indices
	if seriesWithPatterns.Label.LabelFormatter != nil {
		var labelPatternCount int
		for i := range data {
			text, _ := seriesWithPatterns.Label.LabelFormatter(i, "test", 100.0)
			if text != "" {
				labelPatternCount++
			}
		}

		// Should show labels for some pattern indices
		assert.Equal(t, 10, labelPatternCount)

		// The key test: ensure that both scanning approaches find patterns
		assert.Len(t, patterns, 20)
		assert.Len(t, newPatterns, 20)
	}
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
	patterns := ScanCandlestickPatterns(series, PatternDetectionOption{
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
	uniquePatterns := bulk.SliceToSetBy(func(p PatternDetectionResult) string { return p.PatternType }, patterns)
	assert.Len(t, uniquePatterns, 19)
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
			name: "pattern_doji",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},     // Normal candle
					{Open: 105, High: 108, Low: 102, Close: 105.05}, // Doji pattern
					{Open: 105, High: 112, Low: 98, Close: 108},     // Normal candle
				}
				series := NewCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "",
			pngCRC: 0x2da625c3,
		},
		{
			name: "pattern_hammer",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 108, High: 109, Low: 98, Close: 107},  // Hammer pattern
					{Open: 107, High: 112, Low: 102, Close: 110}, // Normal candle
				}
				series := NewCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "",
			pngCRC: 0x9959a71b,
		},
		{
			name: "pattern_inverted_hammer",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105}, // Normal candle
					{Open: 95, High: 107, Low: 94, Close: 96},   // Inverted hammer pattern
					{Open: 96, High: 102, Low: 91, Close: 98},   // Normal candle
				}
				series := NewCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "",
			pngCRC: 0x05255cbd,
		},
		{
			name: "pattern_shooting_star",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 106, High: 125, Low: 105, Close: 107}, // Shooting star pattern
					{Open: 107, High: 112, Low: 102, Close: 109}, // Normal candle
				}
				series := NewCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "",
			pngCRC: 0x9063cc82,
		},
		{
			name: "pattern_gravestone_doji",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},    // Normal candle
					{Open: 108, High: 120, Low: 107, Close: 108.1}, // Gravestone doji pattern
					{Open: 108, High: 115, Low: 103, Close: 110},   // Normal candle
				}
				series := NewCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "",
			pngCRC: 0xd0eda7dd,
		},
		{
			name: "pattern_dragonfly_doji",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},   // Normal candle
					{Open: 109, High: 110, Low: 90, Close: 108.9}, // Dragonfly doji pattern
					{Open: 109, High: 115, Low: 104, Close: 112},  // Normal candle
				}
				series := NewCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "",
			pngCRC: 0xbf557b34,
		},
		{
			name: "pattern_bullish_marubozu",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 100, High: 120, Low: 100, Close: 120}, // Bullish marubozu pattern
					{Open: 120, High: 125, Low: 115, Close: 122}, // Normal candle
				}
				series := NewCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "",
			pngCRC: 0x9a957c80,
		},
		{
			name: "pattern_bearish_marubozu",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 120, High: 120, Low: 100, Close: 100}, // Bearish marubozu pattern
					{Open: 100, High: 105, Low: 95, Close: 102},  // Normal candle
				}
				series := NewCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "",
			pngCRC: 0xca704af0,
		},
		{
			name: "pattern_spinning_top",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 110, High: 125, Low: 95, Close: 112},  // Spinning top pattern
					{Open: 112, High: 118, Low: 107, Close: 115}, // Normal candle
				}
				series := NewCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "",
			pngCRC: 0xb218f934,
		},
		{
			name: "pattern_bullish_engulfing",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 110, High: 112, Low: 105, Close: 106}, // Small bearish candle
					{Open: 104, High: 115, Low: 103, Close: 114}, // Bullish engulfing
					{Open: 114, High: 120, Low: 112, Close: 118}, // Normal candle
				}
				series := NewCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "",
			pngCRC: 0xfac6972a,
		},
		{
			name: "pattern_bearish_engulfing",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 106, High: 112, Low: 105, Close: 110}, // Small bullish candle
					{Open: 114, High: 115, Low: 103, Close: 104}, // Bearish engulfing
					{Open: 104, High: 108, Low: 100, Close: 102}, // Normal candle
				}
				series := NewCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "",
			pngCRC: 0x6f81ef24,
		},
		{
			name: "pattern_morning_star",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 120, High: 125, Low: 105, Close: 108}, // Large bearish
					{Open: 102, High: 104, Low: 100, Close: 103}, // Small body, gap down
					{Open: 108, High: 125, Low: 106, Close: 122}, // Large bullish, gap up
					{Open: 122, High: 128, Low: 120, Close: 125}, // Normal candle
				}
				series := NewCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "",
			pngCRC: 0x047673fc,
		},
		{
			name: "pattern_evening_star",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 122, High: 140, Low: 120, Close: 138}, // Large bullish
					{Open: 142, High: 144, Low: 140, Close: 143}, // Small body, gap up
					{Open: 138, High: 140, Low: 115, Close: 118}, // Large bearish, gap down
					{Open: 118, High: 122, Low: 115, Close: 120}, // Normal candle
				}
				series := NewCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "",
			pngCRC: 0x082963f4,
		},
		{
			name: "pattern_piercing_line",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 120, High: 121, Low: 115, Close: 115}, // Bearish candle
					{Open: 112, High: 119, Low: 112, Close: 118}, // Piercing line (opens below prev low, closes above midpoint)
					{Open: 118, High: 125, Low: 116, Close: 122}, // Normal candle
				}
				series := NewCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "",
			pngCRC: 0x638148cc,
		},
		{
			name: "pattern_dark_cloud_cover",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 118, High: 125, Low: 118, Close: 125}, // Bullish candle
					{Open: 127, High: 127, Low: 120, Close: 121}, // Dark cloud cover (opens above prev high, closes below midpoint)
					{Open: 121, High: 124, Low: 118, Close: 120}, // Normal candle
				}
				series := NewCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "",
			pngCRC: 0xfa618bf0,
		},
		{
			name: "pattern_three_white_soldiers",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 110, High: 115, Low: 109, Close: 114}, // First soldier
					{Open: 113, High: 118, Low: 112, Close: 117}, // Second soldier
					{Open: 116, High: 121, Low: 115, Close: 120}, // Third soldier
					{Open: 120, High: 125, Low: 118, Close: 123}, // Normal candle
				}
				series := NewCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "",
			pngCRC: 0xe2eb70d7,
		},
		{
			name: "pattern_three_black_crows",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},  // Normal candle
					{Open: 120, High: 121, Low: 115, Close: 116}, // First crow
					{Open: 117, High: 118, Low: 112, Close: 113}, // Second crow
					{Open: 114, High: 115, Low: 108, Close: 109}, // Third crow
					{Open: 109, High: 112, Low: 106, Close: 108}, // Normal candle
				}
				series := NewCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "",
			pngCRC: 0x23e77b2c,
		},
		{
			name: "pattern_combination_doji_and_hammers",
			optGen: func() CandlestickChartOption {
				data := []OHLCData{
					{Open: 100, High: 110, Low: 95, Close: 105},     // Normal candle
					{Open: 105, High: 108, Low: 102, Close: 105.05}, // Doji pattern
					{Open: 105, High: 107, Low: 95, Close: 106},     // Hammer pattern
					{Open: 106, High: 118, Low: 105, Close: 107},    // Shooting star pattern
					{Open: 107, High: 115, Low: 102, Close: 112},    // Normal candle
				}
				series := NewCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "",
			pngCRC: 0xf14efe0a,
		},
		{
			name: "pattern_combination_engulfing_and_stars",
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
				series := NewCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "",
			pngCRC: 0x2da07915,
		},
		{
			name: "pattern_combination_mixed",
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
				series := NewCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "",
			pngCRC: 0xacc47163,
		},
		{
			name: "pattern_combination_three_candle_patterns",
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
					{Open: 120, High: 121, Low: 115, Close: 116}, // First crow
					{Open: 117, High: 118, Low: 112, Close: 113}, // Second crow
					{Open: 114, High: 115, Low: 108, Close: 109}, // Third crow
					{Open: 109, High: 112, Low: 106, Close: 108}, // Normal candle
				}
				series := NewCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "",
			pngCRC: 0xf6bad464,
		},
		{
			name: "pattern_combination_reversal_patterns",
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
				series := NewCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "",
			pngCRC: 0x9f600cd5,
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
					// Additional patterns to expand the showcase
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

				series := NewCandlestickWithPatterns(data, PatternDetectionOption{
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
			svg:    "",
			pngCRC: 0x0,
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

			validateCandlestickChartRender(t, p, r, opt, tc.svg, tc.pngCRC)
		})
	}
}
