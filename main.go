package main

import (
	"flag"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/library"
	"vincit.fi/image-sorter/ui"
)

func main() {
	flag.Parse()
	root := flag.Arg(0)
	broker := event.InitBus(1000)
	handles := common.LoadImages(root)
	library := library.ForHandles(handles)
	gui := ui.Init(library, broker)

	broker.Subscribe(event.NEXT_IMAGE, func(message event.Message) {
		library.NextImage()
		gui.UpdateImages()
	})
	broker.Subscribe(event.PREV_IMAGE, func(message event.Message) {
		library.PrevImage()
		gui.UpdateImages()
	})

	gui.Run([]string{})
}

