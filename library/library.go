package library

import (
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/util"
)

type Manager struct {
	imageList []*common.Handle
	index int
	imageCategory map[*common.Handle]*category.CategorizedImage
	sender event.Sender
}

const (
	IMAGE_LIST_SIZE = 5
)

func ForHandles(handles []*common.Handle, sender *event.Broker) *Manager {
	var manager = Manager{
		imageList: handles,
		index: 0,
		imageCategory: map[*common.Handle]*category.CategorizedImage{},
		sender: sender,
	}
	return &manager
}

func (s *Manager) SetCategory(handle *common.Handle, entry *category.Entry, operation category.Operation) {
	if val, ok := s.imageCategory[handle]; ok {
		if operation != category.NONE {
			val.SetOperation(operation)
		} else {
			delete(s.imageCategory, handle)
		}
	} else {
		s.imageCategory[handle] = category.CategorizedImageNew(entry, operation)
	}
}

func (s *Manager) NextImage() *common.Handle {
	s.index++
	if s.index >= len(s.imageList) {
		s.index = len(s.imageList) - 1
	}
	s.sender.Send(event.NewWithSubAndData(event.IMAGES_UPDATED, event.NEXT_IMAGE, s.GetNextImages(IMAGE_LIST_SIZE)))
	s.sender.Send(event.NewWithSubAndData(event.IMAGES_UPDATED, event.PREV_IMAGE, s.GetPrevImages(IMAGE_LIST_SIZE)))
	s.sender.Send(event.NewWithSubAndData(event.IMAGES_UPDATED, event.CURRENT_IMAGE, []*common.Handle {s.GetCurrentImage()}))
	return s.GetCurrentImage()
}

func (s *Manager) PrevImage() *common.Handle {
	s.index--
	if s.index < 0 {
		s.index = 0
	}
	s.sender.Send(event.NewWithSubAndData(event.IMAGES_UPDATED, event.NEXT_IMAGE, s.GetNextImages(IMAGE_LIST_SIZE)))
	s.sender.Send(event.NewWithSubAndData(event.IMAGES_UPDATED, event.PREV_IMAGE, s.GetPrevImages(IMAGE_LIST_SIZE)))
	s.sender.Send(event.NewWithSubAndData(event.IMAGES_UPDATED, event.CURRENT_IMAGE, []*common.Handle {s.GetCurrentImage()}))
	return s.GetCurrentImage()
}

func (s *Manager) GetCurrentImage() *common.Handle {
	var currentImage *common.Handle
	if s.index < len(s.imageList) {
		currentImage = s.imageList[s.index]
	} else {
		currentImage = common.GetEmptyHandle()
	}

	s.sender.Send(event.NewWithSubAndData(event.IMAGES_UPDATED, event.NEXT_IMAGE, s.GetNextImages(IMAGE_LIST_SIZE)))
	s.sender.Send(event.NewWithSubAndData(event.IMAGES_UPDATED, event.PREV_IMAGE, s.GetPrevImages(IMAGE_LIST_SIZE)))
	s.sender.Send(event.NewWithSubAndData(event.IMAGES_UPDATED, event.CURRENT_IMAGE, []*common.Handle {currentImage}))

	return currentImage
}


var (
	EMPTY_HANDLES []*common.Handle
)

type ImageList func(number int) []*common.Handle

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
