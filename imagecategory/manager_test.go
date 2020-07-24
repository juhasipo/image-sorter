package imagecategory

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"image"
	"testing"
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/filter"
	"vincit.fi/image-sorter/imageloader"
	"vincit.fi/image-sorter/imageloader/goimage"
	"vincit.fi/image-sorter/library"
)

type MockSender struct {
	event.Sender
	mock.Mock
}

type MockLibrary struct {
	library.Library
	mock.Mock
}

type MockImageCache struct {
	imageloader.ImageCache
	mock.Mock
}

type MockImageLoader struct {
	goimage.ImageLoader
	mock.Mock
}

func (s *MockSender) SendToTopic(topic event.Topic) {
	s.Called(topic)
}

func (s *MockSender) SendToTopicWithData(topic event.Topic, data ...interface{}) {
	s.Called(topic, data)
}

func (s *MockImageLoader) LoadImage(*common.Handle) (image.Image, error) {
	return nil, nil
}

func (s *MockImageLoader) LoadExifData(*common.Handle) (*common.ExifData, error) {
	return nil, nil
}

//// Basic cases

func TestCategorizeOne(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", event.IMAGE_REQUEST_NEXT).Return()
	lib := new(MockLibrary)
	filterManager := filter.FilterManagerNew()
	imageLoader := new(MockImageLoader)

	sut := ManagerNew(sender, lib, filterManager, imageLoader)

	cat1 := common.CategoryEntryNew("Cat 1", "c1", "C")
	handle := common.HandleNew("/tmp", "foo")
	cmd := category.CategorizeCommandNewWithStayAttr(handle, cat1, common.MOVE, false, false)
	sut.SetCategory(cmd)

	result := sut.GetCategories(handle)

	a.Equal(1, len(result))
	a.Equal("Cat 1", result["Cat 1"].GetEntry().GetName())
}

func TestCategorizeOneToTwoCategories(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", event.IMAGE_REQUEST_NEXT).Return()
	lib := new(MockLibrary)
	filterManager := filter.FilterManagerNew()
	imageLoader := new(MockImageLoader)

	sut := ManagerNew(sender, lib, filterManager, imageLoader)

	cat1 := common.CategoryEntryNew("Cat 1", "c1", "C")
	cat2 := common.CategoryEntryNew("Cat 2", "c2", "D")
	handle := common.HandleNew("/tmp", "foo")
	cmd1 := category.CategorizeCommandNewWithStayAttr(handle, cat1, common.MOVE, false, false)
	cmd2 := category.CategorizeCommandNewWithStayAttr(handle, cat2, common.MOVE, false, false)
	sut.SetCategory(cmd1)
	sut.SetCategory(cmd2)

	result := sut.GetCategories(handle)

	a.Equal(2, len(result))
	a.Equal("Cat 1", result["Cat 1"].GetEntry().GetName())
	a.Equal("Cat 2", result["Cat 2"].GetEntry().GetName())
}

