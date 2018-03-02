package main

import (
	"../../notificator"
	"log"
)

var notify *notificator.Notificator

func main() {
	notify = notificator.New(notificator.Options{
		DefaultIcon: "icon/default.png",
		AppName:     "My test App",
	})

	notify.Push("title", "text", "/home/user/icon.png", notificator.UR_NORMAL)

	// Check errors
	err := notify.Push("error", "ops =(", "/home/user/icon.png", notificator.UR_CRITICAL)

	if err != nil {
		log.Fatal(err)
	}
}
