package api

import (
	"vincit.fi/image-sorter/api/apitype"
)

type Library interface {
	InitializeFromDirectory(directory string)

	RequestImages()
	RequestNextImage()
	RequestPrevImage()
	RequestNextImageWithOffset(int)
	RequestPrevImageWithOffset(int)
	RequestImage(*apitype.Handle)
	RequestImageAt(int)

	RequestGenerateHashes()
	RequestStopHashes()

	GetHandles() []*apitype.Handle
	AddHandles(imageList []*apitype.Handle)
	GetHandleById(handleId apitype.HandleId) *apitype.Handle

	ShowAllImages()
	ShowOnlyImages(string)

	SetImageListSize(imageListSize int)
	SetSendSimilarImages(sendSimilarImages bool)

	Close()
}

type ImageCommand struct {
	handles []*apitype.Handle
	apitype.Command
}

func (s *ImageCommand) GetHandles() []*apitype.Handle {
	return s.handles
}
