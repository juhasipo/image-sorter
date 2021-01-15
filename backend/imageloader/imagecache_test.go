package imageloader

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/database"
)

type StubImageHandleConverter struct {
	database.ImageHandleConverter
}

func (s *StubImageHandleConverter) HandleToImage(handle *apitype.Handle) (*database.Image, map[string]string, error) {
	if jsonData, err := json.Marshal(handle.GetMetaData()); err != nil {
		return nil, nil, err
	} else {
		return &database.Image{
			Id:              0,
			Name:            handle.GetFile(),
			FileName:        handle.GetFile(),
			Directory:       handle.GetDir(),
			ByteSize:        1234,
			ExifOrientation: 1,
			ImageAngle:      90,
			ImageFlip:       true,
			CreatedTime:     time.Now(),
			Width:           1024,
			Height:          2048,
			ModifiedTime:    time.Now(),
			ExifData:        jsonData,
		}, handle.GetMetaData(), nil
	}
}

func (s *StubImageHandleConverter) GetHandleFileStats(handle *apitype.Handle) (os.FileInfo, error) {
	return &StubFileInfo{modTime: time.Now()}, nil
}

type StubFileInfo struct {
	os.FileInfo

	modTime time.Time
}

func (s *StubFileInfo) ModTime() time.Time {
	return s.modTime
}

func TestDefaultImageStore_Initialize(t *testing.T) {
	a := assert.New(t)

	db := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(db, &StubImageHandleConverter{})

	loader := NewImageLoader(imageStore)
	cache := NewImageCache(loader)

	a.Equal(uint64(0), cache.GetByteSize())

	handles := []*apitype.Handle{
		apitype.NewHandle(testAssetsDir, "horizontal.jpg"),
		apitype.NewHandle(testAssetsDir, "vertical.jpg"),
	}
	imageStore.AddImages(handles)
	storedImages, _ := imageStore.GetAllImages()
	cache.Initialize(storedImages)

	a.Equal(uint64(60000), cache.GetByteSize())
	a.InDelta(0.06, cache.GetSizeInMB(), 0.1)
}

func TestDefaultImageStore_Purge(t *testing.T) {
	a := assert.New(t)

	db := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(db, &StubImageHandleConverter{})

	loader := NewImageLoader(imageStore)
	cache := NewImageCache(loader)

	a.Equal(uint64(0), cache.GetByteSize())

	handles := []*apitype.Handle{
		apitype.NewHandle(testAssetsDir, "horizontal.jpg"),
		apitype.NewHandle(testAssetsDir, "vertical.jpg"),
	}
	_ = imageStore.AddImages(handles)
	storedImages, _ := imageStore.GetAllImages()
	handle1 := storedImages[0]
	handle2 := storedImages[1]

	cache.Initialize(storedImages)

	a.Equal(uint64(60000), cache.GetByteSize())

	_, _ = cache.GetFull(handle1.GetId())
	_, _ = cache.GetFull(handle2.GetId())
	size := apitype.SizeOf(100, 100)
	_, _ = cache.GetScaled(handle1.GetId(), size)
	_, _ = cache.GetScaled(handle2.GetId(), size)

	a.Equal(uint64(79967424), cache.GetByteSize())
	a.InDelta(76.3, cache.GetSizeInMB(), 0.1)

	cache.Purge()
	a.Equal(uint64(60000), cache.GetByteSize())
}

func TestDefaultImageStore_GetFull(t *testing.T) {
	a := assert.New(t)

	db := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(db, &StubImageHandleConverter{})

	loader := NewImageLoader(imageStore)
	cache := NewImageCache(loader)

	t.Run("Valid", func(t *testing.T) {
		handle, _ := imageStore.AddImage(apitype.NewHandle(testAssetsDir, "horizontal.jpg"))
		img, err := cache.GetFull(handle.GetId())

		a.Nil(err)
		a.NotNil(img)
	})
	t.Run("No exif", func(t *testing.T) {
		handle, _ := imageStore.AddImage(apitype.NewHandle(testAssetsDir, "no-exif.jpg"))
		img, err := cache.GetFull(handle.GetId())

		a.Nil(err)
		a.NotNil(img)
	})
	t.Run("Invalid", func(t *testing.T) {
		handle := apitype.NewHandle("", "")
		img, err := cache.GetFull(handle.GetId())

		a.NotNil(err)
		a.Nil(img)
	})
}

func TestDefaultImageStore_GetScaled(t *testing.T) {
	a := assert.New(t)

	db := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(db, &StubImageHandleConverter{})

	loader := NewImageLoader(imageStore)
	cache := NewImageCache(loader)

	t.Run("Valid", func(t *testing.T) {
		handle, _ := imageStore.AddImage(apitype.NewHandle(testAssetsDir, "horizontal.jpg"))
		size := apitype.SizeOf(400, 400)
		img, err := cache.GetScaled(handle.GetId(), size)

		a.Nil(err)
		a.NotNil(img)
	})
	t.Run("No exif", func(t *testing.T) {
		handle, _ := imageStore.AddImage(apitype.NewHandle(testAssetsDir, "no-exif.jpg"))
		size := apitype.SizeOf(400, 400)
		img, err := cache.GetScaled(handle.GetId(), size)

		a.Nil(err)
		a.NotNil(img)
	})
	t.Run("Invalid", func(t *testing.T) {
		handle := apitype.NewHandle("", "")
		size := apitype.SizeOf(400, 400)
		img, err := cache.GetScaled(handle.GetId(), size)

		a.NotNil(err)
		a.Nil(img)
	})
}

func TestDefaultImageStore_GetThumbnail(t *testing.T) {
	a := assert.New(t)

	db := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(db, &StubImageHandleConverter{})

	loader := NewImageLoader(imageStore)
	cache := NewImageCache(loader)

	t.Run("Valid", func(t *testing.T) {
		handle, _ := imageStore.AddImage(apitype.NewHandle(testAssetsDir, "horizontal.jpg"))
		img, err := cache.GetThumbnail(handle.GetId())

		a.Nil(err)
		a.NotNil(img)
	})
	t.Run("No exif", func(t *testing.T) {
		handle, _ := imageStore.AddImage(apitype.NewHandle(testAssetsDir, "no-exif.jpg"))
		img, err := cache.GetThumbnail(handle.GetId())

		a.Nil(err)
		a.NotNil(img)
	})
	t.Run("Invalid", func(t *testing.T) {
		handle := apitype.NewHandle("", "")
		img, err := cache.GetThumbnail(handle.GetId())

		a.NotNil(err)
		a.Nil(img)
	})
}
