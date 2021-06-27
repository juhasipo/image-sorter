package database

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
	"vincit.fi/image-sorter/api/apitype"
)

var (
	imageStoreImageFileConverter *StubImageFileConverter
	isCategoryStore              *CategoryStore
	isImageCategoryStore         *ImageCategoryStore
)

func initImageStoreTest() *ImageStore {
	database := NewInMemoryDatabase()
	imageStoreImageFileConverter = &StubImageFileConverter{}
	isCategoryStore = NewCategoryStore(database)
	isImageCategoryStore = NewImageCategoryStore(database)
	return NewImageStore(database, imageStoreImageFileConverter)
}

func TestImageStore_AddImage_GetImageById(t *testing.T) {
	a := require.New(t)

	t.Run("Add image and get it by ID", func(t *testing.T) {
		sut := initImageStoreTest()

		image1, err := sut.AddImage(apitype.NewImageFile("images", "image1"))
		_, err = sut.AddImage(apitype.NewImageFile("images", "image2"))

		a.Nil(err)

		image := sut.GetImageById(image1.Id())

		a.Equal(image1.Id(), image.Id())
		a.Equal(image1.FileName(), image.FileName())
		a.Equal(image1.Directory(), image.Directory())
		a.Equal(image1.ByteSize(), image.ByteSize())
		a.Equal(image1.Path(), image.Path())
	})

	t.Run("Re-add same image not modified", func(t *testing.T) {
		sut := initImageStoreTest()

		image1, err := sut.AddImage(apitype.NewImageFile("images", "image1"))
		a.Nil(err)
		_, err = sut.AddImage(apitype.NewImageFile("images", "image1"))
		a.Nil(err)

		images, err := sut.GetAllImages()
		a.Equal(1, len(images))

		image := sut.GetImageById(image1.Id())

		a.Equal(image1.Id(), image.Id())
		a.Equal(image1.FileName(), image.FileName())
		a.Equal(image1.Directory(), image.Directory())
		a.Equal(image1.ByteSize(), image.ByteSize())
		a.Equal(image1.Path(), image.Path())
	})

	t.Run("Re-add same image modified", func(t *testing.T) {
		sut := initImageStoreTest()

		imageStoreImageFileConverter.SetIncrementModTimeRequest(true)
		image1, err := sut.AddImage(apitype.NewImageFile("images", "image1"))
		a.Nil(err)
		_, err = sut.AddImage(apitype.NewImageFile("images", "image1"))
		a.Nil(err)

		images, err := sut.GetAllImages()
		a.Equal(1, len(images))

		image := sut.GetImageById(image1.Id())

		a.Equal(image1.Id(), image.Id())
		a.Equal(image1.FileName(), image.FileName())
		a.Equal(image1.Directory(), image.Directory())
		a.Equal(image1.ByteSize(), image.ByteSize())
		a.Equal(image1.Path(), image.Path())
	})
}

