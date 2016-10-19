package handlers

import (
	"github.com/gizak/termui"
	"github.com/nlopes/slack"
	termbox "github.com/nsf/termbox-go"

	"github.com/erroneousboat/slack-term/context"
	"github.com/erroneousboat/slack-term/views"
)

func RegisterEventHandlers(ctx *context.AppContext) {
	anyKeyHandler(ctx)
	incomingMessageHandler(ctx)
	termui.Handle("/sys/wnd/resize", resizeHandler(ctx))
}

func anyKeyHandler(ctx *context.AppContext) {
	go func() {
		for {
			ev := termbox.PollEvent()

			if ev.Type == termbox.EventKey {
				if ctx.Mode == context.CommandMode {
					switch ev.Key {
					case termbox.KeyPgup:
						actionScrollUpChat(ctx)
					case termbox.KeyCtrlB:
						actionScrollUpChat(ctx)
					case termbox.KeyCtrlU:
						actionScrollUpChat(ctx)
					case termbox.KeyPgdn:
						actionScrollDownChat(ctx)
					case termbox.KeyCtrlF:
						actionScrollDownChat(ctx)
					case termbox.KeyCtrlD:
						actionScrollDownChat(ctx)
					default:
						switch ev.Ch {
						case 'q':
							actionAskQuit(ctx)
						case 'j':
							actionMoveCursorDownChannels(ctx)
						case 'k':
							actionMoveCursorUpChannels(ctx)
						case 'g':
							actionMoveCursorTopChannels(ctx)
						case 'G':
							actionMoveCursorBottomChannels(ctx)
						case 'i':
							actionInsertMode(ctx)
						}
					}
				} else if ctx.Mode == context.InsertMode {
					switch ev.Key {
					case termbox.KeyEsc:
						actionCommandMode(ctx)
					case termbox.KeyEnter:
						actionSend(ctx)
					case termbox.KeySpace:
						actionInput(ctx.View, " ")
					case termbox.KeyBackspace, termbox.KeyBackspace2:
						actionBackSpace(ctx.View)
					case termbox.KeyDelete:
						actionDelete(ctx.View)
					case termbox.KeyArrowRight:
						actionMoveCursorRight(ctx.View)
					case termbox.KeyArrowLeft:
						actionMoveCursorLeft(ctx.View)
					default:
						actionInput(ctx.View, string(ev.Ch))
					}
				} else if ctx.Mode == context.QuitMode {
					switch ev.Key {
					case termbox.KeyEsc:
						actionCommandMode(ctx)
					case termbox.KeyEnter:
						actionCommandMode(ctx)
					default:
						switch ev.Ch {
						case 'y', 'Y':
							actionQuit()
						case 'n', 'N':
							actionCommandMode(ctx)
						}
					}
				}
			}
		}
	}()
}

func resizeHandler(ctx *context.AppContext) func(termui.Event) {
	return func(e termui.Event) {
		actionResize(ctx)
	}
}

func incomingMessageHandler(ctx *context.AppContext) {
	go func() {
		for {
			select {
			case msg := <-ctx.Service.RTM.IncomingEvents:
				switch ev := msg.Data.(type) {
				case *slack.MessageEvent:

					// Construct message
					msg := ctx.Service.CreateMessageFromMessageEvent(ev)

					// Add message to the selected channel
					if ev.Channel == ctx.Service.Channels[ctx.View.Channels.SelectedChannel].ID {

						// reverse order of messages, mainly done
						// when attachments are added to message
						for i := len(msg) - 1; i >= 0; i-- {
							ctx.View.Chat.AddMessage(msg[i])
						}

						termui.Render(ctx.View.Chat)

						// TODO: set Chat.Offset to 0, to automatically scroll
						// down?
					}

					// Set new message indicator for channel, I'm leaving
					// this here because I also want to be notified when
					// I'm currently in a channel but not in the terminal
					// window (tmux). But only create a notification when
					// it comes from someone else but the current user.
					if ev.User != ctx.Service.CurrentUserID {
						actionNewMessage(ctx, ev.Channel)
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
	view.Input.Backspace()
	termui.Render(view.Input)
}

func actionDelete(view *views.View) {
	view.Input.Delete()
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

		// Clear message before sending, to combat
		// quick succession of actionSend
		message := ctx.View.Input.Text()
		ctx.View.Input.Clear()
		ctx.View.Refresh()

		ctx.View.Input.SendMessage(
			ctx.Service,
			ctx.Service.Channels[ctx.View.Channels.SelectedChannel].ID,
			message,
		)
	}
}

// Ask before quit
func actionAskQuit(ctx *context.AppContext) {
	ctx.Mode = context.QuitMode
	ctx.View.Mode.Par.Text = "Quit? [y/N]"
	termui.Render(ctx.View.Mode)
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
		ctx.Service.Channels[ctx.View.Channels.SelectedChannel],
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

func actionMoveCursorTopChannels(ctx *context.AppContext) {
	ctx.View.Channels.MoveCursorTop()
	actionChangeChannel(ctx)
}

func actionMoveCursorBottomChannels(ctx *context.AppContext) {
	ctx.View.Channels.MoveCursorBottom()
	actionChangeChannel(ctx)
}

func actionChangeChannel(ctx *context.AppContext) {
	// Clear messages from Chat pane
	ctx.View.Chat.ClearMessages()

	// Get message for the new channel
	ctx.View.Chat.GetMessages(
		ctx.Service,
		ctx.Service.SlackChannels[ctx.View.Channels.SelectedChannel],
	)

	// Set channel name for the Chat pane
	ctx.View.Chat.SetBorderLabel(
		ctx.Service.Channels[ctx.View.Channels.SelectedChannel].Name,
	)

	termui.Render(ctx.View.Channels)
	termui.Render(ctx.View.Chat)
}

func actionNewMessage(ctx *context.AppContext, channelID string) {
	ctx.View.Channels.NewMessage(ctx.Service, channelID)
	termui.Render(ctx.View.Channels)
}

func actionScrollUpChat(ctx *context.AppContext) {
	ctx.View.Chat.ScrollUp()
	termui.Render(ctx.View.Chat)
}

func actionScrollDownChat(ctx *context.AppContext) {
	ctx.View.Chat.ScrollDown()
	termui.Render(ctx.View.Chat)
}
