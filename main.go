package main

import (
	"os/user"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/backend/caster"
	"vincit.fi/image-sorter/backend/category"
	"vincit.fi/image-sorter/backend/database"
	"vincit.fi/image-sorter/backend/filter"
	"vincit.fi/image-sorter/backend/imagecategory"
	"vincit.fi/image-sorter/backend/imageloader"
	"vincit.fi/image-sorter/backend/library"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/common/event"
	"vincit.fi/image-sorter/common/logger"
	gtkUi "vincit.fi/image-sorter/ui/gtk"
)

const EventBusQueueSize = 1000

func main() {
	params := common.ParseParams()

	logger.Initialize(logger.StringToLogLevel(params.LogLevel()))

	configDb := database.NewDatabase()
	if currentUser, err := user.Current(); err != nil {
		logger.Error.Fatal("Cannot load user")
	} else {
		if err := configDb.InitializeForDirectory(currentUser.HomeDir, "image-sorter.db"); err != nil {
			logger.Error.Fatal("Error opening database", err)
		} else {
			_ = configDb.Migrate()
		}
	}
	defaultCategoryStore := database.NewCategoryStore(configDb)

	imageDb := database.NewDatabase()
	defer imageDb.Close()

	imageStore := database.NewImageStore(imageDb, &database.FileSystemImageFileConverter{})
	similarityIndex := database.NewSimilarityIndex(imageDb)
	categoryStore := database.NewCategoryStore(imageDb)
	imageCategoryStore := database.NewImageCategoryStore(imageDb)

	broker := event.InitBus(EventBusQueueSize)
	devNull := event.InitDevNullBus()

	categoryManager := category.NewCategoryManager(params, broker, categoryStore)
	defaultCategoryManager := category.NewCategoryManager(params, devNull, defaultCategoryStore)
	imageLoader := imageloader.NewImageLoader(imageStore)
	imageCache := imageloader.NewImageCache(imageLoader)
	imageLibrary := library.NewLibrary(broker, imageCache, imageLoader, similarityIndex, imageStore)
	filterManager := filter.NewFilterManager()
	imageCategoryManager := imagecategory.NewImageCategoryManager(broker, imageLibrary, filterManager, imageLoader, imageCategoryStore)

	casterInstance := caster.NewCaster(params, broker, imageCache)

	defer categoryManager.Close()
	defer imageCategoryManager.Close()
	defer imageLibrary.Close()
	defer casterInstance.Close()

	gui := gtkUi.NewUi(params, broker, imageCache)

	// Startup
	broker.Subscribe(api.DirectoryChanged, func(directory string) {
		logger.Info.Printf("Directory changed to '%s'", directory)

		if err := imageDb.InitializeForDirectory(directory, "image-sorter.db"); err != nil {
			logger.Error.Fatal("Error opening database", err)
		} else {
			if tableExist := imageDb.Migrate(); tableExist == database.TableNotExist {
				if defaultCategories, err := defaultCategoryStore.GetCategories(); err != nil {
					logger.Error.Print("Error while trying to load default categories ", err)
				} else {
					categoryManager.InitializeFromDirectory(params.Categories(), defaultCategories)
				}
			}
			imageLibrary.InitializeFromDirectory(directory)

			if len(imageLibrary.GetImageFiles()) > 0 {
				imageCache.Initialize(imageLibrary.GetImageFiles()[:5])
			}

			imageCategoryManager.InitializeForDirectory(directory)

			categoryManager.RequestCategories()
		}
	})

	// UI -> Library
	broker.Subscribe(api.ImageRequestNext, imageLibrary.RequestNextImage)
	broker.Subscribe(api.ImageRequestNextOffset, imageLibrary.RequestNextImageWithOffset)
	broker.Subscribe(api.ImageRequestPrev, imageLibrary.RequestPrevImage)
	broker.Subscribe(api.ImageRequestPrevOffset, imageLibrary.RequestPrevImageWithOffset)
	broker.Subscribe(api.ImageRequestCurrent, imageLibrary.RequestImages)
	broker.Subscribe(api.ImageRequest, imageLibrary.RequestImage)
	broker.Subscribe(api.ImageRequestAtIndex, imageLibrary.RequestImageAt)
	broker.Subscribe(api.ImageListSizeChanged, imageLibrary.SetImageListSize)
	broker.Subscribe(api.ImageShowAll, imageLibrary.ShowAllImages)
	broker.Subscribe(api.ImageShowOnly, imageLibrary.ShowOnlyImages)

	broker.Subscribe(api.SimilarRequestSearch, imageLibrary.RequestGenerateHashes)
	broker.Subscribe(api.SimilarRequestStop, imageLibrary.RequestStopHashes)
	broker.Subscribe(api.SimilarSetShowImages, imageLibrary.SetSendSimilarImages)

	// Library -> UI
	broker.ConnectToGui(api.ImageListUpdated, gui.SetImages)
	broker.ConnectToGui(api.ImageCurrentUpdated, gui.SetCurrentImage)
	broker.ConnectToGui(api.ProcessStatusUpdated, gui.UpdateProgress)
	broker.ConnectToGui(api.ShowError, gui.ShowError)

	// UI -> Image Categorization
	broker.Subscribe(api.CategorizeImage, imageCategoryManager.SetCategory)
	broker.Subscribe(api.CategoryPersistAll, imageCategoryManager.PersistImageCategories)
	broker.Subscribe(api.ImageChanged, imageCategoryManager.RequestCategory)
	broker.Subscribe(api.CategoriesShowOnly, imageCategoryManager.ShowOnlyCategoryImages)

	// Image Categorization -> UI
	broker.ConnectToGui(api.CategoryImageUpdate, gui.SetImageCategory)

	// UI -> Caster
	broker.Subscribe(api.CastDeviceSearch, casterInstance.FindDevices)
	broker.Subscribe(api.CastDeviceSelect, casterInstance.SelectDevice)
	broker.Subscribe(api.ImageChanged, casterInstance.CastImage)

	// Caster -> UI
	broker.ConnectToGui(api.CastDeviceFound, gui.DeviceFound)
	broker.ConnectToGui(api.CastReady, gui.CastReady)
	broker.ConnectToGui(api.CastDevicesSearchDone, gui.CastFindDone)

	// UI -> Category
	broker.Subscribe(api.CategoriesSave, categoryManager.Save)
	broker.Subscribe(api.CategoriesSaveDefault, defaultCategoryManager.Save)

	// Category -> UI
	broker.ConnectToGui(api.CategoriesUpdated, gui.UpdateCategories)

	gui.Run()
}
