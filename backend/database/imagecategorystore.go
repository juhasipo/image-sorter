package database

import (
	"github.com/upper/db/v4"
	"vincit.fi/image-sorter/api/apitype"
)

type ImageCategoryStore struct {
	collection db.Collection
}

func NewImageCategoryStore(store *Store) *ImageCategoryStore {
	return &ImageCategoryStore{
		collection: store.imageCategoryCollection,
	}
}

func (s *ImageCategoryStore) RemoveImageCategories(imageId apitype.HandleId) error {
	_, err := s.collection.Session().SQL().Exec(`
			DELETE FROM image_category WHERE image_id = ?
		`, imageId)
	return err
}

func (s *ImageCategoryStore) CategorizeImage(imageId apitype.HandleId, categoryId apitype.CategoryId, operation apitype.Operation) error {
	if operation == apitype.NONE {
		_, err := s.collection.Session().SQL().Exec(`
			DELETE FROM image_category WHERE image_id = ? AND category_id = ?
		`, imageId, categoryId)
		return err
	} else {
		_, err := s.collection.Session().SQL().Exec(`
		INSERT INTO image_category (image_id, category_id, operation)
		VALUES(?, ?, ?)
		ON CONFLICT(image_id, category_id) DO 
		UPDATE SET operation = ?
		WHERE image_id = ? AND category_id = ?
	`, imageId, categoryId, operation, operation, imageId, categoryId)
		return err
	}
}

func (s *ImageCategoryStore) GetImagesCategories(imageId apitype.HandleId) ([]*apitype.CategorizedImage, error) {
	var categories []CategorizedImage
	err := s.collection.Session().SQL().
		Select("image_category.image_id AS image_id",
			"category.id AS category_id",
			"category.name AS name",
			"category.sub_path AS sub_path",
			"category.shortcut AS shortcut",
			"image_category.operation AS operation").
		From("category").
		Join("image_category").On("image_category.category_id = category.id").
		Where("image_category.image_id", imageId).
		OrderBy("category.name").
		All(&categories)

	if err != nil {
		return nil, err
	}

	return toApiCategorizedImages(categories), nil
}

func (s *ImageCategoryStore) GetCategorizedImages() (map[apitype.HandleId]map[apitype.CategoryId]*apitype.CategorizedImage, error) {
	var categorizedImages []CategorizedImage
	err := s.collection.Session().SQL().
		Select("image_category.image_id AS image_id",
			"category.id AS category_id",
			"category.name AS name",
			"category.sub_path AS sub_path",
			"category.shortcut AS shortcut",
			"image_category.operation AS operation").
		From("category").
		Join("image_category").On("image_category.category_id = category.id").
		OrderBy("category.name").
		All(&categorizedImages)

	if err != nil {
		return nil, err
	}

	var catImagesByHandleIdAndCategoryId = map[apitype.HandleId]map[apitype.CategoryId]*apitype.CategorizedImage{}
	for _, categorizedImage := range categorizedImages {
		var categorizedImagesByCategoryId map[apitype.CategoryId]*apitype.CategorizedImage
		if val, ok := catImagesByHandleIdAndCategoryId[categorizedImage.ImageId]; ok {
			categorizedImagesByCategoryId = val
		} else {
			categorizedImagesByCategoryId = map[apitype.CategoryId]*apitype.CategorizedImage{}
			catImagesByHandleIdAndCategoryId[categorizedImage.ImageId] = categorizedImagesByCategoryId
		}
		categorizedImagesByCategoryId[categorizedImage.CategoryId] = toApiCategorizedImage(&categorizedImage)
	}
	return catImagesByHandleIdAndCategoryId, nil
}
