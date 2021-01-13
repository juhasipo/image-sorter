package database

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"vincit.fi/image-sorter/api/apitype"
)

var (
	imageStoreImageHandleConverter *StubImageHandleConverter
	isCategoryStore                *CategoryStore
	isImageCategoryStore           *ImageCategoryStore
)

func initImageStoreTest() *ImageStore {
	store := NewInMemoryStore()
	imageStoreImageHandleConverter = &StubImageHandleConverter{}
	isCategoryStore = NewCategoryStore(store)
	isImageCategoryStore = NewImageCategoryStore(store)
	return NewImageStore(store, imageStoreImageHandleConverter)
}

func TestImageStore_AddImage_GetImageById(t *testing.T) {
	a := require.New(t)

	t.Run("Add image and get it by ID", func(t *testing.T) {
		sut := initImageStoreTest()

		image1, err := sut.AddImage(apitype.NewHandleWithMetaData("images", "image1", map[string]string{"foo": "bar1"}))
		_, err = sut.AddImage(apitype.NewHandleWithMetaData("images", "image2", map[string]string{"foo": "bar2"}))

		a.Nil(err)

		image := sut.GetImageById(image1.GetId())

		a.Equal(image1.GetId(), image.GetId())
		a.Equal(image1.GetFile(), image.GetFile())
		a.Equal(image1.GetDir(), image.GetDir())
		a.Equal(image1.GetByteSize(), image.GetByteSize())
		a.Equal(image1.GetPath(), image.GetPath())
		a.Equal(1, len(image.GetMetaData()))
		a.Equal("bar1", image.GetMetaData()["foo"])
	})

	t.Run("Re-add same image not modified", func(t *testing.T) {
		sut := initImageStoreTest()

		image1, err := sut.AddImage(apitype.NewHandle("images", "image1"))
		a.Nil(err)
		_, err = sut.AddImage(apitype.NewHandle("images", "image1"))
		a.Nil(err)

		images, err := sut.GetAllImages()
		a.Equal(1, len(images))

		image := sut.GetImageById(image1.GetId())

		a.Equal(image1.GetId(), image.GetId())
		a.Equal(image1.GetFile(), image.GetFile())
		a.Equal(image1.GetDir(), image.GetDir())
		a.Equal(image1.GetByteSize(), image.GetByteSize())
		a.Equal(image1.GetPath(), image.GetPath())
	})

	t.Run("Re-add same image modified", func(t *testing.T) {
		sut := initImageStoreTest()

		imageStoreImageHandleConverter.SetIncrementModTimeRequest(true)
		image1, err := sut.AddImage(apitype.NewHandle("images", "image1"))
		a.Nil(err)
		_, err = sut.AddImage(apitype.NewHandle("images", "image1"))
		a.Nil(err)

		images, err := sut.GetAllImages()
		a.Equal(1, len(images))

		image := sut.GetImageById(image1.GetId())

		a.Equal(image1.GetId(), image.GetId())
		a.Equal(image1.GetFile(), image.GetFile())
		a.Equal(image1.GetDir(), image.GetDir())
		a.Equal(image1.GetByteSize(), image.GetByteSize())
		a.Equal(image1.GetPath(), image.GetPath())
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
		err := sut.AddImages([]*apitype.Handle{
			apitype.NewHandle("images", "image0"),
			apitype.NewHandle("images", "image1"),
			apitype.NewHandle("images", "image2"),
			apitype.NewHandle("images", "image3"),
			apitype.NewHandle("images", "image4"),
			apitype.NewHandle("images", "image5"),
			apitype.NewHandle("images", "image6"),
		})

		a.Nil(err)

		t.Run("Query less than max images", func(t *testing.T) {
			images, err := sut.GetNextImagesInCategory(5, 0, apitype.NoCategory)
			a.Nil(err)
			if a.Equal(5, len(images)) {
				a.Equal("image1", images[0].GetFile())
				a.Equal("image2", images[1].GetFile())
				a.Equal("image3", images[2].GetFile())
				a.Equal("image4", images[3].GetFile())
				a.Equal("image5", images[4].GetFile())
			}
		})

		t.Run("Query more than max images", func(t *testing.T) {
			images, err := sut.GetNextImagesInCategory(10, 0, apitype.NoCategory)
			a.Nil(err)
			if a.Equal(6, len(images)) {
				a.Equal("image1", images[0].GetFile())
				a.Equal("image2", images[1].GetFile())
				a.Equal("image3", images[2].GetFile())
				a.Equal("image4", images[3].GetFile())
				a.Equal("image5", images[4].GetFile())
				a.Equal("image6", images[5].GetFile())
			}
		})

		t.Run("Query less than max images with offset", func(t *testing.T) {
			images, err := sut.GetNextImagesInCategory(4, 2, apitype.NoCategory)
			a.Nil(err)
			if a.Equal(4, len(images)) {
				a.Equal("image3", images[0].GetFile())
				a.Equal("image4", images[1].GetFile())
				a.Equal("image5", images[2].GetFile())
				a.Equal("image6", images[3].GetFile())
			}
		})

		t.Run("Query more than max images with offset", func(t *testing.T) {
			images, err := sut.GetNextImagesInCategory(10, 2, apitype.NoCategory)
			a.Nil(err)
			if a.Equal(4, len(images)) {
				a.Equal("image3", images[0].GetFile())
				a.Equal("image4", images[1].GetFile())
				a.Equal("image5", images[2].GetFile())
				a.Equal("image6", images[3].GetFile())
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
				a.Equal("image0", images[0].GetFile())
				a.Equal("image1", images[1].GetFile())
			}
		})
	})
}

