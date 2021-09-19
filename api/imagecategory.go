package api

import (
	"time"
	"vincit.fi/image-sorter/api/apitype"
)

type CategorizeCommand struct {
	ImageId         apitype.ImageId
	CategoryId      apitype.CategoryId
	Operation       apitype.Operation
	StayOnSameImage bool
	NextImageDelay  time.Duration
	ForceToCategory bool

	apitype.NotThrottled
}

type CategoriesCommand struct {
	Categories []*apitype.Category

	apitype.NotThrottled
}

type CategorizedImage struct {
	Category  *apitype.Category
	Operation apitype.Operation

	apitype.NotThrottled
}

type ImageCategoryQuery struct {
	ImageId apitype.ImageId

	apitype.NotThrottled
}

type PersistCategorizationCommand struct {
	KeepOriginals  bool
	FixOrientation bool
	Quality        int

	apitype.NotThrottled
}

type ImageCategoryService interface {
	InitializeForDirectory(directory string)

	RequestCategory(*ImageCategoryQuery)
	GetCategories(*ImageCategoryQuery) map[apitype.CategoryId]*CategorizedImage
	SetCategory(*CategorizeCommand)

	PersistImageCategories(*PersistCategorizationCommand)
	PersistImageCategory(*apitype.ImageFile, map[apitype.CategoryId]*CategorizedImage)

	PersistCategorization()
	LoadCategorization(ImageService, CategoryService)

	ShowOnlyCategoryImages(*SelectCategoryCommand)

	ResolveFileOperations(map[apitype.ImageId]map[apitype.CategoryId]*CategorizedImage, *PersistCategorizationCommand, func(current int, total int)) []*apitype.ImageOperationGroup
	ResolveOperationsForGroup(*apitype.ImageFile, map[apitype.CategoryId]*CategorizedImage, *PersistCategorizationCommand) (*apitype.ImageOperationGroup, error)

	Close()
}
