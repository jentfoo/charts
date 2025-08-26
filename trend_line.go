package charts

import (
	"errors"
	"math"

	"github.com/go-analyze/charts/chartdraw"
	"github.com/go-analyze/charts/chartdraw/matrix"
)

// NewTrendLine returns a trend line for the provided type. Set on a specific Series instance.
func NewTrendLine(trendType string) []SeriesTrendLine {
	return []SeriesTrendLine{
		{
			Type: trendType,
		},
	}
}

// trendLinePainter is responsible for rendering trend lines on the chart.
type trendLinePainter struct {
	p       *Painter
	options []trendLineRenderOption
}

// newTrendLinePainter returns a new trend line renderer.
func newTrendLinePainter(p *Painter) *trendLinePainter {
	return &trendLinePainter{
		p: p,
	}
}

// add appends a trend line render option.
func (t *trendLinePainter) add(opt trendLineRenderOption) {
	t.options = append(t.options, opt)
}

// trendLineRenderOption holds configuration for rendering trend lines.
type trendLineRenderOption struct {
	defaultStrokeColor Color
	// xValues are the x-coordinates for each data sample.
	xValues []int
	// seriesValues are the raw data values.
	seriesValues []float64
	// axisRange is used to transform a raw data value into a screen y-coordinate.
	axisRange axisRange
	// trends are the list of trend lines to render for this series.
	trends []SeriesTrendLine
	// dashed indicates if the trend line will be a dashed line.
	dashed bool
}

// Render computes and draws all configured trend lines.
func (t *trendLinePainter) Render() (Box, error) {
	painter := t.p
	for _, opt := range t.options {
		if len(opt.trends) == 0 || len(opt.seriesValues) == 0 || len(opt.xValues) == 0 {
			continue
		}

		for _, trend := range opt.trends {
			if trend.Window != 0 && trend.Period == 0 {
				trend.Period = trend.Window
			}
			var fitted []float64
			var err error
			switch trend.Type {
			case SeriesTrendTypeLinear:
				fitted, err = linearTrend(opt.seriesValues)
			case SeriesTrendTypeCubic:
				fitted, err = cubicTrend(opt.seriesValues)
			case SeriesTrendTypeSMA, "average" /* long term backwards compatibility */ :
				fitted, err = movingAverageTrend(opt.seriesValues, trend.Period)
			case SeriesTrendTypeEMA:
				fitted, err = exponentialMovingAverageTrend(opt.seriesValues, trend.Period)
			case SeriesTrendTypeBollingerUpper:
				fitted, err = bollingerUpperTrend(opt.seriesValues, trend.Period)
			case SeriesTrendTypeBollingerLower:
				fitted, err = bollingerLowerTrend(opt.seriesValues, trend.Period)
			case SeriesTrendTypeRSI:
				fitted, err = rsiTrend(opt.seriesValues, trend.Period)
			default:
				// Unknown trend type; skip.
				continue
			}
			if err != nil {
				return BoxZero, err
			} else if len(fitted) != len(opt.xValues) {
				return BoxZero, errors.New("mismatched data length in trend line computation")
			}

			color := trend.LineColor
			if color.IsTransparent() {
				color = opt.defaultStrokeColor
			}
			strokeWidth := trend.LineStrokeWidth
			if strokeWidth == 0 {
				strokeWidth = defaultStrokeWidth
			}

			// Convert fitted data to screen points.
			points := make([]Point, len(fitted))
			for i, val := range fitted {
				points[i] = Point{
					X: opt.xValues[i],
					Y: opt.axisRange.getRestHeight(val),
				}
			}

			// Determine if this trend line should be dashed
			isDashed := opt.dashed // start with chart default
			if trend.DashedLine != nil {
				isDashed = *trend.DashedLine
			}

			if isDashed {
				// Calculate dash size based on painter dimensions for better visibility
				avgDimension := float64(t.p.box.Width()+t.p.box.Height()) / 2
				dashLength := math.Max(avgDimension*0.02, 4.0) // Minimum 4px, scale with size
				gapLength := dashLength * 0.8
				dashArray := []float64{dashLength, gapLength}
				if trend.StrokeSmoothingTension > 0 {
					painter.SmoothDashedLineStroke(points, trend.StrokeSmoothingTension, color, strokeWidth, dashArray)
				} else {
					painter.DashedLineStroke(points, color, strokeWidth, dashArray)
				}
			} else {
				if trend.StrokeSmoothingTension > 0 {
					painter.SmoothLineStroke(points, trend.StrokeSmoothingTension, color, strokeWidth)
				} else {
					painter.LineStroke(points, color, strokeWidth)
				}
			}
		}
	}
	return BoxZero, nil
}

