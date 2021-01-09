package category

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/database"
	"vincit.fi/image-sorter/backend/util"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/common/constants"
	"vincit.fi/image-sorter/common/logger"
)

type Manager struct {
	commandLineCategories []string
	sender                api.Sender
	rootDir               string
	categoryStore         *database.CategoryStore

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

func New(params *common.Params, sender api.Sender, categoryStore *database.CategoryStore) api.CategoryManager {
	manager := Manager{
		sender:                sender,
		commandLineCategories: params.GetCategories(),
		categoryStore:         categoryStore,
	}
	return &manager
}

func (s *Manager) InitializeFromDirectory(defaultCategories []string, rootDir string) {
	var loadedCategories []*apitype.Category
	s.rootDir = filepath.Join(rootDir, constants.ImageSorterDir)

	if len(defaultCategories) > 0 && defaultCategories[0] != "" {
		logger.Info.Printf("Reading from command line parameters")
		loadedCategories = fromCategoriesStrings(defaultCategories)
	} else {
		loadedCategories = s.loadCategoriesFromFile()
	}

	for i, category := range loadedCategories {
		if category, err := s.categoryStore.AddCategory(category); err != nil {
			s.sender.SendError("Error while loading categories", err)
		} else {
			loadedCategories[i] = category
		}
	}
}

func (s *Manager) GetCategories() []*apitype.Category {
	if categories, err := s.categoryStore.GetCategories(); err != nil {
		s.sender.SendError("Cannot get categories", err)
		return nil
	} else {
		return categories
	}
}

func (s *Manager) RequestCategories() {
	s.sender.SendToTopicWithData(api.CategoriesUpdated, apitype.NewCategoriesCommand(s.GetCategories()))
}

func (s *Manager) Save(categories []*apitype.Category) {
	s.resetCategories(categories)

	s.sender.SendToTopicWithData(api.CategoriesUpdated, apitype.NewCategoriesCommand(s.GetCategories()))
}
func (s *Manager) SaveDefault(categories []*apitype.Category) {
	s.resetCategories(categories)

	if currentUser, err := user.Current(); err != nil {
		s.sender.SendError("Could not find current user", err)
	} else {
		categoryFile := filepath.Join(currentUser.HomeDir, constants.ImageSorterDir)

		if err := s.saveCategoriesToFile(categoryFile, constants.CategoriesFileName, categories); err != nil {
			s.sender.SendError("Could not save categories", err)
		} else {
			s.sender.SendToTopicWithData(api.CategoriesUpdated, apitype.NewCategoriesCommand(s.GetCategories()))
		}
	}
}

func (s *Manager) resetCategories(categories []*apitype.Category) {
	if err := s.categoryStore.ResetCategories(categories); err != nil {
		s.sender.SendError("Error while resetting categories", err)
	}
}

func (s *Manager) Close() {
	logger.Info.Print("Shutting down category manager")
}

func (s *Manager) GetCategoryById(id apitype.CategoryId) *apitype.Category {
	return s.categoryStore.GetCategoryById(id)
}

func (s *Manager) saveCategoriesToFile(fileDir string, fileName string, categories []*apitype.Category) (err error) {
	if _, err := os.Stat(fileDir); os.IsNotExist(err) {
		return os.Mkdir(fileDir, 0666)
	}

	filePath := filepath.Join(fileDir, fileName)

	logger.Info.Printf("Saving categories to file '%s'", filePath)
	if f, err := os.Create(filePath); err != nil {
		s.sender.SendError("Can't write to file", err)
	} else {
		defer func() {
			err = f.Close()
		}()

		w := bufio.NewWriter(f)
		if err = writeCategoriesToBuffer(w, categories); err != nil {
			return err
		} else if err := f.Close(); err != nil {
			return err
		}
	}
	return nil
}

func fromCategoriesStrings(categories []string) []*apitype.Category {
	var categoryEntries []*apitype.Category
	for _, categoryName := range categories {
		if len(categoryName) > 0 {
			name, subPath, shorcut := Parse(categoryName)
			categoryEntries = append(categoryEntries, apitype.NewCategory(name, subPath, shorcut))
		}
	}
	logger.Debug.Printf("Parsed %d categories", len(categoryEntries))
	for _, entry := range categoryEntries {
		logger.Trace.Printf(" - %s", entry.GetName())
	}
	return categoryEntries
}

func (s *Manager) loadCategoriesFromFile() []*apitype.Category {
	if currentUser, err := user.Current(); err == nil {
		filePaths := []string{
			filepath.Join(currentUser.HomeDir, constants.ImageSorterDir, constants.CategoriesFileName),
		}

		filePath := util.GetFirstExistingFilePath(filePaths)

		if filePath == "" {
			logger.Warn.Printf("No category files found: %s", filePaths)
			return []*apitype.Category{}
		}

		logger.Info.Printf("Reading categories from file '%s'", filePath)

		if f, err := os.OpenFile(filePath, os.O_RDONLY, 0666); err == nil {
			defer func() {
				_ = f.Close()
			}()

			return readCategoriesFromReader(f)
		} else {
			s.sender.SendError(fmt.Sprintf("Could not open category file %s ", filePath), err)
			return []*apitype.Category{}
		}
	} else {
		s.sender.SendError("Could not find current user", err)
		return []*apitype.Category{}
	}
}

func writeCategoriesToBuffer(w *bufio.Writer, categories []*apitype.Category) error {
	if _, err := w.WriteString("#version:1\n"); err != nil {
		return err
	}
	for _, category := range categories {
		if _, err := w.WriteString(fmt.Sprintf("%s\n", category.Serialize())); err != nil {
			return err
		}
	}
	return w.Flush()
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
}
