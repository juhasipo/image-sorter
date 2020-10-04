package imagecategory

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"image"
	"path/filepath"
	"testing"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/filter"
	"vincit.fi/image-sorter/backend/library"
	"vincit.fi/image-sorter/common/event"
)

type MockSender struct {
	event.Sender
	mock.Mock
}

type MockLibrary struct {
	api.Library
	mock.Mock
}

type MockImageCache struct {
	api.ImageStore
	mock.Mock
}

type MockImageLoader struct {
	api.ImageLoader
	mock.Mock
}

func (s *MockSender) SendToTopic(topic event.Topic) {
	s.Called(topic)
}

func (s *MockSender) SendToTopicWithData(topic event.Topic, data ...interface{}) {
	s.Called(topic, data)
}

func (s *MockImageLoader) LoadImage(*apitype.Handle) (image.Image, error) {
	return nil, nil
}

func (s *MockImageLoader) LoadExifData(*apitype.Handle) (*apitype.ExifData, error) {
	return nil, nil
}

//// Basic cases

func TestCategorizeOne(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", event.ImageRequestNext).Return()
	sender.On("SendToTopicWithData", event.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.NewFilterManager()
	imageLoader := new(MockImageLoader)

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader)

	cat1 := apitype.NewCategory("Cat 1", "c1", "C")
	handle := apitype.NewHandle("/tmp", "foo")
	cmd := apitype.NewCategorizeCommand(handle, cat1, apitype.MOVE)
	sut.SetCategory(cmd)

	result := sut.GetCategories(handle)

	a.Equal(1, len(result))
	a.Equal("Cat 1", result["Cat 1"].GetEntry().GetName())
}

func TestCategorizeOneToTwoCategories(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", event.ImageRequestNext).Return()
	sender.On("SendToTopicWithData", event.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.NewFilterManager()
	imageLoader := new(MockImageLoader)

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader)

	cat1 := apitype.NewCategory("Cat 1", "c1", "C")
	cat2 := apitype.NewCategory("Cat 2", "c2", "D")
	handle := apitype.NewHandle("/tmp", "foo")
	cmd1 := apitype.NewCategorizeCommand(handle, cat1, apitype.MOVE)
	cmd2 := apitype.NewCategorizeCommand(handle, cat2, apitype.MOVE)
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
	sender.On("SendToTopic", event.ImageRequestNext).Return()
	sender.On("SendToTopicWithData", event.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.NewFilterManager()
	imageLoader := new(MockImageLoader)

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader)

	cat1 := apitype.NewCategory("Cat 1", "c1", "C")
	cat2 := apitype.NewCategory("Cat 2", "c2", "D")
	handle := apitype.NewHandle("/tmp", "foo")
	sut.SetCategory(apitype.NewCategorizeCommand(handle, cat1, apitype.MOVE))
	sut.SetCategory(apitype.NewCategorizeCommand(handle, cat2, apitype.MOVE))
	sut.SetCategory(apitype.NewCategorizeCommand(handle, cat1, apitype.NONE))

	result := sut.GetCategories(handle)

	a.Equal(1, len(result))
	a.Equal("Cat 2", result["Cat 2"].GetEntry().GetName())
}

func TestCategorizeOneRemoveAll(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", event.ImageRequestNext).Return()
	sender.On("SendToTopicWithData", event.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.NewFilterManager()
	imageLoader := new(MockImageLoader)

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader)

	cat1 := apitype.NewCategory("Cat 1", "c1", "C")
	cat2 := apitype.NewCategory("Cat 2", "c2", "D")
	handle := apitype.NewHandle("/tmp", "foo")
	sut.SetCategory(apitype.NewCategorizeCommand(handle, cat1, apitype.MOVE))
	sut.SetCategory(apitype.NewCategorizeCommand(handle, cat2, apitype.MOVE))
	sut.SetCategory(apitype.NewCategorizeCommand(handle, cat1, apitype.NONE))
	sut.SetCategory(apitype.NewCategorizeCommand(handle, cat2, apitype.NONE))

	result := sut.GetCategories(handle)

	a.Equal(0, len(result))
}

//// Force category

