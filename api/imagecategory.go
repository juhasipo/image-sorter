package api

import "vincit.fi/image-sorter/api/apitype"

type ImageCategoryManager interface {
	InitializeForDirectory(directory string)

	RequestCategory(handle *apitype.Handle)
	GetCategories(handle *apitype.Handle) map[int64]*apitype.CategorizedImage
	SetCategory(command *apitype.CategorizeCommand)

	PersistImageCategories(apitype.PersistCategorizationCommand)
	PersistImageCategory(handle *apitype.Handle, categories map[int64]*apitype.CategorizedImage)

	PersistCategorization()
	LoadCategorization(handleManager Library, categoryManager CategoryManager)

	ShowOnlyCategoryImages(*apitype.Category)

	ResolveFileOperations(map[int64]map[int64]*apitype.CategorizedImage, apitype.PersistCategorizationCommand) []*apitype.ImageOperationGroup
	ResolveOperationsForGroup(*apitype.Handle, map[int64]*apitype.CategorizedImage, apitype.PersistCategorizationCommand) (*apitype.ImageOperationGroup, error)

	Close()
}
