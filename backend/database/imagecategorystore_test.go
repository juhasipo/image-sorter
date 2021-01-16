package database

import (
	"github.com/stretchr/testify/require"
	"testing"
	"vincit.fi/image-sorter/api/apitype"
)

var (
	icsImageStore    *ImageStore
	icsCategoryStore *CategoryStore
)

func initImageCategoryStoreTest() *ImageCategoryStore {
	database := NewInMemoryDatabase()
	icsImageStore = NewImageStore(database, &StubImageFileConverter{})
	icsCategoryStore = NewCategoryStore(database)

	return NewImageCategoryStore(database)
}

func TestImageCategoryStore_CategorizeImage(t *testing.T) {
	a := require.New(t)

	t.Run("Simple cases", func(t *testing.T) {
		sut := initImageCategoryStoreTest()

		images := createImages()
		categories := createCategories()

		t.Run("No categories", func(t *testing.T) {
			imagesCategories, err := sut.GetImagesCategories(images[0].GetId())
			a.Nil(err)
			a.Equal(0, len(imagesCategories))
		})

		t.Run("One category", func(t *testing.T) {
			err := sut.CategorizeImage(images[0].GetId(), categories[0].GetId(), apitype.MOVE)
			a.Nil(err)

			imagesCategories, err := sut.GetImagesCategories(images[0].GetId())
			a.Nil(err)
			a.Equal(1, len(imagesCategories))
			a.Equal(apitype.MOVE, imagesCategories[0].Operation)
			a.Equal(categories[0].GetId(), imagesCategories[0].Category.GetId())
		})

		t.Run("Many categories", func(t *testing.T) {
			err := sut.CategorizeImage(images[0].GetId(), categories[1].GetId(), apitype.MOVE)
			a.Nil(err)
			err = sut.CategorizeImage(images[0].GetId(), categories[2].GetId(), apitype.MOVE)
			a.Nil(err)

			imagesCategories, err := sut.GetImagesCategories(images[0].GetId())
			a.Nil(err)
			a.Equal(3, len(imagesCategories))
			a.Equal(apitype.MOVE, imagesCategories[0].Operation)
			a.Equal(categories[0].GetId(), imagesCategories[0].Category.GetId())
			a.Equal(categories[1].GetId(), imagesCategories[1].Category.GetId())
			a.Equal(categories[2].GetId(), imagesCategories[2].Category.GetId())
		})

		t.Run("Many categories many images", func(t *testing.T) {
			err := sut.CategorizeImage(images[1].GetId(), categories[1].GetId(), apitype.MOVE)
			a.Nil(err)
			err = sut.CategorizeImage(images[2].GetId(), categories[1].GetId(), apitype.MOVE)
			a.Nil(err)
			err = sut.CategorizeImage(images[2].GetId(), categories[2].GetId(), apitype.MOVE)
			a.Nil(err)

			imagesCategories, err := sut.GetImagesCategories(images[0].GetId())
			a.Nil(err)
			a.Equal(3, len(imagesCategories))
			a.Equal(apitype.MOVE, imagesCategories[0].Operation)
			a.Equal(categories[0].GetId(), imagesCategories[0].Category.GetId())
			a.Equal(categories[1].GetId(), imagesCategories[1].Category.GetId())
			a.Equal(categories[2].GetId(), imagesCategories[2].Category.GetId())

			imagesCategories, err = sut.GetImagesCategories(images[1].GetId())
			a.Nil(err)
			a.Equal(1, len(imagesCategories))
			a.Equal(apitype.MOVE, imagesCategories[0].Operation)
			a.Equal(categories[1].GetId(), imagesCategories[0].Category.GetId())

			imagesCategories, err = sut.GetImagesCategories(images[2].GetId())
			a.Nil(err)
			a.Equal(2, len(imagesCategories))
			a.Equal(apitype.MOVE, imagesCategories[0].Operation)
			a.Equal(categories[1].GetId(), imagesCategories[0].Category.GetId())
			a.Equal(categories[2].GetId(), imagesCategories[1].Category.GetId())
		})
	})

	t.Run("Re-categorize", func(t *testing.T) {
		sut := initImageCategoryStoreTest()

		images := createImages()
		categories := createCategories()

		err := sut.CategorizeImage(images[0].GetId(), categories[1].GetId(), apitype.MOVE)
		a.Nil(err)
		err = sut.CategorizeImage(images[0].GetId(), categories[1].GetId(), apitype.MOVE)
		a.Nil(err)

		imagesCategories, err := sut.GetImagesCategories(images[0].GetId())
		a.Nil(err)
		a.Equal(1, len(imagesCategories))
		a.Equal(apitype.MOVE, imagesCategories[0].Operation)
		a.Equal(categories[1].GetId(), imagesCategories[0].Category.GetId())
	})

	t.Run("None category", func(t *testing.T) {
		sut := initImageCategoryStoreTest()

		images := createImages()
		categories := createCategories()

		err := sut.CategorizeImage(images[0].GetId(), categories[0].GetId(), apitype.MOVE)
		a.Nil(err)
		err = sut.CategorizeImage(images[0].GetId(), categories[1].GetId(), apitype.MOVE)
		a.Nil(err)
		err = sut.CategorizeImage(images[0].GetId(), categories[2].GetId(), apitype.MOVE)
		a.Nil(err)

		t.Run("First time", func(t *testing.T) {
			err = sut.CategorizeImage(images[0].GetId(), categories[1].GetId(), apitype.NONE)
			a.Nil(err)

			imagesCategories, err := sut.GetImagesCategories(images[0].GetId())
			a.Nil(err)
			a.Equal(2, len(imagesCategories))
			a.Equal(apitype.MOVE, imagesCategories[0].Operation)
			a.Equal(categories[0].GetId(), imagesCategories[0].Category.GetId())
			a.Equal(categories[2].GetId(), imagesCategories[1].Category.GetId())
		})

		t.Run("Second time", func(t *testing.T) {
			err = sut.CategorizeImage(images[0].GetId(), categories[1].GetId(), apitype.NONE)
			a.Nil(err)

			imagesCategories, err := sut.GetImagesCategories(images[0].GetId())
			a.Nil(err)
			a.Equal(2, len(imagesCategories))
			a.Equal(apitype.MOVE, imagesCategories[0].Operation)
			a.Equal(categories[0].GetId(), imagesCategories[0].Category.GetId())
			a.Equal(categories[2].GetId(), imagesCategories[1].Category.GetId())
		})
	})
}

