package components

import "github.com/erroneousboat/termui"

// Mode is the definition of Mode component
type Mode struct {
	Par *termui.Par
}

// CreateMode is the constructor of the Mode struct
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

	// Center text
	space := m.Par.InnerWidth()
	word := len(m.Par.Text)

	midSpace := space / 2
	midWord := word / 2

	start := midSpace - midWord

	cells := termui.DefaultTxBuilder.Build(
		m.Par.Text, m.Par.TextFgColor, m.Par.TextBgColor)

	i, j := 0, 0
	x := m.Par.InnerBounds().Min.X
	for x < m.Par.InnerBounds().Max.X {
		if i < start {
			buf.Set(
				x, m.Par.InnerY(),
				termui.Cell{
					Ch: ' ',
					Fg: m.Par.TextFgColor,
					Bg: m.Par.TextBgColor,
				},
			)
			x++
			i++
		} else {
			if j < len(cells) {
				buf.Set(x, m.Par.InnerY(), cells[j])
				i++
				j++
			}
			x++
		}
	}

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
