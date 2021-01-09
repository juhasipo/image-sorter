package library

import (
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

func (s *MockImageStore) GetFull(handle *apitype.Handle) (image.Image, error) {
	return nil, nil
}

func (s *MockImageStore) GetThumbnail(handle *apitype.Handle) (image.Image, error) {
	return nil, nil
}

type MockImageLoader struct {
	api.ImageLoader
	mock.Mock
}

type StubImageHandleConverter struct {
	database.ImageHandleConverter
}

func (s *StubImageHandleConverter) HandleToImage(handle *apitype.Handle) (*database.Image, error) {
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
	}, nil
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
	dbStore := database.NewInMemoryStore()
	imageStore = database.NewImageStore(dbStore, &StubImageHandleConverter{})
	categoryStore = database.NewCategoryStore(dbStore)
	imageCategoryStore = database.NewImageCategoryStore(dbStore)

	return newLibrary(store, loader, nil, imageStore)
}

func TestGetCurrentImage_Navigate_Empty(t *testing.T) {
	a := assert.New(t)

	sut := initializeSut()

	img, index, _ := sut.getCurrentImage()
	a.NotNil(img)
	a.Equal(0, index)
	a.False(img.GetHandle().IsValid())
}

func TestGetCurrentImage_Navigate_OneImage(t *testing.T) {
	a := assert.New(t)

	sut = initializeSut()

	handles := []*apitype.Handle{
		apitype.NewHandle("/tmp", "foo1"),
	}
	sut.AddHandles(handles)

	t.Run("Initial image", func(t *testing.T) {
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(0, index)
		a.Equal("foo1", img.GetHandle().GetFile())
	})

	t.Run("Next stays on same", func(t *testing.T) {
		sut.RequestNextImage()
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(0, index)
		a.Equal("foo1", img.GetHandle().GetFile())
	})

	t.Run("Previous stays on same", func(t *testing.T) {
		sut.RequestPrevImage()
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(0, index)
		a.Equal("foo1", img.GetHandle().GetFile())
	})
}

func TestGetCurrentImage_Navigate_ManyImages(t *testing.T) {
	a := assert.New(t)

	sut = initializeSut()

	handles := []*apitype.Handle{
		apitype.NewHandle("/tmp", "foo1"),
		apitype.NewHandle("/tmp", "foo2"),
		apitype.NewHandle("/tmp", "foo3"),
	}
	sut.AddHandles(handles)

	t.Run("Initial image", func(t *testing.T) {
		sut.RequestPrevImage()
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(0, index)
		a.Equal("foo1", img.GetHandle().GetFile())

	})

	t.Run("Next image", func(t *testing.T) {
		sut.RequestNextImage()
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(1, index)
		a.Equal("foo2", img.GetHandle().GetFile())

	})

	t.Run("Next again", func(t *testing.T) {
		sut.RequestNextImage()
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(2, index)
		a.Equal("foo3", img.GetHandle().GetFile())
	})

	t.Run("Next again should stay", func(t *testing.T) {
		sut.RequestNextImage()
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(2, index)
		a.Equal("foo3", img.GetHandle().GetFile())
	})

	t.Run("Previous", func(t *testing.T) {
		sut.RequestPrevImage()
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(1, index)
		a.Equal("foo2", img.GetHandle().GetFile())
	})
}

func TestGetCurrentImage_Navigate_Jump(t *testing.T) {
	a := assert.New(t)

	sut = initializeSut()

	handles := []*apitype.Handle{
		apitype.NewHandle("/tmp", "foo0"),
		apitype.NewHandle("/tmp", "foo1"),
		apitype.NewHandle("/tmp", "foo2"),
		apitype.NewHandle("/tmp", "foo3"),
		apitype.NewHandle("/tmp", "foo4"),
		apitype.NewHandle("/tmp", "foo5"),
		apitype.NewHandle("/tmp", "foo6"),
		apitype.NewHandle("/tmp", "foo7"),
		apitype.NewHandle("/tmp", "foo8"),
		apitype.NewHandle("/tmp", "foo9"),
	}
	sut.AddHandles(handles)

	t.Run("Jump to forward 5 images", func(t *testing.T) {
		sut.MoveToNextImageWithOffset(5)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(5, index)
		a.Equal("foo5", img.GetHandle().GetFile())
	})

	t.Run("Jump beyond the last", func(t *testing.T) {
		sut.MoveToNextImageWithOffset(10)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(9, index)
		a.Equal("foo9", img.GetHandle().GetFile())
	})

	t.Run("Jump back to 5 images", func(t *testing.T) {
		sut.MoveToPrevImageWithOffset(5)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(4, index)
		a.Equal("foo4", img.GetHandle().GetFile())
	})

	t.Run("Jump beyond the first", func(t *testing.T) {
		sut.MoveToPrevImageWithOffset(10)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(0, index)
		a.Equal("foo0", img.GetHandle().GetFile())
	})
}

