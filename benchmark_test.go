package charts

import (
	"errors"
	"testing"
)

func makeDefaultMultiChartOptions() ChartOption {
	opt := ChartOption{
		Legend: LegendOption{
			Offset: OffsetStr{
				Top: "-90",
			},
			SeriesNames: []string{"Milk Tea", "Matcha Latte", "Cheese Cocoa", "Walnut Brownie"},
		},
		Padding: NewBoxEqual(10),
		XAxis: XAxisOption{
			Labels: []string{"2012", "2013", "2014", "2015", "2016", "2017"},
		},
		YAxis: []YAxisOption{
			{
				Min: Ptr(0.0),
				Max: Ptr(90.0),
			},
		},
		SeriesList: append(
			NewSeriesListLine([][]float64{
				{56.5, 82.1, 88.7, 70.1, 53.4, 85.1},
				{51.1, 51.4, 55.1, 53.3, 73.8, 68.7},
			}).ToGenericSeriesList(),
			NewSeriesListBar([][]float64{
				{40.1, 62.2, 69.5, 36.4, 45.2, 32.5},
				{25.2, 37.1, 41.2, 18, 33.9, 49.1},
			}).ToGenericSeriesList()...),
		Children: []ChartOption{
			{
				Legend: LegendOption{
					Show:        Ptr(false),
					SeriesNames: []string{"Milk Tea", "Matcha Latte", "Cheese Cocoa", "Walnut Brownie"},
				},
				Box: Box{
					Top:    20,
					Left:   400,
					Right:  500,
					Bottom: 120,
				},
				Radius: "35%",
				SeriesList: NewSeriesListDoughnut([]float64{
					435.9, 354.3, 285.9, 204.5,
				}).ToGenericSeriesList(),
			},
		},
	}
	pointMarks := NewSeriesMarkList(SeriesMarkTypeMin, SeriesMarkTypeMax)
	lineMarks := NewSeriesMarkList(SeriesMarkTypeAverage)
	opt.SeriesList[0].MarkPoint.Points = pointMarks
	opt.SeriesList[0].MarkLine.Lines = lineMarks
	opt.SeriesList[2].MarkPoint.Points = pointMarks
	opt.SeriesList[2].MarkLine.Lines = lineMarks
	return opt
}

func BenchmarkChartOptionMultiChartPNGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		opt := makeDefaultMultiChartOptions()
		opt.OutputFormat = ChartOutputPNG

		if d, err := Render(opt); err != nil {
			panic(err)
		} else if buf, err := d.Bytes(); err != nil {
			panic(err)
		} else if len(buf) == 0 {
			panic(errors.New("data is nil"))
		}
	}
}

func BenchmarkChartOptionMultiChartJPGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		opt := makeDefaultMultiChartOptions()
		opt.OutputFormat = ChartOutputJPG

		if d, err := Render(opt); err != nil {
			panic(err)
		} else if buf, err := d.Bytes(); err != nil {
			panic(err)
		} else if len(buf) == 0 {
			panic(errors.New("data is nil"))
		}
	}
}

func BenchmarkChartOptionMultiChartSVGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		opt := makeDefaultMultiChartOptions()
		opt.OutputFormat = ChartOutputSVG

		if d, err := Render(opt); err != nil {
			panic(err)
		} else if buf, err := d.Bytes(); err != nil {
			panic(err)
		} else if len(buf) == 0 {
			panic(errors.New("data is nil"))
		}
	}
}

func BenchmarkPainterFunnelChartPNGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		painter := NewPainter(PainterOptions{
			OutputFormat: ChartOutputPNG,
		})

		renderPainterFunnel(painter)
	}
}

func BenchmarkPainterFunnelChartJPGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		painter := NewPainter(PainterOptions{
			OutputFormat: ChartOutputJPG,
		})

		renderPainterFunnel(painter)
	}
}

func BenchmarkPainterFunnelChartSVGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		painter := NewPainter(PainterOptions{
			OutputFormat: ChartOutputSVG,
		})

		renderPainterFunnel(painter)
	}
}

