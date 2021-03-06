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

	apitype.Command
}

type CategoriesCommand struct {
	Categories []*apitype.Category

	apitype.Command
}

type CategorizedImage struct {
	Category  *apitype.Category
	Operation apitype.Operation
}

type ImageCategoryQuery struct {
	ImageId apitype.ImageId
}

type PersistCategorizationCommand struct {
	KeepOriginals  bool
	FixOrientation bool
	Quality        int
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

	ResolveFileOperations(map[apitype.ImageId]map[apitype.CategoryId]*CategorizedImage, *PersistCategorizationCommand) []*apitype.ImageOperationGroup
	ResolveOperationsForGroup(*apitype.ImageFile, map[apitype.CategoryId]*CategorizedImage, *PersistCategorizationCommand) (*apitype.ImageOperationGroup, error)

	Close()
}
