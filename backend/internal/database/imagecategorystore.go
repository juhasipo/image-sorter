package database

import (
	"github.com/upper/db/v4"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
)

type ImageCategoryStore struct {
	database   *Database
	collection db.Collection
}

func NewImageCategoryStore(database *Database) *ImageCategoryStore {
	return &ImageCategoryStore{
		database: database,
	}
}

func (s *ImageCategoryStore) getCollection() db.Collection {
	if s.collection == nil {
		s.collection = s.database.Session().Collection("image_category")
	}
	return s.collection
}

func (s *ImageCategoryStore) RemoveImageCategories(imageId apitype.ImageId) error {
	_, err := s.getCollection().Session().SQL().Exec(`
			DELETE FROM image_category WHERE image_id = ?
		`, imageId)
	return err
}

func (s *ImageCategoryStore) CategorizeImage(imageId apitype.ImageId, categoryId apitype.CategoryId, operation apitype.Operation) error {
	if operation == apitype.UNCATEGORIZE {
		_, err := s.getCollection().Session().SQL().Exec(`
			DELETE FROM image_category WHERE image_id = ? AND category_id = ?
		`, imageId, categoryId)
		return err
	} else {
		_, err := s.getCollection().Session().SQL().Exec(`
		INSERT INTO image_category (image_id, category_id, operation)
		VALUES(?, ?, ?)
		ON CONFLICT(image_id, category_id) DO 
		UPDATE SET operation = ?
		WHERE image_id = ? AND category_id = ?
	`, imageId, categoryId, operation, operation, imageId, categoryId)
		return err
	}
}

func (s *ImageCategoryStore) GetImagesCategories(imageId apitype.ImageId) ([]*api.CategorizedImage, error) {
	var categories []CategorizedImage
	err := s.getCollection().Session().SQL().
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

func (s *ImageCategoryStore) GetCategorizedImages() (map[apitype.ImageId]map[apitype.CategoryId]*api.CategorizedImage, error) {
	var categorizedImages []CategorizedImage
	err := s.getCollection().Session().SQL().
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

	var categoryImagesByImageIdAndCategoryId = map[apitype.ImageId]map[apitype.CategoryId]*api.CategorizedImage{}
	for _, categorizedImage := range categorizedImages {
		var categorizedImagesByCategoryId map[apitype.CategoryId]*api.CategorizedImage
		if val, ok := categoryImagesByImageIdAndCategoryId[categorizedImage.ImageId]; ok {
			categorizedImagesByCategoryId = val
		} else {
			categorizedImagesByCategoryId = map[apitype.CategoryId]*api.CategorizedImage{}
			categoryImagesByImageIdAndCategoryId[categorizedImage.ImageId] = categorizedImagesByCategoryId
		}
		categorizedImagesByCategoryId[categorizedImage.CategoryId] = toApiCategorizedImage(&categorizedImage)
	}
	return categoryImagesByImageIdAndCategoryId, nil
}
