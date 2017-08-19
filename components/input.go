package components

import (
	"github.com/gizak/termui"

	"github.com/erroneousboat/slack-term/service"
)

// Input is the definition of an Input component
type Input struct {
	Par            *termui.Par
	Text           []rune
	CursorPosition int
}

// CreateInput is the constructor of the Input struct
func CreateInput() *Input {
	input := &Input{
		Par:            termui.NewPar(""),
		Text:           make([]rune, 0),
		CursorPosition: 0,
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
		termui.Cell{
			Ch: char.Ch,
			Fg: i.Par.TextBgColor,
			Bg: i.Par.TextFgColor,
		},
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

// SendMessage send the input text through the SlackService
func (i *Input) SendMessage(svc *service.SlackService, channel string, message string) {
	svc.SendMessage(channel, message)
}

// Insert will insert a given key at the place of the current CursorPosition
func (i *Input) Insert(key rune) {
	if len(i.Text) < i.Par.InnerBounds().Dx()-1 {

		left := make([]rune, len(i.Text[0:i.CursorPosition]))
		copy(left, i.Text[0:i.CursorPosition])
		left = append(left, key)

		i.Text = append(left, i.Text[i.CursorPosition:]...)

		i.Par.Text = string(i.Text)
		i.MoveCursorRight()
	}
}

// Backspace will remove a character in front of the CursorPosition
func (i *Input) Backspace() {
	if i.CursorPosition > 0 {
		i.Text = append(i.Text[0:i.CursorPosition-1], i.Text[i.CursorPosition:]...)
		i.Par.Text = string(i.Text)
		i.MoveCursorLeft()
	}
}

// Delete will remove a character at the CursorPosition
func (i *Input) Delete() {
	if i.CursorPosition < len(i.Text) {
		i.Text = append(i.Text[0:i.CursorPosition], i.Text[i.CursorPosition+1:]...)
		i.Par.Text = string(i.Text)
	}
}

// MoveCursorRight will increase the current CursorPosition with 1
func (i *Input) MoveCursorRight() {
	if i.CursorPosition < len(i.Text) {
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
	i.Text = make([]rune, 0)
	i.Par.Text = ""
	i.CursorPosition = 0
}

// GetText returns the text currently in the input
func (i *Input) GetText() string {
	return i.Par.Text
}
