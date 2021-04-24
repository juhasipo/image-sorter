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
