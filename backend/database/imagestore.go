package database

import (
	"github.com/upper/db/v4"
	"time"
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
	return s.collection.Session().Tx(func(sess db.Session) error {
		for _, handle := range handles {
			if _, err := s.addImage(sess, handle); err != nil {
				logger.Error.Printf("Error while adding image '%s' to DB", handle.GetPath())
				return err
			}
		}
		return nil
	})
}

func (s *ImageStore) AddImage(handle *apitype.Handle) (*apitype.Handle, error) {
	return s.addImage(s.collection.Session(), handle)
}

func (s *ImageStore) addImage(session db.Session, handle *apitype.Handle) (*apitype.Handle, error) {
	collection := s.getCollectionForSession(session)

	logger.Trace.Printf("Adding image '%s'", handle.String())

	existStart := time.Now()
	exists, err := s.exists(collection, handle)
	if err != nil {
		return nil, err
	}
	existsEnd := time.Now()
	logger.Trace.Printf(" - Checked if image exists %s", existsEnd.Sub(existStart))

	if !exists {
		handleToImageStart := time.Now()
		image, err := s.imageHandleConverter.HandleToImage(handle)
		if err != nil {
			return nil, err
		}

		handleToImageEnd := time.Now()
		logger.Trace.Printf(" - Loaded image meta data in %s", handleToImageEnd.Sub(handleToImageStart))

		insertStart := time.Now()
		if _, err := collection.Insert(image); err != nil {
			return nil, err
		}
		insertEnd := time.Now()
		logger.Trace.Printf(" - Added image to DB in %s", insertEnd.Sub(insertStart))

		return s.findByDirAndFile(collection, handle)
	}

	modifiedId, err := s.findModifiedId(collection, handle)
	if err != nil {
		return nil, err
	}

	if modifiedId > 0 {
		logger.Trace.Printf(" - Image exists with ID %d but is modified", modifiedId)

		handleToImageStart := time.Now()
		image, err := s.imageHandleConverter.HandleToImage(handle)
		if err != nil {
			return nil, err
		}

		handleToImageEnd := time.Now()
		logger.Trace.Printf(" - Loaded image meta data in %s", handleToImageEnd.Sub(handleToImageStart))

		updateStart := time.Now()
		err = s.update(collection, modifiedId, image)
		if err != nil {
			return nil, err
		}
		updateEnd := time.Now()
		logger.Trace.Printf(" - Image meta data updated %s", updateEnd.Sub(updateStart))

		return apitype.NewPersistedHandle(modifiedId, handle), nil
	} else {
		return s.findByDirAndFile(collection, handle)
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
		Select("image.*").
		From("image")

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
	return s.findByDirAndFile(s.collection, handle)
}

func (s *ImageStore) findByDirAndFile(collection db.Collection, handle *apitype.Handle) (*apitype.Handle, error) {
	var handles []Image
	err := collection.
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
	return s.exists(s.collection, handle)
}

func (s *ImageStore) exists(collection db.Collection, handle *apitype.Handle) (bool, error) {
	return collection.
		Find(db.Cond{
			"directory": handle.GetDir(),
			"file_name": handle.GetFile(),
		}).
		Exists()
}

func (s *ImageStore) getCollectionForSession(session db.Session) db.Collection {
	return session.Collection(s.collection.Name())
}

func (s *ImageStore) FindModifiedId(handle *apitype.Handle) (apitype.HandleId, error) {
	return s.findModifiedId(s.collection, handle)
}

func (s *ImageStore) findModifiedId(collection db.Collection, handle *apitype.Handle) (apitype.HandleId, error) {
	stat, err := getHandleFileStats(handle)
	if err != nil {
		return -1, err
	}

	var images []Image
	err = collection.
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

func (s *ImageStore) update(collection db.Collection, id apitype.HandleId, image *Image) error {
	return collection.Find(db.Cond{"id": id}).Update(image)
}

func (s *ImageStore) GetImageById(id apitype.HandleId) *apitype.Handle {
	var image Image
	err := s.collection.Find(db.Cond{"id": id}).One(&image)

	if err != nil {
		logger.Error.Print("Could not find image ", err)
	}

	return toApiHandle(&image)

}

func (s *ImageStore) RemoveImage(handleId apitype.HandleId) error {
	return s.collection.Find(db.Cond{"id": handleId}).Delete()
}
