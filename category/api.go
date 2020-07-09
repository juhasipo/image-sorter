package category

import "vincit.fi/image-sorter/common"

type CategoryManager interface {
	GetCategories() []*common.Category
	RequestCategories()
	Save(categories []*common.Category)
	SaveDefault(categories []*common.Category)
	Close()
	GetCategoryById(id string) *common.Category
}
