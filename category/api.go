package category

import "vincit.fi/image-sorter/common"

type CategoryManager interface {
	InitializeFromDirectory(categories []string, rootDir string)
	GetCategories() []*common.Category
	RequestCategories()
	Save(categories []*common.Category)
	SaveDefault(categories []*common.Category)
	Close()
	GetCategoryById(id string) *common.Category
}
