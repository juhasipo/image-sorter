package category

import (
	"bufio"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/logger"
)

const IMAGE_SORTER_DIR = ".image-sorter"
const CATEGORIES_FILE_NAME = ".categories"

type Manager struct {
	commandLineCategories []string
	categories            []*common.Category
	categoriesById        map[string]*common.Category
	sender                event.Sender
	rootDir               string

	CategoryManager
}

func Parse(name string) (string, string, string) {
	parts := strings.Split(name, ":")

	if len(parts) == 2 {
		return parts[0], parts[0], parts[1]
	} else {
		return parts[0], parts[1], parts[2]
	}
}

func New(sender event.Sender, categories []string) CategoryManager {
	manager := Manager{
		sender:                sender,
		commandLineCategories: categories,
	}
	return &manager
}

func (s *Manager) InitializeFromDirectory(categories []string, rootDir string) {
	var loadedCategories []*common.Category
	var categoriesByName = map[string]*common.Category{}
	s.rootDir = filepath.Join(rootDir, IMAGE_SORTER_DIR)

	if len(categories) > 0 && categories[0] != "" {
		logger.Info.Printf("Reading from command line parameters")
		loadedCategories = fromCategoriesStrings(categories)
	} else {
		loadedCategories = loadCategoriesFromFile(s.rootDir)
	}

	for _, category := range loadedCategories {
		categoriesByName[category.GetName()] = category
	}

	s.categories = loadedCategories
	s.categoriesById = categoriesByName
}

func (s *Manager) GetCategories() []*common.Category {
	return s.categories
}

func (s *Manager) RequestCategories() {
	s.sender.SendToTopicWithData(event.CATEGORIES_UPDATED, &CategoriesCommand{
		categories: s.categories,
	})
}

func (s *Manager) Save(categories []*common.Category) {
	s.resetCategories(categories)

	saveCategoriesToFile(s.rootDir, CATEGORIES_FILE_NAME, categories)
	s.sender.SendToTopicWithData(event.CATEGORIES_UPDATED, &CategoriesCommand{
		categories: categories,
	})
}
func (s *Manager) SaveDefault(categories []*common.Category) {
	s.resetCategories(categories)

	if currentUser, err := user.Current(); err != nil {
		logger.Error.Println("Could not find current user", err)
	} else {
		categoryFile := filepath.Join(currentUser.HomeDir, IMAGE_SORTER_DIR)

		saveCategoriesToFile(categoryFile, CATEGORIES_FILE_NAME, categories)
		s.sender.SendToTopicWithData(event.CATEGORIES_UPDATED, &CategoriesCommand{
			categories: categories,
		})
	}
}

func (s *Manager) resetCategories(categories []*common.Category) {
	s.categories = categories
	for _, category := range categories {
		s.categoriesById[category.GetId()] = category
	}
}

func (s *Manager) Close() {
	logger.Info.Print("Shutting down category manager")
	saveCategoriesToFile(s.rootDir, CATEGORIES_FILE_NAME, s.categories)
}

func (s *Manager) GetCategoryById(id string) *common.Category {
	return s.categoriesById[id]
}

func saveCategoriesToFile(fileDir string, fileName string, categories []*common.Category) {
	if _, err := os.Stat(fileDir); os.IsNotExist(err) {
		os.Mkdir(fileDir, 0666)
	}

	filePath := filepath.Join(fileDir, fileName)

	logger.Info.Printf("Saving categories to file '%s'", filePath)
	f, err := os.Create(filePath)
	if err != nil {
		logger.Error.Panic("Can't write file ", filePath, err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	w.WriteString("#version:1")
	w.WriteString("\n")
	for _, category := range categories {
		w.WriteString(category.Serialize())
		w.WriteString("\n")
	}
	w.Flush()
}

func fromCategoriesStrings(categories []string) []*common.Category {
	var categoryEntries []*common.Category
	for _, categoryName := range categories {
		if len(categoryName) > 0 {
			categoryEntries = append(categoryEntries, common.CategoryEntryNew(Parse(categoryName)))
		}
	}
	logger.Debug.Printf("Parsed %d categories", len(categoryEntries))
	for _, entry := range categoryEntries {
		logger.Trace.Printf(" - %s", entry.GetName())
	}
	return categoryEntries
}

func loadCategoriesFromFile(fileDir string) []*common.Category {
	if currentUser, err := user.Current(); err == nil {
		filePaths := []string{
			filepath.Join(fileDir, CATEGORIES_FILE_NAME),
			filepath.Join(currentUser.HomeDir, IMAGE_SORTER_DIR, CATEGORIES_FILE_NAME),
		}

		filePath := common.GetFirstExistingFilePath(filePaths)

		logger.Info.Printf("Reading categories from file '%s'", filePath)

		if f, err := os.OpenFile(filePath, os.O_RDONLY, 0666); err == nil {
			defer f.Close()
			var lines []string
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}

			if lines != nil {
				return fromCategoriesStrings(lines[1:])
			} else {
				return []*common.Category{}
			}
		} else {
			logger.Error.Println("Could not open file: "+filePath, err)
			return []*common.Category{}
		}
	} else {
		logger.Error.Println("Could not find current user", err)
		return []*common.Category{}
	}
}