func TestGetCurrentImage_Navigate_AtIndex(t *testing.T) {
	a := assert.New(t)

	sut = initializeSut()

	handles := []*apitype.Handle{
		apitype.NewHandle("/tmp", "foo0"),
		apitype.NewHandle("/tmp", "foo1"),
		apitype.NewHandle("/tmp", "foo2"),
		apitype.NewHandle("/tmp", "foo3"),
		apitype.NewHandle("/tmp", "foo4"),
		apitype.NewHandle("/tmp", "foo5"),
		apitype.NewHandle("/tmp", "foo6"),
		apitype.NewHandle("/tmp", "foo7"),
		apitype.NewHandle("/tmp", "foo8"),
		apitype.NewHandle("/tmp", "foo9"),
	}
	sut.AddHandles(handles)

	t.Run("Index first image", func(t *testing.T) {
		sut.MoveToImageAt(0)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(0, index)
		a.Equal("foo0", img.GetHandle().GetFile())
	})

	t.Run("Index 5", func(t *testing.T) {
		sut.MoveToImageAt(5)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(5, index)
		a.Equal("foo5", img.GetHandle().GetFile())
	})

	t.Run("Index last image", func(t *testing.T) {
		sut.MoveToImageAt(9)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(9, index)
		a.Equal("foo9", img.GetHandle().GetFile())
	})

	t.Run("Index after the last gives the last image", func(t *testing.T) {
		sut.MoveToImageAt(10)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(9, index)
		a.Equal("foo9", img.GetHandle().GetFile())
	})

	t.Run("Index last image with negative index", func(t *testing.T) {
		sut.MoveToImageAt(-1)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(9, index)
		a.Equal("foo9", img.GetHandle().GetFile())
	})

	t.Run("Index second to last image with negative index", func(t *testing.T) {
		sut.MoveToImageAt(-2)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(8, index)
		a.Equal("foo8", img.GetHandle().GetFile())
	})

	t.Run("Too big negative index returns the first", func(t *testing.T) {
		sut.MoveToImageAt(-100)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(0, index)
		a.Equal("foo0", img.GetHandle().GetFile())
	})
}

func TestGetCurrentImage_Navigate_Handle(t *testing.T) {
	a := assert.New(t)

	sut = initializeSut()

	handles := []*apitype.Handle{
		apitype.NewHandle("/tmp", "foo0"),
		apitype.NewHandle("/tmp", "foo1"),
		apitype.NewHandle("/tmp", "foo2"),
		apitype.NewHandle("/tmp", "foo3"),
		apitype.NewHandle("/tmp", "foo4"),
	}
	sut.AddHandles(handles)
	handles, _ = imageStore.GetAllImages()

	t.Run("foo1", func(t *testing.T) {
		sut.MoveToImage(handles[1])
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(1, index)
		a.Equal("foo1", img.GetHandle().GetFile())
	})
	t.Run("foo3", func(t *testing.T) {
		sut.MoveToImage(handles[3])
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(3, index)
		a.Equal("foo3", img.GetHandle().GetFile())
	})

	t.Run("Nil stays on the current image", func(t *testing.T) {
		sut.MoveToImage(nil)
		img, index, _ := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(3, index)
		a.Equal("foo3", img.GetHandle().GetFile())
	})
}

