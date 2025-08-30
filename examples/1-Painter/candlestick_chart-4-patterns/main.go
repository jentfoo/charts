package main

import (
	"fmt"
	"os"

	"github.com/go-analyze/charts"
)

// patternExample holds configuration for a pattern demonstration
type patternExample struct {
	name       string
	filename   string
	title      string
	width      int
	height     int
	series     charts.CandlestickSeries
	showLegend bool
}

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

	// Create all pattern configuration examples
	examples := createPatternExamples(ohlcData)

	// Generate all example charts
	generateExampleCharts(examples, ohlcData)
}

// createPatternExamples creates various pattern configuration examples
func createPatternExamples(ohlcData []charts.OHLCData) []patternExample {
	return []patternExample{
		{
			name:     "Basic All Patterns",
			filename: "candlestick_patterns.png",
			title:    "Candlestick Patterns Detection",
			width:    900,
			height:   650,
			series: charts.CandlestickSeries{
				Data:          ohlcData,
				Name:          "Stock Price with Patterns",
				CandleStyle:   charts.CandleStyleFilled,
				PatternConfig: charts.PatternsAll(),
				Label: charts.SeriesLabel{
					Show: charts.Ptr(true),
				},
				CloseMarkLine: charts.SeriesMarkLine{
					Lines: []charts.SeriesMark{
						{Type: charts.SeriesMarkTypeAverage}, // Resistance level
						{Type: charts.SeriesMarkTypeMin},     // Support level
					},
				},
			},
			showLegend: true,
		},
		{
			name:     "Important Patterns Only",
			filename: "patterns_important.png",
			title:    "Important Reversal Patterns",
			width:    800,
			height:   400,
			series: charts.CandlestickSeries{
				Data:          ohlcData,
				Name:          "Important Patterns",
				PatternConfig: charts.PatternsCore(),
				Label: charts.SeriesLabel{
					Show: charts.Ptr(true),
				},
			},
		},
		{
			name:     "Custom Pattern Selection",
			filename: "patterns_custom.png",
			title:    "Custom Pattern Selection",
			width:    800,
			height:   400,
			series: charts.CandlestickSeries{
				Data: ohlcData,
				Name: "Custom Selection",
				PatternConfig: &charts.CandlestickPatternConfig{
					PreferPatternLabels: true,
					EnabledPatterns: []string{
						charts.CandlestickPatternDoji,
						charts.CandlestickPatternHammer,
						charts.CandlestickPatternEngulfingBull,
					},
				},
				Label: charts.SeriesLabel{
					Show: charts.Ptr(true),
				},
			},
		},
		{
			name:     "Patterns Complement User Labels",
			filename: "patterns_complement.png",
			title:    "Patterns + Price Labels",
			width:    800,
			height:   400,
			series: charts.CandlestickSeries{
				Data: ohlcData,
				Name: "Complement Mode",
				PatternConfig: func() *charts.CandlestickPatternConfig {
					config := charts.PatternsCore()
					config.PreferPatternLabels = false // Complement mode
					return config
				}(),
				Label: charts.SeriesLabel{
					Show: charts.Ptr(true),
					LabelFormatter: func(index int, name string, val float64) (string, *charts.LabelStyle) {
						// Only label specific candles (every 3rd one)
						if index%3 == 0 {
							return fmt.Sprintf("$%.2f", val), nil
						}
						return "", nil // No user label - patterns will show
					},
				},
			},
		},
		{
			name:     "Custom Pattern Formatter",
			filename: "patterns_custom_format.png",
			title:    "Custom Pattern Formatting",
			width:    800,
			height:   400,
			series: charts.CandlestickSeries{
				Data: ohlcData,
				Name: "Custom Format",
				PatternConfig: &charts.CandlestickPatternConfig{
					PreferPatternLabels: true,
					EnabledPatterns:     charts.PatternsAll().EnabledPatterns,
					PatternFormatter: func(patterns []charts.PatternDetectionResult, seriesName string, value float64) (string, *charts.LabelStyle) {
						if len(patterns) == 0 {
							return "", nil
						}

						var names []string
						for _, p := range patterns {
							names = append(names, p.PatternName)
						}

						labelText := names[0]
						if len(patterns) > 1 {
							labelText += fmt.Sprintf(" +%d", len(patterns)-1)
						}

						return labelText, &charts.LabelStyle{
							FontStyle: charts.FontStyle{
								FontColor: charts.Color{R: 255, G: 255, B: 255, A: 255},
								FontSize:  8,
							},
							BackgroundColor: charts.Color{R: 0, G: 0, B: 255, A: 180},
							CornerRadius:    2,
						}
					},
				},
				Label: charts.SeriesLabel{
					Show: charts.Ptr(true),
				},
			},
		},
		{
			name:     "Bullish Patterns Only",
			filename: "patterns_bullish.png",
			title:    "Bullish Patterns Only",
			width:    800,
			height:   400,
			series: charts.CandlestickSeries{
				Data:          ohlcData,
				Name:          "Bullish Only",
				PatternConfig: charts.PatternsBullish(),
				Label: charts.SeriesLabel{
					Show: charts.Ptr(true),
				},
			},
		},
	}
}

// generateExampleCharts creates and saves all the pattern example charts
func generateExampleCharts(examples []patternExample, ohlcData []charts.OHLCData) {
	for _, example := range examples {
		fmt.Printf("Generating %s (%s)...\n", example.name, example.filename)

		// Create chart option
		opt := charts.CandlestickChartOption{
			Title: charts.TitleOption{
				Text: example.title,
				FontStyle: charts.FontStyle{
					FontSize: 16,
				},
			},
			XAxis: charts.XAxisOption{
				Labels: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
			},
			YAxis: []charts.YAxisOption{
				{Unit: 1},
			},
			SeriesList: charts.CandlestickSeriesList{example.series},
			Padding:    charts.NewBoxEqual(20),
		}

		// Add legend if requested
		if example.showLegend {
			opt.Legend = charts.LegendOption{
				SeriesNames: []string{example.series.Name},
				Show:        charts.Ptr(true),
			}
		}

		// Create painter
		painterOptions := charts.PainterOptions{
			OutputFormat: charts.ChartOutputPNG,
			Width:        example.width,
			Height:       example.height,
		}
		p := charts.NewPainter(painterOptions)

		// Render the candlestick chart
		if err := p.CandlestickChart(opt); err != nil {
			panic(fmt.Errorf("failed to render %s: %v", example.name, err))
		}

		// Save the chart to file
		buf, err := p.Bytes()
		if err != nil {
			panic(fmt.Errorf("failed to get bytes for %s: %v", example.name, err))
		}

		if err := os.WriteFile(example.filename, buf, 0644); err != nil {
			panic(fmt.Errorf("failed to write %s: %v", example.filename, err))
		}

		fmt.Printf("✓ Generated %s\n", example.filename)
	}
}
