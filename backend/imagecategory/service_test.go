package imagecategory

import (
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
	api.ImageService
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

func (s *MockImageLoader) LoadExifData(*apitype.ImageFile) (*apitype.ExifData, error) {
	return apitype.NewInvalidExifData(), nil
}

type StubProgressReporter struct {
	api.ProgressReporter
}

func (s StubProgressReporter) Update(name string, current int, total int, canCancel bool) {
}

func (s StubProgressReporter) Error(error string, err error) {
}

//// Basic cases

func TestCategorizeOne(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", api.ImageRequestNext).Return()
	sender.On("SendCommandToTopic", api.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterService := filter.NewFilterService()
	imageLoader := new(MockImageLoader)
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)

	sut := NewImageCategoryService(sender, lib, filterService, imageLoader, imageCategoryStore)

	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("/tmp", "foo"))
	cat1, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 1", "c1", "C"))
	cmd := api.CategorizeCommand{
		ImageId:    imageFile.Id(),
		CategoryId: cat1.Id(),
		Operation:  apitype.CATEGORIZE,
	}
	sut.SetCategory(&cmd)

	result := sut.GetCategories(&api.ImageCategoryQuery{ImageId: imageFile.Id()})

	if a.Equal(1, len(result)) {
		a.Equal("Cat 1", result[1].Category.Name())
	}
}

func TestCategorizeOne_InvalidImageId(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", api.ImageRequestNext).Return()
	sender.On("SendCommandToTopic", api.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterService := filter.NewFilterService()
	imageLoader := new(MockImageLoader)
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)

	sut := NewImageCategoryService(sender, lib, filterService, imageLoader, imageCategoryStore)

	_, _ = imageStore.AddImage(apitype.NewImageFile("/tmp", "foo"))
	cat1, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 1", "c1", "C"))
	cmd := api.CategorizeCommand{
		ImageId:    0,
		CategoryId: cat1.Id(),
		Operation:  apitype.CATEGORIZE,
	}
	sut.SetCategory(&cmd)

	result := sut.GetCategories(&api.ImageCategoryQuery{ImageId: 0})

	a.Equal(0, len(result))
}

func TestCategorizeOne_InvalidCategoryId(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", api.ImageRequestNext).Return()
	sender.On("SendCommandToTopic", api.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterService := filter.NewFilterService()
	imageLoader := new(MockImageLoader)
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)

	sut := NewImageCategoryService(sender, lib, filterService, imageLoader, imageCategoryStore)

	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("/tmp", "foo"))
	_, _ = categoryStore.AddCategory(apitype.NewCategory("Cat 1", "c1", "C"))
	cmd := api.CategorizeCommand{
		ImageId:    imageFile.Id(),
		CategoryId: 0,
		Operation:  apitype.CATEGORIZE,
	}
	sut.SetCategory(&cmd)

	result := sut.GetCategories(&api.ImageCategoryQuery{ImageId: imageFile.Id()})

	a.Equal(0, len(result))
}

func TestCategorizeOneToTwoCategories(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", api.ImageRequestNext).Return()
	sender.On("SendCommandToTopic", api.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterService := filter.NewFilterService()
	imageLoader := new(MockImageLoader)
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)

	sut := NewImageCategoryService(sender, lib, filterService, imageLoader, imageCategoryStore)

	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("/tmp", "foo"))
	cat1, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 1", "c1", "C"))
	cat2, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 2", "c2", "D"))
	cmd1 := &api.CategorizeCommand{ImageId: imageFile.Id(), CategoryId: cat1.Id(), Operation: apitype.CATEGORIZE}
	cmd2 := &api.CategorizeCommand{ImageId: imageFile.Id(), CategoryId: cat2.Id(), Operation: apitype.CATEGORIZE}
	sut.SetCategory(cmd1)
	sut.SetCategory(cmd2)

	result := sut.GetCategories(&api.ImageCategoryQuery{ImageId: imageFile.Id()})

	if a.Equal(2, len(result)) {
		a.Equal("Cat 1", result[1].Category.Name())
		a.Equal("Cat 2", result[2].Category.Name())
	}
}

