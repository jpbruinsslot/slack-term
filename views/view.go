package views

import (
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
	Threads  *components.Threads
	Mode     *components.Mode
	Debug    *components.Debug
}

func CreateView(config *config.Config, svc *service.SlackService) (*View, error) {
	// Create Input component
	input := components.CreateInputComponent()

	// Channels: create the component
	sideBarHeight := termui.TermHeight() - input.Par.Height
	channels := components.CreateChannelsComponent(sideBarHeight, config.UnreadOnly)

	// Channels: fill the component
	slackChans, err := svc.GetChannels()
	if err != nil {
		return nil, err
	}

	// Channels: set channels in component
	channels.SetChannels(slackChans)

	// Threads: create component
	threads := components.CreateThreadsComponent(sideBarHeight)

	// Chat: create the component
	chat := components.CreateChatComponent(input.Par.Height)

	// Chat: fill the component
	if chn, ok := channels.GetSelectedChannel(); ok {
		msgs, thr, err := svc.GetMessages(
			chn.ID,
			chat.GetMaxItems(),
		)
		if err != nil {
			return nil, err
		}

		// Chat: set messages in component
		chat.SetMessages(msgs)
		chat.SetBorderLabel(chn.GetChannelName())

		// Threads: set threads in component
		if len(thr) > 0 {

			// Make the first thread the current Channel
			threads.SetChannels(
				append(
					[]*components.ChannelItem{chn},
					thr...,
				),
			)
		}
	}

	// Debug: create the component
	debug := components.CreateDebugComponent(input.Par.Height)

	// Mode: create the component
	mode := components.CreateModeComponent()

	view := &View{
		Config:   config,
		Input:    input,
		Channels: channels,
		Threads:  threads,
		Chat:     chat,
		Mode:     mode,
		Debug:    debug,
	}

	return view, nil
}

func (v *View) Refresh() {
	termui.Render(
		v.Input,
		v.Chat,
		v.Channels,
		v.Threads,
		v.Mode,
	)
}
