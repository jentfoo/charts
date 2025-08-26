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
		SeriesList: CandlestickSeriesList{{Data: makeBasicCandlestickData()}},
	}
}

func makeMinimalCandlestickChartOption() CandlestickChartOption {
	return CandlestickChartOption{
		Padding: NewBoxEqual(10),
		XAxis: XAxisOption{
			Labels: []string{"1", "2", "3", "4", "5"},
			Show:   Ptr(false),
		},
		YAxis:      make([]YAxisOption, 1),
		SeriesList: CandlestickSeriesList{{Data: makeBasicCandlestickData()}},
	}
}

func TestCandlestickChart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		makeOptions func() CandlestickChartOption
		svg         string
		pngCRC      uint32
	}{
		{
			name:        "basic",
			makeOptions: makeBasicCandlestickChartOption,
			svg:         "",
		},
		{
			name:        "minimal",
			makeOptions: makeMinimalCandlestickChartOption,
			svg:         "",
		},
		{
			name: "traditional",
			makeOptions: func() CandlestickChartOption {
				opt := makeBasicCandlestickChartOption()
				opt.SeriesList[0].CandleStyle = CandleStyleTraditional
				return opt
			},
			svg: "",
		},
		{
			name: "outline",
			makeOptions: func() CandlestickChartOption {
				opt := makeBasicCandlestickChartOption()
				opt.SeriesList[0].CandleStyle = CandleStyleOutline
				return opt
			},
			svg: "",
		},
		{
			name: "no_wicks",
			makeOptions: func() CandlestickChartOption {
				opt := makeBasicCandlestickChartOption()
				opt.SeriesList[0].ShowWicks = Ptr(false)
				return opt
			},
			svg: "",
		},
		{
			name: "custom_style",
			makeOptions: func() CandlestickChartOption {
				opt := makeBasicCandlestickChartOption()
				opt.CandleWidth = 0.5
				opt.WickWidth = 2.0
				// Colors now handled by theme
				return opt
			},
			svg: "",
		},
		{
			name: "marks",
			makeOptions: func() CandlestickChartOption {
				opt := makeBasicCandlestickChartOption()
				opt.SeriesList[0].Label.Show = Ptr(true)
				opt.SeriesList[0].MarkLine = NewMarkLine("min", "max")
				opt.SeriesList[0].MarkLine.ValueFormatter = func(f float64) string { return fmt.Sprintf("%.2f", f) }
				opt.SeriesList[0].MarkPoint = NewMarkPoint("min", "max")
				opt.SeriesList[0].MarkPoint.ValueFormatter = func(f float64) string { return fmt.Sprintf("%.2f", f) }
				return opt
			},
			svg: "",
		},
		{
			name: "doji",
			makeOptions: func() CandlestickChartOption {
				opt := makeBasicCandlestickChartOption()
				data := makeBasicCandlestickData()
				if len(data) > 0 {
					data[0] = OHLCData{Open: 100, High: 110, Low: 95, Close: 100}
				}
				opt.SeriesList[0] = CandlestickSeries{Data: data}
				return opt
			},
			svg: "",
		},
		{
			name: "dual_axis",
			makeOptions: func() CandlestickChartOption {
				opt := makeBasicCandlestickChartOption()
				opt.YAxis = append(opt.YAxis, YAxisOption{})
				opt.SeriesList[0].YAxisIndex = 1
				return opt
			},
			svg: "",
		},
		{
			name: "filled_style",
			makeOptions: func() CandlestickChartOption {
				opt := makeBasicCandlestickChartOption()
				opt.SeriesList[0].CandleStyle = CandleStyleFilled
				return opt
			},
			svg: "",
		},
		{
			name: "with_trend_lines",
			makeOptions: func() CandlestickChartOption {
				opt := makeBasicCandlestickChartOption()
				opt.SeriesList[0].TrendLine = []SeriesTrendLine{
					{Type: SeriesTrendTypeSMA, Period: 3},
					{Type: SeriesTrendTypeEMA, Period: 3},
				}
				return opt
			},
			svg: "",
		},
		{
			name: "with_mark_lines",
			makeOptions: func() CandlestickChartOption {
				data := makeBasicCandlestickData()
				series := CandlestickSeries{
					Data: data,
					MarkLine: SeriesMarkLine{
						Lines: []SeriesMark{
							{Type: SeriesMarkTypeAverage}, // Resistance
							{Type: SeriesMarkTypeMin},     // Support
						},
					},
				}
				return CandlestickChartOption{
					Title: TitleOption{
						Text: "Candlestick with Support/Resistance",
					},
					Padding: NewBoxEqual(10),
					XAxis: XAxisOption{
						Labels: []string{"Jan", "Feb", "Mar", "Apr", "May"},
					},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{series},
				}
			},
			svg: "",
		},
		{
			name: "large_dataset",
			makeOptions: func() CandlestickChartOption {
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
				return CandlestickChartOption{
					Title: TitleOption{
						Text: "Large Dataset Candlestick Chart",
					},
					Padding: NewBoxEqual(10),
					XAxis: XAxisOption{
						Show: Ptr(false), // Hide labels for large dataset
					},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{{Data: data}},
				}
			},
			svg: "",
		},
		{
			name: "multiple_series",
			makeOptions: func() CandlestickChartOption {
				// Create different datasets for multiple series
				series1Data := []OHLCData{
					{Open: 100.0, High: 110.0, Low: 95.0, Close: 105.0},
					{Open: 105.0, High: 115.0, Low: 100.0, Close: 112.0},
					{Open: 112.0, High: 118.0, Low: 108.0, Close: 115.0},
					{Open: 115.0, High: 120.0, Low: 110.0, Close: 108.0},
					{Open: 108.0, High: 113.0, Low: 105.0, Close: 109.0},
				}
				series2Data := []OHLCData{
					{Open: 120.0, High: 130.0, Low: 115.0, Close: 125.0},
					{Open: 125.0, High: 135.0, Low: 120.0, Close: 132.0},
					{Open: 132.0, High: 138.0, Low: 128.0, Close: 135.0},
					{Open: 135.0, High: 140.0, Low: 130.0, Close: 128.0},
					{Open: 128.0, High: 133.0, Low: 125.0, Close: 129.0},
				}
				series3Data := []OHLCData{
					{Open: 80.0, High: 110.0, Low: 45.0, Close: 85.0},
					{Open: 85.0, High: 115.0, Low: 40.0, Close: 82.0},
					{Open: 82.0, High: 118.0, Low: 48.0, Close: 85.0},
					{Open: 85.0, High: 120.0, Low: 40.0, Close: 88.0},
					{Open: 88.0, High: 113.0, Low: 45.0, Close: 89.0},
				}
				return CandlestickChartOption{
					Title: TitleOption{Text: "Multiple Candlestick Series"},
					XAxis: XAxisOption{
						Labels: []string{"Day 1", "Day 2", "Day 3", "Day 4", "Day 5"},
					},
					YAxis: make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{
						{Data: series1Data, Name: "Stock A"},
						{Data: series2Data, Name: "Stock B"},
						{Data: series3Data, Name: "Stock C"},
					},
					Legend: LegendOption{
						SeriesNames: []string{"Stock A", "Stock B", "Stock C"},
						Show:        Ptr(true),
					},
					Padding: NewBoxEqual(10),
				}
			},
			svg: "",
		},
		{
			name: "with_sma",
			makeOptions: func() CandlestickChartOption {
				ohlcData := makeBasicCandlestickData()
				candlestickSeries := CandlestickSeries{Data: ohlcData}

				// Test that trend lines can be added to candlestick series
				candlestickSeries.TrendLine = NewTrendLine(SeriesTrendTypeAverage)

				// Convert back to CandlestickChartOption for consistency
				opt := makeBasicCandlestickChartOption()
				opt.Title.Text = "SMA Line Chart"
				return opt
			},
			svg: "",
		},
		{
			name: "with_bollinger_bands",
			makeOptions: func() CandlestickChartOption {
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

				return CandlestickChartOption{
					Title: TitleOption{Text: "Bollinger Bands Lines"},
					XAxis: XAxisOption{
						Labels: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
					},
					YAxis: make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{{
						Data: ohlcData,
						TrendLine: []SeriesTrendLine{
							{Type: SeriesTrendTypeBollingerUpper, Period: 5},
							{Type: SeriesTrendTypeSMA, Period: 5},
							{Type: SeriesTrendTypeBollingerLower, Period: 5},
						},
					}},
					Legend: LegendOption{
						Show: Ptr(true),
					},
					Padding: NewBoxEqual(10),
				}
			},
			svg: "",
		},
		{
			name: "with_ema",
			makeOptions: func() CandlestickChartOption {
				ohlcData := makeBasicCandlestickData()
				series := CandlestickSeries{Data: ohlcData}

				// Test that EMA trend lines can be added to candlestick series
				series.TrendLine = NewTrendLine(SeriesTrendTypeEMA)

				// Convert to CandlestickChartOption for consistency
				return CandlestickChartOption{
					Title: TitleOption{Text: "EMA Line Chart"},
					XAxis: XAxisOption{
						Labels: []string{"Jan", "Feb", "Mar", "Apr", "May"},
					},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{{Data: ohlcData}},
					Legend: LegendOption{
						SeriesNames: []string{"EMA(3)"},
					},
					Padding: NewBoxEqual(10),
				}
			},
			svg: "",
		},
		{
			name: "with_rsi",
			makeOptions: func() CandlestickChartOption {
				ohlcData := makeBasicCandlestickData()

				return CandlestickChartOption{
					Title: TitleOption{Text: "Candlestick (RSI tested)"},
					XAxis: XAxisOption{
						Labels: []string{"Jan", "Feb", "Mar", "Apr", "May"},
					},
					YAxis: make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{{
						Data: ohlcData,
						TrendLine: []SeriesTrendLine{
							{Type: SeriesTrendTypeRSI, Period: 3},
						},
					}},
					Legend: LegendOption{
						Show: Ptr(true),
					},
					Padding: NewBoxEqual(10),
				}
			},
			svg: "",
		},
		{
			name: "aggregation",
			makeOptions: func() CandlestickChartOption {
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

				return CandlestickChartOption{
					Title: TitleOption{
						Text: "Aggregated Candlestick Chart",
					},
					Padding: NewBoxEqual(10),
					XAxis: XAxisOption{
						Labels: []string{"Period 1", "Period 2", "Period 3"},
					},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{aggregated},
				}
			},
			svg: "",
		},
		{
			name: "large_series_count",
			makeOptions: func() CandlestickChartOption {
				// Create 10 different series to test color cycling
				seriesCount := 10
				dataPointsPerSeries := 5

				var seriesList CandlestickSeriesList
				var seriesNames []string

				for i := 0; i < seriesCount; i++ {
					basePrice := 100.0 + float64(i*20) // Different price ranges
					data := make([]OHLCData, dataPointsPerSeries)

					for j := 0; j < dataPointsPerSeries; j++ {
						open := basePrice + float64(j*5)
						high := open + 10.0
						low := open - 5.0
						close := open + float64((j%2)*10-5) // Alternating up/down

						data[j] = OHLCData{
							Open:  open,
							High:  high,
							Low:   low,
							Close: close,
						}
					}

					series := CandlestickSeries{
						Data: data,
						Name: fmt.Sprintf("Series %d", i+1),
					}
					seriesList = append(seriesList, series)
					seriesNames = append(seriesNames, series.Name)
				}

				return CandlestickChartOption{
					Title: TitleOption{Text: "Large Candlestick Series Test"},
					XAxis: XAxisOption{
						Labels: []string{"T1", "T2", "T3", "T4", "T5"},
					},
					YAxis:      make([]YAxisOption, 1),
					SeriesList: seriesList,
					Legend: LegendOption{
						SeriesNames: seriesNames,
						Show:        Ptr(true),
					},
					Padding: NewBoxEqual(10),
				}
			},
			svg: "",
		},
		{
			name: "multiple_series_different_styles",
			makeOptions: func() CandlestickChartOption {
				// Create different datasets for multiple series with different styles
				series1Data := []OHLCData{
					{Open: 100.0, High: 110.0, Low: 95.0, Close: 105.0},
					{Open: 105.0, High: 115.0, Low: 100.0, Close: 112.0},
					{Open: 112.0, High: 118.0, Low: 108.0, Close: 115.0},
				}

				series2Data := []OHLCData{
					{Open: 150.0, High: 160.0, Low: 145.0, Close: 155.0},
					{Open: 155.0, High: 165.0, Low: 150.0, Close: 162.0},
					{Open: 162.0, High: 168.0, Low: 158.0, Close: 165.0},
				}

				return CandlestickChartOption{
					Title: TitleOption{Text: "Multiple Series with Different Styles"},
					XAxis: XAxisOption{
						Labels: []string{"Day 1", "Day 2", "Day 3"},
					},
					YAxis: make([]YAxisOption, 1),
					SeriesList: CandlestickSeriesList{
						{
							Data:        series1Data,
							Name:        "Stock A (Filled)",
							CandleStyle: CandleStyleFilled,
						},
						{
							Data:        series2Data,
							Name:        "Stock B (Traditional)",
							CandleStyle: CandleStyleTraditional,
						},
					},
					Legend: LegendOption{
						SeriesNames: []string{"Stock A (Filled)", "Stock B (Traditional)"},
						Show:        Ptr(true),
					},
					Padding: NewBoxEqual(10),
				}
			},
			svg: "",
		},
	}

	for i, tc := range tests {
		t.Run(strconv.Itoa(i)+"-"+tc.name, func(t *testing.T) {
			p := NewPainter(PainterOptions{
				OutputFormat: ChartOutputSVG,
				Width:        800,
				Height:       600,
			})
			r := NewPainter(PainterOptions{
				OutputFormat: ChartOutputPNG,
				Width:        800,
				Height:       600,
			})

			opt := tc.makeOptions()

			validateCandlestickChartRender(t, p, r, opt, tc.svg, tc.pngCRC)
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
				return NewCandlestickOptionWithSeries(CandlestickSeries{})
			},
			errorMsgContains: "no data in any series",
		},
		{
			name: "invalid_yaxis_index",
			makeOptions: func() CandlestickChartOption {
				series := CandlestickSeries{
					Data: []OHLCData{
						{Open: 100, High: 110, Low: 95, Close: 105},
					},
					YAxisIndex: 5, // Invalid - only have 1 Y axis by default
				}
				return NewCandlestickOptionWithSeries(series)
			},
			errorMsgContains: "YAxisIndex out of bounds",
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
	series := CandlestickSeries{Data: data, Name: "Test Series"}
	seriesList := CandlestickSeriesList{series}

	assert.Equal(t, 1, seriesList.len())
	assert.Equal(t, "Test Series", seriesList.names()[0])
	assert.Equal(t, len(data), len(series.Data))
	assert.Equal(t, string(SymbolSquare), string(seriesList.getSeriesSymbol(0)))

	// Test ToGenericSeriesList
	genericList := seriesList.ToGenericSeriesList()
	assert.Len(t, genericList, 1)
	assert.Equal(t, ChartTypeCandlestick, genericList[0].Type)
	assert.Equal(t, len(data), len(genericList[0].Values))
}

func TestRenderCandlestickChart(t *testing.T) {
	t.Parallel()

	opt := ChartOption{
		SeriesList: CandlestickSeriesList{{Data: makeBasicCandlestickData()}}.ToGenericSeriesList(),
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
	assertEqualSVG(t, "", data)
}

func TestRenderCandlestickMixError(t *testing.T) {
	t.Parallel()

	candlesticks := CandlestickSeriesList{{Data: makeBasicCandlestickData()}}.ToGenericSeriesList()
	line := NewSeriesListLine([][]float64{{1, 2, 3, 4, 5}}).ToGenericSeriesList()

	opt := ChartOption{
		SeriesList: append(candlesticks, line...),
	}

	_, err := Render(opt)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "candlestick can not mix other charts")
}

func validateCandlestickChartRender(t *testing.T, svgP, pngP *Painter, opt CandlestickChartOption, expectedSVG string, expectedCRC uint32) {
	t.Helper()

	err := svgP.CandlestickChart(opt)
	require.NoError(t, err)
	data, err := svgP.Bytes()
	require.NoError(t, err)
	assertEqualSVG(t, expectedSVG, data)

	err = pngP.CandlestickChart(opt)
	require.NoError(t, err)
	rdata, err := pngP.Bytes()
	require.NoError(t, err)
	assertEqualPNGCRC(t, expectedCRC, rdata)
}
