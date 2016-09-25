package main

import (
	"flag"
	"log"
	"os/user"
	"path"

	"github.com/erroneousboat/slack-term/src/context"
	"github.com/erroneousboat/slack-term/src/handlers"

	"github.com/gizak/termui"
)

func main() {
	// Start terminal user interface
	err := termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()

	// Get home dir for config file default
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	// Parse flags
	flgConfig := flag.String(
		"config",
		path.Join(usr.HomeDir, "slack-term.json"),
		"location of config file",
	)
	flag.Parse()

	// Create context
	ctx := context.CreateAppContext(*flgConfig)

	// Setup body
	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(1, 0, ctx.View.Channels),
			termui.NewCol(11, 0, ctx.View.Chat),
		),
		termui.NewRow(
			termui.NewCol(1, 0, ctx.View.Mode),
			termui.NewCol(11, 0, ctx.View.Input),
		),
	)
	termui.Body.Align()
	termui.Render(termui.Body)

	// Set body in context
	ctx.Body = termui.Body

	// Register handlers
	handlers.RegisterEventHandlers(ctx)

	termui.Loop()
}
