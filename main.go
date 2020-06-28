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
	//categories := category.FromCategories([]string{"Good", "Maybe", "Bad"})
	imageLibrary := library.ForHandles(handles, broker)
	gui := ui.Init(broker)

	broker.Subscribe(event.NEXT_IMAGE, func(message event.Message) {
		imageLibrary.NextImage()
	})
	broker.Subscribe(event.PREV_IMAGE, func(message event.Message) {
		imageLibrary.PrevImage()
	})
	broker.Subscribe(event.CURRENT_IMAGE, func(message event.Message) {
		imageLibrary.GetCurrentImage()
	})
	broker.SubscribeGuiEvent(event.IMAGES_UPDATED, func(message event.Message) {
		gui.SetImages(message.GetData().([]*common.Handle), message.GetSubTopic())
	})

	gui.Run([]string{})
}

