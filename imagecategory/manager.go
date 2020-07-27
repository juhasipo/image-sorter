package imagecategory

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/filter"
	"vincit.fi/image-sorter/imageloader/goimage"
	"vincit.fi/image-sorter/library"
)

const CATEGORIZATION_FILE_NAME = ".categorization"

type Manager struct {
	rootDir       string
	imageCategory map[string]map[string]*category.CategorizedImage
	sender        event.Sender
	library       library.Library
	filterManager *filter.Manager
	imageLoader   goimage.ImageLoader

	ImageCategoryManager
}

func ManagerNew(sender event.Sender, lib library.Library, filterManager *filter.Manager, imageLoader goimage.ImageLoader) ImageCategoryManager {
	var manager = Manager{
		imageCategory: map[string]map[string]*category.CategorizedImage{},
		sender:        sender,
		library:       lib,
		filterManager: filterManager,
		imageLoader:   imageLoader,
	}
	return &manager
}

func (s *Manager) InitializeForDirectory(directory string) {
	s.rootDir = directory
	s.imageCategory = map[string]map[string]*category.CategorizedImage{}
}

func (s *Manager) RequestCategory(handle *common.Handle) {
	s.sendCategories(handle)
}

func (s *Manager) GetCategories(handle *common.Handle) map[string]*category.CategorizedImage {
	if categories, ok := s.imageCategory[handle.GetId()]; ok {
		categorizedEntries := map[string]*category.CategorizedImage{}
		for _, categorizedImage := range categories {
			categorizedEntries[categorizedImage.GetEntry().GetId()] = categorizedImage
		}
		return categorizedEntries
	} else {
		return map[string]*category.CategorizedImage{}
	}
}

func (s *Manager) SetCategory(command *category.CategorizeCommand) {
	handle := command.GetHandle()
	categoryEntry := command.GetEntry()
	operation := command.GetOperation()

	// Find existing entry for the image
	var image = s.imageCategory[handle.GetId()]
	var categorizedImage *category.CategorizedImage = nil
	if command.ShouldForceToCategory() {
		log.Printf("Force to category for '%s'", handle.GetPath())
		image = map[string]*category.CategorizedImage{}
		s.imageCategory[handle.GetId()] = image
		if operation != common.NONE {
			categorizedImage = category.CategorizedImageNew(categoryEntry, operation)
			image[categoryEntry.GetId()] = categorizedImage
		}
	} else if image != nil {
		categorizedImage = image[categoryEntry.GetId()]
	}

	// Case entry was not found or should force to use only the new category
	if categorizedImage == nil && operation != common.NONE {
		if image == nil {
			log.Printf("Create category entry for '%s'", handle.GetPath())
			image = map[string]*category.CategorizedImage{}
			s.imageCategory[handle.GetId()] = image
		}
		log.Printf("Create category entry for '%s:%s'", handle.GetPath(), categoryEntry.GetName())
		categorizedImage = category.CategorizedImageNew(categoryEntry, operation)
		image[categoryEntry.GetId()] = categorizedImage
	}

	if operation == common.NONE || categorizedImage == nil {
		// Case entry is removed
		log.Printf("Remove entry for '%s:%s'", handle.GetPath(), categoryEntry.GetName())
		delete(s.imageCategory[handle.GetId()], categoryEntry.GetId())
		if len(s.imageCategory[handle.GetId()]) == 0 {
			log.Printf("Remove entry for '%s'", handle.GetPath())
			delete(s.imageCategory, handle.GetId())
		}
		s.sendCategories(command.GetHandle())
	} else {
		// Case entry found and not removed
		log.Printf("Update entry for '%s:%s' to %d", handle.GetPath(), categoryEntry.GetName(), operation)
		categorizedImage.SetOperation(operation)
		if command.ShouldStayOnSameImage() {
			s.sendCategories(command.GetHandle())
		} else {
			s.sendCategories(handle)
			time.Sleep(command.GetNextImageDelay())
			s.sender.SendToTopic(event.IMAGE_REQUEST_NEXT)
		}
	}
}

func (s *Manager) PersistImageCategories(options common.PersistCategorizationCommand) {
	log.Printf("Persisting files to categories")
	operationsByImage := s.ResolveFileOperations(s.imageCategory, options)

	total := len(operationsByImage)
	s.sender.SendToTopicWithData(event.UPDATE_PROCESS_STATUS, "categorize", 0, total)
	for i, operationGroup := range operationsByImage {
		err := operationGroup.Apply()
		if err != nil {
			log.Println("Error", err)
		}
		s.sender.SendToTopicWithData(event.UPDATE_PROCESS_STATUS, "categorize", i+1, total)
	}
	s.sender.SendToTopicWithData(event.DIRECTORY_CHANGED, s.rootDir)
}

