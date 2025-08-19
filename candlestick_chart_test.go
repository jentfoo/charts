package charts

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeBasicCandlestickData() []OHLCData {
	return []OHLCData{
		{Open: 100, High: 110, Low: 95, Close: 105},
		{Open: 105, High: 115, Low: 100, Close: 112},
		{Open: 112, High: 118, Low: 108, Close: 115},
		{Open: 115, High: 120, Low: 105, Close: 108}, // bearish
		{Open: 108, High: 113, Low: 105, Close: 109},
	}
}

func makeBasicCandlestickChartOption() CandlestickChartOption {
	return CandlestickChartOption{
		Title: TitleOption{
			Text: "Candlestick Chart",
		},
		Padding: NewBoxEqual(10),
		XAxis: XAxisOption{
			Labels: []string{"Jan", "Feb", "Mar", "Apr", "May"},
		},
		YAxis: make([]YAxisOption, 1),
		Legend: LegendOption{
			SeriesNames: []string{"Price"},
		},
		Series: NewSeriesCandlestick(makeBasicCandlestickData()),
	}
}

func makeMinimalCandlestickChartOption() CandlestickChartOption {
	return CandlestickChartOption{
		Padding: NewBoxEqual(10),
		XAxis: XAxisOption{
			Labels: []string{"1", "2", "3", "4", "5"},
			Show:   Ptr(false),
		},
		YAxis:  make([]YAxisOption, 1),
		Series: NewSeriesCandlestick(makeBasicCandlestickData()),
	}
}

func TestCandlestickChart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		makeOptions func() CandlestickChartOption
		expectedSVG string
	}{
		{
			name:        "basic",
			makeOptions: makeBasicCandlestickChartOption,
			expectedSVG: "",
		},
		{
			name:        "minimal",
			makeOptions: makeMinimalCandlestickChartOption,
			expectedSVG: "",
		},
		{
			name: "traditional",
			makeOptions: func() CandlestickChartOption {
				opt := makeBasicCandlestickChartOption()
				opt.Series.CandleStyle = CandleStyleTraditional
				return opt
			},
			expectedSVG: "",
		},
		{
			name: "outline",
			makeOptions: func() CandlestickChartOption {
				opt := makeBasicCandlestickChartOption()
				opt.Series.CandleStyle = CandleStyleOutline
				return opt
			},
			expectedSVG: "",
		},
		{
			name: "no_wicks",
			makeOptions: func() CandlestickChartOption {
				opt := makeBasicCandlestickChartOption()
				opt.Series.ShowWicks = Ptr(false)
				return opt
			},
			expectedSVG: "",
		},
		{
			name: "custom_style",
			makeOptions: func() CandlestickChartOption {
				opt := makeBasicCandlestickChartOption()
				opt.CandleWidth = 0.5
				opt.WickWidth = 2.0
				opt.UpColor = ColorGreen
				opt.DownColor = ColorRed
				return opt
			},
			expectedSVG: "",
		},
		{
			name: "marks",
			makeOptions: func() CandlestickChartOption {
				opt := makeBasicCandlestickChartOption()
				opt.Series.Label.Show = Ptr(true)
				opt.Series.MarkLine = NewMarkLine("min", "max")
				opt.Series.MarkLine.ValueFormatter = func(f float64) string { return fmt.Sprintf("%.2f", f) }
				opt.Series.MarkPoint = NewMarkPoint("min", "max")
				opt.Series.MarkPoint.ValueFormatter = func(f float64) string { return fmt.Sprintf("%.2f", f) }
				return opt
			},
			expectedSVG: "",
		},
		{
			name: "doji",
			makeOptions: func() CandlestickChartOption {
				opt := makeBasicCandlestickChartOption()
				data := makeBasicCandlestickData()
				if len(data) > 0 {
					data[0] = OHLCData{Open: 100, High: 110, Low: 95, Close: 100}
				}
				opt.Series = NewSeriesCandlestick(data)
				return opt
			},
			expectedSVG: "",
		},
		{
			name: "dual_axis",
			makeOptions: func() CandlestickChartOption {
				opt := makeBasicCandlestickChartOption()
				opt.YAxis = append(opt.YAxis, YAxisOption{})
				opt.Series.YAxisIndex = 1
				return opt
			},
			expectedSVG: "",
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			painterOptions := PainterOptions{
				OutputFormat: ChartOutputSVG,
				Width:        600,
				Height:       400,
			}
			p := NewPainter(painterOptions)
			opt := tt.makeOptions()

			err := p.CandlestickChart(opt)
			require.NoError(t, err)
			data, err := p.Bytes()
			require.NoError(t, err)
			assertEqualSVG(t, tt.expectedSVG, data)
		})
	}
}

