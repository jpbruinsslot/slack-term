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
	"github.com/erroneousboat/slack-term/token"
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
    https://github.com/jpbruinsslot/slack-term

GLOBAL OPTIONS:
   -config [path-to-config-file]
   -token [slack-token]
   -token-url
   -debug
   -help, -h
`
)

var (
	flgConfig   string
	flgToken    string
	flgTokenURL bool
	flgDebug    bool
	flgUsage    bool
)

func init() {

	// Find the default config file
	configFile := xdg.New("slack-term", "").QueryConfig("config")

	// Parse flags
	flag.StringVar(
		&flgConfig,
		"config",
		configFile,
		"location of config file",
	)

	flag.StringVar(
		&flgToken,
		"token",
		"",
		"the slack token",
	)

	flag.BoolVar(
		&flgTokenURL,
		"token-url",
		false,
		"print OAuth authorization URL to obtain a token",
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
	if flgTokenURL {
		fmt.Println(token.Method1AuthorizeURL())
		return
	}

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
		flgConfig, flgToken, flgDebug, VERSION, usage,
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
