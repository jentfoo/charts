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

// CandlestickChartOption defines options for rendering candlestick charts. Render the chart using Painter.CandlestickChart.
type CandlestickChartOption struct {
	// Theme specifies the colors used for the candlestick chart.
	Theme ColorPalette
	// Padding specifies the padding around the chart.
	Padding Box
	// SeriesList provides the OHLC data population for the chart. Typically constructed using NewSeriesListCandlestick.
	SeriesList CandlestickSeriesList
	// XAxis contains options for the x-axis.
	XAxis XAxisOption
	// YAxis contains options for the y-axis. At most two y-axes are supported.
	YAxis []YAxisOption
	// Title contains options for rendering the chart title.
	Title TitleOption
	// Legend contains options for the data legend.
	Legend LegendOption
	// CandleWidth specifies the default width of candlestick bodies as a ratio (0.0-1.0).
	CandleWidth float64
	// ShowWicks controls whether high-low wicks are displayed by default. When nil, wicks are shown.
	// Individual series can override this setting.
	ShowWicks *bool
	// WickWidth is the stroke width for high-low wicks in pixels.
	WickWidth float64
	// ValueFormatter defines how float values are rendered to strings, notably for numeric axis labels.
	ValueFormatter ValueFormatter
}

// NewCandlestickOption returns an initialized CandlestickChartOption with default settings.
func NewCandlestickOption() CandlestickChartOption {
	return CandlestickChartOption{
		SeriesList:     CandlestickSeriesList{},
		Padding:        defaultPadding,
		Theme:          GetDefaultTheme(),
		YAxis:          make([]YAxisOption, 1),
		ValueFormatter: defaultValueFormatter,
		CandleWidth:    0.8, // Default 80% of available space
		WickWidth:      1.0,
	}
}

// NewCandlestickOptionWithData returns an initialized CandlestickChartOption with the SeriesList set with the provided data slices.
func NewCandlestickOptionWithData(data ...[]OHLCData) CandlestickChartOption {
	seriesList := make(CandlestickSeriesList, len(data))
	for i, ohlcData := range data {
		seriesList[i] = CandlestickSeries{Data: ohlcData}
	}
	return NewCandlestickOptionWithSeries(seriesList...)
}

// NewCandlestickOptionWithSeries returns an initialized CandlestickChartOption with the provided Series.
func NewCandlestickOptionWithSeries(series ...CandlestickSeries) CandlestickChartOption {
	seriesList := make(CandlestickSeriesList, len(series))
	copy(seriesList, series)
	return CandlestickChartOption{
		SeriesList:     seriesList,
		Padding:        defaultPadding,
		Theme:          GetDefaultTheme(),
		YAxis:          make([]YAxisOption, len(series)), // Y axis count based on series count
		ValueFormatter: defaultValueFormatter,
		CandleWidth:    0.8, // Default 80% of available space
		WickWidth:      1.0,
	}
}

