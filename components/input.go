package components

import (
	"github.com/erroneousboat/gocui"
)

type Input struct {
	Component
	// Text []rune
}

func CreateInputComponent(x, y, w, h int) *Input {
	input := &Input{}

	input.Name = "input"
	input.Y = y
	input.X = x
	input.Width = w
	input.Height = h

	return input
}

// Layout will setup the visible part of the Channels component and implements
// the gocui.Manager interface
func (i *Input) Layout(g *gocui.Gui) error {
	if v, err := g.SetView(i.Name, i.X, i.Y, i.X+i.Width, i.Y+i.Height); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Editable = true

		i.View = v

	}
	return nil
}
