package library

import (
	"log"
	"path/filepath"
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
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
	sender        event.Sender

	Library
}

func ForHandles(rootDir string, sender event.Sender) Library {
	var manager = Manager{
		rootDir: rootDir,
		index: 0,
		imageCategory: map[*common.Handle]map[*category.Entry]*category.CategorizedImage{},
		sender: sender,
	}
	manager.LoadImagesFromRootDir()
	return &manager
}

func (s *Manager) LoadImagesFromRootDir() {
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

func (s *Manager) SetCategory(command *category.CategorizeCommand) {
	defer s.SendCategories(command.GetHandle())

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
	s.index++
	if s.index >= len(s.imageList) {
		s.index = len(s.imageList) - 1
	}
	s.SendImages()
}

func (s *Manager) SendImages() {
	s.sender.SendToTopicWithData(event.IMAGES_UPDATED, event.NEXT_IMAGE, s.GetNextImages(IMAGE_LIST_SIZE))
	s.sender.SendToTopicWithData(event.IMAGES_UPDATED, event.PREV_IMAGE, s.GetPrevImages(IMAGE_LIST_SIZE))
	currentImage := s.GetCurrentImage()
	s.sender.SendToTopicWithData(event.IMAGES_UPDATED, event.CURRENT_IMAGE, []*common.Handle{currentImage})
	s.SendCategories(currentImage)
}

func (s *Manager) SendCategories(currentImage *common.Handle) {
	var categories = s.GetCategories(currentImage)
	var commands []*category.CategorizeCommand
	for _, image := range categories {
		commands = append(commands, category.CategorizeCommandNew(currentImage, image.GetEntry(), image.GetOperation()))
	}
	s.sender.SendToTopicWithData(event.IMAGE_CATEGORIZED, commands)
}


func (s *Manager) RequestPrevImage() {
	s.index--
	if s.index < 0 {
		s.index = 0
	}
	s.SendImages()
}

func (s *Manager) RequestImages() {
	s.SendImages()
}

func (s *Manager) GetCurrentImage() *common.Handle {
	var currentImage *common.Handle
	if s.index < len(s.imageList) {
		currentImage = s.imageList[s.index]
	} else {
		currentImage = common.GetEmptyHandle()
	}
	return currentImage
}

func (s* Manager) GetNextImages(number int) []*common.Handle {
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

func (s* Manager) GetPrevImages(number int) []*common.Handle {
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

func (s* Manager) PersistImageCategories() {
	log.Printf("Persisting files to categories")
	for handle, categoryEntries := range s.imageCategory {
		s.PersistImageCategory(handle, categoryEntries)
	}

	s.LoadImagesFromRootDir()

	s.SendImages()
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

func (s *Manager) GetCategories(image *common.Handle) []*category.CategorizedImage {
	var categories []*category.CategorizedImage

	if i, ok := s.imageCategory[image]; ok {
		for _, categorizedImage := range i {
			categories = append(categories, categorizedImage)
		}
	}

	return categories
}
