package main

import (
	"flag"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/image"
	"vincit.fi/image-sorter/ui"
)

func main() {
	flag.Parse()
	root := flag.Arg(0)
	broker := event.InitBus(1000)
	manager := image.ManagerForDir(root)
	gui := ui.Init(&manager, broker)

	broker.Subscribe(event.NEXT_IMAGE, func(message event.Message) {
		manager.NextImage()
		gui.UpdateImages()
	})
	broker.Subscribe(event.PREV_IMAGE, func(message event.Message) {
		manager.PrevImage()
		gui.UpdateImages()
	})

	gui.Run([]string{})
}

