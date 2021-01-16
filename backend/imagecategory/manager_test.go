package imagecategory

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"image"
	"os"
	"path/filepath"
	"testing"
	"time"
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

func (s *MockSender) SendToTopic(topic api.Topic) {
	s.Called(topic)
}

func (s *MockSender) SendCommandToTopic(topic api.Topic, command apitype.Command) {
	s.Called(topic, command)
}

func (s *MockSender) SendError(message string, err error) {
}

func (s *MockImageLoader) LoadImage(apitype.ImageId) (image.Image, error) {
	return nil, nil
}

func (s *MockImageLoader) LoadExifData(apitype.ImageId) (*apitype.ExifData, error) {
	return nil, nil
}

//// Basic cases

func TestCategorizeOne(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", api.ImageRequestNext).Return()
	sender.On("SendCommandToTopic", api.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.NewFilterManager()
	imageLoader := new(MockImageLoader)
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, imageCategoryStore)

	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("/tmp", "foo"))
	cat1, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 1", "c1", "C"))
	cmd := api.CategorizeCommand{
		ImageId:    imageFile.GetImageId(),
		CategoryId: cat1.GetId(),
		Operation:  apitype.MOVE,
	}
	sut.SetCategory(&cmd)

	result := sut.GetCategories(&api.ImageCategoryQuery{ImageId: imageFile.GetImageId()})

	if a.Equal(1, len(result)) {
		a.Equal("Cat 1", result[1].Category.GetName())
	}
}

func TestCategorizeOneToTwoCategories(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", api.ImageRequestNext).Return()
	sender.On("SendCommandToTopic", api.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.NewFilterManager()
	imageLoader := new(MockImageLoader)
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, imageCategoryStore)

	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("/tmp", "foo"))
	cat1, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 1", "c1", "C"))
	cat2, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 2", "c2", "D"))
	cmd1 := &api.CategorizeCommand{ImageId: imageFile.GetImageId(), CategoryId: cat1.GetId(), Operation: apitype.MOVE}
	cmd2 := &api.CategorizeCommand{ImageId: imageFile.GetImageId(), CategoryId: cat2.GetId(), Operation: apitype.MOVE}
	sut.SetCategory(cmd1)
	sut.SetCategory(cmd2)

	result := sut.GetCategories(&api.ImageCategoryQuery{ImageId: imageFile.GetImageId()})

	if a.Equal(2, len(result)) {
		a.Equal("Cat 1", result[1].Category.GetName())
		a.Equal("Cat 2", result[2].Category.GetName())
	}
}

func TestCategorizeOneRemoveCategory(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", api.ImageRequestNext).Return()
	sender.On("SendCommandToTopic", api.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.NewFilterManager()
	imageLoader := new(MockImageLoader)
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, imageCategoryStore)

	cat1, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 1", "c1", "C"))
	cat2, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 2", "c2", "D"))
	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("/tmp", "foo"))
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.GetImageId(), CategoryId: cat1.GetId(), Operation: apitype.MOVE})
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.GetImageId(), CategoryId: cat2.GetId(), Operation: apitype.MOVE})
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.GetImageId(), CategoryId: cat1.GetId(), Operation: apitype.NONE})

	result := sut.GetCategories(&api.ImageCategoryQuery{ImageId: imageFile.GetImageId()})

	if a.Equal(1, len(result)) {
		a.Equal("Cat 2", result[2].Category.GetName())
	}
}

func TestCategorizeOneRemoveAll(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", api.ImageRequestNext).Return()
	sender.On("SendCommandToTopic", api.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.NewFilterManager()
	imageLoader := new(MockImageLoader)
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, imageCategoryStore)

	cat1, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 1", "c1", "C"))
	cat2, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 2", "c2", "D"))
	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("/tmp", "foo"))
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.GetImageId(), CategoryId: cat1.GetId(), Operation: apitype.MOVE})
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.GetImageId(), CategoryId: cat2.GetId(), Operation: apitype.MOVE})
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.GetImageId(), CategoryId: cat1.GetId(), Operation: apitype.NONE})
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.GetImageId(), CategoryId: cat2.GetId(), Operation: apitype.NONE})

	result := sut.GetCategories(&api.ImageCategoryQuery{ImageId: imageFile.GetImageId()})

	a.Equal(0, len(result))
}

