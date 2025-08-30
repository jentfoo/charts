package charts

import (
	"math"
	"strings"
)

const (
	/** Single candle patterns **/

	// CandlestickPatternDoji represents a doji candle where open and close prices are nearly equal, indicating market indecision.
	CandlestickPatternDoji = "doji"
	// CandlestickPatternHammer represents a hammer candle with a small body and long lower shadow, signaling potential bullish reversal.
	CandlestickPatternHammer = "hammer"
	// CandlestickPatternInvertedHammer represents an inverted hammer with a small body and long upper shadow, signaling potential bullish reversal.
	CandlestickPatternInvertedHammer = "inverted_hammer"
	// CandlestickPatternShootingStar represents a shooting star with a small body and long upper shadow, signaling potential bearish reversal.
	CandlestickPatternShootingStar = "shooting_star"
	// CandlestickPatternGravestone represents a gravestone doji with long upper shadow and no lower shadow, indicating bearish sentiment.
	CandlestickPatternGravestone = "gravestone_doji"
	// CandlestickPatternDragonfly represents a dragonfly doji with long lower shadow and no upper shadow, indicating bullish sentiment.
	CandlestickPatternDragonfly = "dragonfly_doji"
	// CandlestickPatternMarubozuBull represents a bullish marubozu with no shadows and closing at the high, showing strong buying pressure.
	CandlestickPatternMarubozuBull = "marubozu_bull"
	// CandlestickPatternMarubozuBear represents a bearish marubozu with no shadows and closing at the low, showing strong selling pressure.
	CandlestickPatternMarubozuBear = "marubozu_bear"

	/** Two candle patterns **/

	// CandlestickPatternEngulfingBull represents a bullish engulfing pattern where a large bullish candle engulfs the previous bearish candle.
	CandlestickPatternEngulfingBull = "engulfing_bull"
	// CandlestickPatternEngulfingBear represents a bearish engulfing pattern where a large bearish candle engulfs the previous bullish candle.
	CandlestickPatternEngulfingBear = "engulfing_bear"
	// CandlestickPatternPiercingLine represents a piercing line where a bullish candle closes above the midpoint of the previous bearish candle.
	CandlestickPatternPiercingLine = "piercing_line"
	// CandlestickPatternDarkCloudCover represents a dark cloud cover where a bearish candle closes below the midpoint of the previous bullish candle.
	CandlestickPatternDarkCloudCover = "dark_cloud_cover"

	/** Three candle patterns **/

	// CandlestickPatternMorningStar represents a bullish morning star pattern with a doji or small candle between two opposite-colored candles.
	CandlestickPatternMorningStar = "morning_star"
	// CandlestickPatternEveningStar represents a bearish evening star pattern with a doji or small candle between two opposite-colored candles.
	CandlestickPatternEveningStar = "evening_star"
)

// PatternFormatter allows custom formatting of detected patterns.
type PatternFormatter func(patterns []PatternDetectionResult, seriesName string, value float64) (string, *LabelStyle)

// CandlestickPatternConfig configures automatic pattern detection.
// EXPERIMENTAL: Pattern detection logic is under active development and may change in future versions.
type CandlestickPatternConfig struct {
	// PreferPatternLabels controls pattern/user label precedence
	// true = pattern labels have priority over user labels, false = user labels have priority over pattern labels
	PreferPatternLabels bool

	// PatternFormatter allows custom formatting, if nil, uses default formatting with theme colors.
	PatternFormatter PatternFormatter

	// EnabledPatterns lists specific patterns to detect
	// nil or empty = no patterns detected (PatternConfig must be set to enable)
	// ["doji", "hammer"] = only these patterns
	EnabledPatterns []string

	// Sensitivity controls how strict pattern detection is when thresholds are not manually set.
	// "strict" = fewer but more reliable patterns detected
	// "normal" = balanced detection (default if empty)
	// "loose" = more patterns detected, potentially more false positives
	// This affects automatic threshold calculation based on recent volatility (ATR).
	Sensitivity string

	// DojiThreshold is the body-to-range ratio threshold for doji pattern detection.
	// If set to 0 or unset, will be automatically calculated based on recent volatility and Sensitivity.
	// Manual setting: 0.0005-0.002 (0.05%-0.2%) for strict to loose detection.
	// Automatic: Uses ATR-based calculation adjusted by Sensitivity level.
	DojiThreshold float64

	// ShadowTolerance is the shadow-to-range ratio threshold for patterns requiring minimal shadows.
	// Used by marubozu patterns to determine acceptable shadow size.
	// If set to 0 or unset, will be automatically calculated based on volatility.
	// Manual setting: 0.005-0.03 (0.5%-3%) for strict to loose detection.
	ShadowTolerance float64

	// ShadowRatio is the minimum shadow-to-body ratio for patterns requiring long shadows.
	// Used by hammer, shooting star, and similar patterns.
	// If set to 0 or unset, will be automatically calculated based on Sensitivity.
	// Manual setting: 2.5-1.5 for strict to loose detection.
	ShadowRatio float64

	// EngulfingMinSize is the minimum size ratio for engulfing patterns.
	// The engulfing candle body must be at least this percentage of the engulfed candle body.
	// If set to 0 or unset, will be automatically calculated based on Sensitivity.
	// Manual setting: 0.9-0.6 (90%-60%) for strict to loose detection.
	EngulfingMinSize float64
}

