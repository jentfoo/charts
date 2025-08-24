package main

import (
	"os"

	"github.com/go-analyze/charts"
)

func main() {
	// Create OHLC data designed to showcase various candlestick patterns
	ohlcData := []charts.OHLCData{
		// Normal candle
		{Open: 100.0, High: 110.0, Low: 95.0, Close: 105.0},

		// Doji pattern (open ≈ close)
		{Open: 105.0, High: 108.0, Low: 102.0, Close: 105.1},

		// Hammer pattern (long lower shadow, small body at top)
		{Open: 108.0, High: 109.0, Low: 98.0, Close: 107.0},

		// Normal bearish candle for engulfing setup
		{Open: 107.0, High: 108.0, Low: 103.0, Close: 104.0},

		// Bullish engulfing (current candle engulfs previous bearish candle)
		{Open: 102.0, High: 115.0, Low: 101.0, Close: 113.0},

		// Inverted hammer (long upper shadow, small body at bottom)
		{Open: 113.0, High: 125.0, Low: 112.0, Close: 114.0},

		// Normal bullish candle for engulfing setup
		{Open: 114.0, High: 118.0, Low: 113.0, Close: 117.0},

		// Bearish engulfing (current candle engulfs previous bullish candle)
		{Open: 119.0, High: 120.0, Low: 108.0, Close: 110.0},

		// Another doji
		{Open: 110.0, High: 113.0, Low: 107.0, Close: 109.9},

		// Recovery candle
		{Open: 109.0, High: 118.0, Low: 108.0, Close: 116.0},
	}

	// Create candlestick series with automatic pattern detection
	series := charts.NewCandlestickWithPatterns(ohlcData, charts.PatternDetectionOption{
		DojiThreshold:    0.005, // 0.5% threshold for doji detection
		ShadowRatio:      2.0,   // 2:1 shadow to body ratio for hammer patterns
		EngulfingMinSize: 0.7,   // 70% minimum size for engulfing patterns
	})

	// Configure series appearance
	series.Name = "Stock Price with Patterns"
	series.CandleStyle = charts.CandleStyleFilled

	// Add horizontal support/resistance lines
	series.MarkLine = charts.SeriesMarkLine{
		Lines: []charts.SeriesMark{
			{Type: charts.SeriesMarkTypeAverage}, // Resistance level
			{Type: charts.SeriesMarkTypeMin},     // Support level
		},
	}

	// Create chart option
	opt := charts.CandlestickChartOption{
		Title: charts.TitleOption{
			Text: "Candlestick Patterns Detection",
			FontStyle: charts.FontStyle{
				FontSize: 18,
			},
		},
		XAxis: charts.XAxisOption{
			Labels: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
		},
		YAxis: []charts.YAxisOption{
			{
				Unit: 1,
			},
		},
		Legend: charts.LegendOption{
			SeriesNames: []string{"Price with Patterns"},
			Show:        charts.Ptr(true),
		},
		SeriesList: charts.CandlestickSeriesList{series},
		Padding:    charts.NewBoxEqual(20),
	}

	// Create painter
	painterOptions := charts.PainterOptions{
		OutputFormat: charts.ChartOutputPNG,
		Width:        900,
		Height:       650,
	}
	p := charts.NewPainter(painterOptions)

	// Render the candlestick chart
	if err := p.CandlestickChart(opt); err != nil {
		panic(err)
	}

	// Save the chart to file
	buf, err := p.Bytes()
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("candlestick_patterns.png", buf, 0644); err != nil {
		panic(err)
	}
}
