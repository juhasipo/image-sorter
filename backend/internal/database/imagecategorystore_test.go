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
			imagesCategories, err := sut.GetImagesCategories(images[0].Id())
			a.Nil(err)
			a.Equal(0, len(imagesCategories))
		})

		t.Run("One category", func(t *testing.T) {
			err := sut.CategorizeImage(images[0].Id(), categories[0].Id(), apitype.CATEGORIZE)
			a.Nil(err)

			imagesCategories, err := sut.GetImagesCategories(images[0].Id())
			a.Nil(err)
			a.Equal(1, len(imagesCategories))
			a.Equal(apitype.CATEGORIZE, imagesCategories[0].Operation)
			a.Equal(categories[0].Id(), imagesCategories[0].Category.Id())
		})

		t.Run("Many categories", func(t *testing.T) {
			err := sut.CategorizeImage(images[0].Id(), categories[1].Id(), apitype.CATEGORIZE)
			a.Nil(err)
			err = sut.CategorizeImage(images[0].Id(), categories[2].Id(), apitype.CATEGORIZE)
			a.Nil(err)

			imagesCategories, err := sut.GetImagesCategories(images[0].Id())
			a.Nil(err)
			a.Equal(3, len(imagesCategories))
			a.Equal(apitype.CATEGORIZE, imagesCategories[0].Operation)
			a.Equal(categories[0].Id(), imagesCategories[0].Category.Id())
			a.Equal(categories[1].Id(), imagesCategories[1].Category.Id())
			a.Equal(categories[2].Id(), imagesCategories[2].Category.Id())
		})

		t.Run("Many categories many images", func(t *testing.T) {
			err := sut.CategorizeImage(images[1].Id(), categories[1].Id(), apitype.CATEGORIZE)
			a.Nil(err)
			err = sut.CategorizeImage(images[2].Id(), categories[1].Id(), apitype.CATEGORIZE)
			a.Nil(err)
			err = sut.CategorizeImage(images[2].Id(), categories[2].Id(), apitype.CATEGORIZE)
			a.Nil(err)

			imagesCategories, err := sut.GetImagesCategories(images[0].Id())
			a.Nil(err)
			a.Equal(3, len(imagesCategories))
			a.Equal(apitype.CATEGORIZE, imagesCategories[0].Operation)
			a.Equal(categories[0].Id(), imagesCategories[0].Category.Id())
			a.Equal(categories[1].Id(), imagesCategories[1].Category.Id())
			a.Equal(categories[2].Id(), imagesCategories[2].Category.Id())

			imagesCategories, err = sut.GetImagesCategories(images[1].Id())
			a.Nil(err)
			a.Equal(1, len(imagesCategories))
			a.Equal(apitype.CATEGORIZE, imagesCategories[0].Operation)
			a.Equal(categories[1].Id(), imagesCategories[0].Category.Id())

			imagesCategories, err = sut.GetImagesCategories(images[2].Id())
			a.Nil(err)
			a.Equal(2, len(imagesCategories))
			a.Equal(apitype.CATEGORIZE, imagesCategories[0].Operation)
			a.Equal(categories[1].Id(), imagesCategories[0].Category.Id())
			a.Equal(categories[2].Id(), imagesCategories[1].Category.Id())
		})
	})

	t.Run("Re-categorize", func(t *testing.T) {
		sut := initImageCategoryStoreTest()

		images := createImages()
		categories := createCategories()

		err := sut.CategorizeImage(images[0].Id(), categories[1].Id(), apitype.CATEGORIZE)
		a.Nil(err)
		err = sut.CategorizeImage(images[0].Id(), categories[1].Id(), apitype.CATEGORIZE)
		a.Nil(err)

		imagesCategories, err := sut.GetImagesCategories(images[0].Id())
		a.Nil(err)
		a.Equal(1, len(imagesCategories))
		a.Equal(apitype.CATEGORIZE, imagesCategories[0].Operation)
		a.Equal(categories[1].Id(), imagesCategories[0].Category.Id())
	})

	t.Run("None category", func(t *testing.T) {
		sut := initImageCategoryStoreTest()

		images := createImages()
		categories := createCategories()

		err := sut.CategorizeImage(images[0].Id(), categories[0].Id(), apitype.CATEGORIZE)
		a.Nil(err)
		err = sut.CategorizeImage(images[0].Id(), categories[1].Id(), apitype.CATEGORIZE)
		a.Nil(err)
		err = sut.CategorizeImage(images[0].Id(), categories[2].Id(), apitype.CATEGORIZE)
		a.Nil(err)

		t.Run("First time", func(t *testing.T) {
			err = sut.CategorizeImage(images[0].Id(), categories[1].Id(), apitype.UNCATEGORIZE)
			a.Nil(err)

			imagesCategories, err := sut.GetImagesCategories(images[0].Id())
			a.Nil(err)
			a.Equal(2, len(imagesCategories))
			a.Equal(apitype.CATEGORIZE, imagesCategories[0].Operation)
			a.Equal(categories[0].Id(), imagesCategories[0].Category.Id())
			a.Equal(categories[2].Id(), imagesCategories[1].Category.Id())
		})

		t.Run("Second time", func(t *testing.T) {
			err = sut.CategorizeImage(images[0].Id(), categories[1].Id(), apitype.UNCATEGORIZE)
			a.Nil(err)

			imagesCategories, err := sut.GetImagesCategories(images[0].Id())
			a.Nil(err)
			a.Equal(2, len(imagesCategories))
			a.Equal(apitype.CATEGORIZE, imagesCategories[0].Operation)
			a.Equal(categories[0].Id(), imagesCategories[0].Category.Id())
			a.Equal(categories[2].Id(), imagesCategories[1].Category.Id())
		})
	})
}

