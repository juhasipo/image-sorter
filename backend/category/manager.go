package category

import (
	"strings"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/database"
	"vincit.fi/image-sorter/common"
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

func NewCategoryManager(params *common.Params, sender api.Sender, categoryStore *database.CategoryStore) api.CategoryManager {
	return newManager(params, sender, categoryStore)
}

// For tests where some private methods are tested
func newManager(params *common.Params, sender api.Sender, categoryStore *database.CategoryStore) *Manager {
	manager := Manager{
		sender:                sender,
		commandLineCategories: params.Categories(),
		categoryStore:         categoryStore,
	}
	return &manager
}

func (s *Manager) InitializeFromDirectory(cmdLineCategories []string, dbCategories []*apitype.Category) {
	var loadedCategories []*apitype.Category

	if len(cmdLineCategories) > 0 && cmdLineCategories[0] != "" {
		logger.Info.Printf("Reading from command line parameters")
		loadedCategories = fromCategoriesStrings(cmdLineCategories)
	} else {
		loadedCategories = dbCategories
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
	s.sender.SendCommandToTopic(
		api.CategoriesUpdated,
		&api.UpdateCategoriesCommand{Categories: s.GetCategories()},
	)
}

func (s *Manager) Save(command *api.SaveCategoriesCommand) {
	if err := s.categoryStore.ResetCategories(command.Categories); err != nil {
		s.sender.SendError("Error while resetting categories", err)
	}
	s.sender.SendCommandToTopic(
		api.CategoriesUpdated,
		&api.UpdateCategoriesCommand{Categories: s.GetCategories()},
	)
}

func (s *Manager) Close() {
	logger.Info.Print("Shutting down category manager")
}

func (s *Manager) GetCategoryById(query *api.CategoryQuery) *apitype.Category {
	return s.categoryStore.GetCategoryById(query.Id)
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
		logger.Trace.Printf(" - %s", entry.Name())
	}
	return categoryEntries
}
