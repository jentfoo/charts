package charts

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-analyze/charts/chartdraw"
	"github.com/go-analyze/charts/chartdraw/drawing"
)

func TestNewLegend(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		render func(*Painter) ([]byte, error)
		result string
	}{
		{
			name: "basic",
			render: func(p *Painter) ([]byte, error) {
				_, err := NewLegendPainter(p, LegendOption{
					Data: []string{"One", "Two", "Three"},
				}).Render()
				if err != nil {
					return nil, err
				}
				return p.Bytes()
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 184 9\nL 214 9\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><circle cx=\"199\" cy=\"9\" r=\"5\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><text x=\"216\" y=\"15\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">One</text><path  d=\"M 264 9\nL 294 9\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><circle cx=\"279\" cy=\"9\" r=\"5\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><text x=\"296\" y=\"15\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Two</text><path  d=\"M 346 9\nL 376 9\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><circle cx=\"361\" cy=\"9\" r=\"5\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><text x=\"378\" y=\"15\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Three</text></svg>",
		},
		{
			name: "position_left",
			render: func(p *Painter) ([]byte, error) {
				_, err := NewLegendPainter(p, LegendOption{
					Data:   []string{"One", "Two", "Three"},
					Offset: OffsetLeft,
				}).Render()
				if err != nil {
					return nil, err
				}
				return p.Bytes()
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 0 9\nL 30 9\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><circle cx=\"15\" cy=\"9\" r=\"5\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><text x=\"32\" y=\"15\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">One</text><path  d=\"M 80 9\nL 110 9\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><circle cx=\"95\" cy=\"9\" r=\"5\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><text x=\"112\" y=\"15\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Two</text><path  d=\"M 162 9\nL 192 9\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><circle cx=\"177\" cy=\"9\" r=\"5\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><text x=\"194\" y=\"15\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Three</text></svg>",
		},
		{
			name: "position_vertical_with_rect",
			render: func(p *Painter) ([]byte, error) {
				_, err := NewLegendPainter(p, LegendOption{
					Data:     []string{"One", "Two", "Three"},
					Vertical: true,
					Icon:     IconRect,
					Offset: OffsetStr{
						Left: "10%",
					},
				}).Render()
				if err != nil {
					return nil, err
				}
				return p.Bytes()
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 60 3\nL 90 3\nL 90 16\nL 60 16\nL 60 3\" style=\"stroke-width:0;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><text x=\"92\" y=\"15\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">One</text><path  d=\"M 60 23\nL 90 23\nL 90 36\nL 60 36\nL 60 23\" style=\"stroke-width:0;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><text x=\"92\" y=\"35\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Two</text><path  d=\"M 60 43\nL 90 43\nL 90 56\nL 60 56\nL 60 43\" style=\"stroke-width:0;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><text x=\"92\" y=\"55\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Three</text></svg>",
		},
		{
			name: "custom_padding_and_font",
			render: func(p *Painter) ([]byte, error) {
				_, err := NewLegendPainter(p, LegendOption{
					Data: []string{"One", "Two", "Three"},
					FontStyle: FontStyle{
						FontSize:  20.0,
						FontColor: drawing.ColorBlue,
					},
					Padding: chartdraw.NewBox(200, 20, 20, 20),
				}).Render()
				if err != nil {
					return nil, err
				}
				return p.Bytes()
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 151 204\nL 181 204\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><circle cx=\"166\" cy=\"204\" r=\"5\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><text x=\"183\" y=\"210\" style=\"stroke-width:0;stroke:none;fill:rgba(0,0,255,1.0);font-size:25.6px;font-family:'Roboto Medium',sans-serif\">One</text><path  d=\"M 249 204\nL 279 204\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><circle cx=\"264\" cy=\"204\" r=\"5\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><text x=\"281\" y=\"210\" style=\"stroke-width:0;stroke:none;fill:rgba(0,0,255,1.0);font-size:25.6px;font-family:'Roboto Medium',sans-serif\">Two</text><path  d=\"M 351 204\nL 381 204\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><circle cx=\"366\" cy=\"204\" r=\"5\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><text x=\"383\" y=\"210\" style=\"stroke-width:0;stroke:none;fill:rgba(0,0,255,1.0);font-size:25.6px;font-family:'Roboto Medium',sans-serif\">Three</text></svg>",
		},
		{
			name: "hidden",
			render: func(p *Painter) ([]byte, error) {
				_, err := NewLegendPainter(p, LegendOption{
					Data: []string{"A", "B", "C"},
					Show: False(),
				}).Render()
				if err != nil {
					return nil, err
				}
				return p.Bytes()
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"></svg>",
		},
		{
			name: "vertical_right_position",
			render: func(p *Painter) ([]byte, error) {
				_, err := NewLegendPainter(p, LegendOption{
					Data:     []string{"One", "Two Word", "Three Word Item", "Four Words Is Longer"},
					Vertical: true,
					Offset:   OffsetRight,
				}).Render()
				if err != nil {
					return nil, err
				}
				return p.Bytes()
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 421 9\nL 451 9\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><circle cx=\"436\" cy=\"9\" r=\"5\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><text x=\"453\" y=\"15\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">One</text><path  d=\"M 421 29\nL 451 29\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><circle cx=\"436\" cy=\"29\" r=\"5\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><text x=\"453\" y=\"35\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Two Word</text><path  d=\"M 421 49\nL 451 49\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><circle cx=\"436\" cy=\"49\" r=\"5\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><text x=\"453\" y=\"55\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Three Word Item</text><path  d=\"M 421 69\nL 451 69\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/><circle cx=\"436\" cy=\"69\" r=\"5\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/><text x=\"453\" y=\"75\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Four Words Is Longer</text></svg>",
		},
		{
			name: "vertical_right_position_custom_font_size",
			render: func(p *Painter) ([]byte, error) {
				_, err := NewLegendPainter(p, LegendOption{
					Data:     []string{"One", "Two Word", "Three Word Item", "Four Words Is Longer"},
					Vertical: true,
					Offset:   OffsetRight,
					FontStyle: FontStyle{
						FontSize: 6.0,
					},
				}).Render()
				if err != nil {
					return nil, err
				}
				return p.Bytes()
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 494 9\nL 524 9\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><circle cx=\"509\" cy=\"9\" r=\"5\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><text x=\"526\" y=\"15\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:7.7px;font-family:'Roboto Medium',sans-serif\">One</text><path  d=\"M 494 29\nL 524 29\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><circle cx=\"509\" cy=\"29\" r=\"5\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><text x=\"526\" y=\"35\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:7.7px;font-family:'Roboto Medium',sans-serif\">Two Word</text><path  d=\"M 494 49\nL 524 49\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><circle cx=\"509\" cy=\"49\" r=\"5\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><text x=\"526\" y=\"55\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:7.7px;font-family:'Roboto Medium',sans-serif\">Three Word Item</text><path  d=\"M 494 69\nL 524 69\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/><circle cx=\"509\" cy=\"69\" r=\"5\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/><text x=\"526\" y=\"75\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:7.7px;font-family:'Roboto Medium',sans-serif\">Four Words Is Longer</text></svg>",
		},
		{
			name: "vertical_right_position_with_padding",
			render: func(p *Painter) ([]byte, error) {
				_, err := NewLegendPainter(p, LegendOption{
					Data:     []string{"One", "Two Word", "Three Word Item", "Four Words Is Longer"},
					Vertical: true,
					Offset:   OffsetRight,
					Padding:  Box{Top: 120, Left: 120, Right: 120, Bottom: 120},
				}).Render()
				if err != nil {
					return nil, err
				}
				return p.Bytes()
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 301 124\nL 331 124\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><circle cx=\"316\" cy=\"124\" r=\"5\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><text x=\"333\" y=\"130\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">One</text><path  d=\"M 301 144\nL 331 144\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><circle cx=\"316\" cy=\"144\" r=\"5\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><text x=\"333\" y=\"150\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Two Word</text><path  d=\"M 301 164\nL 331 164\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><circle cx=\"316\" cy=\"164\" r=\"5\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><text x=\"333\" y=\"170\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Three Word Item</text><path  d=\"M 301 184\nL 331 184\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/><circle cx=\"316\" cy=\"184\" r=\"5\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/><text x=\"333\" y=\"190\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Four Words Is Longer</text></svg>",
		},
		{
			name: "left_position_overflow",
			render: func(p *Painter) ([]byte, error) {
				_, err := NewLegendPainter(p, LegendOption{
					Data: []string{"One", "Two Word", "Three Word Item", "Four Words Is Longer",
						"Five Words Is Even Longer", "Six Words Is The Longest Tested"},
					Offset: OffsetLeft,
				}).Render()
				if err != nil {
					return nil, err
				}
				return p.Bytes()
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 0 9\nL 30 9\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><circle cx=\"15\" cy=\"9\" r=\"5\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><text x=\"32\" y=\"15\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">One</text><path  d=\"M 80 9\nL 110 9\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><circle cx=\"95\" cy=\"9\" r=\"5\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><text x=\"112\" y=\"15\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Two Word</text><path  d=\"M 202 9\nL 232 9\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><circle cx=\"217\" cy=\"9\" r=\"5\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><text x=\"234\" y=\"15\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Three Word Item</text><path  d=\"M 369 9\nL 399 9\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/><circle cx=\"384\" cy=\"9\" r=\"5\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/><text x=\"401\" y=\"15\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Four Words Is Longer</text><path  d=\"M 0 24\nL 30 24\" style=\"stroke-width:3;stroke:rgba(115,192,222,1.0);fill:rgba(115,192,222,1.0)\"/><circle cx=\"15\" cy=\"24\" r=\"5\" style=\"stroke-width:3;stroke:rgba(115,192,222,1.0);fill:rgba(115,192,222,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(115,192,222,1.0);fill:rgba(115,192,222,1.0)\"/><text x=\"32\" y=\"30\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Five Words Is Even Longer</text><path  d=\"M 233 24\nL 263 24\" style=\"stroke-width:3;stroke:rgba(59,162,114,1.0);fill:rgba(59,162,114,1.0)\"/><circle cx=\"248\" cy=\"24\" r=\"5\" style=\"stroke-width:3;stroke:rgba(59,162,114,1.0);fill:rgba(59,162,114,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(59,162,114,1.0);fill:rgba(59,162,114,1.0)\"/><text x=\"265\" y=\"30\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Six Words Is The Longest Tested</text></svg>",
		},
		{
			name: "vertical_right_position_overflow",
			render: func(p *Painter) ([]byte, error) {
				_, err := NewLegendPainter(p, LegendOption{
					Data: []string{"One", "Two Word", "Three Word Item", "Four Words Is Longer",
						"Five Words Is Even Longer", "Six Words Is The Longest Tested"},
					Vertical: true,
					Offset: OffsetStr{
						Left: "440",
					},
				}).Render()
				if err != nil {
					return nil, err
				}
				return p.Bytes()
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><path  d=\"M 440 9\nL 470 9\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><circle cx=\"455\" cy=\"9\" r=\"5\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><text x=\"472\" y=\"15\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">One</text><path  d=\"M 440 29\nL 470 29\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><circle cx=\"455\" cy=\"29\" r=\"5\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><text x=\"472\" y=\"35\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Two Word</text><path  d=\"M 440 49\nL 470 49\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><circle cx=\"455\" cy=\"49\" r=\"5\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><text x=\"472\" y=\"55\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Three Word Item</text><path  d=\"M 440 69\nL 470 69\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/><circle cx=\"455\" cy=\"69\" r=\"5\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/><text x=\"472\" y=\"75\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Four Words Is Longer</text><path  d=\"M 440 89\nL 470 89\" style=\"stroke-width:3;stroke:rgba(115,192,222,1.0);fill:rgba(115,192,222,1.0)\"/><circle cx=\"455\" cy=\"89\" r=\"5\" style=\"stroke-width:3;stroke:rgba(115,192,222,1.0);fill:rgba(115,192,222,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(115,192,222,1.0);fill:rgba(115,192,222,1.0)\"/><text x=\"472\" y=\"95\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Five Words Is Even Longer</text><path  d=\"M 440 109\nL 470 109\" style=\"stroke-width:3;stroke:rgba(59,162,114,1.0);fill:rgba(59,162,114,1.0)\"/><circle cx=\"455\" cy=\"109\" r=\"5\" style=\"stroke-width:3;stroke:rgba(59,162,114,1.0);fill:rgba(59,162,114,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(59,162,114,1.0);fill:rgba(59,162,114,1.0)\"/><text x=\"472\" y=\"115\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Six Words Is The Longest Tested</text></svg>",
		},
		{
			name: "right_alignment",
			render: func(p *Painter) ([]byte, error) {
				_, err := NewLegendPainter(p, LegendOption{
					Data:  []string{"One", "Two Word", "Three Word Item", "Four Words Is Longer"},
					Align: AlignRight,
				}).Render()
				if err != nil {
					return nil, err
				}
				return p.Bytes()
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><text x=\"27\" y=\"15\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">One</text><path  d=\"M 57 9\nL 87 9\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><circle cx=\"72\" cy=\"9\" r=\"5\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><text x=\"107\" y=\"15\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Two Word</text><path  d=\"M 179 9\nL 209 9\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><circle cx=\"194\" cy=\"9\" r=\"5\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><text x=\"229\" y=\"15\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Three Word Item</text><path  d=\"M 346 9\nL 376 9\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><circle cx=\"361\" cy=\"9\" r=\"5\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><text x=\"396\" y=\"15\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Four Words Is Longer</text><path  d=\"M 545 9\nL 575 9\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/><circle cx=\"560\" cy=\"9\" r=\"5\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/></svg>",
		},
		{
			name: "vertical_right_alignment",
			render: func(p *Painter) ([]byte, error) {
				_, err := NewLegendPainter(p, LegendOption{
					Data:     []string{"One", "Two Word", "Three Word Item", "Four Words Is Longer"},
					Vertical: true,
					Align:    AlignRight,
				}).Render()
				if err != nil {
					return nil, err
				}
				return p.Bytes()
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><text x=\"540\" y=\"15\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">One</text><path  d=\"M 570 9\nL 600 9\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><circle cx=\"585\" cy=\"9\" r=\"5\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><text x=\"498\" y=\"35\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Two Word</text><path  d=\"M 570 29\nL 600 29\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><circle cx=\"585\" cy=\"29\" r=\"5\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><text x=\"453\" y=\"55\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Three Word Item</text><path  d=\"M 570 49\nL 600 49\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><circle cx=\"585\" cy=\"49\" r=\"5\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><text x=\"421\" y=\"75\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Four Words Is Longer</text><path  d=\"M 570 69\nL 600 69\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/><circle cx=\"585\" cy=\"69\" r=\"5\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/></svg>",
		},
		{
			name: "vertical_right_alignment_left_position",
			render: func(p *Painter) ([]byte, error) {
				_, err := NewLegendPainter(p, LegendOption{
					Data:     []string{"One", "Two Word", "Three Word Item", "Four Words Is Longer"},
					Vertical: true,
					Offset:   OffsetLeft,
					Align:    AlignRight,
				}).Render()
				if err != nil {
					return nil, err
				}
				return p.Bytes()
			},
			result: "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 600 400\"><text x=\"119\" y=\"15\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">One</text><path  d=\"M 149 9\nL 179 9\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><circle cx=\"164\" cy=\"9\" r=\"5\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(84,112,198,1.0);fill:rgba(84,112,198,1.0)\"/><text x=\"77\" y=\"35\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Two Word</text><path  d=\"M 149 29\nL 179 29\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><circle cx=\"164\" cy=\"29\" r=\"5\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(145,204,117,1.0);fill:rgba(145,204,117,1.0)\"/><text x=\"32\" y=\"55\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Three Word Item</text><path  d=\"M 149 49\nL 179 49\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><circle cx=\"164\" cy=\"49\" r=\"5\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(250,200,88,1.0);fill:rgba(250,200,88,1.0)\"/><text x=\"0\" y=\"75\" style=\"stroke-width:0;stroke:none;fill:rgba(70,70,70,1.0);font-size:15.3px;font-family:'Roboto Medium',sans-serif\">Four Words Is Longer</text><path  d=\"M 149 69\nL 179 69\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/><circle cx=\"164\" cy=\"69\" r=\"5\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/><path  d=\"\" style=\"stroke-width:3;stroke:rgba(238,102,102,1.0);fill:rgba(238,102,102,1.0)\"/></svg>",
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i)+"-"+tt.name, func(t *testing.T) {
			p, err := NewPainter(PainterOptions{
				OutputFormat: ChartOutputSVG,
				Width:        600,
				Height:       400,
			}, PainterThemeOption(GetTheme(ThemeLight)))
			require.NoError(t, err)
			data, err := tt.render(p)
			require.NoError(t, err)
			assertEqualSVG(t, tt.result, string(data))
		})
	}
}