// linearTrend computes a linear regression over the provided data.
func linearTrend(y []float64) ([]float64, error) {
	n := float64(len(y))
	if n < 2 {
		return nil, errors.New("not enough data points for linear trend")
	}

	var sumX, sumY, sumXY, sumXX float64
	for i, v := range y {
		x := float64(i)
		sumX += x
		sumY += v
		sumXY += x * v
		sumXX += x * x
	}

	denom := n*sumXX - sumX*sumX
	if math.Abs(denom) < matrix.DefaultEpsilon {
		return nil, errors.New("degenerate x values for linear regression")
	}
	slope := (n*sumXY - sumX*sumY) / denom
	intercept := (sumY - slope*sumX) / n

	fitted := make([]float64, len(y))
	for i := range y {
		fitted[i] = intercept + slope*float64(i)
	}
	return fitted, nil
}

// cubicTrend computes a cubic (degree 3) polynomial regression over the data.
// If there are fewer than 4 points, it falls back to a linear trend.
func cubicTrend(y []float64) ([]float64, error) {
	n := len(y)
	if n < 2 {
		return nil, errors.New("not enough data points for cubic trend")
	} else if n < 4 {
		return linearTrend(y)
	}

	// Compute sums of powers of x.
	var S [7]float64 // S[k] = Σ x^k for k = 0..6.
	for i := 0; i < n; i++ {
		x := float64(i)
		xp := 1.0
		for k := 0; k <= 6; k++ {
			S[k] += xp
			xp *= x
		}
	}

	// Compute the right-hand side vector B.
	var B [4]float64
	for i := 0; i < n; i++ {
		x := float64(i)
		xp := 1.0
		for j := 0; j < 4; j++ {
			B[j] += y[i] * xp
			xp *= x
		}
	}

	// Build the augmented matrix for the normal equations.
	M := make([][]float64, 4)
	for j := 0; j < 4; j++ {
		M[j] = make([]float64, 5)
		for k := 0; k < 4; k++ {
			M[j][k] = S[j+k]
		}
		M[j][4] = B[j]
	}

	coeffs, err := solveLinearSystem(M)
	if err != nil {
		// fallback to linear
		return linearTrend(y)
	}

	fitted := make([]float64, n)
	for i := 0; i < n; i++ {
		x := float64(i)
		fitted[i] = coeffs[0] + coeffs[1]*x + coeffs[2]*x*x + coeffs[3]*x*x*x
	}
	return fitted, nil
}

// movingAverageTrend computes a moving average over the data using the given window size.
// If window is <= 0, a default based on the data size is used.
func movingAverageTrend(y []float64, window int) ([]float64, error) {
	n := len(y)
	if n < 2 {
		return nil, errors.New("not enough data points for average trend")
	} else if n < 4 {
		return linearTrend(y)
	}
	if window <= 0 {
		window = chartdraw.MaxInt(2, n/5)
	}

	fitted := make([]float64, n)
	var sum float64
	for i := 0; i < n; i++ {
		sum += y[i]
		if i >= window {
			sum -= y[i-window]
			fitted[i] = sum / float64(window)
		} else {
			fitted[i] = sum / float64(i+1)
		}
	}
	return fitted, nil
}

// exponentialMovingAverageTrend computes an exponential moving average over the data using the given window size.
// If window is <= 0, a default based on the data size is used.
func exponentialMovingAverageTrend(y []float64, window int) ([]float64, error) {
	n := len(y)
	if n < 2 {
		return nil, errors.New("not enough data points for exponential trend")
	} else if n < 4 {
		return linearTrend(y)
	}
	if window <= 0 {
		window = chartdraw.MaxInt(2, n/5)
	}

	multiplier := 2.0 / (float64(window) + 1.0)
	fitted := make([]float64, n)

	// First value is the same as input
	fitted[0] = y[0]

	// Calculate EMA for each subsequent value
	for i := 1; i < n; i++ {
		fitted[i] = (y[i] * multiplier) + (fitted[i-1] * (1 - multiplier))
	}

	return fitted, nil
}

