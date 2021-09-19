package api

import (
	"time"
	"vincit.fi/image-sorter/api/apitype"
)

type DirectoryChangedCommand struct {
	Directory string

	apitype.NotThrottled
}

type ImagesQuery struct {
	ImageFiles []*apitype.ImageFile

	apitype.NotThrottled
}

type ImageQuery struct {
	Id apitype.ImageId

	apitype.NotThrottled
}
type ImageAtQuery struct {
	Index int

	apitype.NotThrottled
}

type SelectCategoryCommand struct {
	CategoryId apitype.CategoryId

	apitype.NotThrottled
}

type ImageListCommand struct {
	ImageListSize int

	apitype.NotThrottled
}
type SimilarImagesCommand struct {
	SendSimilarImages bool

	apitype.NotThrottled
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
	InitializeFromDirectory(directory string) (time.Time, error)

	AddImageFiles(imageList []*apitype.ImageFile) error

	GetImages() []*apitype.ImageFile
	GetTotalImages(categoryId apitype.CategoryId) int

	GetImagesInCategory(number int, offset int, categoryId apitype.CategoryId) ([]*apitype.ImageFile, error)
	GetImageFileById(imageId apitype.ImageId) *apitype.ImageFile
	GetImageAtIndex(index int, categoryId apitype.CategoryId) (*apitype.ImageFile, *apitype.ImageMetaData, int, error)
	GetNextImages(index int, count int, categoryId apitype.CategoryId) ([]*apitype.ImageFile, error)
	GetPreviousImages(index int, count int, categoryId apitype.CategoryId) ([]*apitype.ImageFile, error)

	GenerateHashes() bool
	GetSimilarImages(imageId apitype.ImageId) ([]*apitype.ImageFile, bool, error)
	StopHashes()
}
