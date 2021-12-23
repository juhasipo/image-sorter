package backend

import (
	"os/user"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/backend/dbapi"
	"vincit.fi/image-sorter/backend/internal/caster"
	"vincit.fi/image-sorter/backend/internal/category"
	"vincit.fi/image-sorter/backend/internal/database"
	"vincit.fi/image-sorter/backend/internal/filter"
	"vincit.fi/image-sorter/backend/internal/imagecategory"
	"vincit.fi/image-sorter/backend/internal/imageloader"
	"vincit.fi/image-sorter/backend/internal/library"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/common/event"
	"vincit.fi/image-sorter/common/logger"
)

type Stores struct {
	ImageStore           *database.ImageStore
	ImageMetaDataStore   *database.ImageMetaDataStore
	SimilarityIndex      *database.SimilarityIndex
	CategoryStore        *database.CategoryStore
	DefaultCategoryStore *database.CategoryStore
	ImageCategoryStore   *database.ImageCategoryStore
	StatusStore          *database.StatusStore
	homeDirDb            *database.Database
	workDirDb            *database.Database
}

func (s *Stores) Close() {
	defer s.homeDirDb.Close()
	defer s.workDirDb.Close()
}

func (s *Stores) InitializeForDirectory(directory string, file string) error {
	return s.workDirDb.InitializeForDirectory(directory, file)
}

func (s *Stores) Migrate() dbapi.TableExist {
	return s.workDirDb.Migrate()
}

type Services struct {
	CategoryService        api.CategoryService
	DefaultCategoryService api.CategoryService
	ImageService           api.ImageService
	ImageLibrary           api.ImageLibrary
	FilterService          *filter.FilterService
	ImageCategoryService   api.ImageCategoryService
	CasterInstance         api.Caster
	ImageLoader            api.ImageLoader
	ImageCache             api.ImageStore
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

func InitializeEventBrokers(eventBusQueueSize int) *Brokers {
	logger.Debug.Printf("Initialize event brokers...")
	brokers := &Brokers{
		Broker:        event.InitBus(eventBusQueueSize),
		DevNullBroker: event.InitDevNullBus(),
	}
	logger.Debug.Printf("Event brokers initialized")
	return brokers
}

func InitializeServices(params *common.Params, stores *Stores, brokers *Brokers) *Services {
	logger.Debug.Printf("Initialize services...")
	imageLoader := imageloader.NewImageLoader(stores.ImageStore)
	imageCache := imageloader.NewImageCache(imageLoader)

	filterService := filter.NewFilterService()
	progressReporter := api.NewSenderProgressReporter(brokers.Broker)
	imageLibrary := library.NewImageLibrary(imageCache, imageLoader, stores.SimilarityIndex, stores.ImageStore, stores.ImageMetaDataStore, progressReporter)
	imageService := library.NewImageService(brokers.Broker, imageLibrary, stores.StatusStore)
	services := &Services{
		CategoryService:        category.NewCategoryService(params, brokers.Broker, stores.CategoryStore),
		DefaultCategoryService: category.NewCategoryService(params, brokers.DevNullBroker, stores.DefaultCategoryStore),
		ImageService:           imageService,
		ImageLibrary:           imageLibrary,
		FilterService:          filterService,
		ImageCategoryService:   imagecategory.NewImageCategoryService(brokers.Broker, imageService, filterService, imageLoader, stores.ImageCategoryStore),
		CasterInstance:         caster.NewCaster(params, brokers.Broker, imageCache),
		ImageLoader:            imageLoader,
		ImageCache:             imageCache,
	}
	logger.Debug.Printf("Services initialized")
	return services
}

// Initialize the configuration DB in user's home folder and
// DB for the working directory. Home dir DB can be initialized
// right away. workDirDb will be initialized once the work dir
// is known.
func InitializeStores(databaseFileName string) *Stores {
	logger.Debug.Printf("Initialize databases...")
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

	logger.Debug.Printf("Initialize backend stores...")
	stores := &Stores{
		ImageStore:           database.NewImageStore(workDirDb, &database.FileSystemImageFileConverter{}),
		ImageMetaDataStore:   database.NewImageMetaDataStore(workDirDb),
		SimilarityIndex:      database.NewSimilarityIndex(workDirDb),
		CategoryStore:        database.NewCategoryStore(workDirDb),
		ImageCategoryStore:   database.NewImageCategoryStore(workDirDb),
		DefaultCategoryStore: database.NewCategoryStore(homeDirDb),
		StatusStore:          database.NewStatusStore(workDirDb),
		homeDirDb:            homeDirDb,
		workDirDb:            workDirDb,
	}
	logger.Debug.Printf("Stores and databases initialized")
	return stores
}