func TestCategorizeOneRemoveCategory(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", api.ImageRequestNext).Return()
	sender.On("SendCommandToTopic", api.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterService := filter.NewFilterService()
	imageLoader := new(MockImageLoader)
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)

	sut := NewImageCategoryService(sender, lib, filterService, imageLoader, imageCategoryStore)

	cat1, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 1", "c1", "C"))
	cat2, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 2", "c2", "D"))
	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("/tmp", "foo"))
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.Id(), CategoryId: cat1.Id(), Operation: apitype.CATEGORIZE})
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.Id(), CategoryId: cat2.Id(), Operation: apitype.CATEGORIZE})
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.Id(), CategoryId: cat1.Id(), Operation: apitype.UNCATEGORIZE})

	result := sut.GetCategories(&api.ImageCategoryQuery{ImageId: imageFile.Id()})

	if a.Equal(1, len(result)) {
		a.Equal("Cat 2", result[2].Category.Name())
	}
}

func TestCategorizeOneRemoveAll(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", api.ImageRequestNext).Return()
	sender.On("SendCommandToTopic", api.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterService := filter.NewFilterService()
	imageLoader := new(MockImageLoader)
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)

	sut := NewImageCategoryService(sender, lib, filterService, imageLoader, imageCategoryStore)

	cat1, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 1", "c1", "C"))
	cat2, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 2", "c2", "D"))
	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("/tmp", "foo"))
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.Id(), CategoryId: cat1.Id(), Operation: apitype.CATEGORIZE})
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.Id(), CategoryId: cat2.Id(), Operation: apitype.CATEGORIZE})
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.Id(), CategoryId: cat1.Id(), Operation: apitype.UNCATEGORIZE})
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.Id(), CategoryId: cat2.Id(), Operation: apitype.UNCATEGORIZE})

	result := sut.GetCategories(&api.ImageCategoryQuery{ImageId: imageFile.Id()})

	a.Equal(0, len(result))
}

//// Force category

func TestCategorizeForceToCategory(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", api.ImageRequestNext).Return()
	sender.On("SendCommandToTopic", api.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterService := filter.NewFilterService()
	imageLoader := new(MockImageLoader)
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)

	sut := NewImageCategoryService(sender, lib, filterService, imageLoader, imageCategoryStore)

	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("/tmp", "foo"))
	cat1, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 1", "c1", "C"))
	cat2, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 2", "c2", "D"))
	cat3, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 3", "c3", "E"))
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.Id(), CategoryId: cat1.Id(), Operation: apitype.CATEGORIZE})
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.Id(), CategoryId: cat2.Id(), Operation: apitype.CATEGORIZE})
	command := &api.CategorizeCommand{
		ImageId:         imageFile.Id(),
		CategoryId:      cat3.Id(),
		Operation:       apitype.CATEGORIZE,
		ForceToCategory: true,
	}
	sut.SetCategory(command)

	result := sut.GetCategories(&api.ImageCategoryQuery{ImageId: imageFile.Id()})

	a.Equal(1, len(result))
	if a.NotNil(result[3]) {
		a.Equal("Cat 3", result[3].Category.Name())
	}
}

func TestCategorizeForceToExistingCategory(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", api.ImageRequestNext).Return()
	sender.On("SendCommandToTopic", api.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterService := filter.NewFilterService()
	imageLoader := new(MockImageLoader)
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)

	sut := NewImageCategoryService(sender, lib, filterService, imageLoader, imageCategoryStore)

	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("/tmp", "foo"))
	cat1, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 1", "c1", "C"))
	cat2, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 2", "c2", "D"))
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.Id(), CategoryId: cat1.Id(), Operation: apitype.CATEGORIZE})
	command := &api.CategorizeCommand{ImageId: imageFile.Id(), CategoryId: cat2.Id(), Operation: apitype.CATEGORIZE, ForceToCategory: true}
	sut.SetCategory(command)

	result := sut.GetCategories(&api.ImageCategoryQuery{ImageId: imageFile.Id()})

	if a.Equal(1, len(result)) {
		a.Equal("Cat 2", result[2].Category.Name())
	}
}

