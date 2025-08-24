package main

import (
	"os"

	"github.com/go-analyze/charts"
)

func main() {
	// Create sample OHLC data representing stock price movements
	ohlcData := []charts.OHLCData{
		{Open: 100.0, High: 110.0, Low: 95.0, Close: 105.0},
		{Open: 105.0, High: 115.0, Low: 100.0, Close: 112.0},
		{Open: 112.0, High: 118.0, Low: 108.0, Close: 115.0},
		{Open: 115.0, High: 120.0, Low: 110.0, Close: 108.0}, // bearish candle
		{Open: 108.0, High: 113.0, Low: 105.0, Close: 109.0},
		{Open: 109.0, High: 116.0, Low: 106.0, Close: 114.0},
		{Open: 114.0, High: 121.0, Low: 111.0, Close: 119.0},
		{Open: 119.0, High: 125.0, Low: 116.0, Close: 122.0},
		{Open: 122.0, High: 128.0, Low: 119.0, Close: 125.0},
		{Open: 125.0, High: 130.0, Low: 122.0, Close: 127.0},
	}

	// Create candlestick chart option with basic configuration
	opt := charts.NewCandlestickOptionWithData(ohlcData)

	// Configure chart appearance
	opt.Title = charts.TitleOption{
		Text: "Basic Candlestick Chart",
		FontStyle: charts.FontStyle{
			FontSize: 18,
		},
	}

	opt.XAxis = charts.XAxisOption{
		Labels: []string{"Day 1", "Day 2", "Day 3", "Day 4", "Day 5",
			"Day 6", "Day 7", "Day 8", "Day 9", "Day 10"},
	}

	opt.YAxis = []charts.YAxisOption{
		{
			Unit: 1,
		},
	}

	opt.Legend = charts.LegendOption{
		SeriesNames: []string{"Stock Price"},
		Show:        charts.Ptr(true),
	}

	// Set candle style to traditional (hollow bullish, filled bearish)
	opt.SeriesList[0].CandleStyle = charts.CandleStyleTraditional

	// Configure padding for better appearance
	opt.Padding = charts.NewBoxEqual(20)

	// Create painter with specific dimensions and output format
	painterOptions := charts.PainterOptions{
		OutputFormat: charts.ChartOutputPNG,
		Width:        800,
		Height:       600,
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

	if err := os.WriteFile("candlestick_basic.png", buf, 0644); err != nil {
		panic(err)
	}
}
