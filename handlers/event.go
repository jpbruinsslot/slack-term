package handlers

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/erroneousboat/termui"
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
	"space":               actionSpace,
	"backspace":           actionBackSpace,
	"delete":              actionDelete,
	"cursor-right":        actionMoveCursorRight,
	"cursor-left":         actionMoveCursorLeft,
	"send":                actionSend,
	"quit":                actionQuit,
	"mode-insert":         actionInsertMode,
	"mode-command":        actionCommandMode,
	"mode-search":         actionSearchMode,
	"clear-input":         actionClearInput,
	"channel-up":          actionMoveCursorUpChannels,
	"channel-down":        actionMoveCursorDownChannels,
	"channel-top":         actionMoveCursorTopChannels,
	"channel-bottom":      actionMoveCursorBottomChannels,
	"channel-search-next": actionSearchNextChannels,
	"channel-search-prev": actionSearchPrevChannels,
	"chat-up":             actionScrollUpChat,
	"chat-down":           actionScrollDownChat,
	"help":                actionHelp,
}

func RegisterEventHandlers(ctx *context.AppContext) {
	eventHandler(ctx)
	messageHandler(ctx)
}

// eventHandler will handle events created by the user
func eventHandler(ctx *context.AppContext) {
	go func() {
		for {
			ctx.EventQueue <- termbox.PollEvent()
		}
	}()

	go func() {
		for {
			ev := <-ctx.EventQueue
			handleTermboxEvents(ctx, ev)
			handleMoreTermboxEvents(ctx, ev)

			// Place your debugging statements here
			if ctx.Debug {
				ctx.View.Debug.Println(
					"event received",
				)
			}
		}
	}()
}

func handleTermboxEvents(ctx *context.AppContext, ev termbox.Event) bool {
	switch ev.Type {
	case termbox.EventKey:
		actionKeyEvent(ctx, ev)
	case termbox.EventResize:
		actionResizeEvent(ctx, ev)
	}

	return true
}

func handleMoreTermboxEvents(ctx *context.AppContext, ev termbox.Event) bool {
	for {
		select {
		case ev := <-ctx.EventQueue:
			ok := handleTermboxEvents(ctx, ev)
			if !ok {
				return false
			}
		default:
			return true
		}
	}
}

// messageHandler will handle events created by the service
func messageHandler(ctx *context.AppContext) {
	go func() {
		for {
			select {
			case msg := <-ctx.Service.RTM.IncomingEvents:
				switch ev := msg.Data.(type) {
				case *slack.MessageEvent:

					// Construct message
					msg, err := ctx.Service.CreateMessageFromMessageEvent(ev)
					if err != nil {
						continue
					}

					// Add message to the selected channel
					if ev.Channel == ctx.Service.Channels[ctx.View.Channels.SelectedChannel].ID {

						// reverse order of messages, mainly done
						// when attachments are added to message
						for i := len(msg) - 1; i >= 0; i-- {
							ctx.View.Chat.AddMessage(
								msg[i].ToString(),
							)
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
				case *slack.PresenceChangeEvent:
					actionSetPresence(ctx, ev.User, ev.Presence)
				}
			}
		}
	}()
}

func actionKeyEvent(ctx *context.AppContext, ev termbox.Event) {

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
		} else if ctx.Mode == context.SearchMode && ev.Ch != 0 {
			actionSearch(ctx, ev.Ch)
		}
	}
}

func actionResizeEvent(ctx *context.AppContext, ev termbox.Event) {
	// When terminal window is too small termui will panic, here
	// we won't resize when the terminal window is too small.
	if termui.TermWidth() < 25 || termui.TermHeight() < 5 {
		return
	}

	termui.Body.Width = termui.TermWidth()

	// Vertical resize components
	ctx.View.Channels.List.Height = termui.TermHeight() - ctx.View.Input.Par.Height
	ctx.View.Chat.List.Height = termui.TermHeight() - ctx.View.Input.Par.Height
	ctx.View.Debug.List.Height = termui.TermHeight() - ctx.View.Input.Par.Height

	termui.Body.Align()
	termui.Render(termui.Body)
}

func actionInput(view *views.View, key rune) {
	view.Input.Insert(key)
	termui.Render(view.Input)
}

func actionClearInput(ctx *context.AppContext) {
	// Clear input
	ctx.View.Input.Clear()
	ctx.View.Refresh()

	// Set command mode
	actionCommandMode(ctx)
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

		// Send message
		ctx.Service.SendMessage(
			ctx.View.Channels.SelectedChannel,
			message,
		)

		// Clear notification icon if there is any
		ctx.Service.MarkAsRead(ctx.View.Channels.SelectedChannel)
		ctx.View.Channels.SetChannels(ctx.Service.ChannelsToString())
		termui.Render(ctx.View.Channels)
	}
}