func TestCandlestickChartError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		makeOptions      func() CandlestickChartOption
		errorMsgContains string
	}{
		{
			name: "empty_series",
			makeOptions: func() CandlestickChartOption {
				return NewCandlestickChartOptionWithSeries(CandlestickSeries{})
			},
			errorMsgContains: "empty series data",
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			painterOptions := PainterOptions{
				OutputFormat: ChartOutputSVG,
				Width:        600,
				Height:       400,
			}
			p := NewPainter(painterOptions)
			opt := tt.makeOptions()

			err := p.CandlestickChart(opt)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsgContains)
		})
	}
}

func TestOHLCDataValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ohlc     OHLCData
		expected bool
	}{
		{
			name:     "valid_bullish",
			ohlc:     OHLCData{Open: 100, High: 110, Low: 95, Close: 105},
			expected: true,
		},
		{
			name:     "valid_bearish",
			ohlc:     OHLCData{Open: 110, High: 115, Low: 100, Close: 105},
			expected: true,
		},
		{
			name:     "valid_doji",
			ohlc:     OHLCData{Open: 100, High: 105, Low: 95, Close: 100},
			expected: true,
		},
		{
			name:     "invalid_high_too_low",
			ohlc:     OHLCData{Open: 100, High: 98, Low: 95, Close: 105},
			expected: false,
		},
		{
			name:     "invalid_low_too_high",
			ohlc:     OHLCData{Open: 100, High: 110, Low: 102, Close: 105},
			expected: false,
		},
		{
			name:     "invalid_null_values",
			ohlc:     OHLCData{Open: GetNullValue(), High: 110, Low: 95, Close: 105},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, validateOHLCData(tt.ohlc))
		})
	}
}

func TestCandlestickSeriesListMethods(t *testing.T) {
	t.Parallel()

	data := makeBasicCandlestickData()
	series := NewSeriesCandlestick(data)
	series.Name = "Test Series"

	assert.Equal(t, 1, series.len())
	assert.Equal(t, "Test Series", series.names()[0])
	assert.Equal(t, len(data), len(series.Data))
	assert.Equal(t, string(SymbolSquare), string(series.getSeriesSymbol(0)))

	// Test ToGenericSeriesList
	genericList := series.ToGenericSeriesList()
	assert.Len(t, genericList, 1)
	assert.Equal(t, ChartTypeCandlestick, genericList[0].Type)
	assert.Equal(t, len(data), len(genericList[0].Values))
}

func TestExtractClosePrices(t *testing.T) {
	t.Parallel()

	data := makeBasicCandlestickData()
	series := CandlestickSeries{Data: data}

	closePrices := ExtractClosePrices(series)
	expected := []float64{105, 112, 115, 108, 109}

	assert.Equal(t, expected, closePrices)
}

func TestCandlestickStyles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		style       string
		expectedSVG string
	}{
		{
			style:       CandleStyleFilled,
			expectedSVG: "",
		},
		{
			style:       CandleStyleTraditional,
			expectedSVG: "",
		},
		{
			style:       CandleStyleOutline,
			expectedSVG: "",
		},
	}
	for i, tc := range tests {
		t.Run(strconv.Itoa(i)+"-"+tc.style, func(t *testing.T) {
			opt := makeBasicCandlestickChartOption()
			opt.Series.CandleStyle = tc.style

			painterOptions := PainterOptions{
				OutputFormat: ChartOutputSVG,
				Width:        600,
				Height:       400,
			}
			p := NewPainter(painterOptions)

			err := p.CandlestickChart(opt)
			require.NoError(t, err)
			data, err := p.Bytes()
			require.NoError(t, err)
			assertEqualSVG(t, tc.expectedSVG, data)
		})
	}
}

