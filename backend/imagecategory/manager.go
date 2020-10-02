package imagecategory

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"time"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/backend/filter"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/common/constants"
	"vincit.fi/image-sorter/common/event"
	"vincit.fi/image-sorter/common/logger"
)

type Manager struct {
	rootDir       string
	settingDir    string
	imageCategory map[string]map[string]*api.CategorizedImage
	sender        event.Sender
	library       api.Library
	filterManager *filter.Manager
	imageLoader   api.ImageLoader

	api.ImageCategoryManager
}

func NewImageCategoryManager(sender event.Sender, lib api.Library, filterManager *filter.Manager, imageLoader api.ImageLoader) api.ImageCategoryManager {
	var manager = Manager{
		imageCategory: map[string]map[string]*api.CategorizedImage{},
		sender:        sender,
		library:       lib,
		filterManager: filterManager,
		imageLoader:   imageLoader,
	}
	return &manager
}

func (s *Manager) InitializeForDirectory(directory string) {
	s.rootDir = directory
	s.settingDir = filepath.Join(directory, constants.ImageSorterDir)
	s.imageCategory = map[string]map[string]*api.CategorizedImage{}
}

func (s *Manager) RequestCategory(handle *common.Handle) {
	s.sendCategories(handle)
}

func (s *Manager) GetCategories(handle *common.Handle) map[string]*api.CategorizedImage {
	if categories, ok := s.imageCategory[handle.GetId()]; ok {
		categorizedEntries := map[string]*api.CategorizedImage{}
		for _, categorizedImage := range categories {
			categorizedEntries[categorizedImage.GetEntry().GetId()] = categorizedImage
		}
		return categorizedEntries
	} else {
		return map[string]*api.CategorizedImage{}
	}
}

func (s *Manager) SetCategory(command *api.CategorizeCommand) {
	handle := command.GetHandle()
	categoryEntry := command.GetEntry()
	operation := command.GetOperation()

	// Find existing entry for the image
	var image = s.imageCategory[handle.GetId()]
	var categorizedImage *api.CategorizedImage = nil
	if command.ShouldForceToCategory() {
		logger.Debug.Printf("Force to category for '%s'", handle.GetPath())
		image = map[string]*api.CategorizedImage{}
		s.imageCategory[handle.GetId()] = image
		if operation != common.NONE {
			categorizedImage = api.NewCategorizedImage(categoryEntry, operation)
			image[categoryEntry.GetId()] = categorizedImage
		}
	} else if image != nil {
		categorizedImage = image[categoryEntry.GetId()]
	}

	// Case entry was not found or should force to use only the new category
	if categorizedImage == nil && operation != common.NONE {
		if image == nil {
			logger.Debug.Printf("Create category entry for '%s'", handle.GetPath())
			image = map[string]*api.CategorizedImage{}
			s.imageCategory[handle.GetId()] = image
		}
		logger.Debug.Printf("Create category entry for '%s:%s'", handle.GetPath(), categoryEntry.GetName())
		categorizedImage = api.NewCategorizedImage(categoryEntry, operation)
		image[categoryEntry.GetId()] = categorizedImage
	}

	if operation == common.NONE || categorizedImage == nil {
		// Case entry is removed
		logger.Debug.Printf("Remove entry for '%s:%s'", handle.GetPath(), categoryEntry.GetName())
		delete(s.imageCategory[handle.GetId()], categoryEntry.GetId())
		if len(s.imageCategory[handle.GetId()]) == 0 {
			logger.Debug.Printf("Remove entry for '%s'", handle.GetPath())
			delete(s.imageCategory, handle.GetId())
		}
		s.sendCategories(command.GetHandle())
	} else {
		// Case entry found and not removed
		logger.Debug.Printf("Update entry for '%s:%s' to %d", handle.GetPath(), categoryEntry.GetName(), operation)
		categorizedImage.SetOperation(operation)
		if command.ShouldStayOnSameImage() {
			s.sendCategories(command.GetHandle())
		} else {
			s.sendCategories(handle)
			time.Sleep(command.GetNextImageDelay())
			s.sender.SendToTopic(event.ImageRequestNext)
		}
	}
}

func (s *Manager) PersistImageCategories(options common.PersistCategorizationCommand) {
	logger.Debug.Printf("Persisting files to categories")
	operationsByImage := s.ResolveFileOperations(s.imageCategory, options)

	total := len(operationsByImage)
	s.sender.SendToTopicWithData(event.ProcessStatusUpdated, "categorize", 0, total)
	for i, operationGroup := range operationsByImage {
		err := operationGroup.Apply()
		if err != nil {
			logger.Error.Println("Error", err)
		}
		s.sender.SendToTopicWithData(event.ProcessStatusUpdated, "categorize", i+1, total)
	}
	s.sender.SendToTopicWithData(event.DirectoryChanged, s.rootDir)
}

