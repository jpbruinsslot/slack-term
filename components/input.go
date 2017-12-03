package components

import (
	"github.com/erroneousboat/termui"
	runewidth "github.com/mattn/go-runewidth"

	"github.com/erroneousboat/slack-term/service"
)

// Input is the definition of an Input component
type Input struct {
	Par                  *termui.Par
	Text                 []rune
	CursorPositionScreen int
	CursorPositionText   int
	Offset               int
}

// CreateInput is the constructor of the Input struct
func CreateInputComponent() *Input {
	input := &Input{
		Par:                  termui.NewPar(""),
		Text:                 make([]rune, 0),
		CursorPositionScreen: 0,
		CursorPositionText:   0,
		Offset:               0,
	}

	input.Par.Height = 3

	return input
}

// Buffer implements interface termui.Bufferer
func (i *Input) Buffer() termui.Buffer {
	buf := i.Par.Buffer()

	// Set visible cursor, get char at screen cursor position
	char := buf.At(i.Par.InnerX()+i.CursorPositionScreen, i.Par.Block.InnerY())

	buf.Set(
		i.Par.InnerX()+i.CursorPositionScreen,
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

// Insert will insert a given key at the place of the current CursorPositionText
func (i *Input) Insert(key rune) {
	// Append key to the left side
	left := make([]rune, len(i.Text[0:i.CursorPositionText]))
	copy(left, i.Text[0:i.CursorPositionText])
	left = append(left, key)

	// Combine left and right side
	i.Text = append(left, i.Text[i.CursorPositionText:]...)

	i.MoveCursorRight()
}

// Backspace will remove a character in front of the CursorPositionText
func (i *Input) Backspace() {
	if i.CursorPositionText > 0 {
		i.MoveCursorLeft()
		i.Text = append(i.Text[0:i.CursorPositionText], i.Text[i.CursorPositionText+1:]...)
		i.Par.Text = string(i.Text[i.Offset:])
	}
}

// Delete will remove a character at the CursorPositionText
func (i *Input) Delete() {
	if i.CursorPositionText < len(i.Text) {
		i.Text = append(i.Text[0:i.CursorPositionText], i.Text[i.CursorPositionText+1:]...)
		i.Par.Text = string(i.Text[i.Offset:])
	}
}

// MoveCursorRight will increase the current CursorPositionText with 1
func (i *Input) MoveCursorRight() {
	if i.CursorPositionText < len(i.Text) {
		i.CursorPositionText++
		i.ScrollRight()
	}

	i.Par.Text = string(i.Text[i.Offset:])
}

// MoveCursorLeft will decrease the current CursorPositionText with 1
func (i *Input) MoveCursorLeft() {
	if i.CursorPositionText > 0 {
		i.CursorPositionText--
		i.ScrollLeft()
	}

	i.Par.Text = string(i.Text[i.Offset:])
}

func (i *Input) ScrollLeft() {
	// Is the cursor at the far left of the Input component?
	if i.CursorPositionScreen == 0 {

		// Decrease offset to show what is on the left side
		if i.Offset > 0 {
			i.Offset--
		}
	} else {
		i.CursorPositionScreen -= i.GetRuneWidthRight()
	}
}

func (i *Input) ScrollRight() {
	// Is the cursor at the far right of the Input component, cursor
	// isn't at the end of the text
	if i.CursorPositionScreen == i.Par.InnerBounds().Dx()-1 {

		// Increase offset to show what is on the right side
		if i.Offset < len(i.Text) {
			i.Offset++
		}
	} else {
		i.CursorPositionScreen += i.GetRuneWidthLeft()
	}
}

// GetRuneWidthLeft will get the width of a rune on the left side
// of the CursorPositionText
func (i *Input) GetRuneWidthLeft() int {
	return runewidth.RuneWidth(i.Text[i.CursorPositionText-1])
}

// GetRuneWidthLeft will get the width of a rune on the right side
// of the CursorPositionText
func (i *Input) GetRuneWidthRight() int {
	return runewidth.RuneWidth(i.Text[i.CursorPositionText])
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
	i.CursorPositionScreen = 0
	i.CursorPositionText = 0
}

// GetText returns the text currently in the input
func (i *Input) GetText() string {
	return i.Par.Text
}
