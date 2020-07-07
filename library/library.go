package library

import (
	"log"
	"path/filepath"
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

const (
	IMAGE_LIST_SIZE = 5
)

type ImageList func(number int) []*common.Handle

type Manager struct {
	rootDir string
	imageList     []*common.Handle
	index         int
	imageCategory map[*common.Handle]map[*category.Entry]*category.CategorizedImage
	imageHash     *duplo.Store
	sender        event.Sender

	Library
}

func ForHandles(rootDir string, sender event.Sender) Library {
	var manager = Manager{
		rootDir:       rootDir,
		index:         0,
		imageCategory: map[*common.Handle]map[*category.Entry]*category.CategorizedImage{},
		sender:        sender,
		imageHash:     duplo.New(),
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
	s.sender.SendToTopicWithData(event.UPDATE_HASH_STATUS, "hash", 0, hashExpected)

	// Just to make things consistent in case Go decides to change the default
	cpuCores := runtime.NumCPU()
	log.Printf(" * Using %d threads", cpuCores)
	runtime.GOMAXPROCS(cpuCores)

	inputChannel := make(chan *common.Handle, hashExpected)
	outputChannel := make(chan *HashResult)

	// Add images to input queue for goroutines
	for _, handle := range s.imageList {
		inputChannel <- handle
	}

	// Spin up goroutines which will process the data
	// only same number as CPU cores so that we will only max X hashes are
	// processed at once. Otherwise the goroutines might start processing
	// all images at once which would use all available RAM
	for i := 0; i < cpuCores; i++ {
		go hashImage(inputChannel, outputChannel)
	}

	var i = 0
	for result := range outputChannel {
		i++
		result.handle.SetHash(result.hash)

		//log.Printf(" * Got hash %d/%d", i, hashExpected)
		s.sender.SendToTopicWithData(event.UPDATE_HASH_STATUS, "hash", i, hashExpected)
		s.imageHash.Add(result.handle, *result.hash)

		if i == hashExpected {
			close(outputChannel)
			close(inputChannel)
		}
	}

	endTime := time.Now()
	d := endTime.Sub(startTime)
	log.Printf("%d hashes created in %s", hashExpected, d.String())

	avg := d.Milliseconds() / int64(hashExpected)
	// Remember to take thread count otherwise the avg time is too small
	f := time.Millisecond * time.Duration(avg) * time.Duration(cpuCores)
	log.Printf("  On average: %s/image", f.String())

	s.sendSimilarImages(s.getCurrentImage())
}

func (s *Manager) SetCategory(command *category.CategorizeCommand) {
	defer s.sendCategories(command.GetHandle())

	handle := command.GetHandle()
	categoryEntry := command.GetEntry()
	operation := command.GetOperation()

	var image = s.imageCategory[handle]
	var categorizedImage *category.CategorizedImage = nil
	if image != nil {
		categorizedImage = image[categoryEntry]
	}

	if categorizedImage == nil && operation != category.NONE {
		if image == nil {
			log.Printf("Create category entry for '%s'", handle.GetPath())
			image = map[*category.Entry]*category.CategorizedImage{}
			s.imageCategory[handle] = image
		}
		log.Printf("Create category entry for '%s:%s'", handle.GetPath(), categoryEntry.GetName())
		categorizedImage = category.CategorizedImageNew(categoryEntry, operation)
		image[categoryEntry] = categorizedImage
	}

	if operation == category.NONE || categorizedImage == nil {
		log.Printf("Remove entry for '%s:%s'", handle.GetPath(), categoryEntry.GetName())
		delete(s.imageCategory[handle], categoryEntry)
		if len(s.imageCategory[handle]) == 0 {
			log.Printf("Remove entry for '%s'", handle.GetPath())
			delete(s.imageCategory, handle)
		}
	} else {
		log.Printf("Update entry for '%s:%s' to %d", handle.GetPath(), categoryEntry.GetName(), operation)
		categorizedImage.SetOperation(operation)
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
	s.sendImages()
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
	s.sendImages()
}

func (s *Manager) RequestImages() {
	s.sendImages()
}

func (s* Manager) PersistImageCategories() {
	log.Printf("Persisting files to categories")
	for handle, categoryEntries := range s.imageCategory {
		s.PersistImageCategory(handle, categoryEntries)
	}

	s.loadImagesFromRootDir()

	s.sendImages()
}

func (s* Manager) PersistImageCategory(handle *common.Handle, categories map[*category.Entry]*category.CategorizedImage) {
	log.Printf(" - Persisting '%s'", handle.GetPath())
	dir, file := filepath.Split(handle.GetPath())

	var hasMove = false
	for _, image := range categories {
		targetDirName := image.GetEntry().GetSubPath()
		targetDir := filepath.Join(dir, targetDirName)

		// Always copy
		if image.GetOperation() != category.NONE {
			common.CopyFile(dir, file, targetDir, file)
		}

		// Check if any one is marked to be moved, in that case delete it later
		if image.GetOperation() == category.MOVE {
			hasMove = true
		}
	}
	if hasMove {
		common.RemoveFile(handle.GetPath())
	}
}

// Private API

func (s *Manager) getCategories(image *common.Handle) []*category.CategorizedImage {
	var categories []*category.CategorizedImage

	if i, ok := s.imageCategory[image]; ok {
		for _, categorizedImage := range i {
			categories = append(categories, categorizedImage)
		}
	}

	return categories
}

func (s *Manager) sendImages() {
	s.sender.SendToTopicWithData(event.IMAGES_UPDATED, event.NEXT_IMAGE, s.getNextImages(IMAGE_LIST_SIZE))
	s.sender.SendToTopicWithData(event.IMAGES_UPDATED, event.PREV_IMAGE, s.getPrevImages(IMAGE_LIST_SIZE))
	currentImage := s.getCurrentImage()
	s.sender.SendToTopicWithData(event.IMAGES_UPDATED, event.CURRENT_IMAGE, []*common.Handle{currentImage})
	s.sendCategories(currentImage)
	s.sendSimilarImages(currentImage)
}

func (s *Manager) sendCategories(currentImage *common.Handle) {
	var categories = s.getCategories(currentImage)
	var commands []*category.CategorizeCommand
	for _, image := range categories {
		commands = append(commands, category.CategorizeCommandNew(currentImage, image.GetEntry(), image.GetOperation()))
	}
	s.sender.SendToTopicWithData(event.IMAGE_CATEGORIZED, commands)
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

		s.sender.SendToTopicWithData(event.IMAGES_UPDATED, event.SIMILAR_IMAGE, found)
	}
}

func (s *Manager) loadImagesFromRootDir() {
	log.Printf("Loading images from '%s'", s.rootDir)
	s.imageList = common.LoadImages(s.rootDir)
	// Remove non existing files from the categories in case they have been moved
	for _, handle := range s.imageList {
		if _, ok := s.imageCategory[handle]; !ok {
			delete(s.imageCategory, handle)
		}
	}
	s.index = 0
}
