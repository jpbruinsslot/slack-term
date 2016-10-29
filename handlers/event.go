package handlers

import (
	"strconv"
	"time"

	"github.com/gizak/termui"
	"github.com/nlopes/slack"
	termbox "github.com/nsf/termbox-go"

	"github.com/erroneousboat/slack-term/context"
	"github.com/erroneousboat/slack-term/views"
)

var timer *time.Timer

// actionMap binds specific action names to the function counterparts,
// these action names can then be used to bind them to specific keys
// in the Config.
var actionMap = map[string]func(*context.AppContext){
	"space":          actionSpace,
	"backspace":      actionBackSpace,
	"delete":         actionDelete,
	"cursor-right":   actionMoveCursorRight,
	"cursor-left":    actionMoveCursorLeft,
	"send":           actionSend,
	"quit":           actionQuit,
	"mode-insert":    actionInsertMode,
	"mode-command":   actionCommandMode,
	"channel-up":     actionMoveCursorUpChannels,
	"channel-down":   actionMoveCursorDownChannels,
	"channel-top":    actionMoveCursorTopChannels,
	"channel-bottom": actionMoveCursorBottomChannels,
	"chat-up":        actionScrollUpChat,
	"chat-down":      actionScrollDownChat,
}

func RegisterEventHandlers(ctx *context.AppContext) {
	anyKeyHandler(ctx)
	incomingMessageHandler(ctx)
	termui.Handle("/sys/wnd/resize", resizeHandler(ctx))
}

func anyKeyHandler(ctx *context.AppContext) {
	go func() {
		for {
			ev := termbox.PollEvent()

			if ev.Type != termbox.EventKey {
				continue
			}

			keyStr := getKeyString(ev)

			// Get the action name (actionStr) from the key that
			// has been pressed. If this is found try to uncover
			// the associated function with this key and execute
			// it.
			actionStr, ok := ctx.Config.KeyMap[ctx.Mode][keyStr]
			if ok {
				action, ok := actionMap[actionStr]
				if ok {
					action(ctx)
				}
			} else {
				if ctx.Mode == context.InsertMode && ev.Ch != 0 {
					actionInput(ctx.View, ev.Ch)
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

func actionInput(view *views.View, key rune) {
	view.Input.Insert(key)
	termui.Render(view.Input)
}

func actionSpace(ctx *context.AppContext) {
	actionInput(ctx.View, ' ')
}

func actionBackSpace(ctx *context.AppContext) {
	ctx.View.Input.Backspace()
	termui.Render(ctx.View.Input)
}

func actionDelete(ctx *context.AppContext) {
	ctx.View.Input.Delete()
	termui.Render(ctx.View.Input)
}

func actionMoveCursorRight(ctx *context.AppContext) {
	ctx.View.Input.MoveCursorRight()
	termui.Render(ctx.View.Input)
}

func actionMoveCursorLeft(ctx *context.AppContext) {
	ctx.View.Input.MoveCursorLeft()
	termui.Render(ctx.View.Input)
}

func actionSend(ctx *context.AppContext) {
	if !ctx.View.Input.IsEmpty() {

		// Clear message before sending, to combat
		// quick succession of actionSend
		message := ctx.View.Input.GetText()
		ctx.View.Input.Clear()
		ctx.View.Refresh()

		ctx.View.Input.SendMessage(
			ctx.Service,
			ctx.Service.Channels[ctx.View.Channels.SelectedChannel].ID,
			message,
		)
	}
}

func actionQuit(*context.AppContext) {
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

func actionMoveCursorUpChannels(ctx *context.AppContext) {
	go func() {
		if timer != nil {
			timer.Stop()
		}

		ctx.View.Channels.MoveCursorUp()
		termui.Render(ctx.View.Channels)

		timer = time.NewTimer(time.Second / 4)
		<-timer.C

		actionChangeChannel(ctx)
	}()
}

func actionMoveCursorDownChannels(ctx *context.AppContext) {
	go func() {
		if timer != nil {
			timer.Stop()
		}

		ctx.View.Channels.MoveCursorDown()
		termui.Render(ctx.View.Channels)

		timer = time.NewTimer(time.Second / 4)
		<-timer.C

		actionChangeChannel(ctx)
	}()
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

// GetKeyString will return a string that resembles the key event from
// termbox. This is blatanly copied from termui because it is an unexported
// function.
//
// See:
// - https://github.com/gizak/termui/blob/a7e3aeef4cdf9fa2edb723b1541cb69b7bb089ea/events.go#L31-L72
// - https://github.com/nsf/termbox-go/blob/master/api_common.go
func getKeyString(e termbox.Event) string {
	var ek string

	k := string(e.Ch)
	pre := ""
	mod := ""

	if e.Mod == termbox.ModAlt {
		mod = "M-"
	}
	if e.Ch == 0 {
		if e.Key > 0xFFFF-12 {
			k = "<f" + strconv.Itoa(0xFFFF-int(e.Key)+1) + ">"
		} else if e.Key > 0xFFFF-25 {
			ks := []string{"<insert>", "<delete>", "<home>", "<end>", "<previous>", "<next>", "<up>", "<down>", "<left>", "<right>"}
			k = ks[0xFFFF-int(e.Key)-12]
		}

		if e.Key <= 0x7F {
			pre = "C-"
			k = string('a' - 1 + int(e.Key))
			kmap := map[termbox.Key][2]string{
				termbox.KeyCtrlSpace:     {"C-", "<space>"},
				termbox.KeyBackspace:     {"", "<backspace>"},
				termbox.KeyTab:           {"", "<tab>"},
				termbox.KeyEnter:         {"", "<enter>"},
				termbox.KeyEsc:           {"", "<escape>"},
				termbox.KeyCtrlBackslash: {"C-", "\\"},
				termbox.KeyCtrlSlash:     {"C-", "/"},
				termbox.KeySpace:         {"", "<space>"},
				termbox.KeyCtrl8:         {"C-", "8"},
			}
			if sk, ok := kmap[e.Key]; ok {
				pre = sk[0]
				k = sk[1]
			}
		}
	}

	ek = pre + mod + k
	return ek
}