func TestGetNextImages(t *testing.T) {
	a := assert.New(t)

	sut = initializeSut()

	handles := []*apitype.Handle{
		apitype.NewHandle("/tmp", "foo0"),
		apitype.NewHandle("/tmp", "foo1"),
		apitype.NewHandle("/tmp", "foo2"),
		apitype.NewHandle("/tmp", "foo3"),
		apitype.NewHandle("/tmp", "foo4"),
		apitype.NewHandle("/tmp", "foo5"),
		apitype.NewHandle("/tmp", "foo6"),
		apitype.NewHandle("/tmp", "foo7"),
		apitype.NewHandle("/tmp", "foo8"),
		apitype.NewHandle("/tmp", "foo9"),
	}
	sut.AddHandles(handles)
	sut.SetImageListSize(5)
	a.Equal(5, sut.getImageListSize())

	t.Run("Initial image count", func(t *testing.T) {
		imgList, _ := sut.getNextImages()
		a.NotNil(imgList)
		if a.Equal(5, len(imgList)) {
			a.Equal("foo1", imgList[0].GetHandle().GetFile())
			a.Equal("foo2", imgList[1].GetHandle().GetFile())
			a.Equal("foo3", imgList[2].GetHandle().GetFile())
			a.Equal("foo4", imgList[3].GetHandle().GetFile())
			a.Equal("foo5", imgList[4].GetHandle().GetFile())
		}
	})

	t.Run("Next requested gives the next 5", func(t *testing.T) {
		sut.RequestNextImage()
		imgList, _ := sut.getNextImages()
		a.NotNil(imgList)
		if a.Equal(5, len(imgList)) {
			a.Equal("foo2", imgList[0].GetHandle().GetFile())
			a.Equal("foo3", imgList[1].GetHandle().GetFile())
			a.Equal("foo4", imgList[2].GetHandle().GetFile())
			a.Equal("foo5", imgList[3].GetHandle().GetFile())
			a.Equal("foo6", imgList[4].GetHandle().GetFile())
		}
	})

	t.Run("If no more next images, dont return more", func(t *testing.T) {
		sut.MoveToImageAt(6)
		imgList, _ := sut.getNextImages()
		a.NotNil(imgList)
		if a.Equal(3, len(imgList)) {
			a.Equal("foo7", imgList[0].GetHandle().GetFile())
			a.Equal("foo8", imgList[1].GetHandle().GetFile())
			a.Equal("foo9", imgList[2].GetHandle().GetFile())
		}
	})

	t.Run("Second to last", func(t *testing.T) {
		sut.MoveToImageAt(8)
		imgList, _ := sut.getNextImages()
		a.NotNil(imgList)
		if a.Equal(1, len(imgList)) {
			a.Equal("foo9", imgList[0].GetHandle().GetFile())
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

	handles := []*apitype.Handle{
		apitype.NewHandle("/tmp", "foo0"),
		apitype.NewHandle("/tmp", "foo1"),
		apitype.NewHandle("/tmp", "foo2"),
		apitype.NewHandle("/tmp", "foo3"),
		apitype.NewHandle("/tmp", "foo4"),
		apitype.NewHandle("/tmp", "foo5"),
		apitype.NewHandle("/tmp", "foo6"),
		apitype.NewHandle("/tmp", "foo7"),
		apitype.NewHandle("/tmp", "foo8"),
		apitype.NewHandle("/tmp", "foo9"),
	}
	sut.AddHandles(handles)
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
			a.Equal("foo0", imgList[0].GetHandle().GetFile())
		}
	})

	t.Run("Image at 5 gives the first 5 images", func(t *testing.T) {
		sut.MoveToImageAt(5)
		imgList, _ := sut.getPrevImages()
		a.NotNil(imgList)
		if a.Equal(5, len(imgList)) {
			a.Equal("foo4", imgList[0].GetHandle().GetFile())
			a.Equal("foo3", imgList[1].GetHandle().GetFile())
			a.Equal("foo2", imgList[2].GetHandle().GetFile())
			a.Equal("foo1", imgList[3].GetHandle().GetFile())
			a.Equal("foo0", imgList[4].GetHandle().GetFile())
		}
	})

	t.Run("Second to last image ", func(t *testing.T) {
		sut.MoveToImageAt(8)
		imgList, _ := sut.getPrevImages()
		a.NotNil(imgList)
		if a.Equal(5, len(imgList)) {
			a.Equal("foo7", imgList[0].GetHandle().GetFile())
			a.Equal("foo6", imgList[1].GetHandle().GetFile())
			a.Equal("foo5", imgList[2].GetHandle().GetFile())
			a.Equal("foo4", imgList[3].GetHandle().GetFile())
			a.Equal("foo3", imgList[4].GetHandle().GetFile())
		}
	})

	t.Run("The last", func(t *testing.T) {
		sut.MoveToImageAt(9)
		imgList, _ := sut.getPrevImages()
		a.NotNil(imgList)
		if a.Equal(5, len(imgList)) {
			a.Equal("foo8", imgList[0].GetHandle().GetFile())
			a.Equal("foo7", imgList[1].GetHandle().GetFile())
			a.Equal("foo6", imgList[2].GetHandle().GetFile())
			a.Equal("foo5", imgList[3].GetHandle().GetFile())
			a.Equal("foo4", imgList[4].GetHandle().GetFile())
		}
	})
}

