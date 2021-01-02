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
	"vincit.fi/image-sorter/backend/database"
	"vincit.fi/image-sorter/backend/filter"
	"vincit.fi/image-sorter/backend/library"
)

type MockSender struct {
	api.Sender
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

func (s *MockSender) SendToTopic(topic api.Topic) {
	s.Called(topic)
}

func (s *MockSender) SendToTopicWithData(topic api.Topic, data ...interface{}) {
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
	sender.On("SendToTopic", api.ImageRequestNext).Return()
	sender.On("SendToTopicWithData", api.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.NewFilterManager()
	imageLoader := new(MockImageLoader)
	store := database.NewInMemoryStore()

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, store)

	handle, _ := store.AddImage(apitype.NewHandle(-1, "/tmp", "foo"))
	cat1, _ := store.AddCategory(apitype.NewCategory(-1, "Cat 1", "c1", "C"))
	cmd := apitype.NewCategorizeCommand(handle, cat1, apitype.MOVE)
	sut.SetCategory(cmd)

	result := sut.GetCategories(handle)

	if a.Equal(1, len(result)) {
		a.Equal("Cat 1", result[1].GetEntry().GetName())
	}
}

func TestCategorizeOneToTwoCategories(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", api.ImageRequestNext).Return()
	sender.On("SendToTopicWithData", api.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.NewFilterManager()
	imageLoader := new(MockImageLoader)
	store := database.NewInMemoryStore()

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, store)

	handle, _ := store.AddImage(apitype.NewHandle(-1, "/tmp", "foo"))
	cat1, _ := store.AddCategory(apitype.NewCategory(-1, "Cat 1", "c1", "C"))
	cat2, _ := store.AddCategory(apitype.NewCategory(-1, "Cat 2", "c2", "D"))
	cmd1 := apitype.NewCategorizeCommand(handle, cat1, apitype.MOVE)
	cmd2 := apitype.NewCategorizeCommand(handle, cat2, apitype.MOVE)
	sut.SetCategory(cmd1)
	sut.SetCategory(cmd2)

	result := sut.GetCategories(handle)

	if a.Equal(2, len(result)) {
		a.Equal("Cat 1", result[1].GetEntry().GetName())
		a.Equal("Cat 2", result[2].GetEntry().GetName())
	}
}

func TestCategorizeOneRemoveCategory(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", api.ImageRequestNext).Return()
	sender.On("SendToTopicWithData", api.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.NewFilterManager()
	imageLoader := new(MockImageLoader)
	store := database.NewInMemoryStore()

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, store)

	cat1, _ := store.AddCategory(apitype.NewCategory(-1, "Cat 1", "c1", "C"))
	cat2, _ := store.AddCategory(apitype.NewCategory(-1, "Cat 2", "c2", "D"))
	handle, _ := store.AddImage(apitype.NewHandle(-1, "/tmp", "foo"))
	sut.SetCategory(apitype.NewCategorizeCommand(handle, cat1, apitype.MOVE))
	sut.SetCategory(apitype.NewCategorizeCommand(handle, cat2, apitype.MOVE))
	sut.SetCategory(apitype.NewCategorizeCommand(handle, cat1, apitype.NONE))

	result := sut.GetCategories(handle)

	if a.Equal(1, len(result)) {
		a.Equal("Cat 2", result[2].GetEntry().GetName())
	}
}

func TestCategorizeOneRemoveAll(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", api.ImageRequestNext).Return()
	sender.On("SendToTopicWithData", api.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.NewFilterManager()
	imageLoader := new(MockImageLoader)
	store := database.NewInMemoryStore()

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, store)

	cat1, _ := store.AddCategory(apitype.NewCategory(-1, "Cat 1", "c1", "C"))
	cat2, _ := store.AddCategory(apitype.NewCategory(-1, "Cat 2", "c2", "D"))
	handle, _ := store.AddImage(apitype.NewHandle(-1, "/tmp", "foo"))
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
	sender.On("SendToTopic", api.ImageRequestNext).Return()
	sender.On("SendToTopicWithData", api.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.NewFilterManager()
	imageLoader := new(MockImageLoader)
	store := database.NewInMemoryStore()

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, store)

	handle, _ := store.AddImage(apitype.NewHandle(-1, "/tmp", "foo"))
	cat1, _ := store.AddCategory(apitype.NewCategory(-1, "Cat 1", "c1", "C"))
	cat2, _ := store.AddCategory(apitype.NewCategory(-1, "Cat 2", "c2", "D"))
	cat3, _ := store.AddCategory(apitype.NewCategory(-1, "Cat 3", "c3", "E"))
	sut.SetCategory(apitype.NewCategorizeCommand(handle, cat1, apitype.MOVE))
	sut.SetCategory(apitype.NewCategorizeCommand(handle, cat2, apitype.MOVE))
	command := apitype.NewCategorizeCommand(handle, cat3, apitype.MOVE)
	command.SetForceToCategory(true)
	sut.SetCategory(command)

	result := sut.GetCategories(handle)

	a.Equal(1, len(result))
	if a.NotNil(result[3]) {
		a.Equal("Cat 3", result[3].GetEntry().GetName())
	}
}

func TestCategorizeForceToExistingCategory(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", api.ImageRequestNext).Return()
	sender.On("SendToTopicWithData", api.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.NewFilterManager()
	imageLoader := new(MockImageLoader)
	store := database.NewInMemoryStore()

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, store)

	handle, _ := store.AddImage(apitype.NewHandle(-1, "/tmp", "foo"))
	cat1, _ := store.AddCategory(apitype.NewCategory(-1, "Cat 1", "c1", "C"))
	cat2, _ := store.AddCategory(apitype.NewCategory(-1, "Cat 2", "c2", "D"))
	sut.SetCategory(apitype.NewCategorizeCommand(handle, cat1, apitype.MOVE))
	command := apitype.NewCategorizeCommand(handle, cat2, apitype.MOVE)
	command.SetForceToCategory(true)
	sut.SetCategory(command)

	result := sut.GetCategories(handle)

	if a.Equal(1, len(result)) {
		a.Equal("Cat 2", result[2].GetEntry().GetName())
	}
}

func TestCategorizeForceToCategory_None(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", api.ImageRequestNext).Return()
	sender.On("SendToTopicWithData", api.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.NewFilterManager()
	imageLoader := new(MockImageLoader)
	store := database.NewInMemoryStore()

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, store)

	handle, _ := store.AddImage(apitype.NewHandle(-1, "/tmp", "foo"))
	cat1, _ := store.AddCategory(apitype.NewCategory(-1, "Cat 1", "c1", "C"))
	cat2, _ := store.AddCategory(apitype.NewCategory(-1, "Cat 2", "c2", "D"))
	cat3, _ := store.AddCategory(apitype.NewCategory(-1, "Cat 3", "c3", "E"))
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
	imageLoader.On("LoadImage", api.ImageRequestNext).Return(nil, nil)
	store := database.NewInMemoryStore()
	lib := library.NewLibrary(sender, imageCache, imageLoader, store)
	filterManager := filter.NewFilterManager()

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, store)
	handle, _ := store.AddImage(apitype.NewHandle(-1, "filepath", "filename"))
	lib.AddHandles([]*apitype.Handle{handle})

	var imageCategories = map[int64]map[int64]*apitype.CategorizedImage{
		1: {
			1: apitype.NewCategorizedImage(apitype.NewCategory(-1, "cat1", "cat_1", ""), apitype.MOVE),
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
	imageLoader.On("LoadImage", api.ImageRequestNext).Return(nil, nil)
	store := database.NewInMemoryStore()
	lib := library.NewLibrary(sender, imageCache, imageLoader, store)
	filterManager := filter.NewFilterManager()

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, store)

	handle, _ := store.AddImage(apitype.NewHandle(-1, "filepath", "filename"))
	cat, _ := store.AddCategory(apitype.NewCategory(-1, "cat1", "cat_1", ""))
	_ = store.CategorizeImage(handle.GetId(), cat.GetId(), apitype.MOVE)
	imageCategories, _ := store.GetCategorizedImages()

	command := apitype.NewPersistCategorizationCommand(true, false, 100)
	operations, err := sut.ResolveOperationsForGroup(handle, imageCategories[handle.GetId()], command)

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
	imageLoader.On("LoadImage", api.ImageRequestNext).Return(nil, nil)
	store := database.NewInMemoryStore()
	lib := library.NewLibrary(sender, imageCache, imageLoader, store)
	filterManager := filter.NewFilterManager()

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, store)

	handle, _ := store.AddImage(apitype.NewHandle(-1, "filepath", "filename"))
	cat, _ := store.AddCategory(apitype.NewCategory(-1, "cat1", "cat_1", ""))
	_ = store.CategorizeImage(handle.GetId(), cat.GetId(), apitype.MOVE)
	imageCategories, _ := store.GetCategorizedImages()

	command := apitype.NewPersistCategorizationCommand(false, false, 100)
	operations, err := sut.ResolveOperationsForGroup(handle, imageCategories[handle.GetId()], command)

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
	imageLoader.On("LoadImage", api.ImageRequestNext).Return(nil, nil)
	store := database.NewInMemoryStore()
	lib := library.NewLibrary(sender, imageCache, imageLoader, store)
	filterManager := filter.NewFilterManager()

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, store)

	handle, _ := store.AddImage(apitype.NewHandle(-1, "filepath", "filename"))
	cat, _ := store.AddCategory(apitype.NewCategory(-1, "cat1", "cat_1", ""))
	_ = store.CategorizeImage(handle.GetId(), cat.GetId(), apitype.MOVE)
	imageCategories, _ := store.GetCategorizedImages()

	command := apitype.NewPersistCategorizationCommand(true, true, 100)
	operations, err := sut.ResolveOperationsForGroup(handle, imageCategories[handle.GetId()], command)

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
	imageLoader.On("LoadImage", api.ImageRequestNext).Return(nil, nil)
	store := database.NewInMemoryStore()
	lib := library.NewLibrary(sender, imageCache, imageLoader, store)
	filterManager := filter.NewFilterManager()

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, store)

	handle, _ := store.AddImage(apitype.NewHandle(-1, "filepath", "filename"))
	cat, _ := store.AddCategory(apitype.NewCategory(-1, "cat1", "cat_1", ""))
	_ = store.CategorizeImage(handle.GetId(), cat.GetId(), apitype.MOVE)
	imageCategories, _ := store.GetCategorizedImages()

	command := apitype.NewPersistCategorizationCommand(false, true, 100)
	operations, err := sut.ResolveOperationsForGroup(handle, imageCategories[handle.GetId()], command)

	a.Nil(err)
	ops := operations.GetOperations()
	a.Equal(3, len(ops))
	a.Equal("Exif Rotate", ops[0].String())
	a.Equal(fmt.Sprintf("Copy file 'filename' to '%s'", filepath.Join("filepath", "cat_1")), ops[1].String())
	a.Equal("Remove", ops[2].String())
}
