package charts

import (
	"math"
	"strings"
)

// Candlestick pattern constants
const (
	// Single candle patterns
	PatternDoji           = "doji"
	PatternHammer         = "hammer"
	PatternInvertedHammer = "inverted_hammer"
	PatternShootingStar   = "shooting_star"
	PatternGravestone     = "gravestone_doji"
	PatternDragonfly      = "dragonfly_doji"
	PatternMarubozuBull   = "marubozu_bullish"
	PatternMarubozuBear   = "marubozu_bearish"
	PatternSpinningTop    = "spinning_top"

	// Two candle patterns
	PatternEngulfingBull  = "engulfing_bullish"
	PatternEngulfingBear  = "engulfing_bearish"
	PatternHarami         = "harami"
	PatternPiercingLine   = "piercing_line"
	PatternDarkCloudCover = "dark_cloud_cover"
	PatternTweezerTop     = "tweezer_top"
	PatternTweezerBottom  = "tweezer_bottom"

	// Three candle patterns
	PatternMorningStar        = "morning_star"
	PatternEveningStar        = "evening_star"
	PatternThreeWhiteSoldiers = "three_white_soldiers"
	PatternThreeBlackCrows    = "three_black_crows"
)

// PatternFormatter allows custom formatting of detected patterns
type PatternFormatter func(patterns []PatternDetectionResult, seriesName string, value float64) (string, *LabelStyle)

// CandlestickPatternConfig configures automatic pattern detection.
type CandlestickPatternConfig struct {
	// ReplaceSeriesLabel controls pattern/user label interaction
	// true = patterns replace user labels, false = patterns complement user labels
	ReplaceSeriesLabel bool

	// EnabledPatterns lists specific patterns to detect
	// nil or empty = no patterns detected (PatternConfig must be set to enable)
	// ["doji", "hammer"] = only these patterns
	EnabledPatterns []string

	// DetectionOptions configures detection sensitivity thresholds
	DetectionOptions PatternDetectionOption

	// PatternFormatter allows custom formatting (optional)
	// If nil, uses default formatting with theme colors
	PatternFormatter PatternFormatter
}

// TODO - merge with config above?
// PatternDetectionOption configures pattern detection sensitivity
type PatternDetectionOption struct {
	// DojiThreshold is the percentage threshold for doji detection (default: 0.1%)
	DojiThreshold float64
	// ShadowRatio is the minimum shadow-to-body ratio for hammer patterns (default: 2.0)
	ShadowRatio float64
	// EngulfingMinSize is minimum engulfing percentage (default: 80%)
	EngulfingMinSize float64
}

// =============================================================================
// SINGLE CANDLE PATTERNS
// =============================================================================

// DetectDoji identifies doji patterns where open ≈ close
func DetectDoji(ohlc OHLCData, threshold float64) bool {
	if !validateOHLCData(ohlc) {
		return false
	}
	if threshold <= 0 {
		threshold = 0.001 // 0.1% default
	}
	bodySize := math.Abs(ohlc.Close - ohlc.Open)
	priceRange := ohlc.High - ohlc.Low
	if priceRange == 0 {
		return false
	}
	return (bodySize / priceRange) <= threshold
}

// DetectHammer identifies hammer patterns (long lower shadow, small body at top)
func DetectHammer(ohlc OHLCData, shadowRatio float64) bool {
	if !validateOHLCData(ohlc) {
		return false
	}
	if shadowRatio <= 0 {
		shadowRatio = 2.0 // default
	}

	bodySize := math.Abs(ohlc.Close - ohlc.Open)
	lowerShadow := math.Min(ohlc.Open, ohlc.Close) - ohlc.Low
	upperShadow := ohlc.High - math.Max(ohlc.Open, ohlc.Close)

	// Hammer: long lower shadow, short upper shadow, small body
	return lowerShadow >= shadowRatio*bodySize && upperShadow <= lowerShadow*0.3
}

// DetectInvertedHammer identifies inverted hammer patterns
func DetectInvertedHammer(ohlc OHLCData, shadowRatio float64) bool {
	if !validateOHLCData(ohlc) {
		return false
	}
	if shadowRatio <= 0 {
		shadowRatio = 2.0
	}

	bodySize := math.Abs(ohlc.Close - ohlc.Open)
	lowerShadow := math.Min(ohlc.Open, ohlc.Close) - ohlc.Low
	upperShadow := ohlc.High - math.Max(ohlc.Open, ohlc.Close)

	// Inverted hammer: long upper shadow, short lower shadow, small body
	return upperShadow >= shadowRatio*bodySize && lowerShadow <= upperShadow*0.3
}

// DetectShootingStar identifies shooting star patterns (bearish reversal with long upper shadow)
func DetectShootingStar(ohlc OHLCData, shadowRatio float64) bool {
	if !validateOHLCData(ohlc) {
		return false
	}
	if shadowRatio <= 0 {
		shadowRatio = 2.0 // default
	}

	bodySize := math.Abs(ohlc.Close - ohlc.Open)
	lowerShadow := math.Min(ohlc.Open, ohlc.Close) - ohlc.Low
	upperShadow := ohlc.High - math.Max(ohlc.Open, ohlc.Close)

	// Shooting star: long upper shadow, relatively small lower shadow, small body near the low
	hasLongUpperShadow := upperShadow >= shadowRatio*bodySize
	hasShortLowerShadow := lowerShadow <= upperShadow*0.3

	// Body should be in lower third of the total range
	totalRange := ohlc.High - ohlc.Low
	if totalRange == 0 {
		return false
	}
	bodyPosition := (math.Min(ohlc.Open, ohlc.Close) - ohlc.Low) / totalRange
	isNearLow := bodyPosition <= 0.33

	return hasLongUpperShadow && hasShortLowerShadow && isNearLow
}

