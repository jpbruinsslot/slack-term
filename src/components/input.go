package components

import "github.com/gizak/termui"

type Input struct {
	Block          *termui.Par
	CursorPosition int
	CursorFgColor  termui.Attribute
	CursorBgColor  termui.Attribute
}

func CreateInput() *Input {
	input := &Input{
		Block:          termui.NewPar(""),
		CursorPosition: 0,
		CursorBgColor:  termui.ColorBlack,
		CursorFgColor:  termui.ColorWhite,
	}

	input.Block.Height = 3

	return input
}

// implements interface termui.Bufferer
func (i *Input) Buffer() termui.Buffer {
	return i.Block.Buffer()
}

// implements interface termui.GridBufferer
func (i *Input) GetHeight() int {
	return i.Block.GetHeight()
}

// implements interface termui.GridBufferer
func (i *Input) SetWidth(w int) {
	i.Block.SetWidth(w)
}

// implements interface termui.GridBufferer
func (i *Input) SetX(x int) {
	i.Block.SetX(x)
}

// implements interface termui.GridBufferer
func (i *Input) SetY(y int) {
	i.Block.SetY(y)
}

func (i *Input) Insert(key string) {
	i.Block.Text = i.Block.Text + key
	i.MoveCursorRight()
}

func (i *Input) Remove() {
	if i.CursorPosition > 0 {
		i.Block.Text = i.Block.Text[0:i.CursorPosition-1] + i.Block.Text[i.CursorPosition:len(i.Block.Text)]
		i.MoveCursorLeft()
	}
}

func (i *Input) MoveCursorRight() {
	if i.CursorPosition < len(i.Block.Text) {
		i.CursorPosition++
		i.Block.Block.Buffer().Set(
			i.CursorPosition,       // x
			i.Block.Block.InnerY(), // y
			termui.Cell{Ch: rune('$'), Fg: termui.ColorBlack, Bg: termui.ColorWhite},
		)
	}
}

func (i *Input) MoveCursorLeft() {
	if i.CursorPosition > 0 {
		i.CursorPosition--
		i.Block.Block.Buffer().Set(
			i.CursorPosition,       // x
			i.Block.Block.InnerY(), // y
			termui.Cell{Ch: rune('$'), Fg: termui.ColorBlack, Bg: termui.ColorWhite},
		)
	}
}

func (i *Input) IsEmpty() bool {
	if i.Block.Text == "" {
		return true
	}
	return false
}

func (i *Input) Clear() {
	i.Block.Text = ""
	i.CursorPosition = 0
}

func (i *Input) Text() string {
	return i.Block.Text
}
