package api

import "vincit.fi/image-sorter/api/apitype"

type ImageCategoryManager interface {
	InitializeForDirectory(directory string)

	RequestCategory(handle *apitype.Handle)
	GetCategories(handle *apitype.Handle) map[apitype.CategoryId]*apitype.CategorizedImage
	SetCategory(command *apitype.CategorizeCommand)

	PersistImageCategories(apitype.PersistCategorizationCommand)
	PersistImageCategory(handle *apitype.Handle, categories map[apitype.CategoryId]*apitype.CategorizedImage)

	PersistCategorization()
	LoadCategorization(handleManager Library, categoryManager CategoryManager)

	ShowOnlyCategoryImages(*apitype.Category)

	ResolveFileOperations(map[apitype.HandleId]map[apitype.CategoryId]*apitype.CategorizedImage, apitype.PersistCategorizationCommand) []*apitype.ImageOperationGroup
	ResolveOperationsForGroup(*apitype.Handle, map[apitype.CategoryId]*apitype.CategorizedImage, apitype.PersistCategorizationCommand) (*apitype.ImageOperationGroup, error)

	Close()
}
