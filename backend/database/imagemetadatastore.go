package database

import (
	"github.com/upper/db/v4"
	"time"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
)

type ImageMetaDataStore struct {
	database   *Database
	collection db.Collection
}

func NewImageMetaDataStore(database *Database) *ImageMetaDataStore {
	return &ImageMetaDataStore{
		database: database,
	}
}

func (s *ImageMetaDataStore) getCollection() db.Collection {
	if s.collection == nil {
		s.collection = s.database.Session().Collection("image_meta_data")
	}
	return s.collection
}

func (s *ImageMetaDataStore) getCollectionForSession(session db.Session) db.Collection {
	return session.Collection(s.getCollection().Name())
}

func (s *ImageMetaDataStore) GetMetaDataByImageId(imageId apitype.ImageId) (*apitype.ImageMetaData, error) {
	var metaData []ImageMetaData
	err := s.getCollection().Find(db.Cond{"image_id": imageId}).All(&metaData)

	if err != nil {
		logger.Error.Print("Could not find image meta data ", err)
	}

	var md = map[string]string{}
	for _, m := range metaData {
		md[m.Key] = m.Value
	}

	return apitype.NewImageMetaData(md), nil
}

func (s *ImageMetaDataStore) GetAllImagesWithoutMetaData() ([]*apitype.ImageFile, error) {
	var images []Image
	res := s.getCollection().Session().SQL().
		Select("image.*").
		From("image")

	res = res.
		LeftJoin("image_meta_data").On("image_meta_data.image_id = image.id").
		Where("image_meta_data.image_id IS NULL").
		OrderBy("image.name")

	if err := res.All(&images); err != nil {
		return nil, err
	} else {
		return toImageFiles(images), nil
	}
}

func (s *ImageMetaDataStore) AddMetaData(imageId apitype.ImageId, metaData *apitype.ImageMetaData) error {
	return s.getCollection().Session().Tx(func(sess db.Session) error {
		for key, value := range metaData.MetaData() {
			if _, err := s.addMetaData(sess, imageId, key, value); err != nil {
				logger.Error.Printf("Error while adding image metadata for ImageId '%d'", imageId)
				return err
			}
		}
		return nil
	})
}

func (s *ImageMetaDataStore) AddMetaDataForImages(images []*apitype.ImageFile, loadExifCb func(imageFile *apitype.ImageFile) (*apitype.ExifData, error)) error {
	return s.getCollection().Session().Tx(func(session db.Session) error {
		if err := s.startAddingMetaData(session); err != nil {
			return err
		} else {
			for _, image := range images {
				if err := s.clearMetaDataForImage(session, image.Id()); err != nil {
					logger.Error.Print("cannot remove old meta data ", image, err)
					return err
				} else if data, err := loadExifCb(image); err != nil {
					logger.Error.Print("cannot read meta data ", image, err)
					return err
				} else if data != nil && data.Values() != nil {
					metaData := apitype.NewImageMetaData(data.Values())
					for key, value := range metaData.MetaData() {
						if _, err := s.addMetaData(session, image.Id(), key, value); err != nil {
							logger.Error.Printf("Error while adding image metadata for ImageId '%d'", image.Id())
							return err
						}
					}
				}
			}
		}

		if err := s.endAddingMetaData(session); err != nil {
			return err
		}
		return nil
	})
}

func (s *ImageMetaDataStore) addMetaData(session db.Session, imageId apitype.ImageId, key string, value string) (*db.InsertResult, error) {
	collection := s.getCollectionForSession(session)

	logger.Trace.Printf("Adding metadata for imageId %d", imageId)

	return collection.Insert(&ImageMetaData{ImageId: imageId, Key: key, Value: value})
}

func (s *ImageMetaDataStore) startAddingMetaData(session db.Session) error {
	logger.Trace.Print("Dropping image meta data indices")
	if _, err := session.SQL().Exec("DROP INDEX IF EXISTS image_meta_data_idx"); err != nil {
		return err
	} else if _, err := session.SQL().Exec("DROP INDEX IF EXISTS image_meta_data_uq"); err != nil {
		return err
	} else {
		return nil
	}
}

func (s *ImageMetaDataStore) endAddingMetaData(session db.Session) error {
	start := time.Now()
	logger.Trace.Print("Creating indices for image meta data")
	if _, err := session.SQL().Exec("CREATE INDEX image_meta_data_idx ON image_meta_data (key, value)"); err != nil {
		return err
	} else if _, err := session.SQL().Exec("CREATE UNIQUE INDEX image_meta_data_uq ON image_meta_data (image_id, key)"); err != nil {
		return err
	}

	end := time.Now()
	logger.Trace.Printf(" - Creating indices: %s", end.Sub(start))
	return nil
}

func (s *ImageMetaDataStore) clearMetaDataForImage(session db.Session, imageId apitype.ImageId) error {
	collection := s.getCollectionForSession(session)
	return collection.Find(db.Cond{"image_id": imageId}).Delete()
}