// solveLinearSystem solves a 4x4 linear system represented as an augmented matrix.
// The input matrix has 4 rows and 5 columns (last column is the constants vector).
func solveLinearSystem(mat [][]float64) ([]float64, error) {
	n := len(mat)
	// Forward elimination
	for i := 0; i < n; i++ {
		// Find the pivot row
		maxRow := i
		for j := i + 1; j < n; j++ {
			if math.Abs(mat[j][i]) > math.Abs(mat[maxRow][i]) {
				maxRow = j
			}
		}
		mat[i], mat[maxRow] = mat[maxRow], mat[i]
		if math.Abs(mat[i][i]) < matrix.DefaultEpsilon {
			return nil, errors.New("singular matrix in cubic regression")
		}
		// Eliminate below
		for j := i + 1; j < n; j++ {
			factor := mat[j][i] / mat[i][i]
			for k := i; k <= n; k++ {
				mat[j][k] -= factor * mat[i][k]
			}
		}
	}
	// Back substitution
	sol := make([]float64, n)
	for i := n - 1; i >= 0; i-- {
		sol[i] = mat[i][n]
		for j := i + 1; j < n; j++ {
			sol[i] -= mat[i][j] * sol[j]
		}
		sol[i] /= mat[i][i]
	}
	return sol, nil
}

// bollingerUpperTrend computes the upper Bollinger Band (SMA + 2 * standard deviation).
func bollingerUpperTrend(y []float64, period int) ([]float64, error) {
	if len(y) < 2 {
		return nil, errors.New("not enough data points for Bollinger upper band")
	}
	if period <= 0 {
		period = chartdraw.MaxInt(2, len(y)/5)
	}
	if period > len(y) {
		return nil, errors.New("invalid period for Bollinger upper band")
	}

	// Calculate SMA first
	sma, err := movingAverageTrend(y, period)
	if err != nil {
		return nil, err
	}

	result := make([]float64, len(y))
	for i := 0; i < len(y); i++ {
		if i < period-1 {
			result[i] = GetNullValue()
			continue
		}

		// Calculate standard deviation for this period
		mean := sma[i]
		variance := 0.0
		for j := i - period + 1; j <= i; j++ {
			diff := y[j] - mean
			variance += diff * diff
		}
		stddev := math.Sqrt(variance / float64(period))

		result[i] = mean + (stddev * 2.0) // Using fixed 2.0 multiplier
	}

	return result, nil
}

// bollingerLowerTrend computes the lower Bollinger Band (SMA - 2 * standard deviation).
func bollingerLowerTrend(y []float64, period int) ([]float64, error) {
	if len(y) < 2 {
		return nil, errors.New("not enough data points for Bollinger lower band")
	}
	if period <= 0 {
		period = chartdraw.MaxInt(2, len(y)/5)
	}
	if period > len(y) {
		return nil, errors.New("invalid period for Bollinger lower band")
	}

	// Calculate SMA first
	sma, err := movingAverageTrend(y, period)
	if err != nil {
		return nil, err
	}

	result := make([]float64, len(y))
	for i := 0; i < len(y); i++ {
		if i < period-1 {
			result[i] = GetNullValue()
			continue
		}

		// Calculate standard deviation for this period
		mean := sma[i]
		variance := 0.0
		for j := i - period + 1; j <= i; j++ {
			diff := y[j] - mean
			variance += diff * diff
		}
		stddev := math.Sqrt(variance / float64(period))

		result[i] = mean - (stddev * 2.0) // Using fixed 2.0 multiplier
	}

	return result, nil
}

// rsiTrend computes the Relative Strength Index momentum oscillator.
func rsiTrend(y []float64, period int) ([]float64, error) {
	if len(y) < 2 {
		return nil, errors.New("not enough data points for RSI")
	}
	if period <= 0 {
		period = chartdraw.MaxInt(2, len(y)/5)
	}
	if len(y) < period+1 {
		return nil, errors.New("insufficient data for RSI")
	}

	result := make([]float64, len(y))

	// Initialize first values as null
	for i := 0; i < period; i++ {
		result[i] = GetNullValue()
	}

	// Calculate price changes
	gains := make([]float64, len(y)-1)
	losses := make([]float64, len(y)-1)

	for i := 1; i < len(y); i++ {
		change := y[i] - y[i-1]
		if change > 0 {
			gains[i-1] = change
			losses[i-1] = 0
		} else {
			gains[i-1] = 0
			losses[i-1] = -change
		}
	}

	// Calculate initial averages
	var avgGain, avgLoss float64
	for i := 0; i < period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	// Calculate RSI
	for i := period; i < len(y); i++ {
		if avgLoss == 0 {
			result[i] = 100
		} else {
			rs := avgGain / avgLoss
			result[i] = 100 - (100 / (1 + rs))
		}

		// Update averages for next iteration
		if i < len(gains) {
			avgGain = ((avgGain * float64(period-1)) + gains[i]) / float64(period)
			avgLoss = ((avgLoss * float64(period-1)) + losses[i]) / float64(period)
		}
	}

	return result, nil
}