func TestImageStore_GetNextImagesInCategory_NoCategorySet(t *testing.T) {
	a := assert.New(t)

	t.Run("No images", func(t *testing.T) {
		sut := initImageStoreTest()

		t.Run("Query ", func(t *testing.T) {
			images, err := sut.GetNextImagesInCategory(5, 0, apitype.NoCategory)
			a.Nil(err)
			a.Equal(0, len(images))
		})
	})

	t.Run("Next images without category", func(t *testing.T) {
		sut := initImageStoreTest()
		err := sut.AddImages([]*apitype.ImageFile{
			apitype.NewImageFile("images", "image0"),
			apitype.NewImageFile("images", "image1"),
			apitype.NewImageFile("images", "image2"),
			apitype.NewImageFile("images", "image3"),
			apitype.NewImageFile("images", "image4"),
			apitype.NewImageFile("images", "image5"),
			apitype.NewImageFile("images", "image6"),
		})

		a.Nil(err)

		t.Run("Query less than max images", func(t *testing.T) {
			images, err := sut.GetNextImagesInCategory(5, 0, apitype.NoCategory)
			a.Nil(err)
			if a.Equal(5, len(images)) {
				a.Equal("image1", images[0].FileName())
				a.Equal("image2", images[1].FileName())
				a.Equal("image3", images[2].FileName())
				a.Equal("image4", images[3].FileName())
				a.Equal("image5", images[4].FileName())
			}
		})

		t.Run("Query more than max images", func(t *testing.T) {
			images, err := sut.GetNextImagesInCategory(10, 0, apitype.NoCategory)
			a.Nil(err)
			if a.Equal(6, len(images)) {
				a.Equal("image1", images[0].FileName())
				a.Equal("image2", images[1].FileName())
				a.Equal("image3", images[2].FileName())
				a.Equal("image4", images[3].FileName())
				a.Equal("image5", images[4].FileName())
				a.Equal("image6", images[5].FileName())
			}
		})

		t.Run("Query less than max images with offset", func(t *testing.T) {
			images, err := sut.GetNextImagesInCategory(4, 2, apitype.NoCategory)
			a.Nil(err)
			if a.Equal(4, len(images)) {
				a.Equal("image3", images[0].FileName())
				a.Equal("image4", images[1].FileName())
				a.Equal("image5", images[2].FileName())
				a.Equal("image6", images[3].FileName())
			}
		})

		t.Run("Query more than max images with offset", func(t *testing.T) {
			images, err := sut.GetNextImagesInCategory(10, 2, apitype.NoCategory)
			a.Nil(err)
			if a.Equal(4, len(images)) {
				a.Equal("image3", images[0].FileName())
				a.Equal("image4", images[1].FileName())
				a.Equal("image5", images[2].FileName())
				a.Equal("image6", images[3].FileName())
			}
		})

		t.Run("Query beyond images", func(t *testing.T) {
			images, err := sut.GetNextImagesInCategory(100, 100, apitype.NoCategory)
			a.Nil(err)
			a.Equal(0, len(images))
		})

		t.Run("Query negative offset", func(t *testing.T) {
			images, err := sut.GetNextImagesInCategory(2, -1, apitype.NoCategory)
			a.Nil(err)
			if a.Equal(2, len(images)) {
				a.Equal("image0", images[0].FileName())
				a.Equal("image1", images[1].FileName())
			}
		})
	})
}