// MergePatterns creates a new CandlestickPatternConfig by combining the enabled patterns config with another.
// It returns a union of both pattern sets with the current config taking precedence for other settings.
func (c *CandlestickPatternConfig) MergePatterns(other *CandlestickPatternConfig) *CandlestickPatternConfig {
	if c == nil && other == nil {
		return nil
	} else if c == nil {
		result := *other // Return a copy of other
		result.EnabledPatterns = make([]string, len(other.EnabledPatterns))
		copy(result.EnabledPatterns, other.EnabledPatterns)
		return &result
	} else if other == nil {
		result := *c // Return a copy of c
		result.EnabledPatterns = make([]string, len(c.EnabledPatterns))
		copy(result.EnabledPatterns, c.EnabledPatterns)
		return &result
	}

	// Create union of patterns, preserving order and avoiding duplicates
	seen := make(map[string]bool)
	var mergedPatterns []string
	// Add patterns from current config first (in order)
	for _, pattern := range c.EnabledPatterns {
		if !seen[pattern] {
			mergedPatterns = append(mergedPatterns, pattern)
			seen[pattern] = true
		}
	}
	// Add patterns from other config that aren't already present
	for _, pattern := range other.EnabledPatterns {
		if !seen[pattern] {
			mergedPatterns = append(mergedPatterns, pattern)
			seen[pattern] = true
		}
	}

	// Merge numeric configuration fields (keeping > 0 values with priority to c if both are set)
	dojiThreshold := c.DojiThreshold
	if dojiThreshold <= 0 {
		dojiThreshold = other.DojiThreshold
	}
	shadowTolerance := c.ShadowTolerance
	if shadowTolerance <= 0 {
		shadowTolerance = other.ShadowTolerance
	}
	shadowRatio := c.ShadowRatio
	if shadowRatio <= 0 {
		shadowRatio = other.ShadowRatio
	}
	engulfingMinSize := c.EngulfingMinSize
	if engulfingMinSize <= 0 {
		engulfingMinSize = other.EngulfingMinSize
	}

	return &CandlestickPatternConfig{
		PreferPatternLabels: c.PreferPatternLabels,
		EnabledPatterns:     mergedPatterns,
		PatternFormatter:    c.PatternFormatter,
		Sensitivity:         c.Sensitivity,
		DojiThreshold:       dojiThreshold,
		ShadowTolerance:     shadowTolerance,
		ShadowRatio:         shadowRatio,
		EngulfingMinSize:    engulfingMinSize,
	}
}

// PatternDetectionResult holds detected pattern information.
type PatternDetectionResult struct {
	// Index is the pattern's series data point position.
	Index int
	// PatternName is the display name (e.g., "Doji", "Hammer").
	PatternName string
	// PatternType is the identifier constant (e.g., CandlestickPatternDoji).
	PatternType string
}

// PatternsAll enables all standard patterns.
func PatternsAll() *CandlestickPatternConfig {
	return &CandlestickPatternConfig{
		PreferPatternLabels: true,
		EnabledPatterns: []string{
			// Strong reversal patterns (highest priority)
			CandlestickPatternEngulfingBull, CandlestickPatternEngulfingBear, CandlestickPatternHammer,
			CandlestickPatternMorningStar, CandlestickPatternEveningStar, CandlestickPatternShootingStar,
			// Moderate patterns
			CandlestickPatternDarkCloudCover, CandlestickPatternDragonfly, CandlestickPatternGravestone,
			CandlestickPatternMarubozuBear, CandlestickPatternMarubozuBull, CandlestickPatternPiercingLine,
			// Neutral/indecision patterns
			CandlestickPatternDoji, CandlestickPatternInvertedHammer,
		},
	}
}

