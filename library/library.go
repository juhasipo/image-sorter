package library

import (
	"log"
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
	imageList     []*common.Handle
	index         int
	imageCategory map[*common.Handle]*category.CategorizedImage
	sender        event.Sender
}

func ForHandles(handles []*common.Handle, sender event.Sender) *Manager {
	var manager = Manager{
		imageList: handles,
		index: 0,
		imageCategory: map[*common.Handle]*category.CategorizedImage{},
		sender: sender,
	}
	return &manager
}

func (s *Manager) SetCategory(command *category.CategorizeCommand) {
	log.Print("Categorize ", command.GetHandle().GetPath(), " as ", command.GetEntry().GetName())
	if val, ok := s.imageCategory[command.GetHandle()]; ok {
		if command.GetOperation() != category.NONE {
			val.SetOperation(command.GetOperation())
		} else {
			delete(s.imageCategory, command.GetHandle())
		}
	} else {
		s.imageCategory[command.GetHandle()] = category.CategorizedImageNew(command.GetEntry(), command.GetOperation())
	}
	s.sender.SendToTopicWithData(
		event.IMAGE_CATEGORIZED,
		category.CategorizeCommandNew(command.GetHandle(), command.GetEntry(), command.GetOperation()))
}

func (s *Manager) NextImage() *common.Handle {
	s.index++
	if s.index >= len(s.imageList) {
		s.index = len(s.imageList) - 1
	}
	s.sender.SendToSubTopicWithData(event.IMAGES_UPDATED, event.NEXT_IMAGE, &ImageCommand{handles: s.GetNextImages(IMAGE_LIST_SIZE)})
	s.sender.SendToSubTopicWithData(event.IMAGES_UPDATED, event.PREV_IMAGE, &ImageCommand{handles: s.GetPrevImages(IMAGE_LIST_SIZE)})
	s.sender.SendToSubTopicWithData(event.IMAGES_UPDATED, event.CURRENT_IMAGE, &ImageCommand{handles: []*common.Handle {s.GetCurrentImage()}})
	return s.GetCurrentImage()
}

func (s *Manager) PrevImage() *common.Handle {
	s.index--
	if s.index < 0 {
		s.index = 0
	}
	s.sender.SendToSubTopicWithData(event.IMAGES_UPDATED, event.NEXT_IMAGE, &ImageCommand{handles: s.GetNextImages(IMAGE_LIST_SIZE)})
	s.sender.SendToSubTopicWithData(event.IMAGES_UPDATED, event.PREV_IMAGE, &ImageCommand{handles: s.GetPrevImages(IMAGE_LIST_SIZE)})
	s.sender.SendToSubTopicWithData(event.IMAGES_UPDATED, event.CURRENT_IMAGE, &ImageCommand{handles: []*common.Handle {s.GetCurrentImage()}})
	return s.GetCurrentImage()
}

func (s *Manager) GetCurrentImage() *common.Handle {
	var currentImage *common.Handle
	if s.index < len(s.imageList) {
		currentImage = s.imageList[s.index]
	} else {
		currentImage = common.GetEmptyHandle()
	}

	s.sender.SendToSubTopicWithData(event.IMAGES_UPDATED, event.NEXT_IMAGE, &ImageCommand{handles: s.GetNextImages(IMAGE_LIST_SIZE)})
	s.sender.SendToSubTopicWithData(event.IMAGES_UPDATED, event.PREV_IMAGE, &ImageCommand{handles: s.GetPrevImages(IMAGE_LIST_SIZE)})
	s.sender.SendToSubTopicWithData(event.IMAGES_UPDATED, event.CURRENT_IMAGE, &ImageCommand{handles: []*common.Handle {currentImage}})

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
