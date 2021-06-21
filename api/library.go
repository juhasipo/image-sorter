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

	GetImageFiles() []*apitype.ImageFile
	AddImageFiles([]*apitype.ImageFile)
	GetImageFileById(apitype.ImageId) *apitype.ImageFile

	ShowAllImages()
	ShowOnlyImages(*SelectCategoryCommand)

	SetImageListSize(*ImageListCommand)
	SetSendSimilarImages(*SimilarImagesCommand)

	Close()
}

type ImageLibrary interface {
	InitializeFromDirectory(directory string, sender Sender)

	AddImageFiles(imageList []*apitype.ImageFile, sender Sender) error

	GetImages() []*apitype.ImageFile
	GetTotalImages(categoryId apitype.CategoryId) int

	GetImagesInCategory(number int, offset int, categoryId apitype.CategoryId) ([]*apitype.ImageFile, error)
	GetImageFileById(imageId apitype.ImageId) *apitype.ImageFile
	GetImageAtIndex(index int, categoryId apitype.CategoryId) (*apitype.ImageFileAndData, *apitype.ImageMetaData, int, error)
	GetNextImages(index int, count int, categoryId apitype.CategoryId) ([]*apitype.ImageFileAndData, error)
	GetPreviousImages(index int, count int, categoryId apitype.CategoryId) ([]*apitype.ImageFileAndData, error)

	GenerateHashes(sender Sender) bool
	GetSimilarImages(imageId apitype.ImageId) ([]*apitype.ImageFileAndData, bool, error)
	StopHashes()
}
