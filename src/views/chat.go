package views

import (
	"github.com/erroneousboat/slack-term/src/components"

	"github.com/gizak/termui"
)

type View struct {
	Input    *components.Input
	Chat     *components.Chat
	Channels *termui.List
	Mode     *termui.Par
}

func CreateChatView() *View {
	input := components.CreateInput()
	channels := components.CreateChannelsComponent(input.Par.Height)
	chat := components.CreateChat(input.Par.Height)
	mode := components.CreateModeComponent()

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
