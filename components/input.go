package components

import (
	"strings"

	"github.com/erroneousboat/termui"
	runewidth "github.com/mattn/go-runewidth"
)

// Input is the definition of an Input component
type Input struct {
	Par                  *termui.Par
	Text                 []rune
	Lines                [][]rune
	CursorPositionScreen int
	CursorPositionText   int
	Offset               int
	Line                 int
}

// CreateInput is the constructor of the Input struct
func CreateInputComponent() *Input {
	input := &Input{
		Par:                  termui.NewPar(""),
		Text:                 make([]rune, 0),
		Lines:                make([][]rune, 1),
		CursorPositionScreen: 0,
		CursorPositionText:   0,
		Offset:               0,
		Line:                 0,
	}

	input.Lines[0] = input.Text
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

// Insert will insert a given key at the place of the current CursorPositionText
func (i *Input) Insert(key rune) {
	// Append key to the left side
	left := make([]rune, len(i.Text[0:i.CursorPositionText]))
	copy(left, i.Text[0:i.CursorPositionText])
	left = append(left, key)

	// Combine left and right side
	i.UpdateText(append(left, i.Text[i.CursorPositionText:]...))

	i.MoveCursorRight()
}

func (i *Input) UpdateText(txt []rune) {
	i.Text = txt
	i.Lines[i.Line] = i.Text
}

func (i *Input) UpdatePar() {
	i.Par.Text = string(i.Text[i.Offset:])
}

// Backspace will remove a character in front of the CursorPositionText
func (i *Input) Backspace() {
	if i.CursorPositionText > 0 {
		i.MoveCursorLeft()
		i.UpdateText(append(i.Text[0:i.CursorPositionText], i.Text[i.CursorPositionText+1:]...))
		i.UpdatePar()
	} else if i.Line > 0 {
		// Delete this line if there's one before
		var prev = i.Lines[:i.Line + 1]
		var newlines [][]rune
		if i.Line + 1 < len(i.Lines) {
			newlines = append(prev, i.Lines[i.Line + 1:]...)
		} else {
			newlines = prev
		}
		i.Lines = newlines
		i.MoveCursorUp()
	}
}

// Delete will remove a character at the CursorPositionText
func (i *Input) Delete() {
	if i.CursorPositionText < len(i.Text) {
		i.UpdateText(append(i.Text[0:i.CursorPositionText], i.Text[i.CursorPositionText+1:]...))
		i.UpdatePar()
	}
}

func (i *Input) MoveCursorUp() {
	if i.Line > 0 {
		i.Line--
		i.CursorPositionText = 0
		i.CursorPositionScreen = 0
		i.Offset = 0
		i.Text = i.Lines[i.Line]
	}

	i.UpdatePar()
}

func (i *Input) MoveCursorDown() {
	if i.Line + 1 >= len(i.Lines) {
		// Make a new line
		i.Lines = append(i.Lines, make([]rune, 0))
	}
	// Go to the line below
	i.Line++
	i.CursorPositionText = 0
	i.CursorPositionScreen = 0
	i.Offset = 0
	i.Text = i.Lines[i.Line]
	i.UpdatePar()
}

// MoveCursorRight will increase the current CursorPositionText with 1
func (i *Input) MoveCursorRight() {
	if i.CursorPositionText < len(i.Text) {
		i.CursorPositionText++
		i.ScrollRight()
	}

	i.UpdatePar()
}

// MoveCursorLeft will decrease the current CursorPositionText with 1
func (i *Input) MoveCursorLeft() {
	if i.CursorPositionText > 0 {
		i.CursorPositionText--
		i.ScrollLeft()
	}

	i.UpdatePar()
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
	if (i.CursorPositionScreen + i.GetRuneWidthLeft()) > i.Par.InnerBounds().Dx()-1 {

		// Increase offset to show what is on the right side
		if i.Offset < len(i.Text) {
			i.Offset = i.CalculateOffset()
			i.CursorPositionScreen = i.GetRuneWidthOffsetToCursor()
		}
	} else {
		i.CursorPositionScreen += i.GetRuneWidthLeft()
	}
}

// CalculateOffset will, based on the width of the runes on the
// left of the text cursor, calculate the offset that needs to
// be used by the Inpute Component
func (i *Input) CalculateOffset() int {
	var offset int

	var currentRuneWidth int
	for j := (i.CursorPositionText - 1); currentRuneWidth < i.GetMaxWidth()-1; j-- {
		currentRuneWidth += runewidth.RuneWidth(i.Text[j])
		offset = j
	}

	return offset
}

// GetRunWidthOffsetToCursor will get the rune width of all
// the runes from the offset until the text cursor
func (i *Input) GetRuneWidthOffsetToCursor() int {
	return runewidth.StringWidth(string(i.Text[i.Offset:i.CursorPositionText]))
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
	return len(i.Text) == 0 && len(i.Lines) == 1
}

// Clear will empty the input and move the cursor to the start position
func (i *Input) Clear() {
	i.Text = make([]rune, 0)
	i.Lines = make([][]rune, 1)
	i.Lines[0] = i.Text
	i.Par.Text = ""
	i.CursorPositionScreen = 0
	i.CursorPositionText = 0
	i.Offset = 0
	i.Line = 0
}

// GetText returns the text currently in the input
func (i *Input) GetText() string {
	var fulltext = make([]string, len(i.Lines))
	for i, txt := range i.Lines {
		fulltext[i] = string(txt)
	}
	return strings.Join(fulltext, "\n")
}

// GetMaxWidth returns the maximum number of positions
// the Input component can display
func (i *Input) GetMaxWidth() int {
	return i.Par.InnerBounds().Dx() - 1
}
