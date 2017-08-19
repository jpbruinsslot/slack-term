package components

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

type ChannelsNew struct {
	Component
	Items           []string
	SelectedChannel int // index of which channel is selected from the Items
	Offset          int // from what offset are channels rendered
	CursorPosition  int // the y position of the 'cursor'
}

func CreateChannelsComponent(x, y, w, h int) *ChannelsNew {

	channels := &ChannelsNew{}
	channels.Name = "channels"
	channels.Width = 10
	channels.Height = h

	return channels
}

// ... and implements the gocui.Manager interface
func (c *ChannelsNew) Layout(g *gocui.Gui) error {
	if v, err := g.SetView(c.Name, c.X, c.Y, c.X+c.Width, c.Y+c.Height); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

		for _, item := range c.Items {
			fmt.Fprintln(v, item)
		}

	}
	return nil
}
