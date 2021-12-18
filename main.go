package main

import (
	"fmt"
	"strings"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/backend"
	"vincit.fi/image-sorter/backend/dbapi"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/common/logger"
	"vincit.fi/image-sorter/common/util"
	giuUi "vincit.fi/image-sorter/ui/giu"
)

const databaseFileName = "image-sorter.db"
const EventBusQueueSize = 1000

func main() {
	params := common.ParseParams()
	logger.Initialize(logger.StringToLogLevel(params.LogLevel()))

	initAndRun(params)
}

func initAndRun(params *common.Params) {
	printHeaderToLogger()

	stores := backend.InitializeStores(databaseFileName)
	defer stores.Close()

	brokers := backend.InitializeEventBrokers(EventBusQueueSize)

	services := backend.InitializeServices(params, stores, brokers)
	defer services.Close()

	// UI
	logger.Debug.Printf("Initialize GUI")
	gui := giuUi.NewUi(params, brokers.Broker, services.ImageCache)

	connectUiAndServices(params, stores, services, brokers, gui)

	// Everything has been initialized, so it's time to start the UI
	logger.Debug.Printf("Backend initialized, run GUI")
	gui.Run()
}

func printHeaderToLogger() {
	appName := common.AppName
	appVersion := fmt.Sprintf("version: %s", common.Version)

	separatorLength := util.MaxInt(
		len(appName), len(appVersion),
	)
	separator := strings.Repeat("=", separatorLength)

	logger.Info.Printf(separator)
	logger.Info.Print(appName)
	logger.Info.Print(appVersion)
	logger.Info.Printf(separator)
}

func connectUiAndServices(params *common.Params, stores *backend.Stores, services *backend.Services, brokers *backend.Brokers, gui api.Gui) {
	logger.Debug.Printf("Connecting events to handlers...")
	// Initialize startup procedure run when the directory has been chosen
	handleDirectoryChanged := func(command *api.DirectoryChangedCommand) {
		directory := command.Directory
		brokers.Broker.SendToTopic(api.BackendLoading)
		logger.Info.Printf("Directory changed to '%s'", directory)

		if err := stores.WorkDirDb.InitializeForDirectory(directory, databaseFileName); err != nil {
			logger.Error.Fatal("Error opening database", err)
		} else {
			if tableExist := stores.WorkDirDb.Migrate(); tableExist == dbapi.TableNotExist {
				if defaultCategories, err := stores.DefaultCategoryStore.GetCategories(); err != nil {
					logger.Error.Print("Error while trying to load default categories ", err)
				} else {
					services.CategoryService.InitializeFromDirectory(params.Categories(), defaultCategories)
				}
			}
			services.ImageService.InitializeFromDirectory(directory)

			if len(services.ImageService.GetImageFiles()) > 0 {
				services.ImageCache.Initialize(services.ImageService.GetImageFiles(), api.NewSenderProgressReporter(brokers.Broker))
			}

			services.ImageCategoryService.InitializeForDirectory(directory)

			services.CategoryService.RequestCategories()
			brokers.Broker.SendToTopic(api.BackendReady)
		}
	}
	brokers.Broker.Subscribe(api.BackendLoading, gui.Pause)
	brokers.Broker.Subscribe(api.BackendReady, gui.Ready)
	brokers.Broker.Subscribe(api.DirectoryChanged, handleDirectoryChanged)

	// Connect Topics to methods

	// UI -> ImageService
	brokers.Broker.Subscribe(api.ImageRequestNext, services.ImageService.RequestNextImage)
	brokers.Broker.Subscribe(api.ImageRequestNextOffset, services.ImageService.RequestNextImageWithOffset)
	brokers.Broker.Subscribe(api.ImageRequestPrevious, services.ImageService.RequestPreviousImage)
	brokers.Broker.Subscribe(api.ImageRequestPreviousOffset, services.ImageService.RequestPreviousImageWithOffset)
	brokers.Broker.Subscribe(api.ImageRequestCurrent, services.ImageService.RequestImages)
	brokers.Broker.Subscribe(api.ImageRequest, services.ImageService.RequestImage)
	brokers.Broker.Subscribe(api.ImageRequestAtIndex, services.ImageService.RequestImageAt)
	brokers.Broker.Subscribe(api.ImageListSizeChanged, services.ImageService.SetImageListSize)
	brokers.Broker.Subscribe(api.ImageShowAll, services.ImageService.ShowAllImages)
	brokers.Broker.Subscribe(api.ImageShowOnly, services.ImageService.ShowOnlyImages)

	brokers.Broker.Subscribe(api.SimilarRequestSearch, services.ImageService.RequestGenerateHashes)
	brokers.Broker.Subscribe(api.SimilarRequestStop, services.ImageService.RequestStopHashes)
	brokers.Broker.Subscribe(api.SimilarSetShowImages, services.ImageService.SetSendSimilarImages)

	// ImageService -> UI
	brokers.Broker.Subscribe(api.ImageListUpdated, gui.SetImages)
	brokers.Broker.Subscribe(api.ImageCurrentUpdated, gui.SetCurrentImage)
	brokers.Broker.Subscribe(api.ProcessStatusUpdated, gui.UpdateProgress)
	brokers.Broker.Subscribe(api.ShowError, gui.ShowError)

	// UI -> Image Categorization
	brokers.Broker.Subscribe(api.CategorizeImage, services.ImageCategoryService.SetCategory)
	brokers.Broker.Subscribe(api.CategoryPersistAll, services.ImageCategoryService.PersistImageCategories)
	brokers.Broker.Subscribe(api.ImageChanged, services.ImageCategoryService.RequestCategory)
	brokers.Broker.Subscribe(api.CategoriesShowOnly, services.ImageCategoryService.ShowOnlyCategoryImages)

	// Image Categorization -> UI
	brokers.Broker.Subscribe(api.CategoryImageUpdate, gui.SetImageCategory)

	// UI -> Caster
	brokers.Broker.Subscribe(api.CastDeviceSearch, services.CasterInstance.FindDevices)
	brokers.Broker.Subscribe(api.CastDeviceSelect, services.CasterInstance.SelectDevice)
	brokers.Broker.Subscribe(api.ImageChanged, services.CasterInstance.CastImage)

	// Caster -> UI
	brokers.Broker.Subscribe(api.CastDeviceFound, gui.DeviceFound)
	brokers.Broker.Subscribe(api.CastReady, gui.CastReady)
	brokers.Broker.Subscribe(api.CastDevicesSearchDone, gui.CastFindDone)

	// UI -> Category
	brokers.Broker.Subscribe(api.CategoriesSave, services.CategoryService.Save)
	brokers.Broker.Subscribe(api.CategoriesSaveDefault, services.DefaultCategoryService.Save)

	// Category -> UI
	brokers.Broker.Subscribe(api.CategoriesUpdated, gui.UpdateCategories)

	logger.Debug.Printf("Events connected to handlers")
}
