package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os/user"
	"path"

	"github.com/erroneousboat/slack-term/context"
	"github.com/erroneousboat/slack-term/handlers"
)

const (
	VERSION = "v2.0.0"
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
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// Create context
	appCTX, err := context.CreateAppContext(flgConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer appCTX.View.GUI.Close()

	// Register handlers
	handlers.RegisterEventHandlers(appCTX)
}