//// Force category

func TestCategorizeForceToCategory(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", api.ImageRequestNext).Return()
	sender.On("SendCommandToTopic", api.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.NewFilterManager()
	imageLoader := new(MockImageLoader)
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, imageCategoryStore)

	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("/tmp", "foo"))
	cat1, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 1", "c1", "C"))
	cat2, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 2", "c2", "D"))
	cat3, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 3", "c3", "E"))
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.GetImageId(), CategoryId: cat1.GetId(), Operation: apitype.MOVE})
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.GetImageId(), CategoryId: cat2.GetId(), Operation: apitype.MOVE})
	command := &api.CategorizeCommand{
		ImageId:         imageFile.GetImageId(),
		CategoryId:      cat3.GetId(),
		Operation:       apitype.MOVE,
		ForceToCategory: true,
	}
	sut.SetCategory(command)

	result := sut.GetCategories(&api.ImageCategoryQuery{ImageId: imageFile.GetImageId()})

	a.Equal(1, len(result))
	if a.NotNil(result[3]) {
		a.Equal("Cat 3", result[3].Category.GetName())
	}
}

func TestCategorizeForceToExistingCategory(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", api.ImageRequestNext).Return()
	sender.On("SendCommandToTopic", api.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.NewFilterManager()
	imageLoader := new(MockImageLoader)
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, imageCategoryStore)

	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("/tmp", "foo"))
	cat1, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 1", "c1", "C"))
	cat2, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 2", "c2", "D"))
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.GetImageId(), CategoryId: cat1.GetId(), Operation: apitype.MOVE})
	command := &api.CategorizeCommand{ImageId: imageFile.GetImageId(), CategoryId: cat2.GetId(), Operation: apitype.MOVE, ForceToCategory: true}
	sut.SetCategory(command)

	result := sut.GetCategories(&api.ImageCategoryQuery{ImageId: imageFile.GetImageId()})

	if a.Equal(1, len(result)) {
		a.Equal("Cat 2", result[2].Category.GetName())
	}
}

func TestCategorizeForceToCategory_None(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", api.ImageRequestNext).Return()
	sender.On("SendCommandToTopic", api.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterManager := filter.NewFilterManager()
	imageLoader := new(MockImageLoader)
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, imageCategoryStore)

	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("/tmp", "foo"))
	cat1, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 1", "c1", "C"))
	cat2, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 2", "c2", "D"))
	cat3, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 3", "c3", "E"))
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.GetImageId(), CategoryId: cat1.GetId(), Operation: apitype.MOVE})
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.GetImageId(), CategoryId: cat2.GetId(), Operation: apitype.MOVE})
	command := &api.CategorizeCommand{ImageId: imageFile.GetImageId(), CategoryId: cat3.GetId(), Operation: apitype.NONE, ForceToCategory: true}
	sut.SetCategory(command)

	result := sut.GetCategories(&api.ImageCategoryQuery{ImageId: imageFile.GetImageId()})

	a.Equal(0, len(result))
}

