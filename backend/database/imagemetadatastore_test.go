package database

import (
	"github.com/stretchr/testify/require"
	"testing"
	"vincit.fi/image-sorter/api/apitype"
)

var (
	imdsImageStore *ImageStore
)

func initImageMetaDataStoreTest() *ImageMetaDataStore {
	database := NewInMemoryDatabase()

	imdsImageStore = NewImageStore(database, &StubImageFileConverter{})

	return NewImageMetaDataStore(database)
}

func TestImageMetaDataStore_GetMetaDataByImageId(t *testing.T) {
	a := require.New(t)

	t.Run("Simple cases", func(t *testing.T) {
		sut := initImageMetaDataStoreTest()

		image1, _ := imdsImageStore.AddImage(apitype.NewImageFile("images", "image1"))

		t.Run("No meta data", func(t *testing.T) {
			metaData, err := sut.GetMetaDataByImageId(image1.Id())
			a.Nil(err)
			a.NotNil(metaData)
			a.Equal(0, len(metaData.MetaData()))
		})

		t.Run("Meta data exists", func(t *testing.T) {
			initMetaData := apitype.NewImageMetaData(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			})
			err := sut.AddMetaData(image1.Id(), initMetaData)
			a.Nil(err)

			metaData, err := sut.GetMetaDataByImageId(image1.Id())
			a.Nil(err)
			a.NotNil(metaData)

			a.Equal("value1", metaData.MetaData()["key1"])
			a.Equal("value2", metaData.MetaData()["key2"])
			a.Equal("value3", metaData.MetaData()["key3"])
		})

	})
}

func TestImageMetaDataStore_GetAllImagesWithoutMetaData(t *testing.T) {
	a := require.New(t)

	t.Run("Meta data exists for one", func(t *testing.T) {
		sut := initImageMetaDataStoreTest()

		image1, _ := imdsImageStore.AddImage(apitype.NewImageFile("images", "image1"))
		imdsImageStore.AddImage(apitype.NewImageFile("images", "image2"))
		imdsImageStore.AddImage(apitype.NewImageFile("images", "image3"))

		initMetaData := apitype.NewImageMetaData(map[string]string{
			"key1": "value1",
		})
		err := sut.AddMetaData(image1.Id(), initMetaData)
		a.Nil(err)

		images, err := sut.GetAllImagesWithoutMetaData()
		a.Nil(err)
		a.NotNil(images)

		a.Equal(2, len(images))
		a.Equal("image2", images[0].FileName())
		a.Equal("image3", images[1].FileName())
	})

	t.Run("Meta data exists for all", func(t *testing.T) {
		sut := initImageMetaDataStoreTest()

		image1, _ := imdsImageStore.AddImage(apitype.NewImageFile("images", "image1"))
		image2, _ := imdsImageStore.AddImage(apitype.NewImageFile("images", "image2"))
		image3, _ := imdsImageStore.AddImage(apitype.NewImageFile("images", "image3"))

		initMetaData1 := apitype.NewImageMetaData(map[string]string{
			"key1": "value1",
		})
		err := sut.AddMetaData(image1.Id(), initMetaData1)
		a.Nil(err)

		initMetaData2 := apitype.NewImageMetaData(map[string]string{
			"key1": "value1",
		})
		err = sut.AddMetaData(image2.Id(), initMetaData2)
		a.Nil(err)

		initMetaData3 := apitype.NewImageMetaData(map[string]string{
			"key1": "value1",
		})
		err = sut.AddMetaData(image3.Id(), initMetaData3)
		a.Nil(err)

		images, err := sut.GetAllImagesWithoutMetaData()
		a.Nil(err)
		a.NotNil(images)

		a.Equal(0, len(images))
	})

}

func TestImageMetaDataStore_AddMetaDataBatch(t *testing.T) {
	a := require.New(t)

	t.Run("Simple cases", func(t *testing.T) {
		sut := initImageMetaDataStoreTest()

		image1, _ := imdsImageStore.AddImage(apitype.NewImageFile("images", "image1"))
		image2, _ := imdsImageStore.AddImage(apitype.NewImageFile("images", "image2"))

		t.Run("No meta data", func(t *testing.T) {
			images := []*apitype.ImageFile{image1, image2}
			cb := func(imageFile *apitype.ImageFile) (*apitype.ExifData, error) {
				if imageFile.Id() == image1.Id() {
					return nil, nil
				} else if imageFile.Id() == image2.Id() {
					return apitype.NewExifDataFromMap(map[string]string{}), nil
				} else {
					return nil, nil
				}
			}
			err := sut.AddMetaDataForImages(images, cb)
			a.Nil(err)

			metaData1, err := sut.GetMetaDataByImageId(image1.Id())
			a.Nil(err)
			a.NotNil(metaData1)
			a.Equal(0, len(metaData1.MetaData()))

			metaData2, err := sut.GetMetaDataByImageId(image2.Id())
			a.Nil(err)
			a.NotNil(metaData2)
			a.Equal(0, len(metaData2.MetaData()))
		})

		t.Run("Add meta data", func(t *testing.T) {
			images := []*apitype.ImageFile{image1, image2}
			cb := func(imageFile *apitype.ImageFile) (*apitype.ExifData, error) {
				if imageFile.Id() == image1.Id() {
					return apitype.NewExifDataFromMap(map[string]string{
						"image1-key1": "value1",
						"image1-key2": "value2",
					}), nil
				} else if imageFile.Id() == image2.Id() {
					return apitype.NewExifDataFromMap(map[string]string{
						"image2-key1": "value1",
						"image2-key2": "value2",
					}), nil
				} else {
					return nil, nil
				}
			}
			err := sut.AddMetaDataForImages(images, cb)
			a.Nil(err)

			metaData1, err := sut.GetMetaDataByImageId(image1.Id())
			a.Nil(err)
			a.NotNil(metaData1)

			a.Equal("value1", metaData1.MetaData()["image1-key1"])
			a.Equal("value2", metaData1.MetaData()["image1-key2"])

			metaData2, err := sut.GetMetaDataByImageId(image2.Id())
			a.Nil(err)
			a.NotNil(metaData2)

			a.Equal("value1", metaData2.MetaData()["image2-key1"])
			a.Equal("value2", metaData2.MetaData()["image2-key2"])
		})

		t.Run("Re-add meta data", func(t *testing.T) {
			images := []*apitype.ImageFile{image1, image2}
			cb := func(imageFile *apitype.ImageFile) (*apitype.ExifData, error) {
				if imageFile.Id() == image1.Id() {
					return apitype.NewExifDataFromMap(map[string]string{
						"image1-key1": "value1",
						"image1-key2": "value2",
					}), nil
				} else if imageFile.Id() == image2.Id() {
					return apitype.NewExifDataFromMap(map[string]string{
						"image2-key1": "value1",
						"image2-key2": "value2",
					}), nil
				} else {
					return nil, nil
				}
			}
			err := sut.AddMetaDataForImages(images, cb)
			err = sut.AddMetaDataForImages(images, cb)
			a.Nil(err)

			metaData1, err := sut.GetMetaDataByImageId(image1.Id())
			a.Nil(err)
			a.NotNil(metaData1)

			a.Equal("value1", metaData1.MetaData()["image1-key1"])
			a.Equal("value2", metaData1.MetaData()["image1-key2"])

			metaData2, err := sut.GetMetaDataByImageId(image2.Id())
			a.Nil(err)
			a.NotNil(metaData2)

			a.Equal("value1", metaData2.MetaData()["image2-key1"])
			a.Equal("value2", metaData2.MetaData()["image2-key2"])
		})

	})
}
