package charts

import (
	"github.com/golang/freetype/truetype"
)

// NewMarkPoint returns a mark point for the provided types. Set on a specific Series instance.
func NewMarkPoint(markPointTypes ...string) SeriesMarkPoint {
	return SeriesMarkPoint{
		Points: NewSeriesMarkList(markPointTypes...),
	}
}

type markPointPainter struct {
	p       *Painter
	options []markPointRenderOption
}

func (m *markPointPainter) add(opt markPointRenderOption) {
	if opt.valueFormatter == nil {
		opt.valueFormatter = defaultValueFormatter
	}
	if opt.symbolSize == 0 {
		opt.symbolSize = 28
	}
	m.options = append(m.options, opt)
}

type markPointRenderOption struct {
	fillColor          Color
	font               *truetype.Font
	symbolSize         int
	seriesValues       []float64
	markpoints         []SeriesMark
	seriesLabelPainter *seriesLabelPainter
	points             []Point
	valueFormatter     ValueFormatter
}

// newMarkPointPainter returns a mark point renderer.
func newMarkPointPainter(p *Painter) *markPointPainter {
	return &markPointPainter{
		p: p,
	}
}

func (m *markPointPainter) Render() (Box, error) {
	painter := m.p
	for _, opt := range m.options {
		if len(opt.markpoints) == 0 {
			continue
		}
		summary := summarizePopulationData(opt.seriesValues)
		textStyle := FontStyle{
			FontSize: defaultLabelFontSize,
			Font:     opt.font,
		}
		if isLightColor(opt.fillColor) {
			textStyle.FontColor = defaultLightFontColor
		} else {
			textStyle.FontColor = defaultDarkFontColor
		}
		for _, markPointData := range opt.markpoints {
			textStyle.FontSize = defaultLabelFontSize
			index := summary.MinIndex
			value := summary.Min
			text := opt.valueFormatter(value)

			switch markPointData.Type {
			case SeriesMarkTypeMax:
				index = summary.MaxIndex
				value = summary.Max
				text = opt.valueFormatter(value)
			case SeriesMarkTypePattern:
				// For pattern marks, use the specific index and pattern value
				if markPointData.Index != nil {
					index = *markPointData.Index
					if index >= len(opt.points) {
						continue // skip invalid index
					}

					// Use pattern name as value if available
					if markPointData.Value != nil {
						if str, ok := markPointData.Value.(string); ok {
							text = str
						}
					} else {
						text = markPointData.PatternType
					}
				} else {
					continue // pattern marks require an index
				}
			}

			if index >= len(opt.points) {
				continue // skip invalid index
			}

			p := opt.points[index]
			if opt.seriesLabelPainter != nil {
				// the series label has been replaced by our MarkPoint
				// This is why MarkPoints must be rendered BEFORE series labels
				opt.seriesLabelPainter.values[index].Text = ""
			}

			// Use different colors for bullish vs bearish patterns
			fillColor := opt.fillColor
			if markPointData.Type == SeriesMarkTypePattern {
				switch markPointData.PatternType {
				case PatternEngulfingBull, PatternHammer:
					fillColor = Color{R: 34, G: 197, B: 94, A: 255} // Green
				case PatternEngulfingBear, PatternShootingStar:
					fillColor = Color{R: 239, G: 68, B: 68, A: 255} // Red
				case PatternDoji, PatternGravestone, PatternDragonfly:
					fillColor = Color{R: 255, G: 193, B: 7, A: 255} // Yellow/Orange
				}
			}

			painter.Pin(p.X, p.Y-opt.symbolSize>>1, opt.symbolSize, fillColor, fillColor, 0.0)
			textBox := painter.MeasureText(text, 0, textStyle)
			if textStyle.FontSize > smallLabelFontSize && textBox.Width() > opt.symbolSize {
				textStyle.FontSize = smallLabelFontSize
				textBox = painter.MeasureText(text, 0, textStyle)
			}
			painter.Text(text, p.X-textBox.Width()>>1, p.Y-opt.symbolSize>>1-2, 0, textStyle)
		}
	}
	return BoxZero, nil
}
