package library

import "vincit.fi/image-sorter/common"

type Manager struct {
	imageList []*common.Handle
	index int
}

func ForHandles(handles []*common.Handle) *Manager {
	var manager = Manager{
		imageList: handles,
		index: 0,
	}
	return &manager
}

func (s *Manager) NextImage() *common.Handle {
	s.index++
	if s.index >= len(s.imageList) {
		s.index = len(s.imageList) - 1
	}
	return s.GetCurrentImage()
}

func (s *Manager) PrevImage() *common.Handle {
	s.index--
	if s.index < 0 {
		s.index = 0
	}
	return s.GetCurrentImage()
}

func (s *Manager) GetCurrentImage() *common.Handle {
	if s.index < len(s.imageList) {
		return s.imageList[s.index]
	} else {
		return common.GetEmptyHandle()
	}
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

	return s.imageList[startIndex:endIndex]
}

func (s* Manager) GetPrevImages(number int) []*common.Handle {
	prevIndex := s.index-number
	if prevIndex < 0 {
		prevIndex = 0
	}
	return s.imageList[prevIndex:s.index]
}

