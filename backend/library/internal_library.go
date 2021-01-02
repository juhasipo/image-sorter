package library

import (
	"runtime"
	"sort"
	"time"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/database"
	"vincit.fi/image-sorter/common/logger"
	"vincit.fi/image-sorter/common/util"
	"vincit.fi/image-sorter/duplo"
)

var (
	emptyHandles = []*apitype.ImageContainer{}
)

type ImageList func(number int) []*apitype.Handle

type internalManager struct {
	rootDir                     string
	imageHandles                map[int64]*apitype.Handle
	imagesTitle                 string
	index                       int
	imageHash                   *duplo.Store
	shouldSendSimilar           bool
	shouldGenerateSimilarHashed bool
	categoryManager             api.CategoryManager
	imageListSize               int
	imageStore                  api.ImageStore
	imageLoader                 api.ImageLoader
	store                       *database.Store

	stopChannel   chan bool
	outputChannel chan *HashResult
}

func newLibrary(imageCache api.ImageStore, imageLoader api.ImageLoader, store *database.Store) *internalManager {
	var manager = internalManager{
		index:                       0,
		imageHash:                   duplo.New(),
		shouldGenerateSimilarHashed: true,
		imageListSize:               0,
		imageStore:                  imageCache,
		imageLoader:                 imageLoader,
		store:                       store,
	}
	return &manager
}

func (s *internalManager) InitializeFromDirectory(directory string) {
	s.rootDir = directory
	s.index = 0
	s.imageHash = duplo.New()
	s.shouldGenerateSimilarHashed = true
	s.loadImagesFromRootDir()
}

func (s *internalManager) GetHandles() []*apitype.Handle {
	images, _ := s.store.GetImages(-1, 0)
	return images
}

func (s *internalManager) ShowOnlyImages(categoryName string, handles []*apitype.Handle) {
	s.imagesTitle = categoryName
	s.index = 0
}

func (s *internalManager) ShowAllImages() {
	s.imagesTitle = ""
}

func (s *internalManager) GenerateHashes(sender api.Sender) bool {
	shouldSendSimilarImages := false
	s.shouldSendSimilar = true
	if s.shouldGenerateSimilarHashed {
		startTime := time.Now()
		images, _ := s.store.GetImagesInCategory(-1, 0, s.imagesTitle)
		hashExpected := len(images)
		logger.Info.Printf("Generate hashes for %d images...", hashExpected)
		sender.SendToTopicWithData(api.ProcessStatusUpdated, "hash", 0, hashExpected)

		// Just to make things consistent in case Go decides to change the default
		cpuCores := s.getTreadCount()
		logger.Info.Printf(" * Using %d threads", cpuCores)
		runtime.GOMAXPROCS(cpuCores)

		s.stopChannel = make(chan bool)
		inputChannel := make(chan *apitype.Handle, hashExpected)
		s.outputChannel = make(chan *HashResult)

		// Add images to input queue for goroutines
		for _, handle := range images {
			inputChannel <- handle
		}

		// Spin up goroutines which will process the data
		// only same number as CPU cores so that we will only max X hashes are
		// processed at once. Otherwise the goroutines might start processing
		// all images at once which would use all available RAM
		for i := 0; i < cpuCores; i++ {
			go hashImage(inputChannel, s.outputChannel, s.stopChannel, s.imageLoader)
		}

		var i = 0
		for result := range s.outputChannel {
			i++
			result.handle.SetHash(result.hash)

			sender.SendToTopicWithData(api.ProcessStatusUpdated, "hash", i, hashExpected)
			s.imageHash.Add(result.handle, *result.hash)

			if i == hashExpected {
				s.StopHashes()
			}
		}
		close(inputChannel)

		endTime := time.Now()
		d := endTime.Sub(startTime)
		logger.Info.Printf("%d hashes created in %s", hashExpected, d.String())

		avg := d.Milliseconds() / int64(hashExpected)
		// Remember to take thread count otherwise the avg time is too small
		f := time.Millisecond * time.Duration(avg) * time.Duration(cpuCores)
		logger.Info.Printf("  On average: %s/image", f.String())

		// Always send 100% status even if cancelled so that the progress bar is hidden
		sender.SendToTopicWithData(api.ProcessStatusUpdated, "hash", hashExpected, hashExpected)
		// Only send if not cancelled
		if i == hashExpected {
			shouldSendSimilarImages = true
		}
		s.shouldGenerateSimilarHashed = false
	} else {
		shouldSendSimilarImages = true
	}

	return shouldSendSimilarImages
}

func (s *internalManager) SetSimilarStatus(sendSimilarImages bool) {
	s.shouldSendSimilar = sendSimilarImages
}

func (s *internalManager) StopHashes() {
	if s.stopChannel != nil {
		for i := 0; i < s.getTreadCount(); i++ {
			s.stopChannel <- true
		}
		close(s.outputChannel)
		close(s.stopChannel)
		s.stopChannel = nil
	}
}

func (s *internalManager) MoveToImage(handle *apitype.Handle) {
	images, _ := s.store.GetImagesInCategory(-1, 0, s.imagesTitle)
	for i, c := range images {
		if handle.GetId() == c.GetId() {
			s.index = i
		}
	}
}

