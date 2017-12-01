package components

import (
	"fmt"
	"strings"

	"github.com/erroneousboat/gocui"

	"github.com/erroneousboat/slack-term/service"
)

const (
	IconOnline       = "●"
	IconOffline      = "○"
	IconChannel      = "#"
	IconGroup        = "☰"
	IconIM           = "●"
	IconNotification = "*"

	PresenceAway   = "away"
	PresenceActive = "active"
)

type Channels struct {
	Component
	View            *gocui.View
	Items           []string
	SelectedChannel int // index of which channel is selected from the Items
	Offset          int // from what offset are channels rendered, FIXME probably not necessary anymore
	CursorPosition  int // the y position of the 'cursor'
	// SelectorBGColor
	// SelectorFGColor
}

// Constructor for the Channels component
func CreateChannelsComponent(x, y, w, h int) *Channels {
	channels := &Channels{}

	channels.Name = "channels"
	channels.Y = y
	channels.X = x
	channels.Width = w
	channels.Height = h

	return channels
}

// Layout will setup the visible part of the Channels component and implements
// the gocui.Manager interface
func (c *Channels) Layout(g *gocui.Gui) error {
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

		c.View = v

	}
	return nil
}

// SetChannels will set the channels from the service, passed as an argument
// to the Items field
// FIXME: maybe rename to LoadChannels?
func (c *Channels) SetChannels(channels []service.Channel) {
	for _, slackChan := range channels {
		label := setChannelLabel(slackChan, false)
		c.Items = append(c.Items, label)
	}
}

// SetPresenceChannels will set the icon for all the IM channels
func (c *Channels) SetPresenceChannels(channels []service.Channel) {
	for _, slackChan := range channels {
		if slackChan.Type == service.ChannelTypeIM {
			c.SetPresenceChannel(channels, slackChan.UserID, slackChan.Presence)
		}
	}
}

// SetPresence will set the correct icon for one IM channel
func (c *Channels) SetPresenceChannel(channels []service.Channel, userID string, presence string) {
	// Get the correct Channel from svc.Channels
	var index int
	for i, channel := range channels {
		if userID == channel.UserID {
			index = i
			break
		}
	}

	switch presence {
	case PresenceActive:
		c.Items[index] = strings.Replace(
			c.Items[index], IconOffline, IconOnline, 1,
		)
	case PresenceAway:
		c.Items[index] = strings.Replace(
			c.Items[index], IconOnline, IconOffline, 1,
		)
	default:
		c.Items[index] = strings.Replace(
			c.Items[index], IconOnline, IconOffline, 1,
		)
	}
}

// TODO: documentation
func (c *Channels) SetSelectedChannel(index int) {
	c.SelectedChannel = index
}

// TODO: documentation
func (c *Channels) GetSelectedChannel() string {
	return c.Items[c.SelectedChannel]
}

// MoveCursorUp will decrease the SelectedChannel by 1
func (c *Channels) MoveCursorUp() error {
	if c.SelectedChannel > 0 {
		c.SetSelectedChannel(c.SelectedChannel - 1)
		c.ScrollUp()
		c.MarkAsRead()
	}
	return nil
}

// MoveCursorDown will increase the SelectedChannel by 1
func (c *Channels) MoveCursorDown() error {
	if c.SelectedChannel < len(c.Items)-1 {
		c.SetSelectedChannel(c.SelectedChannel + 1)
		c.ScrollDown()
		c.MarkAsRead()
	}
	return nil
}

// MoveCursorTop will move the cursor to the top of the channels
// func (c *Channels) MoveCursorTop() {
// 	c.SetSelectedChannel(0)
// 	c.CursorPosition = c.List.InnerBounds().Min.Y // FIXME
// 	c.Offset = 0
// }

// MoveCursorBottom will move the cursor to the bottom of the channels
// func (c *Channels) MoveCursorBottom() {
// 	c.SetSelectedChannel(len(c.Items) - 1)
//
// 	offset := len(c.List.Items) - (c.List.InnerBounds().Max.Y - 1) // FIXME
//
// 	if offset < 0 {
// 		c.Offset = 0
// 		c.CursorPosition = c.SelectedChannel + 1
// 	} else {
// 		c.Offset = offset
// 		c.CursorPosition = c.List.InnerBounds().Max.Y - 1 // FIXME
// 	}
// }

