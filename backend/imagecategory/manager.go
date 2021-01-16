package imagecategory

import (
	"path/filepath"
	"time"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/database"
	"vincit.fi/image-sorter/backend/filter"
	"vincit.fi/image-sorter/common/constants"
	"vincit.fi/image-sorter/common/logger"
)

type Manager struct {
	rootDir            string
	settingDir         string
	sender             api.Sender
	library            api.Library
	filterManager      *filter.Manager
	imageLoader        api.ImageLoader
	imageCategoryStore *database.ImageCategoryStore

	api.ImageCategoryManager
}

func NewImageCategoryManager(sender api.Sender, lib api.Library, filterManager *filter.Manager, imageLoader api.ImageLoader, imageCategoryStore *database.ImageCategoryStore) api.ImageCategoryManager {
	var manager = Manager{
		sender:             sender,
		library:            lib,
		filterManager:      filterManager,
		imageLoader:        imageLoader,
		imageCategoryStore: imageCategoryStore,
	}
	return &manager
}

func (s *Manager) InitializeForDirectory(directory string) {
	s.rootDir = directory
	s.settingDir = filepath.Join(directory, constants.ImageSorterDir)
}

func (s *Manager) RequestCategory(query *api.ImageCategoryQuery) {
	s.sendCategories(query.ImageId)
}

func (s *Manager) GetCategories(query *api.ImageCategoryQuery) map[apitype.CategoryId]*api.CategorizedImage {
	if categories, err := s.imageCategoryStore.GetImagesCategories(query.ImageId); err != nil {
		s.sender.SendError("Error while fetching image's category", err)
		return map[apitype.CategoryId]*api.CategorizedImage{}
	} else {
		categorizedEntries := map[apitype.CategoryId]*api.CategorizedImage{}
		for _, categorizedImage := range categories {
			categorizedEntries[categorizedImage.Category.GetId()] = categorizedImage
		}
		return categorizedEntries
	}
}

func (s *Manager) SetCategory(command *api.CategorizeCommand) {
	imageId := command.ImageId
	categoryId := command.CategoryId
	operation := command.Operation

	if command.ForceToCategory {
		logger.Debug.Printf("Force to category for '%d'", imageId)
		if err := s.imageCategoryStore.RemoveImageCategories(imageId); err != nil {
			s.sender.SendError("Error while removing image categories", err)
		}
	}

	if err := s.imageCategoryStore.CategorizeImage(imageId, categoryId, operation); err != nil {
		s.sender.SendError("Error while setting category", err)
	}

	if command.StayOnSameImage {
		s.sendCategories(command.ImageId)
	} else {
		s.sendCategories(imageId)
		time.Sleep(command.NextImageDelay)
		s.sender.SendToTopic(api.ImageRequestNext)
	}
}

func (s *Manager) PersistImageCategories(options *api.PersistCategorizationCommand) {
	logger.Debug.Printf("Persisting files to categories")
	imageCategory, _ := s.imageCategoryStore.GetCategorizedImages()
	operationsByImage := s.ResolveFileOperations(imageCategory, options)

	total := len(operationsByImage)
	s.sender.SendCommandToTopic(api.ProcessStatusUpdated, &api.UpdateProgressCommand{
		Name:    "categorize",
		Current: 0,
		Total:   total,
	})
	for i, operationGroup := range operationsByImage {
		err := operationGroup.Apply()
		if err != nil {
			s.sender.SendError("Error while applying changes", err)
		}
		s.sender.SendCommandToTopic(api.ProcessStatusUpdated, &api.UpdateProgressCommand{
			Name:    "categorize",
			Current: i + 1,
			Total:   total,
		})
	}

	s.sender.SendCommandToTopic(api.DirectoryChanged, s.rootDir)
}

func (s *Manager) ResolveFileOperations(
	imageCategory map[apitype.ImageId]map[apitype.CategoryId]*api.CategorizedImage,
	options *api.PersistCategorizationCommand) []*apitype.ImageOperationGroup {
	var operationGroups []*apitype.ImageOperationGroup

	for imageId, categoryEntries := range imageCategory {
		imageFile := s.library.GetImageFileById(imageId)
		if newOperationGroup, err := s.ResolveOperationsForGroup(imageFile, categoryEntries, options); err == nil {
			operationGroups = append(operationGroups, newOperationGroup)
		}
	}

	return operationGroups
}

func (s *Manager) ResolveOperationsForGroup(
	imageFile *apitype.ImageFileWithMetaData,
	categoryEntries map[apitype.CategoryId]*api.CategorizedImage,
	options *api.PersistCategorizationCommand) (*apitype.ImageOperationGroup, error) {
	dir, file := imageFile.GetImageFile().GetDir(), imageFile.GetImageFile().GetFile()

	filters := s.filterManager.GetFilters(imageFile.GetImageId(), options)

	var imageOperations []apitype.ImageOperation
	for _, categorizedImage := range categoryEntries {
		targetDirName := categorizedImage.Category.GetSubPath()
		targetDir := filepath.Join(dir, targetDirName)

		for _, f := range filters {
			imageOperations = append(imageOperations, f.GetOperation())
		}
		imageOperations = append(imageOperations, filter.NewImageCopy(targetDir, file, options.Quality))
	}
	if !options.KeepOriginals {
		imageOperations = append(imageOperations, filter.NewImageRemove())
	}

	if fullImage, err := s.imageLoader.LoadImage(imageFile.GetImageId()); err != nil {
		s.sender.SendError("Could not load image", err)
		return nil, err
	} else if exifData, err := s.imageLoader.LoadExifData(imageFile.GetImageId()); err != nil {
		s.sender.SendError("Could not load exif data", err)
		return nil, err
	} else {
		return apitype.NewImageOperationGroup(imageFile, fullImage, exifData, imageOperations), nil
	}
}

func (s *Manager) Close() {
	logger.Info.Print("Shutting down image category manager")
}

func (s *Manager) ShowOnlyCategoryImages(command *api.SelectCategoryCommand) {
	s.sender.SendCommandToTopic(api.ImageShowOnly, command)
}

func (s *Manager) getCategories(imageId apitype.ImageId) []*api.CategorizedImage {
	if categories, err := s.imageCategoryStore.GetImagesCategories(imageId); err != nil {
		s.sender.SendError("Error while fetching categories for image", err)
		return []*api.CategorizedImage{}
	} else {
		return categories
	}
}

func (s *Manager) sendCategories(currentImageId apitype.ImageId) {
	var commands []*apitype.Category
	if currentImageId != apitype.ImageId(-1) {
		var categories = s.getCategories(currentImageId)

		for _, image := range categories {
			commands = append(commands, image.Category)
		}
	}
	s.sender.SendCommandToTopic(api.CategoryImageUpdate, &api.CategoriesCommand{
		Categories: commands,
	})
}
