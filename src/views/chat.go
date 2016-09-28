package views

import (
	"github.com/erroneousboat/slack-term/src/components"
	"github.com/erroneousboat/slack-term/src/service"

	"github.com/gizak/termui"
)

type View struct {
	Input    *components.Input
	Chat     *components.Chat
	Channels *components.Channels
	Mode     *components.Mode
}

func CreateChatView(svc *service.SlackService) *View {
	input := components.CreateInput()

	channels := components.CreateChannels(svc, input.Par.Height)

	chat := components.CreateChat(
		svc, input.Par.Height,
		channels.SlackChannels[channels.SelectedChannel],
	)

	mode := components.CreateMode()

	view := &View{
		Input:    input,
		Channels: channels,
		Chat:     chat,
		Mode:     mode,
	}

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
