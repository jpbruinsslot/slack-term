package context

import (
	"github.com/erroneousboat/slack-term/src/views"
	"github.com/gizak/termui"
)

const (
	CommandMode = "command"
	InsertMode  = "insert"
)

type AppContext struct {
	Body *termui.Grid
	View *views.View
	Mode string
}

// TODO: arg Config
func CreateAppContext() *AppContext {
	view := views.CreateChatView()

	return &AppContext{
		View: view,
		Mode: CommandMode,
	}
}
