package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/OpenPeeDeeP/xdg"
	"github.com/erroneousboat/termui"
	termbox "github.com/nsf/termbox-go"

	"github.com/erroneousboat/slack-term/context"
	"github.com/erroneousboat/slack-term/handlers"
)

const (
	VERSION = "master"
	USAGE   = `NAME:
    slack-term - slack client for your terminal

USAGE:
    slack-term -config [path-to-config]

VERSION:
    %s

WEBSITE:
    https://github.com/erroneousboat/slack-term

GLOBAL OPTIONS:
   -config [path-to-config-file]
   -workspace [slack-workspace]
   -debug
   -help, -h
`
)

var (
	flgConfig    string
	flgWorkspace string
	flgDebug     bool
	flgUsage     bool
)

func init() {

	// Find the default config file
	configFile := xdg.New("slack-term", "").QueryConfig("config")

	// Parse flags.

	// Path to the config file to use.
	// Defaults to ~/.slack-term
	flag.StringVar(
		&flgConfig,
		"config",
		configFile,
		"location of config file",
	)

	// The name of the workspace to use.
	// If none is provided, will pick the alphabetical first (in other words,
	// will pick the only workspace in a list of only one workspace).
	flag.StringVar(
		&flgWorkspace,
		"workspace",
		"",
		"the slack workspace to use",
	)

	// Toggle debug mode.
	flag.BoolVar(
		&flgDebug,
		"debug",
		false,
		"turn on debugging",
	)

	flag.Usage = func() {
		fmt.Printf(USAGE, VERSION)
	}

	flag.Parse()
}

func main() {
	// Start terminal user interface
	err := termui.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer termui.Close()

	// Create custom event stream for termui because
	// termui's one has data race conditions with its
	// event handling. We're circumventing it here until
	// it has been fixed.
	customEvtStream := &termui.EvtStream{
		Handlers: make(map[string]func(termui.Event)),
	}
	termui.DefaultEvtStream = customEvtStream

	// Create context
	usage := fmt.Sprintf(USAGE, VERSION)
	ctx, err := context.CreateAppContext(
		flgConfig, flgWorkspace, flgDebug, VERSION, usage,
	)
	if err != nil {
		termbox.Close()
		log.Println(err)
		os.Exit(0)
	}

	// Initialize handlers
	handlers.Initialize(ctx)

	termui.Loop()
}
