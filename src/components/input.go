package components

import "github.com/gizak/termui"

type Input struct {
	Par            *termui.Par
	CursorPosition int
	CursorFgColor  termui.Attribute
	CursorBgColor  termui.Attribute
}

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

// implements interface termui.Bufferer
func (i *Input) Buffer() termui.Buffer {
	buf := i.Par.Buffer()

	// Set cursor
	char := buf.At(i.Par.InnerX()+i.CursorPosition, i.Par.Block.InnerY())
	buf.Set(
		i.Par.InnerX()+i.CursorPosition,
		i.Par.Block.InnerY(),
		termui.Cell{Ch: char.Ch, Fg: termui.ColorBlack, Bg: termui.ColorWhite},
	)

	return buf
}

// implements interface termui.GridBufferer
func (i *Input) GetHeight() int {
	return i.Par.Block.GetHeight()
}

// implements interface termui.GridBufferer
func (i *Input) SetWidth(w int) {
	i.Par.SetWidth(w)
}

// implements interface termui.GridBufferer
func (i *Input) SetX(x int) {
	i.Par.SetX(x)
}

// implements interface termui.GridBufferer
func (i *Input) SetY(y int) {
	i.Par.SetY(y)
}

func (i *Input) Insert(key string) {
	i.Par.Text = i.Par.Text[0:i.CursorPosition] + key + i.Par.Text[i.CursorPosition:len(i.Par.Text)]

	i.MoveCursorRight()
}

func (i *Input) Remove() {
	if i.CursorPosition > 0 {
		i.Par.Text = i.Par.Text[0:i.CursorPosition-1] + i.Par.Text[i.CursorPosition:len(i.Par.Text)]
		i.MoveCursorLeft()
	}
}

func (i *Input) MoveCursorRight() {
	if i.CursorPosition < len(i.Par.Text) {
		i.CursorPosition++
	}
}

func (i *Input) MoveCursorLeft() {
	if i.CursorPosition > 0 {
		i.CursorPosition--
	}
}

func (i *Input) IsEmpty() bool {
	if i.Par.Text == "" {
		return true
	}
	return false
}

func (i *Input) Clear() {
	i.Par.Text = ""
	i.CursorPosition = 0
}

func (i *Input) Text() string {
	return i.Par.Text
}
