package database

import "vincit.fi/image-sorter/api/apitype"

func idToHandleId(id interface{}) apitype.HandleId {
	return apitype.HandleId(id.(int64))
}

func idToCategoryId(id interface{}) apitype.CategoryId {
	return apitype.CategoryId(id.(int64))
}

func toApiHandle(image *Image) *apitype.Handle {
	return apitype.NewHandleWithId(
		image.Id, image.Directory, image.FileName,
	)
}

func toApiHandles(images []Image) []*apitype.Handle {
	handles := make([]*apitype.Handle, len(images))
	for i, image := range images {
		handles[i] = toApiHandle(&image)
	}
	return handles
}

func toApiCategorizedImages(categories []CategorizedImage) []*apitype.CategorizedImage {
	cats := make([]*apitype.CategorizedImage, len(categories))
	for i, category := range categories {
		cats[i] = toApiCategorizedImage(&category)
	}
	return cats
}

func toApiCategorizedImage(category *CategorizedImage) *apitype.CategorizedImage {
	return apitype.NewCategorizedImage(
		apitype.NewCategoryWithId(
			category.CategoryId, category.Name, category.SubPath, category.Shortcut),
		apitype.OperationFromId(category.Operation),
	)
}

func toApiCategories(categories []Category) []*apitype.Category {
	cats := make([]*apitype.Category, len(categories))
	for i, category := range categories {
		cats[i] = toApiCategory(category)
	}
	return cats
}

func toApiCategory(category Category) *apitype.Category {
	return apitype.NewCategoryWithId(category.Id, category.Name, category.SubPath, category.Shortcut)
}
