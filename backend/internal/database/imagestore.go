package database

import (
	"github.com/upper/db/v4"
	"sync"
	"time"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
	"vincit.fi/image-sorter/common/util"
)

type ImageStore struct {
	database           *Database
	collection         db.Collection
	imageFileConverter ImageFileConverter
	mux                sync.Mutex
}

func NewImageStore(database *Database, imageFileConverter ImageFileConverter) *ImageStore {
	return &ImageStore{
		database:           database,
		imageFileConverter: imageFileConverter,
	}
}

func (s *ImageStore) SetImageFileConverter(imageFileConverter ImageFileConverter) {
	s.imageFileConverter = imageFileConverter
}

func (s *ImageStore) getCollection() db.Collection {
	if s.collection == nil {
		s.collection = s.database.Session().Collection("image")
	}
	return s.collection
}

func (s *ImageStore) AddImages(imageFiles []*apitype.ImageFile) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.getCollection().Session().Tx(func(sess db.Session) error {
		for _, imageFile := range imageFiles {
			if _, err := s.addImage(sess, imageFile); err != nil {
				logger.Error.Printf("Error while adding image '%s' to DB", imageFile.Path())
				return err
			}
		}
		return nil
	})
}

func (s *ImageStore) AddImage(imageFile *apitype.ImageFile) (*apitype.ImageFile, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.addImage(s.getCollection().Session(), imageFile)
}

func (s *ImageStore) addImage(session db.Session, imageFile *apitype.ImageFile) (*apitype.ImageFile, error) {
	collection := s.getCollectionForSession(session)

	logger.Trace.Printf("Adding image '%s'", imageFile.String())

	existStart := time.Now()
	exists, err := s.exists(collection, imageFile)
	if err != nil {
		return nil, err
	}
	existsEnd := time.Now()
	logger.Trace.Printf(" - Checked if image exists %s", existsEnd.Sub(existStart))

	if !exists {
		imageFileToDbImageStart := time.Now()
		image, _, err := s.imageFileConverter.ImageFileToDbImage(imageFile)
		if err != nil {
			return nil, err
		}

		imageFileToImageDbEnd := time.Now()
		logger.Trace.Printf(" - Loaded image meta data in %s", imageFileToImageDbEnd.Sub(imageFileToDbImageStart))

		insertStart := time.Now()
		if _, err := collection.Insert(image); err != nil {
			return nil, err
		}
		insertEnd := time.Now()
		logger.Trace.Printf(" - Added image to DB in %s", insertEnd.Sub(insertStart))

		return s.findByFileName(collection, imageFile)
	}

	modifiedId, err := s.findModifiedId(collection, imageFile)
	if err != nil {
		return nil, err
	}

	if modifiedId > apitype.ImageId(0) {
		logger.Trace.Printf(" - Image exists with ID %d but is modified", modifiedId)

		imageFileToDbImageStart := time.Now()
		image, _, err := s.imageFileConverter.ImageFileToDbImage(imageFile)
		if err != nil {
			return nil, err
		}

		imageFileToDbImageEnd := time.Now()
		logger.Trace.Printf(" - Loaded image meta data in %s", imageFileToDbImageEnd.Sub(imageFileToDbImageStart))

		updateStart := time.Now()
		err = s.update(collection, modifiedId, image)
		if err != nil {
			return nil, err
		}
		updateEnd := time.Now()
		logger.Trace.Printf(" - Image meta data updated %s", updateEnd.Sub(updateStart))

		return s.findByFileName(collection, imageFile)
	} else {
		return s.findByFileName(collection, imageFile)
	}
}

