package components

import (
	"fmt"
	"html"
	"sort"
	"strings"
	"time"

	"github.com/erroneousboat/termui"

	"github.com/erroneousboat/slack-term/config"
)

type Message struct {
	Time    time.Time
	Name    string
	Content string

	StyleTime string
	StyleName string
	StyleText string
}

func (m Message) ToString() string {
	if (m.Time != time.Time{} && m.Name != "") {

		return html.UnescapeString(
			fmt.Sprintf(
				"[[%s]](%s) [<%s>](%s) [%s](%s)",
				m.Time.Format("15:04"),
				m.StyleTime,
				m.Name,
				m.StyleName,
				m.Content,
				m.StyleText,
			),
		)
	} else {
		return html.UnescapeString(
			fmt.Sprintf("[%s](%s)", m.Content, m.StyleText),
		)
	}
}

// Chat is the definition of a Chat component
type Chat struct {
	List   *termui.List
	Offset int
}

// CreateChat is the constructor for the Chat struct
func CreateChatComponent(inputHeight int) *Chat {
	chat := &Chat{
		List:   termui.NewList(),
		Offset: 0,
	}

	chat.List.Height = termui.TermHeight() - inputHeight
	chat.List.Overflow = "wrap"

	return chat
}

// Buffer implements interface termui.Bufferer
func (c *Chat) Buffer() termui.Buffer {
	// Build cells, after every item put a newline
	cells := termui.DefaultTxBuilder.Build(
		strings.Join(c.List.Items, "\n"),
		c.List.ItemFgColor, c.List.ItemBgColor,
	)

	// We will create an array of Line structs, this allows us
	// to more easily render the items in a list. We will range
	// over the cells we've created and create a Line within
	// the bounds of the Chat pane
	type Line struct {
		cells []termui.Cell
	}

	lines := []Line{}
	line := Line{}

	// When we encounter a newline or are at the bounds of the chat view we
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
		x++
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
			x++
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
			x++
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

// SetMessages will put the provided messages into the Items field of the
// Chat view
func (c *Chat) SetMessages(messages []string) {
	// Reset offset first, when scrolling in view and changing channels we
	// want the offset to be 0 when loading new messages
	c.Offset = 0

	for _, msg := range messages {
		c.List.Items = append(c.List.Items, html.UnescapeString(msg))
	}
}

// AddMessage adds a single message to List.Items
func (c *Chat) AddMessage(message string) {
	c.List.Items = append(c.List.Items, html.UnescapeString(message))
}

// ClearMessages clear the List.Items
func (c *Chat) ClearMessages() {
	c.List.Items = []string{}
}

// ScrollUp will render the chat messages based on the Offset of the Chat
// pane.
//
// Offset is 0 when scrolled down. (we loop backwards over the array, so we
// start with rendering last item in the list at the maximum y of the Chat
// pane). Increasing the Offset will thus result in subtracting the offset
// from the len(Chat.List.Items).
func (c *Chat) ScrollUp() {
	c.Offset = c.Offset + 10

	// Protect overscrolling
	if c.Offset > len(c.List.Items) {
		c.Offset = len(c.List.Items)
	}
}

// ScrollDown will render the chat messages based on the Offset of the Chat
// pane.
//
// Offset is 0 when scrolled down. (we loop backwards over the array, so we
// start with rendering last item in the list at the maximum y of the Chat
// pane). Increasing the Offset will thus result in subtracting the offset
// from the len(Chat.List.Items).
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

// Help shows the usage and key bindings in the chat pane
func (c *Chat) Help(cfg *config.Config) {
	help := []string{
		"slack-term - slack client for your terminal",
		"",
		"USAGE:",
		"    slack-term -config [path-to-config]",
		"",
		"KEY BINDINGS:",
		"",
	}

	for mode, mapping := range cfg.KeyMap {
		help = append(help, fmt.Sprintf("    %s", strings.ToUpper(mode)))
		help = append(help, "")

		var keys []string
		for k := range mapping {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			help = append(help, fmt.Sprintf("    %-12s%-15s", k, mapping[k]))
		}
		help = append(help, "")
	}

	c.List.Items = help
}