func renderPainterFunnel(painter *Painter) {
	funnelOpt := NewFunnelChartOptionWithData([]float64{100, 80, 60, 40, 20, 10, 2})
	funnelOpt.Title.Text = "Funnel"
	funnelOpt.Legend.SeriesNames = []string{"Show", "Click", "Visit", "Inquiry", "Order", "Pay", "Cancel"}
	funnelOpt.Legend.Padding = Box{Left: 100}
	if err := painter.FunnelChart(funnelOpt); err != nil {
		panic(err)
	}

	if buf, err := painter.Bytes(); err != nil {
		panic(err)
	} else if len(buf) == 0 {
		panic(errors.New("data is nil"))
	}
}

func BenchmarkPainterLineChartPNGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		painter := NewPainter(PainterOptions{
			OutputFormat: ChartOutputPNG,
		})

		renderPainterLine(painter)
	}
}

func BenchmarkPainterLineChartJPGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		painter := NewPainter(PainterOptions{
			OutputFormat: ChartOutputJPG,
		})

		renderPainterLine(painter)
	}
}

func BenchmarkPainterLineChartSVGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		painter := NewPainter(PainterOptions{
			OutputFormat: ChartOutputSVG,
		})

		renderPainterLine(painter)
	}
}

func renderPainterLine(painter *Painter) {
	lineOpt := NewLineChartOptionWithData([][]float64{
		{56.5, 82.1, 88.7, 70.1, 53.4, 85.1},
		{51.1, 51.4, 55.1, 53.3, 73.8, 68.7},
	})
	pointMarks := NewSeriesMarkList(SeriesMarkTypeMin, SeriesMarkTypeMax)
	lineMarks := NewSeriesMarkList(SeriesMarkTypeAverage)
	lineOpt.SeriesList[0].MarkPoint.Points = pointMarks
	lineOpt.SeriesList[0].MarkLine.Lines = lineMarks
	lineOpt.SeriesList[1].MarkPoint.Points = pointMarks
	lineOpt.SeriesList[1].MarkLine.Lines = lineMarks
	lineOpt.Legend = LegendOption{
		Offset: OffsetStr{
			Top: "-90",
		},
		SeriesNames: []string{"Milk Tea", "Matcha Latte", "Cheese Cocoa", "Walnut Brownie"},
	}
	lineOpt.XAxis = XAxisOption{
		Labels: []string{"2012", "2013", "2014", "2015", "2016", "2017"},
	}
	lineOpt.YAxis = []YAxisOption{
		{
			Min: Ptr(0.0),
			Max: Ptr(90.0),
		},
	}
	if err := painter.LineChart(lineOpt); err != nil {
		panic(err)
	} else if buf, err := painter.Bytes(); err != nil {
		panic(err)
	} else if len(buf) == 0 {
		panic(errors.New("data is nil"))
	}
}

func BenchmarkPainterScatterChartPNGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		painter := NewPainter(PainterOptions{
			OutputFormat: ChartOutputPNG,
		})

		renderPainterScatter(painter)
	}
}

func BenchmarkPainterScatterChartJPGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		painter := NewPainter(PainterOptions{
			OutputFormat: ChartOutputJPG,
		})

		renderPainterScatter(painter)
	}
}

func BenchmarkPainterScatterChartSVGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		painter := NewPainter(PainterOptions{
			OutputFormat: ChartOutputSVG,
		})

		renderPainterScatter(painter)
	}
}

