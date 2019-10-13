package components

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/erroneousboat/termui"
	runewidth "github.com/mattn/go-runewidth"

	"github.com/erroneousboat/slack-term/config"
)

// Chat is the definition of a Chat component
type Chat struct {
	List           *termui.List
	Messages       map[string]Message
	Offset         int
	showMessageIDs bool
}

// CreateChatComponent is the constructor for the Chat struct
func CreateChatComponent(inputHeight int) *Chat {
	chat := &Chat{
		List:           termui.NewList(),
		Messages:       make(map[string]Message),
		Offset:         0,
		showMessageIDs: false,
	}

	chat.List.Height = termui.TermHeight() - inputHeight
	chat.List.Overflow = "wrap"

	return chat
}

// Buffer implements interface termui.Bufferer
func (c *Chat) Buffer() termui.Buffer {
	// Convert Messages into termui.Cell
	cells := c.MessagesToCells(c.Messages)

	// We will create an array of Line structs, this allows us
	// to more easily render the items in a list. We will range
	// over the cells we've created and create a Line within
	// the bounds of the Chat pane
	type Line struct {
		cells []termui.Cell
	}

	lines := []Line{}
	line := Line{}

	// When we encounter a newline or, are at the bounds of the chat view we
	// stop iterating over the cells and add the line to the line array
	x := 0
	for _, cell := range cells {

		// When we encounter a newline we add the line to the array
		if cell.Ch == '\n' {
			lines = append(lines, line)

			// Reset for new line
			line = Line{}
			x = 0
			continue
		}

		if x+cell.Width() > c.List.InnerBounds().Dx() {
			lines = append(lines, line)

			// Reset for new line
			line = Line{}
			x = 0
		}

		line.cells = append(line.cells, cell)
		x += cell.Width()
	}

	// Append the last line to the array when we didn't encounter any
	// newlines or were at the bounds of the chat view
	lines = append(lines, line)

	// We will print lines bottom up, it will loop over the lines
	// backwards and for every line it'll set the cell in that line.
	// Offset is the number which allows us to begin printing the
	// line above the last line.
	buf := c.List.Buffer()
	linesHeight := len(lines)
	paneMinY := c.List.InnerBounds().Min.Y
	paneMaxY := c.List.InnerBounds().Max.Y

	currentY := paneMaxY - 1
	for i := (linesHeight - 1) - c.Offset; i >= 0; i-- {

		if currentY < paneMinY {
			break
		}

		x := c.List.InnerBounds().Min.X
		for _, cell := range lines[i].cells {
			buf.Set(x, currentY, cell)
			x += cell.Width()
		}

		// When we're not at the end of the pane, fill it up
		// with empty characters
		for x < c.List.InnerBounds().Max.X {
			buf.Set(
				x, currentY,
				termui.Cell{
					Ch: ' ',
					Fg: c.List.ItemFgColor,
					Bg: c.List.ItemBgColor,
				},
			)
			x += runewidth.RuneWidth(' ')
		}
		currentY--
	}

	// If the space above currentY is empty we need to fill
	// it up with blank lines, otherwise the List object will
	// render the items top down, and the result will mix.
	for currentY >= paneMinY {
		x := c.List.InnerBounds().Min.X
		for x < c.List.InnerBounds().Max.X {
			buf.Set(
				x, currentY,
				termui.Cell{
					Ch: ' ',
					Fg: c.List.ItemFgColor,
					Bg: c.List.ItemBgColor,
				},
			)
			x += runewidth.RuneWidth(' ')
		}
		currentY--
	}

	return buf
}

// GetHeight implements interface termui.GridBufferer
func (c *Chat) GetHeight() int {
	return c.List.Block.GetHeight()
}

// SetWidth implements interface termui.GridBufferer
func (c *Chat) SetWidth(w int) {
	c.List.SetWidth(w)
}

// SetX implements interface termui.GridBufferer
func (c *Chat) SetX(x int) {
	c.List.SetX(x)
}

// SetY implements interface termui.GridBufferer
func (c *Chat) SetY(y int) {
	c.List.SetY(y)
}

// GetMaxItems return the maximal amount of items can fit in the Chat
// component
func (c *Chat) GetMaxItems() int {
	return c.List.InnerBounds().Max.Y - c.List.InnerBounds().Min.Y
}

// SetMessages will put the provided messages into the Messages field of the
// Chat view
func (c *Chat) SetMessages(messages []Message) {
	// Reset offset first, when scrolling in view and changing channels we
	// want the offset to be 0 when loading new messages
	c.Offset = 0
	for _, msg := range messages {
		c.Messages[msg.ID] = msg
	}
}

// AddMessage adds a single message to Messages
func (c *Chat) AddMessage(message Message) {
	c.Messages[message.ID] = message
}

// AddReply adds a single reply to a parent thread, it also sets
// the thread separator
func (c *Chat) AddReply(parentID string, message Message) {
	// It is possible that a message is received but the parent is not
	// present in the chat view
	if _, ok := c.Messages[parentID]; ok {
		message.Thread = "  "
		c.Messages[parentID].Messages[message.ID] = message
	} else {
		c.AddMessage(message)
	}
}