func TestImageCategoryStore_RemoveImageCategories(t *testing.T) {
	a := require.New(t)

	sut := initImageCategoryStoreTest()

	images := createImages()
	categories := createCategories()

	err := sut.CategorizeImage(images[0].GetId(), categories[1].GetId(), apitype.MOVE)
	a.Nil(err)
	err = sut.CategorizeImage(images[0].GetId(), categories[2].GetId(), apitype.MOVE)
	a.Nil(err)

	err = sut.CategorizeImage(images[1].GetId(), categories[0].GetId(), apitype.MOVE)
	a.Nil(err)
	err = sut.CategorizeImage(images[1].GetId(), categories[1].GetId(), apitype.MOVE)
	a.Nil(err)

	imagesCategories, err := sut.GetImagesCategories(images[0].GetId())
	a.Nil(err)
	a.Equal(2, len(imagesCategories))

	err = sut.RemoveImageCategories(images[0].GetId())
	a.Nil(err)

	imagesCategories, err = sut.GetImagesCategories(images[0].GetId())
	a.Nil(err)
	a.Equal(0, len(imagesCategories))

	imagesCategories, err = sut.GetImagesCategories(images[1].GetId())
	a.Nil(err)
	a.Equal(2, len(imagesCategories))

}

func TestImageCategoryStore_RemoveImageRemovesCategories(t *testing.T) {
	a := require.New(t)

	sut := initImageCategoryStoreTest()

	images := createImages()
	categories := createCategories()

	err := sut.CategorizeImage(images[0].GetId(), categories[0].GetId(), apitype.MOVE)
	a.Nil(err)
	err = sut.CategorizeImage(images[0].GetId(), categories[1].GetId(), apitype.MOVE)
	a.Nil(err)
	err = sut.CategorizeImage(images[1].GetId(), categories[1].GetId(), apitype.MOVE)
	a.Nil(err)
	err = sut.CategorizeImage(images[1].GetId(), categories[2].GetId(), apitype.MOVE)
	a.Nil(err)
	err = sut.CategorizeImage(images[2].GetId(), categories[2].GetId(), apitype.MOVE)
	a.Nil(err)

	imagesCategories, err := sut.GetImagesCategories(images[1].GetId())
	a.Nil(err)
	a.Equal(2, len(imagesCategories))

	_ = icsImageStore.RemoveImage(images[1].GetId())
	reinserted, _ := icsImageStore.AddImage(images[1])

	imagesCategories, err = sut.GetImagesCategories(reinserted.GetId())
	a.Nil(err)
	a.Equal(0, len(imagesCategories))
}

