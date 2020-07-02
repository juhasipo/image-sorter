package main

import (
	"flag"
	"log"
	"runtime"
	"strings"
	"time"
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/library"
	"vincit.fi/image-sorter/pixbuf"
	"vincit.fi/image-sorter/ui"
)

func main() {
	categories := flag.String("categories", "", "Comma separated categories. Each category in format <name>:<shortcut> e.g. Good:G")

	flag.Parse()

	broker := event.InitBus(1000)

	categoryArr := strings.Split(*categories, ",")
	categoryManager := category.New(broker, categoryArr)

	imageLibrary := library.ForHandles(flag.Arg(0), broker)
	pixbufCache := pixbuf.NewPixbufCache(imageLibrary.GetHandles()[:5])
	gui := ui.Init(broker, pixbufCache)

	// Startup
	broker.Subscribe(event.UI_READY, func() {
		categoryManager.RequestCategories()
		imageLibrary.RequestImages()
	})

	// UI -> Library
	broker.Subscribe(event.CATEGORIZE_IMAGE, imageLibrary.SetCategory)
	broker.Subscribe(event.NEXT_IMAGE, imageLibrary.RequestNextImage)
	broker.Subscribe(event.JUMP_NEXT_IMAGE, imageLibrary.RequestNextImageWithOffset)
	broker.Subscribe(event.PREV_IMAGE, imageLibrary.RequestPrevImage)
	broker.Subscribe(event.JUMP_PREV_IMAGE, imageLibrary.RequestPrevImageWithOffset)
	broker.Subscribe(event.CURRENT_IMAGE, imageLibrary.RequestImages)
	broker.Subscribe(event.JUMP_TO_IMAGE, imageLibrary.RequestImage)
	broker.Subscribe(event.PERSIST_CATEGORIES, imageLibrary.PersistImageCategories)
	broker.Subscribe(event.GENERATE_HASHES, imageLibrary.GenerateHashes)

	// Library -> UI
	broker.ConnectToGui(event.IMAGES_UPDATED, gui.SetImages)
	broker.ConnectToGui(event.CATEGORIES_UPDATED, gui.UpdateCategories)
	broker.ConnectToGui(event.IMAGE_CATEGORIZED, gui.SetImageCategory)
	broker.ConnectToGui(event.UPDATE_HASH_STATUS, gui.UpdateProgress)

	StartBackgroundGC()

	gui.Run()
}

func StartBackgroundGC() {
	log.Print("Start GC background process")
	go func() {
		for true {
			log.Printf("Running GC")
			runtime.GC()
			time.Sleep(30 * time.Second)
		}
	}()
}
