package chartdraw

import (
	"bytes"
	"hash/crc32"
	"image"
	"image/png"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	rr.MoveTo(3, 3)
	rr.LineTo(4, 4)
	rr.Circle(5, 10, 10)
	rr.FillStroke()

	iw := &ImageWriter{}
	require.NoError(t, rr.Save(iw))
	img, err := iw.Image()
	require.NoError(t, err)
	rgba := img.(*image.RGBA)

	h := crc32.ChecksumIEEE(rgba.Pix)
	assert.Equal(t, uint32(0xc5117d35), h)
}
