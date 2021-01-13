package database

import (
	"github.com/upper/db/v4"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
)

type CategoryStore struct {
	database   *Database
	collection db.Collection
}

func NewCategoryStore(database *Database) *CategoryStore {
	return &CategoryStore{
		database: database,
	}
}

func (s *CategoryStore) getCollection() db.Collection {
	if s.collection == nil {
		s.collection = s.database.Session().Collection("category")
	}
	return s.collection
}

func (s *CategoryStore) AddCategory(category *apitype.Category) (*apitype.Category, error) {
	return addCategory(s.getCollection(), category)
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

func (s *CategoryStore) ResetCategories(categories []*apitype.Category) error {
	return s.getCollection().Session().Tx(func(sess db.Session) error {
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

func (s *CategoryStore) GetCategories() ([]*apitype.Category, error) {
	var categories []Category
	err := s.getCollection().Find().
		OrderBy("name").
		All(&categories)

	if err != nil {
		return nil, err
	}

	return toApiCategories(categories), nil
}

func (s *CategoryStore) GetCategoryById(id apitype.CategoryId) *apitype.Category {
	var category Category
	if err := s.getCollection().Find(db.Cond{"id": id}).One(&category); err != nil {
		return nil
	} else {
		return toApiCategory(category)
	}
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
