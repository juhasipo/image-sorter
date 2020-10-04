package api

import "vincit.fi/image-sorter/api/apitype"

type CategoryManager interface {
	InitializeFromDirectory(categories []string, rootDir string)
	GetCategories() []*apitype.Category
	RequestCategories()
	Save(categories []*apitype.Category)
	SaveDefault(categories []*apitype.Category)
	Close()
	GetCategoryById(id string) *apitype.Category
}
