package components

import "github.com/gizak/termui"

func CreateInputComponent() *termui.Par {
	compInput := termui.NewPar("")
	compInput.Height = 3
	return compInput
}

func CreateChannelsComponent(inputHeight int) *termui.List {
	channels := []string{
		"general",
		"random",
	}

	compChannels := termui.NewList()
	compChannels.Items = channels
	compChannels.BorderLabel = "Channels"
	compChannels.Height = termui.TermHeight() - inputHeight
	compChannels.Overflow = "wrap"

	return compChannels
}

func CreateChatComponent(inputHeight int) *termui.List {
	messages := []string{
		"[jp] hello world",
		"[erroneousboat] foo bar",
	}

	compChat := termui.NewList()
	compChat.Items = messages
	compChat.BorderLabel = "Channel01"
	compChat.Height = termui.TermHeight() - inputHeight
	compChat.Overflow = "wrap"

	return compChat
}

func CreateModeComponent() *termui.Par {
	compMode := termui.NewPar("NORMAL")
	compMode.Height = 3
	return compMode
}
