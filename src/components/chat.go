package components

import (
	"strings"

	"github.com/gizak/termui"
)

type Chat struct {
	List *termui.List
}

func CreateChat(inputHeight int) *Chat {
	chat := &Chat{
		List: termui.NewList(),
	}

	chat.List.Height = termui.TermHeight() - inputHeight
	chat.List.Overflow = "wrap"
	chat.LoadMessages()

	return chat
}

// Buffer implements interface termui.Bufferer
func (c *Chat) Buffer() termui.Buffer {
	// Build cells, after every item put a newline
	cells := termui.DefaultTxBuilder.Build(
		strings.Join(c.List.Items, "\n"),
		c.List.ItemFgColor, c.List.ItemBgColor,
	)

	type Line struct {
		cells []termui.Cell
	}

	// Uncover how many lines there are in total for all items
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

	// We will print lines bottom up
	buf := c.List.Buffer()
	linesHeight := len(lines)

	windowMinY := c.List.InnerBounds().Min.Y
	windowMaxY := c.List.InnerBounds().Max.Y

	currentY := windowMaxY - 1
	for i := linesHeight - 1; i >= 0; i-- {
		if currentY <= windowMinY {
			break
		}

		x := c.List.InnerBounds().Min.X
		for _, cell := range lines[i].cells {
			buf.Set(x, currentY, cell)
			x += cell.Width()
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

func (c *Chat) LoadMessages() {
	messages := []string{
		"[jp] hello world",
		"[erroneousboat] foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar",
	}

	for _, message := range messages {
		c.AddMessages(message)
	}
}

func (c *Chat) AddMessages(message string) {
	c.List.Items = append(c.List.Items, message)
}

func (c *Chat) ScrollUp() {
}

func (c *Chat) ScrollDown() {}