func makePatternTestData() []OHLCData {
	return []OHLCData{
		// Normal candle
		{Open: 100, High: 110, Low: 95, Close: 105},
		// Doji (open ≈ close)
		{Open: 105, High: 108, Low: 102, Close: 105.1},
		// Hammer (long lower shadow, small body at top)
		{Open: 108, High: 109, Low: 98, Close: 107},
		// Bearish candle for engulfing setup
		{Open: 107, High: 108, Low: 103, Close: 104},
		// Bullish engulfing (engulfs previous bearish candle)
		{Open: 102, High: 115, Low: 101, Close: 113},
	}
}

func TestCandlestickWithPatterns(t *testing.T) {
	t.Parallel()

	data := makePatternTestData()
	series := NewCandlestickWithPatterns(data)

	opt := CandlestickChartOption{
		Title: TitleOption{
			Text: "Candlestick Chart with Patterns",
		},
		Padding: NewBoxEqual(10),
		XAxis: XAxisOption{
			Labels: []string{"1", "2", "3", "4", "5"},
		},
		YAxis:  make([]YAxisOption, 1),
		Series: series,
	}

	painterOptions := PainterOptions{
		OutputFormat: ChartOutputSVG,
		Width:        800,
		Height:       600,
	}
	p := NewPainter(painterOptions)

	err := p.CandlestickChart(opt)
	require.NoError(t, err)
	data2, err := p.Bytes()
	require.NoError(t, err)
	assert.Greater(t, len(data2), 100)

	// Verify patterns were detected
	assert.NotEmpty(t, series.MarkPoint.Points, "Should have detected some patterns")
}

func TestCandlestickWithSMA(t *testing.T) {
	t.Parallel()

	ohlcData := makeBasicCandlestickData()
	candlestickSeries := CandlestickSeries{Data: ohlcData}

	// Generate SMA line series
	closes := ExtractClosePrices(candlestickSeries)
	sma3 := CalculateSMA(closes, 3)

	// Create line chart overlaid with candlestick using line chart as primary
	lineSeriesList := NewSeriesListLine([][]float64{sma3})

	// Test just the line chart with SMA data
	chartOpt := ChartOption{
		SeriesList: lineSeriesList.ToGenericSeriesList(),
		Title:      TitleOption{Text: "SMA Line Chart"},
		XAxis: XAxisOption{
			Labels: []string{"Jan", "Feb", "Mar", "Apr", "May"},
		},
		YAxis: make([]YAxisOption, 1),
		Legend: LegendOption{
			SeriesNames: []string{"SMA(3)"},
		},
	}

	painter, err := Render(chartOpt)
	require.NoError(t, err)
	data, err := painter.Bytes()
	require.NoError(t, err)
	assert.Greater(t, len(data), 100)
}

func TestCandlestickWithBollingerBands(t *testing.T) {
	t.Parallel()

	// Create longer dataset for meaningful Bollinger Bands
	ohlcData := []OHLCData{
		{Open: 100, High: 110, Low: 95, Close: 105},
		{Open: 105, High: 115, Low: 100, Close: 112},
		{Open: 112, High: 118, Low: 108, Close: 115},
		{Open: 115, High: 120, Low: 110, Close: 118},
		{Open: 118, High: 125, Low: 115, Close: 122},
		{Open: 122, High: 128, Low: 119, Close: 125},
		{Open: 125, High: 130, Low: 122, Close: 127},
		{Open: 127, High: 132, Low: 124, Close: 129},
		{Open: 129, High: 135, Low: 126, Close: 131},
		{Open: 131, High: 138, Low: 128, Close: 135},
	}

	candlestickSeries := CandlestickSeries{Data: ohlcData}

	// Calculate Bollinger Bands
	closes := ExtractClosePrices(candlestickSeries)
	bands := CalculateBollingerBands(closes, 5, 2.0)

	// Test Bollinger Bands calculation separately with line chart
	lineSeriesList := NewSeriesListLine([][]float64{bands.Upper, bands.Middle, bands.Lower})

	chartOpt := ChartOption{
		SeriesList: lineSeriesList.ToGenericSeriesList(),
		Title:      TitleOption{Text: "Bollinger Bands Lines"},
		XAxis: XAxisOption{
			Labels: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
		},
		YAxis: make([]YAxisOption, 1),
		Legend: LegendOption{
			SeriesNames: []string{"BB Upper", "BB Middle", "BB Lower"},
		},
	}

	painter, err := Render(chartOpt)
	require.NoError(t, err)
	data, err := painter.Bytes()
	require.NoError(t, err)
	assert.Greater(t, len(data), 100)
}