// DetectGravestoneDoji identifies gravestone doji patterns (doji with long upper shadow)
func DetectGravestoneDoji(ohlc OHLCData, options PatternDetectionOption) bool {
	if !validateOHLCData(ohlc) {
		return false
	}

	// Must be a doji first
	if !DetectDoji(ohlc, options.DojiThreshold) {
		return false
	}

	bodyMidpoint := (ohlc.Open + ohlc.Close) / 2
	upperShadow := ohlc.High - bodyMidpoint
	lowerShadow := bodyMidpoint - ohlc.Low

	shadowRatio := options.ShadowRatio
	if shadowRatio <= 0 {
		shadowRatio = 2.0
	}

	// Gravestone doji: long upper shadow, minimal lower shadow
	hasLongUpperShadow := upperShadow >= shadowRatio*math.Abs(ohlc.Close-ohlc.Open)
	hasMinimalLowerShadow := lowerShadow <= upperShadow*0.3

	return hasLongUpperShadow && hasMinimalLowerShadow
}

// DetectDragonflyDoji identifies dragonfly doji patterns (doji with long lower shadow)
func DetectDragonflyDoji(ohlc OHLCData, options PatternDetectionOption) bool {
	if !validateOHLCData(ohlc) {
		return false
	}

	// Must be a doji first
	if !DetectDoji(ohlc, options.DojiThreshold) {
		return false
	}

	bodyMidpoint := (ohlc.Open + ohlc.Close) / 2
	upperShadow := ohlc.High - bodyMidpoint
	lowerShadow := bodyMidpoint - ohlc.Low

	shadowRatio := options.ShadowRatio
	if shadowRatio <= 0 {
		shadowRatio = 2.0
	}

	// Dragonfly doji: long lower shadow, minimal upper shadow
	hasLongLowerShadow := lowerShadow >= shadowRatio*math.Abs(ohlc.Close-ohlc.Open)
	hasMinimalUpperShadow := upperShadow <= lowerShadow*0.3

	return hasLongLowerShadow && hasMinimalUpperShadow
}

// DetectMarubozu identifies marubozu patterns (no shadows, strong trend)
func DetectMarubozu(ohlc OHLCData, threshold float64) (bool, bool) {
	if !validateOHLCData(ohlc) {
		return false, false
	}
	if threshold <= 0 {
		threshold = 0.01 // 1% default tolerance
	}

	// Calculate shadow sizes
	upper := ohlc.High - math.Max(ohlc.Open, ohlc.Close)
	lower := math.Min(ohlc.Open, ohlc.Close) - ohlc.Low
	body := math.Abs(ohlc.Close - ohlc.Open)
	total := ohlc.High - ohlc.Low

	if total == 0 || body == 0 {
		return false, false
	}

	// Shadows should be minimal compared to total range
	hasMinimalShadows := (upper+lower)/total <= threshold

	if !hasMinimalShadows {
		return false, false
	}

	// Determine bullish or bearish
	bullish := ohlc.Close > ohlc.Open
	bearish := ohlc.Close < ohlc.Open

	return bullish, bearish
}

// DetectSpinningTop identifies spinning top patterns (small body, long shadows)
func DetectSpinningTop(ohlc OHLCData, bodyRatio float64) bool {
	if !validateOHLCData(ohlc) {
		return false
	}
	if bodyRatio <= 0 {
		bodyRatio = 0.3 // Body should be less than 30% of total range
	}

	body := math.Abs(ohlc.Close - ohlc.Open)
	total := ohlc.High - ohlc.Low
	upper := ohlc.High - math.Max(ohlc.Open, ohlc.Close)
	lower := math.Min(ohlc.Open, ohlc.Close) - ohlc.Low

	if total == 0 {
		return false
	}

	// Small body relative to total range
	hasSmallBody := (body / total) <= bodyRatio
	// Both shadows should be at least as long as the body (indicating indecision)
	// AND the total range should be reasonably large to indicate real indecision
	hasLongShadows := upper >= body && lower >= body && total > body*3

	return hasSmallBody && hasLongShadows
}

// =============================================================================
// TWO CANDLE PATTERNS
// =============================================================================

// DetectEngulfing identifies bullish/bearish engulfing patterns
func DetectEngulfing(prev, current OHLCData, minSize float64) (bool, bool) {
	if !validateOHLCData(prev) || !validateOHLCData(current) {
		return false, false
	}
	if minSize <= 0 {
		minSize = 0.8 // 80% default
	}

	prevBody := math.Abs(prev.Close - prev.Open)
	currentBody := math.Abs(current.Close - current.Open)

	// Current candle must engulf previous candle's body
	prevTop := math.Max(prev.Open, prev.Close)
	prevBottom := math.Min(prev.Open, prev.Close)
	currentTop := math.Max(current.Open, current.Close)
	currentBottom := math.Min(current.Open, current.Close)

	isEngulfing := currentTop > prevTop && currentBottom < prevBottom
	isSizeSignificant := currentBody >= minSize*prevBody

	if !isEngulfing || !isSizeSignificant {
		return false, false
	}

	// Determine bullish or bearish
	prevBearish := prev.Close < prev.Open
	currentBullish := current.Close > current.Open

	bullishEngulfing := prevBearish && currentBullish
	bearishEngulfing := !prevBearish && !currentBullish

	return bullishEngulfing, bearishEngulfing
}

// DetectHarami identifies harami patterns (small candle within previous large candle)
func DetectHarami(prev, current OHLCData, minRatio float64) (bool, bool) {
	if !validateOHLCData(prev) || !validateOHLCData(current) {
		return false, false
	}
	if minRatio <= 0 {
		minRatio = 0.3 // 30% default - current body should be at least 30% smaller
	}

	prevBody := math.Abs(prev.Close - prev.Open)
	currentBody := math.Abs(current.Close - current.Open)

	// Current candle body must be significantly smaller than previous
	if currentBody >= prevBody*minRatio {
		return false, false
	}

	// Current candle must be contained within previous candle's body
	prevTop := math.Max(prev.Open, prev.Close)
	prevBottom := math.Min(prev.Open, prev.Close)
	currentTop := math.Max(current.Open, current.Close)
	currentBottom := math.Min(current.Open, current.Close)

	isContained := currentTop <= prevTop && currentBottom >= prevBottom

	if !isContained {
		return false, false
	}

	// Determine bullish or bearish harami
	prevBearish := prev.Close < prev.Open
	currentBullish := current.Close > current.Open

	bullishHarami := prevBearish && currentBullish
	bearishHarami := !prevBearish && !currentBullish

	return bullishHarami, bearishHarami
}

