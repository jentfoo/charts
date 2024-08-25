package chartdraw

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRangeTranslate(t *testing.T) {
	values := []float64{1.0, 2.0, 2.5, 2.7, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0}
	r := ContinuousRange{Domain: 1000}
	r.Min, r.Max = MinMax(values...)

	// delta = ~7.0
	// value = ~5.0
	// domain = ~1000
	// 5/8 * 1000 ~=
	assert.Equal(t, 0, r.Translate(1.0))
	assert.Equal(t, 1000, r.Translate(8.0))
	assert.Equal(t, 572, r.Translate(5.0))
}
