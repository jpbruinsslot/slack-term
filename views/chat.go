package views

import (
	"log"

	"github.com/erroneousboat/gocui"
	"github.com/gizak/termui"

	"github.com/erroneousboat/slack-term/components"
	"github.com/erroneousboat/slack-term/service"
)

type View struct {
	Input    *components.Input
	Chat     *components.Chat
	Channels *components.Channels
	Mode     *components.Mode
	Debug    *components.Debug
	GUI      *gocui.Gui
}

type ViewBKP struct {
	Input    *components.Input
	Chat     *components.ChatBKP
	Channels *components.ChannelsBKP
	Mode     *components.Mode
	Debug    *components.Debug
	GUI      *gocui.Gui
}

func CreateChatView(svc *service.SlackService) *View {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Fatal(err)
	}

	g.Cursor = true

	view := &View{
		GUI: g,
	}

	maxX, maxY := g.Size()

	// Create Channels component
	channels := components.CreateChannelsComponent(0, 0, 10, maxY-1)
	view.Channels = channels

	// Fill Channels component
	slackChans := svc.GetChannels()
	channels.SetChannels(slackChans)
	channels.SetPresenceChannels(slackChans)

	// TODO Input component

	// TODO Mode component

	// Chat component
	chat := components.CreateChatComponent(11, 0, maxX-12, maxY-1)
	view.Chat = chat

	// Fill Chat component
	slackMsgs := svc.GetMessages(svc.GetSlackChannel(channels.SelectedChannel), 10)
	chat.SetMessages(slackMsgs)

	// Create Debug component
	debug := components.CreateDebugComponent(maxX-51, 0, 50, 10)
	view.Debug = debug

	// Render the components
	g.SetManager(channels, chat, debug)

	// Initialize keybindings
	// initKeyBindings(view)

	view.GUI.Flush()

	return view
}

func (v *View) RefreshComponent(name string) {
	v.GUI.Update(
		func(g *gocui.Gui) error {
			_, err := g.View(name)
			if err != nil {
				return (err)
			}
			return nil
		},
	)
}

// func initKeyBindings(view *View) {
// 	if err := view.GUI.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
// 		log.Fatal(err)
// 	}
// 	if err := view.GUI.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, view.Channels.MoveCursorDown); err != nil {
// 		log.Fatal(err)
// 	}
// 	if err := view.GUI.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, view.Channels.MoveCursorUp); err != nil {
// 		log.Fatal(err)
// 	}
// }

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func CreateChatViewBKP(svc *service.SlackService) *ViewBKP {
	input := components.CreateInput()

	channels := components.CreateChannels(svc, input.Par.Height)

	chat := components.CreateChat(
		svc,
		input.Par.Height,
		svc.SlackChannels[channels.SelectedChannel],
		svc.Channels[channels.SelectedChannel],
	)

	mode := components.CreateMode()

	view := &ViewBKP{
		Input:    input,
		Channels: channels,
		Chat:     chat,
		Mode:     mode,
	}

	return view
}

func (v *ViewBKP) RefreshBKP() {
	termui.Render(
		v.Input,
		v.Chat,
		v.Channels,
		v.Mode,
	)
}
