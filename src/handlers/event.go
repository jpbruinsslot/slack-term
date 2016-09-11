package handlers

import (
	"github.com/gizak/termui"

	"github.com/erroneousboat/slack-term/src/context"
	"github.com/erroneousboat/slack-term/src/views"
)

func RegisterEventHandlers(ctx *context.AppContext) {
	termui.Handle("/sys/kbd/", anyKeyHandler(ctx))
	termui.Handle("/sys/wnd/resize", resizeHandler(ctx))
}

func anyKeyHandler(ctx *context.AppContext) func(termui.Event) {
	return func(e termui.Event) {
		key := e.Data.(termui.EvtKbd).KeyStr

		if ctx.Mode == context.CommandMode {
			switch key {
			case "q":
				actionQuit()
				return
			case "i":
				actionInsertMode(ctx)
				return
			}
		} else if ctx.Mode == context.InsertMode {
			switch key {
			case "<escape>":
				actionCommandMode(ctx)
				return
			case "<enter>":
				actionSend(ctx)
				return
			case "<space>":
				actionInput(ctx.View, " ")
				return
			case "<backspace>":
				actionBackSpace(ctx.View)
			case "C-8":
				actionBackSpace(ctx.View)
			case "<right>":
				actionMoveCursorRight(ctx.View)
			case "<left>":
				actionMoveCursorLeft(ctx.View)
			default:
				actionInput(ctx.View, key)
				return
			}

		}

	}
}

func resizeHandler(ctx *context.AppContext) func(termui.Event) {
	return func(e termui.Event) {
		actionResize(ctx)
	}
}

// FIXME: resize only seems to work for width and resizing it too small
// will cause termui to panic
func actionResize(ctx *context.AppContext) {
	termui.Body.Width = termui.TermWidth()
	termui.Body.Align()
	termui.Render(termui.Body)
}

func actionInput(view *views.View, key string) {
	view.Input.Insert(key)
	termui.Render(view.Input)
}

func actionBackSpace(view *views.View) {
	view.Input.Remove()
	termui.Render(view.Input)
}

func actionMoveCursorRight(view *views.View) {
	view.Input.MoveCursorRight()
	termui.Render(view.Input)
}

func actionMoveCursorLeft(view *views.View) {
	view.Input.MoveCursorLeft()
	termui.Render(view.Input)
}

func actionSend(ctx *context.AppContext) {
	if !ctx.View.Input.IsEmpty() {
		// FIXME
		ctx.View.Chat.Items = append(ctx.View.Chat.Items, ctx.View.Input.Text())
		ctx.View.Input.Clear()
		ctx.View.Refresh()
	}
}

func actionQuit() {
	termui.StopLoop()
}

func actionInsertMode(ctx *context.AppContext) {
	ctx.Mode = context.InsertMode
	ctx.View.Mode.Text = "INSERT"
	termui.Render(ctx.View.Mode)
}

func actionCommandMode(ctx *context.AppContext) {
	ctx.Mode = context.CommandMode
	ctx.View.Mode.Text = "NORMAL"
	termui.Render(ctx.View.Mode)
}
