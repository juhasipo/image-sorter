package library

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"image"
	"os"
	"testing"
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

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

var (
	sut     *internalManager
	sender  *MockSender
	store   *MockImageStore
	loader  *MockImageLoader
	dbStore *database.Store
)

func setup() {
	sender = new(MockSender)
	store = new(MockImageStore)
	loader = new(MockImageLoader)

	store.On("GetFull", mock.Anything).Return(new(MockImage))
	store.On("GetThumbnail", mock.Anything).Return(new(MockImage))

	dbStore = database.NewInMemoryStore()

	sut = newLibrary(store, loader, dbStore)
}

func TestGetCurrentImage_Navigate_Empty(t *testing.T) {
	a := assert.New(t)

	img, index := sut.getCurrentImage()
	a.NotNil(img)
	a.Equal(0, index)
	a.False(img.GetHandle().IsValid())
}

func TestGetCurrentImage_Navigate_OneImage(t *testing.T) {
	a := assert.New(t)

	handles := []*apitype.Handle{
		apitype.NewHandle("/tmp", "foo1"),
	}
	sut.AddHandles(handles)

	t.Run("Initial image", func(t *testing.T) {
		img, index := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(0, index)
		a.Equal("foo1", img.GetHandle().GetFile())
	})

	t.Run("Next stays on same", func(t *testing.T) {
		sut.RequestNextImage()
		img, index := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(0, index)
		a.Equal("foo1", img.GetHandle().GetFile())
	})

	t.Run("Previous stays on same", func(t *testing.T) {
		sut.RequestPrevImage()
		img, index := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(0, index)
		a.Equal("foo1", img.GetHandle().GetFile())
	})
}

func TestGetCurrentImage_Navigate_ManyImages(t *testing.T) {
	a := assert.New(t)

	handles := []*apitype.Handle{
		apitype.NewHandle("/tmp", "foo1"),
		apitype.NewHandle("/tmp", "foo2"),
		apitype.NewHandle("/tmp", "foo3"),
	}
	sut.AddHandles(handles)

	t.Run("Initial image", func(t *testing.T) {
		sut.RequestPrevImage()
		img, index := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(0, index)
		a.Equal("foo1", img.GetHandle().GetFile())

	})

	t.Run("Next image", func(t *testing.T) {
		sut.RequestNextImage()
		img, index := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(1, index)
		a.Equal("foo2", img.GetHandle().GetFile())

	})

	t.Run("Next again", func(t *testing.T) {
		sut.RequestNextImage()
		img, index := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(2, index)
		a.Equal("foo3", img.GetHandle().GetFile())
	})

	t.Run("Next again should stay", func(t *testing.T) {
		sut.RequestNextImage()
		img, index := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(2, index)
		a.Equal("foo3", img.GetHandle().GetFile())
	})

	t.Run("Previous", func(t *testing.T) {
		sut.RequestPrevImage()
		img, index := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(1, index)
		a.Equal("foo2", img.GetHandle().GetFile())
	})
}

func TestGetCurrentImage_Navigate_Jump(t *testing.T) {
	a := assert.New(t)

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
		img, index := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(5, index)
		a.Equal("foo5", img.GetHandle().GetFile())
	})

	t.Run("Jump beyond the last", func(t *testing.T) {
		sut.MoveToNextImageWithOffset(10)
		img, index := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(9, index)
		a.Equal("foo9", img.GetHandle().GetFile())
	})

	t.Run("Jump back to 5 images", func(t *testing.T) {
		sut.MoveToPrevImageWithOffset(5)
		img, index := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(4, index)
		a.Equal("foo4", img.GetHandle().GetFile())
	})

	t.Run("Jump beyond the first", func(t *testing.T) {
		sut.MoveToPrevImageWithOffset(10)
		img, index := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(0, index)
		a.Equal("foo0", img.GetHandle().GetFile())
	})
}

func TestGetCurrentImage_Navigate_AtIndex(t *testing.T) {
	a := assert.New(t)

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
		img, index := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(0, index)
		a.Equal("foo0", img.GetHandle().GetFile())
	})

	t.Run("Index 5", func(t *testing.T) {
		sut.MoveToImageAt(5)
		img, index := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(5, index)
		a.Equal("foo5", img.GetHandle().GetFile())
	})

	t.Run("Index last image", func(t *testing.T) {
		sut.MoveToImageAt(9)
		img, index := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(9, index)
		a.Equal("foo9", img.GetHandle().GetFile())
	})

	t.Run("Index after the last gives the last image", func(t *testing.T) {
		sut.MoveToImageAt(10)
		img, index := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(9, index)
		a.Equal("foo9", img.GetHandle().GetFile())
	})

	t.Run("Index last image with negative index", func(t *testing.T) {
		sut.MoveToImageAt(-1)
		img, index := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(9, index)
		a.Equal("foo9", img.GetHandle().GetFile())
	})

	t.Run("Index second to last image with negative index", func(t *testing.T) {
		sut.MoveToImageAt(-2)
		img, index := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(8, index)
		a.Equal("foo8", img.GetHandle().GetFile())
	})

	t.Run("Too big negative index returns the first", func(t *testing.T) {
		sut.MoveToImageAt(-100)
		img, index := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(0, index)
		a.Equal("foo0", img.GetHandle().GetFile())
	})
}

