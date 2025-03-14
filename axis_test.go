package charts

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAxis(t *testing.T) {
	t.Parallel()

	dayLabels := []string{
		"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun",
	}
	letterLabels := []string{"A", "B", "C", "D", "E", "F", "G"}

	tests := []struct {
		name          string
		optionFactory func() axisOption
		padPainter    bool
		result        string
	}{
		{
			name:       "x-axis_bottom",
			padPainter: true,
			optionFactory: func() axisOption {
				opt := XAxisOption{
					Labels:      dayLabels,
					BoundaryGap: Ptr(true),
					FontStyle:   NewFontStyleWithSize(18),
				}
				return opt.toAxisOption(nil)
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 50 325\nL 50 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 121 325\nL 121 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 192 325\nL 192 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 264 325\nL 264 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 335 325\nL 335 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 407 325\nL 407 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 478 325\nL 478 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 550 325\nL 550 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 50 320\nL 550 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"62\" y=\"353\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\">Mon</text><text x=\"136\" y=\"353\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\">Tue</text><text x=\"205\" y=\"353\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\">Wed</text><text x=\"279\" y=\"353\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\">Thu</text><text x=\"358\" y=\"353\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\">Fri</text><text x=\"425\" y=\"353\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\">Sat</text><text x=\"494\" y=\"353\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\">Sun</text></svg>",
		},
		{
			name:       "x-axis_bottom_rotation45",
			padPainter: true,
			optionFactory: func() axisOption {
				opt := XAxisOption{
					Labels:        dayLabels,
					BoundaryGap:   Ptr(true),
					FontStyle:     NewFontStyleWithSize(18),
					LabelRotation: DegreesToRadians(45),
				}
				return opt.toAxisOption(nil)
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 50 325\nL 50 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 121 325\nL 121 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 192 325\nL 192 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 264 325\nL 264 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 335 325\nL 335 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 407 325\nL 407 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 478 325\nL 478 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 550 325\nL 550 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 50 320\nL 550 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"61\" y=\"346\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(45.00,61,346)\">Mon</text><text x=\"135\" y=\"346\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(45.00,135,346)\">Tue</text><text x=\"204\" y=\"346\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(45.00,204,346)\">Wed</text><text x=\"278\" y=\"346\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(45.00,278,346)\">Thu</text><text x=\"354\" y=\"346\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(45.00,354,346)\">Fri</text><text x=\"423\" y=\"346\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(45.00,423,346)\">Sat</text><text x=\"493\" y=\"346\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(45.00,493,346)\">Sun</text></svg>",
		},
		{
			name:       "x-axis_bottom_rotation90",
			padPainter: true,
			optionFactory: func() axisOption {
				opt := XAxisOption{
					Labels:        dayLabels,
					BoundaryGap:   Ptr(true),
					FontStyle:     NewFontStyleWithSize(18),
					LabelRotation: DegreesToRadians(90),
				}
				return opt.toAxisOption(nil)
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 50 325\nL 50 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 121 325\nL 121 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 192 325\nL 192 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 264 325\nL 264 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 335 325\nL 335 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 407 325\nL 407 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 478 325\nL 478 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 550 325\nL 550 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 50 320\nL 550 320\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"74\" y=\"330\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(90.00,74,330)\">Mon</text><text x=\"145\" y=\"330\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(90.00,145,330)\">Tue</text><text x=\"217\" y=\"330\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(90.00,217,330)\">Wed</text><text x=\"288\" y=\"330\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(90.00,288,330)\">Thu</text><text x=\"360\" y=\"330\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(90.00,360,330)\">Fri</text><text x=\"431\" y=\"330\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(90.00,431,330)\">Sat</text><text x=\"503\" y=\"330\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(90.00,503,330)\">Sun</text></svg>",
		},
		{
			name:       "x-axis_top_rotation45",
			padPainter: true,
			optionFactory: func() axisOption {
				opt := XAxisOption{
					Labels:        dayLabels,
					BoundaryGap:   Ptr(true),
					FontStyle:     NewFontStyleWithSize(18),
					Position:      PositionTop,
					LabelRotation: DegreesToRadians(45),
				}
				return opt.toAxisOption(nil)
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 50 314\nL 50 309\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 121 314\nL 121 309\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 192 314\nL 192 309\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 264 314\nL 264 309\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 335 314\nL 335 309\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 407 314\nL 407 309\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 478 314\nL 478 309\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 550 314\nL 550 309\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 50 314\nL 550 314\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"61\" y=\"281\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(45.00,61,281)\">Mon</text><text x=\"135\" y=\"281\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(45.00,135,281)\">Tue</text><text x=\"204\" y=\"281\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(45.00,204,281)\">Wed</text><text x=\"278\" y=\"281\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(45.00,278,281)\">Thu</text><text x=\"354\" y=\"281\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(45.00,354,281)\">Fri</text><text x=\"423\" y=\"281\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(45.00,423,281)\">Sat</text><text x=\"493\" y=\"281\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(45.00,493,281)\">Sun</text></svg>",
		},
		{
			name:       "x-axis_top_rotation90",
			padPainter: true,
			optionFactory: func() axisOption {
				opt := XAxisOption{
					Labels:        dayLabels,
					BoundaryGap:   Ptr(true),
					FontStyle:     NewFontStyleWithSize(18),
					Position:      PositionTop,
					LabelRotation: DegreesToRadians(90),
				}
				return opt.toAxisOption(nil)
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 50 316\nL 50 311\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 121 316\nL 121 311\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 192 316\nL 192 311\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 264 316\nL 264 311\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 335 316\nL 335 311\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 407 316\nL 407 311\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 478 316\nL 478 311\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 550 316\nL 550 311\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 50 316\nL 550 316\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"74\" y=\"269\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(90.00,74,269)\">Mon</text><text x=\"145\" y=\"269\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(90.00,145,269)\">Tue</text><text x=\"217\" y=\"269\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(90.00,217,269)\">Wed</text><text x=\"288\" y=\"269\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(90.00,288,269)\">Thu</text><text x=\"360\" y=\"269\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(90.00,360,269)\">Fri</text><text x=\"431\" y=\"269\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(90.00,431,269)\">Sat</text><text x=\"503\" y=\"269\" style=\"stroke:none;fill:rgb(70,70,70);font-size:23px;font-family:'Roboto Medium',sans-serif\" transform=\"rotate(90.00,503,269)\">Sun</text></svg>",
		},
		{
			name: "x-axis_bottom_splitline",
			optionFactory: func() axisOption {
				return axisOption{
					labels:        dayLabels,
					splitLineShow: true,
				}
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 0 375\nL 0 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 85 375\nL 85 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 171 375\nL 171 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 257 375\nL 257 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 342 375\nL 342 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 428 375\nL 428 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 514 375\nL 514 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 600 375\nL 600 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 0 370\nL 600 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"27\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Mon</text><text x=\"115\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Tue</text><text x=\"199\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Wed</text><text x=\"286\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Thu</text><text x=\"376\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Fri</text><text x=\"460\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Sat</text><text x=\"544\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Sun</text><path  d=\"M 85 0\nL 85 370\" style=\"stroke-width:1;stroke:black;fill:none\"/><path  d=\"M 171 0\nL 171 370\" style=\"stroke-width:1;stroke:black;fill:none\"/><path  d=\"M 257 0\nL 257 370\" style=\"stroke-width:1;stroke:black;fill:none\"/><path  d=\"M 342 0\nL 342 370\" style=\"stroke-width:1;stroke:black;fill:none\"/><path  d=\"M 428 0\nL 428 370\" style=\"stroke-width:1;stroke:black;fill:none\"/><path  d=\"M 514 0\nL 514 370\" style=\"stroke-width:1;stroke:black;fill:none\"/><path  d=\"M 600 0\nL 600 370\" style=\"stroke-width:1;stroke:black;fill:none\"/></svg>",
		},
		{
			name: "x-axis_bottom_left",
			optionFactory: func() axisOption {
				return axisOption{
					labels:      dayLabels,
					boundaryGap: Ptr(false),
				}
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 0 375\nL 0 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 100 375\nL 100 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 200 375\nL 200 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 300 375\nL 300 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 400 375\nL 400 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 500 375\nL 500 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 600 375\nL 600 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 0 370\nL 600 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"-1\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Mon</text><text x=\"99\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Tue</text><text x=\"199\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Wed</text><text x=\"299\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Thu</text><text x=\"399\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Fri</text><text x=\"499\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Sat</text><text x=\"573\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Sun</text></svg>",
		},
		{
			name: "y-axis_left",
			optionFactory: func() axisOption {
				opt := YAxisOption{
					Labels:         dayLabels,
					Position:       PositionLeft,
					isCategoryAxis: true,
				}
				return opt.toAxisOption(nil)
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 36 0\nL 41 0\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 36 57\nL 41 57\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 36 114\nL 41 114\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 36 171\nL 41 171\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 36 228\nL 41 228\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 36 285\nL 41 285\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 36 342\nL 41 342\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 36 400\nL 41 400\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 41 0\nL 41 400\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"0\" y=\"35\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Mon</text><text x=\"4\" y=\"92\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Tue</text><text x=\"0\" y=\"149\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Wed</text><text x=\"4\" y=\"206\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Thu</text><text x=\"13\" y=\"263\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Fri</text><text x=\"8\" y=\"320\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Sat</text><text x=\"4\" y=\"378\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Sun</text></svg>",
		},
		{
			name: "y-axis_center",
			optionFactory: func() axisOption {
				return axisOption{
					labels:        dayLabels,
					position:      PositionLeft,
					boundaryGap:   Ptr(false),
					splitLineShow: true,
				}
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 36 0\nL 41 0\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 36 66\nL 41 66\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 36 133\nL 41 133\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 36 200\nL 41 200\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 36 266\nL 41 266\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 36 333\nL 41 333\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 36 400\nL 41 400\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 41 0\nL 41 400\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"0\" y=\"7\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Mon</text><text x=\"4\" y=\"73\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Tue</text><text x=\"0\" y=\"140\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Wed</text><text x=\"4\" y=\"207\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Thu</text><text x=\"13\" y=\"273\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Fri</text><text x=\"8\" y=\"340\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Sat</text><text x=\"4\" y=\"407\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Sun</text><path  d=\"M 41 0\nL 600 0\" style=\"stroke-width:1;stroke:black;fill:none\"/><path  d=\"M 41 66\nL 600 66\" style=\"stroke-width:1;stroke:black;fill:none\"/><path  d=\"M 41 133\nL 600 133\" style=\"stroke-width:1;stroke:black;fill:none\"/><path  d=\"M 41 200\nL 600 200\" style=\"stroke-width:1;stroke:black;fill:none\"/><path  d=\"M 41 266\nL 600 266\" style=\"stroke-width:1;stroke:black;fill:none\"/><path  d=\"M 41 333\nL 600 333\" style=\"stroke-width:1;stroke:black;fill:none\"/></svg>",
		},
		{
			name: "y-axis_right",
			optionFactory: func() axisOption {
				return axisOption{
					labels:        dayLabels,
					position:      PositionRight,
					boundaryGap:   Ptr(false),
					splitLineShow: true,
				}
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 559 0\nL 564 0\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 559 66\nL 564 66\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 559 133\nL 564 133\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 559 200\nL 564 200\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 559 266\nL 564 266\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 559 333\nL 564 333\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 559 400\nL 564 400\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 559 0\nL 559 400\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"569\" y=\"7\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Mon</text><text x=\"569\" y=\"73\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Tue</text><text x=\"569\" y=\"140\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Wed</text><text x=\"569\" y=\"207\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Thu</text><text x=\"569\" y=\"273\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Fri</text><text x=\"569\" y=\"340\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Sat</text><text x=\"569\" y=\"407\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Sun</text><path  d=\"M 0 0\nL 559 0\" style=\"stroke-width:1;stroke:black;fill:none\"/><path  d=\"M 0 66\nL 559 66\" style=\"stroke-width:1;stroke:black;fill:none\"/><path  d=\"M 0 133\nL 559 133\" style=\"stroke-width:1;stroke:black;fill:none\"/><path  d=\"M 0 200\nL 559 200\" style=\"stroke-width:1;stroke:black;fill:none\"/><path  d=\"M 0 266\nL 559 266\" style=\"stroke-width:1;stroke:black;fill:none\"/><path  d=\"M 0 333\nL 559 333\" style=\"stroke-width:1;stroke:black;fill:none\"/></svg>",
		},
		{
			name: "top",
			optionFactory: func() axisOption {
				return axisOption{
					labels:    dayLabels,
					formatter: "{value} --",
					position:  PositionTop,
				}
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 0 392\nL 0 387\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 85 392\nL 85 387\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 171 392\nL 171 387\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 257 392\nL 257 387\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 342 392\nL 342 387\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 428 392\nL 428 387\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 514 392\nL 514 387\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 600 392\nL 600 387\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 0 392\nL 600 392\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"20\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Mon --</text><text x=\"108\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Tue --</text><text x=\"192\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Wed --</text><text x=\"279\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Thu --</text><text x=\"369\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Fri --</text><text x=\"453\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Sat --</text><text x=\"537\" y=\"375\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Sun --</text></svg>",
		},
		{
			name: "reduced_label_count",
			optionFactory: func() axisOption {
				return axisOption{
					labels:               letterLabels,
					splitLineShow:        false,
					labelCountAdjustment: -1,
				}
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 0 375\nL 0 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 85 375\nL 85 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 171 375\nL 171 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 342 375\nL 342 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 428 375\nL 428 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 514 375\nL 514 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 600 375\nL 600 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 0 370\nL 600 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"-1\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">A</text><text x=\"123\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">B</text><text x=\"209\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">C</text><text x=\"381\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">E</text><text x=\"467\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">F</text><text x=\"589\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">G</text></svg>",
		},
		{
			name: "custom_unit",
			optionFactory: func() axisOption {
				return axisOption{
					labels:        letterLabels,
					splitLineShow: false,
					unit:          10,
				}
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 0 375\nL 0 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 342 375\nL 342 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 600 375\nL 600 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 0 370\nL 600 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"-1\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">A</text><text x=\"381\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">E</text><text x=\"589\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">G</text></svg>",
		},
		{
			name: "custom_font",
			optionFactory: func() axisOption {
				return axisOption{
					labels: letterLabels,
					labelFontStyle: FontStyle{
						FontSize:  40.0,
						FontColor: ColorBlue,
					},
				}
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 0 375\nL 0 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 85 375\nL 85 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 171 375\nL 171 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 257 375\nL 257 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 342 375\nL 342 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 428 375\nL 428 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 514 375\nL 514 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 600 375\nL 600 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 0 370\nL 600 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"25\" y=\"431\" style=\"stroke:none;fill:blue;font-size:51.1px;font-family:'Roboto Medium',sans-serif\">A</text><text x=\"112\" y=\"431\" style=\"stroke:none;fill:blue;font-size:51.1px;font-family:'Roboto Medium',sans-serif\">B</text><text x=\"197\" y=\"431\" style=\"stroke:none;fill:blue;font-size:51.1px;font-family:'Roboto Medium',sans-serif\">C</text><text x=\"282\" y=\"431\" style=\"stroke:none;fill:blue;font-size:51.1px;font-family:'Roboto Medium',sans-serif\">D</text><text x=\"371\" y=\"431\" style=\"stroke:none;fill:blue;font-size:51.1px;font-family:'Roboto Medium',sans-serif\">E</text><text x=\"457\" y=\"431\" style=\"stroke:none;fill:blue;font-size:51.1px;font-family:'Roboto Medium',sans-serif\">F</text><text x=\"540\" y=\"431\" style=\"stroke:none;fill:blue;font-size:51.1px;font-family:'Roboto Medium',sans-serif\">G</text></svg>",
		},
		{
			name: "boundary_gap_disable",
			optionFactory: func() axisOption {
				return axisOption{
					labels:      letterLabels,
					boundaryGap: Ptr(false),
				}
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 0 375\nL 0 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 100 375\nL 100 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 200 375\nL 200 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 300 375\nL 300 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 400 375\nL 400 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 500 375\nL 500 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 600 375\nL 600 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 0 370\nL 600 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"-1\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">A</text><text x=\"99\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">B</text><text x=\"199\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">C</text><text x=\"299\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">D</text><text x=\"399\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">E</text><text x=\"499\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">F</text><text x=\"589\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">G</text></svg>",
		},
		{
			name: "boundary_gap_enable",
			optionFactory: func() axisOption {
				return axisOption{
					labels:      letterLabels,
					boundaryGap: Ptr(true),
				}
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 0 375\nL 0 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 85 375\nL 85 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 171 375\nL 171 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 257 375\nL 257 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 342 375\nL 342 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 428 375\nL 428 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 514 375\nL 514 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 600 375\nL 600 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><path  d=\"M 0 370\nL 600 370\" style=\"stroke-width:1;stroke:rgb(110,112,121);fill:none\"/><text x=\"37\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">A</text><text x=\"123\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">B</text><text x=\"209\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">C</text><text x=\"294\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">D</text><text x=\"381\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">E</text><text x=\"467\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">F</text><text x=\"552\" y=\"395\" style=\"stroke:none;fill:rgb(70,70,70);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">G</text></svg>",
		},
	}

	axisTheme := MakeTheme(ThemeOption{
		IsDarkMode:         false,
		AxisStrokeColor:    Color{R: 110, G: 112, B: 121, A: 255},
		AxisSplitLineColor: ColorBlack,
		BackgroundColor:    ColorWhite,
		TextColor:          Color{R: 70, G: 70, B: 70, A: 255},
	})
	for i, tt := range tests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			p := NewPainter(PainterOptions{
				OutputFormat: ChartOutputSVG,
				Width:        600,
				Height:       400,
			}, PainterThemeOption(axisTheme))
			if tt.padPainter {
				p = p.Child(PainterPaddingOption(NewBoxEqual(50)))
			}

			_, err := newAxisPainter(p, tt.optionFactory()).Render()
			require.NoError(t, err)
			data, err := p.Bytes()
			require.NoError(t, err)
			assertEqualSVG(t, tt.result, data)
		})
	}
}