// DetectPiercingLine identifies piercing line patterns (bullish reversal)
func DetectPiercingLine(prev, current OHLCData) bool {
	if !validateOHLCData(prev) || !validateOHLCData(current) {
		return false
	}

	// Previous candle must be bearish
	if prev.Close >= prev.Open {
		return false
	}

	// Current candle must be bullish
	if current.Close <= current.Open {
		return false
	}

	// Current must open below previous low (gap down)
	if current.Open >= prev.Low {
		return false
	}

	// Current must close above midpoint of previous candle's body
	prevMidpoint := (prev.Open + prev.Close) / 2
	if current.Close <= prevMidpoint {
		return false
	}

	// Current close should not exceed previous open (not engulfing)
	if current.Close >= prev.Open {
		return false
	}

	return true
}

// DetectDarkCloudCover identifies dark cloud cover patterns (bearish reversal)
func DetectDarkCloudCover(prev, current OHLCData) bool {
	if !validateOHLCData(prev) || !validateOHLCData(current) {
		return false
	}

	// Previous candle must be bullish
	if prev.Close <= prev.Open {
		return false
	}

	// Current candle must be bearish
	if current.Close >= current.Open {
		return false
	}

	// Current must open above previous high (gap up)
	if current.Open <= prev.High {
		return false
	}

	// Current must close below midpoint of previous candle's body
	prevMidpoint := (prev.Open + prev.Close) / 2
	if current.Close >= prevMidpoint {
		return false
	}

	// Current close should not go below previous open (not engulfing)
	if current.Close <= prev.Open {
		return false
	}

	return true
}

// DetectTweezerTops identifies tweezer top patterns (bearish reversal)
func DetectTweezerTops(prev, current OHLCData, tolerance float64) bool {
	if !validateOHLCData(prev) || !validateOHLCData(current) {
		return false
	}
	if tolerance <= 0 {
		tolerance = 0.005 // 0.5% default tolerance
	}

	// Both candles should have similar highs (resistance level)
	priceDiff := math.Abs(prev.High - current.High)
	avgHigh := (prev.High + current.High) / 2
	if avgHigh == 0 {
		return false
	}

	similarHighs := (priceDiff / avgHigh) <= tolerance

	// First candle should be bullish, second bearish (reversal)
	prevBullish := prev.Close > prev.Open
	currentBearish := current.Close < current.Open

	return similarHighs && prevBullish && currentBearish
}

// DetectTweezerBottoms identifies tweezer bottom patterns (bullish reversal)
func DetectTweezerBottoms(prev, current OHLCData, tolerance float64) bool {
	if !validateOHLCData(prev) || !validateOHLCData(current) {
		return false
	}
	if tolerance <= 0 {
		tolerance = 0.005 // 0.5% default tolerance
	}

	// Both candles should have similar lows (support level)
	priceDiff := math.Abs(prev.Low - current.Low)
	avgLow := (prev.Low + current.Low) / 2
	if avgLow == 0 {
		return false
	}

	similarLows := (priceDiff / avgLow) <= tolerance

	// First candle should be bearish, second bullish (reversal)
	prevBearish := prev.Close < prev.Open
	currentBullish := current.Close > current.Open

	return similarLows && prevBearish && currentBullish
}

// =============================================================================
// THREE CANDLE PATTERNS
// =============================================================================

// DetectMorningStar identifies morning star patterns (3-candle bullish reversal)
func DetectMorningStar(first, second, third OHLCData, options PatternDetectionOption) bool {
	if !validateOHLCData(first) || !validateOHLCData(second) || !validateOHLCData(third) {
		return false
	}

	// First candle: bearish (long red)
	if first.Close >= first.Open {
		return false
	}
	firstBody := first.Open - first.Close

	// Second candle: small body (doji-like), gaps down
	secondBody := math.Abs(second.Close - second.Open)
	if secondBody > firstBody*0.3 { // Second body should be small
		return false
	}
	// Gap down: second candle's high should be below first candle's low
	if second.High >= first.Close {
		return false
	}

	// Third candle: bullish (long green), gaps up
	if third.Close <= third.Open {
		return false
	}
	thirdBody := third.Close - third.Open

	// Gap up: third candle's low should be above second candle's high
	if third.Open <= second.High {
		return false
	}

	// Third candle should close well into first candle's body
	if third.Close <= (first.Open+first.Close)/2 {
		return false
	}

	// Bodies should be reasonably sized
	if thirdBody < firstBody*0.5 {
		return false
	}

	return true
}

// DetectEveningStar identifies evening star patterns (3-candle bearish reversal)
func DetectEveningStar(first, second, third OHLCData, options PatternDetectionOption) bool {
	if !validateOHLCData(first) || !validateOHLCData(second) || !validateOHLCData(third) {
		return false
	}

	// First candle: bullish (long green)
	if first.Close <= first.Open {
		return false
	}
	firstBody := first.Close - first.Open

	// Second candle: small body (doji-like), gaps up
	secondBody := math.Abs(second.Close - second.Open)
	if secondBody > firstBody*0.3 { // Second body should be small
		return false
	}
	// Gap up: second candle's low should be above first candle's high
	if second.Low <= first.Close {
		return false
	}

	// Third candle: bearish (long red), gaps down
	if third.Close >= third.Open {
		return false
	}
	thirdBody := third.Open - third.Close

	// Gap down: third candle's high should be below second candle's low
	if third.Open >= second.Low {
		return false
	}

	// Third candle should close well into first candle's body
	if third.Close >= (first.Open+first.Close)/2 {
		return false
	}

	// Bodies should be reasonably sized
	if thirdBody < firstBody*0.5 {
		return false
	}

	return true
}

