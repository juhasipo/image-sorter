package database

import (
	"github.com/upper/db/v4"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
)

type ImageStore struct {
	collection           db.Collection
	imageHandleConverter ImageHandleConverter
}

func NewImageStore(store *Store, imageHandleConverter ImageHandleConverter) *ImageStore {
	return &ImageStore{
		collection:           store.imageCollection,
		imageHandleConverter: imageHandleConverter,
	}
}

func (s *ImageStore) SetImageHandleConverter(imageHandleConverter ImageHandleConverter) {
	s.imageHandleConverter = imageHandleConverter
}

func (s *ImageStore) AddImages(handles []*apitype.Handle) error {
	for _, handle := range handles {
		if _, err := s.AddImage(handle); err != nil {
			logger.Error.Printf("Error while adding image '%s' to DB", handle.GetPath())
			return err
		}
	}

	return nil
}

func (s *ImageStore) AddImage(handle *apitype.Handle) (*apitype.Handle, error) {
	if exists, err := s.Exists(handle); err != nil {
		return nil, err
	} else if !exists {
		logger.Debug.Printf("Adding image '%s' to DB", handle.String())
		if image, err := s.imageHandleConverter.HandleToImage(handle); err != nil {
			return nil, err
		} else if _, err := s.collection.Insert(image); err != nil {
			return nil, err
		} else {
			return s.FindByDirAndFile(handle)
		}
	} else if modifiedId, err := s.FindModifiedId(handle); err != nil {
		return nil, err
	} else if modifiedId > 0 {
		logger.Debug.Printf("Updating existing image '%s' in DB", handle.String())
		if image, err := s.imageHandleConverter.HandleToImage(handle); err != nil {
			return nil, err
		} else if err := s.Update(modifiedId, image); err != nil {
			return nil, err
		} else {
			return apitype.NewPersistedHandle(modifiedId, handle), nil
		}
	} else {
		return s.FindByDirAndFile(handle)
	}
}

func (s *ImageStore) GetImageCount(categoryName string) int {
	res := s.collection.Session().SQL().
		Select(db.Raw("count(1) AS c")).
		From("image")

	if categoryName != "" {
		res = res.
			Join("image_category").On("image_category.image_id = image.id").
			Join("category").On("image_category.category_id = category.id").
			Where("category.name", categoryName)
	}

	var counter Count
	if err := res.One(&counter); err != nil {
		logger.Error.Fatal("Cannot resolve image count", err)
	}

	return counter.Count
}

func (s *ImageStore) GetImages(number int, offset int) ([]*apitype.Handle, error) {
	return s.GetImagesInCategory(number, offset, "")
}

func (s *ImageStore) GetImagesInCategory(number int, offset int, categoryName string) ([]*apitype.Handle, error) {
	if number == 0 {
		return make([]*apitype.Handle, 0), nil
	}

	var images []Image
	res := s.collection.Session().SQL().
		SelectFrom("image")

	if categoryName != "" {
		res = res.
			Join("image_category").On("image_category.image_id = image.id").
			Join("category").On("image_category.category_id = category.id").
			Where("category.name", categoryName)
	}
	if number >= 0 {
		res = res.Limit(number).
			Offset(offset)
	}

	res = res.OrderBy("image.name")

	if err := res.All(&images); err != nil {
		return nil, err
	} else {
		handles := toApiHandles(images)
		return handles, nil
	}
}

func (s *ImageStore) FindByDirAndFile(handle *apitype.Handle) (*apitype.Handle, error) {
	var handles []Image
	err := s.collection.
		Find(db.Cond{
			"directory": handle.GetDir(),
			"file_name": handle.GetFile(),
		}).
		All(&handles)
	if err != nil {
		return nil, err
	} else if len(handles) == 1 {
		return toApiHandle(&handles[0]), nil
	} else {
		return nil, nil
	}
}

func (s *ImageStore) Exists(handle *apitype.Handle) (bool, error) {
	return s.collection.
		Find(db.Cond{
			"directory": handle.GetDir(),
			"file_name": handle.GetFile(),
		}).
		Exists()
}

func (s *ImageStore) FindModifiedId(handle *apitype.Handle) (apitype.HandleId, error) {
	stat, err := getHandleFileStats(handle)
	if err != nil {
		return -1, err
	}

	var images []Image
	err = s.collection.
		Find(db.Cond{
			"directory":            handle.GetDir(),
			"file_name":            handle.GetFile(),
			"modified_timestamp <": stat.ModTime(),
		}).All(&images)

	if err != nil {
		return -1, err
	}

	if len(images) > 0 {
		return images[0].Id, nil
	} else {
		return -1, nil
	}
}

func (s *ImageStore) Update(id apitype.HandleId, image *Image) error {
	return s.collection.Find(db.Cond{"id": id}).Update(image)
}

func (s *ImageStore) GetImageById(id apitype.HandleId) *apitype.Handle {
	var image Image
	err := s.collection.Find(db.Cond{"id": id}).One(&image)

	if err != nil {
		logger.Error.Print("Could not find image ", err)
	}

	return toApiHandle(&image)

}
