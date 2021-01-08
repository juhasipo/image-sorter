package database

import (
	"github.com/stretchr/testify/require"
	"testing"
	"vincit.fi/image-sorter/api/apitype"
)

var (
	imageStoreImageHandleConverter *StubImageHandleConverter
)

func initImageStoreTest() *ImageStore {
	store := NewInMemoryStore()
	imageStoreImageHandleConverter = &StubImageHandleConverter{}
	return NewImageStore(store, imageStoreImageHandleConverter)
}

func TestImageStore_AddImage_GetImageById(t *testing.T) {
	a := require.New(t)

	t.Run("Add image and get it by ID", func(t *testing.T) {
		sut := initImageStoreTest()

		image1, err := sut.AddImage(apitype.NewHandle("images", "image1"))
		_, err = sut.AddImage(apitype.NewHandle("images", "image2"))

		a.Nil(err)

		image := sut.GetImageById(image1.GetId())

		a.Equal(image1.GetId(), image.GetId())
		a.Equal(image1.GetFile(), image.GetFile())
		a.Equal(image1.GetDir(), image.GetDir())
		a.Equal(image1.GetByteSize(), image.GetByteSize())
		a.Equal(image1.GetPath(), image.GetPath())
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

func TestImageStore_GetImagesGetImages_NoImages(t *testing.T) {
	a := require.New(t)

	sut := initImageStoreTest()

	t.Run("GetAllImages", func(t *testing.T) {
		images, err := sut.GetAllImages()
		a.Nil(err)
		a.Equal(0, len(images))
	})
	t.Run("GetImages", func(t *testing.T) {
		images, err := sut.GetImages(10, 0)
		a.Nil(err)
		a.Equal(0, len(images))
	})
}

func TestImageStore_AddImages_GetImages(t *testing.T) {
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
		images, err := sut.GetImages(2, 0)

		a.Nil(err)

		a.Equal(2, len(images))

		a.NotNil(images[0].GetId())
		a.Equal("image1", images[0].GetFile())
		a.NotNil(images[1].GetId())
		a.Equal("image2", images[1].GetFile())
	})

	t.Run("Get next 2 images", func(t *testing.T) {
		images, err := sut.GetImages(2, 2)

		a.Nil(err)

		a.Equal(2, len(images))

		a.NotNil(images[0].GetId())
		a.Equal("image3", images[0].GetFile())
		a.NotNil(images[1].GetId())
		a.Equal("image4", images[1].GetFile())
	})

	t.Run("Get last 10 images offset 3", func(t *testing.T) {
		images, err := sut.GetImages(100, 3)

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
		images, err := sut.GetImages(100, 0)

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

		a.NotEqual(apitype.HandleId(-1), id)
	})
	
	t.Run("Not modified", func(t *testing.T) {
		sut := initImageStoreTest()

		image1, err := sut.AddImage(apitype.NewHandle("images", "image1"))

		a.Nil(err)

		id, err := sut.FindModifiedId(image1)
		a.Nil(err)

		a.Equal(apitype.HandleId(-1), id)
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