// DetectThreeWhiteSoldiers identifies three white soldiers patterns (strong bullish trend)
func DetectThreeWhiteSoldiers(first, second, third OHLCData) bool {
	if !validateOHLCData(first) || !validateOHLCData(second) || !validateOHLCData(third) {
		return false
	}

	// All three candles must be bullish
	if first.Close <= first.Open || second.Close <= second.Open || third.Close <= third.Open {
		return false
	}

	// Each candle should close higher than the previous
	if second.Close <= first.Close || third.Close <= second.Close {
		return false
	}

	// Each candle should open within or above the previous body
	if second.Open < first.Open || third.Open < second.Open {
		return false
	}

	// Bodies should be reasonably sized (not dojis)
	firstBody := first.Close - first.Open
	secondBody := second.Close - second.Open
	thirdBody := third.Close - third.Open

	avgBody := (firstBody + secondBody + thirdBody) / 3
	totalRange := (first.High - first.Low + second.High - second.Low + third.High - third.Low) / 3

	if totalRange == 0 {
		return false
	}

	// Bodies should be at least 30% of the average range
	if avgBody/totalRange < 0.3 {
		return false
	}

	// Shadows should be relatively small
	maxUpperShadow := math.Max(math.Max(first.High-first.Close, second.High-second.Close), third.High-third.Close)
	maxLowerShadow := math.Max(math.Max(first.Open-first.Low, second.Open-second.Low), third.Open-third.Low)

	if maxUpperShadow > avgBody*0.5 || maxLowerShadow > avgBody*0.5 {
		return false
	}

	return true
}

// DetectThreeBlackCrows identifies three black crows patterns (strong bearish trend)
func DetectThreeBlackCrows(first, second, third OHLCData) bool {
	if !validateOHLCData(first) || !validateOHLCData(second) || !validateOHLCData(third) {
		return false
	}

	// All three candles must be bearish
	if first.Close >= first.Open || second.Close >= second.Open || third.Close >= third.Open {
		return false
	}

	// Each candle should close lower than the previous
	if second.Close >= first.Close || third.Close >= second.Close {
		return false
	}

	// Each candle should open within or below the previous body
	if second.Open > first.Open || third.Open > second.Open {
		return false
	}

	// Bodies should be reasonably sized (not dojis)
	firstBody := first.Open - first.Close
	secondBody := second.Open - second.Close
	thirdBody := third.Open - third.Close

	avgBody := (firstBody + secondBody + thirdBody) / 3
	totalRange := (first.High - first.Low + second.High - second.Low + third.High - third.Low) / 3

	if totalRange == 0 {
		return false
	}

	// Bodies should be at least 30% of the average range
	if avgBody/totalRange < 0.3 {
		return false
	}

	// Shadows should be relatively small
	maxUpperShadow := math.Max(math.Max(first.High-math.Max(first.Open, first.Close), second.High-math.Max(second.Open, second.Close)), third.High-math.Max(third.Open, third.Close))
	maxLowerShadow := math.Max(math.Max(math.Min(first.Open, first.Close)-first.Low, math.Min(second.Open, second.Close)-second.Low), math.Min(third.Open, third.Close)-third.Low)

	if maxUpperShadow > avgBody*0.5 || maxLowerShadow > avgBody*0.5 {
		return false
	}

	return true
}

// =============================================================================
// PATTERN DETECTION FOR LABELS
// =============================================================================

// PatternDetectionResult holds information about a detected pattern at a specific index
type PatternDetectionResult struct {
	Index       int
	PatternName string
	PatternType string
}

