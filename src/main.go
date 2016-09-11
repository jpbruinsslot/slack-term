package main

import (
	"github.com/erroneousboat/slack-term/src/context"
	"github.com/erroneousboat/slack-term/src/handlers"

	"github.com/gizak/termui"
)

func main() {
	err := termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()

	// create context
	ctx := context.CreateAppContext()

	// setup view
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

	ctx.Body = termui.Body

	// register handlers
	handlers.RegisterEventHandlers(ctx)

	termui.Loop()
}
