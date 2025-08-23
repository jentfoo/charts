package drawing

import (
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/math/fixed"
)

// DrawContour draws the given closed contour at the given sub-pixel offset.
func DrawContour(path PathBuilder, ps []truetype.Point, dx, dy float64) {
	if len(ps) == 0 {
		return
	}
	startX, startY := pointToF64Point(ps[0])
	path.MoveTo(startX+dx, startY+dy)
	q0X, q0Y, on0 := startX, startY, true
	for _, p := range ps[1:] {
		qX, qY := pointToF64Point(p)
		on := p.Flags&0x01 != 0
		if on {
			if on0 {
				path.LineTo(qX+dx, qY+dy)
			} else {
				path.QuadCurveTo(q0X+dx, q0Y+dy, qX+dx, qY+dy)
			}
		} else if !on0 {
			midX := (q0X + qX) / 2
			midY := (q0Y + qY) / 2
			path.QuadCurveTo(q0X+dx, q0Y+dy, midX+dx, midY+dy)
		}
		q0X, q0Y, on0 = qX, qY, on
	}
	// Close the curve.
	if on0 {
		path.LineTo(startX+dx, startY+dy)
	} else {
		path.QuadCurveTo(q0X+dx, q0Y+dy, startX+dx, startY+dy)
	}
}

// FontExtents contains font metric information.
type FontExtents struct {
	// Ascent is the distance that the text
	// extends above the baseline.
	Ascent float64

	// Descent is the distance that the text
	// extends below the baseline.  The descent
	// is given as a negative value.
	Descent float64

	// Height is the distance from the lowest
	// descending point to the highest ascending
	// point.
	Height float64
}

// Extents returns the FontExtents for a font.
// TODO needs to read this https://developer.apple.com/fonts/TrueType-Reference-Manual/RM02/Chap2.html#intro
func Extents(font *truetype.Font, size float64) FontExtents {
	bounds := font.Bounds(fixed.Int26_6(font.FUnitsPerEm()))
	scale := size / float64(font.FUnitsPerEm())
	return FontExtents{
		Ascent:  float64(bounds.Max.Y) * scale,
		Descent: float64(bounds.Min.Y) * scale,
		Height:  float64(bounds.Max.Y-bounds.Min.Y) * scale,
	}
}

// IsEmojiOrSymbol checks if a rune is likely an emoji or symbol that might not be in the font.
func IsEmojiOrSymbol(r rune) bool {
	return (r >= 0x1F600 && r <= 0x1F64F) || // Emoticons
		(r >= 0x1F300 && r <= 0x1F5FF) || // Misc Symbols and Pictographs
		(r >= 0x1F680 && r <= 0x1F6FF) || // Transport and Map Symbols
		(r >= 0x1F700 && r <= 0x1F77F) || // Alchemical Symbols
		(r >= 0x1F780 && r <= 0x1F7FF) || // Geometric Shapes Extended
		(r >= 0x1F800 && r <= 0x1F8FF) || // Supplemental Arrows-C
		(r >= 0x1F900 && r <= 0x1F9FF) || // Supplemental Symbols and Pictographs
		(r >= 0x1FA00 && r <= 0x1FA6F) || // Chess Symbols
		(r >= 0x1FA70 && r <= 0x1FAFF) || // Symbols and Pictographs Extended-A
		(r >= 0x2600 && r <= 0x26FF) || // Miscellaneous Symbols
		(r >= 0x2700 && r <= 0x27BF) || // Dingbats
		(r >= 0xFE00 && r <= 0xFE0F) || // Variation Selectors
		(r >= 0x1F000 && r <= 0x1F02F) || // Mahjong Tiles
		(r >= 0x1F030 && r <= 0x1F09F) || // Domino Tiles
		(r >= 0x1F0A0 && r <= 0x1F0FF) || // Playing Cards
		(r >= 0x23E9 && r <= 0x23EC) || // Play/Pause buttons
		(r >= 0x23F0 && r <= 0x23F3) || // Alarm Clock
		(r >= 0x25A0 && r <= 0x25FF) || // Geometric Shapes
		(r >= 0x2934 && r <= 0x2935) || // Arrow symbols
		(r >= 0x2B05 && r <= 0x2B07) || // Arrow symbols
		(r >= 0x2B1B && r <= 0x2B1C) || // Square symbols
		(r >= 0x2B50 && r <= 0x2B55) || // Star symbols
		(r == 0x3030) || (r == 0x303D) || // Wave dash, Part alternation mark
		(r >= 0x3297 && r <= 0x3299) // Circled ideographs
}
