package components

import (
	"strings"

	"github.com/erroneousboat/slack-term/src/service"
	"github.com/gizak/termui"
)

type Chat struct {
	List            *termui.List
	SelectedChannel string
}

func CreateChat(svc *service.SlackService, inputHeight int) *Chat {
	chat := &Chat{
		List: termui.NewList(),
	}

	chat.List.Height = termui.TermHeight() - inputHeight
	chat.List.Overflow = "wrap"
	// chat.GetMessages(svc)

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

	x := 0
	for _, cell := range cells {

		if cell.Ch == '\n' {
			lines = append(lines, line)
			line = Line{}
			x = 0
			continue
		}

		if x+cell.Width() > c.List.InnerBounds().Dx() {
			lines = append(lines, line)
			line = Line{}
			x = 0
		}

		line.cells = append(line.cells, cell)
		x++
	}
	lines = append(lines, line)

	// We will print lines bottom up, it will loop over the lines
	// backwards and for every line it'll set the cell in that line
	buf := c.List.Buffer()
	linesHeight := len(lines)
	paneMinY := c.List.InnerBounds().Min.Y
	paneMaxY := c.List.InnerBounds().Max.Y

	currentY := paneMaxY - 1
	for i := linesHeight - 1; i >= 0; i-- {
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
			buf.Set(x, currentY, termui.Cell{Ch: ' '})
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
			buf.Set(x, currentY, termui.Cell{Ch: ' '})
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

func (c *Chat) GetMessages(svc *service.SlackService, channel string) {
	// Get the count of message that fit in the pane
	count := c.List.InnerBounds().Max.Y - c.List.InnerBounds().Min.Y
	messages := svc.GetMessages(channel, count)

	for _, message := range messages {
		c.AddMessage(message)
	}
}

func (c *Chat) AddMessage(message string) {
	c.List.Items = append(c.List.Items, message)
}

func (c *Chat) ScrollUp() {
}

func (c *Chat) ScrollDown() {}
