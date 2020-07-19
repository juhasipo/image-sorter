package library

import (
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
)

type Library interface {
	InitializeFromDirectory(directory string)
	RequestImages()
	RequestNextImage()
	RequestNextImageWithOffset(int)
	RequestPrevImage()
	RequestPrevImageWithOffset(int)
	RequestImage(*common.Handle)
	RequestGenerateHashes()
	RequestStopHashes()
	GetHandles() []*common.Handle
	GetHandleById(handleId string) *common.Handle
	ChangeImageListSize(imageListSize int)
	SetSimilarStatus(sendSimilarImages bool)

	ShowAllImages()
	ShowOnlyImages([]*common.Handle)

	Close()
}

type ImageCommand struct {
	handles []*common.Handle
	event.Command
}

func (s *ImageCommand) GetHandles() []*common.Handle {
	return s.handles
}
