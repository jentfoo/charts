package charts

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNullValue(t *testing.T) {
	t.Parallel()

	assert.Equal(t, math.MaxFloat64, GetNullValue())
}