func TestCategorizeForceToCategory(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", event.ImageRequestNext).Return()
	sender.On("SendToTopicWithData", event.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.NewFilterManager()
	imageLoader := new(MockImageLoader)

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader)

	cat1 := apitype.NewCategory("Cat 1", "c1", "C")
	cat2 := apitype.NewCategory("Cat 2", "c2", "D")
	cat3 := apitype.NewCategory("Cat 3", "c3", "E")
	handle := apitype.NewHandle("/tmp", "foo")
	sut.SetCategory(apitype.NewCategorizeCommand(handle, cat1, apitype.MOVE))
	sut.SetCategory(apitype.NewCategorizeCommand(handle, cat2, apitype.MOVE))
	command := apitype.NewCategorizeCommand(handle, cat3, apitype.MOVE)
	command.SetForceToCategory(true)
	sut.SetCategory(command)

	result := sut.GetCategories(handle)

	a.Equal(1, len(result))
	if a.NotNil(result["Cat 3"]) {
		a.Equal("Cat 3", result["Cat 3"].GetEntry().GetName())
	}
}

func TestCategorizeForceToExistingCategory(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", event.ImageRequestNext).Return()
	sender.On("SendToTopicWithData", event.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.NewFilterManager()
	imageLoader := new(MockImageLoader)

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader)

	cat1 := apitype.NewCategory("Cat 1", "c1", "C")
	cat2 := apitype.NewCategory("Cat 2", "c2", "D")
	handle := apitype.NewHandle("/tmp", "foo")
	sut.SetCategory(apitype.NewCategorizeCommand(handle, cat1, apitype.MOVE))
	command := apitype.NewCategorizeCommand(handle, cat2, apitype.MOVE)
	command.SetForceToCategory(true)
	sut.SetCategory(command)

	result := sut.GetCategories(handle)

	a.Equal(1, len(result))
	a.Equal("Cat 2", result["Cat 2"].GetEntry().GetName())
}

func TestCategorizeForceToCategory_None(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", event.ImageRequestNext).Return()
	sender.On("SendToTopicWithData", event.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.NewFilterManager()
	imageLoader := new(MockImageLoader)

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader)

	cat1 := apitype.NewCategory("Cat 1", "c1", "C")
	cat2 := apitype.NewCategory("Cat 2", "c2", "D")
	cat3 := apitype.NewCategory("Cat 3", "c3", "E")
	handle := apitype.NewHandle("/tmp", "foo")
	sut.SetCategory(apitype.NewCategorizeCommand(handle, cat1, apitype.MOVE))
	sut.SetCategory(apitype.NewCategorizeCommand(handle, cat2, apitype.MOVE))
	command := apitype.NewCategorizeCommand(handle, cat3, apitype.NONE)
	command.SetForceToCategory(true)
	sut.SetCategory(command)

	result := sut.GetCategories(handle)

	a.Equal(0, len(result))
}

func TestResolveFileOperations(t *testing.T) {
	a := require.New(t)

	sender := new(MockSender)
	imageCache := new(MockImageCache)
	imageLoader := new(MockImageLoader)
	imageLoader.On("LoadImage", event.ImageRequestNext).Return(nil, nil)
	lib := library.NewLibrary(sender, imageCache, imageLoader)
	filterManager := filter.NewFilterManager()

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader)
	handle := apitype.NewHandle("filepath", "filename")
	lib.AddHandles([]*apitype.Handle{handle})

	var imageCategories = map[string]map[string]*apitype.CategorizedImage{
		"filename": {
			"cat1": apitype.NewCategorizedImage(apitype.NewCategory("cat1", "cat_1", ""), apitype.MOVE),
		},
	}
	command := apitype.NewPersistCategorizationCommand(true, false, 100)
	operations := sut.ResolveFileOperations(imageCategories, command)

	a.Equal(1, len(operations))

	ops := operations[0]
	a.Equal(1, len(ops.GetOperations()))
}

