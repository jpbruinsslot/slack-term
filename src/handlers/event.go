package handlers

import (
	"github.com/gizak/termui"
	"github.com/nlopes/slack"

	"github.com/erroneousboat/slack-term/src/context"
	"github.com/erroneousboat/slack-term/src/views"
)

func RegisterEventHandlers(ctx *context.AppContext) {
	termui.Handle("/sys/kbd/", anyKeyHandler(ctx))
	termui.Handle("/sys/wnd/resize", resizeHandler(ctx))
	termui.Handle("/timer/1s", timeHandler(ctx))
	incomingMessageHandler(ctx)
}

func anyKeyHandler(ctx *context.AppContext) func(termui.Event) {
	return func(e termui.Event) {
		key := e.Data.(termui.EvtKbd).KeyStr

		if ctx.Mode == context.CommandMode {
			switch key {
			case "q":
				actionQuit()
				return
			case "j":
				actionMoveCursorDownChannels(ctx)
				return
			case "k":
				actionMoveCursorUpChannels(ctx)
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

func timeHandler(ctx *context.AppContext) func(termui.Event) {
	return func(e termui.Event) {
	}
}

// TODO: it'll always add the latest message
func incomingMessageHandler(ctx *context.AppContext) {
	go func() {
		for {
			select {
			case msg := <-ctx.Service.RTM.IncomingEvents:
				switch ev := msg.Data.(type) {
				case *slack.MessageEvent:

					// Construct message
					m := ctx.Service.CreateMessageFromMessageEvent(ev)

					// Add message to the selected channel
					if ev.Channel == ctx.View.Channels.SlackChannels[ctx.View.Channels.SelectedChannel].ID {
						ctx.View.Chat.AddMessage(m)
						termui.Render(ctx.View.Chat)
					}
				}
			}
		}
	}()
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
		ctx.View.Input.SendMessage(
			ctx.Service,
			ctx.View.Channels.SlackChannels[ctx.View.Channels.SelectedChannel].ID,
			ctx.View.Input.Text(),
		)
		ctx.View.Input.Clear()
		ctx.View.Refresh()
	}
}

func actionQuit() {
	termui.StopLoop()
}

func actionInsertMode(ctx *context.AppContext) {
	ctx.Mode = context.InsertMode
	ctx.View.Mode.Par.Text = "INSERT"
	termui.Render(ctx.View.Mode)
}

func actionCommandMode(ctx *context.AppContext) {
	ctx.Mode = context.CommandMode
	ctx.View.Mode.Par.Text = "NORMAL"
	termui.Render(ctx.View.Mode)
}

func actionGetMessages(ctx *context.AppContext) {
	ctx.View.Chat.GetMessages(
		ctx.Service,
		ctx.View.Channels.SlackChannels[ctx.View.Channels.SelectedChannel].ID,
	)

	termui.Render(ctx.View.Chat)
}

func actionGetChannels(ctx *context.AppContext) {
	ctx.View.Channels.GetChannels(ctx.Service)
	termui.Render(ctx.View.Channels)
}

func actionMoveCursorUpChannels(ctx *context.AppContext) {
	ctx.View.Channels.MoveCursorUp()
	actionChangeChannel(ctx)
}

func actionMoveCursorDownChannels(ctx *context.AppContext) {
	ctx.View.Channels.MoveCursorDown()
	actionChangeChannel(ctx)
}

func actionChangeChannel(ctx *context.AppContext) {
	// Clear messages from Chat pane
	ctx.View.Chat.ClearMessages()

	// Get message for the new channel
	ctx.View.Chat.GetMessages(
		ctx.Service,
		ctx.View.Channels.SlackChannels[ctx.View.Channels.SelectedChannel].ID,
	)

	// Set channel name for the Chat pane
	ctx.View.Chat.SetBorderLabel(
		ctx.View.Channels.SlackChannels[ctx.View.Channels.SelectedChannel].Name,
	)

	termui.Render(ctx.View.Channels)
	termui.Render(ctx.View.Chat)
}
