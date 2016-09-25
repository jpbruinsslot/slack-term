package context

import (
	"log"

	"github.com/gizak/termui"

	"github.com/erroneousboat/slack-term/src/config"
	"github.com/erroneousboat/slack-term/src/service"
	"github.com/erroneousboat/slack-term/src/views"
)

const (
	CommandMode = "command"
	InsertMode  = "insert"
)

type AppContext struct {
	Service *service.SlackService
	Body    *termui.Grid
	View    *views.View
	Config  *config.Config
	Mode    string
}

func CreateAppContext(flgConfig string) *AppContext {
	// Load config
	config, err := config.NewConfig(flgConfig)
	if err != nil {
		log.Fatalf("ERROR: not able to load config file: %s", flgConfig)
	}

	// Create Service
	svc := service.NewSlackService(config.SlackToken)

	// Create ChatView
	view := views.CreateChatView(svc)

	return &AppContext{
		Service: svc,
		View:    view,
		Config:  config,
		Mode:    CommandMode,
	}
}
