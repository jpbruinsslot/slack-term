package views

import (
	"github.com/erroneousboat/slack-term/src/components"

	"github.com/gizak/termui"
)

type View struct {
	Input    *components.Input
	Chat     *termui.List
	Channels *termui.List
	Mode     *termui.Par
}

func CreateChatView() *View {
	input := components.CreateInput()
	channels := components.CreateChannelsComponent(input.Block.Height)
	chat := components.CreateChatComponent(input.Block.Height)
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
