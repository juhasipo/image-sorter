package library

import (
	"github.com/upper/db/v4"
	"runtime"
	"sort"
	"sync"
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

const maxSimilarImages = 20

type ImageList func(number int) []*apitype.Handle

type internalManager struct {
	rootDir                     string
	imagesTitle                 string
	index                       int
	imageHash                   *duplo.Store
	shouldSendSimilar           bool
	shouldGenerateSimilarHashed bool
	categoryManager             api.CategoryManager
	imageListSize               int
	imageCache                  api.ImageStore
	imageLoader                 api.ImageLoader
	similarityIndex             *database.SimilarityIndex
	imageStore                  *database.ImageStore

	stopChannel   chan bool
	outputChannel chan *HashResult
}

func newLibrary(imageCache api.ImageStore, imageLoader api.ImageLoader,
	similarityIndex *database.SimilarityIndex, imageStore *database.ImageStore) *internalManager {
	var manager = internalManager{
		index:                       0,
		imageHash:                   duplo.New(),
		shouldGenerateSimilarHashed: true,
		shouldSendSimilar:           true,
		imageListSize:               0,
		imageCache:                  imageCache,
		imageLoader:                 imageLoader,
		similarityIndex:             similarityIndex,
		imageStore:                  imageStore,
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
	images, _ := s.imageStore.GetImages(-1, 0)
	return images
}

func (s *internalManager) ShowOnlyImages(categoryName string) {
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
		images, _ := s.imageStore.GetImages(-1, 0)
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

		hashes := map[apitype.HandleId]*duplo.Hash{}

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

		var mux sync.Mutex

		var i = 0
		for result := range s.outputChannel {
			i++
			s.addHashToMap(result, hashes, &mux)

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

		logger.Info.Printf("Building similarity index for %d most similar images for each image", maxSimilarImages)

		startTime = time.Now()

		s.similarityIndex.DoInTransaction(func(session db.Session) error {
			if err := s.similarityIndex.StartRecreateSimilarImageIndex(session); err != nil {
				logger.Error.Print("Error while clearing similar images", err)
				return err
			}
			sender.SendToTopicWithData(api.ProcessStatusUpdated, "similarity-index", 0, len(images))
			for imageIndex, handle := range images {
				hash := hashes[handle.GetId()]
				searchStart := time.Now()
				matches := s.imageHash.Query(*hash)
				searchEnd := time.Now()

				sortStart := time.Now()
				sort.Sort(matches)
				sortEnd := time.Now()

				addStart := time.Now()
				i := 0
				for _, match := range matches {
					similar := match.ID.(*apitype.Handle)
					if handle.GetId() != similar.GetId() {
						if err := s.similarityIndex.
							AddSimilarImage(handle.GetId(), similar.GetId(), i, match.Score); err != nil {
							logger.Error.Print("Error while storing similar images", err)
							return err
						}
						i++
					}
					if i == maxSimilarImages {
						break
					}
				}
				addEnd := time.Now()

				sender.SendToTopicWithData(api.ProcessStatusUpdated, "hash", imageIndex, len(images))

				logger.Debug.Printf("Print added matches for image: %s", handle.String())
				logger.Debug.Printf(" -    Search: %s", searchEnd.Sub(searchStart))
				logger.Debug.Printf(" -      Sort: %s", sortEnd.Sub(sortStart))
				logger.Debug.Printf(" -       Add: %s", addEnd.Sub(addStart))
				logger.Debug.Printf(" - Add/image: %s", addEnd.Sub(addStart)/maxSimilarImages)
			}
			if err := s.similarityIndex.EndRecreateSimilarImageIndex(); err != nil {
				logger.Error.Print("Error while finishing similar images index", err)
				return err
			}

			return nil
		})
		endTime = time.Now()
		d = endTime.Sub(startTime)
		logger.Info.Printf("Similarity index has been built in %s", d.String())

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

func (s *internalManager) addHashToMap(result *HashResult, hashes map[apitype.HandleId]*duplo.Hash, mux *sync.Mutex) {
	mux.Lock()
	defer mux.Unlock()
	hashes[result.handle.GetId()] = result.hash
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
	images, _ := s.imageStore.GetImagesInCategory(-1, 0, s.imagesTitle)
	for imageIndex, image := range images {
		if handle.GetId() == image.GetId() {
			s.index = imageIndex
		}
	}
}

func (s *internalManager) MoveToImageAt(index int) {
	images, _ := s.imageStore.GetImagesInCategory(-1, 0, s.imagesTitle)
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

	images, _ := s.imageStore.GetImagesInCategory(-1, 0, s.imagesTitle)
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
	start := time.Now()
	if err := s.imageStore.AddImages(imageList); err != nil {
		logger.Error.Print("cannot add images", err)
	}
	end := time.Now()

	imageCount := len(imageList)
	duration := end.Sub(start)
	avg := duration / time.Duration(imageCount)
	logger.Debug.Printf("Added %d images in %s (avg. %s/image)", imageCount, duration, avg)
}

func (s *internalManager) GetHandleById(handleId apitype.HandleId) *apitype.Handle {
	return s.imageStore.GetImageById(handleId)
}

func (s *internalManager) GetMetaData(handle *apitype.Handle) *apitype.ExifData {
	return s.imageCache.GetExifData(handle)
}

// Private API

func (s *internalManager) getCurrentImage() (*apitype.ImageContainer, int) {
	images, _ := s.imageStore.GetImagesInCategory(-1, 0, s.imagesTitle)
	if s.index < len(images) {
		handle := images[s.index]
		if full, err := s.imageCache.GetFull(handle); err != nil {
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
	return s.imageStore.GetImageCount(s.imagesTitle)
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
	imageCount := s.imageStore.GetImageCount(s.imagesTitle)
	startIndex := s.index + 1
	endIndex := startIndex + s.imageListSize
	if endIndex > imageCount {
		endIndex = imageCount
	}

	if startIndex >= imageCount {
		return emptyHandles
	}

	slice, _ := s.imageStore.GetImagesInCategory(s.imageListSize, startIndex, s.imagesTitle)
	images := make([]*apitype.ImageContainer, len(slice))
	for i, handle := range slice {
		if thumbnail, err := s.imageCache.GetThumbnail(handle); err != nil {
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
	slice, _ := s.imageStore.GetImagesInCategory(size, prevIndex, s.imagesTitle)
	images := make([]*apitype.ImageContainer, len(slice))
	for i, handle := range slice {
		if thumbnail, err := s.imageCache.GetThumbnail(handle); err != nil {
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
	s.AddHandles(handles)
}

func (s *internalManager) getTreadCount() int {
	cpuCores := runtime.NumCPU()
	return cpuCores
}

func (s *internalManager) getSimilarImages(handle *apitype.Handle) ([]*apitype.ImageContainer, bool) {
	similarImages := s.similarityIndex.GetSimilarImages(handle.GetId())
	if len(similarImages) > 0 {
		containers := make([]*apitype.ImageContainer, len(similarImages))
		i := 0
		for _, similar := range similarImages {
			if thumbnail, err := s.imageCache.GetThumbnail(similar); err != nil {
				logger.Error.Print("Error while loading thumbnail", err)
			} else {
				containers[i] = apitype.NewImageContainer(similar, thumbnail)
			}
			i++
		}

		return containers, true
	} else {
		return []*apitype.ImageContainer{}, false
	}
}