func TestGetTotalCount(t *testing.T) {
	a := assert.New(t)

	sut = initializeSut()

	handles := []*apitype.Handle{
		apitype.NewHandle("/tmp", "foo0"),
		apitype.NewHandle("/tmp", "foo1"),
		apitype.NewHandle("/tmp", "foo2"),
		apitype.NewHandle("/tmp", "foo3"),
		apitype.NewHandle("/tmp", "foo4"),
		apitype.NewHandle("/tmp", "foo5"),
		apitype.NewHandle("/tmp", "foo6"),
		apitype.NewHandle("/tmp", "foo7"),
		apitype.NewHandle("/tmp", "foo8"),
		apitype.NewHandle("/tmp", "foo9"),
	}
	sut.AddHandles(handles)

	a.Equal(10, sut.getTotalImages())
}

// Show only images

func TestShowOnlyImages(t *testing.T) {
	a := assert.New(t)

	sut := initializeSut()

	handles := []*apitype.Handle{
		apitype.NewHandle("/tmp", "foo0"),
		apitype.NewHandle("/tmp", "foo1"),
		apitype.NewHandle("/tmp", "foo2"),
		apitype.NewHandle("/tmp", "foo3"),
		apitype.NewHandle("/tmp", "foo4"),
		apitype.NewHandle("/tmp", "foo5"),
		apitype.NewHandle("/tmp", "foo6"),
		apitype.NewHandle("/tmp", "foo7"),
		apitype.NewHandle("/tmp", "foo8"),
		apitype.NewHandle("/tmp", "foo9"),
	}
	sut.AddHandles(handles)
	handles, _ = imageStore.GetAllImages()
	category1, _ := categoryStore.AddCategory(apitype.NewCategory("category1", "cat1", "C"))
	category2, _ := categoryStore.AddCategory(apitype.NewCategory("category2", "cat2", "D"))
	sut.SetImageListSize(10)

	_ = imageCategoryStore.CategorizeImage(handles[1].GetId(), category1.GetId(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(handles[2].GetId(), category1.GetId(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(handles[6].GetId(), category1.GetId(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(handles[7].GetId(), category1.GetId(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(handles[9].GetId(), category1.GetId(), apitype.MOVE)

	_ = imageCategoryStore.CategorizeImage(handles[0].GetId(), category2.GetId(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(handles[1].GetId(), category2.GetId(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(handles[3].GetId(), category2.GetId(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(handles[9].GetId(), category2.GetId(), apitype.MOVE)

	sut.ShowOnlyImages(category1.GetName())

	a.Equal(5, sut.getTotalImages())
	a.Equal("category1", sut.getCurrentCategoryName())

	t.Run("Next and prev images", func(t *testing.T) {
		nextImages, _ := sut.getNextImages()
		prevImages, _ := sut.getPrevImages()
		a.NotNil(nextImages)
		if a.Equal(4, len(nextImages)) {
			a.Equal(handles[2].GetId(), nextImages[0].GetHandle().GetId())
			a.Equal("foo2", nextImages[0].GetHandle().GetFile())
			a.Equal(handles[6].GetId(), nextImages[1].GetHandle().GetId())
			a.Equal("foo6", nextImages[1].GetHandle().GetFile())
			a.Equal(handles[7].GetId(), nextImages[2].GetHandle().GetId())
			a.Equal("foo7", nextImages[2].GetHandle().GetFile())
			a.Equal(handles[9].GetId(), nextImages[3].GetHandle().GetId())
			a.Equal("foo9", nextImages[3].GetHandle().GetFile())
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
			a.Equal(handles[7].GetId(), nextImages[0].GetHandle().GetId())
			a.Equal("foo7", nextImages[0].GetHandle().GetFile())
			a.Equal(handles[9].GetId(), nextImages[1].GetHandle().GetId())
			a.Equal("foo9", nextImages[1].GetHandle().GetFile())
		}

		a.NotNil(prevImages)
		if a.Equal(2, len(prevImages)) {
			a.Equal(handles[1].GetId(), prevImages[1].GetHandle().GetId())
			a.Equal("foo1", prevImages[1].GetHandle().GetFile())
			a.Equal(handles[2].GetId(), prevImages[0].GetHandle().GetId())
			a.Equal("foo2", prevImages[0].GetHandle().GetFile())
		}
	})
}

func TestShowOnlyImages_ShowAllAgain(t *testing.T) {
	a := assert.New(t)

	sut := initializeSut()

	handles := []*apitype.Handle{
		apitype.NewHandle("/tmp", "foo0"),
		apitype.NewHandle("/tmp", "foo1"),
		apitype.NewHandle("/tmp", "foo2"),
		apitype.NewHandle("/tmp", "foo3"),
		apitype.NewHandle("/tmp", "foo4"),
		apitype.NewHandle("/tmp", "foo5"),
		apitype.NewHandle("/tmp", "foo6"),
		apitype.NewHandle("/tmp", "foo7"),
		apitype.NewHandle("/tmp", "foo8"),
		apitype.NewHandle("/tmp", "foo9"),
	}
	sut.AddHandles(handles)
	sut.SetImageListSize(10)

	handles, _ = imageStore.GetAllImages()
	category1, _ := categoryStore.AddCategory(apitype.NewCategory("category1", "C1", "1"))
	_ = imageCategoryStore.CategorizeImage(handles[1].GetId(), category1.GetId(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(handles[2].GetId(), category1.GetId(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(handles[6].GetId(), category1.GetId(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(handles[7].GetId(), category1.GetId(), apitype.MOVE)
	_ = imageCategoryStore.CategorizeImage(handles[9].GetId(), category1.GetId(), apitype.MOVE)

	sut.ShowOnlyImages("category1")

	a.Equal(5, sut.getTotalImages())
	a.Equal("category1", sut.getCurrentCategoryName())

	sut.ShowAllImages()
	a.Equal(10, sut.getTotalImages())
	a.Equal("", sut.getCurrentCategoryName())

	t.Run("Next and prev images", func(t *testing.T) {
		nextImages, _ := sut.getNextImages()
		prevImages, _ := sut.getPrevImages()
		a.NotNil(nextImages)
		if a.Equal(9, len(nextImages)) {
			a.Equal("foo1", nextImages[0].GetHandle().GetFile())
			a.Equal("foo2", nextImages[1].GetHandle().GetFile())
			a.Equal("foo3", nextImages[2].GetHandle().GetFile())
			a.Equal("foo4", nextImages[3].GetHandle().GetFile())
			a.Equal("foo5", nextImages[4].GetHandle().GetFile())
			a.Equal("foo6", nextImages[5].GetHandle().GetFile())
			a.Equal("foo7", nextImages[6].GetHandle().GetFile())
			a.Equal("foo8", nextImages[7].GetHandle().GetFile())
			a.Equal("foo9", nextImages[8].GetHandle().GetFile())
		}

		a.NotNil(prevImages)
		a.Equal(0, len(prevImages))
	})
}