func renderPainterScatter(painter *Painter) {
	lineOpt := NewScatterChartOptionWithData([][]float64{
		{56.5, 82.1, 88.7, 70.1, 53.4, 85.1},
		{51.1, 51.4, 55.1, 53.3, 73.8, 68.7},
	})
	lineMarks := NewSeriesMarkList(SeriesMarkTypeAverage)
	lineOpt.SeriesList[0].MarkLine.Lines = lineMarks
	lineOpt.SeriesList[0].TrendLine = NewTrendLine(SeriesTrendTypeCubic)
	lineOpt.SeriesList[1].MarkLine.Lines = lineMarks
	lineOpt.SeriesList[1].TrendLine = NewTrendLine(SeriesTrendTypeAverage)
	lineOpt.Legend = LegendOption{
		Offset: OffsetStr{
			Top: "-90",
		},
		SeriesNames: []string{"Milk Tea", "Matcha Latte", "Cheese Cocoa", "Walnut Brownie"},
	}
	lineOpt.XAxis = XAxisOption{
		Labels: []string{"2012", "2013", "2014", "2015", "2016", "2017"},
	}
	lineOpt.YAxis = []YAxisOption{
		{
			Min: Ptr(0.0),
			Max: Ptr(90.0),
		},
	}
	if err := painter.ScatterChart(lineOpt); err != nil {
		panic(err)
	} else if buf, err := painter.Bytes(); err != nil {
		panic(err)
	} else if len(buf) == 0 {
		panic(errors.New("data is nil"))
	}
}

func BenchmarkPainterBarChartPNGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		painter := NewPainter(PainterOptions{
			OutputFormat: ChartOutputPNG,
		})

		renderPainterBar(painter)
	}
}

func BenchmarkPainterBarChartJPGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		painter := NewPainter(PainterOptions{
			OutputFormat: ChartOutputJPG,
		})

		renderPainterBar(painter)
	}
}

func BenchmarkPainterBarChartSVGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		painter := NewPainter(PainterOptions{
			OutputFormat: ChartOutputSVG,
		})

		renderPainterBar(painter)
	}
}

func renderPainterBar(painter *Painter) {
	barOpt := NewBarChartOptionWithData([][]float64{
		{2.0, 4.9, 7.0, 23.2, 25.6, 76.7, 135.6, 162.2, 32.6, 20.0, 6.4, 3.3},
		{2.6, 5.9, 9.0, 26.4, 28.7, 70.7, 175.6, 182.2, 48.7, 18.8, 6.0, 2.3},
	})
	pointMarks := NewSeriesMarkList(SeriesMarkTypeMin, SeriesMarkTypeMax)
	lineMarks := NewSeriesMarkList(SeriesMarkTypeAverage)
	barOpt.SeriesList[0].MarkPoint.Points = pointMarks
	barOpt.SeriesList[0].MarkLine.Lines = lineMarks
	barOpt.SeriesList[1].MarkPoint.Points = pointMarks
	barOpt.SeriesList[1].MarkLine.Lines = lineMarks
	barOpt.XAxis.Labels = []string{
		"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
	}
	barOpt.XAxis.LabelCount = 12
	barOpt.Legend = LegendOption{
		SeriesNames:  []string{"Rainfall", "Evaporation"},
		Offset:       OffsetRight,
		OverlayChart: Ptr(true),
	}
	if err := painter.BarChart(barOpt); err != nil {
		panic(err)
	} else if buf, err := painter.Bytes(); err != nil {
		panic(err)
	} else if len(buf) == 0 {
		panic(errors.New("data is nil"))
	}
}

func BenchmarkPainterPieChartPNGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		painter := NewPainter(PainterOptions{
			OutputFormat: ChartOutputPNG,
		})

		renderPainterPie(painter)
	}
}

func BenchmarkPainterPieChartJPGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		painter := NewPainter(PainterOptions{
			OutputFormat: ChartOutputJPG,
		})

		renderPainterPie(painter)
	}
}

func BenchmarkPainterPieChartSVGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		painter := NewPainter(PainterOptions{
			OutputFormat: ChartOutputSVG,
		})

		renderPainterPie(painter)
	}
}

