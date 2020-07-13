package imagecategory

import (
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/library"
)

type ImageCategoryManager interface {
	RequestCategory(handle *common.Handle)
	SetCategory(command *category.CategorizeCommand)

	PersistImageCategories()
	PersistImageCategory(handle *common.Handle, categories map[string]*category.CategorizedImage)

	PersistCategorization()
	LoadCategorization(handleManager library.Library, categoryManager category.CategoryManager)

	Close()
}