func TestImageStore_GetNextImagesInCategory_CategorySet(t *testing.T) {
	a := assert.New(t)
	t.Run("Next images with category", func(t *testing.T) {
		sut := initImageStoreTest()
		image0, _ := sut.AddImage(apitype.NewImageFile("images", "image0"))
		image1, _ := sut.AddImage(apitype.NewImageFile("images", "image1"))
		image2, _ := sut.AddImage(apitype.NewImageFile("images", "image2"))
		image3, _ := sut.AddImage(apitype.NewImageFile("images", "image3"))
		image4, _ := sut.AddImage(apitype.NewImageFile("images", "image4"))
		image5, _ := sut.AddImage(apitype.NewImageFile("images", "image5"))
		image6, _ := sut.AddImage(apitype.NewImageFile("images", "image6"))
		image7, _ := sut.AddImage(apitype.NewImageFile("images", "image7"))
		image8, _ := sut.AddImage(apitype.NewImageFile("images", "image8"))
		_, _ = sut.AddImage(apitype.NewImageFile("images", "image9"))

		category1, _ := isCategoryStore.AddCategory(apitype.NewCategory("Cat 1", "C1", "C"))
		category2, _ := isCategoryStore.AddCategory(apitype.NewCategory("Cat 2", "C2", "D"))

		_ = isImageCategoryStore.CategorizeImage(image0.Id(), category1.Id(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image1.Id(), category1.Id(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image2.Id(), category1.Id(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image3.Id(), category1.Id(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image4.Id(), category1.Id(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image5.Id(), category1.Id(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image6.Id(), category1.Id(), apitype.MOVE)

		_ = isImageCategoryStore.CategorizeImage(image7.Id(), category2.Id(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image8.Id(), category2.Id(), apitype.MOVE)

		category := category1.Id()

		t.Run("Query less than max images", func(t *testing.T) {
			images, err := sut.GetNextImagesInCategory(5, 0, category)
			a.Nil(err)
			if a.Equal(5, len(images)) {
				a.Equal("image1", images[0].FileName())
				a.Equal("image2", images[1].FileName())
				a.Equal("image3", images[2].FileName())
				a.Equal("image4", images[3].FileName())
				a.Equal("image5", images[4].FileName())
			}
		})

		t.Run("Query more than max images", func(t *testing.T) {
			images, err := sut.GetNextImagesInCategory(10, 0, category)
			a.Nil(err)
			if a.Equal(6, len(images)) {
				a.Equal("image1", images[0].FileName())
				a.Equal("image2", images[1].FileName())
				a.Equal("image3", images[2].FileName())
				a.Equal("image4", images[3].FileName())
				a.Equal("image5", images[4].FileName())
				a.Equal("image6", images[5].FileName())
			}
		})

		t.Run("Query less than max images with offset", func(t *testing.T) {
			images, err := sut.GetNextImagesInCategory(4, 2, category)
			a.Nil(err)
			if a.Equal(4, len(images)) {
				a.Equal("image3", images[0].FileName())
				a.Equal("image4", images[1].FileName())
				a.Equal("image5", images[2].FileName())
				a.Equal("image6", images[3].FileName())
			}
		})

		t.Run("Query more than max images with offset", func(t *testing.T) {
			images, err := sut.GetNextImagesInCategory(10, 2, category)
			a.Nil(err)
			if a.Equal(4, len(images)) {
				a.Equal("image3", images[0].FileName())
				a.Equal("image4", images[1].FileName())
				a.Equal("image5", images[2].FileName())
				a.Equal("image6", images[3].FileName())
			}
		})

		t.Run("Query beyond images", func(t *testing.T) {
			images, err := sut.GetNextImagesInCategory(100, 100, category)
			a.Nil(err)
			a.Equal(0, len(images))
		})

		t.Run("Query negative offset", func(t *testing.T) {
			images, err := sut.GetNextImagesInCategory(2, -1, category)
			a.Nil(err)
			if a.Equal(2, len(images)) {
				a.Equal("image0", images[0].FileName())
				a.Equal("image1", images[1].FileName())
			}
		})
	})
}

func TestImageStore_GetPreviousImagesInCategory_NoCategorySet(t *testing.T) {
	a := assert.New(t)

	t.Run("No images", func(t *testing.T) {
		sut := initImageStoreTest()

		t.Run("Query ", func(t *testing.T) {
			images, err := sut.GetPreviousImagesInCategory(5, 0, apitype.NoCategory)
			a.Nil(err)
			a.Equal(0, len(images))
		})
	})

	t.Run("Previous images without category", func(t *testing.T) {
		sut := initImageStoreTest()
		err := sut.AddImages([]*apitype.ImageFile{
			apitype.NewImageFile("images", "image0"),
			apitype.NewImageFile("images", "image1"),
			apitype.NewImageFile("images", "image2"),
			apitype.NewImageFile("images", "image3"),
			apitype.NewImageFile("images", "image4"),
			apitype.NewImageFile("images", "image5"),
			apitype.NewImageFile("images", "image6"),
		})

		a.Nil(err)

		t.Run("Query less than max images", func(t *testing.T) {
			images, err := sut.GetPreviousImagesInCategory(5, 6, apitype.NoCategory)
			a.Nil(err)
			if a.Equal(5, len(images)) {
				a.Equal("image5", images[0].FileName())
				a.Equal("image4", images[1].FileName())
				a.Equal("image3", images[2].FileName())
				a.Equal("image2", images[3].FileName())
				a.Equal("image1", images[4].FileName())
			}
		})

		t.Run("Query at start", func(t *testing.T) {
			images, err := sut.GetPreviousImagesInCategory(5, 0, apitype.NoCategory)
			a.Nil(err)
			a.Equal(0, len(images))
		})

		t.Run("Query more than max images", func(t *testing.T) {
			images, err := sut.GetPreviousImagesInCategory(10, 6, apitype.NoCategory)
			a.Nil(err)
			if a.Equal(6, len(images)) {
				a.Equal("image5", images[0].FileName())
				a.Equal("image4", images[1].FileName())
				a.Equal("image3", images[2].FileName())
				a.Equal("image2", images[3].FileName())
				a.Equal("image1", images[4].FileName())
				a.Equal("image0", images[5].FileName())
			}
		})

		t.Run("Query less than max images with offset", func(t *testing.T) {
			images, err := sut.GetPreviousImagesInCategory(4, 5, apitype.NoCategory)
			a.Nil(err)
			if a.Equal(4, len(images)) {
				a.Equal("image4", images[0].FileName())
				a.Equal("image3", images[1].FileName())
				a.Equal("image2", images[2].FileName())
				a.Equal("image1", images[3].FileName())
			}
		})

		t.Run("Query more than max images with offset", func(t *testing.T) {
			images, err := sut.GetPreviousImagesInCategory(10, 5, apitype.NoCategory)
			a.Nil(err)
			if a.Equal(5, len(images)) {
				a.Equal("image4", images[0].FileName())
				a.Equal("image3", images[1].FileName())
				a.Equal("image2", images[2].FileName())
				a.Equal("image1", images[3].FileName())
				a.Equal("image0", images[4].FileName())
			}
		})

		t.Run("Query beyond images beyond offset", func(t *testing.T) {
			images, err := sut.GetPreviousImagesInCategory(100, 0, apitype.NoCategory)
			a.Nil(err)
			a.Equal(0, len(images))
		})

		t.Run("Query beyond images at start", func(t *testing.T) {
			images, err := sut.GetPreviousImagesInCategory(10, 100, apitype.NoCategory)
			a.Nil(err)
			a.Equal(0, len(images))
		})

		t.Run("Query negative offset", func(t *testing.T) {
			images, err := sut.GetPreviousImagesInCategory(2, -1, apitype.NoCategory)
			a.Nil(err)
			a.Equal(0, len(images))
		})
	})
}

func TestImageStore_GetPreviousImagesInCategory_CategorySet(t *testing.T) {
	a := assert.New(t)
	t.Run("Previous images with category", func(t *testing.T) {
		sut := initImageStoreTest()
		image0, _ := sut.AddImage(apitype.NewImageFile("images", "image0"))
		image1, _ := sut.AddImage(apitype.NewImageFile("images", "image1"))
		image2, _ := sut.AddImage(apitype.NewImageFile("images", "image2"))
		image3, _ := sut.AddImage(apitype.NewImageFile("images", "image3"))
		image4, _ := sut.AddImage(apitype.NewImageFile("images", "image4"))
		image5, _ := sut.AddImage(apitype.NewImageFile("images", "image5"))
		image6, _ := sut.AddImage(apitype.NewImageFile("images", "image6"))
		image7, _ := sut.AddImage(apitype.NewImageFile("images", "image7"))
		image8, _ := sut.AddImage(apitype.NewImageFile("images", "image8"))
		_, _ = sut.AddImage(apitype.NewImageFile("images", "image9"))

		category1, _ := isCategoryStore.AddCategory(apitype.NewCategory("Cat 1", "C1", "C"))
		category2, _ := isCategoryStore.AddCategory(apitype.NewCategory("Cat 2", "C2", "D"))

		_ = isImageCategoryStore.CategorizeImage(image0.Id(), category1.Id(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image1.Id(), category1.Id(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image2.Id(), category1.Id(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image3.Id(), category1.Id(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image4.Id(), category1.Id(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image5.Id(), category1.Id(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image6.Id(), category1.Id(), apitype.MOVE)

		_ = isImageCategoryStore.CategorizeImage(image7.Id(), category2.Id(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image8.Id(), category2.Id(), apitype.MOVE)

		category := category1.Id()

		t.Run("Query less than max images", func(t *testing.T) {
			images, err := sut.GetPreviousImagesInCategory(5, 6, category)
			a.Nil(err)
			if a.Equal(5, len(images)) {
				a.Equal("image5", images[0].FileName())
				a.Equal("image4", images[1].FileName())
				a.Equal("image3", images[2].FileName())
				a.Equal("image2", images[3].FileName())
				a.Equal("image1", images[4].FileName())
			}
		})

		t.Run("Query at start", func(t *testing.T) {
			images, err := sut.GetPreviousImagesInCategory(5, 0, category)
			a.Nil(err)
			a.Equal(0, len(images))
		})

		t.Run("Query more than max images", func(t *testing.T) {
			images, err := sut.GetPreviousImagesInCategory(10, 6, category)
			a.Nil(err)
			if a.Equal(6, len(images)) {
				a.Equal("image5", images[0].FileName())
				a.Equal("image4", images[1].FileName())
				a.Equal("image3", images[2].FileName())
				a.Equal("image2", images[3].FileName())
				a.Equal("image1", images[4].FileName())
				a.Equal("image0", images[5].FileName())
			}
		})

		t.Run("Query less than max images with offset", func(t *testing.T) {
			images, err := sut.GetPreviousImagesInCategory(4, 5, category)
			a.Nil(err)
			if a.Equal(4, len(images)) {
				a.Equal("image4", images[0].FileName())
				a.Equal("image3", images[1].FileName())
				a.Equal("image2", images[2].FileName())
				a.Equal("image1", images[3].FileName())
			}
		})

		t.Run("Query more than max images with offset", func(t *testing.T) {
			images, err := sut.GetPreviousImagesInCategory(10, 5, category)
			a.Nil(err)
			if a.Equal(5, len(images)) {
				a.Equal("image4", images[0].FileName())
				a.Equal("image3", images[1].FileName())
				a.Equal("image2", images[2].FileName())
				a.Equal("image1", images[3].FileName())
				a.Equal("image0", images[4].FileName())
			}
		})

		t.Run("Query beyond images beyond offset", func(t *testing.T) {
			images, err := sut.GetPreviousImagesInCategory(100, 0, category)
			a.Nil(err)
			a.Equal(0, len(images))
		})

		t.Run("Query beyond images at start", func(t *testing.T) {
			images, err := sut.GetPreviousImagesInCategory(10, 100, category)
			a.Nil(err)
			a.Equal(0, len(images))
		})

		t.Run("Query negative offset", func(t *testing.T) {
			images, err := sut.GetPreviousImagesInCategory(2, -1, category)
			a.Nil(err)
			a.Equal(0, len(images))
		})
	})
}

func TestImageStore_GetImages_NoImages(t *testing.T) {
	a := require.New(t)

	sut := initImageStoreTest()

	t.Run("GetAllImages", func(t *testing.T) {
		images, err := sut.GetAllImages()
		a.Nil(err)
		a.Equal(0, len(images))
	})
	t.Run("GetImages", func(t *testing.T) {
		images, err := sut.GetImagesInCategory(10, 0, apitype.NoCategory)
		a.Nil(err)
		a.Equal(0, len(images))
	})
}

func TestImageStore_AddImages_GetImages(t *testing.T) {
	a := require.New(t)

	t.Run("No category", func(t *testing.T) {

		sut := initImageStoreTest()

		err := sut.AddImages([]*apitype.ImageFile{
			apitype.NewImageFile("images", "image1"),
			apitype.NewImageFile("images", "image2"),
			apitype.NewImageFile("images", "image3"),
			apitype.NewImageFile("images", "image4"),
			apitype.NewImageFile("images", "image5"),
			apitype.NewImageFile("images", "image6"),
		})

		a.Nil(err)

		t.Run("Get all images", func(t *testing.T) {
			images, err := sut.GetAllImages()

			a.Nil(err)

			a.Equal(6, len(images))

			a.NotNil(images[0].Id())
			a.Equal("image1", images[0].FileName())
			a.NotNil(images[1].Id())
			a.Equal("image2", images[1].FileName())
			a.NotNil(images[2].Id())
			a.Equal("image3", images[2].FileName())
			a.NotNil(images[3].Id())
			a.Equal("image4", images[3].FileName())
			a.NotNil(images[4].Id())
			a.Equal("image5", images[4].FileName())
			a.NotNil(images[5].Id())
			a.Equal("image6", images[5].FileName())
		})

		t.Run("Get first 2 images", func(t *testing.T) {
			images, err := sut.GetImagesInCategory(2, 0, apitype.NoCategory)

			a.Nil(err)

			a.Equal(2, len(images))

			a.NotNil(images[0].Id())
			a.Equal("image1", images[0].FileName())
			a.NotNil(images[1].Id())
			a.Equal("image2", images[1].FileName())
		})

		t.Run("Get next 2 images", func(t *testing.T) {
			images, err := sut.GetImagesInCategory(2, 2, apitype.NoCategory)

			a.Nil(err)

			a.Equal(2, len(images))

			a.NotNil(images[0].Id())
			a.Equal("image3", images[0].FileName())
			a.NotNil(images[1].Id())
			a.Equal("image4", images[1].FileName())
		})

		t.Run("Get last 10 images offset 3", func(t *testing.T) {
			images, err := sut.GetImagesInCategory(100, 3, apitype.NoCategory)

			a.Nil(err)

			a.Equal(3, len(images))

			a.NotNil(images[0].Id())
			a.Equal("image4", images[0].FileName())
			a.NotNil(images[1].Id())
			a.Equal("image5", images[1].FileName())
			a.NotNil(images[2].Id())
			a.Equal("image6", images[2].FileName())
		})

		t.Run("Get first 10 images", func(t *testing.T) {
			images, err := sut.GetImagesInCategory(100, 0, apitype.NoCategory)

			a.Nil(err)

			a.Equal(6, len(images))

			a.NotNil(images[0].Id())
			a.Equal("image1", images[0].FileName())
			a.NotNil(images[1].Id())
			a.Equal("image2", images[1].FileName())
			a.NotNil(images[2].Id())
			a.Equal("image3", images[2].FileName())
			a.NotNil(images[3].Id())
			a.Equal("image4", images[3].FileName())
			a.NotNil(images[4].Id())
			a.Equal("image5", images[4].FileName())
			a.NotNil(images[5].Id())
			a.Equal("image6", images[5].FileName())
		})
	})

	t.Run("In category", func(t *testing.T) {
		sut := initImageStoreTest()
		image1, _ := sut.AddImage(apitype.NewImageFile("images", "image1"))
		image2, _ := sut.AddImage(apitype.NewImageFile("images", "image2"))
		image3, _ := sut.AddImage(apitype.NewImageFile("images", "image3"))
		image4, _ := sut.AddImage(apitype.NewImageFile("images", "image4"))
		image5, _ := sut.AddImage(apitype.NewImageFile("images", "image5"))
		image6, _ := sut.AddImage(apitype.NewImageFile("images", "image6"))
		image7, _ := sut.AddImage(apitype.NewImageFile("images", "image7"))
		image8, _ := sut.AddImage(apitype.NewImageFile("images", "image8"))
		_, _ = sut.AddImage(apitype.NewImageFile("images", "image9"))

		category1, _ := isCategoryStore.AddCategory(apitype.NewCategory("Cat 1", "C1", "C"))
		category2, _ := isCategoryStore.AddCategory(apitype.NewCategory("Cat 2", "C2", "D"))

		_ = isImageCategoryStore.CategorizeImage(image1.Id(), category1.Id(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image2.Id(), category1.Id(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image3.Id(), category1.Id(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image4.Id(), category1.Id(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image5.Id(), category1.Id(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image6.Id(), category1.Id(), apitype.MOVE)

		_ = isImageCategoryStore.CategorizeImage(image7.Id(), category2.Id(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image8.Id(), category2.Id(), apitype.MOVE)

		category := category1.Id()

		t.Run("Get all images", func(t *testing.T) {
			images, err := sut.GetAllImages()

			a.Nil(err)

			a.Equal(9, len(images))

			a.NotNil(images[0].Id())
			a.Equal("image1", images[0].FileName())
			a.NotNil(images[1].Id())
			a.Equal("image2", images[1].FileName())
			a.NotNil(images[2].Id())
			a.Equal("image3", images[2].FileName())
			a.NotNil(images[3].Id())
			a.Equal("image4", images[3].FileName())
			a.NotNil(images[4].Id())
			a.Equal("image5", images[4].FileName())
			a.NotNil(images[5].Id())
			a.Equal("image6", images[5].FileName())
			a.NotNil(images[6].Id())
			a.Equal("image7", images[6].FileName())
			a.NotNil(images[7].Id())
			a.Equal("image8", images[7].FileName())
			a.NotNil(images[8].Id())
			a.Equal("image9", images[8].FileName())
		})

		t.Run("Get first 2 images", func(t *testing.T) {
			images, err := sut.GetImagesInCategory(2, 0, category)

			a.Nil(err)

			a.Equal(2, len(images))

			a.NotNil(images[0].Id())
			a.Equal("image1", images[0].FileName())
			a.NotNil(images[1].Id())
			a.Equal("image2", images[1].FileName())
		})

		t.Run("Get next 2 images", func(t *testing.T) {
			images, err := sut.GetImagesInCategory(2, 2, category)

			a.Nil(err)

			a.Equal(2, len(images))

			a.NotNil(images[0].Id())
			a.Equal("image3", images[0].FileName())
			a.NotNil(images[1].Id())
			a.Equal("image4", images[1].FileName())
		})

		t.Run("Get last 10 images offset 3", func(t *testing.T) {
			images, err := sut.GetImagesInCategory(100, 3, category)

			a.Nil(err)

			a.Equal(3, len(images))

			a.NotNil(images[0].Id())
			a.Equal("image4", images[0].FileName())
			a.NotNil(images[1].Id())
			a.Equal("image5", images[1].FileName())
			a.NotNil(images[2].Id())
			a.Equal("image6", images[2].FileName())
		})

		t.Run("Get first 10 images", func(t *testing.T) {
			images, err := sut.GetImagesInCategory(100, 0, category)

			a.Nil(err)

			a.Equal(6, len(images))

			a.NotNil(images[0].Id())
			a.Equal("image1", images[0].FileName())
			a.NotNil(images[1].Id())
			a.Equal("image2", images[1].FileName())
			a.NotNil(images[2].Id())
			a.Equal("image3", images[2].FileName())
			a.NotNil(images[3].Id())
			a.Equal("image4", images[3].FileName())
			a.NotNil(images[4].Id())
			a.Equal("image5", images[4].FileName())
			a.NotNil(images[5].Id())
			a.Equal("image6", images[5].FileName())
		})
	})
}

func TestImageStore_FindByDirAndFile(t *testing.T) {
	a := require.New(t)

	sut := initImageStoreTest()

	err := sut.AddImages([]*apitype.ImageFile{
		apitype.NewImageFile("images", "image1"),
		apitype.NewImageFile("images", "image2"),
		apitype.NewImageFile("images", "image3"),
		apitype.NewImageFile("images", "image4"),
		apitype.NewImageFile("images", "image5"),
		apitype.NewImageFile("images", "image6"),
	})

	a.Nil(err)

	t.Run("Image found", func(t *testing.T) {
		image, err := sut.FindByDirAndFile(apitype.NewImageFile("images", "image3"))

		a.Nil(err)

		a.NotNil(image)
		a.NotNil(image.Id())
		a.Equal("image3", image.FileName())
	})

	t.Run("Image not found", func(t *testing.T) {
		image, err := sut.FindByDirAndFile(apitype.NewImageFile("images", "foo"))

		a.Nil(err)

		a.Nil(image)
	})
}

func TestImageStore_Exists(t *testing.T) {
	a := require.New(t)

	sut := initImageStoreTest()

	err := sut.AddImages([]*apitype.ImageFile{
		apitype.NewImageFile("images", "image1"),
		apitype.NewImageFile("images", "image2"),
		apitype.NewImageFile("images", "image3"),
		apitype.NewImageFile("images", "image4"),
		apitype.NewImageFile("images", "image5"),
		apitype.NewImageFile("images", "image6"),
	})

	a.Nil(err)

	t.Run("Image found", func(t *testing.T) {
		imageExists, err := sut.Exists(apitype.NewImageFile("images", "image3"))

		a.Nil(err)

		a.True(imageExists)
	})

	t.Run("Image not found", func(t *testing.T) {
		imageExists, err := sut.Exists(apitype.NewImageFile("images", "foo"))

		a.Nil(err)

		a.False(imageExists)
	})
}

func TestImageStore_FindModifiedId(t *testing.T) {
	a := require.New(t)

	t.Run("Modified", func(t *testing.T) {
		sut := initImageStoreTest()

		imageStoreImageFileConverter.SetIncrementModTimeRequest(true)
		image1, err := sut.AddImage(apitype.NewImageFile("images", "image1"))

		a.Nil(err)

		id, err := sut.FindModifiedId(image1)
		a.Nil(err)

		a.NotEqual(apitype.NoImage, id)
	})

	t.Run("Not modified", func(t *testing.T) {
		sut := initImageStoreTest()

		image1, err := sut.AddImage(apitype.NewImageFile("images", "image1"))

		a.Nil(err)

		id, err := sut.FindModifiedId(image1)
		a.Nil(err)

		a.Equal(apitype.NoImage, id)
	})
}

func TestImageStore_RemoveImage(t *testing.T) {
	a := require.New(t)

	sut := initImageStoreTest()

	err := sut.AddImages([]*apitype.ImageFile{
		apitype.NewImageFile("images", "image1"),
		apitype.NewImageFile("images", "image2"),
		apitype.NewImageFile("images", "image3"),
		apitype.NewImageFile("images", "image4"),
		apitype.NewImageFile("images", "image5"),
		apitype.NewImageFile("images", "image6"),
	})

	a.Nil(err)

	t.Run("Get all images", func(t *testing.T) {
		images, err := sut.GetAllImages()
		a.Nil(err)

		err = sut.RemoveImage(images[2].Id())
		a.Nil(err)

		images, err = sut.GetAllImages()
		a.Nil(err)
		a.Equal(5, len(images))

		a.NotNil(images[0].Id())
		a.Equal("image1", images[0].FileName())
		a.NotNil(images[1].Id())
		a.Equal("image2", images[1].FileName())
		a.NotNil(images[2].Id())
		a.Equal("image4", images[2].FileName())
		a.NotNil(images[3].Id())
		a.Equal("image5", images[3].FileName())
		a.NotNil(images[4].Id())
		a.Equal("image6", images[4].FileName())
	})

}

func TestImageStore_GetLatestModifiedImage(t *testing.T) {
	a := require.New(t)

	sut := initImageStoreTest()
	imageStoreImageFileConverter.SetNamedStubs(true)
	imageStoreImageFileConverter.AddStubFile("image1", time.Date(2021, 4, 3, 1, 30, 10, 0, time.UTC))
	imageStoreImageFileConverter.AddStubFile("image2", time.Date(2021, 4, 3, 1, 30, 12, 0, time.UTC)) // Newest
	imageStoreImageFileConverter.AddStubFile("image3", time.Date(2021, 4, 3, 1, 30, 11, 0, time.UTC))
	imageStoreImageFileConverter.AddStubFile("image4", time.Date(2021, 4, 3, 1, 30, 9, 0, time.UTC))
	imageStoreImageFileConverter.AddStubFile("image5", time.Date(2021, 4, 3, 1, 30, 1, 0, time.UTC)) // Oldest
	imageStoreImageFileConverter.AddStubFile("image6", time.Date(2021, 4, 3, 1, 30, 10, 0, time.UTC))

	err := sut.AddImages([]*apitype.ImageFile{
		apitype.NewImageFile("images", "image1"),
		apitype.NewImageFile("images", "image2"),
		apitype.NewImageFile("images", "image3"),
		apitype.NewImageFile("images", "image4"),
		apitype.NewImageFile("images", "image5"),
		apitype.NewImageFile("images", "image6"),
	})

	a.Nil(err)

	t.Run("Get latest image modified timestamp", func(t *testing.T) {
		timestamp := sut.GetLatestModifiedImage()
		a.NotNil(timestamp)

		a.Equal(time.Date(2021, 4, 3, 1, 30, 12, 0, time.UTC), timestamp)
	})
}

func TestImageStore_GetAllImagesModifiedAfter(t *testing.T) {
	a := require.New(t)

	sut := initImageStoreTest()
	imageStoreImageFileConverter.SetNamedStubs(true)
	imageStoreImageFileConverter.AddStubFile("image1", time.Date(2021, 4, 3, 1, 30, 10, 0, time.UTC))
	imageStoreImageFileConverter.AddStubFile("image2", time.Date(2021, 4, 3, 1, 30, 12, 0, time.UTC)) // Newest
	imageStoreImageFileConverter.AddStubFile("image3", time.Date(2021, 4, 3, 1, 30, 11, 0, time.UTC))
	imageStoreImageFileConverter.AddStubFile("image4", time.Date(2021, 4, 3, 1, 30, 9, 0, time.UTC))
	imageStoreImageFileConverter.AddStubFile("image5", time.Date(2021, 4, 3, 1, 30, 1, 0, time.UTC)) // Oldest
	imageStoreImageFileConverter.AddStubFile("image6", time.Date(2021, 4, 3, 1, 30, 10, 0, time.UTC))

	err := sut.AddImages([]*apitype.ImageFile{
		apitype.NewImageFile("images", "image1"),
		apitype.NewImageFile("images", "image2"),
		apitype.NewImageFile("images", "image3"),
		apitype.NewImageFile("images", "image4"),
		apitype.NewImageFile("images", "image5"),
		apitype.NewImageFile("images", "image6"),
	})

	a.Nil(err)

	t.Run("Get images modified after: Epoc start", func(t *testing.T) {
		images, err := sut.GetAllImagesModifiedAfter(time.Unix(0, 0))
		a.Nil(err)
		a.NotNil(images)
		a.Equal(6, len(images))
		a.Equal("image1", images[0].FileName())
		a.Equal("image2", images[1].FileName())
		a.Equal("image3", images[2].FileName())
		a.Equal("image4", images[3].FileName())
		a.Equal("image5", images[4].FileName())
		a.Equal("image6", images[5].FileName())
	})

	t.Run("Get images modified after: Mid", func(t *testing.T) {
		images, err := sut.GetAllImagesModifiedAfter(time.Date(2021, 4, 3, 1, 30, 9, 0, time.UTC))
		a.Nil(err)
		a.NotNil(images)
		a.Equal(4, len(images))
		a.Equal("image1", images[0].FileName())
		a.Equal("image2", images[1].FileName())
		a.Equal("image3", images[2].FileName())
		a.Equal("image6", images[3].FileName())
	})

	t.Run("Get images modified after: Oldest", func(t *testing.T) {
		images, err := sut.GetAllImagesModifiedAfter(time.Date(2021, 4, 3, 1, 30, 1, 0, time.UTC))
		a.Nil(err)
		a.NotNil(images)
		a.Equal(5, len(images))
		a.Equal("image1", images[0].FileName())
		a.Equal("image2", images[1].FileName())
		a.Equal("image3", images[2].FileName())
		a.Equal("image4", images[3].FileName())
		a.Equal("image6", images[4].FileName())
	})

	t.Run("Get images modified after: Newest", func(t *testing.T) {
		images, err := sut.GetAllImagesModifiedAfter(time.Date(2021, 4, 3, 1, 30, 12, 0, time.UTC))
		a.Nil(err)
		a.NotNil(images)
		a.Equal(0, len(images))
	})

}
