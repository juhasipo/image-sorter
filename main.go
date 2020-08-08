package main

import (
	"flag"
	"github.com/google/uuid"
	"runtime"
	"strings"
	"time"
	"vincit.fi/image-sorter/caster"
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/filter"
	"vincit.fi/image-sorter/imagecategory"
	"vincit.fi/image-sorter/imageloader"
	"vincit.fi/image-sorter/library"
	"vincit.fi/image-sorter/logger"
	gtkUi "vincit.fi/image-sorter/ui/gtk"
)

func main() {
	categories := flag.String("categories", "", "Comma separated categories. Each category in format <name>:<shortcut> e.g. Good:G")
	httpPort := flag.Int("httpPort", 8080, "HTTP Server port for Chrome Cast")
	secret := flag.String("secret", "", "Override default random secret for casting")
	alwaysStartHttpServer := flag.Bool("alwaysStartHttpServer", false, "Always start HTTP server. Not only when casting.")
	logLevel := flag.String("logLevel", "INFO", "Log level: ERROR, WARN, INFO, DEBUG, Trace")

	flag.Parse()
	rootPath := flag.Arg(0)

	logger.Initialize(logger.StringToLogLevel(*logLevel))

	broker := event.InitBus(1000)

	categoryArr := strings.Split(*categories, ",")
	categoryManager := category.New(broker, categoryArr)
	imageLoader := imageloader.NewImageLoader()
	imageCache := imageloader.NewImageCache(imageLoader)
	imageLibrary := library.NewLibrary(broker, imageCache, imageLoader)
	filterManager := filter.NewFilterManager()
	categorizationManager := imagecategory.NewManager(broker, imageLibrary, filterManager, imageLoader)

	secretValue := resolveSecret(*secret)
	castManager := caster.InitCaster(*httpPort, *alwaysStartHttpServer, secretValue, broker, imageCache)

	gui := gtkUi.Init(rootPath, broker, imageCache)

	// Startup
	broker.Subscribe(event.UiReady, func() {
		categoryManager.RequestCategories()
	})
	broker.Subscribe(event.DirectoryChanged, func(directory string) {
		categoryManager.InitializeFromDirectory([]string{}, directory)
		imageLibrary.InitializeFromDirectory(directory)
		if len(imageLibrary.GetHandles()) > 0 {
			imageCache.Initialize(imageLibrary.GetHandles()[:5])
		}

		categorizationManager.InitializeForDirectory(directory)
		categorizationManager.LoadCategorization(imageLibrary, categoryManager)

		categoryManager.RequestCategories()
	})

	// UI -> Library
	broker.Subscribe(event.ImageRequestNext, imageLibrary.RequestNextImage)
	broker.Subscribe(event.ImageRequestNextOffset, imageLibrary.RequestNextImageWithOffset)
	broker.Subscribe(event.ImageRequestPrev, imageLibrary.RequestPrevImage)
	broker.Subscribe(event.ImageRequestPrevOffset, imageLibrary.RequestPrevImageWithOffset)
	broker.Subscribe(event.ImageRequestCurrent, imageLibrary.RequestImages)
	broker.Subscribe(event.ImageRequest, imageLibrary.RequestImage)
	broker.Subscribe(event.ImageRequestAtIndex, imageLibrary.RequestImageAt)
	broker.Subscribe(event.ImageListSizeChanged, imageLibrary.SetImageListSize)
	broker.Subscribe(event.ImageShowAll, imageLibrary.ShowAllImages)
	broker.Subscribe(event.ImageShowOnly, imageLibrary.ShowOnlyImages)

	broker.Subscribe(event.SimilarRequestSearch, imageLibrary.RequestGenerateHashes)
	broker.Subscribe(event.SimilarRequestStop, imageLibrary.RequestStopHashes)
	broker.Subscribe(event.SimilarSetShowImages, imageLibrary.SetSendSimilarImages)

	// Library -> UI
	broker.ConnectToGui(event.ImageListUpdated, gui.SetImages)
	broker.ConnectToGui(event.ImageCurrentUpdated, gui.SetCurrentImage)
	broker.ConnectToGui(event.ProcessStatusUpdated, gui.UpdateProgress)

	// UI -> Image Categorization
	broker.Subscribe(event.CategorizeImage, categorizationManager.SetCategory)
	broker.Subscribe(event.CategoryPersistAll, categorizationManager.PersistImageCategories)
	broker.Subscribe(event.ImageChanged, categorizationManager.RequestCategory)
	broker.Subscribe(event.CategoriesShowOnly, categorizationManager.ShowOnlyCategoryImages)

	// Image Categorization -> UI
	broker.ConnectToGui(event.CategoryImageUpdate, gui.SetImageCategory)

	// UI -> Caster
	broker.Subscribe(event.CastDeviceSearch, castManager.FindDevices)
	broker.Subscribe(event.CastDeviceSelect, castManager.SelectDevice)
	broker.Subscribe(event.ImageChanged, castManager.CastImage)

	// Caster -> UI
	broker.ConnectToGui(event.CastDeviceFound, gui.DeviceFound)
	broker.ConnectToGui(event.CastReady, gui.CastReady)
	broker.ConnectToGui(event.CastDevicesSearchDone, gui.CastFindDone)

	// UI -> Category
	broker.Subscribe(event.CategoriesSave, categoryManager.Save)
	broker.Subscribe(event.CategoriesSaveDefault, categoryManager.SaveDefault)

	// Category -> UI
	broker.ConnectToGui(event.CategoriesUpdated, gui.UpdateCategories)

	StartBackgroundGC()

	gui.Run()

	categoryManager.Close()
	categorizationManager.Close()
	imageLibrary.Close()
	castManager.Close()
}

func resolveSecret(secret string) string {
	if secret == "" {
		if randomSecret, err := uuid.NewRandom(); err != nil {
			logger.Error.Panic("Could not initialize secret for casting", err)
			return ""
		} else {
			return randomSecret.String()
		}
	} else {
		return secret
	}
}

func StartBackgroundGC() {
	logger.Debug.Print("Start GC background process")
	go func() {
		for true {
			logger.Trace.Printf("Running GC")
			runtime.GC()
			time.Sleep(30 * time.Second)
		}
	}()
}
