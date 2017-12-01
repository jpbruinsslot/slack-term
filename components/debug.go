package components

import "github.com/erroneousboat/termui"

type Debug struct {
	Par *termui.Par
}

func CreateDebugComponent() *Debug {
	debug := &Debug{
		Par: termui.NewPar(""),
	}

	debug.Par.Height = 3

	return debug
}

// Buffer implements interface termui.Bufferer
func (d *Debug) Buffer() termui.Buffer {
	return d.Par.Buffer()
}

// GetHeight implements interface termui.GridBufferer
func (d *Debug) GetHeight() int {
	return d.Par.Block.GetHeight()
}

// SetWidth implements interface termui.GridBufferer
func (d *Debug) SetWidth(w int) {
	d.Par.SetWidth(w)
}

// SetX implements interface termui.GridBufferer
func (d *Debug) SetX(x int) {
	d.Par.SetX(x)
}

// SetY implements interface termui.GridBufferer
func (d *Debug) SetY(y int) {
	d.Par.SetY(y)
}

// SetText will set the text of the Debug component
func (d *Debug) SetText(text string) {
	d.Par.Text = text
}