// scanSeriesForPatterns scans entire series upfront for configured patterns (private)
func scanSeriesForPatterns(series *CandlestickSeries, config *CandlestickPatternConfig) map[int][]PatternDetectionResult {
	// Early exit if no patterns enabled
	if len(config.EnabledPatterns) == 0 {
		return nil
	}

	patternMap := make(map[int][]PatternDetectionResult)
	options := config.DetectionOptions
	for index, ohlc := range series.Data {
		if !validateOHLCData(ohlc) {
			continue
		}

		var patterns []PatternDetectionResult

		// Single candle patterns
		if isPatternEnabled(PatternDoji, config.EnabledPatterns) && DetectDoji(ohlc, options.DojiThreshold) {
			patterns = append(patterns, PatternDetectionResult{
				Index:       index,
				PatternName: "Doji",
				PatternType: PatternDoji,
			})
		}

		if isPatternEnabled(PatternHammer, config.EnabledPatterns) && DetectHammer(ohlc, options.ShadowRatio) {
			patterns = append(patterns, PatternDetectionResult{
				Index:       index,
				PatternName: "Hammer",
				PatternType: PatternHammer,
			})
		}

		if isPatternEnabled(PatternInvertedHammer, config.EnabledPatterns) && DetectInvertedHammer(ohlc, options.ShadowRatio) {
			patterns = append(patterns, PatternDetectionResult{
				Index:       index,
				PatternName: "Inverted Hammer",
				PatternType: PatternInvertedHammer,
			})
		}

		if isPatternEnabled(PatternShootingStar, config.EnabledPatterns) && DetectShootingStar(ohlc, options.ShadowRatio) {
			patterns = append(patterns, PatternDetectionResult{
				Index:       index,
				PatternName: "Shooting Star",
				PatternType: PatternShootingStar,
			})
		}

		if isPatternEnabled(PatternGravestone, config.EnabledPatterns) && DetectGravestoneDoji(ohlc, options) {
			patterns = append(patterns, PatternDetectionResult{
				Index:       index,
				PatternName: "Gravestone Doji",
				PatternType: PatternGravestone,
			})
		}

		if isPatternEnabled(PatternDragonfly, config.EnabledPatterns) && DetectDragonflyDoji(ohlc, options) {
			patterns = append(patterns, PatternDetectionResult{
				Index:       index,
				PatternName: "Dragonfly Doji",
				PatternType: PatternDragonfly,
			})
		}

		// Marubozu patterns
		if isPatternEnabled(PatternMarubozuBull, config.EnabledPatterns) || isPatternEnabled(PatternMarubozuBear, config.EnabledPatterns) {
			bullishMarubozu, bearishMarubozu := DetectMarubozu(ohlc, 0.01)
			if isPatternEnabled(PatternMarubozuBull, config.EnabledPatterns) && bullishMarubozu {
				patterns = append(patterns, PatternDetectionResult{
					Index:       index,
					PatternName: "Bullish Marubozu",
					PatternType: PatternMarubozuBull,
				})
			}
			if isPatternEnabled(PatternMarubozuBear, config.EnabledPatterns) && bearishMarubozu {
				patterns = append(patterns, PatternDetectionResult{
					Index:       index,
					PatternName: "Bearish Marubozu",
					PatternType: PatternMarubozuBear,
				})
			}
		}

		if isPatternEnabled(PatternSpinningTop, config.EnabledPatterns) && DetectSpinningTop(ohlc, 0.3) {
			patterns = append(patterns, PatternDetectionResult{
				Index:       index,
				PatternName: "Spinning Top",
				PatternType: PatternSpinningTop,
			})
		}

		// Two candle patterns (if index > 0)
		if index > 0 && validateOHLCData(series.Data[index-1]) {
			prev := series.Data[index-1]

			// Engulfing patterns
			if isPatternEnabled(PatternEngulfingBull, config.EnabledPatterns) || isPatternEnabled(PatternEngulfingBear, config.EnabledPatterns) {
				bullEngulfing, bearEngulfing := DetectEngulfing(prev, ohlc, options.EngulfingMinSize)
				if isPatternEnabled(PatternEngulfingBull, config.EnabledPatterns) && bullEngulfing {
					patterns = append(patterns, PatternDetectionResult{
						Index:       index,
						PatternName: "Bullish Engulfing",
						PatternType: PatternEngulfingBull,
					})
				}
				if isPatternEnabled(PatternEngulfingBear, config.EnabledPatterns) && bearEngulfing {
					patterns = append(patterns, PatternDetectionResult{
						Index:       index,
						PatternName: "Bearish Engulfing",
						PatternType: PatternEngulfingBear,
					})
				}
			}

			// Harami patterns
			if isPatternEnabled(PatternHarami, config.EnabledPatterns) {
				bullishHarami, bearishHarami := DetectHarami(prev, ohlc, 0.3)
				if bullishHarami {
					patterns = append(patterns, PatternDetectionResult{
						Index:       index,
						PatternName: "Bullish Harami",
						PatternType: PatternHarami,
					})
				}
				if bearishHarami {
					patterns = append(patterns, PatternDetectionResult{
						Index:       index,
						PatternName: "Bearish Harami",
						PatternType: PatternHarami,
					})
				}
			}

			if isPatternEnabled(PatternPiercingLine, config.EnabledPatterns) && DetectPiercingLine(prev, ohlc) {
				patterns = append(patterns, PatternDetectionResult{
					Index:       index,
					PatternName: "Piercing Line",
					PatternType: PatternPiercingLine,
				})
			}

			if isPatternEnabled(PatternDarkCloudCover, config.EnabledPatterns) && DetectDarkCloudCover(prev, ohlc) {
				patterns = append(patterns, PatternDetectionResult{
					Index:       index,
					PatternName: "Dark Cloud Cover",
					PatternType: PatternDarkCloudCover,
				})
			}

			if isPatternEnabled(PatternTweezerTop, config.EnabledPatterns) && DetectTweezerTops(prev, ohlc, 0.005) {
				patterns = append(patterns, PatternDetectionResult{
					Index:       index,
					PatternName: "Tweezer Top",
					PatternType: PatternTweezerTop,
				})
			}

			if isPatternEnabled(PatternTweezerBottom, config.EnabledPatterns) && DetectTweezerBottoms(prev, ohlc, 0.005) {
				patterns = append(patterns, PatternDetectionResult{
					Index:       index,
					PatternName: "Tweezer Bottom",
					PatternType: PatternTweezerBottom,
				})
			}
		}

		// Three candle patterns (if index > 1)
		if index > 1 && validateOHLCData(series.Data[index-1]) && validateOHLCData(series.Data[index-2]) {
			prev := series.Data[index-1]
			prevPrev := series.Data[index-2]

			if isPatternEnabled(PatternMorningStar, config.EnabledPatterns) && DetectMorningStar(prevPrev, prev, ohlc, options) {
				patterns = append(patterns, PatternDetectionResult{
					Index:       index,
					PatternName: "Morning Star",
					PatternType: PatternMorningStar,
				})
			}

			if isPatternEnabled(PatternEveningStar, config.EnabledPatterns) && DetectEveningStar(prevPrev, prev, ohlc, options) {
				patterns = append(patterns, PatternDetectionResult{
					Index:       index,
					PatternName: "Evening Star",
					PatternType: PatternEveningStar,
				})
			}

			if isPatternEnabled(PatternThreeWhiteSoldiers, config.EnabledPatterns) && DetectThreeWhiteSoldiers(prevPrev, prev, ohlc) {
				patterns = append(patterns, PatternDetectionResult{
					Index:       index,
					PatternName: "Three White Soldiers",
					PatternType: PatternThreeWhiteSoldiers,
				})
			}

			if isPatternEnabled(PatternThreeBlackCrows, config.EnabledPatterns) && DetectThreeBlackCrows(prevPrev, prev, ohlc) {
				patterns = append(patterns, PatternDetectionResult{
					Index:       index,
					PatternName: "Three Black Crows",
					PatternType: PatternThreeBlackCrows,
				})
			}
		}

		// Store patterns if any were found
		if len(patterns) > 0 {
			patternMap[index] = patterns
		}
	}

	return patternMap
}

