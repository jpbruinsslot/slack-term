package views

import (
	"time"

	"github.com/bharath-srinivas/termloader"
	"github.com/erroneousboat/termui"

	"github.com/erroneousboat/slack-term/components"
	"github.com/erroneousboat/slack-term/config"
	"github.com/erroneousboat/slack-term/service"
)

type View struct {
	Config   *config.Config
	Input    *components.Input
	Chat     *components.Chat
	Channels *components.Channels
	Mode     *components.Mode
	Debug    *components.Debug
}

func CreateView(config *config.Config, svc *service.SlackService) *View {
	loader := termloader.New(termloader.Charsets[0], 100*time.Millisecond)
	loader.Text = "Loading"
	loader.Color = termloader.Green
	loader.Start()

	// Create Input component
	input := components.CreateInputComponent()

	// Channels: create the component
	channels := components.CreateChannelsComponent(input.Par.Height)

	// Channels: fill the component
	slackChans := svc.GetChannels()
	channels.SetChannels(slackChans)

	// Chat: create the component
	chat := components.CreateChatComponent(input.Par.Height)

	// Chat: fill the component
	msgs := svc.GetMessages(
		svc.GetSlackChannel(channels.SelectedChannel),
		chat.GetMaxItems(),
	)

	var strMsgs []string
	for _, msg := range msgs {
		strMsgs = append(strMsgs, msg.ToString())
	}

	chat.SetMessages(strMsgs)
	chat.SetBorderLabel(svc.Channels[channels.SelectedChannel].GetChannelName())

	// Debug: create the component
	debug := components.CreateDebugComponent(input.Par.Height)

	// Mode: create the component
	mode := components.CreateModeComponent()

	view := &View{
		Config:   config,
		Input:    input,
		Channels: channels,
		Chat:     chat,
		Mode:     mode,
		Debug:    debug,
	}

	loader.Stop()
	return view
}

func (v *View) Refresh() {
	termui.Render(
		v.Input,
		v.Chat,
		v.Channels,
		v.Mode,
	)
}
