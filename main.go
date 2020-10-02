package main

import (
	"runtime"
	"time"
	"vincit.fi/image-sorter/backend/caster"
	"vincit.fi/image-sorter/backend/category"
	"vincit.fi/image-sorter/backend/filter"
	"vincit.fi/image-sorter/backend/imagecategory"
	"vincit.fi/image-sorter/backend/imageloader"
	"vincit.fi/image-sorter/backend/library"
	"vincit.fi/image-sorter/common/event"
	"vincit.fi/image-sorter/common/logger"
	"vincit.fi/image-sorter/common/util"
	gtkUi "vincit.fi/image-sorter/ui/gtk"
)

const EventBusQueueSize = 1000

func main() {
	params := util.ParseParams()

	logger.Initialize(logger.StringToLogLevel(params.GetLogLevel()))

	broker := event.InitBus(EventBusQueueSize)

	categoryManager := category.New(params, broker)
	imageLoader := imageloader.NewImageLoader()
	imageCache := imageloader.NewImageCache(imageLoader)
	imageLibrary := library.NewLibrary(broker, imageCache, imageLoader)
	filterManager := filter.NewFilterManager()
	imageCategoryManager := imagecategory.NewImageCategoryManager(broker, imageLibrary, filterManager, imageLoader)

	casterInstance := caster.NewCaster(params, broker, imageCache)

	defer categoryManager.Close()
	defer imageCategoryManager.Close()
	defer imageLibrary.Close()
	defer casterInstance.Close()

	gui := gtkUi.NewUi(params, broker, imageCache)

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

		imageCategoryManager.InitializeForDirectory(directory)
		imageCategoryManager.LoadCategorization(imageLibrary, categoryManager)

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
	broker.Subscribe(event.CategorizeImage, imageCategoryManager.SetCategory)
	broker.Subscribe(event.CategoryPersistAll, imageCategoryManager.PersistImageCategories)
	broker.Subscribe(event.ImageChanged, imageCategoryManager.RequestCategory)
	broker.Subscribe(event.CategoriesShowOnly, imageCategoryManager.ShowOnlyCategoryImages)

	// Image Categorization -> UI
	broker.ConnectToGui(event.CategoryImageUpdate, gui.SetImageCategory)

	// UI -> Caster
	broker.Subscribe(event.CastDeviceSearch, casterInstance.FindDevices)
	broker.Subscribe(event.CastDeviceSelect, casterInstance.SelectDevice)
	broker.Subscribe(event.ImageChanged, casterInstance.CastImage)

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
