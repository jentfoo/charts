package chartdraw

import (
	"bytes"
	"hash/crc32"
	"image"
	"image/png"
	"math"
	"testing"

	"github.com/go-analyze/charts/chartdraw/drawing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func hashImage(t *testing.T, r *rasterRenderer) uint32 {
	iw := &ImageWriter{}
	require.NoError(t, r.Save(iw))
	img, err := iw.Image()
	require.NoError(t, err)
	rgba := img.(*image.RGBA)
	return crc32.ChecksumIEEE(rgba.Pix)
}

func TestRasterRendererRotationAndSave(t *testing.T) {
	t.Parallel()

	rr := PNG(20, 20).(*rasterRenderer)
	x, y := rr.getCoords(5, 5)
	assert.Equal(t, 5, x)
	assert.Equal(t, 5, y)

	rr.SetTextRotation(math.Pi / 2)
	x, y = rr.getCoords(5, 5)
	assert.Zero(t, x)
	assert.Zero(t, y)

	iw := &ImageWriter{}
	require.NoError(t, rr.Save(iw))
	img, err := iw.Image()
	require.NoError(t, err)
	assert.Equal(t, 20, img.Bounds().Dx())
}

func TestRasterRendererSavePNG(t *testing.T) {
	t.Parallel()

	rr := PNG(10, 10).(*rasterRenderer)
	buf := bytes.Buffer{}
	require.NoError(t, rr.Save(&buf))
	img, err := png.Decode(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	assert.Equal(t, 10, img.Bounds().Dx())
}

func TestRasterRendererCircleHash(t *testing.T) {
	t.Parallel()

	rr := PNG(20, 20).(*rasterRenderer)
	rr.SetFillColor(drawing.ColorWhite)
	rr.SetStrokeColor(drawing.ColorRed)
	rr.MoveTo(3, 3)
	rr.LineTo(4, 4)
	rr.Circle(5, 10, 10)
	rr.FillStroke()

	h := hashImage(t, rr)
	assert.Equal(t, uint32(0xf767b6eb), h)
}

func TestRasterRendererRectangleHash(t *testing.T) {
	t.Parallel()

	rr := PNG(20, 20).(*rasterRenderer)
	rr.SetFillColor(drawing.ColorWhite)
	rr.SetStrokeColor(drawing.ColorRed)
	rr.MoveTo(2, 2)
	rr.LineTo(18, 2)
	rr.LineTo(18, 18)
	rr.LineTo(2, 18)
	rr.Close()
	rr.FillStroke()

	h := hashImage(t, rr)
	assert.Equal(t, uint32(0xcb26bf6d), h)
}

func TestRasterRendererArcHash(t *testing.T) {
	t.Parallel()

	rr := PNG(20, 20).(*rasterRenderer)
	rr.SetFillColor(drawing.ColorWhite)
	rr.SetStrokeColor(drawing.ColorRed)
	rr.MoveTo(10, 10)
	rr.ArcTo(10, 10, 8, 8, 0, math.Pi)
	rr.FillStroke()

	h := hashImage(t, rr)
	assert.Equal(t, uint32(0x8a33cae6), h)
}

func TestRasterRendererQuadHash(t *testing.T) {
	t.Parallel()

	rr := PNG(20, 20).(*rasterRenderer)
	rr.SetStrokeColor(drawing.ColorBlue)
	rr.SetStrokeWidth(1)
	rr.MoveTo(2, 18)
	rr.QuadCurveTo(10, 0, 18, 18)
	rr.Stroke()

	h := hashImage(t, rr)
	assert.Equal(t, uint32(0x02e4fd0e), h)
}

func TestRasterRendererTextHash(t *testing.T) {
	t.Parallel()

	rr := PNG(50, 20).(*rasterRenderer)
	rr.SetFont(GetDefaultFont())
	rr.SetFontSize(10)
	rr.SetFontColor(drawing.ColorBlack)
	rr.Text("hi", 2, 12)

	h := hashImage(t, rr)
	assert.Equal(t, uint32(0x41d7b6c8), h)
}
