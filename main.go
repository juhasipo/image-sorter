package main

import (
	"flag"
	"github.com/google/uuid"
	"log"
	"runtime"
	"strings"
	"time"
	"vincit.fi/image-sorter/caster"
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/imagecategory"
	"vincit.fi/image-sorter/library"
	"vincit.fi/image-sorter/pixbuf"
	"vincit.fi/image-sorter/ui"
)

func main() {
	categories := flag.String("categories", "", "Comma separated categories. Each category in format <name>:<shortcut> e.g. Good:G")
	httpPort := flag.Int("httpPort", 8080, "HTTP Server port for Chrome Cast")

	flag.Parse()
	rootPath := flag.Arg(0)

	broker := event.InitBus(1000)

	categoryArr := strings.Split(*categories, ",")
	categoryManager := category.New(broker, categoryArr, rootPath)
	categorizationManager := imagecategory.ManagerNew(rootPath, broker)
	imageLibrary := library.ForHandles(rootPath, broker)

	categorizationManager.LoadCategorization(imageLibrary, categoryManager)

	secret, _ := uuid.NewRandom()
	secretString := secret.String()
	c, _ := caster.InitCaster(secretString, broker)
	c.StartServer(*httpPort, rootPath)

	pixbufCache := pixbuf.NewPixbufCache(imageLibrary.GetHandles()[:5])
	gui := ui.Init(broker, pixbufCache)

	// Startup
	broker.Subscribe(event.UI_READY, func() {
		categoryManager.RequestCategories()
	})

	// UI -> Library
	broker.Subscribe(event.IMAGE_REQUEST_NEXT, imageLibrary.RequestNextImage)
	broker.Subscribe(event.IMAGE_REQUEST_NEXT_OFFSET, imageLibrary.RequestNextImageWithOffset)
	broker.Subscribe(event.IMAGE_REQUEST_PREV, imageLibrary.RequestPrevImage)
	broker.Subscribe(event.IMAGE_REQUEST_PREV_OFFSET, imageLibrary.RequestPrevImageWithOffset)
	broker.Subscribe(event.IMAGE_REQUEST_CURRENT, imageLibrary.RequestImages)
	broker.Subscribe(event.IMAGE_REQUEST, imageLibrary.RequestImage)
	broker.Subscribe(event.SIMILAR_REQUEST_SEARCH, imageLibrary.RequestGenerateHashes)
	broker.Subscribe(event.SIMILAR_REQUEST_STOP, imageLibrary.RequestStopHashes)
	broker.Subscribe(event.APPLICATION_CLOSE, imageLibrary.Close)

	// Library -> UI
	broker.ConnectToGui(event.IMAGE_UPDATE, gui.SetImages)
	broker.ConnectToGui(event.UPDATE_PROCESS_STATUS, gui.UpdateProgress)

	// UI -> Image Categorization
	broker.Subscribe(event.CATEGORIZE_IMAGE, categorizationManager.SetCategory)
	broker.Subscribe(event.CATEGORY_PERSIST_ALL, categorizationManager.PersistImageCategories)
	broker.Subscribe(event.IMAGE_CHANGED, categorizationManager.RequestCategory)
	broker.Subscribe(event.APPLICATION_CLOSE, categorizationManager.Close)

	// Image Categorization -> UI
	broker.ConnectToGui(event.CATEGORY_IMAGE_UPDATE, gui.SetImageCategory)

	// UI -> Caster
	broker.Subscribe(event.CAST_DEVICE_SEARCH, c.FindDevices)
	broker.Subscribe(event.CAST_DEVICE_SELECT, c.SelectDevice)
	broker.Subscribe(event.IMAGE_CHANGED, c.CastImage)
	broker.Subscribe(event.APPLICATION_CLOSE, c.Close)

	// Caster -> UI
	broker.ConnectToGui(event.CAST_DEVICE_FOUND, gui.DeviceFound)
	broker.ConnectToGui(event.CAST_READY, gui.CastReady)
	broker.ConnectToGui(event.CAST_DEVICES_SEARCH_DONE, gui.CastFindDone)

	// UI -> Category
	broker.Subscribe(event.CATEGORIES_SAVE, categoryManager.Save)
	broker.Subscribe(event.CATEGORIES_SAVE_DEFAULT, categoryManager.SaveDefault)
	broker.Subscribe(event.APPLICATION_CLOSE, categoryManager.Close)

	// Category -> UI
	broker.ConnectToGui(event.CATEGORIES_UPDATED, gui.UpdateCategories)

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
