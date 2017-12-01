package components

import (
	"fmt"

	"github.com/erroneousboat/gocui"
)

type Chat struct {
	Component
	Items []string
}

// Constructor for the Chat component
func CreateChatComponent(x, y, w, h int) *Chat {
	chat := &Chat{}

	chat.Name = "chat"
	chat.Y = y
	chat.X = x
	chat.Width = w
	chat.Height = h

	return chat
}

// Layout will setup the visible part of the Chat component and implements
// the gocui.Manager interface
func (c *Chat) Layout(g *gocui.Gui) error {

	if v, err := g.SetView(c.Name, c.X, c.Y, c.X+c.Width, c.Y+c.Height); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Wrap = true
		v.Autoscroll = true // FIXME: see if you need this

		for _, msg := range c.Items {
			fmt.Fprintln(v, msg)
		}

		c.View = v
	}

	return nil
}

// FIXME: maybe not necessary
func (c *Chat) Refresh() {
	for _, msg := range c.Items {
		fmt.Fprintln(c.View, msg)
	}
}

// SetMessage will put the provided message into the the Items field
// of the Chat view
func (c *Chat) SetMessages(messages []string) {
	for _, msg := range messages {
		c.Items = append(c.Items, msg)
	}
}

// ClearMessages clear the c.Items
func (c *Chat) ClearMessages() {
	c.Items = []string{}
	c.View.Clear()
}
