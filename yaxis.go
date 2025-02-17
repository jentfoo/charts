package charts

type YAxisOption struct {
	// Show specifies if the y-axis should be rendered, set this to *false (through Ptr(false)) to hide the axis.
	Show *bool
	// Theme specifies the colors used for the x-axis.
	Theme ColorPalette
	// Min, if set this will force the minimum value of y-axis.
	Min *float64
	// Max, if set this will force the maximum value of y-axis.
	Max *float64
	// RangeValuePaddingScale suggest a scale of padding added to the max and min values.
	RangeValuePaddingScale *float64
	// Labels provides labels for each value on the y-axis.
	Labels []string
	// Position describes the position of y-axis, it can be 'left' or 'right'.
	Position string
	// FontStyle specifies the font configuration for each label.
	FontStyle FontStyle
	// Formatter for replacing y-axis text values.
	Formatter string
	// Unit is a suggestion for how large the axis step is, this is a recommendation only. Larger numbers result in fewer labels.
	Unit float64
	// LabelCount is the number of labels to show on the axis. Specify a smaller number to reduce writing collisions.
	LabelCount int
	// LabelCountAdjustment specifies a relative influence on how many labels should be rendered.
	// Typically, this is negative to result in cleaner graphs, positive values may result in text collisions.
	LabelCountAdjustment int
	// LabelSkipCount specifies a number of lines between labels where there will be no label and instead just a horizontal line.
	LabelSkipCount int
	isCategoryAxis bool
	// SplitLineShow for showing axis split line, set this to true to show the horizontal axis split lines.
	SplitLineShow *bool
	// SpineLineShow can be set to enforce if the vertical spine on the axis should be shown or not.
	// By default, not shown unless a category axis.
	SpineLineShow *bool
	// ValueFormatter defines how float values should be rendered to strings, notably for numeric axis labels.
	ValueFormatter ValueFormatter
}

func (opt *YAxisOption) toAxisOption(fallbackTheme ColorPalette) axisOption {
	position := PositionLeft
	if opt.Position == PositionRight {
		position = PositionRight
	}
	theme := getPreferredTheme(opt.Theme, fallbackTheme)
	if opt.FontStyle.FontColor.IsZero() {
		opt.FontStyle.FontColor = theme.GetTextColor()
	}
	axisOpt := axisOption{
		show:                 opt.Show,
		labels:               opt.Labels,
		formatter:            opt.Formatter,
		position:             position,
		fontStyle:            opt.FontStyle,
		axisSplitLineColor:   theme.GetAxisSplitLineColor(),
		axisColor:            theme.GetYAxisStrokeColor(),
		strokeWidth:          -1,
		boundaryGap:          Ptr(false),
		unit:                 opt.Unit,
		labelCount:           opt.LabelCount,
		labelCountAdjustment: opt.LabelCountAdjustment,
		labelSkipCount:       opt.LabelSkipCount,
		splitLineShow:        true,
	}
	if opt.isCategoryAxis {
		axisOpt.boundaryGap = Ptr(true)
		axisOpt.strokeWidth = 1
		axisOpt.splitLineShow = false
	}
	if opt.SplitLineShow != nil {
		axisOpt.splitLineShow = *opt.SplitLineShow
	}
	if opt.SpineLineShow != nil {
		if *opt.SpineLineShow {
			axisOpt.strokeWidth = 1
		} else {
			axisOpt.strokeWidth = -1
		}
	}
	return axisOpt
}

// newLeftYAxis returns a left y-axis renderer.
func newLeftYAxis(p *Painter, opt YAxisOption) *axisPainter {
	p = p.Child(PainterPaddingOption(Box{
		Bottom: defaultXAxisHeight,
	}))
	return newAxisPainter(p, opt.toAxisOption(p.theme))
}

// newRightYAxis returns a right y-axis renderer.
func newRightYAxis(p *Painter, opt YAxisOption) *axisPainter {
	p = p.Child(PainterPaddingOption(Box{
		Bottom: defaultXAxisHeight,
	}))
	axisOpt := opt.toAxisOption(p.theme)
	axisOpt.position = PositionRight
	axisOpt.splitLineShow = false
	return newAxisPainter(p, axisOpt)
}
