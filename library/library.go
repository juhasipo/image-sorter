package library

import (
	"log"
	"runtime"
	"sort"
	"time"
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/duplo"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/imageloader"
	"vincit.fi/image-sorter/util"
)

var (
	EMPTY_HANDLES []*common.ImageContainer
)

type ImageList func(number int) []*common.Handle

type Manager struct {
	rootDir                     string
	imageList                   []*common.Handle
	fullImageList               []*common.Handle
	imageHandles                map[string]*common.Handle
	imagesTitle                 string
	index                       int
	imageHash                   *duplo.Store
	shouldSendSimilar           bool
	shouldGenerateSimilarHashed bool
	sender                      event.Sender
	categoryManager             *category.Manager
	imageListSize               int
	imageCache                  *imageloader.ImageCache

	Library

	stopChannel   chan bool
	outputChannel chan *HashResult
}

func LibraryNew(sender event.Sender, imageCache *imageloader.ImageCache) Library {
	var manager = Manager{
		index:                       0,
		sender:                      sender,
		imageHash:                   duplo.New(),
		shouldGenerateSimilarHashed: true,
		imageListSize:               0,
		imageCache:                  imageCache,
	}
	return &manager
}

func ReturnResult(channel chan *HashResult, handle *common.Handle, hash *duplo.Hash) {
	channel <- &HashResult{
		handle: handle,
		hash:   hash,
	}
}

func (s *Manager) InitializeFromDirectory(directory string) {
	s.rootDir = directory
	s.index = 0
	s.imageHash = duplo.New()
	s.loadImagesFromRootDir()
}

func (s *Manager) GetHandles() []*common.Handle {
	return s.imageList
}

func (s *Manager) ShowOnlyImages(title string, handles []*common.Handle) {
	s.imageList = handles
	s.imagesTitle = title
	s.index = 0
	s.sendStatus()
}

func (s *Manager) ShowAllImages() {
	s.imageList = make([]*common.Handle, len(s.fullImageList))
	copy(s.imageList, s.fullImageList)
	s.imagesTitle = ""
	s.sendStatus()
}

func (s *Manager) RequestGenerateHashes() {
	s.shouldSendSimilar = true
	if s.shouldGenerateSimilarHashed {
		startTime := time.Now()
		hashExpected := len(s.imageList)
		log.Printf("Generate hashes for %d images...", hashExpected)
		s.sender.SendToTopicWithData(event.UPDATE_PROCESS_STATUS, "hash", 0, hashExpected)

		// Just to make things consistent in case Go decides to change the default
		cpuCores := s.getTreadCount()
		log.Printf(" * Using %d threads", cpuCores)
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
			go hashImage(inputChannel, s.outputChannel, s.stopChannel)
		}

		var i = 0
		for result := range s.outputChannel {
			i++
			result.handle.SetHash(result.hash)

			//log.Printf(" * Got hash %d/%d", i, hashExpected)
			s.sender.SendToTopicWithData(event.UPDATE_PROCESS_STATUS, "hash", i, hashExpected)
			s.imageHash.Add(result.handle, *result.hash)

			if i == hashExpected {
				s.RequestStopHashes()
			}
		}
		close(inputChannel)

		endTime := time.Now()
		d := endTime.Sub(startTime)
		log.Printf("%d hashes created in %s", hashExpected, d.String())

		avg := d.Milliseconds() / int64(hashExpected)
		// Remember to take thread count otherwise the avg time is too small
		f := time.Millisecond * time.Duration(avg) * time.Duration(cpuCores)
		log.Printf("  On average: %s/image", f.String())

		// Always send 100% status even if cancelled so that the progress bar is hidden
		s.sender.SendToTopicWithData(event.UPDATE_PROCESS_STATUS, "hash", hashExpected, hashExpected)
		// Only send if not cancelled
		if i == hashExpected {
			s.sendSimilarImages(s.getCurrentImage().GetHandle())
		}
		s.shouldGenerateSimilarHashed = false
	} else {
		s.sendSimilarImages(s.getCurrentImage().GetHandle())
	}
}

func (s *Manager) SetSimilarStatus(sendSimilarImages bool) {
	s.shouldSendSimilar = sendSimilarImages
}

func (s *Manager) getTreadCount() int {
	cpuCores := runtime.NumCPU()
	return cpuCores
}

func (s *Manager) RequestStopHashes() {
	if s.stopChannel != nil {
		for i := 0; i < s.getTreadCount(); i++ {
			s.stopChannel <- true
		}
		close(s.outputChannel)
		close(s.stopChannel)
		s.stopChannel = nil
	}
}