// ScrollUp enables us to scroll through the channel list when it overflows
func (c *Channels) ScrollUp() {
	originX, originY := c.View.Origin()
	cursorX, cursorY := c.View.Cursor()

	// When cursor is at the beginning of the view then decrease
	// the origin of the view
	if cursorY-1 < 0 {
		c.View.SetOrigin(originX, originY-1)
	}

	c.View.SetCursor(cursorX, cursorY-1)
}

// ScrollDown enables us to scroll through the channel list when it overflows
func (c *Channels) ScrollDown() {
	originX, originY := c.View.Origin()
	cursorX, cursorY := c.View.Cursor()

	// When cursor is at the end of the view then increase
	// the origin of the view
	if cursorY+1 > c.Height-2 {
		c.View.SetOrigin(originX, originY+1)
	}

	c.View.SetCursor(cursorX, cursorY+1)
}

// Search will search through the channels to find a channel,
// when a match has been found the selected channel will then
// be the channel that has been found
// func (c *Channels) Search(term string) {
// 	for i, item := range c.Items {
// 		if strings.Contains(item, term) {
//
// 			// The new position
// 			newPos := i
//
// 			// Is the new position in range of the current view?
// 			minRange := c.Offset
// 			maxRange := c.Offset + (c.List.InnerBounds().Max.Y - 2) // FIXME
//
// 			if newPos < minRange {
// 				// newPos is above, we need to scroll up.
// 				c.SetSelectedChannel(i)
//
// 				// How much do we need to scroll to get it into range?
// 				c.Offset = c.Offset - (minRange - newPos)
// 			} else if newPos > maxRange {
// 				// newPos is below, we need to scroll down
// 				c.SetSelectedChannel(i)
//
// 				// How much do we need to scroll to get it into range?
// 				c.Offset = c.Offset + (newPos - maxRange)
// 			} else {
// 				// newPos is inside range
// 				c.SetSelectedChannel(i)
// 			}
//
// 			// Set cursor to correct position
// 			c.CursorPosition = (newPos - c.Offset) + 1
//
// 			break
// 		}
// 	}
// }

// MarkAsUnread will be called when a new message arrives and will
// render an notification icon in front of the channel name
func (c *Channels) MarkAsUnRead(channels []service.Channel, channelID string) {
	// Get the correct Channel from svc.Channels
	var index int
	for i, channel := range channels {
		if channelID == channel.ID {
			index = i
			break
		}
	}

	if !strings.Contains(c.Items[index], IconNotification) {
		// The order of svc.Channels relates to the order of
		// List.Items, index will be the index of the channel
		c.Items[index] = fmt.Sprintf(
			"%s %s", IconNotification, strings.TrimSpace(c.Items[index]),
		)
	}

	// Play terminal bell sound
	fmt.Print("\a")
}

// MarkAsRead will remove the notification icon in front of
// a channel that received a new message. This will happen as one will
// move up or down the cursor for Channels
func (c *Channels) MarkAsRead() {
	channelName := strings.Split(
		c.Items[c.SelectedChannel],
		fmt.Sprintf("%s ", IconNotification),
	)

	if len(channelName) > 1 {
		c.Items[c.SelectedChannel] = fmt.Sprintf("  %s", channelName[1])
	} else {
		c.Items[c.SelectedChannel] = channelName[0]
	}
}

// setChannelLabel will set the label of the channel, meaning, how it
// is displayed on screen. Based on the type, different icons are
// shown, as well as an optional notification icon.
// func setChannelLabel(channel service.Channel, notification bool) string {
// 	var prefix string
// 	if notification {
// 		prefix = IconNotification
// 	} else {
// 		prefix = " "
// 	}
//
// 	var label string
// 	switch channel.Type {
// 	case service.ChannelTypeChannel:
// 		label = fmt.Sprintf("%s %s %s", prefix, IconChannel, channel.Name)
// 	case service.ChannelTypeGroup:
// 		label = fmt.Sprintf("%s %s %s", prefix, IconGroup, channel.Name)
// 	case service.ChannelTypeIM:
// 		label = fmt.Sprintf("%s %s %s", prefix, IconIM, channel.Name)
// 	}
//
// 	return label
// }
