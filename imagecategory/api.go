package imagecategory

import (
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/filter"
	"vincit.fi/image-sorter/library"
)

type ImageCategoryManager interface {
	InitializeForDirectory(directory string)

	RequestCategory(handle *common.Handle)
	GetCategories(handle *common.Handle) map[string]*category.CategorizedImage
	SetCategory(command *category.CategorizeCommand)

	PersistImageCategories(common.PersistCategorizationCommand)
	PersistImageCategory(handle *common.Handle, categories map[string]*category.CategorizedImage)

	PersistCategorization()
	LoadCategorization(handleManager library.Library, categoryManager category.CategoryManager)

	ShowOnlyCategoryImages(*common.Category)

	ResolveFileOperations(map[string]map[string]*category.CategorizedImage, common.PersistCategorizationCommand) []*filter.ImageOperationGroup
	ResolveOperationsForGroup(*common.Handle, map[string]*category.CategorizedImage, common.PersistCategorizationCommand) (*filter.ImageOperationGroup, error)

	Close()
}