func (s *Manager) RequestNextImage() {
	s.RequestNextImageWithOffset(1)
}

func (s *Manager) RequestNextImageWithOffset(offset int) {
	s.index += offset
	if s.index >= len(s.imageList) {
		s.index = len(s.imageList) - 1
	}
	if s.index < 0 {
		s.index = 0
	}
	s.sendStatus()
}

func (s *Manager) RequestImage(handle *common.Handle) {
	for i, c := range s.imageList {
		if handle == c {
			s.index = i
		}
	}
	s.RequestImages()
}

func (s *Manager) RequestPrevImage() {
	s.RequestPrevImageWithOffset(1)
}

func (s *Manager) RequestPrevImageWithOffset(offset int) {
	s.index -= offset
	if s.index < 0 {
		s.index = 0
	}
	s.sendStatus()
}

func (s *Manager) RequestImages() {
	s.sendStatus()
}

func (s *Manager) ChangeImageListSize(imageListSize int) {
	if s.imageListSize != imageListSize {
		s.imageListSize = imageListSize
		s.sendStatus()
	}
}

func (s *Manager) Close() {
	log.Print("Shutting down library")
}

// Private API

func (s *Manager) sendStatus() {
	currentImage := s.getCurrentImage()
	s.sender.SendToTopicWithData(event.IMAGE_UPDATE, event.IMAGE_REQUEST_CURRENT, []*common.ImageContainer{currentImage},
		s.index+1, len(s.imageList), s.imagesTitle)

	s.sender.SendToTopicWithData(event.IMAGE_UPDATE, event.IMAGE_REQUEST_NEXT, s.getNextImages(s.imageListSize),
		0, len(s.imageList), "Next")
	s.sender.SendToTopicWithData(event.IMAGE_UPDATE, event.IMAGE_REQUEST_PREV, s.getPrevImages(s.imageListSize),
		0, len(s.imageList), "Previous")

	if s.shouldSendSimilar {
		s.sendSimilarImages(currentImage.GetHandle())
	}
}

func (s *Manager) getCurrentImage() *common.ImageContainer {
	if s.index < len(s.imageList) {
		handle := s.imageList[s.index]
		return common.ImageContainerNew(handle, s.imageCache.GetFull(handle))
	} else {
		return common.ImageContainerNew(common.GetEmptyHandle(), nil)
	}
}

func (s *Manager) getNextImages(number int) []*common.ImageContainer {
	startIndex := s.index + 1
	endIndex := startIndex + number
	if endIndex > len(s.imageList) {
		endIndex = len(s.imageList)
	}

	if startIndex >= len(s.imageList) {
		return EMPTY_HANDLES
	}

	slice := s.imageList[startIndex:endIndex]
	images := make([]*common.ImageContainer, len(slice))
	for i, handle := range slice {
		images[i] = common.ImageContainerNew(handle, s.imageCache.GetThumbnail(handle))
	}
	return images
}

func (s *Manager) getPrevImages(number int) []*common.ImageContainer {
	prevIndex := s.index - number
	if prevIndex < 0 {
		prevIndex = 0
	}
	slice := s.imageList[prevIndex:s.index]
	images := make([]*common.ImageContainer, len(slice))
	for i, handle := range slice {
		images[i] = common.ImageContainerNew(handle, s.imageCache.GetThumbnail(handle))
	}
	util.Reverse(images)
	return images
}

func (s *Manager) sendSimilarImages(handle *common.Handle) {
	if s.imageHash.Size() > 0 {
		matches := s.imageHash.Query(*handle.GetHash())
		sort.Sort(matches)

		const maxImages = 10
		images := make([]*common.ImageContainer, maxImages)
		i := 0
		for _, match := range matches {
			similar := match.ID.(*common.Handle)
			if handle.GetId() != similar.GetId() {
				images[i] = common.ImageContainerNew(similar, s.imageCache.GetThumbnail(similar))
				i++
			}
			if i == maxImages {
				break
			}
		}

		s.sender.SendToTopicWithData(event.IMAGE_UPDATE, event.IMAGE_REQUEST_SIMILAR, images, 0, 0, "")
	}
}

func (s *Manager) loadImagesFromRootDir() {
	s.imageHandles = map[string]*common.Handle{}

	s.imageList = common.LoadImageHandles(s.rootDir)
	s.fullImageList = make([]*common.Handle, len(s.imageList))
	copy(s.fullImageList, s.imageList)

	for _, handle := range s.imageList {
		s.imageHandles[handle.GetId()] = handle
	}

	s.index = 0
}

func (s *Manager) GetHandleById(handleId string) *common.Handle {
	return s.imageHandles[handleId]
}
