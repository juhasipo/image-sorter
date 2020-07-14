package imagecategory

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/library"
)

const CATEGORIZATION_FILE_NAME = ".categorization"

type Manager struct {
	rootDir       string
	imageCategory map[*common.Handle]map[string]*category.CategorizedImage
	sender        event.Sender

	ImageCategoryManager
}

func ManagerNew(sender event.Sender) ImageCategoryManager {
	var manager = Manager{
		imageCategory: map[*common.Handle]map[string]*category.CategorizedImage{},
		sender:        sender,
	}
	return &manager
}

func (s *Manager) InitializeForDirectory(directory string) {
	s.rootDir = directory
	s.imageCategory = map[*common.Handle]map[string]*category.CategorizedImage{}
}

func (s *Manager) RequestCategory(handle *common.Handle) {
	s.sendCategories(handle)
}

func (s *Manager) SetCategory(command *category.CategorizeCommand) {
	handle := command.GetHandle()
	categoryEntry := command.GetEntry()
	operation := command.GetOperation()

	var image = s.imageCategory[handle]
	var categorizedImage *category.CategorizedImage = nil
	if image != nil {
		categorizedImage = image[categoryEntry.GetId()]
	}

	if categorizedImage == nil && operation != common.NONE {
		if image == nil {
			log.Printf("Create category entry for '%s'", handle.GetPath())
			image = map[string]*category.CategorizedImage{}
			s.imageCategory[handle] = image
		}
		log.Printf("Create category entry for '%s:%s'", handle.GetPath(), categoryEntry.GetName())
		categorizedImage = category.CategorizedImageNew(categoryEntry, operation)
		image[categoryEntry.GetId()] = categorizedImage
	}

	if operation == common.NONE || categorizedImage == nil {
		log.Printf("Remove entry for '%s:%s'", handle.GetPath(), categoryEntry.GetName())
		delete(s.imageCategory[handle], categoryEntry.GetId())
		if len(s.imageCategory[handle]) == 0 {
			log.Printf("Remove entry for '%s'", handle.GetPath())
			delete(s.imageCategory, handle)
		}
		s.sendCategories(command.GetHandle())
	} else {
		log.Printf("Update entry for '%s:%s' to %d", handle.GetPath(), categoryEntry.GetName(), operation)
		categorizedImage.SetOperation(operation)
		if command.ShouldStayOnSameImage() {
			s.sendCategories(command.GetHandle())
		} else {
			s.sender.SendToTopic(event.IMAGE_REQUEST_NEXT)
		}
	}
}

func (s* Manager) PersistImageCategories() {
	log.Printf("Persisting files to categories")
	for handle, categoryEntries := range s.imageCategory {
		s.PersistImageCategory(handle, categoryEntries)
	}
}

func (s* Manager) PersistImageCategory(handle *common.Handle, categories map[string]*category.CategorizedImage) {
	log.Printf(" - Persisting '%s'", handle.GetPath())
	dir, file := filepath.Split(handle.GetPath())

	for _, image := range categories {
		targetDirName := image.GetEntry().GetSubPath()
		targetDir := filepath.Join(dir, targetDirName)

		// Always copy first because picture may have multiple categories
		if image.GetOperation() != common.NONE {
			common.CopyFile(dir, file, targetDir, file)
		}
	}
	common.RemoveFile(handle.GetPath())
}

func (s *Manager) Close() {
	log.Print("Shutting down image category manager")
	s.PersistCategorization()
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

		categoryMap := s.imageCategory[handle]
		if categoryMap == nil {
			s.imageCategory[handle] = map[string]*category.CategorizedImage{}
			categoryMap = s.imageCategory[handle]
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
		log.Panic("Can't write file ", filePath, err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	w.WriteString("#version:1")
	w.WriteString("\n")
	for handle, categorization := range s.imageCategory {
		if handle != nil {
			w.WriteString(handle.GetId())
			w.WriteString(":")
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

	if i, ok := s.imageCategory[image]; ok {
		for _, categorizedImage := range i {
			categories = append(categories, categorizedImage)
		}
	}

	return categories
}

func (s *Manager) sendCategories(currentImage *common.Handle) {
	var categories = s.getCategories(currentImage)
	var commands []*category.CategorizeCommand
	for _, image := range categories {
		commands = append(commands, category.CategorizeCommandNew(currentImage, image.GetEntry(), image.GetOperation()))
	}
	s.sender.SendToTopicWithData(event.CATEGORY_IMAGE_UPDATE, commands)
}
