package category

import (
	"bufio"
	"log"
	"os"
	"path"
	"strings"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
)

const CATEGORIES_FILE_NAME = ".categories"

type Manager struct {
	categories     []*common.Category
	categoriesById map[string]*common.Category
	sender         event.Sender
	rootDir        string

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

func New(sender event.Sender, categories []string, rootDir string) CategoryManager {
	var loadedCategories []*common.Category
	var categoriesByName = map[string]*common.Category{}

	if len(categories) > 0 && categories[0] != "" {
		log.Printf("Reading from command line parameters")
		loadedCategories = fromCategoriesStrings(categories)
	} else {
		loadedCategories = loadCategoriesToFile(rootDir)
	}

	for _, category := range loadedCategories {
		categoriesByName[category.GetName()] = category
	}

	return &Manager {
		categories:     loadedCategories,
		sender:         sender,
		rootDir:        rootDir,
		categoriesById: categoriesByName,
	}
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

	// TODO: Find user's home dir
	saveCategoriesToFile(s.rootDir, CATEGORIES_FILE_NAME, categories)
	s.sender.SendToTopicWithData(event.CATEGORIES_UPDATED, &CategoriesCommand{
		categories: categories,
	})
}

func (s *Manager) resetCategories(categories []*common.Category) {
	s.categories = categories
	for _, category := range categories {
		s.categoriesById[category.GetId()] = category
	}
}


func (s *Manager) Close() {
	log.Print("Shutting down category manager")
	saveCategoriesToFile(s.rootDir, CATEGORIES_FILE_NAME, s.categories)
}

func (s *Manager) GetCategoryById(id string) *common.Category {
	return s.categoriesById[id]
}

func saveCategoriesToFile(fileDir string, fileName string, categories []*common.Category) {
	filePath := path.Join(fileDir, fileName)

	log.Printf("Saving categories to file '%s'", filePath)
	f, err := os.Create(filePath)
	if err != nil {
		log.Panic("Can't write file ", filePath, err)
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
	log.Printf("Parsed %d categories:", len(categoryEntries))
	for _, entry := range categoryEntries {
		log.Printf(" - %s", entry.GetName())
	}
	return categoryEntries
}


func loadCategoriesToFile(fileDir string) []*common.Category {
	filePath := path.Join(fileDir, CATEGORIES_FILE_NAME)
	log.Printf("Reading categories from file '%s'", filePath)

	f, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return []*common.Category{}
	}
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
}