func TestGetCurrentImage_Navigate_Handle(t *testing.T) {
	a := assert.New(t)

	handles := []*apitype.Handle{
		apitype.NewHandle("/tmp", "foo0"),
		apitype.NewHandle("/tmp", "foo1"),
		apitype.NewHandle("/tmp", "foo2"),
		apitype.NewHandle("/tmp", "foo3"),
		apitype.NewHandle("/tmp", "foo4"),
	}
	sut.AddHandles(handles)
	handles, _ = dbStore.GetImages(-1, 0)

	t.Run("foo1", func(t *testing.T) {
		sut.MoveToImage(handles[1])
		img, index := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(1, index)
		a.Equal("foo1", img.GetHandle().GetFile())
	})
	t.Run("foo3", func(t *testing.T) {
		sut.MoveToImage(handles[3])
		img, index := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(3, index)
		a.Equal("foo3", img.GetHandle().GetFile())
	})

	t.Run("Nil stays on the current image", func(t *testing.T) {
		sut.MoveToImage(nil)
		img, index := sut.getCurrentImage()
		a.NotNil(img)
		a.Equal(3, index)
		a.Equal("foo3", img.GetHandle().GetFile())
	})
}

func TestGetNextImages(t *testing.T) {
	a := assert.New(t)

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
		imgList := sut.getNextImages()
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
		imgList := sut.getNextImages()
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
		imgList := sut.getNextImages()
		a.NotNil(imgList)
		if a.Equal(3, len(imgList)) {
			a.Equal("foo7", imgList[0].GetHandle().GetFile())
			a.Equal("foo8", imgList[1].GetHandle().GetFile())
			a.Equal("foo9", imgList[2].GetHandle().GetFile())
		}
	})

	t.Run("Second to last", func(t *testing.T) {
		sut.MoveToImageAt(8)
		imgList := sut.getNextImages()
		a.NotNil(imgList)
		if a.Equal(1, len(imgList)) {
			a.Equal("foo9", imgList[0].GetHandle().GetFile())
		}
	})

	t.Run("The last", func(t *testing.T) {
		sut.MoveToImageAt(9)
		imgList := sut.getNextImages()
		a.NotNil(imgList)
		a.Equal(0, len(imgList))
	})

}

func TestGetPrevImages(t *testing.T) {
	a := assert.New(t)

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
		imgList := sut.getPrevImages()
		a.NotNil(imgList)
		a.Equal(0, len(imgList))
	})

	t.Run("Next requested gives the first image", func(t *testing.T) {
		sut.RequestNextImage()
		imgList := sut.getPrevImages()
		a.NotNil(imgList)
		if a.Equal(1, len(imgList)) {
			a.Equal("foo0", imgList[0].GetHandle().GetFile())
		}
	})

	t.Run("Image at 5 gives the first 5 images", func(t *testing.T) {
		sut.MoveToImageAt(5)
		imgList := sut.getPrevImages()
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
		imgList := sut.getPrevImages()
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
		imgList := sut.getPrevImages()
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
	handles, _ = dbStore.GetImages(-1, 0)
	category, _ := dbStore.AddCategory(apitype.NewCategory("category1", "cat", "C"))
	sut.SetImageListSize(10)

	_ = dbStore.CategorizeImage(handles[1].GetId(), category.GetId(), apitype.MOVE)
	_ = dbStore.CategorizeImage(handles[2].GetId(), category.GetId(), apitype.MOVE)
	_ = dbStore.CategorizeImage(handles[6].GetId(), category.GetId(), apitype.MOVE)
	_ = dbStore.CategorizeImage(handles[7].GetId(), category.GetId(), apitype.MOVE)
	_ = dbStore.CategorizeImage(handles[9].GetId(), category.GetId(), apitype.MOVE)
	sut.ShowOnlyImages(category.GetName(), nil)

	a.Equal(5, sut.getTotalImages())
	a.Equal("category1", sut.getCurrentCategoryName())

	t.Run("Next and prev images", func(t *testing.T) {
		nextImages := sut.getNextImages()
		prevImages := sut.getPrevImages()
		a.NotNil(nextImages)
		if a.Equal(4, len(nextImages)) {
			a.Equal("foo2", nextImages[0].GetHandle().GetFile())
			a.Equal("foo6", nextImages[1].GetHandle().GetFile())
			a.Equal("foo7", nextImages[2].GetHandle().GetFile())
			a.Equal("foo9", nextImages[3].GetHandle().GetFile())
		}

		a.NotNil(prevImages)
		a.Equal(0, len(prevImages))
	})

	t.Run("Next and prev images at 2", func(t *testing.T) {
		sut.MoveToImageAt(2)
		nextImages := sut.getNextImages()
		prevImages := sut.getPrevImages()
		a.NotNil(nextImages)
		if a.Equal(2, len(nextImages)) {
			a.Equal("foo7", nextImages[0].GetHandle().GetFile())
			a.Equal("foo9", nextImages[1].GetHandle().GetFile())
		}

		a.NotNil(prevImages)
		if a.Equal(2, len(prevImages)) {
			a.Equal("foo1", prevImages[1].GetHandle().GetFile())
			a.Equal("foo2", prevImages[0].GetHandle().GetFile())
		}
	})
}

func TestShowOnlyImages_ShowAllAgain(t *testing.T) {
	a := assert.New(t)

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

	categoryHandles := []*apitype.Handle{
		apitype.NewHandle("/tmp", "foo1"),
		apitype.NewHandle("/tmp", "foo2"),
		apitype.NewHandle("/tmp", "foo6"),
		apitype.NewHandle("/tmp", "foo7"),
		apitype.NewHandle("/tmp", "foo9"),
	}
	sut.ShowOnlyImages("category1", categoryHandles)

	a.Equal(5, sut.getTotalImages())
	a.Equal("category1", sut.getCurrentCategoryName())

	sut.ShowAllImages()
	a.Equal(10, sut.getTotalImages())
	a.Equal("", sut.getCurrentCategoryName())

	t.Run("Next and prev images", func(t *testing.T) {
		nextImages := sut.getNextImages()
		prevImages := sut.getPrevImages()
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
