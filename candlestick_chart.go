package charts

import (
	"errors"
	"math"
)

type candlestickChart struct {
	p   *Painter
	opt *CandlestickChartOption
}

// newCandlestickChart returns a candlestick chart renderer.
func newCandlestickChart(p *Painter, opt CandlestickChartOption) *candlestickChart {
	return &candlestickChart{
		p:   p,
		opt: &opt,
	}
}

// CandlestickChartOption defines options for rendering candlestick charts.
type CandlestickChartOption struct {
	Theme   ColorPalette
	Padding Box
	Series  CandlestickSeries
	XAxis   XAxisOption
	YAxis   []YAxisOption
	Title   TitleOption
	Legend  LegendOption
	// CandleWidth specifies default width of candlestick bodies (0.0-1.0)
	CandleWidth float64
	// UpColor default color for bullish candles
	UpColor Color
	// DownColor default color for bearish candles
	DownColor Color
	// WickWidth stroke width for high-low wicks
	WickWidth      float64
	ValueFormatter ValueFormatter
}

// NewCandlestickChartOptionWithData returns an initialized CandlestickChartOption with the Series set with the provided data slice.
func NewCandlestickChartOptionWithData(data []OHLCData) CandlestickChartOption {
	return NewCandlestickChartOptionWithSeries(NewSeriesCandlestick(data))
}

// NewCandlestickChartOptionWithSeries returns an initialized CandlestickChartOption with the provided Series.
func NewCandlestickChartOptionWithSeries(s CandlestickSeries) CandlestickChartOption {
	return CandlestickChartOption{
		Series:         s,
		Padding:        defaultPadding,
		Theme:          GetDefaultTheme(),
		YAxis:          make([]YAxisOption, 1), // Single series = single Y axis
		ValueFormatter: defaultValueFormatter,
		CandleWidth:    0.8, // Default 80% of available space
		WickWidth:      1.0,
	}
}

