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
	"vincit.fi/image-sorter/util"
)

var (
	EMPTY_HANDLES []*common.Handle
)

type ImageList func(number int) []*common.Handle

type Manager struct {
	rootDir       string
	imageList     []*common.Handle
	imageHandles  map[string]*common.Handle
	index         int
	imageHash     *duplo.Store
	sender        event.Sender
	categoryManager *category.Manager
	imageListSize int

	Library

	stopChannel     chan bool
	outputChannel   chan *HashResult
}

func ForHandles(rootDir string, sender event.Sender) Library {
	var manager = Manager{
		rootDir:       rootDir,
		index:         0,
		sender:        sender,
		imageHash:     duplo.New(),
		imageListSize: 5,
	}
	manager.loadImagesFromRootDir()
	return &manager
}

func ReturnResult(channel chan *HashResult, handle *common.Handle, hash *duplo.Hash) {
	channel <-&HashResult{
		handle: handle,
		hash:   hash,
	}
}

func (s *Manager) GetHandles() []*common.Handle {
	return s.imageList
}

func (s *Manager) RequestGenerateHashes() {
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
		s.sendSimilarImages(s.getCurrentImage())
	}
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
	s.sender.SendToTopicWithData(event.IMAGE_UPDATE, event.IMAGE_REQUEST_NEXT, s.getNextImages(s.imageListSize))
	s.sender.SendToTopicWithData(event.IMAGE_UPDATE, event.IMAGE_REQUEST_PREV, s.getPrevImages(s.imageListSize))
	currentImage := s.getCurrentImage()
	s.sender.SendToTopicWithData(event.IMAGE_UPDATE, event.IMAGE_REQUEST_CURRENT, []*common.Handle{currentImage})
	s.sendSimilarImages(currentImage)
}

func (s *Manager) getCurrentImage() *common.Handle {
	var currentImage *common.Handle
	if s.index < len(s.imageList) {
		currentImage = s.imageList[s.index]
	} else {
		currentImage = common.GetEmptyHandle()
	}
	return currentImage
}

func (s* Manager) getNextImages(number int) []*common.Handle {
	startIndex := s.index + 1
	endIndex := startIndex + number
	if endIndex > len(s.imageList) {
		endIndex = len(s.imageList)
	}

	if startIndex >= len(s.imageList) - 1 {
		return EMPTY_HANDLES
	}

	slice := s.imageList[startIndex:endIndex]
	arr := make([]*common.Handle, len(slice))
	copy(arr[:], slice)
	return arr
}

func (s* Manager) getPrevImages(number int) []*common.Handle {
	prevIndex := s.index-number
	if prevIndex < 0 {
		prevIndex = 0
	}
	slice := s.imageList[prevIndex:s.index]
	arr := make([]*common.Handle, len(slice))
	copy(arr[:], slice)
	util.Reverse(arr)
	return arr
}


func (s *Manager) sendSimilarImages(handle *common.Handle) {
	if s.imageHash.Size() > 0 {
		matches := s.imageHash.Query(*handle.GetHash())
		sort.Sort(matches)

		var found []*common.Handle
		for _, match := range matches {
			similar := match.ID.(*common.Handle)
			if handle != similar {
				found = append(found, similar)
			}
			if len(found) == 10 {
				break
			}
		}

		s.sender.SendToTopicWithData(event.IMAGE_UPDATE, event.IMAGE_REQUEST_SIMILAR, found)
	}
}

func (s *Manager) loadImagesFromRootDir() {
	s.imageHandles = map[string]*common.Handle{}

	s.imageList = LoadImages(s.rootDir)

	for _, handle := range s.imageList {
		s.imageHandles[handle.GetId()] = handle
	}

	s.index = 0
}

func (s *Manager) GetHandleById(handleId string) *common.Handle {
	return s.imageHandles[handleId]
}