func (s *ImageStore) GetImageCount(categoryId apitype.CategoryId) int {
	s.mux.Lock()
	defer s.mux.Unlock()
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

func (s *ImageStore) GetLatestModifiedImage() time.Time {
	s.mux.Lock()
	defer s.mux.Unlock()
	res := s.getCollection().Session().SQL().
		Select(db.Raw("modified_timestamp AS t")).
		From("image").
		OrderBy("modified_timestamp DESC").
		Limit(1)

	var timestamps []Timestamp
	if err := res.All(&timestamps); err != nil {
		logger.Error.Fatal("Cannot resolve image count", err)
		return time.Unix(0, 0)
	} else if len(timestamps) > 0 {
		return timestamps[0].Timestamp
	} else {
		return time.Unix(0, 0)
	}
}

func (s *ImageStore) GetAllImages() ([]*apitype.ImageFile, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.getImagesInCategory(-1, 0, apitype.NoCategory)
}

func (s *ImageStore) GetAllImagesModifiedAfter(timestamp time.Time) ([]*apitype.ImageFile, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	var images []Image
	searchCondition := db.Cond{
		"modified_timestamp >": timestamp,
	}

	if err := s.getCollection().Find(searchCondition).OrderBy("name").All(&images); err != nil {
		return nil, err
	} else {
		return toImageFiles(images, s.database.BasePath()), nil
	}
}

func (s *ImageStore) GetNextImagesInCategory(number int, currentIndex int, categoryId apitype.CategoryId) ([]*apitype.ImageFile, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	startIndex := currentIndex + 1
	return s.getImagesInCategory(number, startIndex, categoryId)
}

func (s *ImageStore) GetPreviousImagesInCategory(number int, currentIndex int, categoryId apitype.CategoryId) ([]*apitype.ImageFile, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	prevIndex := currentIndex - number
	if prevIndex < 0 {
		prevIndex = 0
	}
	size := currentIndex - prevIndex

	if size < 0 {
		return []*apitype.ImageFile{}, nil
	} else if images, err := s.getImagesInCategory(size, prevIndex, categoryId); err != nil {
		return nil, err
	} else {
		util.Reverse(images)
		return images, nil
	}
}

func (s *ImageStore) GetImagesInCategory(number int, offset int, categoryId apitype.CategoryId) ([]*apitype.ImageFile, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.getImagesInCategory(number, offset, categoryId)
}

func (s *ImageStore) getImagesInCategory(number int, offset int, categoryId apitype.CategoryId) ([]*apitype.ImageFile, error) {
	if number == 0 {
		return make([]*apitype.ImageFile, 0), nil
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
		return toImageFiles(images, s.database.BasePath()), nil
	}
}

type sortDir string

const (
	asc  sortDir = "ASC"
	desc sortDir = "DESC"
)

func (s *ImageStore) FindByFileName(imageFile *apitype.ImageFile) (*apitype.ImageFile, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.findByFileName(s.getCollection(), imageFile)
}

func (s *ImageStore) findByFileName(collection db.Collection, imageFile *apitype.ImageFile) (*apitype.ImageFile, error) {
	var imageFiles []Image
	err := collection.
		Find(db.Cond{
			"file_name": imageFile.FileName(),
		}).
		All(&imageFiles)
	if err != nil {
		return nil, err
	} else if len(imageFiles) == 1 {
		return toImageFile(&imageFiles[0], s.database.BasePath())
	} else {
		return nil, nil
	}
}

func (s *ImageStore) Exists(imageFile *apitype.ImageFile) (bool, error) {
	return s.exists(s.getCollection(), imageFile)
}

func (s *ImageStore) exists(collection db.Collection, imageFile *apitype.ImageFile) (bool, error) {
	return collection.
		Find(db.Cond{
			"file_name": imageFile.FileName(),
		}).
		Exists()
}

func (s *ImageStore) getCollectionForSession(session db.Session) db.Collection {
	return session.Collection(s.getCollection().Name())
}

func (s *ImageStore) FindModifiedId(imageFile *apitype.ImageFile) (apitype.ImageId, error) {
	return s.findModifiedId(s.getCollection(), imageFile)
}

func (s *ImageStore) findModifiedId(collection db.Collection, imageFile *apitype.ImageFile) (apitype.ImageId, error) {
	stat, err := s.imageFileConverter.GetImageFileStats(imageFile)
	if err != nil {
		return apitype.NoImage, err
	}

	var images []Image
	err = collection.
		Find(db.Cond{
			"file_name":            imageFile.FileName(),
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

func (s *ImageStore) update(collection db.Collection, imageId apitype.ImageId, image *Image) error {
	return collection.Find(db.Cond{"id": imageId}).Update(image)
}

func (s *ImageStore) GetImageById(imageId apitype.ImageId) *apitype.ImageFile {
	s.mux.Lock()
	defer s.mux.Unlock()
	var image Image
	err := s.getCollection().Find(db.Cond{"id": imageId}).One(&image)

	if err != nil {
		logger.Error.Print("Could not find image ", err)
	}

	imageFile, err := toImageFile(&image, s.database.BasePath())
	if err != nil {
		logger.Error.Print("Could not find image ", err)
	}

	return imageFile

}

func (s *ImageStore) RemoveImage(imageId apitype.ImageId) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.getCollection().Find(db.Cond{"id": imageId}).Delete()
}
