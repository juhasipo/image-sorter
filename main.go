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

const databaseFileName = "image-sorter.db"
const EventBusQueueSize = 1000

func main() {
	params := common.ParseParams()
	logger.Initialize(logger.StringToLogLevel(params.LogLevel()))

	initAndRun(params)
}

func initAndRun(params *common.Params) {
	stores := initializeStores()
	defer stores.Close()

	brokers := initializeEventBrokers()

	imageLoader := imageloader.NewImageLoader(stores.ImageStore)
	imageCache := imageloader.NewImageCache(imageLoader)
	managers := initializeManagers(params, stores, brokers, imageCache, imageLoader)
	defer managers.Close()

	// UI
	gui := gtkUi.NewUi(params, brokers.Broker, imageCache)

	connectUiAndManagers(params, stores, managers, imageCache, brokers, gui)

	// Everything has been initialized so it is time to start the UI
	gui.Run()
}

type Stores struct {
	ImageStore           *database.ImageStore
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

type Managers struct {
	CategoryManager        api.CategoryManager
	DefaultCategoryManager api.CategoryManager
	ImageLibrary           api.Library
	FilterManager          *filter.Manager
	ImageCategoryManager   api.ImageCategoryManager
	CasterInstance         api.Caster
}

func (s *Managers) Close() {
	defer s.CategoryManager.Close()
	defer s.DefaultCategoryManager.Close()
	defer s.ImageLibrary.Close()
	defer s.ImageCategoryManager.Close()
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

func connectUiAndManagers(params *common.Params, stores *Stores, managers *Managers, imageCache api.ImageStore, brokers *Brokers, gui api.Gui) {
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
					managers.CategoryManager.InitializeFromDirectory(params.Categories(), defaultCategories)
				}
			}
			managers.ImageLibrary.InitializeFromDirectory(directory)

			if len(managers.ImageLibrary.GetImageFiles()) > 0 {
				imageCache.Initialize(managers.ImageLibrary.GetImageFiles()[:5])
			}

			managers.ImageCategoryManager.InitializeForDirectory(directory)

			managers.CategoryManager.RequestCategories()
		}
	}
	brokers.Broker.Subscribe(api.DirectoryChanged, handleDirectoryChanged)

	// Connect Topics to methods

	// UI -> Library
	brokers.Broker.Subscribe(api.ImageRequestNext, managers.ImageLibrary.RequestNextImage)
	brokers.Broker.Subscribe(api.ImageRequestNextOffset, managers.ImageLibrary.RequestNextImageWithOffset)
	brokers.Broker.Subscribe(api.ImageRequestPrev, managers.ImageLibrary.RequestPrevImage)
	brokers.Broker.Subscribe(api.ImageRequestPrevOffset, managers.ImageLibrary.RequestPrevImageWithOffset)
	brokers.Broker.Subscribe(api.ImageRequestCurrent, managers.ImageLibrary.RequestImages)
	brokers.Broker.Subscribe(api.ImageRequest, managers.ImageLibrary.RequestImage)
	brokers.Broker.Subscribe(api.ImageRequestAtIndex, managers.ImageLibrary.RequestImageAt)
	brokers.Broker.Subscribe(api.ImageListSizeChanged, managers.ImageLibrary.SetImageListSize)
	brokers.Broker.Subscribe(api.ImageShowAll, managers.ImageLibrary.ShowAllImages)
	brokers.Broker.Subscribe(api.ImageShowOnly, managers.ImageLibrary.ShowOnlyImages)

	brokers.Broker.Subscribe(api.SimilarRequestSearch, managers.ImageLibrary.RequestGenerateHashes)
	brokers.Broker.Subscribe(api.SimilarRequestStop, managers.ImageLibrary.RequestStopHashes)
	brokers.Broker.Subscribe(api.SimilarSetShowImages, managers.ImageLibrary.SetSendSimilarImages)

	// Library -> UI
	brokers.Broker.ConnectToGui(api.ImageListUpdated, gui.SetImages)
	brokers.Broker.ConnectToGui(api.ImageCurrentUpdated, gui.SetCurrentImage)
	brokers.Broker.ConnectToGui(api.ProcessStatusUpdated, gui.UpdateProgress)
	brokers.Broker.ConnectToGui(api.ShowError, gui.ShowError)

	// UI -> Image Categorization
	brokers.Broker.Subscribe(api.CategorizeImage, managers.ImageCategoryManager.SetCategory)
	brokers.Broker.Subscribe(api.CategoryPersistAll, managers.ImageCategoryManager.PersistImageCategories)
	brokers.Broker.Subscribe(api.ImageChanged, managers.ImageCategoryManager.RequestCategory)
	brokers.Broker.Subscribe(api.CategoriesShowOnly, managers.ImageCategoryManager.ShowOnlyCategoryImages)

	// Image Categorization -> UI
	brokers.Broker.ConnectToGui(api.CategoryImageUpdate, gui.SetImageCategory)

	// UI -> Caster
	brokers.Broker.Subscribe(api.CastDeviceSearch, managers.CasterInstance.FindDevices)
	brokers.Broker.Subscribe(api.CastDeviceSelect, managers.CasterInstance.SelectDevice)
	brokers.Broker.Subscribe(api.ImageChanged, managers.CasterInstance.CastImage)

	// Caster -> UI
	brokers.Broker.ConnectToGui(api.CastDeviceFound, gui.DeviceFound)
	brokers.Broker.ConnectToGui(api.CastReady, gui.CastReady)
	brokers.Broker.ConnectToGui(api.CastDevicesSearchDone, gui.CastFindDone)

	// UI -> Category
	brokers.Broker.Subscribe(api.CategoriesSave, managers.CategoryManager.Save)
	brokers.Broker.Subscribe(api.CategoriesSaveDefault, managers.DefaultCategoryManager.Save)

	// Category -> UI
	brokers.Broker.ConnectToGui(api.CategoriesUpdated, gui.UpdateCategories)
}

func initializeManagers(params *common.Params, stores *Stores, brokers *Brokers, imageCache api.ImageStore, imageLoader api.ImageLoader) *Managers {
	filterManager := filter.NewFilterManager()
	imageLibrary := library.NewLibrary(brokers.Broker, imageCache, imageLoader, stores.SimilarityIndex, stores.ImageStore)
	managers := &Managers{
		CategoryManager:        category.NewCategoryManager(params, brokers.Broker, stores.CategoryStore),
		DefaultCategoryManager: category.NewCategoryManager(params, brokers.DevNullBroker, stores.DefaultCategoryStore),
		ImageLibrary:           imageLibrary,
		FilterManager:          filterManager,
		ImageCategoryManager:   imagecategory.NewImageCategoryManager(brokers.Broker, imageLibrary, filterManager, imageLoader, stores.ImageCategoryStore),
		CasterInstance:         caster.NewCaster(params, brokers.Broker, imageCache),
	}
	return managers
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
		SimilarityIndex:      database.NewSimilarityIndex(workDirDb),
		CategoryStore:        database.NewCategoryStore(workDirDb),
		ImageCategoryStore:   database.NewImageCategoryStore(workDirDb),
		DefaultCategoryStore: database.NewCategoryStore(homeDirDb),
		HomeDirDb:            homeDirDb,
		WorkDirDb:            workDirDb,
	}
	return stores
}