func (s *Manager) ResolveFileOperations(imageCategory map[string]map[string]*category.CategorizedImage, options common.PersistCategorizationCommand) []*filter.ImageOperationGroup {
	var operationGroups []*filter.ImageOperationGroup

	for handleId, categoryEntries := range imageCategory {
		handle := s.library.GetHandleById(handleId)
		if newOperationGroup, err := s.ResolveOperationsForGroup(handle, categoryEntries, options); err == nil {
			operationGroups = append(operationGroups, newOperationGroup)
		}
	}

	return operationGroups
}

func (s *Manager) ResolveOperationsForGroup(handle *common.Handle,
	categoryEntries map[string]*category.CategorizedImage,
	options common.PersistCategorizationCommand) (*filter.ImageOperationGroup, error) {
	dir, file := filepath.Split(handle.GetPath())

	filters := s.filterManager.GetFilters(handle, options)

	var imageOperations []filter.ImageOperation
	for _, categorizedImage := range categoryEntries {
		targetDirName := categorizedImage.GetEntry().GetSubPath()
		targetDir := filepath.Join(dir, targetDirName)

		for _, f := range filters {
			imageOperations = append(imageOperations, f.GetOperation())
		}
		imageOperations = append(imageOperations, filter.ImageCopyNew(targetDir, file, options.GetQuality()))
	}
	if !options.ShouldKeepOriginals() {
		imageOperations = append(imageOperations, filter.ImageRemoveNew())
	}

	if fullImage, err := s.imageLoader.LoadImage(handle); err != nil {
		log.Println("Could not load image", err)
		return nil, err
	} else if exifData, err := s.imageLoader.LoadExifData(handle); err != nil {
		log.Println("Could not load exif data")
		return nil, err
	} else {
		return filter.ImageOperationGroupNew(handle, fullImage, exifData, imageOperations), nil
	}
}

func (s *Manager) Close() {
	log.Print("Shutting down image category manager")
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
	s.sender.SendToTopicWithData(event.IMAGE_SHOW_ONLY, cat.GetName(), handles)
}

func (s *Manager) LoadCategorization(handleManager library.Library, categoryManager category.CategoryManager) {
	filePath := filepath.Join(s.rootDir, CATEGORIZATION_FILE_NAME)

	log.Printf("Loading categozation from file '%s'", filePath)
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		log.Print("Can't read file ", filePath, err)
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
			s.imageCategory[handle.GetId()] = map[string]*category.CategorizedImage{}
			categoryMap = s.imageCategory[handle.GetId()]
		}

		for _, c := range categories {
			if c != "" {
				entry := categoryManager.GetCategoryById(c)
				if entry != nil {
					categoryMap[entry.GetId()] = category.CategorizedImageNew(entry, common.MOVE)
				}
			}
		}
	}
}

func (s *Manager) PersistCategorization() {
	filePath := filepath.Join(s.rootDir, CATEGORIZATION_FILE_NAME)

	log.Printf("Saving image categorization to file '%s'", filePath)
	f, err := os.Create(filePath)
	if err != nil {
		log.Println("Can't write file ", filePath, err)
		return
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	w.WriteString("#version:1")
	w.WriteString("\n")
	for handleId, categorization := range s.imageCategory {
		if handleId != "" {
			w.WriteString(handleId)
			w.WriteString(";")
			for entry, categorizedImage := range categorization {
				if categorizedImage.GetOperation() == common.MOVE {
					w.WriteString(entry)
					w.WriteString(";")
				}
			}
			w.WriteString("\n")
		}
	}
	w.Flush()
}

func (s *Manager) getCategories(image *common.Handle) []*category.CategorizedImage {
	var categories []*category.CategorizedImage

	if i, ok := s.imageCategory[image.GetId()]; ok {
		for _, categorizedImage := range i {
			categories = append(categories, categorizedImage)
		}
	}

	return categories
}

func (s *Manager) sendCategories(currentImage *common.Handle) {
	var commands []*category.CategorizeCommand
	if currentImage != nil {
		var categories = s.getCategories(currentImage)

		for _, image := range categories {
			commands = append(commands, category.CategorizeCommandNew(currentImage, image.GetEntry(), image.GetOperation()))
		}
	}
	s.sender.SendToTopicWithData(event.CATEGORY_IMAGE_UPDATE, commands)
}
