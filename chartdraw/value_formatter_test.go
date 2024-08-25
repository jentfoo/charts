package chartdraw

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeValueFormatterWithFormat(t *testing.T) {
	d := time.Now()
	di := TimeToFloat64(d)
	df := float64(di)

	s := formatTime(d, DefaultDateFormat)
	si := formatTime(di, DefaultDateFormat)
	sf := formatTime(df, DefaultDateFormat)
	assert.Equal(t, s, si)
	assert.Equal(t, s, sf)

	sd := TimeValueFormatter(d)
	sdi := TimeValueFormatter(di)
	sdf := TimeValueFormatter(df)
	assert.Equal(t, s, sd)
	assert.Equal(t, s, sdi)
	assert.Equal(t, s, sdf)
}

func TestFloatValueFormatter(t *testing.T) {
	assert.Equal(t, "1234.00", FloatValueFormatter(1234.00))
}

func TestFloatValueFormatterWithFloat32Input(t *testing.T) {
	assert.Equal(t, "1234.00", FloatValueFormatter(float32(1234.00)))
}

func TestFloatValueFormatterWithIntegerInput(t *testing.T) {
	assert.Equal(t, "1234.00", FloatValueFormatter(1234))
}

func TestFloatValueFormatterWithInt64Input(t *testing.T) {
	assert.Equal(t, "1234.00", FloatValueFormatter(int64(1234)))
}

func TestFloatValueFormatterWithFormat(t *testing.T) {
	v := 123.456
	sv := FloatValueFormatterWithFormat(v, "%.3f")
	assert.Equal(t, "123.456", sv)
	assert.Equal(t, "123.000", FloatValueFormatterWithFormat(123, "%.3f"))
}

func TestExponentialValueFormatter(t *testing.T) {
	assert.Equal(t, "1.23e+02", ExponentialValueFormatter(123.456))
	assert.Equal(t, "1.24e+07", ExponentialValueFormatter(12421243.424))
	assert.Equal(t, "4.50e-01", ExponentialValueFormatter(0.45))
}
