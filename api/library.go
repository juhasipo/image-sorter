package api

import (
	"vincit.fi/image-sorter/api/apitype"
)

type ImagesQuery struct {
	ImageFiles []*apitype.ImageFile
	apitype.Command
}

type ImageQuery struct {
	Id apitype.ImageId
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

type ImageService interface {
	InitializeFromDirectory(directory string)

	RequestImages()
	RequestNextImage()
	RequestPreviousImage()
	RequestNextImageWithOffset(*ImageAtQuery)
	RequestPreviousImageWithOffset(*ImageAtQuery)
	RequestImage(*ImageQuery)
	RequestImageAt(*ImageAtQuery)

	RequestGenerateHashes()
	RequestStopHashes()

	GetImageFiles() []*apitype.ImageFileWithMetaData
	AddImageFiles([]*apitype.ImageFile)
	GetImageFileById(apitype.ImageId) *apitype.ImageFileWithMetaData

	ShowAllImages()
	ShowOnlyImages(*SelectCategoryCommand)

	SetImageListSize(*ImageListCommand)
	SetSendSimilarImages(*SimilarImagesCommand)

	Close()
}
