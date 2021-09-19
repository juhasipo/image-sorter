package api

import "vincit.fi/image-sorter/api/apitype"

type CategoryQuery struct {
	Id apitype.CategoryId
	apitype.NotThrottled
}

type SaveCategoriesCommand struct {
	Categories []*apitype.Category
	apitype.NotThrottled
}

type CategoryService interface {
	InitializeFromDirectory(cmdLineCategories []string, dbCategories []*apitype.Category)
	GetCategories() []*apitype.Category
	RequestCategories()
	Save(*SaveCategoriesCommand)
	Close()
	GetCategoryById(*CategoryQuery) *apitype.Category
}