func TestImageCategoryStore_RemoveImageCategories(t *testing.T) {
	a := require.New(t)

	sut := initImageCategoryStoreTest()

	images := createImages()
	categories := createCategories()

	err := sut.CategorizeImage(images[0].Id(), categories[1].Id(), apitype.CATEGORIZE)
	a.Nil(err)
	err = sut.CategorizeImage(images[0].Id(), categories[2].Id(), apitype.CATEGORIZE)
	a.Nil(err)

	err = sut.CategorizeImage(images[1].Id(), categories[0].Id(), apitype.CATEGORIZE)
	a.Nil(err)
	err = sut.CategorizeImage(images[1].Id(), categories[1].Id(), apitype.CATEGORIZE)
	a.Nil(err)

	imagesCategories, err := sut.GetImagesCategories(images[0].Id())
	a.Nil(err)
	a.Equal(2, len(imagesCategories))

	err = sut.RemoveImageCategories(images[0].Id())
	a.Nil(err)

	imagesCategories, err = sut.GetImagesCategories(images[0].Id())
	a.Nil(err)
	a.Equal(0, len(imagesCategories))

	imagesCategories, err = sut.GetImagesCategories(images[1].Id())
	a.Nil(err)
	a.Equal(2, len(imagesCategories))

}

func TestImageCategoryStore_RemoveImageRemovesCategories(t *testing.T) {
	a := require.New(t)

	sut := initImageCategoryStoreTest()

	images := createImages()
	categories := createCategories()

	err := sut.CategorizeImage(images[0].Id(), categories[0].Id(), apitype.CATEGORIZE)
	a.Nil(err)
	err = sut.CategorizeImage(images[0].Id(), categories[1].Id(), apitype.CATEGORIZE)
	a.Nil(err)
	err = sut.CategorizeImage(images[1].Id(), categories[1].Id(), apitype.CATEGORIZE)
	a.Nil(err)
	err = sut.CategorizeImage(images[1].Id(), categories[2].Id(), apitype.CATEGORIZE)
	a.Nil(err)
	err = sut.CategorizeImage(images[2].Id(), categories[2].Id(), apitype.CATEGORIZE)
	a.Nil(err)

	imagesCategories, err := sut.GetImagesCategories(images[1].Id())
	a.Nil(err)
	a.Equal(2, len(imagesCategories))

	_ = icsImageStore.RemoveImage(images[1].Id())
	reinserted, _ := icsImageStore.AddImage(images[1])

	imagesCategories, err = sut.GetImagesCategories(reinserted.Id())
	a.Nil(err)
	a.Equal(0, len(imagesCategories))
}

func TestImageCategoryStore_GetCategorizedImages(t *testing.T) {
	a := require.New(t)

	sut := initImageCategoryStoreTest()

	images := createImages()
	categories := createCategories()

	err := sut.CategorizeImage(images[0].Id(), categories[0].Id(), apitype.CATEGORIZE)
	a.Nil(err)
	err = sut.CategorizeImage(images[0].Id(), categories[1].Id(), apitype.CATEGORIZE)
	a.Nil(err)

	err = sut.CategorizeImage(images[1].Id(), categories[0].Id(), apitype.CATEGORIZE)
	a.Nil(err)
	err = sut.CategorizeImage(images[1].Id(), categories[1].Id(), apitype.CATEGORIZE)
	a.Nil(err)
	err = sut.CategorizeImage(images[1].Id(), categories[2].Id(), apitype.CATEGORIZE)
	a.Nil(err)

	err = sut.CategorizeImage(images[2].Id(), categories[2].Id(), apitype.CATEGORIZE)
	a.Nil(err)

	imagesCategories, err := sut.GetCategorizedImages()
	a.Nil(err)
	a.Equal(3, len(imagesCategories))

	a.Equal(2, len(imagesCategories[images[0].Id()]))
	a.Equal("C1", imagesCategories[images[0].Id()][categories[0].Id()].Category.Name())
	a.Equal("C2", imagesCategories[images[0].Id()][categories[1].Id()].Category.Name())

	a.Equal(3, len(imagesCategories[images[1].Id()]))
	a.Equal("C1", imagesCategories[images[1].Id()][categories[0].Id()].Category.Name())
	a.Equal("C2", imagesCategories[images[1].Id()][categories[1].Id()].Category.Name())
	a.Equal("C3", imagesCategories[images[1].Id()][categories[2].Id()].Category.Name())

	a.Equal(1, len(imagesCategories[images[2].Id()]))
	a.Equal("C3", imagesCategories[images[2].Id()][categories[2].Id()].Category.Name())
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
		image1,
		image2,
		image3,
		image4,
		image5,
	}
}