// actionSearch will search through the channels based on the users
// input. A time is implemented to make sure the actual searching
// and changing of channels is done when the user's typing is paused.
func actionSearch(ctx *context.AppContext, key rune) {
	actionInput(ctx.View, key)

	go func() {
		if timer != nil {
			timer.Stop()
		}

		timer = time.NewTimer(time.Second / 4)
		<-timer.C

		// Only actually search when the time expires
		term := ctx.View.Input.GetText()
		ctx.View.Channels.Search(term)
		actionChangeChannel(ctx)
	}()
}

// actionQuit will exit the program by using os.Exit, this is
// done because we are using a custom termui EvtStream. Which
// we won't be able to call termui.StopLoop() on. See main.go
// for the customEvtStream and why this is done.
func actionQuit(ctx *context.AppContext) {
	termbox.Close()
	os.Exit(0)
}

func actionInsertMode(ctx *context.AppContext) {
	ctx.Mode = context.InsertMode
	ctx.View.Mode.SetInsertMode()
}

func actionCommandMode(ctx *context.AppContext) {
	ctx.Mode = context.CommandMode
	ctx.View.Mode.SetCommandMode()
}

func actionSearchMode(ctx *context.AppContext) {
	ctx.Mode = context.SearchMode
	ctx.View.Mode.SetSearchMode()
}

func actionGetMessages(ctx *context.AppContext) {
	msgs := ctx.Service.GetMessages(
		ctx.Service.Channels[ctx.View.Channels.SelectedChannel],
		ctx.View.Chat.GetMaxItems(),
	)

	var strMsgs []string
	for _, msg := range msgs {
		strMsgs = append(strMsgs, msg.ToString())
	}

	ctx.View.Chat.SetMessages(strMsgs)

	termui.Render(ctx.View.Chat)
}

// actionMoveCursorUpChannels will execute the actionChangeChannel
// function. A timer is implemented to support fast scrolling through
// the list without executing the actionChangeChannel event
func actionMoveCursorUpChannels(ctx *context.AppContext) {
	go func() {
		if timer != nil {
			timer.Stop()
		}

		ctx.View.Channels.MoveCursorUp()
		termui.Render(ctx.View.Channels)

		timer = time.NewTimer(time.Second / 4)
		<-timer.C

		// Only actually change channel when the timer expires
		actionChangeChannel(ctx)
	}()
}

// actionMoveCursorDownChannels will execute the actionChangeChannel
// function. A timer is implemented to support fast scrolling through
// the list without executing the actionChangeChannel event
func actionMoveCursorDownChannels(ctx *context.AppContext) {
	go func() {
		if timer != nil {
			timer.Stop()
		}

		ctx.View.Channels.MoveCursorDown()
		termui.Render(ctx.View.Channels)

		timer = time.NewTimer(time.Second / 4)
		<-timer.C

		// Only actually change channel when the timer expires
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

func actionSearchNextChannels(ctx *context.AppContext) {
	ctx.View.Channels.SearchNext()
	actionChangeChannel(ctx)
}

func actionSearchPrevChannels(ctx *context.AppContext) {
	ctx.View.Channels.SearchPrev()
	actionChangeChannel(ctx)
}

func actionChangeChannel(ctx *context.AppContext) {
	// Clear messages from Chat pane
	ctx.View.Chat.ClearMessages()

	// Get messages of the SelectedChannel, and get the count of messages
	// that fit into the Chat component
	msgs := ctx.Service.GetMessages(
		ctx.Service.GetSlackChannel(ctx.View.Channels.SelectedChannel),
		ctx.View.Chat.GetMaxItems(),
	)

	var strMsgs []string
	for _, msg := range msgs {
		strMsgs = append(strMsgs, msg.ToString())
	}

	// Set messages for the channel
	ctx.View.Chat.SetMessages(strMsgs)

	// Set channel name for the Chat pane
	ctx.View.Chat.SetBorderLabel(
		ctx.Service.Channels[ctx.View.Channels.SelectedChannel].GetChannelName(),
	)

	// Clear notification icon if there is any
	ctx.Service.MarkAsRead(ctx.View.Channels.SelectedChannel)
	ctx.View.Channels.SetChannels(ctx.Service.ChannelsToString())

	termui.Render(ctx.View.Channels)
	termui.Render(ctx.View.Chat)
}

func actionNewMessage(ctx *context.AppContext, channelID string) {
	ctx.Service.MarkAsUnread(channelID)
	ctx.View.Channels.SetChannels(ctx.Service.ChannelsToString())
	termui.Render(ctx.View.Channels)
	fmt.Print("\a")
}

func actionSetPresence(ctx *context.AppContext, channelID string, presence string) {
	ctx.Service.SetPresenceChannelEvent(channelID, presence)
	ctx.View.Channels.SetChannels(ctx.Service.ChannelsToString())
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

func actionHelp(ctx *context.AppContext) {
	ctx.View.Chat.Help(ctx.Config)
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
