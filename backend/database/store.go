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
}

func NewStore(database *Database) *Store {
	return &Store{
		database:                database,
		imageCollection:         database.Session().Collection("image"),
		categoryCollection:      database.Session().Collection("category"),
		imageCategoryCollection: database.Session().Collection("image_category"),
	}
}

func (s *Store) AddImages(handles []*apitype.Handle) ([]*apitype.Handle, error) {
	var persistedHandles []*apitype.Handle
	for _, handle := range handles {
		persistedHandle, err := s.AddImage(handle)
		if err != nil {
			logger.Error.Printf("Error while adding image '%s' to DB", handle.GetPath())
			return nil, err
		}
		persistedHandles = append(persistedHandles, persistedHandle)
	}

	return persistedHandles, nil
}

func (s *Store) AddImage(handle *apitype.Handle) (*apitype.Handle, error) {
	if existing, err := s.FindByDirAndFile(handle.GetDir(), handle.GetFile()); err != nil {
		return nil, err
	} else if existing != nil {
		logger.Trace.Printf("Image %s/%s already in DB", handle.GetDir(), handle.GetFile())
		return existing, nil
	}

	result, err := s.imageCollection.Insert(Image{
		Name:      handle.GetFile(),
		FileName:  handle.GetFile(),
		Directory: handle.GetDir(),
		ByteSize:  handle.GetByteSize(),
	})
	if err != nil {
		return nil, err
	}

	var image Image
	err = s.imageCollection.Find("id", result.ID()).One(&image)
	if err != nil {
		return nil, err
	}

	return apitype.NewHandle(image.Id, image.Directory, image.FileName), err
}

func (s *Store) GetImages(number int, offset int) ([]*apitype.Handle, error) {
	return s.GetImagesInCategories(number, offset)
}

func (s *Store) GetImagesInCategories(number int, offset int, categories ...int64) ([]*apitype.Handle, error) {
	if number == 0 {
		return make([]*apitype.Handle, 0), nil
	}

	var images []Image
	res := s.database.Session().SQL().
		SelectFrom("image")

	if len(categories) > 0 {
		res = res.
			Join("image_category").On("image_category.image_id = image.id").
			Where("image_category.category_id IN ", categories)
	}
	res = res.Limit(number).
		Offset(offset).
		OrderBy("name")

	if err := res.All(&images); err != nil {
		return nil, err
	} else {
		handles := make([]*apitype.Handle, len(images))
		for i, image := range images {
			handles[i] = imageToHandle(&image)
		}
		return handles, nil
	}
}

func (s *Store) RemoveImageCategories(imageId int64) error {
	_, err := s.database.Session().SQL().Exec(`
			DELETE FROM image_category WHERE image_id = ?
		`, imageId)
	return err
}

func (s *Store) CategorizeImage(imageId int64, categoryId int64, operation apitype.Operation) error {
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
	var existing []Category
	if err := s.categoryCollection.Find(db.Cond{"name": category.GetName()}).
		All(&existing); err != nil {
		return nil, err
	} else if len(existing) > 0 {
		return toApiCategory(existing[0]), nil
	}

	result, err := s.categoryCollection.Insert(Category{
		Name:     category.GetName(),
		SubPath:  category.GetSubPath(),
		Shortcut: category.GetShortcutAsString(),
	})
	if err != nil {
		return nil, err
	}

	var cat Category
	err = s.categoryCollection.Find("id", result.ID()).One(&cat)
	if err != nil {
		return nil, err
	}

	logger.Debug.Printf("Stored category %s (%d) to DB", cat.Name, cat.Id)
	return apitype.NewCategory(cat.Id, cat.Name, cat.SubPath, cat.Shortcut), err
}

func (s *Store) ResetCategories(categories []*apitype.Category) error {
	return s.database.Session().Tx(func(sess db.Session) error {
		for _, category := range categories {
			if _, err := s.AddCategory(category); err != nil {
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

func toApiCategories(categories []Category) []*apitype.Category {
	cats := make([]*apitype.Category, len(categories))
	for i, category := range categories {
		cats[i] = toApiCategory(category)
	}
	return cats
}

func toApiCategory(category Category) *apitype.Category {
	return apitype.NewCategory(category.Id, category.Name, category.SubPath, category.Shortcut)
}

func (s *Store) GetImagesCategories(imageId int64) ([]*apitype.CategorizedImage, error) {
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

func toApiCategorizedImages(categories []CategorizedImage) []*apitype.CategorizedImage {
	cats := make([]*apitype.CategorizedImage, len(categories))
	for i, category := range categories {
		cats[i] = toApiCategorizedImage(&category)
	}
	return cats
}

func toApiCategorizedImage(category *CategorizedImage) *apitype.CategorizedImage {
	return apitype.NewCategorizedImage(
		apitype.NewCategory(
			category.CategoryId, category.Name, category.Name, category.Shortcut),
		apitype.OperationFromId(category.Operation),
	)
}

func (s *Store) FindByDirAndFile(directory string, fileName string) (*apitype.Handle, error) {
	var handles []Image
	err := s.imageCollection.
		Find(db.Cond{
			"directory": directory,
			"file_name": fileName,
		}).
		All(&handles)
	if err != nil {
		return nil, err
	} else if len(handles) == 1 {
		return imageToHandle(&handles[0]), nil
	} else {
		return nil, nil
	}
}

func (s *Store) GetCategoryById(id int64) *apitype.Category {
	var category Category
	if err := s.categoryCollection.Find(db.Cond{"id": id}).One(&category); err != nil {
		return toApiCategory(category)
	} else {
		return nil
	}
}

func (s *Store) GetCategorizedImages() (map[int64]map[int64]*apitype.CategorizedImage, error) {
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
		OrderBy("category.name").
		All(&categories)

	if err != nil {
		return nil, err
	}

	var a = map[int64]map[int64]*apitype.CategorizedImage{}
	for _, image := range categories {
		var m map[int64]*apitype.CategorizedImage
		if val, ok := a[image.ImageId]; ok {
			m = val
		} else {
			m = map[int64]*apitype.CategorizedImage{}
			a[image.ImageId] = m
		}
		m[image.CategoryId] = toApiCategorizedImage(&image)
	}
	return a, nil
}

func imageToHandle(image *Image) *apitype.Handle {
	return apitype.NewHandle(
		image.Id, image.Directory, image.FileName,
	)
}