func (s *internalManager) MoveToImageAt(index int) {
	images, _ := s.store.GetImagesInCategory(-1, 0, s.imagesTitle)
	if index >= 0 {
		s.index = index
	} else {
		s.index = len(images) + index
	}

	if s.index >= len(images) {
		s.index = len(images) - 1
	}
	if s.index < 0 {
		s.index = 0
	}
}

func (s *internalManager) RequestNextImage() {
	s.MoveToNextImageWithOffset(1)
}

func (s *internalManager) RequestPrevImage() {
	s.MoveToPrevImageWithOffset(1)
}

func (s *internalManager) MoveToNextImageWithOffset(offset int) {
	s.requestImageWithOffset(offset)
}

func (s *internalManager) MoveToPrevImageWithOffset(offset int) {
	s.requestImageWithOffset(-offset)
}

func (s *internalManager) requestImageWithOffset(offset int) {
	s.index += offset

	images, _ := s.store.GetImagesInCategory(-1, 0, s.imagesTitle)
	if s.index >= len(images) {
		s.index = len(images) - 1
	}
	if s.index < 0 {
		s.index = 0
	}
}

func (s *internalManager) SetImageListSize(imageListSize int) bool {
	if s.imageListSize != imageListSize {
		s.imageListSize = imageListSize
		return true
	} else {
		return false
	}
}

func (s *internalManager) AddHandles(imageList []*apitype.Handle) {
	s.index = 0
}

func (s *internalManager) GetHandleById(handleId int64) *apitype.Handle {
	return s.imageHandles[handleId]
}

func (s *internalManager) GetMetaData(handle *apitype.Handle) *apitype.ExifData {
	return s.imageStore.GetExifData(handle)
}

// Private API

func (s *internalManager) getCurrentImage() (*apitype.ImageContainer, int) {
	images, _ := s.store.GetImagesInCategory(-1, 0, s.imagesTitle)
	if s.index < len(images) {
		handle := images[s.index]
		if full, err := s.imageStore.GetFull(handle); err != nil {
			logger.Error.Print("Error while loading full image", err)
			return apitype.NewImageContainer(apitype.GetEmptyHandle(), nil), 0
		} else {
			return apitype.NewImageContainer(handle, full), s.index
		}
	} else {
		return apitype.NewImageContainer(apitype.GetEmptyHandle(), nil), 0
	}
}

func (s *internalManager) getTotalImages() int {
	return s.store.GetImageCount(s.imagesTitle)
}
func (s *internalManager) getCurrentCategoryName() string {
	return s.imagesTitle
}
func (s *internalManager) shouldSendSimilarImages() bool {
	return s.shouldSendSimilar
}
func (s *internalManager) getImageListSize() int {
	return s.imageListSize
}

func (s *internalManager) getNextImages() []*apitype.ImageContainer {
	imageCount := s.store.GetImageCount(s.imagesTitle)
	startIndex := s.index + 1
	endIndex := startIndex + s.imageListSize
	if endIndex > imageCount {
		endIndex = imageCount
	}

	if startIndex >= imageCount {
		return emptyHandles
	}

	slice, _ := s.store.GetImages(s.imageListSize, startIndex)
	images := make([]*apitype.ImageContainer, len(slice))
	for i, handle := range slice {
		if thumbnail, err := s.imageStore.GetThumbnail(handle); err != nil {
			logger.Error.Print("Error while loading thumbnail", err)
		} else {
			images[i] = apitype.NewImageContainer(handle, thumbnail)
		}
	}

	return images
}

func (s *internalManager) getPrevImages() []*apitype.ImageContainer {
	prevIndex := s.index - s.imageListSize
	if prevIndex < 0 {
		prevIndex = 0
	}
	size := s.index - prevIndex
	slice, _ := s.store.GetImages(size, prevIndex)
	images := make([]*apitype.ImageContainer, len(slice))
	for i, handle := range slice {
		if thumbnail, err := s.imageStore.GetThumbnail(handle); err != nil {
			logger.Error.Print("Error while loading thumbnail", err)
		} else {
			images[i] = apitype.NewImageContainer(handle, thumbnail)
		}
	}
	util.Reverse(images)
	return images
}

func (s *internalManager) loadImagesFromRootDir() {
	handles := apitype.LoadImageHandles(s.rootDir)
	persistedHandles, err := s.store.AddImages(handles)
	if err != nil {
		logger.Error.Fatal("Error while persisting images", err)
	}
	s.AddHandles(persistedHandles)
}

func (s *internalManager) getTreadCount() int {
	cpuCores := runtime.NumCPU()
	return cpuCores
}

func (s *internalManager) getSimilarImages(handle *apitype.Handle) ([]*apitype.ImageContainer, bool) {
	if s.imageHash.Size() > 0 {
		matches := s.imageHash.Query(*handle.GetHash())
		sort.Sort(matches)

		const maxImages = 10
		images := make([]*apitype.ImageContainer, maxImages)
		i := 0
		for _, match := range matches {
			similar := match.ID.(*apitype.Handle)
			if handle.GetId() != similar.GetId() {
				if thumbnail, err := s.imageStore.GetThumbnail(similar); err != nil {
					logger.Error.Print("Error while loading thumbnail", err)
				} else {
					images[i] = apitype.NewImageContainer(similar, thumbnail)
				}
				i++
			}
			if i == maxImages {
				break
			}
		}

		return images, true
	} else {
		return []*apitype.ImageContainer{}, false
	}
}