func renderPainterPie(painter *Painter) {
	pieOpt := NewPieChartOptionWithData([]float64{1048, 735, 580, 484, 300})
	pieOpt.Title = TitleOption{
		Text:             "Pie Chart",
		Subtext:          "(Fake Data)",
		Offset:           OffsetCenter,
		FontStyle:        NewFontStyleWithSize(16),
		SubtextFontStyle: NewFontStyleWithSize(10),
	}
	pieOpt.Padding = NewBoxEqual(20)
	pieOpt.Legend = LegendOption{
		SeriesNames: []string{"Search Engine", "Direct", "Email", "Union Ads", "Video Ads"},
		Vertical:    Ptr(true),
		Offset: OffsetStr{
			Left: "80%",
			Top:  PositionBottom,
		},
		FontStyle: NewFontStyleWithSize(10),
	}
	if err := painter.PieChart(pieOpt); err != nil {
		panic(err)
	} else if buf, err := painter.Bytes(); err != nil {
		panic(err)
	} else if len(buf) == 0 {
		panic(errors.New("data is nil"))
	}
}

func BenchmarkPainterDoughnutChartPNGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		painter := NewPainter(PainterOptions{
			OutputFormat: ChartOutputPNG,
		})

		renderPainterDoughnut(painter)
	}
}

func BenchmarkPainterDoughnutChartJPGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		painter := NewPainter(PainterOptions{
			OutputFormat: ChartOutputJPG,
		})

		renderPainterDoughnut(painter)
	}
}

func BenchmarkPainterDoughnutChartSVGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		painter := NewPainter(PainterOptions{
			OutputFormat: ChartOutputSVG,
		})

		renderPainterDoughnut(painter)
	}
}

func renderPainterDoughnut(painter *Painter) {
	doughnutOpt := NewDoughnutChartOptionWithData([]float64{1048, 735, 580, 484, 300})
	doughnutOpt.Title = TitleOption{
		Text:             "Doughnut Chart",
		Subtext:          "(Fake Data)",
		Offset:           OffsetCenter,
		FontStyle:        NewFontStyleWithSize(16),
		SubtextFontStyle: NewFontStyleWithSize(10),
	}
	doughnutOpt.Padding = NewBoxEqual(20)
	doughnutOpt.Legend = LegendOption{
		SeriesNames: []string{"Search Engine", "Direct", "Email", "Union Ads", "Video Ads"},
		Vertical:    Ptr(true),
		Offset: OffsetStr{
			Left: "80%",
			Top:  PositionBottom,
		},
		FontStyle: NewFontStyleWithSize(10),
	}
	if err := painter.DoughnutChart(doughnutOpt); err != nil {
		panic(err)
	} else if buf, err := painter.Bytes(); err != nil {
		panic(err)
	} else if len(buf) == 0 {
		panic(errors.New("data is nil"))
	}
}

func BenchmarkPainterRadarChartPNGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		painter := NewPainter(PainterOptions{
			OutputFormat: ChartOutputPNG,
		})

		renderPainterRadar(painter)
	}
}

func BenchmarkPainterRadarChartJPGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		painter := NewPainter(PainterOptions{
			OutputFormat: ChartOutputJPG,
		})

		renderPainterRadar(painter)
	}
}

func BenchmarkPainterRadarChartSVGRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		painter := NewPainter(PainterOptions{
			OutputFormat: ChartOutputSVG,
		})

		renderPainterRadar(painter)
	}
}

func renderPainterRadar(painter *Painter) {
	radarOpt := NewRadarChartOptionWithData([][]float64{
		{4200, 3000, 20000, 35000, 50000, 18000},
		{5000, 14000, 28000, 26000, 42000, 21000},
	}, []string{
		"Sales",
		"Administration",
		"Information Technology",
		"Customer Support",
		"Development",
		"Marketing",
	}, []float64{6500, 16000, 30000, 38000, 52000, 25000})
	radarOpt.Title = TitleOption{
		Text:      "Basic Radar Chart",
		FontStyle: NewFontStyleWithSize(16),
	}
	radarOpt.Legend = LegendOption{
		SeriesNames: []string{"Allocated Budget", "Actual Spending"},
		Offset:      OffsetRight,
	}
	if err := painter.RadarChart(radarOpt); err != nil {
		panic(err)
	} else if buf, err := painter.Bytes(); err != nil {
		panic(err)
	} else if len(buf) == 0 {
		panic(errors.New("data is nil"))
	}
}
