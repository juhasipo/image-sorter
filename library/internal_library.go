package library

import (
	"runtime"
	"sort"
	"time"
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/duplo"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/imageloader"
	"vincit.fi/image-sorter/logger"
	"vincit.fi/image-sorter/util"
)

var (
	EMPTY_HANDLES = []*common.ImageContainer{}
)

type ImageList func(number int) []*common.Handle

type internalManager struct {
	rootDir                     string
	imageList                   []*common.Handle
	fullImageList               []*common.Handle
	imageHandles                map[string]*common.Handle
	imagesTitle                 string
	index                       int
	imageHash                   *duplo.Store
	shouldSendSimilar           bool
	shouldGenerateSimilarHashed bool
	categoryManager             *category.Manager
	imageListSize               int
	imageStore                  imageloader.ImageStore
	imageLoader                 imageloader.ImageLoader

	stopChannel   chan bool
	outputChannel chan *HashResult
}

func newLibrary(imageCache imageloader.ImageStore, imageLoader imageloader.ImageLoader) *internalManager {
	var manager = internalManager{
		index:                       0,
		imageHash:                   duplo.New(),
		shouldGenerateSimilarHashed: true,
		imageListSize:               0,
		imageStore:                  imageCache,
		imageLoader:                 imageLoader,
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

func (s *internalManager) GetHandles() []*common.Handle {
	return s.imageList
}

func (s *internalManager) ShowOnlyImages(title string, handles []*common.Handle) {
	s.imageList = make([]*common.Handle, len(handles))
	copy(s.imageList, handles)
	sort.Slice(s.imageList, func(i, j int) bool {
		return s.imageList[i].GetId() < s.imageList[j].GetId()
	})
	s.imagesTitle = title
	s.index = 0
}

func (s *internalManager) ShowAllImages() {
	s.imageList = make([]*common.Handle, len(s.fullImageList))
	copy(s.imageList, s.fullImageList)
	s.imagesTitle = ""
}

func (s *internalManager) GenerateHashes(sender event.Sender) bool {
	shouldSendSimilarImages := false
	s.shouldSendSimilar = true
	if s.shouldGenerateSimilarHashed {
		startTime := time.Now()
		hashExpected := len(s.imageList)
		logger.Info.Printf("Generate hashes for %d images...", hashExpected)
		sender.SendToTopicWithData(event.UPDATE_PROCESS_STATUS, "hash", 0, hashExpected)

		// Just to make things consistent in case Go decides to change the default
		cpuCores := s.getTreadCount()
		logger.Info.Printf(" * Using %d threads", cpuCores)
		runtime.GOMAXPROCS(cpuCores)

		s.stopChannel = make(chan bool)
		inputChannel := make(chan *common.Handle, hashExpected)
		s.outputChannel = make(chan *HashResult)

		// Add images to input queue for goroutines
		for _, handle := range s.imageList {
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

			sender.SendToTopicWithData(event.UPDATE_PROCESS_STATUS, "hash", i, hashExpected)
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
		sender.SendToTopicWithData(event.UPDATE_PROCESS_STATUS, "hash", hashExpected, hashExpected)
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

func (s *internalManager) MoveToImage(handle *common.Handle) {
	for i, c := range s.imageList {
		if handle.GetId() == c.GetId() {
			s.index = i
		}
	}
}

func (s *internalManager) MoveToImageAt(index int) {
	if index >= 0 {
		s.index = index
	} else {
		s.index = len(s.imageList) + index
	}

	if s.index >= len(s.imageList) {
		s.index = len(s.imageList) - 1
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

	if s.index >= len(s.imageList) {
		s.index = len(s.imageList) - 1
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

func (s *internalManager) AddHandles(imageList []*common.Handle) {
	s.imageList = imageList
	s.imageHandles = map[string]*common.Handle{}
	s.fullImageList = make([]*common.Handle, len(s.imageList))
	copy(s.fullImageList, s.imageList)

	for _, handle := range s.imageList {
		s.imageHandles[handle.GetId()] = handle
	}

	s.index = 0
}

func (s *internalManager) GetHandleById(handleId string) *common.Handle {
	return s.imageHandles[handleId]
}

func (s *internalManager) GetMetaData(handle *common.Handle) *common.ExifData {
	return s.imageStore.GetExifData(handle)
}

// Private API

func (s *internalManager) getCurrentImage() (*common.ImageContainer, int) {
	if s.index < len(s.imageList) {
		handle := s.imageList[s.index]
		return common.NewImageContainer(handle, s.imageStore.GetFull(handle)), s.index
	} else {
		return common.NewImageContainer(common.GetEmptyHandle(), nil), 0
	}
}
func (s *internalManager) getTotalImages() int {
	return len(s.imageList)
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

func (s *internalManager) getNextImages() []*common.ImageContainer {
	startIndex := s.index + 1
	endIndex := startIndex + s.imageListSize
	if endIndex > len(s.imageList) {
		endIndex = len(s.imageList)
	}

	if startIndex >= len(s.imageList) {
		return EMPTY_HANDLES
	}

	slice := s.imageList[startIndex:endIndex]
	images := make([]*common.ImageContainer, len(slice))
	for i, handle := range slice {
		images[i] = common.NewImageContainer(handle, s.imageStore.GetThumbnail(handle))
	}

	return images
}

func (s *internalManager) getPrevImages() []*common.ImageContainer {
	prevIndex := s.index - s.imageListSize
	if prevIndex < 0 {
		prevIndex = 0
	}
	slice := s.imageList[prevIndex:s.index]
	images := make([]*common.ImageContainer, len(slice))
	for i, handle := range slice {
		images[i] = common.NewImageContainer(handle, s.imageStore.GetThumbnail(handle))
	}
	util.Reverse(images)
	return images
}

func (s *internalManager) loadImagesFromRootDir() {
	s.AddHandles(common.LoadImageHandles(s.rootDir))
}

func (s *internalManager) getTreadCount() int {
	cpuCores := runtime.NumCPU()
	return cpuCores
}

func (s *internalManager) getSimilarImages(handle *common.Handle) ([]*common.ImageContainer, bool) {
	if s.imageHash.Size() > 0 {
		matches := s.imageHash.Query(*handle.GetHash())
		sort.Sort(matches)

		const maxImages = 10
		images := make([]*common.ImageContainer, maxImages)
		i := 0
		for _, match := range matches {
			similar := match.ID.(*common.Handle)
			if handle.GetId() != similar.GetId() {
				images[i] = common.NewImageContainer(similar, s.imageStore.GetThumbnail(similar))
				i++
			}
			if i == maxImages {
				break
			}
		}

		return images, true
	} else {
		return []*common.ImageContainer{}, false
	}
}
