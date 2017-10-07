package components

import (
	"fmt"

	"github.com/erroneousboat/gocui"
)

type Debug struct {
	Component
	View *gocui.View
	Text string
}

func CreateDebugComponent(x, y, w, h int) *Debug {
	debug := &Debug{}
	debug.Name = "debug"
	debug.X = x
	debug.Y = y
	debug.Width = w
	debug.Height = h

	return debug
}

// Layout will setup the visible part of the Debug component and implements
// the gocui.Manager interface
func (d *Debug) Layout(g *gocui.Gui) error {
	if v, err := g.SetView(d.Name, d.X, d.Y, d.X+d.Width, d.Y+d.Height); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Wrap = true
		v.Autoscroll = true

		fmt.Fprintln(v, d.Text)

		d.View = v

	}
	return nil
}

func (d *Debug) SetText(text string) {
	fmt.Fprintln(d.View, text)
}
