package library

import (
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
)

type Library interface {
	RequestImages()
	SetCategory(command *category.CategorizeCommand)
	RequestNextImage()
	RequestPrevImage()
	PersistImageCategories()

}

type ImageCommand struct {
	handles []*common.Handle
	event.Command
}

func (s *ImageCommand) GetHandles() []*common.Handle {
	return s.handles
}
