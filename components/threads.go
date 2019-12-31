package components

import (
	"github.com/erroneousboat/termui"
)

type Threads struct {
	*Channels
}

func CreateThreadsComponent(height int) *Threads {
	threads := &Threads{
		Channels: &Channels{
			List: termui.NewList(),
		},
	}

	threads.List.BorderLabel = "Threads"
	threads.List.Height = height

	threads.SelectedChannel = 0
	threads.Offset = 0
	threads.CursorPosition = threads.List.InnerBounds().Min.Y

	return threads
}