func TestImageStore_GetNextImagesInCategory_CategorySet(t *testing.T) {
	a := assert.New(t)
	t.Run("Next images with category", func(t *testing.T) {
		sut := initImageStoreTest()
		image0, _ := sut.AddImage(apitype.NewHandle("images", "image0"))
		image1, _ := sut.AddImage(apitype.NewHandle("images", "image1"))
		image2, _ := sut.AddImage(apitype.NewHandle("images", "image2"))
		image3, _ := sut.AddImage(apitype.NewHandle("images", "image3"))
		image4, _ := sut.AddImage(apitype.NewHandle("images", "image4"))
		image5, _ := sut.AddImage(apitype.NewHandle("images", "image5"))
		image6, _ := sut.AddImage(apitype.NewHandle("images", "image6"))
		image7, _ := sut.AddImage(apitype.NewHandle("images", "image7"))
		image8, _ := sut.AddImage(apitype.NewHandle("images", "image8"))
		_, _ = sut.AddImage(apitype.NewHandle("images", "image9"))

		category1, _ := isCategoryStore.AddCategory(apitype.NewCategory("Cat 1", "C1", "C"))
		category2, _ := isCategoryStore.AddCategory(apitype.NewCategory("Cat 2", "C2", "D"))

		_ = isImageCategoryStore.CategorizeImage(image0.GetId(), category1.GetId(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image1.GetId(), category1.GetId(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image2.GetId(), category1.GetId(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image3.GetId(), category1.GetId(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image4.GetId(), category1.GetId(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image5.GetId(), category1.GetId(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image6.GetId(), category1.GetId(), apitype.MOVE)

		_ = isImageCategoryStore.CategorizeImage(image7.GetId(), category2.GetId(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image8.GetId(), category2.GetId(), apitype.MOVE)

		category := category1.GetId()

		t.Run("Query less than max images", func(t *testing.T) {
			images, err := sut.GetNextImagesInCategory(5, 0, category)
			a.Nil(err)
			if a.Equal(5, len(images)) {
				a.Equal("image1", images[0].GetFile())
				a.Equal("image2", images[1].GetFile())
				a.Equal("image3", images[2].GetFile())
				a.Equal("image4", images[3].GetFile())
				a.Equal("image5", images[4].GetFile())
			}
		})

		t.Run("Query more than max images", func(t *testing.T) {
			images, err := sut.GetNextImagesInCategory(10, 0, category)
			a.Nil(err)
			if a.Equal(6, len(images)) {
				a.Equal("image1", images[0].GetFile())
				a.Equal("image2", images[1].GetFile())
				a.Equal("image3", images[2].GetFile())
				a.Equal("image4", images[3].GetFile())
				a.Equal("image5", images[4].GetFile())
				a.Equal("image6", images[5].GetFile())
			}
		})

		t.Run("Query less than max images with offset", func(t *testing.T) {
			images, err := sut.GetNextImagesInCategory(4, 2, category)
			a.Nil(err)
			if a.Equal(4, len(images)) {
				a.Equal("image3", images[0].GetFile())
				a.Equal("image4", images[1].GetFile())
				a.Equal("image5", images[2].GetFile())
				a.Equal("image6", images[3].GetFile())
			}
		})

		t.Run("Query more than max images with offset", func(t *testing.T) {
			images, err := sut.GetNextImagesInCategory(10, 2, category)
			a.Nil(err)
			if a.Equal(4, len(images)) {
				a.Equal("image3", images[0].GetFile())
				a.Equal("image4", images[1].GetFile())
				a.Equal("image5", images[2].GetFile())
				a.Equal("image6", images[3].GetFile())
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
				a.Equal("image0", images[0].GetFile())
				a.Equal("image1", images[1].GetFile())
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
		err := sut.AddImages([]*apitype.Handle{
			apitype.NewHandle("images", "image0"),
			apitype.NewHandle("images", "image1"),
			apitype.NewHandle("images", "image2"),
			apitype.NewHandle("images", "image3"),
			apitype.NewHandle("images", "image4"),
			apitype.NewHandle("images", "image5"),
			apitype.NewHandle("images", "image6"),
		})

		a.Nil(err)

		t.Run("Query less than max images", func(t *testing.T) {
			images, err := sut.GetPreviousImagesInCategory(5, 6, apitype.NoCategory)
			a.Nil(err)
			if a.Equal(5, len(images)) {
				a.Equal("image5", images[0].GetFile())
				a.Equal("image4", images[1].GetFile())
				a.Equal("image3", images[2].GetFile())
				a.Equal("image2", images[3].GetFile())
				a.Equal("image1", images[4].GetFile())
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
				a.Equal("image5", images[0].GetFile())
				a.Equal("image4", images[1].GetFile())
				a.Equal("image3", images[2].GetFile())
				a.Equal("image2", images[3].GetFile())
				a.Equal("image1", images[4].GetFile())
				a.Equal("image0", images[5].GetFile())
			}
		})

		t.Run("Query less than max images with offset", func(t *testing.T) {
			images, err := sut.GetPreviousImagesInCategory(4, 5, apitype.NoCategory)
			a.Nil(err)
			if a.Equal(4, len(images)) {
				a.Equal("image4", images[0].GetFile())
				a.Equal("image3", images[1].GetFile())
				a.Equal("image2", images[2].GetFile())
				a.Equal("image1", images[3].GetFile())
			}
		})

		t.Run("Query more than max images with offset", func(t *testing.T) {
			images, err := sut.GetPreviousImagesInCategory(10, 5, apitype.NoCategory)
			a.Nil(err)
			if a.Equal(5, len(images)) {
				a.Equal("image4", images[0].GetFile())
				a.Equal("image3", images[1].GetFile())
				a.Equal("image2", images[2].GetFile())
				a.Equal("image1", images[3].GetFile())
				a.Equal("image0", images[4].GetFile())
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
		image0, _ := sut.AddImage(apitype.NewHandle("images", "image0"))
		image1, _ := sut.AddImage(apitype.NewHandle("images", "image1"))
		image2, _ := sut.AddImage(apitype.NewHandle("images", "image2"))
		image3, _ := sut.AddImage(apitype.NewHandle("images", "image3"))
		image4, _ := sut.AddImage(apitype.NewHandle("images", "image4"))
		image5, _ := sut.AddImage(apitype.NewHandle("images", "image5"))
		image6, _ := sut.AddImage(apitype.NewHandle("images", "image6"))
		image7, _ := sut.AddImage(apitype.NewHandle("images", "image7"))
		image8, _ := sut.AddImage(apitype.NewHandle("images", "image8"))
		_, _ = sut.AddImage(apitype.NewHandle("images", "image9"))

		category1, _ := isCategoryStore.AddCategory(apitype.NewCategory("Cat 1", "C1", "C"))
		category2, _ := isCategoryStore.AddCategory(apitype.NewCategory("Cat 2", "C2", "D"))

		_ = isImageCategoryStore.CategorizeImage(image0.GetId(), category1.GetId(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image1.GetId(), category1.GetId(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image2.GetId(), category1.GetId(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image3.GetId(), category1.GetId(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image4.GetId(), category1.GetId(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image5.GetId(), category1.GetId(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image6.GetId(), category1.GetId(), apitype.MOVE)

		_ = isImageCategoryStore.CategorizeImage(image7.GetId(), category2.GetId(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image8.GetId(), category2.GetId(), apitype.MOVE)

		category := category1.GetId()

		t.Run("Query less than max images", func(t *testing.T) {
			images, err := sut.GetPreviousImagesInCategory(5, 6, category)
			a.Nil(err)
			if a.Equal(5, len(images)) {
				a.Equal("image5", images[0].GetFile())
				a.Equal("image4", images[1].GetFile())
				a.Equal("image3", images[2].GetFile())
				a.Equal("image2", images[3].GetFile())
				a.Equal("image1", images[4].GetFile())
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
				a.Equal("image5", images[0].GetFile())
				a.Equal("image4", images[1].GetFile())
				a.Equal("image3", images[2].GetFile())
				a.Equal("image2", images[3].GetFile())
				a.Equal("image1", images[4].GetFile())
				a.Equal("image0", images[5].GetFile())
			}
		})

		t.Run("Query less than max images with offset", func(t *testing.T) {
			images, err := sut.GetPreviousImagesInCategory(4, 5, category)
			a.Nil(err)
			if a.Equal(4, len(images)) {
				a.Equal("image4", images[0].GetFile())
				a.Equal("image3", images[1].GetFile())
				a.Equal("image2", images[2].GetFile())
				a.Equal("image1", images[3].GetFile())
			}
		})

		t.Run("Query more than max images with offset", func(t *testing.T) {
			images, err := sut.GetPreviousImagesInCategory(10, 5, category)
			a.Nil(err)
			if a.Equal(5, len(images)) {
				a.Equal("image4", images[0].GetFile())
				a.Equal("image3", images[1].GetFile())
				a.Equal("image2", images[2].GetFile())
				a.Equal("image1", images[3].GetFile())
				a.Equal("image0", images[4].GetFile())
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

		err := sut.AddImages([]*apitype.Handle{
			apitype.NewHandle("images", "image1"),
			apitype.NewHandle("images", "image2"),
			apitype.NewHandle("images", "image3"),
			apitype.NewHandle("images", "image4"),
			apitype.NewHandle("images", "image5"),
			apitype.NewHandle("images", "image6"),
		})

		a.Nil(err)

		t.Run("Get all images", func(t *testing.T) {
			images, err := sut.GetAllImages()

			a.Nil(err)

			a.Equal(6, len(images))

			a.NotNil(images[0].GetId())
			a.Equal("image1", images[0].GetFile())
			a.NotNil(images[1].GetId())
			a.Equal("image2", images[1].GetFile())
			a.NotNil(images[2].GetId())
			a.Equal("image3", images[2].GetFile())
			a.NotNil(images[3].GetId())
			a.Equal("image4", images[3].GetFile())
			a.NotNil(images[4].GetId())
			a.Equal("image5", images[4].GetFile())
			a.NotNil(images[5].GetId())
			a.Equal("image6", images[5].GetFile())
		})

		t.Run("Get first 2 images", func(t *testing.T) {
			images, err := sut.GetImagesInCategory(2, 0, apitype.NoCategory)

			a.Nil(err)

			a.Equal(2, len(images))

			a.NotNil(images[0].GetId())
			a.Equal("image1", images[0].GetFile())
			a.NotNil(images[1].GetId())
			a.Equal("image2", images[1].GetFile())
		})

		t.Run("Get next 2 images", func(t *testing.T) {
			images, err := sut.GetImagesInCategory(2, 2, apitype.NoCategory)

			a.Nil(err)

			a.Equal(2, len(images))

			a.NotNil(images[0].GetId())
			a.Equal("image3", images[0].GetFile())
			a.NotNil(images[1].GetId())
			a.Equal("image4", images[1].GetFile())
		})

		t.Run("Get last 10 images offset 3", func(t *testing.T) {
			images, err := sut.GetImagesInCategory(100, 3, apitype.NoCategory)

			a.Nil(err)

			a.Equal(3, len(images))

			a.NotNil(images[0].GetId())
			a.Equal("image4", images[0].GetFile())
			a.NotNil(images[1].GetId())
			a.Equal("image5", images[1].GetFile())
			a.NotNil(images[2].GetId())
			a.Equal("image6", images[2].GetFile())
		})

		t.Run("Get first 10 images", func(t *testing.T) {
			images, err := sut.GetImagesInCategory(100, 0, apitype.NoCategory)

			a.Nil(err)

			a.Equal(6, len(images))

			a.NotNil(images[0].GetId())
			a.Equal("image1", images[0].GetFile())
			a.NotNil(images[1].GetId())
			a.Equal("image2", images[1].GetFile())
			a.NotNil(images[2].GetId())
			a.Equal("image3", images[2].GetFile())
			a.NotNil(images[3].GetId())
			a.Equal("image4", images[3].GetFile())
			a.NotNil(images[4].GetId())
			a.Equal("image5", images[4].GetFile())
			a.NotNil(images[5].GetId())
			a.Equal("image6", images[5].GetFile())
		})
	})

	t.Run("In category", func(t *testing.T) {
		sut := initImageStoreTest()
		image1, _ := sut.AddImage(apitype.NewHandle("images", "image1"))
		image2, _ := sut.AddImage(apitype.NewHandle("images", "image2"))
		image3, _ := sut.AddImage(apitype.NewHandle("images", "image3"))
		image4, _ := sut.AddImage(apitype.NewHandle("images", "image4"))
		image5, _ := sut.AddImage(apitype.NewHandle("images", "image5"))
		image6, _ := sut.AddImage(apitype.NewHandle("images", "image6"))
		image7, _ := sut.AddImage(apitype.NewHandle("images", "image7"))
		image8, _ := sut.AddImage(apitype.NewHandle("images", "image8"))
		_, _ = sut.AddImage(apitype.NewHandle("images", "image9"))

		category1, _ := isCategoryStore.AddCategory(apitype.NewCategory("Cat 1", "C1", "C"))
		category2, _ := isCategoryStore.AddCategory(apitype.NewCategory("Cat 2", "C2", "D"))

		_ = isImageCategoryStore.CategorizeImage(image1.GetId(), category1.GetId(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image2.GetId(), category1.GetId(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image3.GetId(), category1.GetId(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image4.GetId(), category1.GetId(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image5.GetId(), category1.GetId(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image6.GetId(), category1.GetId(), apitype.MOVE)

		_ = isImageCategoryStore.CategorizeImage(image7.GetId(), category2.GetId(), apitype.MOVE)
		_ = isImageCategoryStore.CategorizeImage(image8.GetId(), category2.GetId(), apitype.MOVE)

		category := category1.GetId()

		t.Run("Get all images", func(t *testing.T) {
			images, err := sut.GetAllImages()

			a.Nil(err)

			a.Equal(9, len(images))

			a.NotNil(images[0].GetId())
			a.Equal("image1", images[0].GetFile())
			a.NotNil(images[1].GetId())
			a.Equal("image2", images[1].GetFile())
			a.NotNil(images[2].GetId())
			a.Equal("image3", images[2].GetFile())
			a.NotNil(images[3].GetId())
			a.Equal("image4", images[3].GetFile())
			a.NotNil(images[4].GetId())
			a.Equal("image5", images[4].GetFile())
			a.NotNil(images[5].GetId())
			a.Equal("image6", images[5].GetFile())
			a.NotNil(images[6].GetId())
			a.Equal("image7", images[6].GetFile())
			a.NotNil(images[7].GetId())
			a.Equal("image8", images[7].GetFile())
			a.NotNil(images[8].GetId())
			a.Equal("image9", images[8].GetFile())
		})

		t.Run("Get first 2 images", func(t *testing.T) {
			images, err := sut.GetImagesInCategory(2, 0, category)

			a.Nil(err)

			a.Equal(2, len(images))

			a.NotNil(images[0].GetId())
			a.Equal("image1", images[0].GetFile())
			a.NotNil(images[1].GetId())
			a.Equal("image2", images[1].GetFile())
		})

		t.Run("Get next 2 images", func(t *testing.T) {
			images, err := sut.GetImagesInCategory(2, 2, category)

			a.Nil(err)

			a.Equal(2, len(images))

			a.NotNil(images[0].GetId())
			a.Equal("image3", images[0].GetFile())
			a.NotNil(images[1].GetId())
			a.Equal("image4", images[1].GetFile())
		})

		t.Run("Get last 10 images offset 3", func(t *testing.T) {
			images, err := sut.GetImagesInCategory(100, 3, category)

			a.Nil(err)

			a.Equal(3, len(images))

			a.NotNil(images[0].GetId())
			a.Equal("image4", images[0].GetFile())
			a.NotNil(images[1].GetId())
			a.Equal("image5", images[1].GetFile())
			a.NotNil(images[2].GetId())
			a.Equal("image6", images[2].GetFile())
		})

		t.Run("Get first 10 images", func(t *testing.T) {
			images, err := sut.GetImagesInCategory(100, 0, category)

			a.Nil(err)

			a.Equal(6, len(images))

			a.NotNil(images[0].GetId())
			a.Equal("image1", images[0].GetFile())
			a.NotNil(images[1].GetId())
			a.Equal("image2", images[1].GetFile())
			a.NotNil(images[2].GetId())
			a.Equal("image3", images[2].GetFile())
			a.NotNil(images[3].GetId())
			a.Equal("image4", images[3].GetFile())
			a.NotNil(images[4].GetId())
			a.Equal("image5", images[4].GetFile())
			a.NotNil(images[5].GetId())
			a.Equal("image6", images[5].GetFile())
		})
	})
}

func TestImageStore_FindByDirAndFile(t *testing.T) {
	a := require.New(t)

	sut := initImageStoreTest()

	err := sut.AddImages([]*apitype.Handle{
		apitype.NewHandle("images", "image1"),
		apitype.NewHandle("images", "image2"),
		apitype.NewHandle("images", "image3"),
		apitype.NewHandle("images", "image4"),
		apitype.NewHandle("images", "image5"),
		apitype.NewHandle("images", "image6"),
	})

	a.Nil(err)

	t.Run("Image found", func(t *testing.T) {
		image, err := sut.FindByDirAndFile(apitype.NewHandle("images", "image3"))

		a.Nil(err)

		a.NotNil(image)
		a.NotNil(image.GetId())
		a.Equal("image3", image.GetFile())
	})

	t.Run("Image not found", func(t *testing.T) {
		image, err := sut.FindByDirAndFile(apitype.NewHandle("images", "foo"))

		a.Nil(err)

		a.Nil(image)
	})
}

func TestImageStore_Exists(t *testing.T) {
	a := require.New(t)

	sut := initImageStoreTest()

	err := sut.AddImages([]*apitype.Handle{
		apitype.NewHandle("images", "image1"),
		apitype.NewHandle("images", "image2"),
		apitype.NewHandle("images", "image3"),
		apitype.NewHandle("images", "image4"),
		apitype.NewHandle("images", "image5"),
		apitype.NewHandle("images", "image6"),
	})

	a.Nil(err)

	t.Run("Image found", func(t *testing.T) {
		imageExists, err := sut.Exists(apitype.NewHandle("images", "image3"))

		a.Nil(err)

		a.True(imageExists)
	})

	t.Run("Image not found", func(t *testing.T) {
		imageExists, err := sut.Exists(apitype.NewHandle("images", "foo"))

		a.Nil(err)

		a.False(imageExists)
	})
}

func TestImageStore_FindModifiedId(t *testing.T) {
	a := require.New(t)

	t.Run("Modified", func(t *testing.T) {
		sut := initImageStoreTest()

		imageStoreImageHandleConverter.SetIncrementModTimeRequest(true)
		image1, err := sut.AddImage(apitype.NewHandle("images", "image1"))

		a.Nil(err)

		id, err := sut.FindModifiedId(image1)
		a.Nil(err)

		a.NotEqual(apitype.NoHandle, id)
	})

	t.Run("Not modified", func(t *testing.T) {
		sut := initImageStoreTest()

		image1, err := sut.AddImage(apitype.NewHandle("images", "image1"))

		a.Nil(err)

		id, err := sut.FindModifiedId(image1)
		a.Nil(err)

		a.Equal(apitype.NoHandle, id)
	})
}

func TestImageStore_RemoveImage(t *testing.T) {
	a := require.New(t)

	sut := initImageStoreTest()

	err := sut.AddImages([]*apitype.Handle{
		apitype.NewHandle("images", "image1"),
		apitype.NewHandle("images", "image2"),
		apitype.NewHandle("images", "image3"),
		apitype.NewHandle("images", "image4"),
		apitype.NewHandle("images", "image5"),
		apitype.NewHandle("images", "image6"),
	})

	a.Nil(err)

	t.Run("Get all images", func(t *testing.T) {
		images, err := sut.GetAllImages()
		a.Nil(err)

		err = sut.RemoveImage(images[2].GetId())
		a.Nil(err)

		images, err = sut.GetAllImages()
		a.Nil(err)
		a.Equal(5, len(images))

		a.NotNil(images[0].GetId())
		a.Equal("image1", images[0].GetFile())
		a.NotNil(images[1].GetId())
		a.Equal("image2", images[1].GetFile())
		a.NotNil(images[2].GetId())
		a.Equal("image4", images[2].GetFile())
		a.NotNil(images[3].GetId())
		a.Equal("image5", images[3].GetFile())
		a.NotNil(images[4].GetId())
		a.Equal("image6", images[4].GetFile())
	})

}