// PatternsCore enables only the most reliable patterns that work well without volume.
func PatternsCore() *CandlestickPatternConfig {
	return &CandlestickPatternConfig{
		PreferPatternLabels: true,
		EnabledPatterns: []string{
			// Most reliable without volume (6-8 patterns)
			CandlestickPatternEngulfingBull, CandlestickPatternEngulfingBear, // Strong reversal, clear visual
			CandlestickPatternHammer, CandlestickPatternShootingStar, // Single bar reversal, location matters
			CandlestickPatternMorningStar, CandlestickPatternEveningStar, // Multi-candle reversal confirmation
		},
	}
}

// PatternsBullish enables only bullish patterns.
func PatternsBullish() *CandlestickPatternConfig {
	return &CandlestickPatternConfig{
		PreferPatternLabels: true,
		EnabledPatterns: []string{
			CandlestickPatternHammer, CandlestickPatternInvertedHammer, CandlestickPatternDragonfly,
			CandlestickPatternMarubozuBull, CandlestickPatternEngulfingBull, CandlestickPatternPiercingLine,
			CandlestickPatternMorningStar,
		},
	}
}

// PatternsBearish enables only bearish patterns.
func PatternsBearish() *CandlestickPatternConfig {
	return &CandlestickPatternConfig{
		PreferPatternLabels: true,
		EnabledPatterns: []string{
			CandlestickPatternShootingStar, CandlestickPatternGravestone, CandlestickPatternMarubozuBear,
			CandlestickPatternEngulfingBear, CandlestickPatternDarkCloudCover, CandlestickPatternEveningStar,
		},
	}
}

// PatternsReversal enables patterns that signal potential trend reversals.
func PatternsReversal() *CandlestickPatternConfig {
	return &CandlestickPatternConfig{
		PreferPatternLabels: true,
		EnabledPatterns: []string{
			// Single candle reversals
			CandlestickPatternHammer, CandlestickPatternShootingStar,
			CandlestickPatternDragonfly, CandlestickPatternGravestone,
			// Two candle reversals
			CandlestickPatternEngulfingBull, CandlestickPatternEngulfingBear,
			CandlestickPatternPiercingLine, CandlestickPatternDarkCloudCover,
			// Three candle reversals
			CandlestickPatternMorningStar, CandlestickPatternEveningStar,
		},
	}
}

// PatternsTrend enables patterns that signal trend continuation.
func PatternsTrend() *CandlestickPatternConfig {
	return &CandlestickPatternConfig{
		PreferPatternLabels: true,
		EnabledPatterns: []string{
			// Strong directional patterns
			CandlestickPatternMarubozuBull, CandlestickPatternMarubozuBear,
		},
	}
}

// scanForCandlestickPatterns scans entire series upfront for configured patterns (private)
func scanForCandlestickPatterns(data []OHLCData, config CandlestickPatternConfig) map[int][]PatternDetectionResult {
	if len(config.EnabledPatterns) == 0 {
		return nil
	}

	patternMap := make(map[int][]PatternDetectionResult)
	for _, patternType := range config.EnabledPatterns {
		detector, ok := patternDetectors[patternType]
		if !ok {
			continue
		}
		// Scan series for this specific pattern
		for i := detector.minCandles - 1; i < len(data); i++ {
			if detector.detectFunc(data, i, config) {
				patternMap[i] = append(patternMap[i], PatternDetectionResult{
					Index:       i,
					PatternName: detector.patternName,
					PatternType: patternType,
				})
			}
		}
	}

	return patternMap
}

// calculateATR calculates the Average True Range for volatility measurement.
// Period is typically 14 days. ATR does not require volume data.
func calculateATR(data []OHLCData, endIndex int, period int) float64 {
	if endIndex < 1 || period <= 0 {
		return 0
	}

	// Adjust period if not enough data
	if endIndex < period {
		period = endIndex
	}

	startIndex := endIndex - period + 1
	if startIndex < 1 {
		startIndex = 1 // Need at least 1 previous candle for true range
	}

	var sumTR float64
	var count int
	for i := startIndex; i <= endIndex; i++ {
		current := data[i]
		prev := data[i-1]

		if !validateOHLCData(current) || !validateOHLCData(prev) {
			continue
		}

		// True Range = max of:
		// 1. Current High - Current Low
		// 2. |Current High - Previous Close|
		// 3. |Current Low - Previous Close|
		hl := current.High - current.Low
		hc := math.Abs(current.High - prev.Close)
		lc := math.Abs(current.Low - prev.Close)

		tr := math.Max(hl, math.Max(hc, lc))
		sumTR += tr
		count++
	}

	if count == 0 {
		return 0
	}

	return sumTR / float64(count)
}

