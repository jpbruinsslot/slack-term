package components

import "github.com/erroneousboat/termui"

type Debug struct {
	Par  *termui.Par
	List *termui.List
}

func CreateDebugComponent(inputHeight int) *Debug {
	debug := &Debug{
		List: termui.NewList(),
	}

	debug.List.BorderLabel = "Debug"
	debug.List.Height = termui.TermHeight() - inputHeight

	return debug
}

// Buffer implements interface termui.Bufferer
func (d *Debug) Buffer() termui.Buffer {
	return d.List.Buffer()
}

// GetHeight implements interface termui.GridBufferer
func (d *Debug) GetHeight() int {
	return d.List.Block.GetHeight()
}

// SetWidth implements interface termui.GridBufferer
func (d *Debug) SetWidth(w int) {
	d.List.SetWidth(w)
}

// SetX implements interface termui.GridBufferer
func (d *Debug) SetX(x int) {
	d.List.SetX(x)
}

// SetY implements interface termui.GridBufferer
func (d *Debug) SetY(y int) {
	d.List.SetY(y)
}

// Println will add the text to the Debug component
func (d *Debug) Println(text string) {
	d.List.Items = append(d.List.Items, text)
	termui.Render(d)
}
