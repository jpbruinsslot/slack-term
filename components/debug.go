package components

import (
	"fmt"

	"github.com/erroneousboat/termui"
)

// Debug can be used to relay debugging information in the Debug component,
// see event.go on how to use it
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
	debug.List.Overflow = "wrap"

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

	// When at the end remove first item
	if len(d.List.Items) > d.List.InnerBounds().Max.Y-1 {
		d.List.Items = d.List.Items[1:]
	}

	termui.Render(d)
}

func (d *Debug) Sprintf(format string, a ...interface{}) {
	text := fmt.Sprintf(format, a...)
	d.List.Items = append(d.List.Items, text)

	// When at the end remove first item
	if len(d.List.Items) > d.List.InnerBounds().Max.Y-1 {
		d.List.Items = d.List.Items[1:]
	}

	termui.Render(d)
}
