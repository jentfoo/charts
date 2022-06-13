// MIT License

// Copyright (c) 2022 Tree Xie

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package charts

import (
	"github.com/golang/freetype/truetype"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

type lineChart struct {
	p   *Painter
	opt *LineChartOption
}

func NewLineChart(p *Painter, opt LineChartOption) *lineChart {
	if opt.Theme == nil {
		opt.Theme = NewTheme("")
	}
	return &lineChart{
		p:   p,
		opt: &opt,
	}
}

type LineChartOption struct {
	Theme ColorPalette
	// The font size
	Font *truetype.Font
	// The data series list
	SeriesList SeriesList
	// The x axis option
	XAxis XAxisOption
	// The padding of line chart
	Padding Box
	// The y axis option
	YAxisOptions []YAxisOption
	// The option of title
	TitleOption TitleOption
	// The legend option
	LegendOption LegendOption
}

func (l *lineChart) Render() (Box, error) {
	p := l.p
	opt := l.opt
	seriesList := opt.SeriesList
	seriesList.init()
	renderResult, err := defaultRender(p, defaultRenderOption{
		Theme:        opt.Theme,
		Padding:      opt.Padding,
		SeriesList:   seriesList,
		XAxis:        opt.XAxis,
		YAxisOptions: opt.YAxisOptions,
		TitleOption:  opt.TitleOption,
		LegendOption: opt.LegendOption,
	})
	if err != nil {
		return chart.BoxZero, err
	}

	seriesList = seriesList.Filter(ChartTypeLine)

	seriesPainter := renderResult.p

	xDivideValues := autoDivide(seriesPainter.Width(), len(opt.XAxis.Data))
	xValues := make([]int, len(xDivideValues)-1)
	for i := 0; i < len(xDivideValues)-1; i++ {
		xValues[i] = (xDivideValues[i] + xDivideValues[i+1]) >> 1
	}
	markPointPainter := NewMarkPointPainter(seriesPainter)
	markLinePainter := NewMarkLinePainter(seriesPainter)
	rendererList := []Renderer{
		markPointPainter,
		markLinePainter,
	}
	for index, series := range seriesList {
		seriesColor := opt.Theme.GetSeriesColor(index)
		drawingStyle := Style{
			StrokeColor: seriesColor,
			StrokeWidth: defaultStrokeWidth,
		}

		seriesPainter.SetDrawingStyle(drawingStyle)
		yr := renderResult.axisRanges[series.AxisIndex]
		points := make([]Point, 0)
		for i, item := range series.Data {
			h := yr.getRestHeight(item.Value)
			p := Point{
				X: xValues[i],
				Y: h,
			}
			points = append(points, p)
		}
		// 画线
		seriesPainter.LineStroke(points)

		// 画点
		if opt.Theme.IsDark() {
			drawingStyle.FillColor = drawingStyle.StrokeColor
		} else {
			drawingStyle.FillColor = drawing.ColorWhite
		}
		drawingStyle.StrokeWidth = 1
		seriesPainter.SetDrawingStyle(drawingStyle)
		seriesPainter.Dots(points)
		markPointPainter.Add(markPointRenderOption{
			FillColor: seriesColor,
			Font:      opt.Font,
			Points:    points,
			Series:    series,
		})
		markLinePainter.Add(markLineRenderOption{
			FillColor:   seriesColor,
			FontColor:   opt.Theme.GetTextColor(),
			StrokeColor: seriesColor,
			Font:        opt.Font,
			Series:      series,
			Range:       yr,
		})
	}
	// 最大、最小的mark point
	for _, renderer := range rendererList {
		_, err = renderer.Render()
		if err != nil {
			return chart.BoxZero, err
		}
	}

	return p.box, nil
}
