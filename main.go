package main

import (
	"flag"
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/library"
	"vincit.fi/image-sorter/ui"
)

func main() {
	flag.Parse()
	broker := event.InitBus(1000)

	flag.Parse()
	root := flag.Arg(0)
	categoryManager := category.New(broker)
	imageLibrary := library.ForHandles(root, broker)
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
	broker.Subscribe(event.PERSIST_CATEGORIES, imageLibrary.PersistImageCategories)
	broker.Subscribe(event.GENERATE_HASHES, imageLibrary.GenerateHashes)

	// Library -> UI
	broker.ConnectToGui(event.IMAGES_UPDATED, gui.SetImages)
	broker.ConnectToGui(event.CATEGORIES_UPDATED, gui.UpdateCategories)
	broker.ConnectToGui(event.IMAGE_CATEGORIZED, gui.SetImageCategory)
	broker.ConnectToGui(event.UPDATE_HASH_STATUS, gui.UpdateProgress)

	gui.Run([]string{})
}

