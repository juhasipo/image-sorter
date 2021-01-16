package library

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/database"
	"vincit.fi/image-sorter/backend/imageloader"
)

const testAssetsDir = "../../testassets"

func TestHashCalculator_GenerateHashes(t *testing.T) {
	a := assert.New(t)

	memoryDatabase := database.NewInMemoryDatabase()
	similarityIndex := database.NewSimilarityIndex(memoryDatabase)
	imageStore := database.NewImageStore(memoryDatabase, &StubImageHandleConverter{})

	imageLoader := imageloader.NewImageLoader(imageStore)

	t.Run("No images in store", func(t *testing.T) {
		sut := NewHashCalculator(similarityIndex, imageLoader, 1)

		hashes, err := sut.GenerateHashes([]*apitype.ImageFileWithMetaData{}, func(current int, total int) {})

		if a.Nil(err) {
			a.Equal(0, len(hashes))
		}
	})

	t.Run("Images in store", func(t *testing.T) {
		sut := NewHashCalculator(similarityIndex, imageLoader, 1)
		i1, _ := imageStore.AddImage(apitype.NewHandle(testAssetsDir, "horizontal.jpg"))
		i2, _ := imageStore.AddImage(apitype.NewHandle(testAssetsDir, "no-exif.jpg"))
		i3, _ := imageStore.AddImage(apitype.NewHandle(testAssetsDir, "vertical.jpg"))

		hashes, err := sut.GenerateHashes([]*apitype.ImageFileWithMetaData{i1, i2, i3}, func(current int, total int) {})

		if a.Nil(err) {
			a.Equal(3, len(hashes))
		}
	})
}

/*
TODO: Create a test for stopping the hash calculation
      The problem is to make it reliable since the whole thing
      is asynchronous
func TestHashCalculator_StopHashes(t *testing.T) {
	a := assert.New(t)

	store := database.NewInMemoryStore()
	similarityIndex := database.NewSimilarityIndex(store)
	imageStore := database.NewImageStore(store, &StubImageHandleConverter{})

	imageLoader := imageloader.NewImageLoader()

	sut := NewHashCalculator(similarityIndex, imageLoader, 1)
	i1, _ := imageStore.AddImage(apitype.NewHandle(testAssetsDir, "horizontal.jpg"))
	i2, _ := imageStore.AddImage(apitype.NewHandle(testAssetsDir, "no-exif.jpg"))
	i3, _ := imageStore.AddImage(apitype.NewHandle(testAssetsDir, "vertical.jpg"))

	hashes, err := sut.GenerateHashes([]*apitype.ImageFile{i1, i2, i3}, func(current int, total int) {
		go sut.StopHashes()
	})

	if a.Nil(err) {
		a.Less(3, len(hashes))
	}
}
*/

func TestHashCalculator_BuildSimilarityIndex(t *testing.T) {
	a := assert.New(t)

	memoryDatabase := database.NewInMemoryDatabase()
	similarityIndex := database.NewSimilarityIndex(memoryDatabase)
	imageStore := database.NewImageStore(memoryDatabase, &StubImageHandleConverter{})

	imageLoader := imageloader.NewImageLoader(imageStore)

	t.Run("No images in store", func(t *testing.T) {
		sut := NewHashCalculator(similarityIndex, imageLoader, 1)

		hashes, err := sut.GenerateHashes([]*apitype.ImageFileWithMetaData{}, func(current int, total int) {})

		if a.Nil(err) {

			err := sut.BuildSimilarityIndex(hashes, func(current int, total int) {})

			if a.Nil(err) {
				size, err := similarityIndex.GetIndexSize()
				if a.Nil(err) {
					a.Equal(uint64(0), size)
				}
			}
		}
	})

	t.Run("Images in store", func(t *testing.T) {
		sut := NewHashCalculator(similarityIndex, imageLoader, 1)
		i1, _ := imageStore.AddImage(apitype.NewHandle(testAssetsDir, "horizontal.jpg"))
		i2, _ := imageStore.AddImage(apitype.NewHandle(testAssetsDir, "no-exif.jpg"))
		i3, _ := imageStore.AddImage(apitype.NewHandle(testAssetsDir, "vertical.jpg"))

		hashes, err := sut.GenerateHashes([]*apitype.ImageFileWithMetaData{i1, i2, i3}, func(current int, total int) {})

		if a.Nil(err) {
			err := sut.BuildSimilarityIndex(hashes, func(current int, total int) {})

			if a.Nil(err) {
				size, err := similarityIndex.GetIndexSize()
				if a.Nil(err) {
					a.Equal(uint64(6), size)
				}

				images := similarityIndex.GetSimilarImages(i1.GetImageId())
				a.Equal(2, len(images))

				images = similarityIndex.GetSimilarImages(i2.GetImageId())
				a.Equal(2, len(images))

				images = similarityIndex.GetSimilarImages(i3.GetImageId())
				a.Equal(2, len(images))
			}
		}
	})
}
