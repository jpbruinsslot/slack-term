package context

import (
	"errors"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/0xAX/notificator"
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
	Version    string
	Usage      string
	EventQueue chan termbox.Event
	Service    *service.SlackService
	Body       *termui.Grid
	View       *views.View
	Config     *config.Config
	Debug      bool
	Mode       string
	Notify     *notificator.Notificator
}

// CreateAppContext creates an application context which can be passed
// and referenced througout the application
func CreateAppContext(flgConfig string, flgToken string, flgDebug bool, version string, usage string) (*AppContext, error) {
	if flgDebug {
		go func() {
			http.ListenAndServe(":6060", nil)
		}()
	}

	// Loading screen
	views.Loading()

	// Load config
	config, err := config.NewConfig(flgConfig)
	if err != nil {
		return nil, err
	}

	// When slack token isn't set in the config file, we'll check
	// the command-line flag or the environment variable
	if config.SlackToken == "" {
		if flgToken != "" {
			config.SlackToken = flgToken
		} else {
			config.SlackToken = os.Getenv("SLACK_TOKEN")
		}
	}

	// Create desktop notifier
	var notify *notificator.Notificator
	if config.Notify != "" {
		notify = notificator.New(notificator.Options{AppName: "slack-term"})
		if notify == nil {
			return nil, errors.New(
				"desktop notifications are not supported for your OS",
			)
		}
	}

	// Create Service
	svc, err := service.NewSlackService(config)
	if err != nil {
		return nil, err
	}

	// Create the main view
	view, err := views.CreateView(config, svc)
	if err != nil {
		return nil, err
	}

	// Setup the interface
	if flgDebug {
		// FIXME: threads width configurable
		termui.Body.AddRows(
			termui.NewRow(
				termui.NewCol(config.SidebarWidth, 0, view.Channels),
				termui.NewCol(config.MainWidth-5, 0, view.Chat),
				termui.NewCol(config.MainWidth-10, 0, view.Threads),
				termui.NewCol(config.MainWidth-7, 0, view.Debug),
			),
			termui.NewRow(
				termui.NewCol(config.SidebarWidth, 0, view.Mode),
				termui.NewCol(config.MainWidth, 0, view.Input),
			),
		)
	} else {
		// FIXME: threads width configurable
		termui.Body.AddRows(
			termui.NewRow(
				termui.NewCol(config.SidebarWidth, 0, view.Channels),
				termui.NewCol(config.MainWidth-1, 0, view.Chat),
				termui.NewCol(config.MainWidth-10, 0, view.Threads),
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
		Version:    version,
		Usage:      usage,
		EventQueue: make(chan termbox.Event, 20),
		Service:    svc,
		Body:       termui.Body,
		View:       view,
		Config:     config,
		Debug:      flgDebug,
		Mode:       CommandMode,
		Notify:     notify,
	}, nil
}