func (s *Manager) ResolveFileOperations(imageCategory map[string]map[string]*api.CategorizedImage, options common.PersistCategorizationCommand) []*api.ImageOperationGroup {
	var operationGroups []*api.ImageOperationGroup

	for handleId, categoryEntries := range imageCategory {
		handle := s.library.GetHandleById(handleId)
		if newOperationGroup, err := s.ResolveOperationsForGroup(handle, categoryEntries, options); err == nil {
			operationGroups = append(operationGroups, newOperationGroup)
		}
	}

	return operationGroups
}

func (s *Manager) ResolveOperationsForGroup(handle *common.Handle,
	categoryEntries map[string]*api.CategorizedImage,
	options common.PersistCategorizationCommand) (*api.ImageOperationGroup, error) {
	dir, file := filepath.Split(handle.GetPath())

	filters := s.filterManager.GetFilters(handle, options)

	var imageOperations []api.ImageOperation
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
		return api.NewImageOperationGroup(handle, fullImage, exifData, imageOperations), nil
	}
}

func (s *Manager) Close() {
	logger.Info.Print("Shutting down image category manager")
	s.PersistCategorization()
}

func (s *Manager) ShowOnlyCategoryImages(cat *common.Category) {
	var handles []*common.Handle
	for key, img := range s.imageCategory {
		if _, ok := img[cat.GetId()]; ok {
			handle := s.library.GetHandleById(key)
			handles = append(handles, handle)
		}
	}
	s.sender.SendToTopicWithData(event.ImageShowOnly, cat.GetName(), handles)
}

func (s *Manager) LoadCategorization(handleManager api.Library, categoryManager api.CategoryManager) {
	filePath := filepath.Join(s.settingDir, constants.CategorizationFileName)

	logger.Info.Printf("Loading categozation from file '%s'", filePath)
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		logger.Error.Print("Can't read file ", filePath, err)
		return
	}

	var lines []string
	scanner := bufio.NewScanner(f)
	// Read version even though it is not used yet
	scanner.Scan()

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	for _, line := range lines {
		parts := strings.Split(line, ";")
		handle := handleManager.GetHandleById(parts[0])
		categories := parts[1:]

		if !handle.IsValid() {
			continue
		}

		categoryMap := s.imageCategory[handle.GetId()]
		if categoryMap == nil {
			s.imageCategory[handle.GetId()] = map[string]*api.CategorizedImage{}
			categoryMap = s.imageCategory[handle.GetId()]
		}

		for _, c := range categories {
			if c != "" {
				entry := categoryManager.GetCategoryById(c)
				if entry != nil {
					categoryMap[entry.GetId()] = api.NewCategorizedImage(entry, common.MOVE)
				}
			}
		}
	}
}

func (s *Manager) PersistCategorization() {
	filePath := filepath.Join(s.settingDir, constants.CategorizationFileName)

	logger.Info.Printf("Saving image categorization to file '%s'", filePath)
	f, err := os.Create(filePath)
	if err != nil {
		logger.Error.Println("Can't write file ", filePath, err)
		return
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	_, _ = w.WriteString("#version:1")
	_, _ = w.WriteString("\n")
	for handleId, categorization := range s.imageCategory {
		if handleId != "" {
			_, _ = w.WriteString(handleId)
			_, _ = w.WriteString(";")
			for entry, categorizedImage := range categorization {
				if categorizedImage.GetOperation() == common.MOVE {
					_, _ = w.WriteString(entry)
					_, _ = w.WriteString(";")
				}
			}
			_, _ = w.WriteString("\n")
		}
	}
	_ = w.Flush()
}

func (s *Manager) getCategories(image *common.Handle) []*api.CategorizedImage {
	var categories []*api.CategorizedImage

	if i, ok := s.imageCategory[image.GetId()]; ok {
		for _, categorizedImage := range i {
			categories = append(categories, categorizedImage)
		}
	}

	return categories
}

func (s *Manager) sendCategories(currentImage *common.Handle) {
	var commands []*api.CategorizeCommand
	if currentImage != nil {
		var categories = s.getCategories(currentImage)

		for _, image := range categories {
			commands = append(commands, api.NewCategorizeCommand(currentImage, image.GetEntry(), image.GetOperation()))
		}
	}
	s.sender.SendToTopicWithData(event.CategoryImageUpdate, commands)
}