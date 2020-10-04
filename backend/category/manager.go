package category

import (
	"bufio"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/common/constants"
	"vincit.fi/image-sorter/common/event"
	"vincit.fi/image-sorter/common/logger"
	"vincit.fi/image-sorter/common/util"
)

type Manager struct {
	commandLineCategories []string
	categories            []*apitype.Category
	categoriesById        map[string]*apitype.Category
	sender                event.Sender
	rootDir               string

	api.CategoryManager
}

func Parse(value string) (name string, path string, shortcut string) {
	parts := strings.Split(value, ":")

	if len(parts) == 2 {
		name = parts[0]
		path = parts[0]
		shortcut = parts[1]
	} else {
		name = parts[0]
		path = parts[1]
		shortcut = parts[2]
	}
	return
}

func New(params *util.Params, sender event.Sender) api.CategoryManager {
	manager := Manager{
		sender:                sender,
		commandLineCategories: params.GetCategories(),
	}
	return &manager
}

func (s *Manager) InitializeFromDirectory(defaultCategories []string, rootDir string) {
	var loadedCategories []*apitype.Category
	var categoriesByName = map[string]*apitype.Category{}
	s.rootDir = filepath.Join(rootDir, constants.ImageSorterDir)

	if len(defaultCategories) > 0 && defaultCategories[0] != "" {
		logger.Info.Printf("Reading from command line parameters")
		loadedCategories = fromCategoriesStrings(defaultCategories)
	} else {
		loadedCategories = loadCategoriesFromFile(s.rootDir)
	}

	for _, category := range loadedCategories {
		categoriesByName[category.GetName()] = category
	}

	s.categories = loadedCategories
	s.categoriesById = categoriesByName
}

func (s *Manager) GetCategories() []*apitype.Category {
	return s.categories
}

func (s *Manager) RequestCategories() {
	s.sender.SendToTopicWithData(event.CategoriesUpdated, apitype.NewCategoriesCommand(s.categories))
}

func (s *Manager) Save(categories []*apitype.Category) {
	s.resetCategories(categories)

	saveCategoriesToFile(s.rootDir, constants.CategoriesFileName, categories)
	s.sender.SendToTopicWithData(event.CategoriesUpdated, apitype.NewCategoriesCommand(s.categories))
}
func (s *Manager) SaveDefault(categories []*apitype.Category) {
	s.resetCategories(categories)

	if currentUser, err := user.Current(); err != nil {
		logger.Error.Println("Could not find current user", err)
	} else {
		categoryFile := filepath.Join(currentUser.HomeDir, constants.ImageSorterDir)

		saveCategoriesToFile(categoryFile, constants.CategoriesFileName, categories)
		s.sender.SendToTopicWithData(event.CategoriesUpdated, apitype.NewCategoriesCommand(s.categories))
	}
}

func (s *Manager) resetCategories(categories []*apitype.Category) {
	s.categories = categories
	for _, category := range categories {
		s.categoriesById[category.GetId()] = category
	}
}

func (s *Manager) Close() {
	logger.Info.Print("Shutting down category manager")
	saveCategoriesToFile(s.rootDir, constants.CategoriesFileName, s.categories)
}

func (s *Manager) GetCategoryById(id string) *apitype.Category {
	return s.categoriesById[id]
}

func saveCategoriesToFile(fileDir string, fileName string, categories []*apitype.Category) {
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

	writeCategoriesToBuffer(w, categories)
}

func fromCategoriesStrings(categories []string) []*apitype.Category {
	var categoryEntries []*apitype.Category
	for _, categoryName := range categories {
		if len(categoryName) > 0 {
			categoryEntries = append(categoryEntries, apitype.NewCategory(Parse(categoryName)))
		}
	}
	logger.Debug.Printf("Parsed %d categories", len(categoryEntries))
	for _, entry := range categoryEntries {
		logger.Trace.Printf(" - %s", entry.GetName())
	}
	return categoryEntries
}

func loadCategoriesFromFile(fileDir string) []*apitype.Category {
	if currentUser, err := user.Current(); err == nil {
		filePaths := []string{
			filepath.Join(fileDir, constants.CategoriesFileName),
			filepath.Join(currentUser.HomeDir, constants.ImageSorterDir, constants.CategoriesFileName),
		}

		filePath := common.GetFirstExistingFilePath(filePaths)

		logger.Info.Printf("Reading categories from file '%s'", filePath)

		if f, err := os.OpenFile(filePath, os.O_RDONLY, 0666); err == nil {
			defer f.Close()
			return readCategoriesFromReader(f)
		} else {
			logger.Error.Println("Could not open file: "+filePath, err)
			return []*apitype.Category{}
		}
	} else {
		logger.Error.Println("Could not find current user", err)
		return []*apitype.Category{}
	}
}

func writeCategoriesToBuffer(w *bufio.Writer, categories []*apitype.Category) {
	w.WriteString("#version:1")
	w.WriteString("\n")
	for _, category := range categories {
		w.WriteString(category.Serialize())
		w.WriteString("\n")
	}
	w.Flush()
}

func readCategoriesFromReader(f io.Reader) []*apitype.Category {
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if lines != nil {
		return fromCategoriesStrings(lines[1:])
	} else {
		return []*apitype.Category{}
	}
	return nil
}
