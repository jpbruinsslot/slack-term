package views

import (
	termbox "github.com/nsf/termbox-go"
)

func Loading() {
	const loading string = "LOADING"

	w, h := termbox.Size()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	offset := (w / 2) - (len(loading) / 2)
	y := h / 2

	for x := 0; x < len(loading); x++ {
		termbox.SetCell(offset+x, y, rune(loading[x]), termbox.ColorDefault, termbox.ColorDefault)
	}

	termbox.Flush()
}
