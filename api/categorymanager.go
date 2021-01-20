package api

import "vincit.fi/image-sorter/api/apitype"

type CategoryQuery struct {
	Id apitype.CategoryId
}

type SaveCategoriesCommand struct {
	Categories []*apitype.Category
}

type CategoryManager interface {
	InitializeFromDirectory(cmdLineCategories []string, dbCategories []*apitype.Category)
	GetCategories() []*apitype.Category
	RequestCategories()
	Save(*SaveCategoriesCommand)
	Close()
	GetCategoryById(*CategoryQuery) *apitype.Category
}
