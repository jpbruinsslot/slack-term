package components

import (
	"fmt"
	"strings"

	"github.com/erroneousboat/slack-term/src/service"
	"github.com/gizak/termui"
)

type Channels struct {
	List            *termui.List
	SelectedChannel int
}

type SlackChannel struct {
	Name string
	ID   string
}

func CreateChannels(svc *service.SlackService, inputHeight int) *Channels {
	channels := &Channels{
		List: termui.NewList(),
	}

	channels.List.BorderLabel = "Channels"
	channels.List.Height = termui.TermHeight() - inputHeight

	channels.SelectedChannel = 0

	channels.GetChannels(svc)

	return channels
}

// Buffer implements interface termui.Bufferer
func (c *Channels) Buffer() termui.Buffer {
	buf := c.List.Buffer()

	for y, item := range c.List.Items {
		var cells []termui.Cell
		if y == c.SelectedChannel {
			cells = termui.DefaultTxBuilder.Build(
				item, termui.ColorBlack, termui.ColorWhite)
		} else {
			cells = termui.DefaultTxBuilder.Build(
				item, c.List.ItemFgColor, c.List.ItemBgColor)
		}

		cells = termui.DTrimTxCls(cells, c.List.InnerWidth())

		x := 0
		for _, cell := range cells {
			width := cell.Width()
			buf.Set(
				c.List.InnerBounds().Min.X+x,
				c.List.InnerBounds().Min.Y+y,
				cell,
			)
			x += width
		}
	}

	return buf
}

// GetHeight implements interface termui.GridBufferer
func (c *Channels) GetHeight() int {
	return c.List.Block.GetHeight()
}

// SetWidth implements interface termui.GridBufferer
func (c *Channels) SetWidth(w int) {
	c.List.SetWidth(w)
}

// SetX implements interface termui.GridBufferer
func (c *Channels) SetX(x int) {
	c.List.SetX(x)
}

// SetY implements interface termui.GridBufferer
func (c *Channels) SetY(y int) {
	c.List.SetY(y)
}

// GetChannels will get all available channels from the SlackService
// and add them to the List as well as to the SlackChannels, this is done
// to better relate the ID and name given to Channels, for Chat.GetMessages.
// See event.go actionChangeChannel for more explanation
func (c *Channels) GetChannels(svc *service.SlackService) {
	for _, slackChan := range svc.GetChannels() {
		c.List.Items = append(c.List.Items, fmt.Sprintf("  %s", slackChan.Name))
	}
}

// SetSelectedChannel sets the SelectedChannel given the index
func (c *Channels) SetSelectedChannel(index int) {
	c.SelectedChannel = index
}

// MoveCursorUp will decrease the SelectedChannel by 1
func (c *Channels) MoveCursorUp() {
	if c.SelectedChannel > 0 {
		c.SetSelectedChannel(c.SelectedChannel - 1)
		c.ClearNewMessageIndicator()
	}
}

// MoveCursorDown will increase the SelectedChannel by 1
func (c *Channels) MoveCursorDown() {
	if c.SelectedChannel < len(c.List.Items)-1 {
		c.SetSelectedChannel(c.SelectedChannel + 1)
		c.ClearNewMessageIndicator()
	}
}

// NewMessage will be called when a new message arrives and will
// render an asterisk in front of the channel name
func (c *Channels) NewMessage(svc *service.SlackService, channelID string) {
	var index int

	// Get the correct Channel from SlackChannels
	for i, channel := range svc.Channels {
		if channelID == channel.ID {
			index = i
			break
		}
	}

	if !strings.Contains(c.List.Items[index], "*") {
		// The order of SlackChannels relates to the order of
		// List.Items, index will be the index of the channel
		c.List.Items[index] = fmt.Sprintf("* %s", strings.TrimSpace(c.List.Items[index]))
	}

	// Play terminal bell sound
	fmt.Print("\a")
}

// ClearNewMessageIndicator will remove the asterisk in front of a channel that
// received a new message. This will happen as one will move up or down the
// cursor for Channels
func (c *Channels) ClearNewMessageIndicator() {
	channelName := strings.Split(c.List.Items[c.SelectedChannel], "* ")
	if len(channelName) > 1 {
		c.List.Items[c.SelectedChannel] = fmt.Sprintf("  %s", channelName[1])
	} else {
		c.List.Items[c.SelectedChannel] = channelName[0]
	}
}
