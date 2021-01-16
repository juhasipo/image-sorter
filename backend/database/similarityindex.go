package database

import (
	"github.com/upper/db/v4"
	"time"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
)

type SimilarityIndex struct {
	database   *Database
	session    db.Session
	collection db.Collection
}

func NewSimilarityIndex(database *Database) *SimilarityIndex {
	return &SimilarityIndex{
		database: database,
	}
}

func (s *SimilarityIndex) getCollection() db.Collection {
	if s.collection == nil {
		s.collection = s.database.Session().Collection("image_similar")
	}
	return s.collection
}

func (s *SimilarityIndex) DoInTransaction(fn func(session db.Session) error) error {
	return s.getCollection().Session().Tx(fn)
}

func (s *SimilarityIndex) StartRecreateSimilarImageIndex(session db.Session) error {
	s.session = session
	collection := s.session.Collection(s.getCollection().Name())

	logger.Trace.Print("Truncate similar image index")
	if err := collection.Truncate(); err != nil {
		return err
	}

	logger.Trace.Print("Dropping index")
	if _, err := s.session.SQL().Exec("DROP INDEX IF EXISTS image_similar_uq"); err != nil {
		return err
	} else {
		return nil
	}
}

func (s *SimilarityIndex) EndRecreateSimilarImageIndex() error {
	defer func() {
		s.session = nil
	}()

	start := time.Now()
	logger.Trace.Print("Creating indices for similar images")
	if _, err := s.session.SQL().Exec("CREATE UNIQUE INDEX image_similar_uq ON image_similar(image_id, similar_image_id)"); err != nil {
		return err
	}

	end := time.Now()
	logger.Trace.Printf(" - Creating index: %s", end.Sub(start))
	return nil
}

func (s *SimilarityIndex) AddSimilarImage(imageId apitype.ImageId, similarId apitype.ImageId, rank int, score float64) error {
	collection := s.session.Collection(s.getCollection().Name())
	_, err := collection.Insert(&ImageSimilar{
		ImageId:        imageId,
		SimilarImageId: similarId,
		Rank:           rank,
		Score:          score,
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *SimilarityIndex) GetSimilarImages(imageId apitype.ImageId) []*apitype.Handle {
	var images []Image
	s.getCollection().Session().SQL().
		Select("image.*").
		From("image").
		Join("image_similar").On("image_similar.similar_image_id = image.id").
		Where("image_similar.image_id", imageId).
		OrderBy("image_similar.rank").
		All(&images)

	return toApiHandles(images)
}

func (s *SimilarityIndex) GetIndexSize() (uint64, error) {
	return s.getCollection().Count()
}
