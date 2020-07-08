package category

import (
	"bufio"
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"log"
	"os"
	"path"
	"strings"
	"vincit.fi/image-sorter/common"
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

func (s *Entry) Serialize() string {
	shortcut := strings.ToUpper(common.KeyvalName(s.shortcuts[0]))
	return fmt.Sprintf("%s:%s:%s", s.name, s.subPath, shortcut)
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
	rootDir string
}

func FromCategoriesStrings(categories []string) []*Entry {
	var categoryEntries []*Entry
	for _, categoryName := range categories {
		if len(categoryName) > 0 {
			name, subPath, keys := Parse(categoryName)
			categoryEntries = append(categoryEntries, &Entry{
				name:      name,
				subPath:   subPath,
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

func Parse(name string) (string, string, []uint) {
	parts := strings.Split(name, ":")

	if len(parts) == 2 {
		return parts[0], parts[0], KeyToUint(parts[1])
	} else {
		return parts[0], parts[1], KeyToUint(parts[2])
	}
}
func ParseToEntry(name string) *Entry {
	name, subPath, key := Parse(name)

	return &Entry{
		name:      name,
		subPath:   subPath,
		shortcuts: key,
	}
}

func KeyToUint(key string) []uint {
	return []uint {
		gdk.KeyvalFromName(strings.ToLower(key)),
		gdk.KeyvalFromName(strings.ToUpper(key)),
	}
}

func New(sender event.Sender, categories []string, rootDir string) *Manager {
	var loadedCategories []*Entry

	if len(categories) > 0 && categories[0] != "" {
		log.Printf("Reading from command line parameters")
		loadedCategories = FromCategoriesStrings(categories)
	} else {
		loadedCategories = FromFile(rootDir)
	}

	return &Manager {
		categories: loadedCategories,
		sender: sender,
		rootDir: rootDir,
	}
}

const CATEGORIES_FILE_NAME = ".categories"

func FromFile(fileDir string) []*Entry {
	filePath := path.Join(fileDir, CATEGORIES_FILE_NAME)
	log.Printf("Reading categories from file '%s'", filePath)

	f, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return []*Entry{}
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if lines != nil {
		return FromCategoriesStrings(lines[1:])
	} else {
		return []*Entry{}
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
	saveCategoriesToFile(s.rootDir, CATEGORIES_FILE_NAME, categories)
	s.sender.SendToTopicWithData(event.CATEGORIES_UPDATED, &CategoriesCommand{
		categories: categories,
	})
}

func (s *Manager) SaveDefault(categories []*Entry) {
	s.categories = categories
	// TODO: Find user's home dir
	saveCategoriesToFile(s.rootDir, CATEGORIES_FILE_NAME, categories)
	s.sender.SendToTopicWithData(event.CATEGORIES_UPDATED, &CategoriesCommand{
		categories: categories,
	})
}

func saveCategoriesToFile(fileDir string, fileName string, categories []*Entry) {
	filePath := path.Join(fileDir, fileName)
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
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
