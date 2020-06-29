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
	broker.Subscribe(event.UI_READY, func() {
		categoryManager.RequestCategories()
		imageLibrary.RequestImages()
	})

	// UI -> Library
	broker.Subscribe(event.CATEGORIZE_IMAGE, imageLibrary.SetCategory)
	broker.Subscribe(event.NEXT_IMAGE, imageLibrary.RequestNextImage)
	broker.Subscribe(event.PREV_IMAGE, imageLibrary.RequestPrevImage)
	broker.Subscribe(event.CURRENT_IMAGE, imageLibrary.RequestImages)

	// Library -> UI
	broker.ConnectToGui(event.IMAGES_UPDATED, gui.SetImages)
	broker.ConnectToGui(event.CATEGORIES_UPDATED, gui.UpdateCategories)
	broker.ConnectToGui(event.IMAGE_CATEGORIZED, gui.SetImageCategory)

	gui.Run([]string{})
}

