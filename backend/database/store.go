package database

import (
	"github.com/upper/db/v4"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
)

type Store struct {
	database                *Database
	imageCollection         db.Collection
	categoryCollection      db.Collection
	imageCategoryCollection db.Collection
	imageSimilarCollection  db.Collection
	imageHandleConverter    ImageHandleConverter
}

func NewStore(database *Database) *Store {
	return &Store{
		database:                database,
		imageCollection:         database.Session().Collection("image"),
		categoryCollection:      database.Session().Collection("category"),
		imageCategoryCollection: database.Session().Collection("image_category"),
		imageSimilarCollection:  database.Session().Collection("image_similar"),
		imageHandleConverter:    &FileSystemImageHandleConverter{},
	}
}

func NewInMemoryStore() *Store {
	memoryDb := NewInMemoryDatabase()
	memoryDb.Migrate()
	memoryStore := NewStore(memoryDb)
	return memoryStore
}

func (s *Store) RemoveImageCategories(imageId apitype.HandleId) error {
	_, err := s.database.Session().SQL().Exec(`
			DELETE FROM image_category WHERE image_id = ?
		`, imageId)
	return err
}

func (s *Store) CategorizeImage(imageId apitype.HandleId, categoryId apitype.CategoryId, operation apitype.Operation) error {
	if operation == apitype.NONE {
		_, err := s.database.Session().SQL().Exec(`
			DELETE FROM image_category WHERE image_id = ? AND category_id = ?
		`, imageId, categoryId)
		return err
	} else {
		_, err := s.database.Session().SQL().Exec(`
		INSERT INTO image_category (image_id, category_id, operation)
		VALUES(?, ?, ?)
		ON CONFLICT(image_id, category_id) DO 
		UPDATE SET operation = ?
		WHERE image_id = ? AND category_id = ?
	`, imageId, categoryId, operation, operation, imageId, categoryId)
		return err
	}
}

func (s *Store) AddCategory(category *apitype.Category) (*apitype.Category, error) {
	return addCategory(s.categoryCollection, category)
}

func addCategory(collection db.Collection, category *apitype.Category) (*apitype.Category, error) {
	var existing []Category
	if err := collection.Find(db.Cond{"name": category.GetName()}).
		All(&existing); err != nil {
		return nil, err
	} else if len(existing) > 0 {
		return toApiCategory(existing[0]), nil
	}

	result, err := collection.Insert(Category{
		Name:     category.GetName(),
		SubPath:  category.GetSubPath(),
		Shortcut: category.GetShortcutAsString(),
	})

	if err != nil {
		return nil, err
	}

	logger.Debug.Printf("Stored category %s (%d) to DB", category.GetName(), category.GetId())
	return apitype.NewPersistedCategory(idToCategoryId(result.ID()), category), err
}

func (s *Store) ResetCategories(categories []*apitype.Category) error {
	return s.database.Session().Tx(func(sess db.Session) error {
		collection := sess.Collection("category")
		var persistedCategories []Category
		if err := collection.Find().All(&persistedCategories); err != nil {
			return err
		}

		var existingCategoriesById = map[apitype.CategoryId]*apitype.Category{}
		for _, category := range persistedCategories {
			existingCategoriesById[category.Id] = toApiCategory(category)
		}

		for _, category := range categories {
			categoryKey := category.GetId()
			if _, ok := existingCategoriesById[categoryKey]; ok {
				if err := updateCategory(collection, category); err != nil {
					return err
				}
				delete(existingCategoriesById, categoryKey)
			} else {
				if _, err := addCategory(collection, category); err != nil {
					return err
				}
			}
		}

		for _, category := range existingCategoriesById {
			if err := removeCategory(collection, category.GetId()); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) GetCategories() ([]*apitype.Category, error) {
	var categories []Category
	err := s.categoryCollection.Find().
		OrderBy("name").
		All(&categories)

	if err != nil {
		return nil, err
	}

	return toApiCategories(categories), nil
}

func (s *Store) GetImagesCategories(imageId apitype.HandleId) ([]*apitype.CategorizedImage, error) {
	var categories []CategorizedImage
	err := s.database.Session().SQL().
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

func (s *Store) GetCategoryById(id apitype.CategoryId) *apitype.Category {
	var category Category
	if err := s.categoryCollection.Find(db.Cond{"id": id}).One(&category); err != nil {
		return toApiCategory(category)
	} else {
		return nil
	}
}

func (s *Store) GetCategorizedImages() (map[apitype.HandleId]map[apitype.CategoryId]*apitype.CategorizedImage, error) {
	var categorizedImages []CategorizedImage
	err := s.database.Session().SQL().
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

func (s *Store) DoInTransaction(fn func(session db.Session) error) error {
	return s.database.Session().Tx(fn)
}

func removeCategory(collection db.Collection, categoryId apitype.CategoryId) error {
	return collection.Find(db.Cond{"id": categoryId}).Delete()
}

func updateCategory(collection db.Collection, category *apitype.Category) error {
	return collection.Find(db.Cond{"id": category.GetId()}).Update(&Category{
		Id:       category.GetId(),
		Name:     category.GetName(),
		SubPath:  category.GetSubPath(),
		Shortcut: category.GetShortcutAsString(),
	})
}