func TestCategorizeForceToCategory_None(t *testing.T) {
	a := assert.New(t)

	sender := new(MockSender)
	sender.On("SendToTopic", api.ImageRequestNext).Return()
	sender.On("SendCommandToTopic", api.CategoryImageUpdate, mock.Anything).Return()
	lib := new(MockLibrary)
	filterService := filter.NewFilterService()
	imageLoader := new(MockImageLoader)
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)

	sut := NewImageCategoryService(sender, lib, filterService, imageLoader, imageCategoryStore)

	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("/tmp", "foo"))
	cat1, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 1", "c1", "C"))
	cat2, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 2", "c2", "D"))
	cat3, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 3", "c3", "E"))
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.Id(), CategoryId: cat1.Id(), Operation: apitype.CATEGORIZE})
	sut.SetCategory(&api.CategorizeCommand{ImageId: imageFile.Id(), CategoryId: cat2.Id(), Operation: apitype.CATEGORIZE})
	command := &api.CategorizeCommand{ImageId: imageFile.Id(), CategoryId: cat3.Id(), Operation: apitype.UNCATEGORIZE, ForceToCategory: true}
	sut.SetCategory(command)

	result := sut.GetCategories(&api.ImageCategoryQuery{ImageId: imageFile.Id()})

	a.Equal(0, len(result))
}

func TestResolveFileOperations(t *testing.T) {
	a := require.New(t)

	sender := new(MockSender)
	imageCache := new(MockImageCache)
	imageLoader := new(MockImageLoader)
	imageLoader.On("LoadImage", api.ImageRequestNext).Return(nil, nil)
	sender.On("SendCommandToTopic", mock.Anything, mock.Anything)
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	imageMetaDataStore := database.NewImageMetaDataStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)
	statusStore := database.NewStatusStore(memoryDatabase)
	lib := library.NewImageService(
		sender,
		library.NewImageLibrary(imageCache, imageLoader, nil, imageStore, imageMetaDataStore, StubProgressReporter{}),
		statusStore,
	)
	filterService := filter.NewFilterService()

	sut := NewImageCategoryService(sender, lib, filterService, imageLoader, imageCategoryStore)
	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("filepath", "filename"))
	lib.AddImageFiles([]*apitype.ImageFile{imageFile})

	var imageCategories = map[apitype.ImageId]map[apitype.CategoryId]*api.CategorizedImage{
		1: {
			1: &api.CategorizedImage{
				Category:  apitype.NewCategory("cat1", "cat_1", ""),
				Operation: apitype.CATEGORIZE,
			},
		},
	}
	command := &api.PersistCategorizationCommand{
		KeepOriginals:  true,
		FixOrientation: false,
		Quality:        100,
	}
	operations := sut.ResolveFileOperations(imageCategories, command, func(int, int) {})

	a.Equal(1, len(operations))

	ops := operations[0]
	a.Equal(1, len(ops.Operations()))
}

