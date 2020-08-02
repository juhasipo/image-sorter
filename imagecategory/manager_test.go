package imagecategory

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"image"
	"path/filepath"
	"testing"
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/filter"
	"vincit.fi/image-sorter/imageloader"
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
	imageloader.ImageStore
	mock.Mock
}

type MockImageLoader struct {
	imageloader.ImageLoader
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
	sender.On("SendToTopic", event.ImageRequestNext).Return()
	sender.On("SendToTopicWithData", event.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.NewFilterManager()
	imageLoader := new(MockImageLoader)

	sut := NewManager(sender, lib, filterManager, imageLoader)

	cat1 := common.NewCategory("Cat 1", "c1", "C")
	handle := common.NewHandle("/tmp", "foo")
	cmd := category.NewCategorizeCommand(handle, cat1, common.MOVE)
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

	sut := NewManager(sender, lib, filterManager, imageLoader)

	cat1 := common.NewCategory("Cat 1", "c1", "C")
	cat2 := common.NewCategory("Cat 2", "c2", "D")
	handle := common.NewHandle("/tmp", "foo")
	cmd1 := category.NewCategorizeCommand(handle, cat1, common.MOVE)
	cmd2 := category.NewCategorizeCommand(handle, cat2, common.MOVE)
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

	sut := NewManager(sender, lib, filterManager, imageLoader)

	cat1 := common.NewCategory("Cat 1", "c1", "C")
	cat2 := common.NewCategory("Cat 2", "c2", "D")
	handle := common.NewHandle("/tmp", "foo")
	sut.SetCategory(category.NewCategorizeCommand(handle, cat1, common.MOVE))
	sut.SetCategory(category.NewCategorizeCommand(handle, cat2, common.MOVE))
	sut.SetCategory(category.NewCategorizeCommand(handle, cat1, common.NONE))

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

	sut := NewManager(sender, lib, filterManager, imageLoader)

	cat1 := common.NewCategory("Cat 1", "c1", "C")
	cat2 := common.NewCategory("Cat 2", "c2", "D")
	handle := common.NewHandle("/tmp", "foo")
	sut.SetCategory(category.NewCategorizeCommand(handle, cat1, common.MOVE))
	sut.SetCategory(category.NewCategorizeCommand(handle, cat2, common.MOVE))
	sut.SetCategory(category.NewCategorizeCommand(handle, cat1, common.NONE))
	sut.SetCategory(category.NewCategorizeCommand(handle, cat2, common.NONE))

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

	sut := NewManager(sender, lib, filterManager, imageLoader)

	cat1 := common.NewCategory("Cat 1", "c1", "C")
	cat2 := common.NewCategory("Cat 2", "c2", "D")
	cat3 := common.NewCategory("Cat 3", "c3", "E")
	handle := common.NewHandle("/tmp", "foo")
	sut.SetCategory(category.NewCategorizeCommand(handle, cat1, common.MOVE))
	sut.SetCategory(category.NewCategorizeCommand(handle, cat2, common.MOVE))
	command := category.NewCategorizeCommand(handle, cat3, common.MOVE)
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

	sut := NewManager(sender, lib, filterManager, imageLoader)

	cat1 := common.NewCategory("Cat 1", "c1", "C")
	cat2 := common.NewCategory("Cat 2", "c2", "D")
	handle := common.NewHandle("/tmp", "foo")
	sut.SetCategory(category.NewCategorizeCommand(handle, cat1, common.MOVE))
	command := category.NewCategorizeCommand(handle, cat2, common.MOVE)
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

	sut := NewManager(sender, lib, filterManager, imageLoader)

	cat1 := common.NewCategory("Cat 1", "c1", "C")
	cat2 := common.NewCategory("Cat 2", "c2", "D")
	cat3 := common.NewCategory("Cat 3", "c3", "E")
	handle := common.NewHandle("/tmp", "foo")
	sut.SetCategory(category.NewCategorizeCommand(handle, cat1, common.MOVE))
	sut.SetCategory(category.NewCategorizeCommand(handle, cat2, common.MOVE))
	command := category.NewCategorizeCommand(handle, cat3, common.NONE)
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

	sut := NewManager(sender, lib, filterManager, imageLoader)
	handle := common.NewHandle("filepath", "filename")
	lib.AddHandles([]*common.Handle{handle})

	var imageCategories = map[string]map[string]*category.CategorizedImage{
		"filename": {
			"cat1": category.NewCategorizedImage(common.NewCategory("cat1", "cat_1", ""), common.MOVE),
		},
	}
	command := common.NewPersistCategorizationCommand(true, false, 100)
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

	sut := NewManager(sender, lib, filterManager, imageLoader)

	var imageCategories = map[string]*category.CategorizedImage{
		"cat1": category.NewCategorizedImage(common.NewCategory("cat1", "cat_1", ""), common.MOVE),
	}
	handle := common.NewHandle("filepath", "filename")
	command := common.NewPersistCategorizationCommand(true, false, 100)
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

	sut := NewManager(sender, lib, filterManager, imageLoader)

	var imageCategories = map[string]*category.CategorizedImage{
		"cat1": category.NewCategorizedImage(common.NewCategory("cat1", "cat_1", ""), common.MOVE),
	}
	handle := common.NewHandle("filepath", "filename")
	command := common.NewPersistCategorizationCommand(false, false, 100)
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

	sut := NewManager(sender, lib, filterManager, imageLoader)

	var imageCategories = map[string]*category.CategorizedImage{
		"cat1": category.NewCategorizedImage(common.NewCategory("cat1", "cat_1", ""), common.MOVE),
	}
	handle := common.NewHandle("filepath", "filename")
	command := common.NewPersistCategorizationCommand(true, true, 100)
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

	sut := NewManager(sender, lib, filterManager, imageLoader)

	var imageCategories = map[string]*category.CategorizedImage{
		"cat1": category.NewCategorizedImage(common.NewCategory("cat1", "cat_1", ""), common.MOVE),
	}
	handle := common.NewHandle("filepath", "filename")
	command := common.NewPersistCategorizationCommand(false, true, 100)
	operations, err := sut.ResolveOperationsForGroup(handle, imageCategories, command)

	a.Nil(err)
	ops := operations.GetOperations()
	a.Equal(3, len(ops))
	a.Equal("Exif Rotate", ops[0].String())
	a.Equal(fmt.Sprintf("Copy file 'filename' to '%s'", filepath.Join("filepath", "cat_1")), ops[1].String())
	a.Equal("Remove", ops[2].String())
}
