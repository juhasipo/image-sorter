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
	database := NewInMemoryDatabase()
	sut = NewSimilarityIndex(database)
	imageStore = NewImageStore(database, &StubImageFileConverter{})
}

func TestSimilarityIndex_AddAndGetSimilarImages(t *testing.T) {
	a := require.New(t)

	initSimilarityIndexTest()

	image1, _ := imageStore.AddImage(apitype.NewImageFile("images", "image1"))
	image2, _ := imageStore.AddImage(apitype.NewImageFile("images", "image2"))
	image3, _ := imageStore.AddImage(apitype.NewImageFile("images", "image3"))
	image4, _ := imageStore.AddImage(apitype.NewImageFile("images", "image4"))
	image5, _ := imageStore.AddImage(apitype.NewImageFile("images", "image5"))

	err := sut.DoInTransaction(func(session db.Session) error {
		if err := sut.StartRecreateSimilarImageIndex(session); err != nil {
			return err
		} else if err = sut.AddSimilarImage(image1.GetImageId(), image2.GetImageId(), 0, -10.12); err != nil {
			return err
		} else if err = sut.AddSimilarImage(image1.GetImageId(), image3.GetImageId(), 1, 1); err != nil {
			return err
		} else if err = sut.AddSimilarImage(image1.GetImageId(), image4.GetImageId(), 2, 12); err != nil {
			return err
		} else if err = sut.AddSimilarImage(image2.GetImageId(), image4.GetImageId(), 0, 1); err != nil {
			return err
		}
		return sut.EndRecreateSimilarImageIndex()
	})
	a.Nil(err)

	t.Run("One similar image found for an image", func(t *testing.T) {
		images := sut.GetSimilarImages(image2.GetImageId())
		a.Equal(1, len(images))
		a.Equal(images[0].GetImageId(), image4.GetImageId())
	})

	t.Run("Many similar images found for an image", func(t *testing.T) {
		images := sut.GetSimilarImages(image1.GetImageId())
		a.Equal(3, len(images))
		a.Equal(images[0].GetImageId(), image2.GetImageId())
		a.Equal(images[1].GetImageId(), image3.GetImageId())
		a.Equal(images[2].GetImageId(), image4.GetImageId())
	})

	t.Run("No similar found for an image", func(t *testing.T) {
		images := sut.GetSimilarImages(image5.GetImageId())
		a.Equal(0, len(images))
	})

}