// isPatternEnabled checks if a pattern is in the enabled list (private)
func isPatternEnabled(patternType string, enabledPatterns []string) bool {
	// Must have explicit patterns listed
	if len(enabledPatterns) == 0 {
		return false
	}
	// Check if pattern is in the enabled list
	for _, p := range enabledPatterns {
		if p == patternType {
			return true
		}
	}
	return false
}

// formatPatternsDefault provides default pattern formatting (private)
func formatPatternsDefault(patterns []PatternDetectionResult, seriesIndex int, theme ColorPalette) (string, *LabelStyle) {
	if len(patterns) == 0 {
		return "", nil
	}

	// Build display names and determine color
	displayNames := make([]string, len(patterns))
	var bullishCount, bearishCount, neutralCount int
	for i, pattern := range patterns {
		displayNames[i] = getPatternDisplayName(pattern.PatternName)

		// Count pattern types to determine color
		switch pattern.PatternType {
		case PatternHammer, PatternMorningStar, PatternEngulfingBull, PatternDragonfly, PatternMarubozuBull, PatternPiercingLine, PatternTweezerBottom, PatternThreeWhiteSoldiers:
			bullishCount++
		case PatternShootingStar, PatternEveningStar, PatternEngulfingBear, PatternGravestone, PatternMarubozuBear, PatternDarkCloudCover, PatternTweezerTop, PatternThreeBlackCrows:
			bearishCount++
		default: // Doji, spinning top and other neutral patterns
			neutralCount++
		}
	}

	// Determine color based on predominant pattern type
	upColor, downColor := theme.GetSeriesUpDownColors(seriesIndex)
	var color Color
	if bullishCount > bearishCount && bullishCount > neutralCount {
		color = upColor
	} else if bearishCount > bullishCount && bearishCount > neutralCount {
		color = downColor
	} else {
		if theme.IsDark() {
			color = Color{R: 100, G: 100, B: 100, A: 255} // dark gray
		} else {
			color = Color{R: 200, G: 200, B: 200, A: 255} // light gray
		}
	}

	// Use theme-appropriate background based on dark/light mode
	var backgroundColor, fontColor Color
	if theme.IsDark() {
		backgroundColor = ColorBlack.WithAlpha(180)
		fontColor = color.WithAdjustHSL(0, 0, 0.28) // Lighter for dark backgrounds
	} else {
		backgroundColor = ColorWhite.WithAlpha(180)
		fontColor = color.WithAdjustHSL(0, 0, -0.28) // Darker for light backgrounds
	}

	return strings.Join(displayNames, "\n"), &LabelStyle{
		FontStyle: FontStyle{
			FontColor: fontColor,
			FontSize:  10,
		},
		BackgroundColor: backgroundColor,
		CornerRadius:    4,
		BorderColor:     color,
		BorderWidth:     1.2,
	}
}

// scanCandlestickPatterns scans entire series for patterns and returns detected patterns (private)
func scanCandlestickPatterns(series CandlestickSeries, options ...PatternDetectionOption) []PatternDetectionResult {
	var opt PatternDetectionOption
	if len(options) > 0 {
		opt = options[0]
	}

	var patterns []PatternDetectionResult

	for i, ohlc := range series.Data {
		// Single candle patterns
		if DetectDoji(ohlc, opt.DojiThreshold) {
			patterns = append(patterns, PatternDetectionResult{
				Index:       i,
				PatternName: "Doji",
				PatternType: PatternDoji,
			})
		}

		if DetectHammer(ohlc, opt.ShadowRatio) {
			patterns = append(patterns, PatternDetectionResult{
				Index:       i,
				PatternName: "Hammer",
				PatternType: PatternHammer,
			})
		}

		if DetectInvertedHammer(ohlc, opt.ShadowRatio) {
			patterns = append(patterns, PatternDetectionResult{
				Index:       i,
				PatternName: "Inverted Hammer",
				PatternType: PatternInvertedHammer,
			})
		}

		if DetectShootingStar(ohlc, opt.ShadowRatio) {
			patterns = append(patterns, PatternDetectionResult{
				Index:       i,
				PatternName: "Shooting Star",
				PatternType: PatternShootingStar,
			})
		}

		if DetectGravestoneDoji(ohlc, opt) {
			patterns = append(patterns, PatternDetectionResult{
				Index:       i,
				PatternName: "Gravestone Doji",
				PatternType: PatternGravestone,
			})
		}

		if DetectDragonflyDoji(ohlc, opt) {
			patterns = append(patterns, PatternDetectionResult{
				Index:       i,
				PatternName: "Dragonfly Doji",
				PatternType: PatternDragonfly,
			})
		}

		// Marubozu patterns
		bullishMarubozu, bearishMarubozu := DetectMarubozu(ohlc, 0.01)
		if bullishMarubozu {
			patterns = append(patterns, PatternDetectionResult{
				Index:       i,
				PatternName: "Bullish Marubozu",
				PatternType: PatternMarubozuBull,
			})
		}
		if bearishMarubozu {
			patterns = append(patterns, PatternDetectionResult{
				Index:       i,
				PatternName: "Bearish Marubozu",
				PatternType: PatternMarubozuBear,
			})
		}

		// Spinning top pattern
		if DetectSpinningTop(ohlc, 0.3) {
			patterns = append(patterns, PatternDetectionResult{
				Index:       i,
				PatternName: "Spinning Top",
				PatternType: PatternSpinningTop,
			})
		}

		// Two candle patterns (need previous candle)
		if i > 0 {
			bullishEngulfing, bearishEngulfing := DetectEngulfing(series.Data[i-1], ohlc, opt.EngulfingMinSize)
			if bullishEngulfing {
				patterns = append(patterns, PatternDetectionResult{
					Index:       i,
					PatternName: "Bullish Engulfing",
					PatternType: PatternEngulfingBull,
				})
			}
			if bearishEngulfing {
				patterns = append(patterns, PatternDetectionResult{
					Index:       i,
					PatternName: "Bearish Engulfing",
					PatternType: PatternEngulfingBear,
				})
			}

			// Harami patterns
			bullishHarami, bearishHarami := DetectHarami(series.Data[i-1], ohlc, 0.3)
			if bullishHarami {
				patterns = append(patterns, PatternDetectionResult{
					Index:       i,
					PatternName: "Bullish Harami",
					PatternType: PatternHarami,
				})
			}
			if bearishHarami {
				patterns = append(patterns, PatternDetectionResult{
					Index:       i,
					PatternName: "Bearish Harami",
					PatternType: PatternHarami,
				})
			}

			// Piercing Line pattern
			if DetectPiercingLine(series.Data[i-1], ohlc) {
				patterns = append(patterns, PatternDetectionResult{
					Index:       i,
					PatternName: "Piercing Line",
					PatternType: PatternPiercingLine,
				})
			}

			// Dark Cloud Cover pattern
			if DetectDarkCloudCover(series.Data[i-1], ohlc) {
				patterns = append(patterns, PatternDetectionResult{
					Index:       i,
					PatternName: "Dark Cloud Cover",
					PatternType: PatternDarkCloudCover,
				})
			}

			// Tweezer patterns
			if DetectTweezerTops(series.Data[i-1], ohlc, 0.005) {
				patterns = append(patterns, PatternDetectionResult{
					Index:       i,
					PatternName: "Tweezer Top",
					PatternType: PatternTweezerTop,
				})
			}

			if DetectTweezerBottoms(series.Data[i-1], ohlc, 0.005) {
				patterns = append(patterns, PatternDetectionResult{
					Index:       i,
					PatternName: "Tweezer Bottom",
					PatternType: PatternTweezerBottom,
				})
			}
		}

		// Three candle patterns (need two previous candles)
		if i > 1 {
			// Morning Star pattern
			if DetectMorningStar(series.Data[i-2], series.Data[i-1], ohlc, opt) {
				patterns = append(patterns, PatternDetectionResult{
					Index:       i,
					PatternName: "Morning Star",
					PatternType: PatternMorningStar,
				})
			}

			// Evening Star pattern
			if DetectEveningStar(series.Data[i-2], series.Data[i-1], ohlc, opt) {
				patterns = append(patterns, PatternDetectionResult{
					Index:       i,
					PatternName: "Evening Star",
					PatternType: PatternEveningStar,
				})
			}

			// Three White Soldiers pattern
			if DetectThreeWhiteSoldiers(series.Data[i-2], series.Data[i-1], ohlc) {
				patterns = append(patterns, PatternDetectionResult{
					Index:       i,
					PatternName: "Three White Soldiers",
					PatternType: PatternThreeWhiteSoldiers,
				})
			}

			// Three Black Crows pattern
			if DetectThreeBlackCrows(series.Data[i-2], series.Data[i-1], ohlc) {
				patterns = append(patterns, PatternDetectionResult{
					Index:       i,
					PatternName: "Three Black Crows",
					PatternType: PatternThreeBlackCrows,
				})
			}
		}
	}

	return patterns
}

