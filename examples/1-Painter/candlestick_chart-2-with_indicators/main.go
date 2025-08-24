package main

import (
	"os"

	"github.com/go-analyze/charts"
)

func main() {
	// Create longer OHLC dataset for meaningful technical indicators
	ohlcData := []charts.OHLCData{
		{Open: 100.0, High: 110.0, Low: 95.0, Close: 105.0},
		{Open: 105.0, High: 115.0, Low: 100.0, Close: 112.0},
		{Open: 112.0, High: 118.0, Low: 108.0, Close: 115.0},
		{Open: 115.0, High: 120.0, Low: 110.0, Close: 118.0},
		{Open: 118.0, High: 125.0, Low: 115.0, Close: 122.0},
		{Open: 122.0, High: 128.0, Low: 119.0, Close: 125.0},
		{Open: 125.0, High: 130.0, Low: 122.0, Close: 127.0},
		{Open: 127.0, High: 132.0, Low: 124.0, Close: 129.0},
		{Open: 129.0, High: 135.0, Low: 126.0, Close: 131.0},
		{Open: 131.0, High: 138.0, Low: 128.0, Close: 135.0},
		{Open: 135.0, High: 140.0, Low: 132.0, Close: 137.0},
		{Open: 137.0, High: 142.0, Low: 134.0, Close: 139.0},
		{Open: 139.0, High: 145.0, Low: 136.0, Close: 141.0},
		{Open: 141.0, High: 148.0, Low: 138.0, Close: 145.0},
		{Open: 145.0, High: 150.0, Low: 142.0, Close: 147.0},
		{Open: 147.0, High: 152.0, Low: 144.0, Close: 149.0},
		{Open: 149.0, High: 155.0, Low: 146.0, Close: 151.0},
		{Open: 151.0, High: 158.0, Low: 148.0, Close: 155.0},
		{Open: 155.0, High: 160.0, Low: 152.0, Close: 157.0},
		{Open: 157.0, High: 162.0, Low: 154.0, Close: 159.0},
	}

	// Create candlestick series
	candlestickSeries := charts.CandlestickSeries{Data: ohlcData}

	// Calculate technical indicators
	closes := charts.ExtractClosePrices(candlestickSeries)
	sma10 := charts.CalculateSMA(closes, 10)
	sma20 := charts.CalculateSMA(closes, 20)
	ema10 := charts.CalculateEMA(closes, 10)

	// Create mixed chart using generic chart option
	seriesList := append(
		charts.CandlestickSeriesList{{Data: ohlcData}}.ToGenericSeriesList(),
		charts.NewSeriesListLine([][]float64{sma10, sma20, ema10}).ToGenericSeriesList()...,
	)

	chartOpt := charts.ChartOption{
		SeriesList: seriesList,
		Title: charts.TitleOption{
			Text: "Candlestick Chart with Technical Indicators",
			FontStyle: charts.FontStyle{
				FontSize: 18,
			},
		},
		XAxis: charts.XAxisOption{
			Labels: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10",
				"11", "12", "13", "14", "15", "16", "17", "18", "19", "20"},
		},
		YAxis: []charts.YAxisOption{
			{
				Unit: 1,
			},
		},
		Legend: charts.LegendOption{
			SeriesNames: []string{"Price", "SMA(10)", "SMA(20)", "EMA(10)"},
			Show:        charts.Ptr(true),
		},
		Padding: charts.NewBoxEqual(20),
	}

	// Render the chart
	painter, err := charts.Render(chartOpt)
	if err != nil {
		panic(err)
	}

	// Save the chart to file
	buf, err := painter.Bytes()
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("candlestick_with_indicators.png", buf, 0644); err != nil {
		panic(err)
	}
}
