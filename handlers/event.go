package handlers

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/0xAX/notificator"
	"github.com/erroneousboat/termui"
	"github.com/nlopes/slack"
	termbox "github.com/nsf/termbox-go"

	"github.com/erroneousboat/slack-term/components"
	"github.com/erroneousboat/slack-term/config"
	"github.com/erroneousboat/slack-term/context"
	"github.com/erroneousboat/slack-term/views"
)

var scrollTimer *time.Timer
var notifyTimer *time.Timer

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
	"channel-jump":        actionJumpChannels,
	"thread-up":           actionMoveCursorUpThreads,
	"thread-down":         actionMoveCursorDownThreads,
	"chat-up":             actionScrollUpChat,
	"chat-down":           actionScrollDownChat,
	"help":                actionHelp,
}

// Initialize will start a combination of event handlers and 'background tasks'
func Initialize(ctx *context.AppContext) {

	// Keyboard events
	eventHandler(ctx)

	// RTM incoming events
	messageHandler(ctx)

	// User presence
	go actionSetPresenceAll(ctx)
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
			case rtmEvent := <-ctx.Service.RTM.IncomingEvents:
				switch ev := rtmEvent.Data.(type) {
				case *slack.MessageEvent:

					// Construct message
					msg, err := ctx.Service.CreateMessageFromMessageEvent(ev, ev.Channel)
					if err != nil {
						continue
					}

					// Add message to the selected channel
					if ev.Channel == ctx.View.Channels.ChannelItems[ctx.View.Channels.SelectedChannel].ID {

						// Get the thread timestamp of the event, we need to
						// check the previous message as well, because edited
						// message don't have the thread timestamp
						var threadTimestamp string
						if ev.ThreadTimestamp != "" {
							threadTimestamp = ev.ThreadTimestamp
						} else if ev.PreviousMessage != nil && ev.PreviousMessage.ThreadTimestamp != "" {
							threadTimestamp = ev.PreviousMessage.ThreadTimestamp
						} else {
							threadTimestamp = ""
						}

						// When timestamp isn't set this is a thread reply,
						// handle as such
						if threadTimestamp != "" {
							ctx.View.Chat.AddReply(threadTimestamp, msg)
						} else if threadTimestamp == "" && ctx.Focus == context.ChatFocus {
							ctx.View.Chat.AddMessage(msg)
						}

						// we (mis)use actionChangeChannel, to rerender, the
						// view when a new thread has been started
						if ctx.View.Chat.IsNewThread(threadTimestamp) {
							actionChangeChannel(ctx)
						} else {
							termui.Render(ctx.View.Chat)
						}

						// TODO: set Chat.Offset to 0, to automatically scroll
						// down?
					}

					// Set new message indicator for channel, I'm leaving
					// this here because I also want to be notified when
					// I'm currently in a channel but not in the terminal
					// window (tmux). But only create a notification when
					// it comes from someone else but the current user.
					if ev.User != ctx.Service.CurrentUserID {
						actionNewMessage(ctx, ev)
					}
				case *slack.PresenceChangeEvent:
					actionSetPresence(ctx, ev.User, ev.Presence)
				case *slack.RTMError:
					ctx.View.Debug.Println(
						ev.Error(),
					)
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

func actionRedrawGrid(ctx *context.AppContext, threads bool, debug bool) {
	termui.Clear()
	termui.Body = termui.NewGrid()
	termui.Body.X = 0
	termui.Body.Y = 0
	termui.Body.BgColor = termui.ThemeAttr("bg")
	termui.Body.Width = termui.TermWidth()

	columns := []*termui.Row{
		termui.NewCol(ctx.Config.SidebarWidth, 0, ctx.View.Channels),
	}

	if threads && debug {
		columns = append(
			columns,
			[]*termui.Row{
				termui.NewCol(ctx.Config.MainWidth-ctx.Config.ThreadsWidth-3, 0, ctx.View.Chat),
				termui.NewCol(ctx.Config.ThreadsWidth, 0, ctx.View.Threads),
				termui.NewCol(3, 0, ctx.View.Debug),
			}...,
		)
	} else if threads {
		columns = append(
			columns,
			[]*termui.Row{
				termui.NewCol(ctx.Config.MainWidth-ctx.Config.ThreadsWidth, 0, ctx.View.Chat),
				termui.NewCol(ctx.Config.ThreadsWidth, 0, ctx.View.Threads),
			}...,
		)
	} else if debug {
		columns = append(
			columns,
			[]*termui.Row{
				termui.NewCol(ctx.Config.MainWidth-5, 0, ctx.View.Chat),
				termui.NewCol(ctx.Config.MainWidth-6, 0, ctx.View.Debug),
			}...,
		)
	} else {
		columns = append(
			columns,
			[]*termui.Row{
				termui.NewCol(ctx.Config.MainWidth, 0, ctx.View.Chat),
			}...,
		)
	}

	termui.Body.AddRows(
		termui.NewRow(columns...),
		termui.NewRow(
			termui.NewCol(ctx.Config.SidebarWidth, 0, ctx.View.Mode),
			termui.NewCol(ctx.Config.MainWidth, 0, ctx.View.Input),
		),
	)

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
		termui.Render(ctx.View.Input)

		// Send slash command
		isCmd, err := ctx.Service.SendCommand(
			ctx.View.Channels.ChannelItems[ctx.View.Channels.SelectedChannel].ID,
			message,
		)
		if err != nil {
			ctx.View.Debug.Println(
				err.Error(),
			)
		}

		// Send message
		if !isCmd {
			if ctx.Focus == context.ChatFocus {
				err := ctx.Service.SendMessage(
					ctx.View.Channels.ChannelItems[ctx.View.Channels.SelectedChannel].ID,
					message,
				)
				if err != nil {
					ctx.View.Debug.Println(
						err.Error(),
					)
				}

			}

			if ctx.Focus == context.ThreadFocus {
				err := ctx.Service.SendReply(
					ctx.View.Channels.ChannelItems[ctx.View.Channels.SelectedChannel].ID,
					ctx.View.Threads.ChannelItems[ctx.View.Threads.SelectedChannel].ID,
					message,
				)
				if err != nil {
					ctx.View.Debug.Println(
						err.Error(),
					)
				}
			}
		}

		// Clear notification icon if there is any
		channelItem := ctx.View.Channels.ChannelItems[ctx.View.Channels.SelectedChannel]
		if channelItem.Notification {
			ctx.Service.MarkAsRead(channelItem)
			ctx.View.Channels.MarkAsRead(ctx.View.Channels.SelectedChannel)
		}
		termui.Render(ctx.View.Channels)
	}
}

// actionSearch will search through the channels based on the users
// input. A time is implemented to make sure the actual searching
// and changing of channels is done when the user's typing is paused.
func actionSearch(ctx *context.AppContext, key rune) {
	actionInput(ctx.View, key)

	go func() {
		if scrollTimer != nil {
			scrollTimer.Stop()
		}

		scrollTimer = time.NewTimer(time.Second / 4)
		<-scrollTimer.C

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
	msgs, _, err := ctx.Service.GetMessages(
		ctx.View.Channels.ChannelItems[ctx.View.Channels.SelectedChannel].ID,
		ctx.View.Chat.GetMaxItems(),
	)
	if err != nil {
		termbox.Close()
		log.Println(err)
		os.Exit(0)
	}

	ctx.View.Chat.SetMessages(msgs)

	termui.Render(ctx.View.Chat)
}

// actionMoveCursorUpChannels will execute the actionChangeChannel
// function. A timer is implemented to support fast scrolling through
// the list without executing the actionChangeChannel event
func actionMoveCursorUpChannels(ctx *context.AppContext) {
	go func() {
		if scrollTimer != nil {
			scrollTimer.Stop()
		}

		ctx.View.Channels.MoveCursorUp()
		termui.Render(ctx.View.Channels)

		scrollTimer = time.NewTimer(time.Second / 4)
		<-scrollTimer.C

		// Only actually change channel when the timer expires
		actionChangeChannel(ctx)
	}()
}

// actionMoveCursorDownChannels will execute the actionChangeChannel
// function. A timer is implemented to support fast scrolling through
// the list without executing the actionChangeChannel event
func actionMoveCursorDownChannels(ctx *context.AppContext) {
	go func() {
		if scrollTimer != nil {
			scrollTimer.Stop()
		}

		ctx.View.Channels.MoveCursorDown()
		termui.Render(ctx.View.Channels)

		scrollTimer = time.NewTimer(time.Second / 4)
		<-scrollTimer.C

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

func actionJumpChannels(ctx *context.AppContext) {
	ctx.View.Channels.Jump()
	actionChangeChannel(ctx)
}

func actionChangeChannel(ctx *context.AppContext) {
	// Clear messages from Chat pane
	ctx.View.Chat.ClearMessages()

	// Get messages of the SelectedChannel, and get the count of messages
	// that fit into the Chat component
	msgs, threads, err := ctx.Service.GetMessages(
		ctx.View.Channels.ChannelItems[ctx.View.Channels.SelectedChannel].ID,
		ctx.View.Chat.GetMaxItems(),
	)
	if err != nil {
		termbox.Close()
		log.Println(err)
		os.Exit(0)
	}

	// Set messages for the channel
	ctx.View.Chat.SetMessages(msgs)

	// Set the threads identifiers in the threads pane
	var haveThreads bool
	if len(threads) > 0 {
		haveThreads = true

		// Make the first thread the current Channel
		ctx.View.Threads.SetChannels(
			append(
				[]components.ChannelItem{ctx.View.Channels.GetSelectedChannel()},
				threads...,
			),
		)

		// Reset position of SelectedChannel
		ctx.View.Threads.MoveCursorTop()
	}

	// Set channel name for the Chat pane
	ctx.View.Chat.SetBorderLabel(
		ctx.View.Channels.ChannelItems[ctx.View.Channels.SelectedChannel].GetChannelName(),
	)

	// Clear notification icon if there is any
	channelItem := ctx.View.Channels.ChannelItems[ctx.View.Channels.SelectedChannel]
	if channelItem.Notification {
		ctx.Service.MarkAsRead(channelItem)
		ctx.View.Channels.MarkAsRead(ctx.View.Channels.SelectedChannel)
	}

	// Redraw grid, necessary when threads and/or debug is set. We will redraw
	// the grid when there are threads, or we just came from a thread and went
	// to a channel without threads. Hence the clearing of ChannelItems of
	// Threads.
	if haveThreads {
		actionRedrawGrid(ctx, haveThreads, ctx.Debug)
	} else if !haveThreads && len(ctx.View.Threads.ChannelItems) > 0 {
		ctx.View.Threads.SetChannels([]components.ChannelItem{})
		actionRedrawGrid(ctx, haveThreads, ctx.Debug)
	} else {
		termui.Render(ctx.View.Threads)
		termui.Render(ctx.View.Channels)
		termui.Render(ctx.View.Chat)
	}

	// Set focus, necessary to know when replying to thread or chat
	ctx.Focus = context.ChatFocus
}

func actionChangeThread(ctx *context.AppContext) {
	// Clear messages from Chat pane
	ctx.View.Chat.ClearMessages()

	// The first channel in the Thread list is current Channel. Set context
	// Focus and messages accordingly.
	var err error
	msgs := []components.Message{}
	if ctx.View.Threads.SelectedChannel == 0 {
		ctx.Focus = context.ChatFocus

		msgs, _, err = ctx.Service.GetMessages(
			ctx.View.Channels.ChannelItems[ctx.View.Channels.SelectedChannel].ID,
			ctx.View.Chat.GetMaxItems(),
		)
		if err != nil {
			termbox.Close()
			log.Println(err)
			os.Exit(0)
		}
	} else {
		ctx.Focus = context.ThreadFocus

		msgs, err = ctx.Service.GetMessageByID(
			ctx.View.Threads.ChannelItems[ctx.View.Threads.SelectedChannel].ID,
			ctx.View.Channels.ChannelItems[ctx.View.Channels.SelectedChannel].ID,
		)
		if err != nil {
			termbox.Close()
			log.Println(err)
			os.Exit(0)
		}
	}

	// Set messages for the channel
	ctx.View.Chat.SetMessages(msgs)

	termui.Render(ctx.View.Channels)
	termui.Render(ctx.View.Threads)
	termui.Render(ctx.View.Chat)
}

func actionMoveCursorUpThreads(ctx *context.AppContext) {
	go func() {
		if scrollTimer != nil {
			scrollTimer.Stop()
		}

		ctx.View.Threads.MoveCursorUp()
		termui.Render(ctx.View.Threads)

		scrollTimer = time.NewTimer(time.Second / 4)
		<-scrollTimer.C

		// Only actually change channel when the timer expires
		actionChangeThread(ctx)
	}()
}

func actionMoveCursorDownThreads(ctx *context.AppContext) {
	go func() {
		if scrollTimer != nil {
			scrollTimer.Stop()
		}

		ctx.View.Threads.MoveCursorDown()
		termui.Render(ctx.View.Threads)

		scrollTimer = time.NewTimer(time.Second / 4)
		<-scrollTimer.C

		// Only actually change thread when the timer expires
		actionChangeThread(ctx)
	}()
}

// actionNewMessage will set the new message indicator for a channel, and
// if configured will also display a desktop notification
func actionNewMessage(ctx *context.AppContext, ev *slack.MessageEvent) {
	ctx.View.Channels.MarkAsUnread(ev.Channel)
	termui.Render(ctx.View.Channels)

	// Terminal bell
	fmt.Print("\a")

	// Desktop notification
	if ctx.Config.Notify == config.NotifyMention {
		if isMention(ctx, ev) {
			createNotifyMessage(ctx, ev)
		}
	} else if ctx.Config.Notify == config.NotifyAll {
		createNotifyMessage(ctx, ev)
	}
}

func actionSetPresence(ctx *context.AppContext, channelID string, presence string) {
	ctx.View.Channels.SetPresence(channelID, presence)
	termui.Render(ctx.View.Channels)
}

// actionPresenceAll will set the presence of the user list. Because the
// requests to the endpoint are rate limited we implement a timeout here.
func actionSetPresenceAll(ctx *context.AppContext) {
	for _, chn := range ctx.Service.Conversations {
		if chn.IsIM {

			presence, err := ctx.Service.GetUserPresence(chn.User)
			if err != nil {
				presence = "away"
			}
			ctx.View.Channels.SetPresence(chn.ID, presence)

			termui.Render(ctx.View.Channels)
			time.Sleep(1200 * time.Millisecond)
		}
	}
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
	ctx.View.Chat.ClearMessages()
	ctx.View.Chat.Help(ctx.Usage, ctx.Config)
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

// isMention check if the message event either contains a
// mention or is posted on an IM channel.
func isMention(ctx *context.AppContext, ev *slack.MessageEvent) bool {
	channel := ctx.View.Channels.ChannelItems[ctx.View.Channels.FindChannel(ev.Channel)]

	if channel.Type == components.ChannelTypeIM {
		return true
	}

	// Mentions have the following format:
	//	<@U12345|erroneousboat>
	// 	<@U12345>
	r := regexp.MustCompile(`\<@(\w+\|*\w+)\>`)
	matches := r.FindAllString(ev.Text, -1)
	for _, match := range matches {
		if strings.Contains(match, ctx.Service.CurrentUserID) {
			return true
		}
	}

	return false
}

func createNotifyMessage(ctx *context.AppContext, ev *slack.MessageEvent) {
	go func() {
		if notifyTimer != nil {
			notifyTimer.Stop()
		}

		// Only actually notify when time expires
		notifyTimer = time.NewTimer(time.Second * 2)
		<-notifyTimer.C

		var message string
		channel := ctx.View.Channels.ChannelItems[ctx.View.Channels.FindChannel(ev.Channel)]
		switch channel.Type {
		case components.ChannelTypeChannel:
			message = fmt.Sprintf("Message received on channel: %s", channel.Name)
		case components.ChannelTypeGroup:
			message = fmt.Sprintf("Message received in group: %s", channel.Name)
		case components.ChannelTypeIM:
			message = fmt.Sprintf("Message received from: %s", channel.Name)
		default:
			message = fmt.Sprintf("Message received from: %s", channel.Name)
		}

		ctx.Notify.Push("slack-term", message, "", notificator.UR_NORMAL)
	}()
}
