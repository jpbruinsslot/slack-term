package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"

	"github.com/erroneousboat/termui"
	termbox "github.com/nsf/termbox-go"

	"github.com/oliverseal/slack-term/context"
	"github.com/oliverseal/slack-term/handlers"
)

const (
	VERSION = "v0.4.0"
	USAGE   = `NAME:
    slack-term - slack client for your terminal

USAGE:
    slack-term -config [path-to-config]

VERSION:
    %s

WEBSITE:
    https://github.com/oliverseal/slack-term

GLOBAL OPTIONS:
   --help, -h
`
)

var (
	flgConfig string
	flgToken  string
	flgDebug  bool
	flgUsage  bool
)

func init() {
	// Get home dir for config file default
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	// Parse flags
	flag.StringVar(
		&flgConfig,
		"config",
		path.Join(usr.HomeDir, ".slack-term"),
		"location of config file",
	)

	flag.StringVar(
		&flgToken,
		"token",
		"",
		"the slack token",
	)

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
	ctx, err := context.CreateAppContext(flgConfig, flgToken, flgDebug)
	if err != nil {
		termbox.Close()
		log.Println(err)
		os.Exit(0)
	}

	// Register handlers
	handlers.RegisterEventHandlers(ctx)

	termui.Loop()
}