// ClearMessages clear the c.Messages
func (c *Chat) ClearMessages() {
	c.Messages = make(map[string]Message)
}

// ToggleMessageIDs toggles the visibility of message IDs on and off in chat view
func (c *Chat) ToggleMessageIDs() {
	if c.showMessageIDs {
		c.showMessageIDs = false
	} else {
		c.showMessageIDs = true
	}
}

// ScrollUp will render the chat messages based on the Offset of the Chat
// pane.
//
// Offset is 0 when scrolled down. (we loop backwards over the array, so we
// start with rendering last item in the list at the maximum y of the Chat
// pane). Increasing the Offset will thus result in substracting the offset
// from the len(Chat.Messages).
func (c *Chat) ScrollUp() {
	c.Offset = c.Offset + 10

	// Protect overscrolling
	if c.Offset > len(c.Messages) {
		c.Offset = len(c.Messages)
	}
}

// ScrollDown will render the chat messages based on the Offset of the Chat
// pane.
//
// Offset is 0 when scrolled down. (we loop backwards over the array, so we
// start with rendering last item in the list at the maximum y of the Chat
// pane). Increasing the Offset will thus result in substracting the offset
// from the len(Chat.Messages).
func (c *Chat) ScrollDown() {
	c.Offset = c.Offset - 10

	// Protect overscrolling
	if c.Offset < 0 {
		c.Offset = 0
	}
}

// SetBorderLabel will set Label of the Chat pane to the specified string
func (c *Chat) SetBorderLabel(channelName string) {
	c.List.BorderLabel = channelName
}

// MessagesToCells is a wrapper around MessageToCells to use for a slice of
// of type Message
func (c *Chat) MessagesToCells(msgs map[string]Message) []termui.Cell {
	cells := make([]termui.Cell, 0)
	sortedMessages := SortMessages(msgs)

	for i, msg := range sortedMessages {
		cells = append(cells, c.MessageToCells(msg)...)

		if len(msg.Messages) > 0 {
			cells = append(cells, termui.Cell{Ch: '\n'})
			cells = append(cells, c.MessagesToCells(msg.Messages)...)
		}

		// Add a newline after every message
		if i < len(sortedMessages)-1 {
			cells = append(cells, termui.Cell{Ch: '\n'})
		}
	}

	return cells
}

// MessageToCells will convert a Message struct to termui.Cell
//
// We're building parts of the message individually, or else DefaultTxBuilder
// will interpret potential markdown usage in a message as well.
func (c *Chat) MessageToCells(msg Message) []termui.Cell {
	cells := make([]termui.Cell, 0)

	// When msg.Time and msg.Name are empty (in the case of attachments)
	// don't add the time and name parts.
	if (msg.Time != time.Time{} && msg.Name != "") {
		// Time
		cells = append(cells, termui.DefaultTxBuilder.Build(
			msg.GetTime(),
			termui.ColorDefault, termui.ColorDefault)...,
		)

		if c.showMessageIDs && msg.ThreadID == "" {
			cells = append(cells, termui.DefaultTxBuilder.Build(
				msg.GetMsgID(),
				termui.ColorDefault, termui.ColorDefault)...,
			)
		}

		// Thread
		cells = append(cells, termui.DefaultTxBuilder.Build(
			msg.GetThread(),
			termui.ColorDefault, termui.ColorDefault)...,
		)

		// Name
		cells = append(cells, termui.DefaultTxBuilder.Build(
			msg.GetName(),
			termui.ColorDefault, termui.ColorDefault)...,
		)
	}

	// Hack, in order to get the correct fg and bg attributes. This is
	// because the readAttr function in termui is unexported.
	txCells := termui.DefaultTxBuilder.Build(
		msg.GetContent(),
		termui.ColorDefault, termui.ColorDefault,
	)

	// Text
	for _, r := range msg.Content {
		cells = append(
			cells,
			termui.Cell{
				Ch: r,
				Fg: txCells[0].Fg,
				Bg: txCells[0].Bg,
			},
		)
	}

	return cells
}

// Help shows the usage and key bindings in the chat pane
func (c *Chat) Help(usage string, cfg *config.Config) {
	msgUsage := Message{
		ID:      fmt.Sprintf("%d", time.Now().UnixNano()),
		Content: usage,
	}

	c.Messages[msgUsage.ID] = msgUsage

	for mode, mapping := range cfg.KeyMap {
		msgMode := Message{
			ID:      fmt.Sprintf("%d", time.Now().UnixNano()),
			Content: fmt.Sprintf("%s", strings.ToUpper(mode)),
		}
		c.Messages[msgMode.ID] = msgMode

		msgNewline := Message{
			ID:      fmt.Sprintf("%d", time.Now().UnixNano()),
			Content: "",
		}
		c.Messages[msgNewline.ID] = msgNewline

		var keys []string
		for k := range mapping {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			msgKey := Message{
				ID:      fmt.Sprintf("%d", time.Now().UnixNano()),
				Content: fmt.Sprintf("    %-12s%-15s", k, mapping[k]),
			}
			c.Messages[msgKey.ID] = msgKey
		}

		msgNewline.ID = fmt.Sprintf("%d", time.Now().UnixNano())
		c.Messages[msgNewline.ID] = msgNewline
	}
}
