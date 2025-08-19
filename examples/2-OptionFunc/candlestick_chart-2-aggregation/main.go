package main

import (
	"os"

	"github.com/go-analyze/charts"
)

func main() {
	// Create 1-minute OHLC data (simulated)
	minuteData := []charts.OHLCData{
		// First 5-minute period
		{Open: 100.0, High: 102.0, Low: 99.0, Close: 101.0},  // Minute 1
		{Open: 101.0, High: 103.0, Low: 100.0, Close: 102.0}, // Minute 2
		{Open: 102.0, High: 105.0, Low: 101.0, Close: 104.0}, // Minute 3
		{Open: 104.0, High: 106.0, Low: 103.0, Close: 105.0}, // Minute 4
		{Open: 105.0, High: 107.0, Low: 104.0, Close: 106.0}, // Minute 5

		// Second 5-minute period
		{Open: 106.0, High: 108.0, Low: 105.0, Close: 107.0}, // Minute 6
		{Open: 107.0, High: 109.0, Low: 106.0, Close: 108.0}, // Minute 7
		{Open: 108.0, High: 110.0, Low: 107.0, Close: 109.0}, // Minute 8
		{Open: 109.0, High: 111.0, Low: 108.0, Close: 110.0}, // Minute 9
		{Open: 110.0, High: 112.0, Low: 109.0, Close: 111.0}, // Minute 10

		// Third 5-minute period
		{Open: 111.0, High: 113.0, Low: 110.0, Close: 112.0}, // Minute 11
		{Open: 112.0, High: 114.0, Low: 111.0, Close: 113.0}, // Minute 12
		{Open: 113.0, High: 115.0, Low: 112.0, Close: 114.0}, // Minute 13
		{Open: 114.0, High: 116.0, Low: 113.0, Close: 115.0}, // Minute 14
		{Open: 115.0, High: 117.0, Low: 114.0, Close: 116.0}, // Minute 15
	}

	// Create 1-minute series
	minuteSeries := charts.CandlestickSeries{Data: minuteData, Name: "1-Minute"}

	// Aggregate to 5-minute candles
	fiveMinuteSeries := charts.AggregateCandlestick(minuteSeries, 5)

	// Create combined chart showing both timeframes
	seriesList := append(
		charts.NewSeriesListCandlestick([][]charts.OHLCData{minuteData}).ToGenericSeriesList(),
		charts.NewSeriesListCandlestick([][]charts.OHLCData{fiveMinuteSeries.Data}).ToGenericSeriesList()...,
	)

	// Render the chart
	painter, err := charts.Render(charts.ChartOption{
		SeriesList: seriesList,
		Title: charts.TitleOption{
			Text: "Candlestick Data Aggregation: 1-Min vs 5-Min",
			FontStyle: charts.FontStyle{
				FontSize: 18,
			},
		},
		XAxis: charts.XAxisOption{
			Labels: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"},
		},
		YAxis: []charts.YAxisOption{
			{
				Unit: 1,
			},
		},
		Legend: charts.LegendOption{
			SeriesNames: []string{"1-Minute Data", "5-Minute Aggregated"},
			Show:        charts.Ptr(true),
		},
		Padding:      charts.NewBoxEqual(20),
		Width:        1200,
		Height:       700,
		OutputFormat: charts.ChartOutputPNG,
	})

	if err != nil {
		panic(err)
	}

	// Save the chart to file
	buf, err := painter.Bytes()
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("candlestick_aggregation.png", buf, 0644); err != nil {
		panic(err)
	}
}
