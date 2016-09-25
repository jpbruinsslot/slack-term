package components

import "github.com/gizak/termui"

type Mode struct {
	Par *termui.Par
}

func CreateMode() *Mode {
	mode := &Mode{
		Par: termui.NewPar("NORMAL"),
	}

	mode.Par.Height = 3

	return mode
}

// Buffer implements interface termui.Bufferer
func (m *Mode) Buffer() termui.Buffer {
	buf := m.Par.Buffer()
	return buf
}

// GetHeight implements interface termui.GridBufferer
func (m *Mode) GetHeight() int {
	return m.Par.Block.GetHeight()
}

// SetWidth implements interface termui.GridBufferer
func (m *Mode) SetWidth(w int) {
	m.Par.SetWidth(w)
}

// SetX implements interface termui.GridBufferer
func (m *Mode) SetX(x int) {
	m.Par.SetX(x)
}

// SetY implements interface termui.GridBufferer
func (m *Mode) SetY(y int) {
	m.Par.SetY(y)
}