func TestCategorizeOneRemoveCategory(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", event.IMAGE_REQUEST_NEXT).Return()
	sender.On("SendToTopicWithData", event.CATEGORY_IMAGE_UPDATE, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.FilterManagerNew()
	imageLoader := new(MockImageLoader)

	sut := ManagerNew(sender, lib, filterManager, imageLoader)

	cat1 := common.CategoryEntryNew("Cat 1", "c1", "C")
	cat2 := common.CategoryEntryNew("Cat 2", "c2", "D")
	handle := common.HandleNew("/tmp", "foo")
	sut.SetCategory(category.CategorizeCommandNewWithStayAttr(handle, cat1, common.MOVE, false, false))
	sut.SetCategory(category.CategorizeCommandNewWithStayAttr(handle, cat2, common.MOVE, false, false))
	sut.SetCategory(category.CategorizeCommandNewWithStayAttr(handle, cat1, common.NONE, false, false))

	result := sut.GetCategories(handle)

	a.Equal(1, len(result))
	a.Equal("Cat 2", result["Cat 2"].GetEntry().GetName())
}

func TestCategorizeOneRemoveAll(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", event.IMAGE_REQUEST_NEXT).Return()
	sender.On("SendToTopicWithData", event.CATEGORY_IMAGE_UPDATE, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.FilterManagerNew()
	imageLoader := new(MockImageLoader)

	sut := ManagerNew(sender, lib, filterManager, imageLoader)

	cat1 := common.CategoryEntryNew("Cat 1", "c1", "C")
	cat2 := common.CategoryEntryNew("Cat 2", "c2", "D")
	handle := common.HandleNew("/tmp", "foo")
	sut.SetCategory(category.CategorizeCommandNewWithStayAttr(handle, cat1, common.MOVE, false, false))
	sut.SetCategory(category.CategorizeCommandNewWithStayAttr(handle, cat2, common.MOVE, false, false))
	sut.SetCategory(category.CategorizeCommandNewWithStayAttr(handle, cat1, common.NONE, false, false))
	sut.SetCategory(category.CategorizeCommandNewWithStayAttr(handle, cat2, common.NONE, false, false))

	result := sut.GetCategories(handle)

	a.Equal(0, len(result))
}

//// Force category

func TestCategorizeForceToCategory(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", event.IMAGE_REQUEST_NEXT).Return()
	sender.On("SendToTopicWithData", event.CATEGORY_IMAGE_UPDATE, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.FilterManagerNew()
	imageLoader := new(MockImageLoader)

	sut := ManagerNew(sender, lib, filterManager, imageLoader)

	cat1 := common.CategoryEntryNew("Cat 1", "c1", "C")
	cat2 := common.CategoryEntryNew("Cat 2", "c2", "D")
	cat3 := common.CategoryEntryNew("Cat 3", "c3", "E")
	handle := common.HandleNew("/tmp", "foo")
	sut.SetCategory(category.CategorizeCommandNewWithStayAttr(handle, cat1, common.MOVE, false, false))
	sut.SetCategory(category.CategorizeCommandNewWithStayAttr(handle, cat2, common.MOVE, false, false))
	sut.SetCategory(category.CategorizeCommandNewWithStayAttr(handle, cat3, common.MOVE, false, true))

	result := sut.GetCategories(handle)

	a.Equal(1, len(result))
	if a.NotNil(result["Cat 3"]) {
		a.Equal("Cat 3", result["Cat 3"].GetEntry().GetName())
	}
}

func TestCategorizeForceToExistingCategory(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", event.IMAGE_REQUEST_NEXT).Return()
	sender.On("SendToTopicWithData", event.CATEGORY_IMAGE_UPDATE, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.FilterManagerNew()
	imageLoader := new(MockImageLoader)

	sut := ManagerNew(sender, lib, filterManager, imageLoader)

	cat1 := common.CategoryEntryNew("Cat 1", "c1", "C")
	cat2 := common.CategoryEntryNew("Cat 2", "c2", "D")
	handle := common.HandleNew("/tmp", "foo")
	sut.SetCategory(category.CategorizeCommandNewWithStayAttr(handle, cat1, common.MOVE, false, false))
	sut.SetCategory(category.CategorizeCommandNewWithStayAttr(handle, cat2, common.MOVE, false, true))

	result := sut.GetCategories(handle)

	a.Equal(1, len(result))
	a.Equal("Cat 2", result["Cat 2"].GetEntry().GetName())
}

func TestCategorizeForceToCategory_None(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", event.IMAGE_REQUEST_NEXT).Return()
	sender.On("SendToTopicWithData", event.CATEGORY_IMAGE_UPDATE, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.FilterManagerNew()
	imageLoader := new(MockImageLoader)

	sut := ManagerNew(sender, lib, filterManager, imageLoader)

	cat1 := common.CategoryEntryNew("Cat 1", "c1", "C")
	cat2 := common.CategoryEntryNew("Cat 2", "c2", "D")
	cat3 := common.CategoryEntryNew("Cat 3", "c3", "E")
	handle := common.HandleNew("/tmp", "foo")
	sut.SetCategory(category.CategorizeCommandNewWithStayAttr(handle, cat1, common.MOVE, false, false))
	sut.SetCategory(category.CategorizeCommandNewWithStayAttr(handle, cat2, common.MOVE, false, false))
	sut.SetCategory(category.CategorizeCommandNewWithStayAttr(handle, cat3, common.NONE, false, true))

	result := sut.GetCategories(handle)

	a.Equal(0, len(result))
}

func TestResolveFileOperations(t *testing.T) {
	a := require.New(t)

	sender := new(MockSender)
	imageCache := new(MockImageCache)
	imageLoader := new(MockImageLoader)
	imageLoader.On("LoadImage", event.IMAGE_REQUEST_NEXT).Return(nil, nil)
	lib := library.LibraryNew(sender, imageCache, imageLoader)
	filterManager := filter.FilterManagerNew()

	sut := ManagerNew(sender, lib, filterManager, imageLoader)
	handle := common.HandleNew("filepath", "filename")
	lib.AddHandles([]*common.Handle{handle})

	var imageCategories = map[string]map[string]*category.CategorizedImage{
		"filename": {
			"cat1": category.CategorizedImageNew(common.CategoryEntryNew("cat1", "cat_1", ""), common.MOVE),
		},
	}
	operations := sut.ResolveFileOperations(imageCategories, false)

	a.Equal(1, len(operations))

	ops := operations[0]
	a.Equal(1, len(ops.GetOperations()))
}

func TestResolveOperationsForGroup(t *testing.T) {
	a := require.New(t)

	sender := new(MockSender)
	imageCache := new(MockImageCache)
	imageLoader := new(MockImageLoader)
	imageLoader.On("LoadImage", event.IMAGE_REQUEST_NEXT).Return(nil, nil)
	lib := library.LibraryNew(sender, imageCache, imageLoader)
	filterManager := filter.FilterManagerNew()

	sut := ManagerNew(sender, lib, filterManager, imageLoader)

	var imageCategories = map[string]*category.CategorizedImage{
		"cat1": category.CategorizedImageNew(common.CategoryEntryNew("cat1", "cat_1", ""), common.MOVE),
	}
	handle := common.HandleNew("filepath", "filename")
	operations, err := sut.ResolveOperationsForGroup(handle, imageCategories, false)

	a.Nil(err)
	ops := operations.GetOperations()
	a.Equal(1, len(ops))
	a.Equal("Copy file 'filename' to 'filepath/cat_1'", ops[0].String())
}