func TestImageCategoryStore_GetCategorizedImages(t *testing.T) {
	a := require.New(t)

	sut := initImageCategoryStoreTest()

	images := createImages()
	categories := createCategories()

	err := sut.CategorizeImage(images[0].GetId(), categories[0].GetId(), apitype.MOVE)
	a.Nil(err)
	err = sut.CategorizeImage(images[0].GetId(), categories[1].GetId(), apitype.MOVE)
	a.Nil(err)

	err = sut.CategorizeImage(images[1].GetId(), categories[0].GetId(), apitype.MOVE)
	a.Nil(err)
	err = sut.CategorizeImage(images[1].GetId(), categories[1].GetId(), apitype.MOVE)
	a.Nil(err)
	err = sut.CategorizeImage(images[1].GetId(), categories[2].GetId(), apitype.MOVE)
	a.Nil(err)

	err = sut.CategorizeImage(images[2].GetId(), categories[2].GetId(), apitype.MOVE)
	a.Nil(err)

	imagesCategories, err := sut.GetCategorizedImages()
	a.Nil(err)
	a.Equal(3, len(imagesCategories))

	a.Equal(2, len(imagesCategories[images[0].GetId()]))
	a.Equal("C1", imagesCategories[images[0].GetId()][categories[0].GetId()].Category.GetName())
	a.Equal("C2", imagesCategories[images[0].GetId()][categories[1].GetId()].Category.GetName())

	a.Equal(3, len(imagesCategories[images[1].GetId()]))
	a.Equal("C1", imagesCategories[images[1].GetId()][categories[0].GetId()].Category.GetName())
	a.Equal("C2", imagesCategories[images[1].GetId()][categories[1].GetId()].Category.GetName())
	a.Equal("C3", imagesCategories[images[1].GetId()][categories[2].GetId()].Category.GetName())

	a.Equal(1, len(imagesCategories[images[2].GetId()]))
	a.Equal("C3", imagesCategories[images[2].GetId()][categories[2].GetId()].Category.GetName())
}

func createCategories() []*apitype.Category {
	category1, _ := icsCategoryStore.AddCategory(apitype.NewCategory("C1", "c1", "1"))
	category2, _ := icsCategoryStore.AddCategory(apitype.NewCategory("C2", "c2", "2"))
	category3, _ := icsCategoryStore.AddCategory(apitype.NewCategory("C3", "c3", "3"))

	return []*apitype.Category{category1, category2, category3}
}

func createImages() []*apitype.ImageFile {
	image1, _ := icsImageStore.AddImage(apitype.NewImageFile("images", "image1"))
	image2, _ := icsImageStore.AddImage(apitype.NewImageFile("images", "image2"))
	image3, _ := icsImageStore.AddImage(apitype.NewImageFile("images", "image3"))
	image4, _ := icsImageStore.AddImage(apitype.NewImageFile("images", "image4"))
	image5, _ := icsImageStore.AddImage(apitype.NewImageFile("images", "image5"))
	return []*apitype.ImageFile{
		&image1.ImageFile,
		&image2.ImageFile,
		&image3.ImageFile,
		&image4.ImageFile,
		&image5.ImageFile,
	}
}