// getPatternDisplayName returns the pattern name with appropriate symbol.
func getPatternDisplayName(patternName string) string {
	switch patternName {
	case "Doji":
		// Current: ± (plus-minus, balance symbol)
		// Alternatives: ≈ (approximately equal), • (bullet), ∏ (product)
		return "± Doji"
	case "Hammer":
		// Current: Γ (Greek gamma, hammer shape)
		// Alternatives: Τ (Greek tau), τ (small tau), Γ (gamma)
		return "Γ Hammer"
	case "Inverted Hammer":
		// Current: Ʇ (turned T, upside-down hammer)
		return "Ʇ Inv. Hammer"
	case "Shooting Star":
		// Current: ※ (reference mark, star-like)
		// Alternatives: * (asterisk), ‣ (triangular bullet), • (bullet)
		return "※ Shooting Star"
	case "Gravestone Doji":
		// Current: † (dagger, cross symbol)
		// Alternatives: ‡ (double dagger)
		return "† Gravestone"
	case "Dragonfly Doji":
		// Current: ψ (small psi, trident-like)
		// Alternatives: Ψ (capital psi), ‡ (double dagger)
		return "ψ Dragonfly"
	case "Bullish Marubozu":
		// Current: ^ (circumflex, upward direction)
		// Alternatives: Λ (lambda), Δ (delta)
		return "^ Bull Marubozu"
	case "Bearish Marubozu":
		// Current: v (lowercase v, downward direction)
		return "v Bear Marubozu"
	case "Spinning Top":
		// Current: ◌ (dotted circle, spinning motion)
		// Alternatives: • (bullet)
		return "◌ Spinning Top"
	case "Bullish Engulfing":
		// Current: Λ (Lambda, upward V shape, engulfing)
		// Alternatives: Δ (delta)
		return "Λ Bull Engulfing"
	case "Bearish Engulfing":
		// Current: V (capital V, downward engulfing)
		// Alternatives: v (lowercase v)
		return "V Bear Engulfing"
	case "Bullish Harami":
		// Current: ʘ (bilabial click, circle with dot - containment)
		// Alternatives: • (bullet), ≈ (approximately equal), ◌ (dotted circle)
		return "ʘ Bull Harami"
	case "Bearish Harami":
		// Current: θ (small theta, circle with horizontal line - containment)
		// Alternatives: Θ (capital theta), ϴ (capital theta symbol), ◌ (dotted circle)
		return "θ Bear Harami"
	case "Morning Star":
		// Current: * (asterisk, star symbol)
		// Alternatives: ※ (reference mark), ‣ (triangular bullet), • (bullet)
		return "* Morning Star"
	case "Evening Star":
		// Current: ⁎ (low asterisk, evening star)
		// Alternatives: ※ (reference mark), ‣ (triangular bullet), • (bullet)
		return "⁎ Evening Star"
	case "Piercing Line":
		// Current: | (vertical bar)
		// Alternatives: ¦ (broken bar), ǀ (dental click)
		return "| Piercing Line"
	case "Dark Cloud Cover":
		// Current: Ξ (Xi, horizontal lines like cloud layers)
		// Alternatives: ≈ (approximately equal), ∞ (infinity)
		return "Ξ Dark Cloud"
	case "Tweezer Top":
		// Current: ‖ (double vertical line, parallel lines like tweezers)
		return "‖ Tweezer Top"
	case "Tweezer Bottom":
		// Current: ǁ (lateral click, parallel lines like tweezers)
		return "ǁ Tweezer Bottom"
	case "Three White Soldiers":
		// Current: Ш (Cyrillic Sha, three vertical lines like soldiers)
		// Alternatives: Ξ (xi)
		return "Ш Three Soldiers"
	case "Three Black Crows":
		// Current: ω (omega, three lines like crow)
		// Alternatives: Ш (capital sha), Ξ (capital xi), ш (small Cyrillic sha)
		return "ω Three Crows"
	default:
		return patternName
	}
}

