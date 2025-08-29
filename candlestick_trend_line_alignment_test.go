package charts

import (
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// extract dashed trend line x coordinates from svg
func extractDashedPathXCoords(svg string) [][]int {
	rePath := regexp.MustCompile(`<path[^>]*stroke-dasharray[^>]*d="([^"]+)"`)
	paths := rePath.FindAllStringSubmatch(svg, -1)
	coordRe := regexp.MustCompile(`[ML] ([0-9]+) [0-9]+`)
	result := make([][]int, len(paths))
	for i, p := range paths {
		coords := coordRe.FindAllStringSubmatch(p[1], -1)
		xs := make([]int, len(coords))
		for j, c := range coords {
			x, _ := strconv.Atoi(c[1])
			xs[j] = x
		}
		result[i] = xs
	}
	return result
}

// compute expected center positions (absolute) for series
func computeCenters(r *defaultRenderResult, opt CandlestickChartOption, seriesIndex int) []int {
	width := r.seriesPainter.Width()
	seriesCount := opt.SeriesList.len()
	maxDataCount := getSeriesMaxDataCount(opt.SeriesList)
	candleWidthRatio := opt.CandleWidth
	if candleWidthRatio <= 0 {
		candleWidthRatio = 0.8
	}
	candleWidth := int(float64(width) * candleWidthRatio / float64(maxDataCount))
	if candleWidth < 1 {
		candleWidth = 1
	}
	candleWidthPerSeries := candleWidth / seriesCount
	if candleWidthPerSeries < 1 {
		candleWidthPerSeries = 1
	}
	divideValues := r.xaxisRange.autoDivide()
	centers := make([]int, len(opt.SeriesList.getSeries(seriesIndex).(*CandlestickSeries).Data))
	for j := range centers {
		if j >= len(divideValues) {
			continue
		}
		var sectionWidth int
		if j < len(divideValues)-1 {
			sectionWidth = divideValues[j+1] - divideValues[j]
		} else if j > 0 {
			sectionWidth = divideValues[j] - divideValues[j-1]
		} else {
			sectionWidth = width / maxDataCount
		}
		var groupMargin, candleMargin, cWidth int
		if seriesCount == 1 {
			cWidth = candleWidthPerSeries
		} else {
			var candleMarginFloat *float64
			if opt.CandleMargin != nil {
				marginPixels := float64(sectionWidth) * (*opt.CandleMargin)
				candleMarginFloat = &marginPixels
			}
			groupMargin, candleMargin, cWidth = calculateCandleMarginsAndSize(seriesCount, sectionWidth, candleWidthPerSeries, candleMarginFloat)
		}
		var center int
		if seriesCount == 1 {
			center = divideValues[j] + sectionWidth/2
		} else {
			x := divideValues[j] + groupMargin + seriesIndex*(cWidth+candleMargin)
			center = x + cWidth/2
		}
		centers[j] = center + r.seriesPainter.box.Left
	}
	return centers
}

func TestCandlestickTrendLineAlignmentSingleSeries(t *testing.T) {
	p := NewPainter(PainterOptions{OutputFormat: ChartOutputSVG, Width: 600, Height: 400})
	data := []OHLCData{{Open: 100, High: 110, Low: 90, Close: 105}, {Open: 105, High: 115, Low: 95, Close: 108}, {Open: 108, High: 118, Low: 100, Close: 112}}
	opt := CandlestickChartOption{
		Theme:   GetDefaultTheme(),
		Padding: NewBoxEqual(0),
		XAxis:   XAxisOption{Labels: []string{"A", "B", "C"}, Show: Ptr(false)},
		YAxis:   make([]YAxisOption, 1),
		SeriesList: CandlestickSeriesList{{
			Data:           data,
			CloseTrendLine: []SeriesTrendLine{{Type: SeriesTrendTypeLinear, DashedLine: Ptr(true)}},
		}},
		ShowWicks: Ptr(false),
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
	require.NoError(t, err)

	_, err = newCandlestickChart(p, opt).renderChart(renderResult)
	require.NoError(t, err)
	svgBytes, err := p.Bytes()
	require.NoError(t, err)

	paths := extractDashedPathXCoords(string(svgBytes))
	require.Len(t, paths, 1)
	expected := computeCenters(renderResult, opt, 0)
	assert.Equal(t, expected, paths[0])
}

func TestCandlestickTrendLineAlignmentMultiSeries(t *testing.T) {
	p := NewPainter(PainterOptions{OutputFormat: ChartOutputSVG, Width: 600, Height: 400})
	data := []OHLCData{{Open: 100, High: 110, Low: 90, Close: 105}, {Open: 105, High: 115, Low: 95, Close: 108}, {Open: 108, High: 118, Low: 100, Close: 112}}
	opt := CandlestickChartOption{
		Theme:   GetDefaultTheme(),
		Padding: NewBoxEqual(0),
		XAxis:   XAxisOption{Labels: []string{"A", "B", "C"}, Show: Ptr(false)},
		YAxis:   make([]YAxisOption, 1),
		SeriesList: CandlestickSeriesList{
			{Data: data, CloseTrendLine: []SeriesTrendLine{{Type: SeriesTrendTypeLinear, DashedLine: Ptr(true)}}},
			{Data: data, CloseTrendLine: []SeriesTrendLine{{Type: SeriesTrendTypeLinear, DashedLine: Ptr(true)}}},
		},
		ShowWicks: Ptr(false),
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
	require.NoError(t, err)

	_, err = newCandlestickChart(p, opt).renderChart(renderResult)
	require.NoError(t, err)
	svgBytes, err := p.Bytes()
	require.NoError(t, err)

	paths := extractDashedPathXCoords(string(svgBytes))
	require.Len(t, paths, 2)
	expected0 := computeCenters(renderResult, opt, 0)
	expected1 := computeCenters(renderResult, opt, 1)
	assert.Equal(t, expected0, paths[0])
	assert.Equal(t, expected1, paths[1])
}
