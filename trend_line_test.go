package charts

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrendLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		render func(*Painter) ([]byte, error)
		result string
	}{
		{
			name: "linear",
			render: func(p *Painter) ([]byte, error) {
				trendLine := newTrendLinePainter(p)
				axisRange := newTestRange(p.Height(), 6, 0.0, 10.0, 0.0, 0.0)
				xValues := []int{50, 150, 250, 350, 450, 550}
				trend := SeriesTrendLine{
					Type: SeriesTrendTypeLinear,
				}
				trendLine.add(trendLineRenderOption{
					defaultStrokeColor: ColorBlack,
					xValues:            xValues,
					seriesValues:       []float64{1, 2, 3, 4, 5, 6},
					axisRange:          axisRange,
					trends:             []SeriesTrendLine{trend},
				})
				if _, err := trendLine.Render(); err != nil {
					return nil, err
				}
				return p.Bytes()
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 70 344\nL 170 308\nL 270 272\nL 370 236\nL 470 200\nL 570 164\" style=\"stroke-width:2;stroke:black;fill:none\"/></svg>",
		},
		{
			name: "cubic",
			render: func(p *Painter) ([]byte, error) {
				trendLine := newTrendLinePainter(p)
				axisRange := newTestRange(p.Height(), 6, 0.0, 40.0, 0.0, 0.0)
				xValues := []int{50, 150, 250, 350, 450, 550}
				trend := SeriesTrendLine{
					Type: SeriesTrendTypeCubic,
				}
				trendLine.add(trendLineRenderOption{
					defaultStrokeColor: ColorBlack,
					xValues:            xValues,
					seriesValues:       []float64{1, 4, 9, 16, 25, 36},
					axisRange:          axisRange,
					trends:             []SeriesTrendLine{trend},
				})
				if _, err := trendLine.Render(); err != nil {
					return nil, err
				}
				return p.Bytes()
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 70 371\nL 170 345\nL 270 300\nL 370 236\nL 470 155\nL 570 57\" style=\"stroke-width:2;stroke:black;fill:none\"/></svg>",
		},
		{
			name: "average",
			render: func(p *Painter) ([]byte, error) {
				trendLine := newTrendLinePainter(p)
				axisRange := newTestRange(p.Height(), 6, 0.0, 6.0, 0.0, 0.0)
				xValues := []int{50, 150, 250, 350, 450, 550}
				trend := SeriesTrendLine{
					Type:   SeriesTrendTypeAverage,
					Window: 3,
				}
				trendLine.add(trendLineRenderOption{
					defaultStrokeColor: ColorBlack,
					xValues:            xValues,
					seriesValues:       []float64{1, 2, 3, 4, 5, 6},
					axisRange:          axisRange,
					trends:             []SeriesTrendLine{trend},
				})
				if _, err := trendLine.Render(); err != nil {
					return nil, err
				}
				return p.Bytes()
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 70 320\nL 170 290\nL 270 260\nL 370 200\nL 470 140\nL 570 80\" style=\"stroke-width:2;stroke:black;fill:none\"/></svg>",
		},
		{
			name: "sma",
			render: func(p *Painter) ([]byte, error) {
				trendLine := newTrendLinePainter(p)
				axisRange := newTestRange(p.Height(), 6, 0.0, 6.0, 0.0, 0.0)
				xValues := []int{50, 150, 250, 350, 450, 550}
				trend := SeriesTrendLine{
					Type:   SeriesTrendTypeSMA,
					Period: 3,
				}
				trendLine.add(trendLineRenderOption{
					defaultStrokeColor: ColorBlack,
					xValues:            xValues,
					seriesValues:       []float64{1, 2, 3, 4, 5, 6},
					axisRange:          axisRange,
					trends:             []SeriesTrendLine{trend},
				})
				if _, err := trendLine.Render(); err != nil {
					return nil, err
				}
				return p.Bytes()
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 70 320\nL 170 290\nL 270 260\nL 370 200\nL 470 140\nL 570 80\" style=\"stroke-width:2;stroke:black;fill:none\"/></svg>",
		},
		{
			name: "ema",
			render: func(p *Painter) ([]byte, error) {
				trendLine := newTrendLinePainter(p)
				axisRange := newTestRange(p.Height(), 6, 0.0, 5.0, 0.0, 0.0)
				xValues := []int{50, 150, 250, 350, 450, 550}
				trend := SeriesTrendLine{
					Type:   SeriesTrendTypeEMA,
					Period: 3,
				}
				trendLine.add(trendLineRenderOption{
					defaultStrokeColor: ColorBlack,
					xValues:            xValues,
					seriesValues:       []float64{1, 2, 3, 4, 5, 6},
					axisRange:          axisRange,
					trends:             []SeriesTrendLine{trend},
				})
				if _, err := trendLine.Render(); err != nil {
					return nil, err
				}
				return p.Bytes()
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 70 308\nL 170 272\nL 270 218\nL 370 155\nL 470 88\nL 570 20\" style=\"stroke-width:2;stroke:black;fill:none\"/></svg>",
		},
		{
			name: "bollinger_upper",
			render: func(p *Painter) ([]byte, error) {
				trendLine := newTrendLinePainter(p)
				axisRange := newTestRange(p.Height(), 6, 0.0, 10.0, 0.0, 0.0)
				xValues := []int{50, 150, 250, 350, 450, 550}
				trend := SeriesTrendLine{
					Type:   SeriesTrendTypeBollingerUpper,
					Period: 3,
				}
				trendLine.add(trendLineRenderOption{
					defaultStrokeColor: ColorBlack,
					xValues:            xValues,
					seriesValues:       []float64{1, 2, 3, 4, 5, 6},
					axisRange:          axisRange,
					trends:             []SeriesTrendLine{trend},
				})
				if _, err := trendLine.Render(); err != nil {
					return nil, err
				}
				return p.Bytes()
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 70 380\nL 170 380\nL 270 250\nL 370 214\nL 470 178\nL 570 142\" style=\"stroke-width:2;stroke:black;fill:none\"/></svg>",
		},
		{
			name: "bollinger_lower",
			render: func(p *Painter) ([]byte, error) {
				trendLine := newTrendLinePainter(p)
				axisRange := newTestRange(p.Height(), 6, 0.0, 10.0, 0.0, 0.0)
				xValues := []int{50, 150, 250, 350, 450, 550}
				trend := SeriesTrendLine{
					Type:   SeriesTrendTypeBollingerLower,
					Period: 3,
				}
				trendLine.add(trendLineRenderOption{
					defaultStrokeColor: ColorBlack,
					xValues:            xValues,
					seriesValues:       []float64{1, 2, 3, 4, 5, 6},
					axisRange:          axisRange,
					trends:             []SeriesTrendLine{trend},
				})
				if _, err := trendLine.Render(); err != nil {
					return nil, err
				}
				return p.Bytes()
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 70 380\nL 170 380\nL 270 367\nL 370 331\nL 470 295\nL 570 259\" style=\"stroke-width:2;stroke:black;fill:none\"/></svg>",
		},
		{
			name: "rsi",
			render: func(p *Painter) ([]byte, error) {
				trendLine := newTrendLinePainter(p)
				axisRange := newTestRange(p.Height(), 6, 0.0, 100.0, 0.0, 0.0)
				xValues := []int{50, 150, 250, 350, 450, 550}
				trend := SeriesTrendLine{
					Type:   SeriesTrendTypeRSI,
					Period: 3,
				}
				trendLine.add(trendLineRenderOption{
					defaultStrokeColor: ColorBlack,
					xValues:            xValues,
					seriesValues:       []float64{44, 44.5, 43.8, 44.2, 44.5, 43.9},
					axisRange:          axisRange,
					trends:             []SeriesTrendLine{trend},
				})
				if _, err := trendLine.Render(); err != nil {
					return nil, err
				}
				return p.Bytes()
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 70 380\nL 170 380\nL 270 380\nL 370 178\nL 470 143\nL 570 238\" style=\"stroke-width:2;stroke:black;fill:none\"/></svg>",
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			p := NewPainter(PainterOptions{
				OutputFormat: ChartOutputSVG,
				Width:        600,
				Height:       400,
			}, PainterThemeOption(GetTheme(ThemeLight)))
			data, err := tt.render(p.Child(PainterPaddingOption(NewBoxEqual(20))))
			require.NoError(t, err)
			assertEqualSVG(t, tt.result, data)
		})
	}
}

func TestLinearTrend(t *testing.T) {
	t.Parallel()

	input := []float64{2, 4, 6, 8}
	expected := []float64{2, 4, 6, 8}

	result, err := linearTrend(input)
	require.NoError(t, err)
	require.Len(t, result, len(expected))
	for i := range expected {
		assert.InDelta(t, expected[i], result[i], 1e-9)
	}
}

func TestCubicTrend(t *testing.T) {
	t.Parallel()

	input := []float64{0, 1, 8, 27}
	expected := []float64{0, 1, 8, 27}

	result, err := cubicTrend(input)
	require.NoError(t, err)
	require.Len(t, result, len(expected))
	for i := range expected {
		assert.InDelta(t, expected[i], result[i], 1e-9)
	}
}

func TestMovingAverageTrend(t *testing.T) {
	t.Parallel()

	input := []float64{1, 2, 3, 4, 5}
	expected := []float64{1, 1.5, 2, 3, 4}

	result, err := movingAverageTrend(input, 3)
	require.NoError(t, err)
	require.Len(t, result, len(expected))
	for i := range expected {
		assert.InDelta(t, expected[i], result[i], 1e-9)
	}

	t.Run("window_larger_than_data", func(t *testing.T) {
		input := []float64{1, 2, 3, 4}
		result, err := movingAverageTrend(input, 10) // window > len(input)
		require.NoError(t, err)
		assert.Len(t, result, len(input))
	})

	t.Run("massive_window", func(t *testing.T) {
		input := []float64{1, 2, 3, 4, 5}
		result, err := movingAverageTrend(input, 1000)
		require.NoError(t, err)
		assert.Len(t, result, len(input))
	})
}

func TestExponentialMovingAverageTrend(t *testing.T) {
	t.Parallel()

	values := []float64{1, 2, 3, 4, 5}
	result, err := exponentialMovingAverageTrend(values, 3)

	require.NoError(t, err)
	require.Len(t, result, 5)

	// First value should equal input
	assert.InDelta(t, 1.0, result[0], 0.001)

	// EMA should be calculated with smoothing factor 2/(3+1) = 0.5
	multiplier := 2.0 / 4.0
	expected := (2.0 * multiplier) + (1.0 * (1 - multiplier))
	assert.InDelta(t, expected, result[1], 0.001)
}

func TestSolveLinearSystem(t *testing.T) {
	t.Parallel()

	mat := [][]float64{
		{0, 1, 0, 0, 2},
		{1, 0, 0, 0, 1},
		{0, 0, 1, 0, 3},
		{0, 0, 0, 1, 4},
	}
	expected := []float64{1, 2, 3, 4}

	result, err := solveLinearSystem(mat)
	require.NoError(t, err)
	require.Len(t, result, len(expected))
	for i := range expected {
		assert.InDelta(t, expected[i], result[i], 1e-9)
	}
}

func TestBollingerUpperTrend(t *testing.T) {
	t.Parallel()

	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	result, err := bollingerUpperTrend(values, 3)

	require.NoError(t, err)
	require.Len(t, result, 10)

	// First two values should be null
	assert.InDelta(t, GetNullValue(), result[0], 0.001)
	assert.InDelta(t, GetNullValue(), result[1], 0.001)

	// Upper band should be greater than SMA
	sma, err := movingAverageTrend(values, 3)
	require.NoError(t, err)
	for i := 2; i < len(result); i++ {
		assert.Greater(t, result[i], sma[i])
	}
}

func TestBollingerLowerTrend(t *testing.T) {
	t.Parallel()

	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	result, err := bollingerLowerTrend(values, 3)

	require.NoError(t, err)
	require.Len(t, result, 10)

	// First two values should be null
	assert.InDelta(t, GetNullValue(), result[0], 0.001)
	assert.InDelta(t, GetNullValue(), result[1], 0.001)

	// Lower band should be less than SMA
	sma, err := movingAverageTrend(values, 3)
	require.NoError(t, err)
	for i := 2; i < len(result); i++ {
		assert.Less(t, result[i], sma[i])
	}
}

func TestRsiTrend(t *testing.T) {
	t.Parallel()

	// Create test data with known gains/losses
	values := []float64{44, 44.5, 43.8, 44.2, 44.5, 43.9, 44.5, 44.9, 44.5, 44.8}
	result, err := rsiTrend(values, 3)

	require.NoError(t, err)
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