// applySensitivity adjusts a threshold based on the sensitivity setting.
func applySensitivity(baseValue float64, sensitivity string, isInverse bool) float64 {
	var multiplier float64

	switch sensitivity {
	case "strict":
		if isInverse {
			multiplier = 1.5 // Stricter = higher threshold for things like shadow ratio
		} else {
			multiplier = 0.6 // Stricter = lower threshold for things like doji
		}
	case "loose":
		if isInverse {
			multiplier = 0.7 // Looser = lower threshold for things like shadow ratio
		} else {
			multiplier = 1.5 // Looser = higher threshold for things like doji
		}
	default: // "normal" or empty
		multiplier = 1.0
	}

	return baseValue * multiplier
}

func detectDojiAt(data []OHLCData, index int, options CandlestickPatternConfig) bool {
	ohlc := data[index]
	if !validateOHLCData(ohlc) {
		return false
	}

	var threshold float64
	if options.DojiThreshold > 0 {
		threshold = options.DojiThreshold
	} else { // Use ATR-based dynamic threshold
		atr := calculateATR(data, index, 14)
		avgPrice := (ohlc.High + ohlc.Low + ohlc.Close) / 3

		if avgPrice > 0 && atr > 0 {
			// Base threshold on volatility relative to price
			baseThreshold := (atr / avgPrice) * 0.1 // 10% of ATR relative to price
			threshold = applySensitivity(baseThreshold, options.Sensitivity, false)
		} else {
			// Fallback to static default
			threshold = applySensitivity(0.001, options.Sensitivity, false)
		}
	}

	bodySize := math.Abs(ohlc.Close - ohlc.Open)
	priceRange := ohlc.High - ohlc.Low
	if priceRange == 0 {
		return false
	}
	return (bodySize / priceRange) <= threshold
}

func detectHammerAt(data []OHLCData, index int, options CandlestickPatternConfig) bool {
	ohlc := data[index]
	if !validateOHLCData(ohlc) {
		return false
	}

	var shadowRatio float64
	if options.ShadowRatio > 0 {
		shadowRatio = options.ShadowRatio
	} else { // Use sensitivity-based threshold
		shadowRatio = applySensitivity(2.0, options.Sensitivity, true)
	}

	bodySize := math.Abs(ohlc.Close - ohlc.Open)
	lowerShadow := math.Min(ohlc.Open, ohlc.Close) - ohlc.Low
	upperShadow := ohlc.High - math.Max(ohlc.Open, ohlc.Close)

	// Hammer: long lower shadow, short upper shadow, small body
	return lowerShadow >= shadowRatio*bodySize && upperShadow <= lowerShadow*0.3
}

func detectInvertedHammerAt(data []OHLCData, index int, options CandlestickPatternConfig) bool {
	ohlc := data[index]
	if !validateOHLCData(ohlc) {
		return false
	}

	var shadowRatio float64
	if options.ShadowRatio > 0 {
		shadowRatio = options.ShadowRatio
	} else { // Use sensitivity-based threshold
		shadowRatio = applySensitivity(2.0, options.Sensitivity, true)
	}

	bodySize := math.Abs(ohlc.Close - ohlc.Open)
	lowerShadow := math.Min(ohlc.Open, ohlc.Close) - ohlc.Low
	upperShadow := ohlc.High - math.Max(ohlc.Open, ohlc.Close)

	// Inverted hammer: long upper shadow, short lower shadow, small body
	return upperShadow >= shadowRatio*bodySize && lowerShadow <= upperShadow*0.3
}

