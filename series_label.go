package charts

import (
	"github.com/golang/freetype/truetype"

	"github.com/go-analyze/charts/chartdraw"
)

type labelRenderValue struct {
	Text      string
	FontStyle FontStyle
	X         int
	Y         int
	Radians   float64
}

type labelValue struct {
	index     int
	value     float64
	x         int
	y         int
	radians   float64
	fontStyle FontStyle
	vertical  bool
	offset    OffsetInt
}

type seriesLabelPainter struct {
	p           *Painter
	seriesNames []string
	label       *SeriesLabel
	theme       ColorPalette
	font        *truetype.Font
	values      []labelRenderValue
}

func newSeriesLabelPainter(p *Painter, seriesNames []string, label SeriesLabel,
	theme ColorPalette, font *truetype.Font) *seriesLabelPainter {
	return &seriesLabelPainter{
		p:           p,
		seriesNames: seriesNames,
		label:       &label,
		theme:       theme,
		font:        font,
		values:      make([]labelRenderValue, 0),
	}
}

func (o *seriesLabelPainter) Add(value labelValue) {
	label := o.label
	distance := label.Distance
	if distance == 0 {
		distance = 5
	}
	text := labelFormatValue(o.seriesNames, label.Formatter, value.index, value.value, -1)
	labelStyle := FontStyle{
		FontColor: o.theme.GetTextColor(),
		FontSize:  labelFontSize,
		Font:      getPreferredFont(label.FontStyle.Font, value.fontStyle.Font, o.font),
	}
	if label.FontStyle.FontSize != 0 {
		labelStyle.FontSize = label.FontStyle.FontSize
	} else if value.fontStyle.FontSize != 0 {
		labelStyle.FontSize = value.fontStyle.FontSize
	}
	if !label.FontStyle.FontColor.IsZero() {
		labelStyle.FontColor = label.FontStyle.FontColor
	} else if !value.fontStyle.FontColor.IsZero() {
		labelStyle.FontColor = value.fontStyle.FontColor
	}
	p := o.p
	p.OverrideDrawingStyle(chartdraw.Style{FontStyle: labelStyle})
	rotated := value.radians != 0
	if rotated {
		p.setTextRotation(value.radians)
	}
	textBox := p.MeasureText(text)
	renderValue := labelRenderValue{
		Text:      text,
		FontStyle: labelStyle,
		X:         value.x,
		Y:         value.y,
		Radians:   value.radians,
	}
	if value.vertical {
		renderValue.X -= textBox.Width() >> 1
		renderValue.Y -= distance
	} else {
		renderValue.X += distance
		renderValue.Y += textBox.Height() >> 1
		renderValue.Y -= 2
	}
	if value.radians != 0 {
		renderValue.X = value.x + (textBox.Width() >> 1) - 1
		p.clearTextRotation()
	} else if textBox.Width()%2 != 0 {
		renderValue.X++
	}
	renderValue.X += value.offset.Left
	renderValue.Y += value.offset.Top
	o.values = append(o.values, renderValue)
}

func (o *seriesLabelPainter) Render() (Box, error) {
	for _, item := range o.values {
		o.p.OverrideFontStyle(item.FontStyle)
		if item.Radians != 0 {
			o.p.TextRotation(item.Text, item.X, item.Y, item.Radians)
		} else {
			o.p.Text(item.Text, item.X, item.Y)
		}
	}
	return chartdraw.BoxZero, nil
}
