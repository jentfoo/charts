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