func detectShootingStarAt(data []OHLCData, index int, options CandlestickPatternConfig) bool {
	ohlc := data[index]
	if !validateOHLCData(ohlc) {
		return false
	}

	var shadowRatio float64
	if options.ShadowRatio > 0 {
		shadowRatio = options.ShadowRatio
	} else { // Use sensitivity-based threshold
		shadowRatio = applySensitivity(2.0, options.Sensitivity, true)
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

func detectGravestoneDojiAt(data []OHLCData, index int, options CandlestickPatternConfig) bool {
	ohlc := data[index]
	if !validateOHLCData(ohlc) {
		return false
	}

	// Must be a doji first - apply dynamic threshold
	var threshold float64
	if options.DojiThreshold > 0 {
		threshold = options.DojiThreshold
	} else {
		atr := calculateATR(data, index, 14)
		avgPrice := (ohlc.High + ohlc.Low + ohlc.Close) / 3

		if avgPrice > 0 && atr > 0 {
			baseThreshold := (atr / avgPrice) * 0.1
			threshold = applySensitivity(baseThreshold, options.Sensitivity, false)
		} else {
			threshold = applySensitivity(0.001, options.Sensitivity, false)
		}
	}

	bodySize := math.Abs(ohlc.Close - ohlc.Open)
	priceRange := ohlc.High - ohlc.Low
	if priceRange == 0 {
		return false
	}
	if (bodySize / priceRange) > threshold {
		return false
	}

	bodyMidpoint := (ohlc.Open + ohlc.Close) / 2
	upperShadow := ohlc.High - bodyMidpoint
	lowerShadow := bodyMidpoint - ohlc.Low

	var shadowRatio float64
	if options.ShadowRatio > 0 {
		shadowRatio = options.ShadowRatio
	} else {
		shadowRatio = applySensitivity(2.0, options.Sensitivity, true)
	}

	// Gravestone doji: long upper shadow, minimal lower shadow
	hasLongUpperShadow := upperShadow >= shadowRatio*math.Abs(ohlc.Close-ohlc.Open)
	hasMinimalLowerShadow := lowerShadow <= upperShadow*0.3

	return hasLongUpperShadow && hasMinimalLowerShadow
}

func detectDragonflyDojiAt(data []OHLCData, index int, options CandlestickPatternConfig) bool {
	ohlc := data[index]
	if !validateOHLCData(ohlc) {
		return false
	}

	// Must be a doji first - apply dynamic threshold
	var threshold float64
	if options.DojiThreshold > 0 {
		threshold = options.DojiThreshold
	} else {
		atr := calculateATR(data, index, 14)
		avgPrice := (ohlc.High + ohlc.Low + ohlc.Close) / 3

		if avgPrice > 0 && atr > 0 {
			baseThreshold := (atr / avgPrice) * 0.1
			threshold = applySensitivity(baseThreshold, options.Sensitivity, false)
		} else {
			threshold = applySensitivity(0.001, options.Sensitivity, false)
		}
	}

	bodySize := math.Abs(ohlc.Close - ohlc.Open)
	priceRange := ohlc.High - ohlc.Low
	if priceRange == 0 {
		return false
	}
	if (bodySize / priceRange) > threshold {
		return false
	}

	bodyMidpoint := (ohlc.Open + ohlc.Close) / 2
	upperShadow := ohlc.High - bodyMidpoint
	lowerShadow := bodyMidpoint - ohlc.Low

	var shadowRatio float64
	if options.ShadowRatio > 0 {
		shadowRatio = options.ShadowRatio
	} else {
		shadowRatio = applySensitivity(2.0, options.Sensitivity, true)
	}

	// Dragonfly doji: long lower shadow, minimal upper shadow
	hasLongLowerShadow := lowerShadow >= shadowRatio*math.Abs(ohlc.Close-ohlc.Open)
	hasMinimalUpperShadow := upperShadow <= lowerShadow*0.3

	return hasLongLowerShadow && hasMinimalUpperShadow
}

func detectBullishMarubozuAt(data []OHLCData, index int, options CandlestickPatternConfig) bool {
	ohlc := data[index]
	if !validateOHLCData(ohlc) {
		return false
	}

	var threshold float64
	if options.ShadowTolerance > 0 {
		threshold = options.ShadowTolerance
	} else { // Use ATR-based dynamic threshold for shadow tolerance
		atr := calculateATR(data, index, 14)
		priceRange := ohlc.High - ohlc.Low

		if priceRange > 0 && atr > 0 {
			// Base threshold on volatility
			baseThreshold := (atr / priceRange) * 0.05 // 5% of ATR relative to candle range
			threshold = applySensitivity(baseThreshold, options.Sensitivity, false)
		} else {
			// Fallback to static default
			threshold = applySensitivity(0.01, options.Sensitivity, false)
		}
	}

	// Calculate shadow sizes
	upper := ohlc.High - math.Max(ohlc.Open, ohlc.Close)
	lower := math.Min(ohlc.Open, ohlc.Close) - ohlc.Low
	body := math.Abs(ohlc.Close - ohlc.Open)
	total := ohlc.High - ohlc.Low

	if total == 0 || body == 0 {
		return false
	}

	// Shadows should be minimal compared to total range
	hasMinimalShadows := (upper+lower)/total <= threshold

	if !hasMinimalShadows {
		return false
	}
	return ohlc.Close > ohlc.Open
}

func detectBearishMarubozuAt(data []OHLCData, index int, options CandlestickPatternConfig) bool {
	ohlc := data[index]
	if !validateOHLCData(ohlc) {
		return false
	}

	var threshold float64
	if options.ShadowTolerance > 0 {
		threshold = options.ShadowTolerance
	} else { // Use ATR-based dynamic threshold for shadow tolerance
		atr := calculateATR(data, index, 14)
		priceRange := ohlc.High - ohlc.Low

		if priceRange > 0 && atr > 0 {
			// Base threshold on volatility
			baseThreshold := (atr / priceRange) * 0.05 // 5% of ATR relative to candle range
			threshold = applySensitivity(baseThreshold, options.Sensitivity, false)
		} else {
			// Fallback to static default
			threshold = applySensitivity(0.01, options.Sensitivity, false)
		}
	}

	// Calculate shadow sizes
	upper := ohlc.High - math.Max(ohlc.Open, ohlc.Close)
	lower := math.Min(ohlc.Open, ohlc.Close) - ohlc.Low
	body := math.Abs(ohlc.Close - ohlc.Open)
	total := ohlc.High - ohlc.Low

	if total == 0 || body == 0 {
		return false
	}

	// Shadows should be minimal compared to total range
	hasMinimalShadows := (upper+lower)/total <= threshold

	if !hasMinimalShadows {
		return false
	}
	return ohlc.Close < ohlc.Open
}

func detectBullishEngulfingAt(data []OHLCData, index int, options CandlestickPatternConfig) bool {
	if index < 1 {
		return false
	}
	prev := data[index-1]
	current := data[index]
	if !validateOHLCData(prev) || !validateOHLCData(current) {
		return false
	}

	var minSize float64
	if options.EngulfingMinSize > 0 {
		minSize = options.EngulfingMinSize
	} else { // Use sensitivity-based threshold
		minSize = applySensitivity(0.8, options.Sensitivity, true)
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
		return false
	}

	// Determine bullish
	prevBearish := prev.Close < prev.Open
	currentBullish := current.Close > current.Open

	return prevBearish && currentBullish
}

func detectBearishEngulfingAt(data []OHLCData, index int, options CandlestickPatternConfig) bool {
	if index < 1 {
		return false
	}
	prev := data[index-1]
	current := data[index]
	if !validateOHLCData(prev) || !validateOHLCData(current) {
		return false
	}

	var minSize float64
	if options.EngulfingMinSize > 0 {
		minSize = options.EngulfingMinSize
	} else { // Use sensitivity-based threshold
		minSize = applySensitivity(0.8, options.Sensitivity, true)
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
		return false
	}

	// Determine bearish
	prevBullish := prev.Close > prev.Open
	currentBearish := current.Close < current.Open

	return prevBullish && currentBearish
}

func detectPiercingLineAt(data []OHLCData, index int, _ CandlestickPatternConfig) bool {
	if index < 1 {
		return false
	}
	prev := data[index-1]
	current := data[index]
	if !validateOHLCData(prev) || !validateOHLCData(current) {
		return false
	} else if prev.Close >= prev.Open { // Previous candle must be bearish
		return false
	} else if current.Close <= current.Open { // Current candle must be bullish
		return false
	} else if current.Open >= prev.Close { // Current must open below previous close (gap down)
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

func detectDarkCloudCoverAt(data []OHLCData, index int, _ CandlestickPatternConfig) bool {
	if index < 1 {
		return false
	}
	prev := data[index-1]
	current := data[index]
	if !validateOHLCData(prev) || !validateOHLCData(current) {
		return false
	} else if prev.Close <= prev.Open { // Previous candle must be bullish
		return false
	} else if current.Close >= current.Open { // Current candle must be bearish
		return false
	} else if current.Open <= prev.Close { // Current must open above previous close (gap up)
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

func detectMorningStarAt(data []OHLCData, index int, _ CandlestickPatternConfig) bool {
	if index < 2 {
		return false
	}
	first := data[index-2]
	second := data[index-1]
	third := data[index]
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
	// Gap down: second candle should open below first candle's body (allow some overlap)
	if second.Open >= first.Close {
		return false
	}
	// Third candle: bullish (long green), gaps up
	if third.Close <= third.Open {
		return false
	}
	thirdBody := third.Close - third.Open
	// Gap up: third candle should open above second candle's body (allow some overlap)
	if third.Open <= math.Max(second.Open, second.Close) {
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

func detectEveningStarAt(data []OHLCData, index int, _ CandlestickPatternConfig) bool {
	if index < 2 {
		return false
	}
	first := data[index-2]
	second := data[index-1]
	third := data[index]
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
	// Gap up: second candle should open above first candle's body (allow some overlap)
	if second.Open <= first.Close {
		return false
	}
	// Third candle: bearish (long red), gaps down
	if third.Close >= third.Open {
		return false
	}
	thirdBody := third.Open - third.Close
	// Gap down: third candle should open below second candle's body (allow some overlap)
	if third.Open >= math.Min(second.Open, second.Close) {
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

// patternDetector defines a single pattern detection function with metadata.
type patternDetector struct {
	patternName string
	detectFunc  func([]OHLCData, int, CandlestickPatternConfig) bool
	minCandles  int
}

// patternDetectors contains all available pattern detectors organized by type
var patternDetectors = map[string]patternDetector{
	// single candle patterns
	CandlestickPatternDoji:           {"Doji", detectDojiAt, 1},
	CandlestickPatternHammer:         {"Hammer", detectHammerAt, 1},
	CandlestickPatternInvertedHammer: {"Inverted Hammer", detectInvertedHammerAt, 1},
	CandlestickPatternShootingStar:   {"Shooting Star", detectShootingStarAt, 1},
	CandlestickPatternGravestone:     {"Gravestone Doji", detectGravestoneDojiAt, 1},
	CandlestickPatternDragonfly:      {"Dragonfly Doji", detectDragonflyDojiAt, 1},
	CandlestickPatternMarubozuBull:   {"Bullish Marubozu", detectBullishMarubozuAt, 1},
	CandlestickPatternMarubozuBear:   {"Bearish Marubozu", detectBearishMarubozuAt, 1},
	// double candle patterns
	CandlestickPatternEngulfingBull:  {"Bullish Engulfing", detectBullishEngulfingAt, 2},
	CandlestickPatternEngulfingBear:  {"Bearish Engulfing", detectBearishEngulfingAt, 2},
	CandlestickPatternPiercingLine:   {"Piercing Line", detectPiercingLineAt, 2},
	CandlestickPatternDarkCloudCover: {"Dark Cloud Cover", detectDarkCloudCoverAt, 2},
	// triple candle patterns
	CandlestickPatternMorningStar: {"Morning Star", detectMorningStarAt, 3},
	CandlestickPatternEveningStar: {"Evening Star", detectEveningStarAt, 3},
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
		displayName := getPatternDisplayName(pattern.PatternType)
		if displayName == "" {
			displayName = pattern.PatternName // fallback name without icon
		}
		displayNames[i] = displayName

		// Count pattern types to determine color
		switch pattern.PatternType {
		case CandlestickPatternHammer, CandlestickPatternMorningStar, CandlestickPatternEngulfingBull, CandlestickPatternDragonfly, CandlestickPatternMarubozuBull, CandlestickPatternPiercingLine:
			bullishCount++
		case CandlestickPatternShootingStar, CandlestickPatternEveningStar, CandlestickPatternEngulfingBear, CandlestickPatternGravestone, CandlestickPatternMarubozuBear, CandlestickPatternDarkCloudCover:
			bearishCount++
		default: // Doji and other neutral patterns
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

/* All symbols currently supported:
Balance: ≈
Volatility: ~
Directional: ^ v
Greek: Α Β Γ Δ Ε Ζ Η Θ Ι Κ Λ Μ Ν Ξ Ο Π Ρ Σ Τ Υ Φ Χ Ψ Ω α β γ δ ε ζ η θ ι κ λ μ ν ξ ο π ρ ς σ τ υ φ χ ψ ω ϑ ϕ ϖ ϗ Ϙ ϙ Ϛ ϛ Ϝ ϝ Ϟ ϟ Ϡ ϡ ϰ ϱ ϲ ϳ ϴ ϵ ϶ Ϸ ϸ Ϲ Ϻ ϻ ϼ Ͻ Ͼ Ͽ
Geometric: ◊ ◌
Math: ∂ ∆ ∏ ∑ √ ∞ ∫ ≠ ≤ ≥
Financial: $ ¢ £ ¤ ¥ ₡ ₢ ₣ ₤ ₥ ₦ ₧ ₨ ₩ ₪ ₫ € ₭ ₮ ₯ ₰ ₱ ₲ ₳ ₴ ₵ ₶ ₷ ₸ ₹ ₺ ₻ ₼ ₽ ₾ ₿
Star: * ※ ⁎
Special: ! " # % & ' ( ) + , - . / : ; < = > ? @ K V [ \ ] _ ` { | } ¡ ¦ § ¨ © ª « ¬ ® ¯ ° ± ² ³ ´ µ ¶ · ¸ ¹ º » ¼ ½ ¾ ¿ Å × ÷ ƒ ǀ ʘ ˆ ˇ ˘ ˙ ˚ ˛ ˜ ˝ – — † ‡ • ‣ ‰ ‱ ′ ″ ‴ ‵ ‶ ‷ ‸ ‹ › ‼ ‽ ⁂ ⁄ ⁅ ⁆ ⁇ ⁈ ⁉ ⁊ ⁋ ⁌ ⁍ ⁏ ℀ ℁ ℂ ℃ ℄ ℅ ℆ ℇ ℈ ℉ ℊ ℋ ℌ ℍ ℎ ℏ ℐ ℑ ℒ ℓ ℔ ℕ № ℗ ℘ ℙ ℚ ℛ ℜ ℝ ℞ ℟ ℠ ℡ ™ ℣ ℤ ℥ ℧ ℨ ℩ ℬ ℭ ℮ ℯ ℰ ℱ Ⅎ ℳ ℴ ℵ ℶ ℷ ℸ ℹ ℺ ℻ ℼ ℽ ℾ ℿ ⅀ ⅁ ⅂ ⅃ ⅄ ⅅ ⅆ ⅇ ⅈ ⅉ ⅊ ⅋ ⅌ ⅍ ⅎ ⅏ Ʇ
*/

// getPatternDisplayName returns the pattern name with appropriate symbol.
func getPatternDisplayName(patternType string) string {
	switch patternType {
	case CandlestickPatternDoji:
		// Current: ± (plus-minus, balance symbol)
		// Alternatives: ≈ (approximately equal), ∏ (product)
		return "± Doji"
	case CandlestickPatternHammer:
		// Current: Γ (Greek gamma, hammer shape)
		// Alternatives: Τ (Greek tau), τ (small tau)
		return "Γ Hammer"
	case CandlestickPatternInvertedHammer:
		// Current: Ʇ (turned T, upside-down hammer)
		return "Ʇ Inv. Hammer"
	case CandlestickPatternShootingStar:
		// Current: ※ (reference mark, star-like)
		// Alternatives: * (asterisk), ⁎ (low asterisk), ‣ (triangular bullet), • (bullet)
		return "※ Shooting Star"
	case CandlestickPatternGravestone:
		// Current: † (dagger, cross symbol)
		// Alternatives: ‡ (double dagger)
		return "† Gravestone"
	case CandlestickPatternDragonfly:
		// Current: ψ (small psi, trident-like)
		// Alternatives: Ψ (capital psi), ‡ (double dagger), ◊ (geometric diamond)
		return "ψ Dragonfly"
	case CandlestickPatternMarubozuBull:
		// Current: ^ (circumflex, upward direction)
		// Alternatives: Λ (lambda), Δ (delta)
		return "^ Bull Marubozu"
	case CandlestickPatternMarubozuBear:
		// Current: v (lowercase v, downward direction)
		// Alternatives: V (capital v)
		return "v Bear Marubozu"
	case CandlestickPatternEngulfingBull:
		// Current: Λ (Lambda, upward V shape, engulfing)
		// Alternatives: Δ (delta), < (less than)
		return "Λ Bull Engulfing"
	case CandlestickPatternEngulfingBear:
		// Current: V (capital V, downward engulfing)
		// Alternatives: v (lowercase v), > (greater than)
		return "V Bear Engulfing"
	case CandlestickPatternMorningStar:
		// Current: * (asterisk, star symbol)
		// Alternatives: ※ (reference mark), ⁎ (low asterisk), ‣ (triangular bullet), • (bullet)
		return "* Morning Star"
	case CandlestickPatternEveningStar:
		// Current: ⁎ (low asterisk, evening star)
		// Alternatives: ※ (reference mark), * (asterisk), ‣ (triangular bullet), • (bullet)
		return "⁎ Evening Star"
	case CandlestickPatternPiercingLine:
		// Current: | (vertical bar)
		// Alternatives: ǀ (dental click), ¦ (broken bar)
		return "| Piercing Line"
	case CandlestickPatternDarkCloudCover:
		// Current: Ξ (Xi, horizontal lines like cloud layers)
		// Alternatives: ≈ (approximately equal), ∞ (infinity), ~ (tilde)
		return "Ξ Dark Cloud"
	default:
		return ""
	}
}