func (k *candlestickChart) renderChart(result *defaultRenderResult) (Box, error) {
	p := k.p
	opt := k.opt
	series := opt.Series
	if len(series.Data) == 0 {
		return BoxZero, errors.New("empty series data")
	}

	seriesPainter := result.seriesPainter

	// Use boundary gap positioning similar to bar charts for proper centering
	dataCount := len(series.Data)

	// Calculate spacing and width using bar chart logic for proper boundary gaps
	width := seriesPainter.Width()
	margin := 10 // margin on each side
	if width < 50 {
		margin = 5
	}

	// Calculate available space and candlestick width
	availableSpace := width - 2*margin
	candleSpacing := availableSpace / dataCount

	// Use default candle width if not set
	candleWidthRatio := opt.CandleWidth
	if candleWidthRatio <= 0 {
		candleWidthRatio = 0.8 // Default 80% of available space
	}

	candleWidth := int(float64(candleSpacing) * candleWidthRatio)
	if candleWidth < 1 {
		candleWidth = 1
	}

	// Use autoDivide for positioning but adjust for proper boundary gaps
	divideValues := result.xaxisRange.autoDivide()

	// Convert divide values to centered positions within available space
	positions := make([]int, dataCount)
	for i := 0; i < dataCount && i < len(divideValues); i++ {
		// Map from divide values to centered positions with margins
		positions[i] = margin + i*candleSpacing + candleSpacing/2
	}

	upColor := opt.UpColor
	downColor := opt.DownColor
	if upColor.IsZero() {
		upColor = opt.Theme.GetUpColor()
	}
	if downColor.IsZero() {
		downColor = opt.Theme.GetDownColor()
	}

	// render list must start with the markPointPainter, as it can influence label painters (if enabled)
	markPointPainter := newMarkPointPainter(seriesPainter)
	markLinePainter := newMarkLinePainter(seriesPainter)
	rendererList := []renderer{markPointPainter, markLinePainter}

	// Single series conversion - no loop needed
	yRange := result.yaxisRanges[series.YAxisIndex]

	var labelPainter *seriesLabelPainter
	if flagIs(true, series.Label.Show) {
		labelPainter = newSeriesLabelPainter(seriesPainter, []string{series.Name}, series.Label, opt.Theme)
		rendererList = append(rendererList, labelPainter)
	}

	points := make([]Point, len(series.Data)) // for mark points

	// Render each candlestick
	for j, ohlc := range series.Data {
		if j >= dataCount || j >= len(positions) {
			continue
		}

		// Skip invalid data
		if !validateOHLCData(ohlc) {
			points[j] = Point{X: positions[j], Y: math.MaxInt32} // Mark as null
			continue
		}

		// Calculate positions using proper centered positioning
		centerX := positions[j]
		leftX := centerX - candleWidth/2
		rightX := centerX + candleWidth/2

		highY := yRange.getRestHeight(ohlc.High)
		lowY := yRange.getRestHeight(ohlc.Low)
		openY := yRange.getRestHeight(ohlc.Open)
		closeY := yRange.getRestHeight(ohlc.Close)

		bodyTop := int(math.Min(float64(openY), float64(closeY)))
		bodyBottom := int(math.Max(float64(openY), float64(closeY)))

		// Determine colors and style
		isBullish := ohlc.Close >= ohlc.Open
		candleStyle := series.CandleStyle
		if candleStyle == "" {
			candleStyle = CandleStyleFilled // Default
		}

		var bodyColor, wickColor Color
		if isBullish {
			bodyColor = upColor
		} else {
			bodyColor = downColor
		}

		wickColor = opt.Theme.GetCandleWickColor()
		if wickColor.IsZero() {
			wickColor = bodyColor
		}

		// Draw high-low wick (if enabled)
		if flagIs(true, series.ShowWicks) || series.ShowWicks == nil {
			wickWidth := opt.WickWidth
			if wickWidth <= 0 {
				wickWidth = 1.0 // Default wick width
			}

			if highY < bodyTop {
				seriesPainter.LineStroke([]Point{
					{X: centerX, Y: highY},
					{X: centerX, Y: bodyTop},
				}, wickColor, wickWidth)
			}
			if lowY > bodyBottom {
				seriesPainter.LineStroke([]Point{
					{X: centerX, Y: bodyBottom},
					{X: centerX, Y: lowY},
				}, wickColor, wickWidth)
			}

			// Calculate cap width (half the candle width)
			capWidth := candleWidth / 4 // 1/4 of candle width on each side
			if capWidth < 1 {
				capWidth = 1 // Minimum cap width
			}

			// Draw horizontal cap at high point
			seriesPainter.LineStroke([]Point{
				{X: centerX - capWidth, Y: highY},
				{X: centerX + capWidth, Y: highY},
			}, wickColor, wickWidth)

			// Draw horizontal cap at low point
			seriesPainter.LineStroke([]Point{
				{X: centerX - capWidth, Y: lowY},
				{X: centerX + capWidth, Y: lowY},
			}, wickColor, wickWidth)
		}

		// Draw open-close body based on style

		if bodyTop == bodyBottom { // Doji (open == close)
			// Draw thin line instead of rectangle
			seriesPainter.LineStroke([]Point{
				{X: leftX, Y: bodyTop},
				{X: rightX, Y: bodyTop},
			}, bodyColor, 1.0)
		} else {
			switch candleStyle {
			case CandleStyleFilled:
				// Always filled
				seriesPainter.FilledRect(leftX, bodyTop, rightX, bodyBottom,
					bodyColor, bodyColor, 0.0)

			case CandleStyleTraditional:
				if isBullish {
					// Hollow body for bullish
					backgroundColor := opt.Theme.GetBackgroundColor()
					seriesPainter.FilledRect(leftX, bodyTop, rightX, bodyBottom,
						backgroundColor, bodyColor, 1.0)
				} else {
					// Filled body for bearish
					seriesPainter.FilledRect(leftX, bodyTop, rightX, bodyBottom,
						bodyColor, bodyColor, 0.0)
				}

			case CandleStyleOutline:
				// Always outlined only
				backgroundColor := opt.Theme.GetBackgroundColor()
				seriesPainter.FilledRect(leftX, bodyTop, rightX, bodyBottom,
					backgroundColor, bodyColor, 1.0)
			}
		}

		// Store point for mark points (use close price position)
		points[j] = Point{
			X: centerX,
			Y: closeY,
		}

		// Add label if enabled
		if labelPainter != nil {
			labelPainter.Add(labelValue{
				index:     0,          // Single series, so index is always 0
				value:     ohlc.Close, // Use close price for label
				x:         centerX,
				y:         closeY,
				fontStyle: series.Label.FontStyle,
			})
		}
	}

	// Handle mark lines (following line_chart.go pattern)
	if len(series.MarkLine.Lines) > 0 {
		markLineValueFormatter := getPreferredValueFormatter(series.MarkLine.ValueFormatter,
			series.Label.ValueFormatter, opt.ValueFormatter)
		seriesMarks := series.MarkLine.Lines.filterGlobal(false)

		if len(seriesMarks) > 0 {
			// Use close prices for mark line calculations
			closeValues := ExtractClosePrices(series)
			seriesColor := opt.Theme.GetSeriesColor(0) // Single series, so index is 0
			markLinePainter.add(markLineRenderOption{
				fillColor:      seriesColor,
				fontColor:      opt.Theme.GetMarkTextColor(),
				strokeColor:    seriesColor,
				font:           getPreferredFont(series.Label.FontStyle.Font),
				marklines:      seriesMarks,
				seriesValues:   closeValues,
				axisRange:      yRange,
				valueFormatter: markLineValueFormatter,
			})
		}
	}

	// Handle mark points (following line_chart.go pattern)
	if len(series.MarkPoint.Points) > 0 {
		markPointValueFormatter := getPreferredValueFormatter(series.MarkPoint.ValueFormatter,
			series.Label.ValueFormatter, opt.ValueFormatter)
		seriesMarks := series.MarkPoint.Points.filterGlobal(false)

		if len(seriesMarks) > 0 {
			// Use close prices for mark point calculations
			closeValues := ExtractClosePrices(series)
			seriesColor := opt.Theme.GetSeriesColor(0) // Single series, so index is 0
			markPointPainter.add(markPointRenderOption{
				fillColor:          seriesColor,
				font:               getPreferredFont(series.Label.FontStyle.Font),
				symbolSize:         series.MarkPoint.SymbolSize,
				points:             points,
				markpoints:         seriesMarks,
				seriesValues:       closeValues,
				valueFormatter:     markPointValueFormatter,
				seriesLabelPainter: labelPainter,
			})
		}
	}

	if err := doRender(rendererList...); err != nil {
		return BoxZero, err
	}
	return p.box, nil
}

func (k *candlestickChart) Render() (Box, error) {
	p := k.p
	opt := k.opt
	if opt.Theme == nil {
		opt.Theme = getPreferredTheme(p.theme)
	}
	if opt.Legend.Symbol == "" {
		opt.Legend.Symbol = SymbolSquare
	}

	renderResult, err := defaultRender(p, defaultRenderOption{
		theme:          opt.Theme,
		padding:        opt.Padding,
		seriesList:     &opt.Series,
		xAxis:          &opt.XAxis,
		yAxis:          opt.YAxis,
		title:          opt.Title,
		legend:         &opt.Legend,
		valueFormatter: opt.ValueFormatter,
	})
	if err != nil {
		return BoxZero, err
	}
	return k.renderChart(renderResult)
}
