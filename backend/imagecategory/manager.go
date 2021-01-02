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
	rootDir       string
	settingDir    string
	sender        api.Sender
	library       api.Library
	filterManager *filter.Manager
	imageLoader   api.ImageLoader
	store         *database.Store

	api.ImageCategoryManager
}

func NewImageCategoryManager(sender api.Sender, lib api.Library, filterManager *filter.Manager, imageLoader api.ImageLoader, store *database.Store) api.ImageCategoryManager {
	var manager = Manager{
		sender:        sender,
		library:       lib,
		filterManager: filterManager,
		imageLoader:   imageLoader,
		store:         store,
	}
	return &manager
}

func (s *Manager) InitializeForDirectory(directory string) {
	s.rootDir = directory
	s.settingDir = filepath.Join(directory, constants.ImageSorterDir)
}

func (s *Manager) RequestCategory(handle *apitype.Handle) {
	s.sendCategories(handle)
}

func (s *Manager) GetCategories(handle *apitype.Handle) map[int64]*apitype.CategorizedImage {
	if categories, err := s.store.GetImagesCategories(handle.GetId()); err != nil {
		logger.Error.Print("Error while fetching images's category", err)
		return map[int64]*apitype.CategorizedImage{}
	} else {
		categorizedEntries := map[int64]*apitype.CategorizedImage{}
		for _, categorizedImage := range categories {
			categorizedEntries[categorizedImage.GetEntry().GetId()] = categorizedImage
		}
		return categorizedEntries
	}
}

func (s *Manager) SetCategory(command *apitype.CategorizeCommand) {
	handle := command.GetHandle()
	categoryEntry := command.GetEntry()
	operation := command.GetOperation()

	if command.ShouldForceToCategory() {
		logger.Debug.Printf("Force to category for '%s'", handle.GetPath())
		if err := s.store.RemoveImageCategories(handle.GetId()); err != nil {
			logger.Error.Print("Error while removing image categories", err)
		}
	}

	if err := s.store.CategorizeImage(handle.GetId(), categoryEntry.GetId(), operation); err != nil {
		logger.Error.Print("Error while setting category", err)
	}

	if command.ShouldStayOnSameImage() {
		s.sendCategories(command.GetHandle())
	} else {
		s.sendCategories(handle)
		time.Sleep(command.GetNextImageDelay())
		s.sender.SendToTopic(api.ImageRequestNext)
	}
}

func (s *Manager) PersistImageCategories(options apitype.PersistCategorizationCommand) {
	logger.Debug.Printf("Persisting files to categories")
	imageCategory, _ := s.store.GetCategorizedImages()
	operationsByImage := s.ResolveFileOperations(imageCategory, options)

	total := len(operationsByImage)
	s.sender.SendToTopicWithData(api.ProcessStatusUpdated, "categorize", 0, total)
	for i, operationGroup := range operationsByImage {
		err := operationGroup.Apply()
		if err != nil {
			logger.Error.Println("Error", err)
		}
		s.sender.SendToTopicWithData(api.ProcessStatusUpdated, "categorize", i+1, total)
	}
	s.sender.SendToTopicWithData(api.DirectoryChanged, s.rootDir)
}

func (s *Manager) ResolveFileOperations(imageCategory map[apitype.HandleId]map[int64]*apitype.CategorizedImage, options apitype.PersistCategorizationCommand) []*apitype.ImageOperationGroup {
	var operationGroups []*apitype.ImageOperationGroup

	for handleId, categoryEntries := range imageCategory {
		handle := s.library.GetHandleById(handleId)
		if newOperationGroup, err := s.ResolveOperationsForGroup(handle, categoryEntries, options); err == nil {
			operationGroups = append(operationGroups, newOperationGroup)
		}
	}

	return operationGroups
}

func (s *Manager) ResolveOperationsForGroup(handle *apitype.Handle,
	categoryEntries map[int64]*apitype.CategorizedImage,
	options apitype.PersistCategorizationCommand) (*apitype.ImageOperationGroup, error) {
	dir, file := filepath.Split(handle.GetPath())

	filters := s.filterManager.GetFilters(handle, options)

	var imageOperations []apitype.ImageOperation
	for _, categorizedImage := range categoryEntries {
		targetDirName := categorizedImage.GetEntry().GetSubPath()
		targetDir := filepath.Join(dir, targetDirName)

		for _, f := range filters {
			imageOperations = append(imageOperations, f.GetOperation())
		}
		imageOperations = append(imageOperations, filter.NewImageCopy(targetDir, file, options.GetQuality()))
	}
	if !options.ShouldKeepOriginals() {
		imageOperations = append(imageOperations, filter.NewImageRemove())
	}

	if fullImage, err := s.imageLoader.LoadImage(handle); err != nil {
		logger.Error.Println("Could not load image", err)
		return nil, err
	} else if exifData, err := s.imageLoader.LoadExifData(handle); err != nil {
		logger.Error.Println("Could not load exif data")
		return nil, err
	} else {
		return apitype.NewImageOperationGroup(handle, fullImage, exifData, imageOperations), nil
	}
}

func (s *Manager) Close() {
	logger.Info.Print("Shutting down image category manager")
}

func (s *Manager) ShowOnlyCategoryImages(cat *apitype.Category) {
	var handles []*apitype.Handle
	categorizedImages, _ := s.store.GetCategorizedImages()
	for key, img := range categorizedImages {
		if _, ok := img[cat.GetId()]; ok {
			handle := s.library.GetHandleById(key)
			handles = append(handles, handle)
		}
	}
	s.sender.SendToTopicWithData(api.ImageShowOnly, cat.GetName(), handles)
}

func (s *Manager) getCategories(image *apitype.Handle) []*apitype.CategorizedImage {
	if cats, err := s.store.GetImagesCategories(image.GetId()); err != nil {
		logger.Error.Print("Error while fetching categories for image", err)
		return []*apitype.CategorizedImage{}
	} else {
		return cats
	}
}

func (s *Manager) sendCategories(currentImage *apitype.Handle) {
	var commands []*apitype.CategorizeCommand
	if currentImage != nil {
		var categories = s.getCategories(currentImage)

		for _, image := range categories {
			commands = append(commands, apitype.NewCategorizeCommand(currentImage, image.GetEntry(), image.GetOperation()))
		}
	}
	s.sender.SendToTopicWithData(api.CategoryImageUpdate, commands)
}
