package components

import (
	"github.com/erroneousboat/slack-term/src/service"
	"github.com/gizak/termui"
)

// Input is the definition of and input box
type Input struct {
	Par            *termui.Par
	CursorPosition int
	CursorFgColor  termui.Attribute
	CursorBgColor  termui.Attribute
}

// CreateInput is the constructor of the Input struct
func CreateInput() *Input {
	input := &Input{
		Par:            termui.NewPar(""),
		CursorPosition: 0,
		CursorBgColor:  termui.ColorBlack,
		CursorFgColor:  termui.ColorWhite,
	}

	input.Par.Height = 3

	return input
}

// Buffer implements interface termui.Bufferer
func (i *Input) Buffer() termui.Buffer {
	buf := i.Par.Buffer()

	// Set visible cursor
	char := buf.At(i.Par.InnerX()+i.CursorPosition, i.Par.Block.InnerY())
	buf.Set(
		i.Par.InnerX()+i.CursorPosition,
		i.Par.Block.InnerY(),
		termui.Cell{Ch: char.Ch, Fg: termui.ColorBlack, Bg: termui.ColorWhite},
	)

	return buf
}

// GetHeight implements interface termui.GridBufferer
func (i *Input) GetHeight() int {
	return i.Par.Block.GetHeight()
}

// SetWidth implements interface termui.GridBufferer
func (i *Input) SetWidth(w int) {
	i.Par.SetWidth(w)
}

// SetX implements interface termui.GridBufferer
func (i *Input) SetX(x int) {
	i.Par.SetX(x)
}

// SetY implements interface termui.GridBufferer
func (i *Input) SetY(y int) {
	i.Par.SetY(y)
}

func (i *Input) SendMessage(svc *service.SlackService, channel string, message string) {
	svc.SendMessage(channel, message)
}

// Insert will insert a given key at the place of the current CursorPosition
func (i *Input) Insert(key string) {
	if len(i.Par.Text) < i.Par.InnerBounds().Dx()-1 {
		i.Par.Text = i.Par.Text[0:i.CursorPosition] + key + i.Par.Text[i.CursorPosition:len(i.Par.Text)]
		i.MoveCursorRight()
	}
}

// Remove will remove a character at the place of the current CursorPosition
func (i *Input) Remove() {
	if i.CursorPosition > 0 {
		i.Par.Text = i.Par.Text[0:i.CursorPosition-1] + i.Par.Text[i.CursorPosition:len(i.Par.Text)]
		i.MoveCursorLeft()
	}
}

// MoveCursorRight will increase the current CursorPosition with 1
func (i *Input) MoveCursorRight() {
	if i.CursorPosition < len(i.Par.Text) {
		i.CursorPosition++
	}
}

// MoveCursorLeft will decrease the current CursorPosition with 1
func (i *Input) MoveCursorLeft() {
	if i.CursorPosition > 0 {
		i.CursorPosition--
	}
}

// IsEmpty will return true when the input is empty
func (i *Input) IsEmpty() bool {
	if i.Par.Text == "" {
		return true
	}
	return false
}

// Clear will empty the input and move the cursor to the start position
func (i *Input) Clear() {
	i.Par.Text = ""
	i.CursorPosition = 0
}

// Text returns the text currently in the input
func (i *Input) Text() string {
	return i.Par.Text
}