func TestResolveOperationsForGroup_KeepOld(t *testing.T) {
	a := require.New(t)

	sender := new(MockSender)
	imageCache := new(MockImageCache)
	imageLoader := new(MockImageLoader)
	imageLoader.On("LoadImage", api.ImageRequestNext).Return(nil, nil)
	memoryDatabase := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(memoryDatabase, &StubImageFileConverter{})
	imageMetaDataStore := database.NewImageMetaDataStore(memoryDatabase)
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)
	statusStore := database.NewStatusStore(memoryDatabase)
	lib := library.NewImageService(
		sender,
		library.NewImageLibrary(imageCache, imageLoader, nil, imageStore, imageMetaDataStore, StubProgressReporter{}),
		statusStore,
	)
	filterService := filter.NewFilterService()

	sut := NewImageCategoryService(sender, lib, filterService, imageLoader, imageCategoryStore)

	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("filepath", "filename"))
	cat, _ := categoryStore.AddCategory(apitype.NewCategory("cat1", "cat_1", ""))
	_ = imageCategoryStore.CategorizeImage(imageFile.Id(), cat.Id(), apitype.CATEGORIZE)
	imageCategories, _ := imageCategoryStore.GetCategorizedImages()

	command := &api.PersistCategorizationCommand{
		KeepOriginals:  true,
		FixOrientation: false,
		Quality:        100,
	}
	operations, err := sut.ResolveOperationsForGroup(imageFile, imageCategories[imageFile.Id()], command)

	a.Nil(err)
	ops := operations.Operations()
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
	imageMetaDataStore := database.NewImageMetaDataStore(memoryDatabase)
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)
	statusStore := database.NewStatusStore(memoryDatabase)
	lib := library.NewImageService(
		sender,
		library.NewImageLibrary(imageCache, imageLoader, nil, imageStore, imageMetaDataStore, StubProgressReporter{}),
		statusStore,
	)
	filterService := filter.NewFilterService()

	sut := NewImageCategoryService(sender, lib, filterService, imageLoader, imageCategoryStore)

	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("filepath", "filename"))
	cat, _ := categoryStore.AddCategory(apitype.NewCategory("cat1", "cat_1", ""))
	_ = imageCategoryStore.CategorizeImage(imageFile.Id(), cat.Id(), apitype.CATEGORIZE)
	imageCategories, _ := imageCategoryStore.GetCategorizedImages()

	command := &api.PersistCategorizationCommand{
		KeepOriginals:  false,
		FixOrientation: false,
		Quality:        100,
	}
	operations, err := sut.ResolveOperationsForGroup(imageFile, imageCategories[imageFile.Id()], command)

	a.Nil(err)
	ops := operations.Operations()
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
	imageMetaDataStore := database.NewImageMetaDataStore(memoryDatabase)
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)
	statusStore := database.NewStatusStore(memoryDatabase)
	lib := library.NewImageService(
		sender,
		library.NewImageLibrary(imageCache, imageLoader, nil, imageStore, imageMetaDataStore, StubProgressReporter{}),
		statusStore,
	)
	filterService := filter.NewFilterService()

	sut := NewImageCategoryService(sender, lib, filterService, imageLoader, imageCategoryStore)

	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("filepath", "filename"))
	cat, _ := categoryStore.AddCategory(apitype.NewCategory("cat1", "cat_1", ""))
	_ = imageCategoryStore.CategorizeImage(imageFile.Id(), cat.Id(), apitype.CATEGORIZE)
	imageCategories, _ := imageCategoryStore.GetCategorizedImages()

	command := &api.PersistCategorizationCommand{
		KeepOriginals:  true,
		FixOrientation: true,
		Quality:        100,
	}
	operations, err := sut.ResolveOperationsForGroup(imageFile, imageCategories[imageFile.Id()], command)

	a.Nil(err)
	ops := operations.Operations()
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
	imageMetaDataStore := database.NewImageMetaDataStore(memoryDatabase)
	categoryStore := database.NewCategoryStore(memoryDatabase)
	imageCategoryStore := database.NewImageCategoryStore(memoryDatabase)
	statusStore := database.NewStatusStore(memoryDatabase)
	lib := library.NewImageService(
		sender,
		library.NewImageLibrary(imageCache, imageLoader, nil, imageStore, imageMetaDataStore, StubProgressReporter{}),
		statusStore,
	)
	filterService := filter.NewFilterService()

	sut := NewImageCategoryService(sender, lib, filterService, imageLoader, imageCategoryStore)

	imageFile, _ := imageStore.AddImage(apitype.NewImageFile("filepath", "filename"))
	cat, _ := categoryStore.AddCategory(apitype.NewCategory("cat1", "cat_1", ""))
	_ = imageCategoryStore.CategorizeImage(imageFile.Id(), cat.Id(), apitype.CATEGORIZE)
	imageCategories, _ := imageCategoryStore.GetCategorizedImages()

	command := &api.PersistCategorizationCommand{
		KeepOriginals:  false,
		FixOrientation: true,
		Quality:        100,
	}
	operations, err := sut.ResolveOperationsForGroup(imageFile, imageCategories[imageFile.Id()], command)

	a.Nil(err)
	ops := operations.Operations()
	a.Equal(3, len(ops))
	a.Equal("Exif Rotate", ops[0].String())
	a.Equal(fmt.Sprintf("Copy file 'filename' to '%s'", filepath.Join("filepath", "cat_1")), ops[1].String())
	a.Equal("Remove", ops[2].String())
}
