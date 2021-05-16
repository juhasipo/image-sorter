package main

import (
	"fmt"
	"os/user"
	"strings"
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
	"vincit.fi/image-sorter/common/util"
	gtkUi "vincit.fi/image-sorter/ui/gtk"
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

	stores := initializeStores()
	defer stores.Close()

	brokers := initializeEventBrokers()

	imageLoader := imageloader.NewImageLoader(stores.ImageStore)
	imageCache := imageloader.NewImageCache(imageLoader)
	services := initializeServices(params, stores, brokers, imageCache, imageLoader)
	defer services.Close()

	// UI
	gui := gtkUi.NewUi(params, brokers.Broker, imageCache)

	connectUiAndServices(params, stores, services, imageCache, brokers, gui)

	// Everything has been initialized so it is time to start the UI
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

type Stores struct {
	ImageStore           *database.ImageStore
	ImageMetaDataStore   *database.ImageMetaDataStore
	SimilarityIndex      *database.SimilarityIndex
	CategoryStore        *database.CategoryStore
	DefaultCategoryStore *database.CategoryStore
	ImageCategoryStore   *database.ImageCategoryStore
	HomeDirDb            *database.Database
	WorkDirDb            *database.Database
}

func (s *Stores) Close() {
	defer s.HomeDirDb.Close()
	defer s.WorkDirDb.Close()
}

type Services struct {
	CategoryService        api.CategoryService
	DefaultCategoryService api.CategoryService
	ImageService           api.ImageService
	FilterService          *filter.FilterService
	ImageCategoryService   api.ImageCategoryService
	CasterInstance         api.Caster
}

func (s *Services) Close() {
	defer s.CategoryService.Close()
	defer s.DefaultCategoryService.Close()
	defer s.ImageService.Close()
	defer s.ImageCategoryService.Close()
	defer s.CasterInstance.Close()
}

type Brokers struct {
	Broker        *event.Broker
	DevNullBroker *event.Broker
}

func initializeEventBrokers() *Brokers {
	brokers := &Brokers{
		Broker:        event.InitBus(EventBusQueueSize),
		DevNullBroker: event.InitDevNullBus(),
	}
	return brokers
}

func connectUiAndServices(params *common.Params, stores *Stores, services *Services, imageCache api.ImageStore, brokers *Brokers, gui api.Gui) {
	// Initialize startup procedure run when the directory has been chosen
	handleDirectoryChanged := func(directory string) {
		logger.Info.Printf("Directory changed to '%s'", directory)

		if err := stores.WorkDirDb.InitializeForDirectory(directory, databaseFileName); err != nil {
			logger.Error.Fatal("Error opening database", err)
		} else {
			if tableExist := stores.WorkDirDb.Migrate(); tableExist == database.TableNotExist {
				if defaultCategories, err := stores.DefaultCategoryStore.GetCategories(); err != nil {
					logger.Error.Print("Error while trying to load default categories ", err)
				} else {
					services.CategoryService.InitializeFromDirectory(params.Categories(), defaultCategories)
				}
			}
			services.ImageService.InitializeFromDirectory(directory)

			if len(services.ImageService.GetImageFiles()) > 0 {
				imageCache.Initialize(services.ImageService.GetImageFiles()[:5])
			}

			services.ImageCategoryService.InitializeForDirectory(directory)

			services.CategoryService.RequestCategories()
		}
	}
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
	brokers.Broker.ConnectToGui(api.ImageListUpdated, gui.SetImages)
	brokers.Broker.ConnectToGui(api.ImageCurrentUpdated, gui.SetCurrentImage)
	brokers.Broker.ConnectToGui(api.ProcessStatusUpdated, gui.UpdateProgress)
	brokers.Broker.ConnectToGui(api.ShowError, gui.ShowError)

	// UI -> Image Categorization
	brokers.Broker.Subscribe(api.CategorizeImage, services.ImageCategoryService.SetCategory)
	brokers.Broker.Subscribe(api.CategoryPersistAll, services.ImageCategoryService.PersistImageCategories)
	brokers.Broker.Subscribe(api.ImageChanged, services.ImageCategoryService.RequestCategory)
	brokers.Broker.Subscribe(api.CategoriesShowOnly, services.ImageCategoryService.ShowOnlyCategoryImages)

	// Image Categorization -> UI
	brokers.Broker.ConnectToGui(api.CategoryImageUpdate, gui.SetImageCategory)

	// UI -> Caster
	brokers.Broker.Subscribe(api.CastDeviceSearch, services.CasterInstance.FindDevices)
	brokers.Broker.Subscribe(api.CastDeviceSelect, services.CasterInstance.SelectDevice)
	brokers.Broker.Subscribe(api.ImageChanged, services.CasterInstance.CastImage)

	// Caster -> UI
	brokers.Broker.ConnectToGui(api.CastDeviceFound, gui.DeviceFound)
	brokers.Broker.ConnectToGui(api.CastReady, gui.CastReady)
	brokers.Broker.ConnectToGui(api.CastDevicesSearchDone, gui.CastFindDone)

	// UI -> Category
	brokers.Broker.Subscribe(api.CategoriesSave, services.CategoryService.Save)
	brokers.Broker.Subscribe(api.CategoriesSaveDefault, services.DefaultCategoryService.Save)

	// Category -> UI
	brokers.Broker.ConnectToGui(api.CategoriesUpdated, gui.UpdateCategories)

	// Services -> Caster
	brokers.Broker.ConnectToGui(api.ImageCurrentUpdated, services.CasterInstance.SetCurrentImage)
	brokers.Broker.ConnectToGui(api.CategoryImageUpdate, services.CasterInstance.SetImageCategory)
	brokers.Broker.ConnectToGui(api.CategoriesUpdated, services.CasterInstance.UpdateCategories)
}

func initializeServices(params *common.Params, stores *Stores, brokers *Brokers, imageCache api.ImageStore, imageLoader api.ImageLoader) *Services {
	filterService := filter.NewFilterService()
	imageService := library.NewImageService(brokers.Broker, imageCache, imageLoader, stores.SimilarityIndex, stores.ImageStore, stores.ImageMetaDataStore)
	services := &Services{
		CategoryService:        category.NewCategoryService(params, brokers.Broker, stores.CategoryStore),
		DefaultCategoryService: category.NewCategoryService(params, brokers.DevNullBroker, stores.DefaultCategoryStore),
		ImageService:           imageService,
		FilterService:          filterService,
		ImageCategoryService:   imagecategory.NewImageCategoryService(brokers.Broker, imageService, filterService, imageLoader, stores.ImageCategoryStore),
		CasterInstance:         caster.NewCaster(params, brokers.Broker, imageCache),
	}
	return services
}

// Initialize the configuration DB in user's home folder and
// DB for the working directory. Home dir DB can be initialized
// right away. workDirDb will be initialized once the work dir
// is known.
func initializeStores() *Stores {
	homeDirDb := database.NewDatabase()
	if currentUser, err := user.Current(); err != nil {
		logger.Error.Fatal("Cannot load user")
	} else {
		if err := homeDirDb.InitializeForDirectory(currentUser.HomeDir, databaseFileName); err != nil {
			logger.Error.Fatal("Error opening database", err)
		} else {
			_ = homeDirDb.Migrate()
		}
	}

	workDirDb := database.NewDatabase()

	stores := &Stores{
		ImageStore:           database.NewImageStore(workDirDb, &database.FileSystemImageFileConverter{}),
		ImageMetaDataStore:   database.NewImageMetaDataStore(workDirDb),
		SimilarityIndex:      database.NewSimilarityIndex(workDirDb),
		CategoryStore:        database.NewCategoryStore(workDirDb),
		ImageCategoryStore:   database.NewImageCategoryStore(workDirDb),
		DefaultCategoryStore: database.NewCategoryStore(homeDirDb),
		HomeDirDb:            homeDirDb,
		WorkDirDb:            workDirDb,
	}
	return stores
}
