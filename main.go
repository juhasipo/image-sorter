package main

import (
	"flag"
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/library"
	"vincit.fi/image-sorter/ui"
)

func main() {
	flag.Parse()
	broker := event.InitBus(1000)

	root := flag.Arg(0)
	handles := common.LoadImages(root)
	categoryManager := category.New(broker)
	imageLibrary := library.ForHandles(handles, broker)
	gui := ui.Init(broker)

	// Startup
	broker.Subscribe(event.UI_READY, func(message event.Message) {
		categoryManager.RequestCategories()
		imageLibrary.RequestImages()
	})

	// UI -> Library
	broker.Subscribe(event.CATEGORIZE_IMAGE, func(message event.Message) {
		imageLibrary.SetCategory(message.GetData().(*category.CategorizeCommand))
	})
	broker.Subscribe(event.NEXT_IMAGE, func(message event.Message) {
		imageLibrary.RequestNextImage()
	})
	broker.Subscribe(event.PREV_IMAGE, func(message event.Message) {
		imageLibrary.RequestPrevImage()
	})
	broker.Subscribe(event.CURRENT_IMAGE, func(message event.Message) {
		imageLibrary.RequestImages()
	})

	// Library -> UI
	broker.SubscribeGuiEvent(event.IMAGES_UPDATED, func(message event.Message) {
		gui.SetImages(message.GetData().(*library.ImageCommand).GetHandles(), message.GetSubTopic())
	})
	broker.SubscribeGuiEvent(event.CATEGORIES_UPDATED, func(message event.Message) {
		gui.UpdateCategories(message.GetData().(*category.CategoriesCommand).GetCategories())
	})
	broker.SubscribeGuiEvent(event.IMAGE_CATEGORIZED, func(message event.Message) {
		gui.SetImageCategory(message.GetData().(*category.CategorizeCommand))
	})

	gui.Run([]string{})
}

