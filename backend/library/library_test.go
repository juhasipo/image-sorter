package library

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"image"
	"os"
	"testing"
	"time"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/database"
)

type MockSender struct {
	api.Sender
	mock.Mock
}

func (s *MockSender) SendCommandToTopic(topic api.Topic, command apitype.Command) {
	// Noop
}

type MockImageStore struct {
	api.ImageStore
	mock.Mock
}

type MockImage struct {
	image.Image
	mock.Mock
}

func (s *MockImageStore) GetFull(apitype.ImageId) (image.Image, error) {
	return nil, nil
}

func (s *MockImageStore) GetThumbnail(apitype.ImageId) (image.Image, error) {
	return nil, nil
}

type MockImageLoader struct {
	api.ImageLoader
	mock.Mock
}

func (s *MockImageLoader) LoadExifData(*apitype.ImageFile) (*apitype.ExifData, error) {
	return apitype.NewInvalidExifData(), nil
}

type StubImageFileConverter struct {
	database.ImageFileConverter
}

func (s *StubImageFileConverter) ImageFileToDbImage(imageFile *apitype.ImageFile) (*database.Image, map[string]string, error) {
	metaData := map[string]string{}
	if _, err := json.Marshal(metaData); err != nil {
		return nil, nil, err
	} else {
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
}

type FakeFile struct {
	name     string
	contents string
	mode     os.FileMode
	offset   int

	os.FileInfo
}

func (f *FakeFile) ModTime() time.Time {
	return time.Time{}
}

func (s *StubImageFileConverter) GetImageFileStats(*apitype.ImageFile) (os.FileInfo, error) {
	return &FakeFile{
		name:     "fake",
		contents: "",
		mode:     0,
		offset:   0,
	}, nil
}

type StubProgressReporter struct {
	api.ProgressReporter
}

func (s StubProgressReporter) Update(name string, current int, total int) {
}

func (s StubProgressReporter) Error(error string, err error) {
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

var (
	sut                *ImageLibrary
	sender             *MockSender
	store              *MockImageStore
	loader             *MockImageLoader
	imageStore         *database.ImageStore
	imageMetaDataStore *database.ImageMetaDataStore
	categoryStore      *database.CategoryStore
	imageCategoryStore *database.ImageCategoryStore
)

func setup() {
	sender = new(MockSender)
	store = new(MockImageStore)
	loader = new(MockImageLoader)

	store.On("GetFull", mock.Anything).Return(new(MockImage))
	store.On("GetThumbnail", mock.Anything).Return(new(MockImage))
}

func initializeSut() *ImageLibrary {
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore = database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	imageMetaDataStore = database.NewImageMetaDataStore(memoryDatabase)
	categoryStore = database.NewCategoryStore(memoryDatabase)
	imageCategoryStore = database.NewImageCategoryStore(memoryDatabase)

	return NewImageLibrary(store, loader, nil, imageStore, imageMetaDataStore, StubProgressReporter{})
}

func TestGetCurrentImage_Navigate_Empty(t *testing.T) {
	a := assert.New(t)

	sut := initializeSut()

	img, metaData, index, _ := sut.GetImageAtIndex(0, apitype.NoCategory)
	a.NotNil(img)
	a.NotNil(metaData)
	a.Equal(0, index)
	a.False(img.ImageFile().IsValid())
}

func TestGetCurrentImage_Navigate_OneImage(t *testing.T) {
	a := assert.New(t)

	sut = initializeSut()

	imageFiles := []*apitype.ImageFile{
		apitype.NewImageFile("/tmp", "foo1"),
	}
	sut.AddImageFiles(imageFiles)

	t.Run("First image", func(t *testing.T) {
		img, metaData, index, _ := sut.GetImageAtIndex(0, apitype.NoCategory)
		a.NotNil(img)
		a.NotNil(metaData)
		a.Equal(0, index)
		a.Equal("foo1", img.ImageFile().FileName())
	})

	t.Run("Positive index", func(t *testing.T) {
		img, metaData, index, _ := sut.GetImageAtIndex(1, apitype.NoCategory)
		a.NotNil(img)
		a.NotNil(metaData)
		a.Equal(0, index)
		a.Equal("", img.ImageFile().FileName())
	})

	t.Run("Negative index", func(t *testing.T) {
		img, metaData, index, _ := sut.GetImageAtIndex(-1, apitype.NoCategory)
		a.NotNil(img)
		a.NotNil(metaData)
		a.Equal(0, index)
		a.Equal("", img.ImageFile().FileName())
	})
}

func TestGetCurrentImage_Navigate_ManyImages(t *testing.T) {
	a := assert.New(t)

	sut = initializeSut()

	imageFiles := []*apitype.ImageFile{
		apitype.NewImageFile("/tmp", "foo1"),
		apitype.NewImageFile("/tmp", "foo2"),
		apitype.NewImageFile("/tmp", "foo3"),
	}
	sut.AddImageFiles(imageFiles)

	t.Run("First image", func(t *testing.T) {
		img, metaData, index, _ := sut.GetImageAtIndex(0, apitype.NoCategory)
		a.NotNil(img)
		a.NotNil(metaData)
		a.Equal(0, index)
		a.Equal("foo1", img.ImageFile().FileName())

	})

	t.Run("Second image", func(t *testing.T) {
		img, metaData, index, _ := sut.GetImageAtIndex(1, apitype.NoCategory)
		a.NotNil(img)
		a.NotNil(metaData)
		a.Equal(1, index)
		a.Equal("foo2", img.ImageFile().FileName())

	})

	t.Run("Last image", func(t *testing.T) {
		img, metaData, index, _ := sut.GetImageAtIndex(2, apitype.NoCategory)
		a.NotNil(img)
		a.NotNil(metaData)
		a.Equal(2, index)
		a.Equal("foo3", img.ImageFile().FileName())
	})

	t.Run("Over-indexing", func(t *testing.T) {
		img, metaData, index, _ := sut.GetImageAtIndex(3, apitype.NoCategory)
		a.NotNil(img)
		a.NotNil(metaData)
		a.Equal(0, index)
		a.Equal("", img.ImageFile().FileName())
	})
}

/*
func TestGetCurrentImage_Navigate_Jump(t *testing.T) {
	a := assert.New(t)

	sut = initializeSut()

	imageFiles := []*apitype.ImageFile{
		apitype.NewImageFile("/tmp", "foo0"),
		apitype.NewImageFile("/tmp", "foo1"),
		apitype.NewImageFile("/tmp", "foo2"),
		apitype.NewImageFile("/tmp", "foo3"),
		apitype.NewImageFile("/tmp", "foo4"),
		apitype.NewImageFile("/tmp", "foo5"),
		apitype.NewImageFile("/tmp", "foo6"),
		apitype.NewImageFile("/tmp", "foo7"),
		apitype.NewImageFile("/tmp", "foo8"),
		apitype.NewImageFile("/tmp", "foo9"),
	}
	sut.AddImageFiles(imageFiles, sender)

	t.Run("Jump to forward 5 images", func(t *testing.T) {
		sut.MoveToNextImageWithOffset(5)
		img, metaData, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.NotNil(metaData)
		a.Equal(5, index)
		a.Equal("foo5", img.ImageFile().FileName())
	})

	t.Run("Jump beyond the last", func(t *testing.T) {
		sut.MoveToNextImageWithOffset(10)
		img, metaData, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.NotNil(metaData)
		a.Equal(9, index)
		a.Equal("foo9", img.ImageFile().FileName())
	})

	t.Run("Jump back to 5 images", func(t *testing.T) {
		sut.MoveToPreviousImageWithOffset(5)
		img, metaData, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.NotNil(metaData)
		a.Equal(4, index)
		a.Equal("foo4", img.ImageFile().FileName())
	})

	t.Run("Jump beyond the first", func(t *testing.T) {
		sut.MoveToPreviousImageWithOffset(10)
		img, metaData, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.NotNil(metaData)
		a.Equal(0, index)
		a.Equal("foo0", img.ImageFile().FileName())
	})
}

func TestGetCurrentImage_Navigate_AtIndex(t *testing.T) {
	a := assert.New(t)

	sut = initializeSut()

	imageFiles := []*apitype.ImageFile{
		apitype.NewImageFile("/tmp", "foo0"),
		apitype.NewImageFile("/tmp", "foo1"),
		apitype.NewImageFile("/tmp", "foo2"),
		apitype.NewImageFile("/tmp", "foo3"),
		apitype.NewImageFile("/tmp", "foo4"),
		apitype.NewImageFile("/tmp", "foo5"),
		apitype.NewImageFile("/tmp", "foo6"),
		apitype.NewImageFile("/tmp", "foo7"),
		apitype.NewImageFile("/tmp", "foo8"),
		apitype.NewImageFile("/tmp", "foo9"),
	}
	sut.AddImageFiles(imageFiles, sender)

	t.Run("Index first image", func(t *testing.T) {
		sut.MoveToImageAt(0)
		img, metaData, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.NotNil(metaData)
		a.Equal(0, index)
		a.Equal("foo0", img.ImageFile().FileName())
	})

	t.Run("Index 5", func(t *testing.T) {
		sut.MoveToImageAt(5)
		img, metaData, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.NotNil(metaData)
		a.Equal(5, index)
		a.Equal("foo5", img.ImageFile().FileName())
	})

	t.Run("Index last image", func(t *testing.T) {
		sut.MoveToImageAt(9)
		img, metaData, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.NotNil(metaData)
		a.Equal(9, index)
		a.Equal("foo9", img.ImageFile().FileName())
	})

	t.Run("Index after the last gives the last image", func(t *testing.T) {
		sut.MoveToImageAt(10)
		img, metaData, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.NotNil(metaData)
		a.Equal(9, index)
		a.Equal("foo9", img.ImageFile().FileName())
	})

	t.Run("Index last image with negative index", func(t *testing.T) {
		sut.MoveToImageAt(-1)
		img, metaData, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.NotNil(metaData)
		a.Equal(9, index)
		a.Equal("foo9", img.ImageFile().FileName())
	})

	t.Run("Index second to last image with negative index", func(t *testing.T) {
		sut.MoveToImageAt(-2)
		img, metaData, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.NotNil(metaData)
		a.Equal(8, index)
		a.Equal("foo8", img.ImageFile().FileName())
	})

	t.Run("Too big negative index returns the first", func(t *testing.T) {
		sut.MoveToImageAt(-100)
		img, metaData, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.NotNil(metaData)
		a.Equal(0, index)
		a.Equal("foo0", img.ImageFile().FileName())
	})
}

func TestGetCurrentImage_Navigate_ImageId(t *testing.T) {
	a := assert.New(t)

	sut = initializeSut()

	sut.AddImageFiles([]*apitype.ImageFile{
		apitype.NewImageFile("/tmp", "foo0"),
		apitype.NewImageFile("/tmp", "foo1"),
		apitype.NewImageFile("/tmp", "foo2"),
		apitype.NewImageFile("/tmp", "foo3"),
		apitype.NewImageFile("/tmp", "foo4"),
	}, sender)
	imageFiles, _ := imageStore.GetAllImages()

	t.Run("foo1", func(t *testing.T) {
		sut.MoveToImage(imageFiles[1].Id())
		img, metaData, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.NotNil(metaData)
		a.Equal(1, index)
		a.Equal("foo1", img.ImageFile().FileName())
	})
	t.Run("foo3", func(t *testing.T) {
		sut.MoveToImage(imageFiles[3].Id())
		img, metaData, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.NotNil(metaData)
		a.Equal(3, index)
		a.Equal("foo3", img.ImageFile().FileName())
	})

	t.Run("NoImage stays on the current image", func(t *testing.T) {
		sut.MoveToImage(apitype.NoImage)
		img, metaData, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.NotNil(metaData)
		a.Equal(3, index)
		a.Equal("foo3", img.ImageFile().FileName())
	})
}
*/
func TestGetNextImages(t *testing.T) {
	a := assert.New(t)

	sut = initializeSut()

	imageFiles := []*apitype.ImageFile{
		apitype.NewImageFile("/tmp", "foo0"),
		apitype.NewImageFile("/tmp", "foo1"),
		apitype.NewImageFile("/tmp", "foo2"),
		apitype.NewImageFile("/tmp", "foo3"),
		apitype.NewImageFile("/tmp", "foo4"),
		apitype.NewImageFile("/tmp", "foo5"),
		apitype.NewImageFile("/tmp", "foo6"),
		apitype.NewImageFile("/tmp", "foo7"),
		apitype.NewImageFile("/tmp", "foo8"),
		apitype.NewImageFile("/tmp", "foo9"),
	}
	sut.AddImageFiles(imageFiles)

	t.Run("Initial image count", func(t *testing.T) {
		imgList, _ := sut.GetNextImages(0, 5, apitype.NoCategory)
		a.NotNil(imgList)
		if a.Equal(5, len(imgList)) {
			a.Equal("foo1", imgList[0].ImageFile().FileName())
			a.Equal("foo2", imgList[1].ImageFile().FileName())
			a.Equal("foo3", imgList[2].ImageFile().FileName())
			a.Equal("foo4", imgList[3].ImageFile().FileName())
			a.Equal("foo5", imgList[4].ImageFile().FileName())
		}
	})

	t.Run("Next requested gives the next 5", func(t *testing.T) {
		imgList, _ := sut.GetNextImages(1, 5, apitype.NoCategory)
		a.NotNil(imgList)
		if a.Equal(5, len(imgList)) {
			a.Equal("foo2", imgList[0].ImageFile().FileName())
			a.Equal("foo3", imgList[1].ImageFile().FileName())
			a.Equal("foo4", imgList[2].ImageFile().FileName())
			a.Equal("foo5", imgList[3].ImageFile().FileName())
			a.Equal("foo6", imgList[4].ImageFile().FileName())
		}
	})

	t.Run("If no more next images, dont return more", func(t *testing.T) {
		imgList, _ := sut.GetNextImages(6, 5, apitype.NoCategory)
		a.NotNil(imgList)
		if a.Equal(3, len(imgList)) {
			a.Equal("foo7", imgList[0].ImageFile().FileName())
			a.Equal("foo8", imgList[1].ImageFile().FileName())
			a.Equal("foo9", imgList[2].ImageFile().FileName())
		}
	})

	t.Run("Second to last", func(t *testing.T) {
		imgList, _ := sut.GetNextImages(8, 5, apitype.NoCategory)
		a.NotNil(imgList)
		if a.Equal(1, len(imgList)) {
			a.Equal("foo9", imgList[0].ImageFile().FileName())
		}
	})

	t.Run("The last", func(t *testing.T) {
		imgList, _ := sut.GetNextImages(9, 5, apitype.NoCategory)
		a.NotNil(imgList)
		a.Equal(0, len(imgList))
	})

}

func TestGetPrevImages(t *testing.T) {
	a := assert.New(t)

	sut = initializeSut()

	imageFiles := []*apitype.ImageFile{
		apitype.NewImageFile("/tmp", "foo0"),
		apitype.NewImageFile("/tmp", "foo1"),
		apitype.NewImageFile("/tmp", "foo2"),
		apitype.NewImageFile("/tmp", "foo3"),
		apitype.NewImageFile("/tmp", "foo4"),
		apitype.NewImageFile("/tmp", "foo5"),
		apitype.NewImageFile("/tmp", "foo6"),
		apitype.NewImageFile("/tmp", "foo7"),
		apitype.NewImageFile("/tmp", "foo8"),
		apitype.NewImageFile("/tmp", "foo9"),
	}
	sut.AddImageFiles(imageFiles)

	t.Run("Initial image count", func(t *testing.T) {
		imgList, _ := sut.GetPreviousImages(0, 5, apitype.NoCategory)
		a.NotNil(imgList)
		a.Equal(0, len(imgList))
	})

	t.Run("Next requested gives the first image", func(t *testing.T) {
		imgList, _ := sut.GetPreviousImages(1, 5, apitype.NoCategory)
		a.NotNil(imgList)
		if a.Equal(1, len(imgList)) {
			a.Equal("foo0", imgList[0].ImageFile().FileName())
		}
	})

	t.Run("Image at 5 gives the first 5 images", func(t *testing.T) {
		imgList, _ := sut.GetPreviousImages(5, 5, apitype.NoCategory)
		a.NotNil(imgList)
		if a.Equal(5, len(imgList)) {
			a.Equal("foo4", imgList[0].ImageFile().FileName())
			a.Equal("foo3", imgList[1].ImageFile().FileName())
			a.Equal("foo2", imgList[2].ImageFile().FileName())
			a.Equal("foo1", imgList[3].ImageFile().FileName())
			a.Equal("foo0", imgList[4].ImageFile().FileName())
		}
	})

	t.Run("Second to last image ", func(t *testing.T) {
		imgList, _ := sut.GetPreviousImages(8, 5, apitype.NoCategory)
		a.NotNil(imgList)
		if a.Equal(5, len(imgList)) {
			a.Equal("foo7", imgList[0].ImageFile().FileName())
			a.Equal("foo6", imgList[1].ImageFile().FileName())
			a.Equal("foo5", imgList[2].ImageFile().FileName())
			a.Equal("foo4", imgList[3].ImageFile().FileName())
			a.Equal("foo3", imgList[4].ImageFile().FileName())
		}
	})

	t.Run("The last", func(t *testing.T) {
		imgList, _ := sut.GetPreviousImages(9, 5, apitype.NoCategory)
		a.NotNil(imgList)
		if a.Equal(5, len(imgList)) {
			a.Equal("foo8", imgList[0].ImageFile().FileName())
			a.Equal("foo7", imgList[1].ImageFile().FileName())
			a.Equal("foo6", imgList[2].ImageFile().FileName())
			a.Equal("foo5", imgList[3].ImageFile().FileName())
			a.Equal("foo4", imgList[4].ImageFile().FileName())
		}
	})
}

func TestGetTotalCount(t *testing.T) {
	a := assert.New(t)

	sut = initializeSut()

	imageFiles := []*apitype.ImageFile{
		apitype.NewImageFile("/tmp", "foo0"),
		apitype.NewImageFile("/tmp", "foo1"),
		apitype.NewImageFile("/tmp", "foo2"),
		apitype.NewImageFile("/tmp", "foo3"),
		apitype.NewImageFile("/tmp", "foo4"),
		apitype.NewImageFile("/tmp", "foo5"),
		apitype.NewImageFile("/tmp", "foo6"),
		apitype.NewImageFile("/tmp", "foo7"),
		apitype.NewImageFile("/tmp", "foo8"),
		apitype.NewImageFile("/tmp", "foo9"),
	}
	sut.AddImageFiles(imageFiles)

	a.Equal(10, sut.GetTotalImages(apitype.NoCategory))
}

// Show only images

func TestShowOnlyImages(t *testing.T) {
	a := assert.New(t)

	sut := initializeSut()

	sut.AddImageFiles([]*apitype.ImageFile{
		apitype.NewImageFile("/tmp", "foo0"),
		apitype.NewImageFile("/tmp", "foo1"),
		apitype.NewImageFile("/tmp", "foo2"),
		apitype.NewImageFile("/tmp", "foo3"),
		apitype.NewImageFile("/tmp", "foo4"),
		apitype.NewImageFile("/tmp", "foo5"),
		apitype.NewImageFile("/tmp", "foo6"),
		apitype.NewImageFile("/tmp", "foo7"),
		apitype.NewImageFile("/tmp", "foo8"),
		apitype.NewImageFile("/tmp", "foo9"),
	})
	imageFiles, _ := imageStore.GetAllImages()
	category1, _ := categoryStore.AddCategory(apitype.NewCategory("category1", "cat1", "C"))
	category2, _ := categoryStore.AddCategory(apitype.NewCategory("category2", "cat2", "D"))

	_ = imageCategoryStore.CategorizeImage(imageFiles[1].Id(), category1.Id(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(imageFiles[2].Id(), category1.Id(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(imageFiles[6].Id(), category1.Id(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(imageFiles[7].Id(), category1.Id(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(imageFiles[9].Id(), category1.Id(), apitype.MOVE)

	_ = imageCategoryStore.CategorizeImage(imageFiles[0].Id(), category2.Id(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(imageFiles[1].Id(), category2.Id(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(imageFiles[3].Id(), category2.Id(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(imageFiles[9].Id(), category2.Id(), apitype.MOVE)

	selectedCategoryId := category1.Id()
	a.Equal(5, sut.GetTotalImages(selectedCategoryId))

	t.Run("Next and prev images", func(t *testing.T) {
		nextImages, _ := sut.GetNextImages(0, 5, selectedCategoryId)
		prevImages, _ := sut.GetPreviousImages(0, 5, selectedCategoryId)
		a.NotNil(nextImages)
		if a.Equal(4, len(nextImages)) {
			a.Equal(imageFiles[2].Id(), nextImages[0].ImageFile().Id())
			a.Equal("foo2", nextImages[0].ImageFile().FileName())
			a.Equal(imageFiles[6].Id(), nextImages[1].ImageFile().Id())
			a.Equal("foo6", nextImages[1].ImageFile().FileName())
			a.Equal(imageFiles[7].Id(), nextImages[2].ImageFile().Id())
			a.Equal("foo7", nextImages[2].ImageFile().FileName())
			a.Equal(imageFiles[9].Id(), nextImages[3].ImageFile().Id())
			a.Equal("foo9", nextImages[3].ImageFile().FileName())
		}

		a.NotNil(prevImages)
		a.Equal(0, len(prevImages))
	})

	t.Run("Next and prev images at 2", func(t *testing.T) {
		nextImages, _ := sut.GetNextImages(2, 5, selectedCategoryId)
		prevImages, _ := sut.GetPreviousImages(2, 5, selectedCategoryId)
		a.NotNil(nextImages)
		if a.Equal(2, len(nextImages)) {
			a.Equal(imageFiles[7].Id(), nextImages[0].ImageFile().Id())
			a.Equal("foo7", nextImages[0].ImageFile().FileName())
			a.Equal(imageFiles[9].Id(), nextImages[1].ImageFile().Id())
			a.Equal("foo9", nextImages[1].ImageFile().FileName())
		}

		a.NotNil(prevImages)
		if a.Equal(2, len(prevImages)) {
			a.Equal(imageFiles[1].Id(), prevImages[1].ImageFile().Id())
			a.Equal("foo1", prevImages[1].ImageFile().FileName())
			a.Equal(imageFiles[2].Id(), prevImages[0].ImageFile().Id())
			a.Equal("foo2", prevImages[0].ImageFile().FileName())
		}
	})
}

/*
func TestShowOnlyImages_ShowAllAgain(t *testing.T) {
	a := assert.New(t)

	sut := initializeSut()

	sut.AddImageFiles([]*apitype.ImageFile{
		apitype.NewImageFile("/tmp", "foo0"),
		apitype.NewImageFile("/tmp", "foo1"),
		apitype.NewImageFile("/tmp", "foo2"),
		apitype.NewImageFile("/tmp", "foo3"),
		apitype.NewImageFile("/tmp", "foo4"),
		apitype.NewImageFile("/tmp", "foo5"),
		apitype.NewImageFile("/tmp", "foo6"),
		apitype.NewImageFile("/tmp", "foo7"),
		apitype.NewImageFile("/tmp", "foo8"),
		apitype.NewImageFile("/tmp", "foo9"),
	}, sender)

	imageFiles, _ := imageStore.GetAllImages()
	category1, _ := categoryStore.AddCategory(apitype.NewCategory("category1", "C1", "1"))
	_ = imageCategoryStore.CategorizeImage(imageFiles[1].Id(), category1.Id(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(imageFiles[2].Id(), category1.Id(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(imageFiles[6].Id(), category1.Id(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(imageFiles[7].Id(), category1.Id(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(imageFiles[9].Id(), category1.Id(), apitype.MOVE)

	selectedCategoryId := category1.Id()
	a.Equal(5, sut.GetTotalImages(selectedCategoryId))

	t.Run("Next and prev images", func(t *testing.T) {
		nextImages, _ := sut.GetNextImages(0, 10, selectedCategoryId)
		prevImages, _ := sut.GetPreviousImages(0, 10, selectedCategoryId)
		a.NotNil(nextImages)
		if a.Equal(9, len(nextImages)) {
			a.Equal("foo1", nextImages[0].ImageFile().FileName())
			a.Equal("foo2", nextImages[1].ImageFile().FileName())
			a.Equal("foo3", nextImages[2].ImageFile().FileName())
			a.Equal("foo4", nextImages[3].ImageFile().FileName())
			a.Equal("foo5", nextImages[4].ImageFile().FileName())
			a.Equal("foo6", nextImages[5].ImageFile().FileName())
			a.Equal("foo7", nextImages[6].ImageFile().FileName())
			a.Equal("foo8", nextImages[7].ImageFile().FileName())
			a.Equal("foo9", nextImages[8].ImageFile().FileName())
		}

		a.NotNil(prevImages)
		a.Equal(0, len(prevImages))
	})
}
*/