func TestResolveFileOperations(t *testing.T) {
	a := require.New(t)

	sender := new(MockSender)
	imageCache := new(MockImageCache)
	imageLoader := new(MockImageLoader)
	imageLoader.On("LoadImage", api.ImageRequestNext).Return(nil, nil)
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)
	lib := library.NewLibrary(sender, imageCache, imageLoader, nil, imageStore)
	filterManager := filter.NewFilterManager()

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, imageCategoryStore)
	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("filepath", "filename"))
	lib.AddImageFiles([]*apitype.ImageFile{imageFile.GetImageFile()})

	var imageCategories = map[apitype.ImageId]map[apitype.CategoryId]*api.CategorizedImage{
		1: {
			1: &api.CategorizedImage{
				Category:  apitype.NewCategory("cat1", "cat_1", ""),
				Operation: apitype.MOVE,
			},
		},
	}
	command := &api.PersistCategorizationCommand{
		KeepOriginals:  true,
		FixOrientation: false,
		Quality:        100,
	}
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
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)
	lib := library.NewLibrary(sender, imageCache, imageLoader, nil, imageStore)
	filterManager := filter.NewFilterManager()

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, imageCategoryStore)

	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("filepath", "filename"))
	cat, _ := categoryStore.AddCategory(apitype.NewCategory("cat1", "cat_1", ""))
	_ = imageCategoryStore.CategorizeImage(imageFile.GetImageId(), cat.GetId(), apitype.MOVE)
	imageCategories, _ := imageCategoryStore.GetCategorizedImages()

	command := &api.PersistCategorizationCommand{
		KeepOriginals:  true,
		FixOrientation: false,
		Quality:        100,
	}
	operations, err := sut.ResolveOperationsForGroup(imageFile, imageCategories[imageFile.GetImageId()], command)

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
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)
	lib := library.NewLibrary(sender, imageCache, imageLoader, nil, imageStore)
	filterManager := filter.NewFilterManager()

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, imageCategoryStore)

	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("filepath", "filename"))
	cat, _ := categoryStore.AddCategory(apitype.NewCategory("cat1", "cat_1", ""))
	_ = imageCategoryStore.CategorizeImage(imageFile.GetImageId(), cat.GetId(), apitype.MOVE)
	imageCategories, _ := imageCategoryStore.GetCategorizedImages()

	command := &api.PersistCategorizationCommand{
		KeepOriginals:  false,
		FixOrientation: false,
		Quality:        100,
	}
	operations, err := sut.ResolveOperationsForGroup(imageFile, imageCategories[imageFile.GetImageId()], command)

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
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)
	lib := library.NewLibrary(sender, imageCache, imageLoader, nil, imageStore)
	filterManager := filter.NewFilterManager()

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, imageCategoryStore)

	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("filepath", "filename"))
	cat, _ := categoryStore.AddCategory(apitype.NewCategory("cat1", "cat_1", ""))
	_ = imageCategoryStore.CategorizeImage(imageFile.GetImageId(), cat.GetId(), apitype.MOVE)
	imageCategories, _ := imageCategoryStore.GetCategorizedImages()

	command := &api.PersistCategorizationCommand{
		KeepOriginals:  true,
		FixOrientation: true,
		Quality:        100,
	}
	operations, err := sut.ResolveOperationsForGroup(imageFile, imageCategories[imageFile.GetImageId()], command)

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
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)
	lib := library.NewLibrary(sender, imageCache, imageLoader, nil, imageStore)
	filterManager := filter.NewFilterManager()

	sut := NewImageCategoryManager(sender, lib, filterManager, imageLoader, imageCategoryStore)

	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("filepath", "filename"))
	cat, _ := categoryStore.AddCategory(apitype.NewCategory("cat1", "cat_1", ""))
	_ = imageCategoryStore.CategorizeImage(imageFile.GetImageId(), cat.GetId(), apitype.MOVE)
	imageCategories, _ := imageCategoryStore.GetCategorizedImages()

	command := &api.PersistCategorizationCommand{
		KeepOriginals:  false,
		FixOrientation: true,
		Quality:        100,
	}
	operations, err := sut.ResolveOperationsForGroup(imageFile, imageCategories[imageFile.GetImageId()], command)

	a.Nil(err)
	ops := operations.GetOperations()
	a.Equal(3, len(ops))
	a.Equal("Exif Rotate", ops[0].String())
	a.Equal(fmt.Sprintf("Copy file 'filename' to '%s'", filepath.Join("filepath", "cat_1")), ops[1].String())
	a.Equal("Remove", ops[2].String())
}
