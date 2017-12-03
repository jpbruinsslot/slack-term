package components

import (
	"strings"

	"github.com/erroneousboat/termui"
)

const (
	IconOnline       = "●"
	IconOffline      = "○"
	IconChannel      = "#"
	IconGroup        = "☰"
	IconIM           = "●"
	IconNotification = "*"
)

// Channels is the definition of a Channels component
type Channels struct {
	List            *termui.List
	SelectedChannel int // index of which channel is selected from the List
	Offset          int // from what offset are channels rendered
	CursorPosition  int // the y position of the 'cursor'
}

// CreateChannels is the constructor for the Channels component
func CreateChannelsComponent(inputHeight int) *Channels {
	channels := &Channels{
		List: termui.NewList(),
	}

	channels.List.BorderLabel = "Channels"
	channels.List.Height = termui.TermHeight() - inputHeight

	channels.SelectedChannel = 0
	channels.Offset = 0
	channels.CursorPosition = channels.List.InnerBounds().Min.Y

	return channels
}

// Buffer implements interface termui.Bufferer
func (c *Channels) Buffer() termui.Buffer {
	buf := c.List.Buffer()

	for i, item := range c.List.Items[c.Offset:] {

		y := c.List.InnerBounds().Min.Y + i

		if y > c.List.InnerBounds().Max.Y-1 {
			break
		}

		var cells []termui.Cell
		if y == c.CursorPosition {
			cells = termui.DefaultTxBuilder.Build(
				item, c.List.ItemBgColor, c.List.ItemFgColor)
		} else {
			cells = termui.DefaultTxBuilder.Build(
				item, c.List.ItemFgColor, c.List.ItemBgColor)
		}

		cells = termui.DTrimTxCls(cells, c.List.InnerWidth())

		x := 0
		for _, cell := range cells {
			width := cell.Width()
			buf.Set(c.List.InnerBounds().Min.X+x, y, cell)
			x += width
		}

		// When not at the end of the pane fill it up empty characters
		for x < c.List.InnerBounds().Max.X-1 {
			if y == c.CursorPosition {
				buf.Set(x+1, y,
					termui.Cell{
						Ch: ' ',
						Fg: c.List.ItemBgColor,
						Bg: c.List.ItemFgColor,
					},
				)
			} else {
				buf.Set(
					x+1, y,
					termui.Cell{
						Ch: ' ',
						Fg: c.List.ItemFgColor,
						Bg: c.List.ItemBgColor,
					},
				)
			}
			x++
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

func (c *Channels) SetChannels(channels []string) {
	c.List.Items = channels
}

// SetSelectedChannel sets the SelectedChannel given the index
func (c *Channels) SetSelectedChannel(index int) {
	c.SelectedChannel = index
}

// GetSelectedChannel returns the SelectedChannel
func (c *Channels) GetSelectedChannel() string {
	return c.List.Items[c.SelectedChannel]
}

// MoveCursorUp will decrease the SelectedChannel by 1
func (c *Channels) MoveCursorUp() {
	if c.SelectedChannel > 0 {
		c.SetSelectedChannel(c.SelectedChannel - 1)
		c.ScrollUp()
	}
}

// MoveCursorDown will increase the SelectedChannel by 1
func (c *Channels) MoveCursorDown() {
	if c.SelectedChannel < len(c.List.Items)-1 {
		c.SetSelectedChannel(c.SelectedChannel + 1)
		c.ScrollDown()
	}
}

// MoveCursorTop will move the cursor to the top of the channels
func (c *Channels) MoveCursorTop() {
	c.SetSelectedChannel(0)
	c.CursorPosition = c.List.InnerBounds().Min.Y
	c.Offset = 0
}

// MoveCursorBottom will move the cursor to the bottom of the channels
func (c *Channels) MoveCursorBottom() {
	c.SetSelectedChannel(len(c.List.Items) - 1)

	offset := len(c.List.Items) - (c.List.InnerBounds().Max.Y - 1)

	if offset < 0 {
		c.Offset = 0
		c.CursorPosition = c.SelectedChannel + 1
	} else {
		c.Offset = offset
		c.CursorPosition = c.List.InnerBounds().Max.Y - 1
	}
}

// ScrollUp enables us to scroll through the channel list when it overflows
func (c *Channels) ScrollUp() {
	// Is cursor at the top of the channel view?
	if c.CursorPosition == c.List.InnerBounds().Min.Y {
		if c.Offset > 0 {
			c.Offset--
		}
	} else {
		c.CursorPosition--
	}
}

// ScrollDown enables us to scroll through the channel list when it overflows
func (c *Channels) ScrollDown() {
	// Is the cursor at the bottom of the channel view?
	if c.CursorPosition == c.List.InnerBounds().Max.Y-1 {
		if c.Offset < len(c.List.Items)-1 {
			c.Offset++
		}
	} else {
		c.CursorPosition++
	}
}

// Search will search through the channels to find a channel,
// when a match has been found the selected channel will then
// be the channel that has been found
func (c *Channels) Search(term string) {
	for i, item := range c.List.Items {
		if strings.Contains(item, term) {

			// The new position
			newPos := i

			// Is the new position in range of the current view?
			minRange := c.Offset
			maxRange := c.Offset + (c.List.InnerBounds().Max.Y - 2)

			if newPos < minRange {
				// newPos is above, we need to scroll up.
				c.SetSelectedChannel(i)

				// How much do we need to scroll to get it into range?
				c.Offset = c.Offset - (minRange - newPos)
			} else if newPos > maxRange {
				// newPos is below, we need to scroll down
				c.SetSelectedChannel(i)

				// How much do we need to scroll to get it into range?
				c.Offset = c.Offset + (newPos - maxRange)
			} else {
				// newPos is inside range
				c.SetSelectedChannel(i)
			}

			// Set cursor to correct position
			c.CursorPosition = (newPos - c.Offset) + 1

			break
		}
	}
}
