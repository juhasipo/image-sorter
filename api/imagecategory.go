package api

import "vincit.fi/image-sorter/api/apitype"

type ImageCategoryManager interface {
	InitializeForDirectory(directory string)

	RequestCategory(handle *apitype.Handle)
	GetCategories(handle *apitype.Handle) map[string]*apitype.CategorizedImage
	SetCategory(command *apitype.CategorizeCommand)

	PersistImageCategories(apitype.PersistCategorizationCommand)
	PersistImageCategory(handle *apitype.Handle, categories map[string]*apitype.CategorizedImage)

	PersistCategorization()
	LoadCategorization(handleManager Library, categoryManager CategoryManager)

	ShowOnlyCategoryImages(*apitype.Category)

	ResolveFileOperations(map[string]map[string]*apitype.CategorizedImage, apitype.PersistCategorizationCommand) []*apitype.ImageOperationGroup
	ResolveOperationsForGroup(*apitype.Handle, map[string]*apitype.CategorizedImage, apitype.PersistCategorizationCommand) (*apitype.ImageOperationGroup, error)

	Close()
}
