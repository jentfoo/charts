package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-analyze/charts"
	"github.com/go-analyze/charts/chartdraw"
)

/*
Example of putting two charts together.
*/

func writeFile(buf []byte) error {
	tmpPath := "./tmp"
	if err := os.MkdirAll(tmpPath, 0700); err != nil {
		return err
	}

	file := filepath.Join(tmpPath, "multiple-charts-3.png")
	return os.WriteFile(file, buf, 0600)
}

func main() {
	values := [][]float64{
		{120, 132, 101, 134, 90, 230, 210},
		{150, 232, 201, 154, 190, 330, 410},
		{320, 332, 301, 334, 390, 330, 320},
	}
	p, err := charts.LineRender(
		values,
		charts.XAxisDataOptionFunc([]string{
			"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun",
		}),
		charts.LegendOptionFunc(charts.LegendOption{
			Data: []string{
				"Email", "Video Ads", "Direct",
			},
			OverlayChart: charts.False(),
			Offset: charts.OffsetStr{
				Top:  charts.PositionBottom,
				Left: "20%",
			},
		}),
		func(opt *charts.ChartOption) {
			opt.YAxis = []charts.YAxisOption{
				{
					Max: charts.FloatPointer(2000),
				},
			}
			opt.SymbolShow = charts.True()
			opt.LineStrokeWidth = 1.2
			opt.ValueFormatter = func(f float64) string {
				return fmt.Sprintf("%.0f", f)
			}

			opt.Children = []charts.ChartOption{
				{
					Box: chartdraw.NewBox(10, 200, 500, 200),
					SeriesList: charts.NewSeriesListDataFromValues([][]float64{
						{70, 90, 110, 130},
						{80, 100, 120, 140},
					}, charts.ChartTypeHorizontalBar),
					Legend: charts.LegendOption{
						Data: []string{
							"2011", "2012",
						},
					},
					YAxis: []charts.YAxisOption{
						{
							Data: []string{
								"USA", "India", "China", "World",
							},
						},
					},
				},
			}
		},
	)
	if err != nil {
		panic(err)
	}

	if buf, err := p.Bytes(); err != nil {
		panic(err)
	} else if err = writeFile(buf); err != nil {
		panic(err)
	}
}