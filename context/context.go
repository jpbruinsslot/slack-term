package context

import (
	"github.com/erroneousboat/termui"
	termbox "github.com/nsf/termbox-go"

	"github.com/erroneousboat/slack-term/config"
	"github.com/erroneousboat/slack-term/service"
	"github.com/erroneousboat/slack-term/views"
)

const (
	CommandMode = "command"
	InsertMode  = "insert"
	SearchMode  = "search"
)

type AppContext struct {
	EventQueue chan termbox.Event
	Service    *service.SlackService
	Body       *termui.Grid
	View       *views.View
	Config     *config.Config
	Debug      bool
	Mode       string
}

// CreateAppContext creates an application context which can be passed
// and referenced througout the application
func CreateAppContext(flgConfig string, flgDebug bool) (*AppContext, error) {
	// Load config
	config, err := config.NewConfig(flgConfig)
	if err != nil {
		return nil, err
	}

	// Create Service
	svc, err := service.NewSlackService(config.SlackToken)
	if err != nil {
		return nil, err
	}

	// Create the main view
	view := views.CreateView(svc)

	// Setup the interface
	if flgDebug {
		termui.Body.AddRows(
			termui.NewRow(
				termui.NewCol(config.SidebarWidth, 0, view.Channels),
				termui.NewCol(config.MainWidth, 0, view.Chat),
			),
			termui.NewRow(
				termui.NewCol(config.SidebarWidth, 0, view.Mode),
				termui.NewCol(config.MainWidth-1, 0, view.Input),
				termui.NewCol(1, 0, view.Debug),
			),
		)
	} else {
		termui.Body.AddRows(
			termui.NewRow(
				termui.NewCol(config.SidebarWidth, 0, view.Channels),
				termui.NewCol(config.MainWidth, 0, view.Chat),
			),
			termui.NewRow(
				termui.NewCol(config.SidebarWidth, 0, view.Mode),
				termui.NewCol(config.MainWidth, 0, view.Input),
			),
		)
	}

	termui.Body.Align()
	termui.Render(termui.Body)

	return &AppContext{
		EventQueue: make(chan termbox.Event, 20),
		Service:    svc,
		Body:       termui.Body,
		View:       view,
		Config:     config,
		Debug:      flgDebug,
		Mode:       CommandMode,
	}, nil
}
