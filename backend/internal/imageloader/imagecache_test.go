package imageloader

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/internal/database"
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

type StubProgressReporter struct {
	Current int
	Total   int
	api.ProgressReporter
}

func (s *StubProgressReporter) Update(name string, current int, total int, canCancel bool, modal bool) {
	s.Current = current
	s.Total = total
}

func (s *StubProgressReporter) Error(error string, err error) {
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

	a.Equal(0, int(cache.GetByteSize()))

	imageFiles := []*apitype.ImageFile{
		apitype.NewImageFile(testAssetsDir, "horizontal.jpg"),
		apitype.NewImageFile(testAssetsDir, "vertical.jpg"),
	}
	err := imageStore.AddImages(imageFiles)

	if a.Nil(err) {
		storedImages, _ := imageStore.GetAllImages()
		reporter := &StubProgressReporter{}
		cache.Initialize(storedImages, reporter)

		waitForCacheToFill(reporter)

		a.Equal(60000, int(cache.GetByteSize()))
		a.InDelta(0.06, cache.GetSizeInMB(), 0.1)
	}
}

func waitForCacheToFill(reporter *StubProgressReporter) {
	deadline := time.Now().Add(5 * time.Minute)
	for {
		if reporter.Current == reporter.Total && reporter.Total > 0 {
			break
		}
		if time.Now().After(deadline) {
			break
		}
	}
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

	err := imageStore.AddImages(imageFiles)

	if err != nil {
		storedImages, _ := imageStore.GetAllImages()
		imageFile0 := storedImages[0]
		imageFile1 := storedImages[1]

		reporter := &StubProgressReporter{}
		cache.Initialize(storedImages, reporter)

		waitForCacheToFill(reporter)

		a.Equal(60000, int(cache.GetByteSize()))

		_, _ = cache.GetFull(imageFile0.Id())
		_, _ = cache.GetFull(imageFile1.Id())
		size := apitype.SizeOf(100, 100)
		_, _ = cache.GetScaled(imageFile0.Id(), size)
		_, _ = cache.GetScaled(imageFile1.Id(), size)

		a.Equal(79967424, int(cache.GetByteSize()))
		a.InDelta(76.3, cache.GetSizeInMB(), 0.1)

		cache.Purge()
		a.Equal(60000, int(cache.GetByteSize()))
	}
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
