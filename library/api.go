package library

import (
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
)

type Library interface {
	RequestImages()
	SetCategory(*category.CategorizeCommand)
	RequestNextImage()
	RequestNextImageWithOffset(int)
	RequestPrevImage()
	RequestPrevImageWithOffset(int)
	RequestImage(*common.Handle)
	PersistImageCategories()
	RequestGenerateHashes()
	RequestStopHashes()
	GetHandles() []*common.Handle
	Close()
}

type ImageCommand struct {
	handles []*common.Handle
	event.Command
}

func (s *ImageCommand) GetHandles() []*common.Handle {
	return s.handles
}
