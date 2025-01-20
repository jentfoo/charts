package drawing

import (
	"math"
)

const (
	// CurveRecursionLimit represents the maximum number of subdivisions.
	CurveRecursionLimit = 32
)

// TraceCubic flattens (subdivides) a cubic Bezier curve into line segments.
// 'cubic' must contain x1, y1, cx1, cy1, cx2, cy2, x2, y2, in that order.
// 'flatteningThreshold' controls how tightly the curve must be approximated.
func TraceCubic(t Liner, cubic []float64, flatteningThreshold float64) {
	// We store up to 32 curves on our stack, each occupying 8 floats (x1,y1,cx1,cy1,cx2,cy2,x2,y2).
	var stack [CurveRecursionLimit * 8]float64

	// Copy the initial cubic curve into the bottom of the stack.
	copy(stack[:8], cubic[:8])

	// 'i' is our top-of-stack index in terms of “which curve” we're processing, not byte offset.
	// Each curve is 8 floats, so an index i means slice [i*8 : i*8+8].
	i := 0

	for i >= 0 {
		// Current curve segment on top of stack.
		off := i * 8
		c := stack[off : off+8]

		// Measure how far the control points deviate from the chord (x1,y1 -> x2,y2).
		dx := c[6] - c[0]
		dy := c[7] - c[1]

		// We check “distance” from control points c[2],c[3], c[4],c[5] using cross-product magnitude.
		d2 := math.Abs((c[2]-c[6])*dy - (c[3]-c[7])*dx)
		d3 := math.Abs((c[4]-c[6])*dy - (c[5]-c[7])*dx)
		flatness := (d2 + d3) * (d2 + d3)

		// If the curve is flat enough or we've hit the limit, we just draw a line to the end.
		if flatness < flatteningThreshold*(dx*dx+dy*dy) || i == CurveRecursionLimit-1 {
			t.LineTo(c[6], c[7])
			i--
		} else {
			// Subdivide: we create two cubics out of the original.
			// We'll store the "first half" in stack[i+1], and reuse stack[i] for the "second half".
			nextOff := (i + 1) * 8

			// c1 is the first half of c, c2 will be c itself (reused).
			c1 := stack[nextOff : nextOff+8]
			c2 := c

			// SubdivideCubic inlined:
			// The first half c1:
			c1[0], c1[1] = c2[0], c2[1]
			c2[6], c2[7] = c2[6], c2[7] // (redundant but left for clarity)

			c1[2] = (c2[0] + c2[2]) / 2
			c1[3] = (c2[1] + c2[3]) / 2

			midX := (c2[2] + c2[4]) / 2
			midY := (c2[3] + c2[5]) / 2

			c2[4] = (c2[4] + c2[6]) / 2
			c2[5] = (c2[5] + c2[7]) / 2

			c1[4] = (c1[2] + midX) / 2
			c1[5] = (c1[3] + midY) / 2

			c2[2] = (midX + c2[4]) / 2
			c2[3] = (midY + c2[5]) / 2

			c1[6] = (c1[4] + c2[2]) / 2
			c1[7] = (c1[5] + c2[3]) / 2

			// The second half c2:
			c2[0], c2[1] = c1[6], c1[7]

			// Push the new first half on top of stack.
			i++
		}
	}
}

// TraceQuad flattens (subdivides) a quadratic Bezier curve into line segments.
// 'quad' must be [x1, y1, cx1, cy1, x2, y2].
// 'flatteningThreshold' is the same concept as with cubic, controlling flatness.
func TraceQuad(t Liner, quad []float64, flatteningThreshold float64) {
	// Each quad is 6 floats, so we can hold up to 32 of them on the stack.
	var stack [CurveRecursionLimit * 6]float64

	// Start by copying the single quad to the bottom of the stack.
	copy(stack[:6], quad[:6])

	i := 0
	for i >= 0 {
		off := i * 6
		c := stack[off : off+6]

		dx := c[4] - c[0]
		dy := c[5] - c[1]
		d := math.Abs((c[2]-c[4])*dy - (c[3]-c[5])*dx)
		// bail early if the distance is 0
		if d == 0 { // TODO - should we continue here?
			return
		}

		if d*d < flatteningThreshold*(dx*dx+dy*dy) || i == CurveRecursionLimit-1 {
			t.LineTo(c[4], c[5])
			i--
		} else {
			// SubdivideQuad inlined
			nextOff := (i + 1) * 6
			c1 := stack[nextOff : nextOff+6]
			c2 := c

			c1[0], c1[1] = c2[0], c2[1]
			c2[4], c2[5] = c2[4], c2[5] // (again, redundant for clarity)

			// Midpoints
			c1[2] = (c2[0] + c2[2]) / 2
			c1[3] = (c2[1] + c2[3]) / 2
			c2[2] = (c2[2] + c2[4]) / 2
			c2[3] = (c2[3] + c2[5]) / 2

			c1[4] = (c1[2] + c2[2]) / 2
			c1[5] = (c1[3] + c2[3]) / 2

			c2[0], c2[1] = c1[4], c1[5]
			i++
		}
	}
}

// TraceArc approximates an arc (x,y at center, radii rx,ry, from angle 'start' by 'angle').
// 'scale' is often used so smaller arcs subdivide fewer times.
func TraceArc(t Liner, x, y, rx, ry, start, angle, scale float64) (lastX, lastY float64) {
	end := start + angle
	clockWise := angle >= 0
	ra := (math.Abs(rx) + math.Abs(ry)) / 2
	da := math.Acos(ra/(ra+0.125/scale)) * 2
	//normalize
	if !clockWise {
		da = -da
	}
	step := start + da
	var curX, curY float64
	for {
		// If going forward clockwise, once 'step' passes 'end' by more than da/4, we stop.
		// If going backward (negative angle), once 'step' is below 'end' by more than da/4, we stop.
		if (step < end-da/4) != clockWise {
			curX = x + math.Cos(end)*rx
			curY = y + math.Sin(end)*ry
			return curX, curY
		}
		curX = x + math.Cos(step)*rx
		curY = y + math.Sin(step)*ry

		step += da
		t.LineTo(curX, curY)
	}
}
