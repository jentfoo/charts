package drawing

// DemuxFlattener is a slice of Flattener.
type DemuxFlattener struct {
	Flatteners []Flattener
}

// MoveTo implements the path builder interface.
func (dc DemuxFlattener) MoveTo(x, y float64) {
	for _, flattener := range dc.Flatteners {
		flattener.MoveTo(x, y)
	}
}

// LineTo implements the path builder interface.
func (dc DemuxFlattener) LineTo(x, y float64) {
	for _, flattener := range dc.Flatteners {
		flattener.LineTo(x, y)
	}
}

// End implements the path builder interface.
func (dc DemuxFlattener) End() {
	for _, flattener := range dc.Flatteners {
		flattener.End()
	}
}
