package main

import (
	"flag"
	"fmt"
	"log"
	"os/user"
	"path"

	"github.com/jroimartin/gocui"

	"github.com/erroneousboat/slack-term/context"
)

const (
	VERSION = "v0.2.3"
	USAGE   = `NAME:
    slack-term - slack client for your terminal

USAGE:
    slack-term -config [path-to-config]

VERSION:
    %s

GLOBAL OPTIONS:
   --help, -h
`
)

var (
	flgConfig string
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
		path.Join(usr.HomeDir, "slack-term.json"),
		"location of config file",
	)

	flag.Usage = func() {
		fmt.Printf(USAGE, VERSION)
	}

	flag.Parse()
}

func main() {
	// Create context
	appCTX, err := context.CreateAppContext(flgConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer appCTX.View.GUI.Close()

	// Create the view
	// view, err := views.CreateView(ctx)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer view.Close()

	// Register handlers
	// handlers.RegisterEventHandlers(app)

	if err := appCTX.View.GUI.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Fatal(err)
	}
}
