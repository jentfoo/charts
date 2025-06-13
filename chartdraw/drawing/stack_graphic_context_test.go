package drawing

import (
	"math"
	"testing"

	"github.com/go-analyze/charts/chartdraw/roboto"
	"github.com/golang/freetype/truetype"
	"github.com/stretchr/testify/assert"
)

func TestStackGraphicContextSaveRestore(t *testing.T) {
	t.Parallel()

	gc := NewStackGraphicContext()
	gc.SetLineWidth(2)
	gc.MoveTo(1, 1)
	gc.Save()
	gc.SetLineWidth(4)
	gc.LineTo(2, 2)
	gc.Restore()
	assert.InDelta(t, 2.0, gc.current.LineWidth, 0.0001)
	x, y := gc.LastPoint()
	assert.InDelta(t, 1.0, x, 0.0001)
	assert.InDelta(t, 1.0, y, 0.0001)
}

func TestStackGraphicContextTransforms(t *testing.T) {
	t.Parallel()

	gc := NewStackGraphicContext()
	gc.Translate(2, 3)
	tr := gc.GetMatrixTransform()
	x, y := tr.TransformPoint(0, 0)
	assert.InDelta(t, 2.0, x, 0.0001)
	assert.InDelta(t, 3.0, y, 0.0001)
	gc.Rotate(math.Pi / 2)
	tr = gc.GetMatrixTransform()
	x, y = tr.TransformPoint(1, 0)
	assert.InDelta(t, 2.0, x, 0.0001)
	assert.InDelta(t, 4.0, y, 0.0001)
}

func TestStackGraphicContextColors(t *testing.T) {
	t.Parallel()

	gc := NewStackGraphicContext()
	gc.SetStrokeColor(ColorRed)
	gc.SetFillColor(ColorBlue)
	assert.Equal(t, ColorRed, gc.current.StrokeColor)
	assert.Equal(t, ColorBlue, gc.current.FillColor)
}

func TestStackMatrixTransform(t *testing.T) {
	t.Parallel()

	gc := NewStackGraphicContext()
	tr := NewTranslationMatrix(5, 7)
	gc.SetMatrixTransform(tr)
	got := gc.GetMatrixTransform()
	x, y := got.TransformPoint(0, 0)
	assert.InDelta(t, 5.0, x, 0.0001)
	assert.InDelta(t, 7.0, y, 0.0001)
}

func TestStackComposeMatrixTransform(t *testing.T) {
	t.Parallel()

	gc := NewStackGraphicContext()
	gc.SetMatrixTransform(NewTranslationMatrix(5, 7))
	gc.ComposeMatrixTransform(NewTranslationMatrix(3, 4))
	got := gc.GetMatrixTransform()
	x, y := got.TransformPoint(0, 0)
	assert.InDelta(t, 8.0, x, 0.0001)
	assert.InDelta(t, 11.0, y, 0.0001)
}

func TestStackLineDash(t *testing.T) {
	t.Parallel()

	gc := NewStackGraphicContext()
	dash := []float64{1, 2, 3}
	gc.SetLineDash(dash, 0.5)
	assert.Equal(t, dash, gc.current.Dash)
	assert.InDelta(t, 0.5, gc.current.DashOffset, 0.0001)
}

func TestStackFontRoundTrip(t *testing.T) {
	t.Parallel()

	f, err := truetype.Parse(roboto.Roboto)
	assert.NoError(t, err)

	gc := NewStackGraphicContext()
	gc.SetFont(f)
	gc.SetFontSize(13.0)
	assert.Equal(t, f, gc.GetFont())
	assert.InDelta(t, 13.0, gc.GetFontSize(), 0.0001)
}