func (k *candlestickChart) renderChart(result *defaultRenderResult) (Box, error) {
	p := k.p
	opt := k.opt
	seriesList := opt.SeriesList
	if seriesList.len() == 0 {
		return BoxZero, errors.New("empty series list")
	}

	seriesPainter := result.seriesPainter

	// Find maximum data count across all series
	maxDataCount := 0
	for seriesIndex := 0; seriesIndex < seriesList.len(); seriesIndex++ {
		dataLen := seriesList.getSeriesLen(seriesIndex)
		if dataLen > maxDataCount {
			maxDataCount = dataLen
		}
	}

	if maxDataCount == 0 {
		return BoxZero, errors.New("no data in any series")
	}

	// Reuse bar chart positioning logic for consistent spacing
	width := seriesPainter.Width()
	seriesCount := seriesList.len()

	// Calculate candle width using CandleWidth ratio (default 80%)
	candleWidthRatio := opt.CandleWidth
	if candleWidthRatio <= 0 {
		candleWidthRatio = 0.8 // Default 80% of available space
	}
	candleWidth := int(float64(width) * candleWidthRatio / float64(maxDataCount))
	if candleWidth < 1 {
		candleWidth = 1
	}

	// Use bar chart margin calculation for consistency
	margin, candleMargin, candleWidthPerSeries := calculateBarMarginsAndSize(seriesCount, width, candleWidth, nil)
	if candleWidthPerSeries < 1 {
		candleWidthPerSeries = 1
	}

	// Use autoDivide for positioning
	divideValues := result.xaxisRange.autoDivide()

	// render list must start with the markPointPainter, as it can influence label painters (if enabled)
	markPointPainter := newMarkPointPainter(seriesPainter)
	markLinePainter := newMarkLinePainter(seriesPainter)
	trendLinePainter := newTrendLinePainter(seriesPainter)
	rendererList := []renderer{markPointPainter, markLinePainter, trendLinePainter}

	// Check if any series has labels enabled
	seriesNames := seriesList.names()
	var labelPainter *seriesLabelPainter
	for seriesIndex := 0; seriesIndex < seriesList.len(); seriesIndex++ {
		series := seriesList.getSeries(seriesIndex).(*CandlestickSeries)
		if flagIs(true, series.Label.Show) {
			labelPainter = newSeriesLabelPainter(seriesPainter, seriesNames, series.Label,
				opt.Theme, opt.Padding.Right)
			rendererList = append(rendererList, labelPainter)
			break
		}
	}

	// Store points for each series (for mark points)
	allSeriesPoints := make([][]Point, seriesList.len())

	// Render each series
	for seriesIndex := 0; seriesIndex < seriesList.len(); seriesIndex++ {
		series := seriesList.getSeries(seriesIndex).(*CandlestickSeries)

		// Bounds check for Y axis index to prevent panic
		if series.YAxisIndex >= len(result.yaxisRanges) {
			return BoxZero, errors.New("candlestick series YAxisIndex out of bounds")
		}
		yRange := result.yaxisRanges[series.YAxisIndex]

		// Get series-specific up/down colors
		upColor, downColor := opt.Theme.GetSeriesUpDownColors(seriesIndex)

		// Initialize points array for this series
		seriesDataLen := len(series.Data)
		allSeriesPoints[seriesIndex] = make([]Point, seriesDataLen)

		// Render each candlestick in this series
		for j, ohlc := range series.Data {
			if j >= maxDataCount {
				continue
			}

			// Bounds check for divideValues to prevent panic
			if j >= len(divideValues) {
				continue
			}

			// Skip invalid data
			if !validateOHLCData(ohlc) {
				allSeriesPoints[seriesIndex][j] = Point{X: divideValues[j], Y: math.MaxInt32} // Mark as null
				continue
			}

			// Position calculation similar to bar chart logic
			centerX := divideValues[j] + margin + seriesIndex*(candleWidthPerSeries+candleMargin)
			leftX := centerX - candleWidthPerSeries/2
			rightX := centerX + candleWidthPerSeries/2

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
			showWicks := !flagIs(false, opt.ShowWicks)
			if series.ShowWicks != nil {
				showWicks = *series.ShowWicks
			}
			if showWicks {
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

				// Calculate cap width (based on series candle width)
				capWidth := candleWidthPerSeries / 4
				if capWidth < 1 {
					capWidth = 1
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
						seriesPainter.FilledRect(leftX, bodyTop, rightX, bodyBottom,
							ColorTransparent, bodyColor, 1.0)
					} else {
						// Filled body for bearish
						seriesPainter.FilledRect(leftX, bodyTop, rightX, bodyBottom,
							bodyColor, bodyColor, 0.0)
					}

				case CandleStyleOutline:
					// Always outlined only
					seriesPainter.FilledRect(leftX, bodyTop, rightX, bodyBottom,
						ColorTransparent, bodyColor, 1.0)
				}
			}

			// Store point for mark points (use close price position)
			allSeriesPoints[seriesIndex][j] = Point{
				X: centerX,
				Y: closeY,
			}

			// Add label if enabled
			if labelPainter != nil {
				labelPainter.Add(labelValue{
					index:     seriesIndex,
					value:     ohlc.Close, // Use close price for label
					x:         centerX,
					y:         closeY,
					fontStyle: series.Label.FontStyle,
				})
			}
		}
	}

	// Handle mark lines for each series
	for seriesIndex := 0; seriesIndex < seriesList.len(); seriesIndex++ {
		series := seriesList.getSeries(seriesIndex).(*CandlestickSeries)
		yRange := result.yaxisRanges[series.YAxisIndex]

		if len(series.MarkLine.Lines) > 0 {
			markLineValueFormatter := getPreferredValueFormatter(series.MarkLine.ValueFormatter,
				series.Label.ValueFormatter, opt.ValueFormatter)
			seriesMarks := series.MarkLine.Lines.filterGlobal(false)

			if len(seriesMarks) > 0 {
				// Use close prices for mark line calculations
				closeValues := ExtractClosePrices(*series)
				seriesColor := opt.Theme.GetSeriesColor(seriesIndex)
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
	}

	// Handle mark points for each series
	for seriesIndex := 0; seriesIndex < seriesList.len(); seriesIndex++ {
		series := seriesList.getSeries(seriesIndex).(*CandlestickSeries)

		if len(series.MarkPoint.Points) > 0 {
			markPointValueFormatter := getPreferredValueFormatter(series.MarkPoint.ValueFormatter,
				series.Label.ValueFormatter, opt.ValueFormatter)
			seriesMarks := series.MarkPoint.Points.filterGlobal(false)

			if len(seriesMarks) > 0 {
				// Use close prices for mark point calculations
				closeValues := ExtractClosePrices(*series)
				seriesColor := opt.Theme.GetSeriesColor(seriesIndex)
				markPointPainter.add(markPointRenderOption{
					fillColor:          seriesColor,
					font:               getPreferredFont(series.Label.FontStyle.Font),
					symbolSize:         series.MarkPoint.SymbolSize,
					points:             allSeriesPoints[seriesIndex],
					markpoints:         seriesMarks,
					seriesValues:       closeValues,
					valueFormatter:     markPointValueFormatter,
					seriesLabelPainter: labelPainter,
				})
			}
		}
	}

	// Handle trend lines for each series
	for seriesIndex := 0; seriesIndex < seriesList.len(); seriesIndex++ {
		series := seriesList.getSeries(seriesIndex).(*CandlestickSeries)
		yRange := result.yaxisRanges[series.YAxisIndex]

		if len(series.TrendLine) > 0 {
			// Use close prices for trend line calculations
			closeValues := ExtractClosePrices(*series)
			trendLinePainter.add(trendLineRenderOption{
				defaultStrokeColor: opt.Theme.GetSeriesTrendColor(seriesIndex),
				xValues:            divideValues,
				seriesValues:       closeValues,
				axisRange:          yRange,
				trends:             series.TrendLine,
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
		seriesList:     &opt.SeriesList,
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
