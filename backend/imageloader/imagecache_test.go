package imageloader

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/database"
)

type StubImageFileConverter struct {
	database.ImageFileConverter
}

func (s *StubImageFileConverter) ImageFileToDbImage(imageFile *apitype.ImageFile) (*database.Image, map[string]string, error) {
	metaData := map[string]string{}
	return &database.Image{
		Id:              0,
		Name:            imageFile.FileName(),
		FileName:        imageFile.FileName(),
		Directory:       imageFile.Directory(),
		ByteSize:        1234,
		ExifOrientation: 1,
		ImageAngle:      90,
		ImageFlip:       true,
		CreatedTime:     time.Now(),
		Width:           1024,
		Height:          2048,
		ModifiedTime:    time.Now(),
	}, metaData, nil
}

func (s *StubImageFileConverter) GetImageFileStats(imageFile *apitype.ImageFile) (os.FileInfo, error) {
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
	imageStore := database.NewImageStore(db, &StubImageFileConverter{})

	loader := NewImageLoader(imageStore)
	cache := NewImageCache(loader)

	a.Equal(uint64(0), cache.GetByteSize())

	imageFiles := []*apitype.ImageFile{
		apitype.NewImageFile(testAssetsDir, "horizontal.jpg"),
		apitype.NewImageFile(testAssetsDir, "vertical.jpg"),
	}
	imageStore.AddImages(imageFiles)
	storedImages, _ := imageStore.GetAllImages()
	cache.Initialize(storedImages)

	a.Equal(uint64(60000), cache.GetByteSize())
	a.InDelta(0.06, cache.GetSizeInMB(), 0.1)
}

func TestDefaultImageStore_Purge(t *testing.T) {
	a := assert.New(t)

	db := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(db, &StubImageFileConverter{})

	loader := NewImageLoader(imageStore)
	cache := NewImageCache(loader)

	a.Equal(uint64(0), cache.GetByteSize())

	imageFiles := []*apitype.ImageFile{
		apitype.NewImageFile(testAssetsDir, "horizontal.jpg"),
		apitype.NewImageFile(testAssetsDir, "vertical.jpg"),
	}
	_ = imageStore.AddImages(imageFiles)
	storedImages, _ := imageStore.GetAllImages()
	imageFile0 := storedImages[0]
	imageFile1 := storedImages[1]

	cache.Initialize(storedImages)

	a.Equal(uint64(60000), cache.GetByteSize())

	_, _ = cache.GetFull(imageFile0.Id())
	_, _ = cache.GetFull(imageFile1.Id())
	size := apitype.SizeOf(100, 100)
	_, _ = cache.GetScaled(imageFile0.Id(), size)
	_, _ = cache.GetScaled(imageFile1.Id(), size)

	a.Equal(uint64(79967424), cache.GetByteSize())
	a.InDelta(76.3, cache.GetSizeInMB(), 0.1)

	cache.Purge()
	a.Equal(uint64(60000), cache.GetByteSize())
}

func TestDefaultImageStore_GetFull(t *testing.T) {
	a := assert.New(t)

	db := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(db, &StubImageFileConverter{})

	loader := NewImageLoader(imageStore)
	cache := NewImageCache(loader)

	t.Run("Valid", func(t *testing.T) {
		imageFile, _ := imageStore.AddImage(apitype.NewImageFile(testAssetsDir, "horizontal.jpg"))
		img, err := cache.GetFull(imageFile.Id())

		a.Nil(err)
		a.NotNil(img)
	})
	t.Run("No exif", func(t *testing.T) {
		imageFile, _ := imageStore.AddImage(apitype.NewImageFile(testAssetsDir, "no-exif.jpg"))
		img, err := cache.GetFull(imageFile.Id())

		a.Nil(err)
		a.NotNil(img)
	})
	t.Run("Invalid", func(t *testing.T) {
		imageFile := apitype.NewImageFile("", "")
		img, err := cache.GetFull(imageFile.Id())

		a.NotNil(err)
		a.Nil(img)
	})
}

func TestDefaultImageStore_GetScaled(t *testing.T) {
	a := assert.New(t)

	db := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(db, &StubImageFileConverter{})

	loader := NewImageLoader(imageStore)
	cache := NewImageCache(loader)

	t.Run("Valid", func(t *testing.T) {
		imageFile, _ := imageStore.AddImage(apitype.NewImageFile(testAssetsDir, "horizontal.jpg"))
		size := apitype.SizeOf(400, 400)
		img, err := cache.GetScaled(imageFile.Id(), size)

		a.Nil(err)
		a.NotNil(img)
	})
	t.Run("No exif", func(t *testing.T) {
		imageFile, _ := imageStore.AddImage(apitype.NewImageFile(testAssetsDir, "no-exif.jpg"))
		size := apitype.SizeOf(400, 400)
		img, err := cache.GetScaled(imageFile.Id(), size)

		a.Nil(err)
		a.NotNil(img)
	})
	t.Run("Invalid", func(t *testing.T) {
		imageFile := apitype.NewImageFile("", "")
		size := apitype.SizeOf(400, 400)
		img, err := cache.GetScaled(imageFile.Id(), size)

		a.NotNil(err)
		a.Nil(img)
	})
}

func TestDefaultImageStore_GetThumbnail(t *testing.T) {
	a := assert.New(t)

	db := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(db, &StubImageFileConverter{})

	loader := NewImageLoader(imageStore)
	cache := NewImageCache(loader)

	t.Run("Valid", func(t *testing.T) {
		imageFile, _ := imageStore.AddImage(apitype.NewImageFile(testAssetsDir, "horizontal.jpg"))
		img, err := cache.GetThumbnail(imageFile.Id())

		a.Nil(err)
		a.NotNil(img)
	})
	t.Run("No exif", func(t *testing.T) {
		imageFile, _ := imageStore.AddImage(apitype.NewImageFile(testAssetsDir, "no-exif.jpg"))
		img, err := cache.GetThumbnail(imageFile.Id())

		a.Nil(err)
		a.NotNil(img)
	})
	t.Run("Invalid", func(t *testing.T) {
		imageFile := apitype.NewImageFile("", "")
		img, err := cache.GetThumbnail(imageFile.Id())

		a.NotNil(err)
		a.Nil(img)
	})
}
