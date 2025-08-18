package charts

import (
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
		SeriesList: NewSeriesListCandlestick([][]OHLCData{makeBasicCandlestickData()}),
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
		SeriesList: NewSeriesListCandlestick([][]OHLCData{makeBasicCandlestickData()}),
	}
}

func TestCandlestickChart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		makeOptions func() CandlestickChartOption
	}{
		{
			name:        "basic",
			makeOptions: makeBasicCandlestickChartOption,
		},
		{
			name:        "minimal",
			makeOptions: makeMinimalCandlestickChartOption,
		},
	}

	for i, tt := range tests {
		tt := tt // capture loop variable
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			t.Parallel()

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
			assert.Greater(t, len(data), 100)
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
				return NewCandlestickChartOptionWithSeries(CandlestickSeriesList{})
			},
			errorMsgContains: "empty series list",
		},
	}

	for i, tt := range tests {
		tt := tt // capture loop variable
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			t.Parallel()

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
		tt := tt // capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, validateOHLCData(tt.ohlc))
		})
	}
}

func TestCandlestickSeriesListMethods(t *testing.T) {
	t.Parallel()

	data := makeBasicCandlestickData()
	seriesList := NewSeriesListCandlestick([][]OHLCData{data}, CandlestickSeriesOption{
		Names: []string{"Test Series"},
	})

	assert.Equal(t, 1, seriesList.len())
	assert.Equal(t, "Test Series", seriesList.getSeriesName(0))
	assert.Equal(t, len(data), seriesList.getSeriesLen(0))
	assert.Equal(t, string(SymbolSquare), string(seriesList.getSeriesSymbol(0)))

	// Test ToGenericSeriesList
	genericList := seriesList.ToGenericSeriesList()
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

	styles := []string{CandleStyleFilled, CandleStyleTraditional, CandleStyleOutline}
	for _, style := range styles {
		style := style // capture loop variable
		t.Run(style, func(t *testing.T) {
			t.Parallel()

			opt := makeBasicCandlestickChartOption()
			opt.SeriesList[0].CandleStyle = style

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
