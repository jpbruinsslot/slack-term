package components

import "github.com/jroimartin/gocui"

type Debug struct {
	Component
	Text string
}

// ... and implements the gocui.Manager interface
func (d *Debug) Layout(g *gocui.Gui) error {
	return nil
}
