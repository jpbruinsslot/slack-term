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
	ID             string
	Name           string
	Topic          string
	Type           string
	UserID         string
	Presence       string
	Notification   bool
	StylePrefix    string
	StyleIcon      string
	StyleText      string
	IsSearchResult bool
}

// ToString will set the label of the channel, how it will be
// displayed on screen. Based on the type, different icons are
// shown, as well as an optional notification icon.
func (c *ChannelItem) ToString() string {
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
func (c *ChannelItem) GetChannelName() string {
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
	ChannelItems    map[string]*ChannelItem // mapping of channel IDs to ChannelItems
	CursorPosition  int                     // the y position of the cursor in this component
	List            *termui.List            // ui of visible channels
	Offset          int                     // offset in channel list from which items are render
	SearchPosition  int                     // current position in search results
	SelectedChannel string                  // index of which channel is selected from the List
	UnreadOnly      bool                    // only show unread messages when on
	itemsRendered   int                     // the number of items currently rendered on screen
	channelIDs      []string                // sorted list of channel IDs; the nth sortedChannels has the nth alphabetically significant ChannelItem.Name
}

// CreateChannels is the constructor for the Channels component
func CreateChannelsComponent(height int, unreadOnly bool) *Channels {
	channels := &Channels{
		List: termui.NewList(),
	}

	channels.List.BorderLabel = "Conversations"
	channels.List.Height = height

	channels.SelectedChannel = ""
	channels.CursorPosition = channels.minY()
	channels.Offset = 0
	channels.UnreadOnly = unreadOnly
	channels.ChannelItems = make(map[string]*ChannelItem)

	return channels
}

// ListChannels lists all visible channels
// if the application has UnreadOnly disabled, list all channels
// if the application has UnreadOnly enabled
//  - always includes the currently selected channel (if one is selected)
//  - always show channels that match the current search
//  - never show channels that are read and match neither of the above conditions
func (c *Channels) ListChannels() (items []*ChannelItem) {

	if !c.UnreadOnly {
		for _, id := range c.channelIDs {
			items = append(items, c.ChannelItems[id])
		}

		return
	}

	// filter by messages that are either unread or search results ( always show selected channel)
	for _, id := range c.channelIDs {

		chn := c.ChannelItems[id]
		if c.shouldBeVisible(chn) {
			items = append(items, chn)
		}
	}

	return
}

// Buffer implements interface termui.Bufferer
func (c *Channels) Buffer() termui.Buffer {
	buf := c.List.Buffer()

	c.itemsRendered = 0

	for i, item := range c.ListChannels()[c.Offset:] {

		y := c.minY() + i

		c.itemsRendered++

		if y > c.maxY()-1 {
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

		x := c.List.InnerBounds().Min.X
		for _, cell := range cells {
			buf.Set(x, y, cell)
			x += cell.Width()
		}

		// When not at the end of the pane fill it up empty characters
		for x < c.List.InnerBounds().Max.X {
			if y == c.CursorPosition {
				buf.Set(x, y,
					termui.Cell{
						Ch: ' ',
						Fg: c.List.ItemBgColor,
						Bg: c.List.ItemFgColor,
					},
				)
			} else {
				buf.Set(
					x, y,
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

// SetChannels sets the channel list for this component.
// `channels` is assumed to be sorted
func (c *Channels) SetChannels(channels []*ChannelItem) {

	c.channelIDs = make([]string, len(channels))

	for i, chn := range channels {
		c.channelIDs[i] = chn.ID
		c.ChannelItems[chn.ID] = chn
	}

	// select the first channel by default when UnreadOnly is disabled
	if len(channels) > 0 && !c.UnreadOnly {
		c.SetSelectedChannel(c.ChannelItems[c.channelIDs[0]].ID)
	}
}

// MarkAsRead marks a channel as read
func (c *Channels) MarkAsRead(channelID string) {

	if chn, ok := c.ChannelItems[channelID]; ok {
		chn.Notification = false
	}
}

// MarkAsUnread marks a channel as unread
func (c *Channels) MarkAsUnread(channelID string) {

	if chn, ok := c.ChannelItems[channelID]; ok {
		chn.Notification = true
	}

	// the cursor may need to be repositioned now that more channels
	// may be visible
	c.ScrollToChannel(c.SelectedChannel)
}

// SetPresence sets the presnce on a given channel
func (c *Channels) SetPresence(channelID string, presence string) {

	if chn, ok := c.ChannelItems[channelID]; ok {
		chn.Presence = presence
	}
}

// SetSelectedChannel sets the SelectedChannel given its ID
func (c *Channels) SetSelectedChannel(channelID string) {

	if _, ok := c.ChannelItems[channelID]; ok {
		c.SelectedChannel = channelID
	}
}

// Get SelectedChannel returns the ChannelItem that is currently selected
// return
// - selected: the selected channel item
// - ok: true if there exists a selected channel item in the channel list
func (c *Channels) GetSelectedChannel() (selected *ChannelItem, ok bool) {

	if chn, ok := c.ChannelItems[c.SelectedChannel]; ok {
		return chn, ok
	}

	return
}

// GetPreviousChannel gets the channel item that preceeds the selected channel item in the visible channel list
func (c *Channels) GetPreviousChannel() (prev *ChannelItem, ok bool) {

	var selectedChannel *ChannelItem

	// return the current channel when at the top of channel list
	if c.CursorPosition == c.minY() {
		if selectedChannel, ok = c.GetSelectedChannel(); ok {
			prev = selectedChannel
			return
		}
	}

	for _, currID := range c.channelIDs {
		if currID == c.SelectedChannel && prev.ID != "" {
			ok = true
			break
		}

		chn := c.ChannelItems[currID]
		if c.shouldBeVisible(chn) {
			prev = c.ChannelItems[currID]
		}
	}

	return
}

// GetNextChannel gets the channel item that proceeds the selected channel item in the visible channel list
func (c *Channels) GetNextChannel() (next *ChannelItem, ok bool) {

	var selectedChannel *ChannelItem

	// return the current channel when at the bottom of channel list
	if c.CursorPosition == c.maxCursorPos() {
		if selectedChannel, ok = c.GetSelectedChannel(); ok {
			next = selectedChannel
			return
		}
	}

	for i := len(c.channelIDs) - 1; i >= 0; i-- {

		currID := c.channelIDs[i]
		if currID == c.SelectedChannel && next.ID != "" {
			ok = true
			break
		}

		chn := c.ChannelItems[currID]
		if c.shouldBeVisible(chn) {
			next = c.ChannelItems[currID]
		}
	}

	return
}

// MoveCursorTop will move the cursor to the top of the channels
func (c *Channels) MoveCursorTop() {

	if list := c.ListChannels(); len(list) > 0 {
		firstChn := list[0]
		c.SetSelectedChannel(firstChn.ID)
		c.ScrollToChannel(firstChn.ID)
	}
}

// MoveCursorBottom will move the cursor to the bottom of the channels
func (c *Channels) MoveCursorBottom() {

	if list := c.ListChannels(); len(list) > 0 {
		index := len(list) - 1
		lastChn := list[index]
		c.SetSelectedChannel(lastChn.ID)
		c.ScrollToChannel(lastChn.ID)
	}
}

// ScrollToChannel repositions the Offset such that when the screen renders, the
// channel identified by channelID is visible.
func (c *Channels) ScrollToChannel(channelID string) {

	if _, index, ok := c.findVisibleChannel(channelID); ok {
		offset := index - (c.maxY() - c.minY()) + 1

		if offset < 0 {
			c.Offset = 0
			c.CursorPosition = c.minY() + index
		} else {
			c.Offset = offset
			c.CursorPosition = c.maxY() - 1
		}
	}
}

// Search will search through the channels to find a channel,
// when a match has been found the selected channel will then
// be the channel that has been found
func (c *Channels) Search(term string) (resultCount int) {

	for _, chn := range c.ChannelItems {
		if chn.IsSearchResult {
			chn.IsSearchResult = false
		}
	}

	targets := make([]string, 0)
	for _, c := range c.ChannelItems {
		targets = append(targets, c.Name)
	}

	matches := fuzzy.Find(term, targets)

	for _, m := range matches {
		for key, chn := range c.ChannelItems {
			if m == chn.Name {
				resultCount = resultCount + 1
				chn.IsSearchResult = true
				c.ChannelItems[key] = chn
				break
			}
		}
	}

	if resultCount > 0 {
		c.GotoPosition(0)
		c.SearchPosition = 0
	}

	return
}

// GotoPosition scrolls the channel list to a new position in the current search results
func (c *Channels) GotoPosition(newPos int) (ok bool) {

	// there's nothing to be done if there are no search results, or the given
	// position is out of range
	var results []string
	if results = c.listSearchResults(); len(results) == 0 ||
		newPos > len(results)-1 ||
		newPos < 0 {
		return
	}

	newChannelID := results[newPos]
	if _, ok = c.ChannelItems[newChannelID]; !ok {
		return
	}

	c.SetSelectedChannel(newChannelID)
	c.ScrollToChannel(newChannelID)

	return true
}

// SearchNext allows us to cycle through search results
func (c *Channels) SearchNext() {
	newPosition := c.SearchPosition + 1
	if ok := c.GotoPosition(newPosition); ok {
		c.SearchPosition = newPosition
	}
}

// SearchPrev allows us to cycle through search results
func (c *Channels) SearchPrev() {
	newPosition := c.SearchPosition - 1
	if ok := c.GotoPosition(newPosition); ok {
		c.SearchPosition = newPosition
	}
}

// Jump to the first channel with a notification
func (c *Channels) Jump() {

	for _, chnID := range c.channelIDs {

		if c.ChannelItems[chnID].Notification {
			c.SetSelectedChannel(chnID)
			c.ScrollToChannel(chnID)
			break
		}
	}
}

// findVisibleChannels finds a channel by ID in the visible channel list
func (c *Channels) findVisibleChannel(channelID string) (chn *ChannelItem, idx int, ok bool) {

	for i, channel := range c.ListChannels() {
		if channel.ID == channelID {
			chn = channel
			idx = i
			ok = true
			break
		}
	}

	return
}

// listSearchResults lists the IDs of the channel that are search results
func (c *Channels) listSearchResults() (channelIDs []string) {
	for _, chn := range c.ListChannels() {
		if chn.IsSearchResult {
			channelIDs = append(channelIDs, chn.ID)
		}
	}

	return
}

// maxCursorPos returns the y coordinate beyond which the cursor may not advance
func (c *Channels) maxCursorPos() int {
	return c.minY() + c.itemsRendered - 1
}

// maxY returns the largest Y value that may be rendered in this component
func (c *Channels) maxY() int {
	return c.List.InnerBounds().Max.Y
}

// minY returns the largets Y value that may be rendered in this component
func (c *Channels) minY() int {
	return c.List.InnerBounds().Min.Y
}

func (c *Channels) shouldBeVisible(chn *ChannelItem) bool {

	if !c.UnreadOnly {
		return true
	}

	return chn.ID == c.SelectedChannel || chn.Notification || chn.IsSearchResult
}
