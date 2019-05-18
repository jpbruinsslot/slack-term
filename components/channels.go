package components

import (
	"fmt"
	"html"

	"github.com/erroneousboat/termui"
	"github.com/renstrom/fuzzysearch/fuzzy"
)

const (
	IconOnline       = "●"
	IconOffline      = "○"
	IconChannel      = "#"
	IconGroup        = "☰"
	IconIM           = "●"
	IconMpIM         = "☰"
	IconNotification = "*"

	PresenceAway   = "away"
	PresenceActive = "active"

	ChannelTypeChannel = "channel"
	ChannelTypeGroup   = "group"
	ChannelTypeIM      = "im"
	ChannelTypeMpIM    = "mpim"
)

type ChannelItem struct {
	ID           string
	Name         string
	Topic        string
	Type         string
	UserID       string
	Presence     string
	Notification bool

	StylePrefix string
	StyleIcon   string
	StyleText   string
}

// ToString will set the label of the channel, how it will be
// displayed on screen. Based on the type, different icons are
// shown, as well as an optional notification icon.
func (c ChannelItem) ToString() string {
	var prefix string
	if c.Notification {
		prefix = IconNotification
	} else {
		prefix = " "
	}

	var icon string
	switch c.Type {
	case ChannelTypeChannel:
		icon = IconChannel
	case ChannelTypeGroup:
		icon = IconGroup
	case ChannelTypeMpIM:
		icon = IconMpIM
	case ChannelTypeIM:
		switch c.Presence {
		case PresenceActive:
			icon = IconOnline
		case PresenceAway:
			icon = IconOffline
		default:
			icon = IconIM
		}
	}

	label := fmt.Sprintf(
		"[%s](%s) [%s](%s) [%s](%s)",
		prefix, c.StylePrefix,
		icon, c.StyleIcon,
		c.Name, c.StyleText,
	)

	return label
}

// GetChannelName will return a formatted representation of the
// name of the channel
func (c ChannelItem) GetChannelName() string {
	var channelName string
	if c.Topic != "" {
		channelName = fmt.Sprintf("%s - %s",
			html.UnescapeString(c.Name),
			html.UnescapeString(c.Topic),
		)
	} else {
		channelName = c.Name
	}
	return channelName
}

// Channels is the definition of a Channels component
type Channels struct {
	ChannelItems    []ChannelItem
	List            *termui.List
	SelectedChannel int // index of which channel is selected from the List
	Offset          int // from what offset are channels rendered
	CursorPosition  int // the y position of the 'cursor'

	SearchMatches  []int // index of the search matches
	SearchPosition int   // current position of a search match
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

	for i, item := range c.ChannelItems[c.Offset:] {

		y := c.List.InnerBounds().Min.Y + i

		if y > c.List.InnerBounds().Max.Y-1 {
			break
		}

		// Set the visible cursor
		var cells []termui.Cell
		if y == c.CursorPosition {
			cells = termui.DefaultTxBuilder.Build(
				item.ToString(), c.List.ItemBgColor, c.List.ItemFgColor)
		} else {
			cells = termui.DefaultTxBuilder.Build(
				item.ToString(), c.List.ItemFgColor, c.List.ItemBgColor)
		}

		// Append ellipsis when overflows
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

func (c *Channels) SetChannels(channels []ChannelItem) {
	c.ChannelItems = channels
}

func (c *Channels) MarkAsRead(channelID int) {
	c.ChannelItems[channelID].Notification = false
}

func (c *Channels) MarkAsUnread(channelID string) {
	index := c.FindChannel(channelID)
	c.ChannelItems[index].Notification = true
}

func (c *Channels) SetPresence(channelID string, presence string) {
	index := c.FindChannel(channelID)
	c.ChannelItems[index].Presence = presence
}

func (c *Channels) FindChannel(channelID string) int {
	var index int
	for i, channel := range c.ChannelItems {
		if channel.ID == channelID {
			index = i
			break
		}
	}
	return index
}

// SetSelectedChannel sets the SelectedChannel given the index
func (c *Channels) SetSelectedChannel(index int) {
	c.SelectedChannel = index
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
	if c.SelectedChannel < len(c.ChannelItems)-1 {
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
	c.SetSelectedChannel(len(c.ChannelItems) - 1)

	offset := len(c.ChannelItems) - (c.List.InnerBounds().Max.Y - 1)

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
		if c.Offset < len(c.ChannelItems)-1 {
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
	c.SearchMatches = make([]int, 0)

	targets := make([]string, 0)
	for _, c := range c.ChannelItems {
		targets = append(targets, c.Name)
	}

	matches := fuzzy.Find(term, targets)

	for _, m := range matches {
		for i, item := range c.ChannelItems {
			if m == item.Name {
				c.SearchMatches = append(c.SearchMatches, i)
				break
			}
		}
	}

	if len(c.SearchMatches) > 0 {
		c.GotoPositionSearch(0)
		c.SearchPosition = 0
	}
}

// GotoPosition is used by to automatically scroll to a specific
// location in the channels component
func (c *Channels) GotoPosition(newPos int) {

	// Is the new position in range of the current view?
	minRange := c.Offset
	maxRange := c.Offset + (c.List.InnerBounds().Max.Y - 2)

	if newPos < minRange {
		// newPos is above, we need to scroll up.
		c.SetSelectedChannel(newPos)

		// How much do we need to scroll to get it into range?
		c.Offset = c.Offset - (minRange - newPos)
	} else if newPos > maxRange {
		// newPos is below, we need to scroll down
		c.SetSelectedChannel(newPos)

		// How much do we need to scroll to get it into range?
		c.Offset = c.Offset + (newPos - maxRange)
	} else {
		// newPos is inside range
		c.SetSelectedChannel(newPos)
	}

	// Set cursor to correct position
	c.CursorPosition = (newPos - c.Offset) + 1
}

// GotoPosition is used by the search functionality to automatically
// scroll to a specific location in the channels component
func (c *Channels) GotoPositionSearch(position int) {
	newPos := c.SearchMatches[position]
	c.GotoPosition(newPos)
}

// SearchNext allows us to cycle through the c.SearchMatches
func (c *Channels) SearchNext() {
	newPosition := c.SearchPosition + 1
	if newPosition <= len(c.SearchMatches)-1 {
		c.GotoPositionSearch(newPosition)
		c.SearchPosition = newPosition
	}
}

// SearchPrev allows us to cycle through the c.SearchMatches
func (c *Channels) SearchPrev() {
	newPosition := c.SearchPosition - 1
	if newPosition >= 0 {
		c.GotoPositionSearch(newPosition)
		c.SearchPosition = newPosition
	}
}

// Jump to the first channel with a notification
func (c *Channels) Jump() {
	for i, channel := range c.ChannelItems {
		if channel.Notification {
			c.GotoPosition(i)
			break
		}
	}
}