func TestCandlestickWithEMA(t *testing.T) {
	t.Parallel()

	ohlcData := makeBasicCandlestickData()
	series := CandlestickSeries{Data: ohlcData}

	emaLines := AddEMAToKlines(series, 3, Color{})

	chartOpt := ChartOption{
		SeriesList: emaLines.ToGenericSeriesList(),
		Title:      TitleOption{Text: "EMA Line Chart"},
		XAxis: XAxisOption{
			Labels: []string{"Jan", "Feb", "Mar", "Apr", "May"},
		},
		YAxis: make([]YAxisOption, 1),
		Legend: LegendOption{
			SeriesNames: []string{"EMA(3)"},
		},
	}

	painter, err := Render(chartOpt)
	require.NoError(t, err)
	data, err := painter.Bytes()
	require.NoError(t, err)
	assert.Greater(t, len(data), 100)
}

func TestCandlestickWithRSI(t *testing.T) {
	t.Parallel()

	ohlcData := makeBasicCandlestickData()
	series := CandlestickSeries{Data: ohlcData}

	rsiLines := AddRSIToKlines(series, 3)

	chartOpt := ChartOption{
		SeriesList: rsiLines.ToGenericSeriesList(),
		Title:      TitleOption{Text: "RSI Line Chart"},
		XAxis: XAxisOption{
			Labels: []string{"Jan", "Feb", "Mar", "Apr", "May"},
		},
		YAxis: make([]YAxisOption, 1),
		Legend: LegendOption{
			SeriesNames: []string{"RSI(3)"},
		},
	}

	painter, err := Render(chartOpt)
	require.NoError(t, err)
	data, err := painter.Bytes()
	require.NoError(t, err)
	assert.Greater(t, len(data), 100)
}

func TestCandlestickWithMarkLines(t *testing.T) {
	t.Parallel()

	data := makeBasicCandlestickData()
	series := CandlestickSeries{
		Data: data,
		MarkLine: SeriesMarkLine{
			Lines: []SeriesMark{
				{Type: SeriesMarkTypeAverage, Value: 110.0}, // Resistance
				{Type: SeriesMarkTypeAverage, Value: 100.0}, // Support
			},
		},
	}

	opt := CandlestickChartOption{
		Title: TitleOption{
			Text: "Candlestick with Support/Resistance",
		},
		Padding: NewBoxEqual(10),
		XAxis: XAxisOption{
			Labels: []string{"Jan", "Feb", "Mar", "Apr", "May"},
		},
		YAxis:  make([]YAxisOption, 1),
		Series: series,
	}

	painterOptions := PainterOptions{
		OutputFormat: ChartOutputSVG,
		Width:        800,
		Height:       600,
	}
	p := NewPainter(painterOptions)

	err := p.CandlestickChart(opt)
	require.NoError(t, err)
	data2, err := p.Bytes()
	require.NoError(t, err)
	assert.Greater(t, len(data2), 100)
}

func TestCandlestickAggregation(t *testing.T) {
	t.Parallel()

	// Create longer dataset to test aggregation
	data := []OHLCData{
		{Open: 100, High: 110, Low: 95, Close: 105},  // Period 1
		{Open: 105, High: 115, Low: 100, Close: 112}, // Period 1
		{Open: 112, High: 118, Low: 108, Close: 115}, // Period 2
		{Open: 115, High: 120, Low: 110, Close: 118}, // Period 2
		{Open: 118, High: 125, Low: 115, Close: 122}, // Period 3
		{Open: 122, High: 128, Low: 119, Close: 125}, // Period 3
	}

	series := CandlestickSeries{Data: data, Name: "1-Period"}
	aggregated := AggregateCandlestick(series, 2) // Aggregate into 2-period candles

	// Should have 3 aggregated candles from 6 original candles
	assert.Len(t, aggregated.Data, 3)
	assert.Equal(t, "1-Period (Aggregated)", aggregated.Name)

	// Test first aggregated candle
	first := aggregated.Data[0]
	assert.InDelta(t, 100.0, first.Open, 0.001)  // First open
	assert.InDelta(t, 112.0, first.Close, 0.001) // Last close of period
	assert.InDelta(t, 115.0, first.High, 0.001)  // Max high
	assert.InDelta(t, 95.0, first.Low, 0.001)    // Min low

	// Render aggregated chart
	opt := CandlestickChartOption{
		Title: TitleOption{
			Text: "Aggregated Candlestick Chart",
		},
		Padding: NewBoxEqual(10),
		XAxis: XAxisOption{
			Labels: []string{"Period 1", "Period 2", "Period 3"},
		},
		YAxis:  make([]YAxisOption, 1),
		Series: aggregated,
	}

	painterOptions := PainterOptions{
		OutputFormat: ChartOutputSVG,
		Width:        600,
		Height:       400,
	}
	p := NewPainter(painterOptions)

	err := p.CandlestickChart(opt)
	require.NoError(t, err)
	data2, err := p.Bytes()
	require.NoError(t, err)
	assert.Greater(t, len(data2), 100)
}

