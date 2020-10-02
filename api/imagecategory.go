package api

import (
	"vincit.fi/image-sorter/common"
)

type ImageCategoryManager interface {
	InitializeForDirectory(directory string)

	RequestCategory(handle *common.Handle)
	GetCategories(handle *common.Handle) map[string]*CategorizedImage
	SetCategory(command *CategorizeCommand)

	PersistImageCategories(common.PersistCategorizationCommand)
	PersistImageCategory(handle *common.Handle, categories map[string]*CategorizedImage)

	PersistCategorization()
	LoadCategorization(handleManager Library, categoryManager CategoryManager)

	ShowOnlyCategoryImages(*common.Category)

	ResolveFileOperations(map[string]map[string]*CategorizedImage, common.PersistCategorizationCommand) []*ImageOperationGroup
	ResolveOperationsForGroup(*common.Handle, map[string]*CategorizedImage, common.PersistCategorizationCommand) (*ImageOperationGroup, error)

	Close()
}
