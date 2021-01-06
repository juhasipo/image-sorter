package database

import (
	"github.com/upper/db/v4"
	"time"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
)

type SimilarityIndex struct {
	session    db.Session
	collection db.Collection
}

func NewSimilarityIndex(store *Store) *SimilarityIndex {
	return &SimilarityIndex{
		collection: store.imageSimilarCollection,
	}
}

func (s *SimilarityIndex) DoInTransaction(fn func(session db.Session) error) error {
	return s.collection.Session().Tx(fn)
}

func (s *SimilarityIndex) StartRecreateSimilarImageIndex(session db.Session) error {
	s.session = session

	logger.Debug.Print("Truncate similar image index")
	if err := s.collection.Truncate(); err != nil {
		return err
	}

	logger.Debug.Print("Dropping index")
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
	logger.Debug.Print("Creating indices for similar images")
	if _, err := s.session.SQL().Exec("CREATE UNIQUE INDEX image_similar_uq ON image_similar(image_id, similar_image_id)"); err != nil {
		return err
	}

	end := time.Now()
	logger.Debug.Printf(" - Creating index: %s", end.Sub(start))
	return nil
}

func (s *SimilarityIndex) AddSimilarImage(imageId apitype.HandleId, similarId apitype.HandleId, rank int, score float64) error {
	collection := s.session.Collection(s.collection.Name())
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

func (s *SimilarityIndex) GetSimilarImages(imageId apitype.HandleId) []*apitype.Handle {
	var images []Image
	s.collection.Session().SQL().
		Select("image.*").
		From("image").
		Join("image_similar").On("image_similar.similar_image_id = image.id").
		Where("image_similar.image_id", imageId).
		OrderBy("image_similar.rank").
		All(&images)

	return toApiHandles(images)
}