func TestCandlestickDifferentThemes(t *testing.T) {
	t.Parallel()

	themes := []ColorPalette{
		GetTheme(ThemeLight),
		GetTheme(ThemeDark),
		GetTheme(ThemeAnt),
		GetTheme(ThemeGrafana),
	}

	for i, theme := range themes {
		theme := theme // capture loop variable
		t.Run(fmt.Sprintf("theme_%d", i), func(t *testing.T) {
			t.Parallel()

			opt := makeBasicCandlestickChartOption()
			opt.Theme = theme

			painterOptions := PainterOptions{
				OutputFormat: ChartOutputSVG,
				Width:        600,
				Height:       400,
			}
			p := NewPainter(painterOptions)

			err := p.CandlestickChart(opt)
			require.NoError(t, err)
			data, err := p.Bytes()
			require.NoError(t, err)
			assert.Greater(t, len(data), 100)
		})
	}
}

func TestCandlestickLargeDataset(t *testing.T) {
	t.Parallel()

	// Generate larger dataset
	var data []OHLCData
	for i := 0; i < 50; i++ {
		basePrice := 100.0 + float64(i)*0.5
		data = append(data, OHLCData{
			Open:  basePrice,
			High:  basePrice + 5,
			Low:   basePrice - 3,
			Close: basePrice + 2,
		})
	}

	opt := CandlestickChartOption{
		Title: TitleOption{
			Text: "Large Dataset Candlestick Chart",
		},
		Padding: NewBoxEqual(10),
		XAxis: XAxisOption{
			Show: Ptr(false), // Hide labels for large dataset
		},
		YAxis:  make([]YAxisOption, 1),
		Series: NewSeriesCandlestick(data),
	}

	painterOptions := PainterOptions{
		OutputFormat: ChartOutputSVG,
		Width:        1200,
		Height:       600,
	}
	p := NewPainter(painterOptions)

	err := p.CandlestickChart(opt)
	require.NoError(t, err)
	data2, err := p.Bytes()
	require.NoError(t, err)
	assert.Greater(t, len(data2), 100)
}

func TestRenderCandlestickChart(t *testing.T) {
	t.Parallel()

	opt := ChartOption{
		SeriesList: NewSeriesCandlestick(makeBasicCandlestickData()).ToGenericSeriesList(),
		Title:      TitleOption{Text: "Price"},
		XAxis: XAxisOption{
			Labels: []string{"Jan", "Feb", "Mar", "Apr", "May"},
		},
		YAxis:  make([]YAxisOption, 1),
		Legend: LegendOption{SeriesNames: []string{"Price"}},
	}

	painter, err := Render(opt, SVGOutputOptionFunc())
	require.NoError(t, err)
	data, err := painter.Bytes()
	require.NoError(t, err)
	assert.Greater(t, len(data), 100)
}

func TestRenderCandlestickMixError(t *testing.T) {
	t.Parallel()

	candlesticks := NewSeriesCandlestick(makeBasicCandlestickData()).ToGenericSeriesList()
	line := NewSeriesListLine([][]float64{{1, 2, 3, 4, 5}}).ToGenericSeriesList()

	opt := ChartOption{
		SeriesList: append(candlesticks, line...),
	}

	_, err := Render(opt)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "candlestick can not mix other charts")
}
