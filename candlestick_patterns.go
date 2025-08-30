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
	// CandlestickPatternSpinningTop represents a spinning top with a small body and long shadows, indicating market indecision.
	CandlestickPatternSpinningTop = "spinning_top"
	// CandlestickPatternLongLeggedDoji represents a doji with very long upper and lower shadows, showing extreme market indecision.
	CandlestickPatternLongLeggedDoji = "long_legged_doji"
	// CandlestickPatternHighWave represents a candle with small body and extremely long shadows, indicating high volatility and indecision.
	CandlestickPatternHighWave = "high_wave"
	// CandlestickPatternBeltHoldBull represents a bullish belt hold opening at the low with no lower shadow, showing strong buying interest.
	CandlestickPatternBeltHoldBull = "belt_hold_bull"
	// CandlestickPatternBeltHoldBear represents a bearish belt hold opening at the high with no upper shadow, showing strong selling interest.
	CandlestickPatternBeltHoldBear = "belt_hold_bear"

	/** Two candle patterns **/

	// CandlestickPatternEngulfingBull represents a bullish engulfing pattern where a large bullish candle engulfs the previous bearish candle.
	CandlestickPatternEngulfingBull = "engulfing_bull"
	// CandlestickPatternEngulfingBear represents a bearish engulfing pattern where a large bearish candle engulfs the previous bullish candle.
	CandlestickPatternEngulfingBear = "engulfing_bear"
	// CandlestickPatternHaramiBull represents a bullish harami where a small bullish candle is contained within the previous large bearish candle.
	CandlestickPatternHaramiBull = "harami_bull"
	// CandlestickPatternHaramiBear represents a bearish harami where a small bearish candle is contained within the previous large bullish candle.
	CandlestickPatternHaramiBear = "harami_bear"
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

	// DojiThreshold is the body-to-range ratio threshold for doji pattern detection.
	// Smaller values make doji detection more strict. Typical range: 0.0005-0.002 (0.05%-0.2%).
	// Default: 0.001 (0.1%).
	DojiThreshold float64

	// ShadowTolerance is the shadow-to-range ratio threshold for patterns requiring minimal shadows.
	// Used by marubozu and belt-hold patterns to determine acceptable shadow size.
	// Smaller values require cleaner candles. Typical range: 0.005-0.03 (0.5%-3%).
	// Default: 0.01 (1%).
	ShadowTolerance float64

	// BodySizeRatio is the body-to-range ratio threshold for patterns with small bodies.
	// Used by spinning tops to determine maximum acceptable body size relative to total range.
	// Smaller values require smaller bodies. Typical range: 0.2-0.4 (20%-40%).
	// Default: 0.3 (30%).
	BodySizeRatio float64

	// ShadowRatio is the minimum shadow-to-body ratio for patterns requiring long shadows.
	// Used by hammer, shooting star, and similar patterns. Higher values require longer shadows.
	// Typical range: 1.5-4.0. Default: 2.0 (shadow must be at least 2x body size).
	ShadowRatio float64

	// EngulfingMinSize is the minimum size ratio for engulfing patterns.
	// The engulfing candle body must be at least this percentage of the engulfed candle body.
	// Higher values require more complete engulfment. Typical range: 0.6-0.9 (60%-90%).
	// Default: 0.8 (80%).
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
	bodySizeRatio := c.BodySizeRatio
	if bodySizeRatio <= 0 {
		bodySizeRatio = other.BodySizeRatio
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
		DojiThreshold:       dojiThreshold,
		ShadowTolerance:     shadowTolerance,
		BodySizeRatio:       bodySizeRatio,
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

func detectDojiAt(data []OHLCData, index int, options CandlestickPatternConfig) bool {
	ohlc := data[index]
	if !validateOHLCData(ohlc) {
		return false
	}
	threshold := options.DojiThreshold
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

func detectHammerAt(data []OHLCData, index int, options CandlestickPatternConfig) bool {
	ohlc := data[index]
	if !validateOHLCData(ohlc) {
		return false
	}
	shadowRatio := options.ShadowRatio
	if shadowRatio <= 0 {
		shadowRatio = 2.0 // default
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
	shadowRatio := options.ShadowRatio
	if shadowRatio <= 0 {
		shadowRatio = 2.0
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
	shadowRatio := options.ShadowRatio
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

func detectGravestoneDojiAt(data []OHLCData, index int, options CandlestickPatternConfig) bool {
	ohlc := data[index]
	if !validateOHLCData(ohlc) {
		return false
	}

	// Must be a doji first
	threshold := options.DojiThreshold
	if threshold <= 0 {
		threshold = 0.001 // 0.1% default
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

	shadowRatio := options.ShadowRatio
	if shadowRatio <= 0 {
		shadowRatio = 2.0
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

	// Must be a doji first
	threshold := options.DojiThreshold
	if threshold <= 0 {
		threshold = 0.001 // 0.1% default
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

	shadowRatio := options.ShadowRatio
	if shadowRatio <= 0 {
		shadowRatio = 2.0
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
	threshold := options.ShadowTolerance
	if threshold <= 0 {
		threshold = 0.01 // 1% default tolerance
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

	// Determine bullish
	bullish := ohlc.Close > ohlc.Open

	return bullish
}

func detectBearishMarubozuAt(data []OHLCData, index int, options CandlestickPatternConfig) bool {
	ohlc := data[index]
	if !validateOHLCData(ohlc) {
		return false
	}
	threshold := options.ShadowTolerance
	if threshold <= 0 {
		threshold = 0.01 // 1% default tolerance
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

	// Determine bearish
	bearish := ohlc.Close < ohlc.Open

	return bearish
}

func detectSpinningTopAt(data []OHLCData, index int, options CandlestickPatternConfig) bool {
	ohlc := data[index]
	if !validateOHLCData(ohlc) {
		return false
	}
	bodyRatio := options.BodySizeRatio
	if bodyRatio <= 0 || bodyRatio > 0.5 {
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

func detectLongLeggedDojiAt(data []OHLCData, index int, options CandlestickPatternConfig) bool {
	ohlc := data[index]
	if !validateOHLCData(ohlc) {
		return false
	}

	dojiThreshold := options.DojiThreshold
	if dojiThreshold <= 0 {
		dojiThreshold = 0.001
	}
	shadowRatio := options.ShadowRatio
	if shadowRatio <= 0 {
		shadowRatio = 3.0 // Long-legged doji needs higher ratio for very long shadows
	}

	// Must be a doji first
	bodySize := math.Abs(ohlc.Close - ohlc.Open)
	priceRange := ohlc.High - ohlc.Low
	if priceRange == 0 || (bodySize/priceRange) > dojiThreshold {
		return false
	}

	// Calculate shadows
	upperShadow := ohlc.High - math.Max(ohlc.Open, ohlc.Close)
	lowerShadow := math.Min(ohlc.Open, ohlc.Close) - ohlc.Low

	// Both shadows should be long (at least shadowRatio times the body)
	// Default shadowRatio = 3.0 for long-legged
	minShadowLength := shadowRatio * bodySize

	// Both shadows must be significant relative to total range
	shadowsSignificant := upperShadow >= priceRange*0.3 && lowerShadow >= priceRange*0.3

	return upperShadow >= minShadowLength && lowerShadow >= minShadowLength && shadowsSignificant
}

func detectHighWaveAt(data []OHLCData, index int, options CandlestickPatternConfig) bool {
	ohlc := data[index]
	if !validateOHLCData(ohlc) {
		return false
	}

	shadowToBodyRatio := options.ShadowRatio
	if shadowToBodyRatio <= 0 {
		shadowToBodyRatio = 3.0 // High wave needs higher ratio for detection
	}
	bodySize := math.Abs(ohlc.Close - ohlc.Open)
	upperShadow := ohlc.High - math.Max(ohlc.Open, ohlc.Close)
	lowerShadow := math.Min(ohlc.Open, ohlc.Close) - ohlc.Low
	totalShadow := upperShadow + lowerShadow
	priceRange := ohlc.High - ohlc.Low

	if priceRange == 0 || bodySize == 0 {
		return false
	}

	// Default shadowToBodyRatio = 3.0
	// Total shadows should be at least 3x the body
	hasLongShadows := totalShadow >= shadowToBodyRatio*bodySize

	// Body should be small relative to total range (but not necessarily a doji)
	smallBody := bodySize/priceRange <= 0.25

	// Both shadows should be meaningful
	bothShadowsPresent := upperShadow > bodySize*0.5 && lowerShadow > bodySize*0.5

	return hasLongShadows && smallBody && bothShadowsPresent
}

func detectBullishBeltHoldAt(data []OHLCData, index int, options CandlestickPatternConfig) bool {
	ohlc := data[index]
	if !validateOHLCData(ohlc) {
		return false
	}

	threshold := options.ShadowTolerance
	if threshold <= 0 {
		threshold = 0.01 // 1% default shadow tolerance
	}
	bodySize := math.Abs(ohlc.Close - ohlc.Open)
	priceRange := ohlc.High - ohlc.Low

	if priceRange == 0 || bodySize == 0 {
		return false
	}

	// Body should be at least 60% of the total range
	if bodySize/priceRange < 0.6 {
		return false
	}

	upperShadow := ohlc.High - math.Max(ohlc.Open, ohlc.Close)
	lowerShadow := math.Min(ohlc.Open, ohlc.Close) - ohlc.Low

	// Bullish Belt Hold: Opens at/near low, closes at/near high
	// Minimal lower shadow, can have small upper shadow
	bullish := ohlc.Close > ohlc.Open &&
		lowerShadow <= priceRange*threshold &&
		upperShadow <= priceRange*0.2

	return bullish
}

func detectBearishBeltHoldAt(data []OHLCData, index int, options CandlestickPatternConfig) bool {
	ohlc := data[index]
	if !validateOHLCData(ohlc) {
		return false
	}

	threshold := options.ShadowTolerance
	if threshold <= 0 {
		threshold = 0.01 // 1% default shadow tolerance
	}
	bodySize := math.Abs(ohlc.Close - ohlc.Open)
	priceRange := ohlc.High - ohlc.Low

	if priceRange == 0 || bodySize == 0 {
		return false
	}

	// Body should be at least 60% of the total range
	if bodySize/priceRange < 0.6 {
		return false
	}

	upperShadow := ohlc.High - math.Max(ohlc.Open, ohlc.Close)
	lowerShadow := math.Min(ohlc.Open, ohlc.Close) - ohlc.Low

	// Bearish Belt Hold: Opens at/near high, closes at/near low
	// Minimal upper shadow, can have small lower shadow
	bearish := ohlc.Close < ohlc.Open &&
		upperShadow <= priceRange*threshold &&
		lowerShadow <= priceRange*0.2

	return bearish
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
	minSize := options.EngulfingMinSize
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
		return false
	}

	// Determine bullish
	prevBearish := prev.Close < prev.Open
	currentBullish := current.Close > current.Open

	bullishEngulfing := prevBearish && currentBullish

	return bullishEngulfing
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
	minSize := options.EngulfingMinSize
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
		return false
	}

	// Determine bearish
	prevBearish := prev.Close < prev.Open
	currentBullish := current.Close > current.Open

	bearishEngulfing := !prevBearish && !currentBullish

	return bearishEngulfing
}

func detectBullishHaramiAt(data []OHLCData, index int, options CandlestickPatternConfig) bool {
	if index < 1 {
		return false
	}
	prev := data[index-1]
	current := data[index]
	if !validateOHLCData(prev) || !validateOHLCData(current) {
		return false
	}
	minRatio := 1.0 - options.EngulfingMinSize // Inverse of engulfing size for harami containment
	if minRatio <= 0 || minRatio >= 1 {
		minRatio = 0.3 // 30% default - current body should be at least 30% smaller
	}

	prevBody := math.Abs(prev.Close - prev.Open)
	currentBody := math.Abs(current.Close - current.Open)

	// Current candle body must be significantly smaller than previous
	if currentBody >= prevBody*minRatio {
		return false
	}

	// Current candle must be contained within previous candle's body
	prevTop := math.Max(prev.Open, prev.Close)
	prevBottom := math.Min(prev.Open, prev.Close)
	currentTop := math.Max(current.Open, current.Close)
	currentBottom := math.Min(current.Open, current.Close)

	isContained := currentTop <= prevTop && currentBottom >= prevBottom

	if !isContained {
		return false
	}

	// Determine bullish harami
	prevBearish := prev.Close < prev.Open
	currentBullish := current.Close > current.Open

	bullishHarami := prevBearish && currentBullish

	return bullishHarami
}

func detectBearishHaramiAt(data []OHLCData, index int, options CandlestickPatternConfig) bool {
	if index < 1 {
		return false
	}
	prev := data[index-1]
	current := data[index]
	if !validateOHLCData(prev) || !validateOHLCData(current) {
		return false
	}
	minRatio := 1.0 - options.EngulfingMinSize // Inverse of engulfing size for harami containment
	if minRatio <= 0 || minRatio >= 1 {
		minRatio = 0.3 // 30% default - current body should be at least 30% smaller
	}

	prevBody := math.Abs(prev.Close - prev.Open)
	currentBody := math.Abs(current.Close - current.Open)

	// Current candle body must be significantly smaller than previous
	if currentBody >= prevBody*minRatio {
		return false
	}

	// Current candle must be contained within previous candle's body
	prevTop := math.Max(prev.Open, prev.Close)
	prevBottom := math.Min(prev.Open, prev.Close)
	currentTop := math.Max(current.Open, current.Close)
	currentBottom := math.Min(current.Open, current.Close)

	isContained := currentTop <= prevTop && currentBottom >= prevBottom

	if !isContained {
		return false
	}

	// Determine bearish harami
	prevBearish := prev.Close < prev.Open
	currentBullish := current.Close > current.Open

	bearishHarami := !prevBearish && !currentBullish

	return bearishHarami
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
	CandlestickPatternSpinningTop:    {"Spinning Top", detectSpinningTopAt, 1},
	CandlestickPatternLongLeggedDoji: {"Long Legged Doji", detectLongLeggedDojiAt, 1},
	CandlestickPatternHighWave:       {"High Wave", detectHighWaveAt, 1},
	CandlestickPatternBeltHoldBull:   {"Bullish Belt Hold", detectBullishBeltHoldAt, 1},
	CandlestickPatternBeltHoldBear:   {"Bearish Belt Hold", detectBearishBeltHoldAt, 1},
	// double candle patterns
	CandlestickPatternEngulfingBull:  {"Bullish Engulfing", detectBullishEngulfingAt, 2},
	CandlestickPatternEngulfingBear:  {"Bearish Engulfing", detectBearishEngulfingAt, 2},
	CandlestickPatternHaramiBull:     {"Bullish Harami", detectBullishHaramiAt, 2},
	CandlestickPatternHaramiBear:     {"Bearish Harami", detectBearishHaramiAt, 2},
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
		displayNames[i] = getPatternDisplayName(pattern.PatternName)

		// Count pattern types to determine color
		switch pattern.PatternType {
		case CandlestickPatternHammer, CandlestickPatternMorningStar, CandlestickPatternEngulfingBull, CandlestickPatternDragonfly, CandlestickPatternMarubozuBull, CandlestickPatternPiercingLine, CandlestickPatternBeltHoldBull:
			bullishCount++
		case CandlestickPatternShootingStar, CandlestickPatternEveningStar, CandlestickPatternEngulfingBear, CandlestickPatternGravestone, CandlestickPatternMarubozuBear, CandlestickPatternDarkCloudCover, CandlestickPatternBeltHoldBear:
			bearishCount++
		default: // Doji, spinning top, long legged doji, high wave and other neutral patterns
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
func getPatternDisplayName(patternName string) string {
	switch patternName {
	case "Doji":
		// Current: ± (plus-minus, balance symbol)
		// Alternatives: ≈ (approximately equal), ∏ (product)
		return "± Doji"
	case "Hammer":
		// Current: Γ (Greek gamma, hammer shape)
		// Alternatives: Τ (Greek tau), τ (small tau)
		return "Γ Hammer"
	case "Inverted Hammer":
		// Current: Ʇ (turned T, upside-down hammer)
		return "Ʇ Inv. Hammer"
	case "Shooting Star":
		// Current: ※ (reference mark, star-like)
		// Alternatives: * (asterisk), ⁎ (low asterisk), ‣ (triangular bullet), • (bullet)
		return "※ Shooting Star"
	case "Gravestone Doji":
		// Current: † (dagger, cross symbol)
		// Alternatives: ‡ (double dagger)
		return "† Gravestone"
	case "Dragonfly Doji":
		// Current: ψ (small psi, trident-like)
		// Alternatives: Ψ (capital psi), ‡ (double dagger), ◊ (geometric diamond)
		return "ψ Dragonfly"
	case "Bullish Marubozu":
		// Current: ^ (circumflex, upward direction)
		// Alternatives: Λ (lambda), Δ (delta)
		return "^ Bull Marubozu"
	case "Bearish Marubozu":
		// Current: v (lowercase v, downward direction)
		// Alternatives: V (capital v)
		return "v Bear Marubozu"
	case "Spinning Top":
		// Current: ◌ (dotted circle, spinning motion)
		// Alternatives: • (bullet)
		return "◌ Spinning Top"
	case "Long Legged Doji":
		// Current: ‡ (double dagger, perfect for doji indecision with long shadows)
		// Alternatives: ‡ (double dagger), ± (plus-minus), ∏ (product)
		return "‡ Long Legged Doji"
	case "High Wave":
		// Current: ~ (tilde, perfect for extreme volatility)
		return "~ High Wave"
	case "Bullish Belt Hold":
		// Current: [ (left bracket, opens at/near low)
		return "[ Bull Belt Hold"
	case "Bearish Belt Hold":
		// Current: ] (right bracket, opens at/near high)
		return "] Bear Belt Hold"
	case "Bullish Engulfing":
		// Current: Λ (Lambda, upward V shape, engulfing)
		// Alternatives: Δ (delta), < (less than)
		return "Λ Bull Engulfing"
	case "Bearish Engulfing":
		// Current: V (capital V, downward engulfing)
		// Alternatives: v (lowercase v), > (greater than)
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
		// Alternatives: ※ (reference mark), ⁎ (low asterisk), ‣ (triangular bullet), • (bullet)
		return "* Morning Star"
	case "Evening Star":
		// Current: ⁎ (low asterisk, evening star)
		// Alternatives: ※ (reference mark), * (asterisk), ‣ (triangular bullet), • (bullet)
		return "⁎ Evening Star"
	case "Piercing Line":
		// Current: | (vertical bar)
		// Alternatives: ǀ (dental click), ¦ (broken bar)
		return "| Piercing Line"
	case "Dark Cloud Cover":
		// Current: Ξ (Xi, horizontal lines like cloud layers)
		// Alternatives: ≈ (approximately equal), ∞ (infinity), ~ (tilde)
		return "Ξ Dark Cloud"
	default:
		return patternName
	}
}

// =============================================================================
// PATTERN CONFIGURATION PRESETS
// =============================================================================

// PatternsAll enables all standard patterns.
func PatternsAll() *CandlestickPatternConfig {
	return &CandlestickPatternConfig{
		PreferPatternLabels: true,
		EnabledPatterns: []string{
			// Strong reversal patterns (highest priority)
			CandlestickPatternEngulfingBull, CandlestickPatternEngulfingBear, CandlestickPatternHammer, CandlestickPatternMorningStar,
			CandlestickPatternEveningStar, CandlestickPatternShootingStar,
			// Moderate patterns
			CandlestickPatternDarkCloudCover, CandlestickPatternDragonfly, CandlestickPatternGravestone, CandlestickPatternMarubozuBear,
			CandlestickPatternMarubozuBull, CandlestickPatternPiercingLine, CandlestickPatternBeltHoldBull, CandlestickPatternBeltHoldBear,
			// Neutral/indecision patterns
			CandlestickPatternDoji, CandlestickPatternHaramiBear, CandlestickPatternHaramiBull, CandlestickPatternInvertedHammer, CandlestickPatternSpinningTop, CandlestickPatternLongLeggedDoji, CandlestickPatternHighWave,
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
			CandlestickPatternBeltHoldBull, CandlestickPatternBeltHoldBear, // Strong directional, clear conviction
			CandlestickPatternMorningStar, CandlestickPatternEveningStar, // Multi-candle reversal confirmation
		},
	}
}

// PatternsBullish enables only bullish patterns.
func PatternsBullish() *CandlestickPatternConfig {
	return &CandlestickPatternConfig{
		PreferPatternLabels: true,
		EnabledPatterns: []string{
			CandlestickPatternHammer, CandlestickPatternInvertedHammer, CandlestickPatternDragonfly, CandlestickPatternMarubozuBull,
			CandlestickPatternEngulfingBull, CandlestickPatternHaramiBull, CandlestickPatternPiercingLine,
			CandlestickPatternMorningStar, CandlestickPatternBeltHoldBull,
		},
	}
}

// PatternsBearish enables only bearish patterns.
func PatternsBearish() *CandlestickPatternConfig {
	return &CandlestickPatternConfig{
		PreferPatternLabels: true,
		EnabledPatterns: []string{
			CandlestickPatternShootingStar, CandlestickPatternGravestone, CandlestickPatternMarubozuBear, CandlestickPatternEngulfingBear,
			CandlestickPatternHaramiBear, CandlestickPatternDarkCloudCover, CandlestickPatternEveningStar, CandlestickPatternBeltHoldBear,
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

// PatternsIndecision enables patterns that signal market indecision or volatility.
func PatternsIndecision() *CandlestickPatternConfig {
	return &CandlestickPatternConfig{
		PreferPatternLabels: true,
		EnabledPatterns: []string{
			CandlestickPatternLongLeggedDoji, // Extreme indecision
			CandlestickPatternHighWave,       // Extreme volatility
			CandlestickPatternDoji,           // Basic indecision
			CandlestickPatternSpinningTop,    // Moderate indecision
		},
	}
}

// PatternsTrend enables patterns that signal trend continuation.
func PatternsTrend() *CandlestickPatternConfig {
	return &CandlestickPatternConfig{
		PreferPatternLabels: true,
		EnabledPatterns: []string{
			// Strong directional patterns
			CandlestickPatternBeltHoldBull, CandlestickPatternBeltHoldBear,
			CandlestickPatternMarubozuBull, CandlestickPatternMarubozuBear,
			// Consolidation patterns
			CandlestickPatternHaramiBull, CandlestickPatternHaramiBear,
		},
	}
}
