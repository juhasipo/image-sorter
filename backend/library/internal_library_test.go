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

type StubImageFileConverter struct {
	database.ImageFileConverter
}

func (s *StubImageFileConverter) ImageFileToDbImage(imageFile *apitype.ImageFile) (*database.Image, map[string]string, error) {
	metaData := map[string]string{}
	if jsonData, err := json.Marshal(metaData); err != nil {
		return nil, nil, err
	} else {
		return &database.Image{
			Id:              0,
			Name:            imageFile.GetFile(),
			FileName:        imageFile.GetFile(),
			Directory:       imageFile.GetDir(),
			ByteSize:        1234,
			ExifOrientation: 1,
			ImageAngle:      90,
			ImageFlip:       true,
			CreatedTime:     time.Now(),
			Width:           1024,
			Height:          2048,
			ModifiedTime:    time.Now(),
			ExifData:        jsonData,
		}, metaData, nil
	}
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

var (
	sut                *internalManager
	sender             *MockSender
	store              *MockImageStore
	loader             *MockImageLoader
	imageStore         *database.ImageStore
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

func initializeSut() *internalManager {
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore = database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	categoryStore = database.NewCategoryStore(memoryDatabase)
	imageCategoryStore = database.NewImageCategoryStore(memoryDatabase)

	return newLibrary(store, loader, nil, imageStore)
}

func TestGetCurrentImage_Navigate_Empty(t *testing.T) {
	a := assert.New(t)

	sut := initializeSut()

	img, index, _ := sut.getCurrentImage()
	a.NotNil(img)
	a.Equal(0, index)
	a.False(img.GetImageFile().IsValid())
}

func TestGetCurrentImage_Navigate_OneImage(t *testing.T) {
	a := assert.New(t)

	sut = initializeSut()

	imageFiles := []*apitype.ImageFile{
		apitype.NewImageFile("/tmp", "foo1"),
	}
	sut.AddImageFiles(imageFiles)

	t.Run("Initial image", func(t *testing.T) {
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(0, index)
		a.Equal("foo1", img.GetImageFile().GetFile())
	})

	t.Run("Next stays on same", func(t *testing.T) {
		sut.RequestNextImage()
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(0, index)
		a.Equal("foo1", img.GetImageFile().GetFile())
	})

	t.Run("Previous stays on same", func(t *testing.T) {
		sut.RequestPrevImage()
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(0, index)
		a.Equal("foo1", img.GetImageFile().GetFile())
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

	t.Run("Initial image", func(t *testing.T) {
		sut.RequestPrevImage()
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(0, index)
		a.Equal("foo1", img.GetImageFile().GetFile())

	})

	t.Run("Next image", func(t *testing.T) {
		sut.RequestNextImage()
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(1, index)
		a.Equal("foo2", img.GetImageFile().GetFile())

	})

	t.Run("Next again", func(t *testing.T) {
		sut.RequestNextImage()
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(2, index)
		a.Equal("foo3", img.GetImageFile().GetFile())
	})

	t.Run("Next again should stay", func(t *testing.T) {
		sut.RequestNextImage()
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(2, index)
		a.Equal("foo3", img.GetImageFile().GetFile())
	})

	t.Run("Previous", func(t *testing.T) {
		sut.RequestPrevImage()
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(1, index)
		a.Equal("foo2", img.GetImageFile().GetFile())
	})
}

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
	sut.AddImageFiles(imageFiles)

	t.Run("Jump to forward 5 images", func(t *testing.T) {
		sut.MoveToNextImageWithOffset(5)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(5, index)
		a.Equal("foo5", img.GetImageFile().GetFile())
	})

	t.Run("Jump beyond the last", func(t *testing.T) {
		sut.MoveToNextImageWithOffset(10)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(9, index)
		a.Equal("foo9", img.GetImageFile().GetFile())
	})

	t.Run("Jump back to 5 images", func(t *testing.T) {
		sut.MoveToPrevImageWithOffset(5)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(4, index)
		a.Equal("foo4", img.GetImageFile().GetFile())
	})

	t.Run("Jump beyond the first", func(t *testing.T) {
		sut.MoveToPrevImageWithOffset(10)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(0, index)
		a.Equal("foo0", img.GetImageFile().GetFile())
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
	sut.AddImageFiles(imageFiles)

	t.Run("Index first image", func(t *testing.T) {
		sut.MoveToImageAt(0)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(0, index)
		a.Equal("foo0", img.GetImageFile().GetFile())
	})

	t.Run("Index 5", func(t *testing.T) {
		sut.MoveToImageAt(5)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(5, index)
		a.Equal("foo5", img.GetImageFile().GetFile())
	})

	t.Run("Index last image", func(t *testing.T) {
		sut.MoveToImageAt(9)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(9, index)
		a.Equal("foo9", img.GetImageFile().GetFile())
	})

	t.Run("Index after the last gives the last image", func(t *testing.T) {
		sut.MoveToImageAt(10)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(9, index)
		a.Equal("foo9", img.GetImageFile().GetFile())
	})

	t.Run("Index last image with negative index", func(t *testing.T) {
		sut.MoveToImageAt(-1)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(9, index)
		a.Equal("foo9", img.GetImageFile().GetFile())
	})

	t.Run("Index second to last image with negative index", func(t *testing.T) {
		sut.MoveToImageAt(-2)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(8, index)
		a.Equal("foo8", img.GetImageFile().GetFile())
	})

	t.Run("Too big negative index returns the first", func(t *testing.T) {
		sut.MoveToImageAt(-100)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(0, index)
		a.Equal("foo0", img.GetImageFile().GetFile())
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
	})
	imageFiles, _ := imageStore.GetAllImages()

	t.Run("foo1", func(t *testing.T) {
		sut.MoveToImage(imageFiles[1].GetId())
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(1, index)
		a.Equal("foo1", img.GetImageFile().GetFile())
	})
	t.Run("foo3", func(t *testing.T) {
		sut.MoveToImage(imageFiles[3].GetId())
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(3, index)
		a.Equal("foo3", img.GetImageFile().GetFile())
	})

	t.Run("NoImage stays on the current image", func(t *testing.T) {
		sut.MoveToImage(apitype.NoImage)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(3, index)
		a.Equal("foo3", img.GetImageFile().GetFile())
	})
}

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
	sut.SetImageListSize(5)
	a.Equal(5, sut.getImageListSize())

	t.Run("Initial image count", func(t *testing.T) {
		imgList, _ := sut.getNextImages()
		a.NotNil(imgList)
		if a.Equal(5, len(imgList)) {
			a.Equal("foo1", imgList[0].GetImageFile().GetFile())
			a.Equal("foo2", imgList[1].GetImageFile().GetFile())
			a.Equal("foo3", imgList[2].GetImageFile().GetFile())
			a.Equal("foo4", imgList[3].GetImageFile().GetFile())
			a.Equal("foo5", imgList[4].GetImageFile().GetFile())
		}
	})

	t.Run("Next requested gives the next 5", func(t *testing.T) {
		sut.RequestNextImage()
		imgList, _ := sut.getNextImages()
		a.NotNil(imgList)
		if a.Equal(5, len(imgList)) {
			a.Equal("foo2", imgList[0].GetImageFile().GetFile())
			a.Equal("foo3", imgList[1].GetImageFile().GetFile())
			a.Equal("foo4", imgList[2].GetImageFile().GetFile())
			a.Equal("foo5", imgList[3].GetImageFile().GetFile())
			a.Equal("foo6", imgList[4].GetImageFile().GetFile())
		}
	})

	t.Run("If no more next images, dont return more", func(t *testing.T) {
		sut.MoveToImageAt(6)
		imgList, _ := sut.getNextImages()
		a.NotNil(imgList)
		if a.Equal(3, len(imgList)) {
			a.Equal("foo7", imgList[0].GetImageFile().GetFile())
			a.Equal("foo8", imgList[1].GetImageFile().GetFile())
			a.Equal("foo9", imgList[2].GetImageFile().GetFile())
		}
	})

	t.Run("Second to last", func(t *testing.T) {
		sut.MoveToImageAt(8)
		imgList, _ := sut.getNextImages()
		a.NotNil(imgList)
		if a.Equal(1, len(imgList)) {
			a.Equal("foo9", imgList[0].GetImageFile().GetFile())
		}
	})

	t.Run("The last", func(t *testing.T) {
		sut.MoveToImageAt(9)
		imgList, _ := sut.getNextImages()
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
	sut.SetImageListSize(5)
	a.Equal(5, sut.getImageListSize())

	t.Run("Initial image count", func(t *testing.T) {
		imgList, _ := sut.getPrevImages()
		a.NotNil(imgList)
		a.Equal(0, len(imgList))
	})

	t.Run("Next requested gives the first image", func(t *testing.T) {
		sut.RequestNextImage()
		imgList, _ := sut.getPrevImages()
		a.NotNil(imgList)
		if a.Equal(1, len(imgList)) {
			a.Equal("foo0", imgList[0].GetImageFile().GetFile())
		}
	})

	t.Run("Image at 5 gives the first 5 images", func(t *testing.T) {
		sut.MoveToImageAt(5)
		imgList, _ := sut.getPrevImages()
		a.NotNil(imgList)
		if a.Equal(5, len(imgList)) {
			a.Equal("foo4", imgList[0].GetImageFile().GetFile())
			a.Equal("foo3", imgList[1].GetImageFile().GetFile())
			a.Equal("foo2", imgList[2].GetImageFile().GetFile())
			a.Equal("foo1", imgList[3].GetImageFile().GetFile())
			a.Equal("foo0", imgList[4].GetImageFile().GetFile())
		}
	})

	t.Run("Second to last image ", func(t *testing.T) {
		sut.MoveToImageAt(8)
		imgList, _ := sut.getPrevImages()
		a.NotNil(imgList)
		if a.Equal(5, len(imgList)) {
			a.Equal("foo7", imgList[0].GetImageFile().GetFile())
			a.Equal("foo6", imgList[1].GetImageFile().GetFile())
			a.Equal("foo5", imgList[2].GetImageFile().GetFile())
			a.Equal("foo4", imgList[3].GetImageFile().GetFile())
			a.Equal("foo3", imgList[4].GetImageFile().GetFile())
		}
	})

	t.Run("The last", func(t *testing.T) {
		sut.MoveToImageAt(9)
		imgList, _ := sut.getPrevImages()
		a.NotNil(imgList)
		if a.Equal(5, len(imgList)) {
			a.Equal("foo8", imgList[0].GetImageFile().GetFile())
			a.Equal("foo7", imgList[1].GetImageFile().GetFile())
			a.Equal("foo6", imgList[2].GetImageFile().GetFile())
			a.Equal("foo5", imgList[3].GetImageFile().GetFile())
			a.Equal("foo4", imgList[4].GetImageFile().GetFile())
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

	a.Equal(10, sut.getTotalImages())
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
	sut.SetImageListSize(10)

	_ = imageCategoryStore.CategorizeImage(imageFiles[1].GetId(), category1.GetId(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(imageFiles[2].GetId(), category1.GetId(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(imageFiles[6].GetId(), category1.GetId(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(imageFiles[7].GetId(), category1.GetId(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(imageFiles[9].GetId(), category1.GetId(), apitype.MOVE)

	_ = imageCategoryStore.CategorizeImage(imageFiles[0].GetId(), category2.GetId(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(imageFiles[1].GetId(), category2.GetId(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(imageFiles[3].GetId(), category2.GetId(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(imageFiles[9].GetId(), category2.GetId(), apitype.MOVE)

	sut.ShowOnlyImages(category1.GetId())

	a.Equal(5, sut.getTotalImages())
	a.Equal(category1.GetId(), sut.getSelectedCategoryId())

	t.Run("Next and prev images", func(t *testing.T) {
		nextImages, _ := sut.getNextImages()
		prevImages, _ := sut.getPrevImages()
		a.NotNil(nextImages)
		if a.Equal(4, len(nextImages)) {
			a.Equal(imageFiles[2].GetId(), nextImages[0].GetImageFile().GetId())
			a.Equal("foo2", nextImages[0].GetImageFile().GetFile())
			a.Equal(imageFiles[6].GetId(), nextImages[1].GetImageFile().GetId())
			a.Equal("foo6", nextImages[1].GetImageFile().GetFile())
			a.Equal(imageFiles[7].GetId(), nextImages[2].GetImageFile().GetId())
			a.Equal("foo7", nextImages[2].GetImageFile().GetFile())
			a.Equal(imageFiles[9].GetId(), nextImages[3].GetImageFile().GetId())
			a.Equal("foo9", nextImages[3].GetImageFile().GetFile())
		}

		a.NotNil(prevImages)
		a.Equal(0, len(prevImages))
	})

	t.Run("Next and prev images at 2", func(t *testing.T) {
		sut.MoveToImageAt(2)
		nextImages, _ := sut.getNextImages()
		prevImages, _ := sut.getPrevImages()
		a.NotNil(nextImages)
		if a.Equal(2, len(nextImages)) {
			a.Equal(imageFiles[7].GetId(), nextImages[0].GetImageFile().GetId())
			a.Equal("foo7", nextImages[0].GetImageFile().GetFile())
			a.Equal(imageFiles[9].GetId(), nextImages[1].GetImageFile().GetId())
			a.Equal("foo9", nextImages[1].GetImageFile().GetFile())
		}

		a.NotNil(prevImages)
		if a.Equal(2, len(prevImages)) {
			a.Equal(imageFiles[1].GetId(), prevImages[1].GetImageFile().GetId())
			a.Equal("foo1", prevImages[1].GetImageFile().GetFile())
			a.Equal(imageFiles[2].GetId(), prevImages[0].GetImageFile().GetId())
			a.Equal("foo2", prevImages[0].GetImageFile().GetFile())
		}
	})
}

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
	})
	sut.SetImageListSize(10)

	imageFiles, _ := imageStore.GetAllImages()
	category1, _ := categoryStore.AddCategory(apitype.NewCategory("category1", "C1", "1"))
	_ = imageCategoryStore.CategorizeImage(imageFiles[1].GetId(), category1.GetId(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(imageFiles[2].GetId(), category1.GetId(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(imageFiles[6].GetId(), category1.GetId(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(imageFiles[7].GetId(), category1.GetId(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(imageFiles[9].GetId(), category1.GetId(), apitype.MOVE)

	sut.ShowOnlyImages(category1.GetId())

	a.Equal(5, sut.getTotalImages())
	a.Equal(category1.GetId(), sut.getSelectedCategoryId())

	sut.ShowAllImages()
	a.Equal(10, sut.getTotalImages())
	a.Equal(apitype.NoCategory, sut.getSelectedCategoryId())

	t.Run("Next and prev images", func(t *testing.T) {
		nextImages, _ := sut.getNextImages()
		prevImages, _ := sut.getPrevImages()
		a.NotNil(nextImages)
		if a.Equal(9, len(nextImages)) {
			a.Equal("foo1", nextImages[0].GetImageFile().GetFile())
			a.Equal("foo2", nextImages[1].GetImageFile().GetFile())
			a.Equal("foo3", nextImages[2].GetImageFile().GetFile())
			a.Equal("foo4", nextImages[3].GetImageFile().GetFile())
			a.Equal("foo5", nextImages[4].GetImageFile().GetFile())
			a.Equal("foo6", nextImages[5].GetImageFile().GetFile())
			a.Equal("foo7", nextImages[6].GetImageFile().GetFile())
			a.Equal("foo8", nextImages[7].GetImageFile().GetFile())
			a.Equal("foo9", nextImages[8].GetImageFile().GetFile())
		}

		a.NotNil(prevImages)
		a.Equal(0, len(prevImages))
	})
}