// =============================================================================
// PATTERN CONFIGURATION PRESETS
// =============================================================================

// PatternsAll enables all standard patterns (replaces user labels)
func PatternsAll() *CandlestickPatternConfig {
	return &CandlestickPatternConfig{
		ReplaceSeriesLabel: true,
		EnabledPatterns: []string{
			PatternDoji, PatternHammer, PatternInvertedHammer, PatternShootingStar,
			PatternGravestone, PatternDragonfly, PatternMarubozuBull, PatternMarubozuBear,
			PatternSpinningTop, PatternEngulfingBull, PatternEngulfingBear, PatternHarami,
			PatternPiercingLine, PatternDarkCloudCover, PatternTweezerTop, PatternTweezerBottom,
			PatternMorningStar, PatternEveningStar, PatternThreeWhiteSoldiers, PatternThreeBlackCrows,
		},
		DetectionOptions: DefaultPatternOptions(),
	}
}

// PatternsAllComplement enables all standard patterns (complements user labels)
func PatternsAllComplement() *CandlestickPatternConfig {
	return &CandlestickPatternConfig{
		ReplaceSeriesLabel: false,
		EnabledPatterns: []string{
			PatternDoji, PatternHammer, PatternInvertedHammer, PatternShootingStar,
			PatternGravestone, PatternDragonfly, PatternMarubozuBull, PatternMarubozuBear,
			PatternSpinningTop, PatternEngulfingBull, PatternEngulfingBear, PatternHarami,
			PatternPiercingLine, PatternDarkCloudCover, PatternTweezerTop, PatternTweezerBottom,
			PatternMorningStar, PatternEveningStar, PatternThreeWhiteSoldiers, PatternThreeBlackCrows,
		},
		DetectionOptions: DefaultPatternOptions(),
	}
}

// PatternsImportant enables only key reversal patterns (replaces user labels)
func PatternsImportant() *CandlestickPatternConfig {
	return &CandlestickPatternConfig{
		ReplaceSeriesLabel: true,
		EnabledPatterns: []string{
			PatternEngulfingBull, PatternEngulfingBear,
			PatternMorningStar, PatternEveningStar,
			PatternHammer, PatternShootingStar,
		},
		DetectionOptions: DefaultPatternOptions(),
	}
}

// PatternsImportantComplement enables only key reversal patterns (complements user labels)
func PatternsImportantComplement() *CandlestickPatternConfig {
	return &CandlestickPatternConfig{
		ReplaceSeriesLabel: false,
		EnabledPatterns: []string{
			PatternEngulfingBull, PatternEngulfingBear,
			PatternMorningStar, PatternEveningStar,
			PatternHammer, PatternShootingStar,
		},
		DetectionOptions: DefaultPatternOptions(),
	}
}

// PatternsBullish enables only bullish patterns (replaces user labels)
func PatternsBullish() *CandlestickPatternConfig {
	return &CandlestickPatternConfig{
		ReplaceSeriesLabel: true,
		EnabledPatterns: []string{
			PatternHammer, PatternInvertedHammer, PatternDragonfly,
			PatternMarubozuBull, PatternEngulfingBull, PatternPiercingLine,
			PatternTweezerBottom, PatternMorningStar, PatternThreeWhiteSoldiers,
		},
		DetectionOptions: DefaultPatternOptions(),
	}
}

// PatternsBearish enables only bearish patterns (replaces user labels)
func PatternsBearish() *CandlestickPatternConfig {
	return &CandlestickPatternConfig{
		ReplaceSeriesLabel: true,
		EnabledPatterns: []string{
			PatternShootingStar, PatternGravestone,
			PatternMarubozuBear, PatternEngulfingBear, PatternDarkCloudCover,
			PatternTweezerTop, PatternEveningStar, PatternThreeBlackCrows,
		},
		DetectionOptions: DefaultPatternOptions(),
	}
}

// EnablePatterns returns a config for specific patterns (replaces user labels)
func EnablePatterns(patterns ...string) *CandlestickPatternConfig {
	return &CandlestickPatternConfig{
		ReplaceSeriesLabel: true,
		EnabledPatterns:    patterns,
		DetectionOptions:   DefaultPatternOptions(),
	}
}

// EnablePatternsComplement returns a config for specific patterns (complements user labels)
func EnablePatternsComplement(patterns ...string) *CandlestickPatternConfig {
	return &CandlestickPatternConfig{
		ReplaceSeriesLabel: false,
		EnabledPatterns:    patterns,
		DetectionOptions:   DefaultPatternOptions(),
	}
}

// DefaultPatternOptions returns sensible default detection options
func DefaultPatternOptions() PatternDetectionOption {
	return PatternDetectionOption{
		DojiThreshold:    0.001, // 0.1% body-to-range ratio
		ShadowRatio:      2.0,   // 2:1 shadow-to-body ratio
		EngulfingMinSize: 0.8,   // 80% minimum engulfing size
	}
}
