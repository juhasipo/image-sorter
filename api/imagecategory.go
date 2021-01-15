package api

import (
	"time"
	"vincit.fi/image-sorter/api/apitype"
)

type CategorizeCommand struct {
	HandleId        apitype.HandleId
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
	HandleId apitype.HandleId
}

type PersistCategorizationCommand struct {
	KeepOriginals  bool
	FixOrientation bool
	Quality        int
}

type ImageCategoryManager interface {
	InitializeForDirectory(directory string)

	RequestCategory(*ImageCategoryQuery)
	GetCategories(*ImageCategoryQuery) map[apitype.CategoryId]*CategorizedImage
	SetCategory(*CategorizeCommand)

	PersistImageCategories(*PersistCategorizationCommand)
	PersistImageCategory(handle *apitype.Handle, categories map[apitype.CategoryId]*CategorizedImage)

	PersistCategorization()
	LoadCategorization(handleManager Library, categoryManager CategoryManager)

	ShowOnlyCategoryImages(*SelectCategoryCommand)

	ResolveFileOperations(map[apitype.HandleId]map[apitype.CategoryId]*CategorizedImage, *PersistCategorizationCommand) []*apitype.ImageOperationGroup
	ResolveOperationsForGroup(*apitype.Handle, map[apitype.CategoryId]*CategorizedImage, *PersistCategorizationCommand) (*apitype.ImageOperationGroup, error)

	Close()
}
