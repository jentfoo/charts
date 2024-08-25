package chartdraw

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockValuesProvider struct {
	X []float64
	Y []float64
}

func (m mockValuesProvider) Len() int {
	return MinInt(len(m.X), len(m.Y))
}

func (m mockValuesProvider) GetValues(index int) (x, y float64) {
	if index < 0 {
		panic("negative index at GetValue()")
	}
	if index >= MinInt(len(m.X), len(m.Y)) {
		panic("index is outside the length of m.X or m.Y")
	}
	x = m.X[index]
	y = m.Y[index]
	return
}

func TestSMASeriesGetValue(t *testing.T) {
	mockSeries := mockValuesProvider{
		LinearRange(1.0, 10.0),
		LinearRange(10, 1.0),
	}
	assert.Equal(t, 10, mockSeries.Len())

	mas := &SMASeries{
		InnerSeries: mockSeries,
		Period:      10,
	}

	var yvalues []float64
	for x := 0; x < mas.Len(); x++ {
		_, y := mas.GetValues(x)
		yvalues = append(yvalues, y)
	}

	assert.Equal(t, 10.0, yvalues[0])
	assert.Equal(t, 9.5, yvalues[1])
	assert.Equal(t, 9.0, yvalues[2])
	assert.Equal(t, 8.5, yvalues[3])
	assert.Equal(t, 8.0, yvalues[4])
	assert.Equal(t, 7.5, yvalues[5])
	assert.Equal(t, 7.0, yvalues[6])
	assert.Equal(t, 6.5, yvalues[7])
	assert.Equal(t, 6.0, yvalues[8])
}

func TestSMASeriesGetLastValueWindowOverlap(t *testing.T) {
	mockSeries := mockValuesProvider{
		LinearRange(1.0, 10.0),
		LinearRange(10, 1.0),
	}
	assert.Equal(t, 10, mockSeries.Len())

	mas := &SMASeries{
		InnerSeries: mockSeries,
		Period:      15,
	}

	var yvalues []float64
	for x := 0; x < mas.Len(); x++ {
		_, y := mas.GetValues(x)
		yvalues = append(yvalues, y)
	}

	lx, ly := mas.GetLastValues()
	assert.Equal(t, 10.0, lx)
	assert.Equal(t, 5.5, ly)
	assert.Equal(t, yvalues[len(yvalues)-1], ly)
}

func TestSMASeriesGetLastValue(t *testing.T) {
	mockSeries := mockValuesProvider{
		LinearRange(1.0, 100.0),
		LinearRange(100, 1.0),
	}
	assert.Equal(t, 100, mockSeries.Len())

	mas := &SMASeries{
		InnerSeries: mockSeries,
		Period:      10,
	}

	var yvalues []float64
	for x := 0; x < mas.Len(); x++ {
		_, y := mas.GetValues(x)
		yvalues = append(yvalues, y)
	}

	lx, ly := mas.GetLastValues()
	assert.Equal(t, 100.0, lx)
	assert.Equal(t, 6.0, ly)
	assert.Equal(t, yvalues[len(yvalues)-1], ly)
}
