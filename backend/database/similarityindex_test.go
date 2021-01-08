package database

import (
	"github.com/stretchr/testify/require"
	"github.com/upper/db/v4"
	"testing"
	"vincit.fi/image-sorter/api/apitype"
)

var (
	sut        *SimilarityIndex
	imageStore *ImageStore
)

func initSimilarityIndexTest() {
	store := NewInMemoryStore()
	sut = NewSimilarityIndex(store)
	imageStore = NewImageStore(store, &StubImageHandleConverter{})
}

func TestSimilarityIndex_AddAndGetSimilarImages(t *testing.T) {
	a := require.New(t)

	initSimilarityIndexTest()

	image1, _ := imageStore.AddImage(apitype.NewHandle("images", "image1"))
	image2, _ := imageStore.AddImage(apitype.NewHandle("images", "image2"))
	image3, _ := imageStore.AddImage(apitype.NewHandle("images", "image3"))
	image4, _ := imageStore.AddImage(apitype.NewHandle("images", "image4"))
	image5, _ := imageStore.AddImage(apitype.NewHandle("images", "image5"))

	err := sut.DoInTransaction(func(session db.Session) error {
		if err := sut.StartRecreateSimilarImageIndex(session); err != nil {
			return err
		} else if err = sut.AddSimilarImage(image1.GetId(), image2.GetId(), 0, -10.12); err != nil {
			return err
		} else if err = sut.AddSimilarImage(image1.GetId(), image3.GetId(), 1, 1); err != nil {
			return err
		} else if err = sut.AddSimilarImage(image1.GetId(), image4.GetId(), 2, 12); err != nil {
			return err
		} else if err = sut.AddSimilarImage(image2.GetId(), image4.GetId(), 0, 1); err != nil {
			return err
		}
		return sut.EndRecreateSimilarImageIndex()
	})
	a.Nil(err)

	t.Run("One similar image found for an image", func(t *testing.T) {
		images := sut.GetSimilarImages(image2.GetId())
		a.Equal(1, len(images))
		a.Equal(images[0].GetId(), image4.GetId())
	})

	t.Run("Many similar images found for an image", func(t *testing.T) {
		images := sut.GetSimilarImages(image1.GetId())
		a.Equal(3, len(images))
		a.Equal(images[0].GetId(), image2.GetId())
		a.Equal(images[1].GetId(), image3.GetId())
		a.Equal(images[2].GetId(), image4.GetId())
	})

	t.Run("No similar found for an image", func(t *testing.T) {
		images := sut.GetSimilarImages(image5.GetId())
		a.Equal(0, len(images))
	})

}
