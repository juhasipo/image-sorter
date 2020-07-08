package category

import (
	"github.com/gotk3/gotk3/gdk"
	"log"
	"strings"
	"vincit.fi/image-sorter/event"
)

type Operation int
const(
	NONE Operation = 0
	MOVE Operation = 1
	COPY Operation = 2
)

func (s *CategorizeCommand) ToLabel() string {
	entryName := s.GetEntry().GetName()
	status := ""
	switch s.GetOperation() {
	case COPY: status = "C"
	case MOVE: status = "M"
	}

	if status != "" {
		return entryName + " (" + status + ")"
	} else {
		return entryName
	}
}

func (s Operation) NextOperation() Operation {
	return (s + 1) % 3
}

type Entry struct {
	name string
	subPath string
	shortcuts []uint
}

func CategoryEntryNew(name string, subPath string, shortcut string) *Entry {
	return &Entry{
		name:      name,
		subPath:   subPath,
		shortcuts: KeyToUint(shortcut),
	}
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
				subPath:   name,
				shortcuts: keys,
			})
		}
	}
	log.Printf("Parsed %d categories:", len(categoryEntries))
	for _, entry := range categoryEntries {
		log.Printf(" - %s", entry.name)
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

func (s *Manager) Save(categories []*Entry) {
	s.categories = categories
	s.sender.SendToTopicWithData(event.CATEGORIES_UPDATED, &CategoriesCommand{
		categories: categories,
	})
}

func (s *Manager) SaveDefault(categories []*Entry) {
	s.Save(categories)
}
