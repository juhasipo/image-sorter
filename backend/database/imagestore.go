package database

import (
	"fmt"
	"github.com/upper/db/v4"
	"time"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
	"vincit.fi/image-sorter/common/util"
)

type ImageStore struct {
	database             *Database
	collection           db.Collection
	imageHandleConverter ImageHandleConverter
}

func NewImageStore(database *Database, imageHandleConverter ImageHandleConverter) *ImageStore {
	return &ImageStore{
		database:             database,
		imageHandleConverter: imageHandleConverter,
	}
}

func (s *ImageStore) SetImageHandleConverter(imageHandleConverter ImageHandleConverter) {
	s.imageHandleConverter = imageHandleConverter
}

func (s *ImageStore) getCollection() db.Collection {
	if s.collection == nil {
		s.collection = s.database.Session().Collection("image")
	}
	return s.collection
}

func (s *ImageStore) AddImages(imageFiles []*apitype.ImageFile) error {
	return s.getCollection().Session().Tx(func(sess db.Session) error {
		for _, imageFile := range imageFiles {
			if _, err := s.addImage(sess, imageFile); err != nil {
				logger.Error.Printf("Error while adding image '%s' to DB", imageFile.GetPath())
				return err
			}
		}
		return nil
	})
}

func (s *ImageStore) AddImage(handle *apitype.ImageFile) (*apitype.ImageFileWithMetaData, error) {
	return s.addImage(s.getCollection().Session(), handle)
}

func (s *ImageStore) addImage(session db.Session, handle *apitype.ImageFile) (*apitype.ImageFileWithMetaData, error) {
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
		image, _, err := s.imageHandleConverter.HandleToImage(handle)
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

	if modifiedId > apitype.ImageId(0) {
		logger.Trace.Printf(" - Image exists with ID %d but is modified", modifiedId)

		handleToImageStart := time.Now()
		image, _, err := s.imageHandleConverter.HandleToImage(handle)
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

		return s.findByDirAndFile(collection, handle)
	} else {
		return s.findByDirAndFile(collection, handle)
	}
}

func (s *ImageStore) GetImageCount(categoryId apitype.CategoryId) int {
	res := s.getCollection().Session().SQL().
		Select(db.Raw("count(1) AS c")).
		From("image")

	if categoryId != apitype.NoCategory {
		res = res.
			Join("image_category").On("image_category.image_id = image.id").
			Join("category").On("image_category.category_id = category.id").
			Where("category.id", categoryId)
	}

	var counter Count
	if err := res.One(&counter); err != nil {
		logger.Error.Fatal("Cannot resolve image count", err)
	}

	return counter.Count
}

func (s *ImageStore) GetAllImages() ([]*apitype.ImageFileWithMetaData, error) {
	return s.GetImagesInCategory(-1, 0, apitype.NoCategory)
}

func (s *ImageStore) GetNextImagesInCategory(number int, currentIndex int, categoryId apitype.CategoryId) ([]*apitype.ImageFileWithMetaData, error) {
	startIndex := currentIndex + 1
	return s.GetImagesInCategory(number, startIndex, categoryId)
}

func (s *ImageStore) GetPreviousImagesInCategory(number int, currentIndex int, categoryId apitype.CategoryId) ([]*apitype.ImageFileWithMetaData, error) {
	prevIndex := currentIndex - number
	if prevIndex < 0 {
		prevIndex = 0
	}
	size := currentIndex - prevIndex

	if size < 0 {
		return []*apitype.ImageFileWithMetaData{}, nil
	} else if images, err := s.GetImagesInCategory(size, prevIndex, categoryId); err != nil {
		return nil, err
	} else {
		util.Reverse(images)
		return images, nil
	}
}

func (s *ImageStore) GetImagesInCategory(number int, offset int, categoryId apitype.CategoryId) ([]*apitype.ImageFileWithMetaData, error) {
	if number == 0 {
		return make([]*apitype.ImageFileWithMetaData, 0), nil
	}

	var images []Image
	res := s.getCollection().Session().SQL().
		Select("image.*").
		From("image")

	if categoryId != apitype.NoCategory {
		res = res.
			Join("image_category").On("image_category.image_id = image.id").
			Join("category").On("image_category.category_id = category.id").
			Where("category.id", categoryId)
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

type sortDir string

const (
	asc  sortDir = "ASC"
	desc sortDir = "DESC"
)

func (s *ImageStore) getImagesInCategory(number int, offset int, categoryName string, sort sortDir) ([]*apitype.ImageFileWithMetaData, error) {
	if number == 0 {
		return make([]*apitype.ImageFileWithMetaData, 0), nil
	}

	var images []Image
	res := s.getCollection().Session().SQL().
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

	res = res.OrderBy(fmt.Sprintf("image.name %s", sort))

	if err := res.All(&images); err != nil {
		return nil, err
	} else {
		handles := toApiHandles(images)
		return handles, nil
	}
}

func (s *ImageStore) FindByDirAndFile(handle *apitype.ImageFile) (*apitype.ImageFileWithMetaData, error) {
	return s.findByDirAndFile(s.getCollection(), handle)
}

func (s *ImageStore) findByDirAndFile(collection db.Collection, handle *apitype.ImageFile) (*apitype.ImageFileWithMetaData, error) {
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
		return toApiHandle(&handles[0])
	} else {
		return nil, nil
	}
}

func (s *ImageStore) Exists(handle *apitype.ImageFile) (bool, error) {
	return s.exists(s.getCollection(), handle)
}

func (s *ImageStore) exists(collection db.Collection, handle *apitype.ImageFile) (bool, error) {
	return collection.
		Find(db.Cond{
			"directory": handle.GetDir(),
			"file_name": handle.GetFile(),
		}).
		Exists()
}

func (s *ImageStore) getCollectionForSession(session db.Session) db.Collection {
	return session.Collection(s.getCollection().Name())
}

func (s *ImageStore) FindModifiedId(handle *apitype.ImageFile) (apitype.ImageId, error) {
	return s.findModifiedId(s.getCollection(), handle)
}

func (s *ImageStore) findModifiedId(collection db.Collection, imageFile *apitype.ImageFile) (apitype.ImageId, error) {
	stat, err := s.imageHandleConverter.GetHandleFileStats(imageFile)
	if err != nil {
		return apitype.NoImage, err
	}

	var images []Image
	err = collection.
		Find(db.Cond{
			"directory":            imageFile.GetDir(),
			"file_name":            imageFile.GetFile(),
			"modified_timestamp <": stat.ModTime(),
		}).All(&images)

	if err != nil {
		return apitype.NoImage, err
	}

	if len(images) > 0 {
		return images[0].Id, nil
	} else {
		return apitype.NoImage, nil
	}
}

func (s *ImageStore) update(collection db.Collection, id apitype.ImageId, image *Image) error {
	return collection.Find(db.Cond{"id": id}).Update(image)
}

func (s *ImageStore) GetImageById(id apitype.ImageId) *apitype.ImageFileWithMetaData {
	var image Image
	err := s.getCollection().Find(db.Cond{"id": id}).One(&image)

	if err != nil {
		logger.Error.Print("Could not find image ", err)
	}

	handle, err := toApiHandle(&image)
	if err != nil {
		logger.Error.Print("Could not find image ", err)
	}

	return handle

}

func (s *ImageStore) RemoveImage(imageId apitype.ImageId) error {
	return s.getCollection().Find(db.Cond{"id": imageId}).Delete()
}