func TestResolveOperationsForGroup_KeepOld(t *testing.T) {
	a := require.New(t)

	sender := new(MockSender)
	imageCache := new(MockImageCache)
	imageLoader := new(MockImageLoader)
	imageLoader.On("LoadImage", event.ImageRequestNext).Return(nil, nil)
	lib := library.NewLibrary(sender, imageCache, imageLoader)
	filterManager := filter.NewFilterManager()

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader)

	var imageCategories = map[string]*apitype.CategorizedImage{
		"cat1": apitype.NewCategorizedImage(apitype.NewCategory("cat1", "cat_1", ""), apitype.MOVE),
	}
	handle := apitype.NewHandle("filepath", "filename")
	command := apitype.NewPersistCategorizationCommand(true, false, 100)
	operations, err := sut.ResolveOperationsForGroup(handle, imageCategories, command)

	a.Nil(err)
	ops := operations.GetOperations()
	a.Equal(1, len(ops))
	a.Equal(fmt.Sprintf("Copy file 'filename' to '%s'", filepath.Join("filepath", "cat_1")), ops[0].String())
}

func TestResolveOperationsForGroup_RemoveOld(t *testing.T) {
	a := require.New(t)

	sender := new(MockSender)
	imageCache := new(MockImageCache)
	imageLoader := new(MockImageLoader)
	imageLoader.On("LoadImage", event.ImageRequestNext).Return(nil, nil)
	lib := library.NewLibrary(sender, imageCache, imageLoader)
	filterManager := filter.NewFilterManager()

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader)

	var imageCategories = map[string]*apitype.CategorizedImage{
		"cat1": apitype.NewCategorizedImage(apitype.NewCategory("cat1", "cat_1", ""), apitype.MOVE),
	}
	handle := apitype.NewHandle("filepath", "filename")
	command := apitype.NewPersistCategorizationCommand(false, false, 100)
	operations, err := sut.ResolveOperationsForGroup(handle, imageCategories, command)

	a.Nil(err)
	ops := operations.GetOperations()
	a.Equal(2, len(ops))
	a.Equal(fmt.Sprintf("Copy file 'filename' to '%s'", filepath.Join("filepath", "cat_1")), ops[0].String())
	a.Equal("Remove", ops[1].String())
}

func TestResolveOperationsForGroup_FixExifRotation(t *testing.T) {
	a := require.New(t)

	sender := new(MockSender)
	imageCache := new(MockImageCache)
	imageLoader := new(MockImageLoader)
	imageLoader.On("LoadImage", event.ImageRequestNext).Return(nil, nil)
	lib := library.NewLibrary(sender, imageCache, imageLoader)
	filterManager := filter.NewFilterManager()

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader)

	var imageCategories = map[string]*apitype.CategorizedImage{
		"cat1": apitype.NewCategorizedImage(apitype.NewCategory("cat1", "cat_1", ""), apitype.MOVE),
	}
	handle := apitype.NewHandle("filepath", "filename")
	command := apitype.NewPersistCategorizationCommand(true, true, 100)
	operations, err := sut.ResolveOperationsForGroup(handle, imageCategories, command)

	a.Nil(err)
	ops := operations.GetOperations()
	a.Equal(2, len(ops))
	a.Equal("Exif Rotate", ops[0].String())
	a.Equal(fmt.Sprintf("Copy file 'filename' to '%s'", filepath.Join("filepath", "cat_1")), ops[1].String())
}

func TestResolveOperationsForGroup_FixExifRotation_RemoveOld(t *testing.T) {
	a := require.New(t)

	sender := new(MockSender)
	imageCache := new(MockImageCache)
	imageLoader := new(MockImageLoader)
	imageLoader.On("LoadImage", event.ImageRequestNext).Return(nil, nil)
	lib := library.NewLibrary(sender, imageCache, imageLoader)
	filterManager := filter.NewFilterManager()

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader)

	var imageCategories = map[string]*apitype.CategorizedImage{
		"cat1": apitype.NewCategorizedImage(apitype.NewCategory("cat1", "cat_1", ""), apitype.MOVE),
	}
	handle := apitype.NewHandle("filepath", "filename")
	command := apitype.NewPersistCategorizationCommand(false, true, 100)
	operations, err := sut.ResolveOperationsForGroup(handle, imageCategories, command)

	a.Nil(err)
	ops := operations.GetOperations()
	a.Equal(3, len(ops))
	a.Equal("Exif Rotate", ops[0].String())
	a.Equal(fmt.Sprintf("Copy file 'filename' to '%s'", filepath.Join("filepath", "cat_1")), ops[1].String())
	a.Equal("Remove", ops[2].String())
}
