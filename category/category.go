package category

import (
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
)

type CategorizeCommand struct {
	handle     *common.Handle
	entry      *Entry
	operation  Operation

	event.Command
}

type CategoriesCommand struct {
	categories []*Entry

	event.Command
}

func (s *CategoriesCommand) GetCategories() []*Entry {
	return s.categories
}

func (s *CategorizeCommand) GetHandle() *common.Handle {
	return s.handle
}
func (s *CategorizeCommand) GetEntry() *Entry {
	return s.entry
}
func (s *CategorizeCommand) GetOperation() Operation {
	return s.operation
}

func CategorizeCommandNew(handle *common.Handle, entry *Entry, operation Operation) *CategorizeCommand {
	return &CategorizeCommand{
		handle:    handle,
		entry:     entry,
		operation: operation,
	}
}

type Operation int
const(
	NONE Operation = 0
	COPY Operation = 1
	MOVE Operation = 2
)

type Entry struct {
	name string
	subPath string
}

func (s *Entry) GetName() string {
	return s.name
}

type CategorizedImage struct {
	category *Entry
	operation Operation
}

func CategorizedImageNew(entry *Entry, operation Operation) *CategorizedImage {
	return &CategorizedImage {
		category: entry,
		operation: operation,
	}
}

func (s* CategorizedImage) GetOperation() Operation {
	return s.operation
}

func (s* CategorizedImage) SetOperation(operation Operation) {
	s.operation = operation
}

func (s* CategorizedImage) GetEntry() *Entry {
	return s.category
}

type Manager struct {
	categories []*Entry
	sender event.Sender
}

func FromCategories(categories []string) []*Entry {
	var categoryEntries []*Entry
	for i := range categories {
		categoryName := categories[i]
		categoryEntries = append(categoryEntries, &Entry {
			name: categoryName,
			subPath: categoryName,
		})
	}
	return categoryEntries
}

func New(sender event.Sender) *Manager {
	return &Manager {
		categories: FromCategories([]string {"Good", "Maybe", "Bad"}),
		sender: sender,
	}
}

func (s *Manager) AddCategory(name string, subPath string) *Entry {
	category := Entry {name: name, subPath: subPath}
	s.categories = append(s.categories, &category)
	return &category
}

func (s *Manager) GetCategories() []*Entry {
	return s.categories
}

func (s *Manager) RequestCategories() {
	s.sender.SendToTopicWithData(event.CATEGORIES_UPDATED, &CategoriesCommand{
		categories: s.categories,
	})
}
