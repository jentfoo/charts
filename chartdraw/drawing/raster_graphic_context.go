package drawing

import (
	"errors"
	"image"
	"image/color"
	"math"

	"github.com/golang/freetype/raster"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// DefaultDPI is the default image DPI.
const DefaultDPI = 96.0

// NewRasterGraphicContext creates a new Graphic context from an image.
func NewRasterGraphicContext(img *image.RGBA) *RasterGraphicContext {
	painter := raster.NewRGBAPainter(img)
	return NewRasterGraphicContextWithPainter(img, painter)
}

// NewRasterGraphicContextWithPainter creates a new Graphic context from an image and a Painter (see Freetype-go).
func NewRasterGraphicContextWithPainter(img draw.Image, painter Painter) *RasterGraphicContext {
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	return &RasterGraphicContext{
		StackGraphicContext: NewStackGraphicContext(),
		img:                 img,
		painter:             painter,
		fillRasterizer:      raster.NewRasterizer(width, height),
		strokeRasterizer:    raster.NewRasterizer(width, height),
		glyphBuf:            &truetype.GlyphBuf{},
		dpi:                 DefaultDPI,
		glyphCache:          make(map[glyphCacheKey]*glyphData),
	}
}

type glyphCacheKey struct {
	font    *truetype.Font
	scale   fixed.Int26_6
	glyph   truetype.Index
	hinting font.Hinting // TODO - keep?
}

type glyphData struct {
	contours [][]truetype.Point // each contour is a slice of truetype.Point
}

// RasterGraphicContext is the implementation of GraphicContext for a raster image.
type RasterGraphicContext struct {
	*StackGraphicContext
	img              draw.Image
	painter          Painter
	fillRasterizer   *raster.Rasterizer
	strokeRasterizer *raster.Rasterizer
	glyphBuf         *truetype.GlyphBuf
	dpi              float64
	glyphCache       map[glyphCacheKey]*glyphData
}

// SetDPI sets the screen resolution in dots per inch.
func (rgc *RasterGraphicContext) SetDPI(dpi float64) {
	rgc.dpi = dpi
	rgc.recalc()
}

// GetDPI returns the resolution of the Image GraphicContext
func (rgc *RasterGraphicContext) GetDPI() float64 {
	return rgc.dpi
}

// Clear fills the current canvas with a default transparent color
func (rgc *RasterGraphicContext) Clear() {
	width, height := rgc.img.Bounds().Dx(), rgc.img.Bounds().Dy()
	rgc.current.FillColor = color.Transparent
	rgc.FillRect(0, 0, width, height)
}

// FillRect draws a filled rectangle with the provided coordinates and the current set FillColor.
func (rgc *RasterGraphicContext) FillRect(x1, y1, x2, y2 int) {
	imageColor := image.NewUniform(rgc.current.FillColor)
	draw.Draw(rgc.img, image.Rect(x1, y1, x2, y2), imageColor, image.Point{}, draw.Over)
}

// DrawImage draws the raster image in the current canvas
func (rgc *RasterGraphicContext) DrawImage(img image.Image) {
	DrawImage(img, rgc.img, rgc.current.Tr, draw.Over, BilinearFilter)
}

// FillString draws the text at point (0, 0)
func (rgc *RasterGraphicContext) FillString(text string) (float64, error) {
	return rgc.FillStringAt(text, 0, 0)
}

// FillStringAt draws the text at the specified point (x, y)
func (rgc *RasterGraphicContext) FillStringAt(text string, x, y float64) (cursor float64, err error) {
	cursor, err = rgc.CreateStringPath(text, x, y)
	rgc.Fill()
	return
}

// StrokeString draws the contour of the text at point (0, 0)
func (rgc *RasterGraphicContext) StrokeString(text string) (float64, error) {
	return rgc.StrokeStringAt(text, 0, 0)
}

// StrokeStringAt draws the contour of the text at point (x, y)
func (rgc *RasterGraphicContext) StrokeStringAt(text string, x, y float64) (cursor float64, err error) {
	cursor, err = rgc.CreateStringPath(text, x, y)
	rgc.Stroke()
	return
}

// drawGlyph fetches (or builds) the path for the given glyph at current scale
// and adds it to rgc.current.Path at offset dx, dy.
func (rgc *RasterGraphicContext) drawGlyph(glyph truetype.Index, dx, dy float64) error {
	scale := fixed.Int26_6(rgc.current.Scale)
	hinting := font.HintingNone
	key := glyphCacheKey{
		font:    rgc.current.Font,
		scale:   scale,
		glyph:   glyph,
		hinting: hinting,
	}

	// Try to retrieve from cache
	gd, ok := rgc.glyphCache[key]
	if !ok {
		// If not found in cache, load the glyph from the font
		if err := rgc.glyphBuf.Load(rgc.current.Font, scale, glyph, hinting); err != nil {
			return err
		}
		// Build the glyphData from glyphBuf
		gd = &glyphData{}
		e0 := 0
		for _, e1 := range rgc.glyphBuf.Ends {
			pts := rgc.glyphBuf.Points[e0:e1]
			// Copy the points for this contour
			contourCopy := make([]truetype.Point, len(pts))
			copy(contourCopy, pts)
			gd.contours = append(gd.contours, contourCopy)
			e0 = e1
		}
		// Store in the cache
		rgc.glyphCache[key] = gd
	}

	// Use your existing DrawContour to add each contour to the path,
	// offset by dx, dy
	for _, contour := range gd.contours {
		DrawContour(rgc, contour, dx, dy)
	}
	return nil
}

// CreateStringPath creates a path from the string s at x, y, and returns the string width.
// The text is placed so that the left edge of the em square of the first character of s
// and the baseline intersect at x, y. The majority of the affected pixels will be
// above and to the right of the point, but some may be below or to the left.
// For example, drawing a string that starts with a 'J' in an italic font may
// affect pixels below and left of the point.
func (rgc *RasterGraphicContext) CreateStringPath(s string, x, y float64) (cursor float64, err error) {
	f := rgc.GetFont()
	if f == nil {
		return 0, errors.New("no font loaded, cannot continue")
	}
	rgc.recalc()

	startx := x
	prev, hasPrev := truetype.Index(0), false
	for _, rc := range s {
		index := f.Index(rc)
		if hasPrev {
			x += fUnitsToFloat64(f.Kern(fixed.Int26_6(rgc.current.Scale), prev, index))
		}
		if err = rgc.drawGlyph(index, x, y); err != nil {
			// Return the offset so far if something fails
			return x - startx, err
		}
		x += fUnitsToFloat64(f.HMetric(fixed.Int26_6(rgc.current.Scale), index).AdvanceWidth)
		prev, hasPrev = index, true
	}
	return x - startx, nil
}

// GetStringBounds returns the approximate pixel bounds of a string.
func (rgc *RasterGraphicContext) GetStringBounds(s string) (left, top, right, bottom float64, err error) {
	f := rgc.GetFont()
	if f == nil {
		return 0, 0, 0, 0, errors.New("no font loaded, cannot continue")
	}
	rgc.recalc()

	left = math.MaxFloat64
	top = math.MaxFloat64
	cursor := 0.0
	prev, hasPrev := truetype.Index(0), false
	for _, rc := range s {
		index := f.Index(rc)
		if hasPrev {
			cursor += fUnitsToFloat64(f.Kern(fixed.Int26_6(rgc.current.Scale), prev, index))
		}

		if err = rgc.glyphBuf.Load(rgc.current.Font, fixed.Int26_6(rgc.current.Scale), index, font.HintingNone); err != nil {
			return
		}
		e0 := 0
		for _, e1 := range rgc.glyphBuf.Ends {
			ps := rgc.glyphBuf.Points[e0:e1]
			for _, p := range ps {
				x, y := pointToF64Point(p)
				top = math.Min(top, y)
				bottom = math.Max(bottom, y)
				left = math.Min(left, x+cursor)
				right = math.Max(right, x+cursor)
			}
			e0 = e1
		}
		cursor += fUnitsToFloat64(f.HMetric(fixed.Int26_6(rgc.current.Scale), index).AdvanceWidth)
		prev, hasPrev = index, true
	}
	return
}

// recalc recalculates scale and bounds values from the font size, screen
// resolution and font metrics, and invalidates the glyph cache if needed.
//
// In this example we do NOT automatically invalidate the entire glyph cache,
// because the cache key includes the scale. If you change the DPI/FontSize
// frequently, you may want to flush or limit the cache.
func (rgc *RasterGraphicContext) recalc() {
	rgc.current.Scale = rgc.current.FontSizePoints * rgc.dpi
}

// SetFont sets the font used to draw text.
func (rgc *RasterGraphicContext) SetFont(font *truetype.Font) {
	rgc.current.Font = font
}

// GetFont returns the font used to draw text.
func (rgc *RasterGraphicContext) GetFont() *truetype.Font {
	return rgc.current.Font
}

// SetFontSize sets the font size in points (as in “a 12 point font”).
func (rgc *RasterGraphicContext) SetFontSize(fontSizePoints float64) {
	rgc.current.FontSizePoints = fontSizePoints
	rgc.recalc()
}

// paint applies the rasterizer to the image with the given color and resets path.
func (rgc *RasterGraphicContext) paint(rasterizer *raster.Rasterizer, c color.Color) {
	rgc.painter.SetColor(c)
	rasterizer.Rasterize(rgc.painter)
	rasterizer.Clear()
	rgc.current.Path.Clear()
}

// Stroke strokes the paths with the color specified by SetStrokeColor
func (rgc *RasterGraphicContext) Stroke(paths ...*Path) {
	if rgc.current.LineWidth == 0 {
		rgc.current.Path.Clear()
		return
	}
	paths = append(paths, rgc.current.Path)

	rgc.strokeRasterizer.UseNonZeroWinding = true

	stroker := NewLineStroker(rgc.current.Cap, rgc.current.Join, Transformer{
		Tr:        rgc.current.Tr,
		Flattener: FtLineBuilder{Adder: rgc.strokeRasterizer},
	})
	stroker.HalfLineWidth = rgc.current.LineWidth / 2

	var liner Flattener
	if len(rgc.current.Dash) > 0 {
		liner = NewDashVertexConverter(rgc.current.Dash, rgc.current.DashOffset, stroker)
	} else {
		liner = stroker
	}
	for _, p := range paths {
		Flatten(p, liner, rgc.current.Tr.GetScale())
	}

	rgc.paint(rgc.strokeRasterizer, rgc.current.StrokeColor)
}

func isRectanglePath(path *Path) bool {
	if len(path.Components) != 5 {
		return false
	} else if path.Components[0] != MoveToComponent {
		return false
	}
	for i := 1; i < 3; i++ {
		if path.Components[i] != LineToComponent {
			return false
		}
	}
	x1, y1 := path.Points[0], path.Points[1]
	x2, y2 := path.Points[2], path.Points[3]
	x3, y3 := path.Points[4], path.Points[5]
	var x4, y4 float64
	if path.Components[4] == LineToComponent {
		x4, y4 = path.Points[6], path.Points[7]
	} else if path.Components[4] == CloseComponent {
		x4 = x1
		y4 = y1
	} else {
		return false
	}

	// Check if opposite sides are equal
	return (x1 == x4 && x2 == x3 && y1 == y2 && y3 == y4) || (x1 == x2 && x3 == x4 && y1 == y4 && y2 == y3)
}

func getRectangleBounds(path *Path) (int, int, int, int) {
	x1, y1 := path.Points[0], path.Points[1]
	x2, y2 := path.Points[4], path.Points[5]
	if x2 < x1 {
		x1, x2 = x2, x1
	}
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	return int(math.Floor(x1)), int(math.Floor(y1)), int(math.Ceil(x2)), int(math.Ceil(y2))
}

// Fill fills the paths with the color specified by SetFillColor
func (rgc *RasterGraphicContext) Fill(paths ...*Path) {
	paths = append(paths, rgc.current.Path)
	pathCount := len(paths)
	if pathCount == 0 {
		return
	} else if pathCount == 1 && isRectanglePath(paths[0]) {
		// we can draw rectangles of a uniform color using a more efficient method
		x1, y1, x2, y2 := getRectangleBounds(paths[0])
		rgc.FillRect(x1, y1, x2, y2)
		rgc.current.Path.Clear() // draw complete
		return
	}

	rgc.fillRasterizer.UseNonZeroWinding = rgc.current.FillRule == FillRuleWinding

	flattener := Transformer{
		Tr:        rgc.current.Tr,
		Flattener: FtLineBuilder{Adder: rgc.fillRasterizer},
	}
	for _, p := range paths {
		Flatten(p, flattener, rgc.current.Tr.GetScale())
	}

	rgc.paint(rgc.fillRasterizer, rgc.current.FillColor)
}

// FillStroke first fills the paths and then strokes them
func (rgc *RasterGraphicContext) FillStroke(paths ...*Path) {
	paths = append(paths, rgc.current.Path)
	pathCount := len(paths)
	if pathCount == 0 {
		return
	} else if pathCount == 1 && isRectanglePath(paths[0]) {
		// we can draw rectangles of a uniform color using a more efficient method, then stroke the line after
		x1, y1, x2, y2 := getRectangleBounds(paths[0])
		rgc.FillRect(x1, y1, x2, y2)
		rgc.Stroke() // draw path for stroke
		return
	}

	rgc.fillRasterizer.UseNonZeroWinding = rgc.current.FillRule == FillRuleWinding
	rgc.strokeRasterizer.UseNonZeroWinding = true

	flattener := Transformer{
		Tr:        rgc.current.Tr,
		Flattener: FtLineBuilder{Adder: rgc.fillRasterizer},
	}

	stroker := NewLineStroker(rgc.current.Cap, rgc.current.Join, Transformer{
		Tr:        rgc.current.Tr,
		Flattener: FtLineBuilder{Adder: rgc.strokeRasterizer},
	})
	stroker.HalfLineWidth = rgc.current.LineWidth / 2

	var liner Flattener
	if len(rgc.current.Dash) > 0 {
		liner = NewDashVertexConverter(rgc.current.Dash, rgc.current.DashOffset, stroker)
	} else {
		liner = stroker
	}

	demux := DemuxFlattener{Flatteners: []Flattener{flattener, liner}}
	for _, p := range paths {
		Flatten(p, demux, rgc.current.Tr.GetScale())
	}

	// Fill
	rgc.paint(rgc.fillRasterizer, rgc.current.FillColor)
	// Stroke
	rgc.paint(rgc.strokeRasterizer, rgc.current.StrokeColor)
}
