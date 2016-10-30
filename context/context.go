package context

import (
	"log"

	"github.com/gizak/termui"

	"../config"
	"../service"
	"../views"
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

// CreateAppContext creates an application context which can be passed
// and referenced througout the application
func CreateAppContext(flgConfig string) *AppContext {
	// Load config
	cfg, err := config.NewConfig(flgConfig)
	if err != nil {
		log.Fatalf("ERROR: not able to load config file (%s): %s", flgConfig, err)
	}

	// Create Service
	svc := service.NewSlackService(cfg.SlackToken)

	// Create ChatView
	view := views.CreateChatView(svc)

	return &AppContext{
		Service: svc,
		View:    view,
		Config:  cfg,
		Mode:    CommandMode,
	}
}
