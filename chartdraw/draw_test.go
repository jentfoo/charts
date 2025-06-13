package chartdraw

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-analyze/charts/chartdraw/drawing"
)

type boundedTest struct{}

func (boundedTest) Len() int { return 3 }
func (boundedTest) GetBoundedValues(i int) (float64, float64, float64) {
	return float64(i), float64(i + 1), float64(i)
}

func TestDrawBoundedAndHistogramSeries(t *testing.T) {
	t.Parallel()

	vr := SVG(60, 60).(*vectorRenderer)
	style := Style{StrokeColor: drawing.ColorBlack, FillColor: drawing.ColorWhite}

	Draw.BoundedSeries(vr, Box{Bottom: 50, Left: 0}, &ContinuousRange{Min: 0, Max: 2, Domain: 50}, &ContinuousRange{Min: 0, Max: 3, Domain: 50}, style, boundedTest{})
	hist := HistogramSeries{InnerSeries: ContinuousSeries{XValues: LinearRange(0, 2), YValues: LinearRange(1, 3)}}
	Draw.HistogramSeries(vr, Box{Bottom: 50, Left: 0}, &ContinuousRange{Min: 0, Max: 2, Domain: 50}, &ContinuousRange{Min: 0, Max: 3, Domain: 50}, style, hist)

	buf := bytes.Buffer{}
	require.NoError(t, vr.Save(&buf))
	out := buf.String()
	assert.Contains(t, out, "<path")
}

func TestDrawAnnotationAndText(t *testing.T) {
	t.Parallel()

	vr := SVG(80, 80).(*vectorRenderer)
	style := Style{FillColor: drawing.ColorWhite, StrokeColor: drawing.ColorBlack, FontStyle: FontStyle{Font: GetDefaultFont(), FontSize: 10}}
	Draw.Annotation(vr, style, 10, 10, "label")
	Draw.Text(vr, "hello", 5, 5, style)
	Draw.BoxRotated(vr, Box{Top: 1, Left: 1, Right: 10, Bottom: 10}, 45, style)
	Draw.BoxCorners(vr, Box{Top: 1, Left: 1, Right: 5, Bottom: 5}.Corners(), style)

	buf := bytes.Buffer{}
	require.NoError(t, vr.Save(&buf))
	out := buf.String()
	assert.Contains(t, out, "label")
	assert.Contains(t, out, "<path")
}

func TestDrawMeasureTextAndTextWithin(t *testing.T) {
	t.Parallel()

	vr := SVG(100, 50).(*vectorRenderer)
	vr.SetFont(GetDefaultFont())
	vr.SetFontSize(10)
	box := Draw.MeasureText(vr, "abc", Style{FontStyle: FontStyle{Font: GetDefaultFont(), FontSize: 10}})
	assert.NotZero(t, box.Width())

	Draw.TextWithin(vr, "abc", Box{Top: 0, Left: 0, Right: 50, Bottom: 20}, Style{FontStyle: FontStyle{Font: GetDefaultFont(), FontSize: 10}})
	buf := bytes.Buffer{}
	require.NoError(t, vr.Save(&buf))
	out := buf.String()
	assert.Contains(t, out, "<text")
}
