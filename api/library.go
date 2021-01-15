package api

import (
	"vincit.fi/image-sorter/api/apitype"
)

type ImagesQuery struct {
	handles []*apitype.Handle
	apitype.Command
}

type ImageQuery struct {
	Id apitype.HandleId
	apitype.Command
}
type ImageAtQuery struct {
	Index int
	apitype.Command
}

type SelectCategoryCommand struct {
	CategoryId apitype.CategoryId
	apitype.Command
}

type ImageListCommand struct {
	ImageListSize int
	apitype.Command
}
type SimilarImagesCommand struct {
	SendSimilarImages bool
	apitype.Command
}

type Library interface {
	InitializeFromDirectory(directory string)

	RequestImages()
	RequestNextImage()
	RequestPrevImage()
	RequestNextImageWithOffset(*ImageAtQuery)
	RequestPrevImageWithOffset(*ImageAtQuery)
	RequestImage(*ImageQuery)
	RequestImageAt(*ImageAtQuery)

	RequestGenerateHashes()
	RequestStopHashes()

	GetHandles() []*apitype.Handle
	AddHandles(imageList []*apitype.Handle)
	GetHandleById(handleId apitype.HandleId) *apitype.Handle

	ShowAllImages()
	ShowOnlyImages(*SelectCategoryCommand)

	SetImageListSize(*ImageListCommand)
	SetSendSimilarImages(*SimilarImagesCommand)

	Close()
}
