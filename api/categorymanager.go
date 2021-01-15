package api

import "vincit.fi/image-sorter/api/apitype"

type CategoryQuery struct {
	Id apitype.CategoryId
}

type SaveCategoriesCommand struct {
	Categories []*apitype.Category
}

type CategoryManager interface {
	InitializeFromDirectory(categories []string, rootDir string)
	GetCategories() []*apitype.Category
	RequestCategories()
	Save(*SaveCategoriesCommand)
	SaveDefault(*SaveCategoriesCommand)
	Close()
	GetCategoryById(*CategoryQuery) *apitype.Category
}
