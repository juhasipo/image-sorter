package category

import (
	"github.com/gotk3/gotk3/gdk"
	"strings"
	"vincit.fi/image-sorter/event"
)

type Operation int
const(
	NONE Operation = 0
	MOVE Operation = 1
	COPY Operation = 2
)

type Entry struct {
	name string
	subPath string
	shortcuts []uint
}

func (s *Entry) GetSubPath() string {
	return s.subPath
}

func (s *Entry) GetName() string {
	return s.name
}

func (s* Entry) GetShortcuts() []uint {
	return s.shortcuts
}

func (s* Entry) HasShortcut(val uint) bool {
	for _, shortcut := range s.shortcuts {
		if shortcut == val {
			return true
		}
	}
	return false
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
	for _, categoryName := range categories {
		if len(categoryName) > 0 {
			name, keys := Parse(categoryName)
			categoryEntries = append(categoryEntries, &Entry{
				name:      name,
				subPath:   categoryName,
				shortcuts: keys,
			})
		}
	}
	return categoryEntries
}

func Parse(name string) (string, []uint) {
	parts := strings.Split(name, ":")

	return parts[0], KeyToUint(parts[1])
}

func KeyToUint(key string) []uint {
	return []uint {
		gdk.KeyvalFromName(strings.ToLower(key)),
		gdk.KeyvalFromName(strings.ToUpper(key)),
	}
}

func New(sender event.Sender, categories []string) *Manager {
	return &Manager {
		categories: FromCategories(categories),
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
