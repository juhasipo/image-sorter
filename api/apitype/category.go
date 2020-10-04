package apitype

import (
	"fmt"
	"strings"
	"time"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/common/event"
)

type Category struct {
	id       string
	name     string
	subPath  string
	shortcut uint
}

func NewCategory(name string, subPath string, shortcut string) *Category {
	return &Category{
		id:       name,
		name:     name,
		subPath:  subPath,
		shortcut: common.KeyToUint(shortcut),
	}
}

func (s *Category) GetId() string {
	return s.id
}

func (s *Category) GetSubPath() string {
	return s.subPath
}

func (s *Category) GetName() string {
	return s.name
}

func (s *Category) String() string {
	return s.name
}

func (s *Category) GetShortcut() uint {
	return s.shortcut
}

func (s *Category) GetShortcutAsString() string {
	return common.KeyvalName(s.shortcut)
}

func (s *Category) HasShortcut(val uint) bool {
	return s.shortcut == val
}

func (s *Category) Serialize() string {
	shortcut := strings.ToUpper(common.KeyvalName(s.shortcut))
	return fmt.Sprintf("%s:%s:%s", s.name, s.subPath, shortcut)
}

type CategorizeCommand struct {
	handle          *Handle
	entry           *Category
	operation       Operation
	stayOnSameImage bool
	nextImageDelay  time.Duration
	forceToCategory bool

	event.Command
}

type CategoriesCommand struct {
	categories []*Category

	event.Command
}

func NewCategoriesCommand(categories []*Category) *CategoriesCommand {
	return &CategoriesCommand{
		categories: categories,
	}
}

func (s *CategoriesCommand) GetCategories() []*Category {
	return s.categories
}

func (s *CategoriesCommand) String() string {
	return fmt.Sprintf("CategoriesCommand{%s}",
		s.categories)
}

func (s *CategorizeCommand) GetHandle() *Handle {
	return s.handle
}
func (s *CategorizeCommand) GetEntry() *Category {
	return s.entry
}
func (s *CategorizeCommand) GetOperation() Operation {
	return s.operation
}
func (s *CategorizeCommand) ShouldStayOnSameImage() bool {
	return s.stayOnSameImage
}
func (s *CategorizeCommand) ShouldForceToCategory() bool {
	return s.forceToCategory
}
func (s *CategorizeCommand) GetNextImageDelay() time.Duration {
	return s.nextImageDelay
}

func NewCategorizeCommand(handle *Handle, entry *Category, operation Operation) *CategorizeCommand {
	return &CategorizeCommand{
		handle:    handle,
		entry:     entry,
		operation: operation,
	}
}

func (s *CategorizeCommand) String() string {
	return fmt.Sprintf("CategorizeCommand{%s:%s:%d}",
		s.handle.GetId(), s.entry.GetName(), s.operation)
}

func (s *CategorizeCommand) SetStayOfSameImage(stayOnSameImage bool) {
	s.stayOnSameImage = stayOnSameImage
}

func (s *CategorizeCommand) SetForceToCategory(forceToCategory bool) {
	s.forceToCategory = forceToCategory
}

func (s *CategorizeCommand) SetNextImageDelay(duration time.Duration) {
	s.nextImageDelay = duration
}

type CategorizedImage struct {
	category  *Category
	operation Operation
}

func NewCategorizedImage(entry *Category, operation Operation) *CategorizedImage {
	return &CategorizedImage{
		category:  entry,
		operation: operation,
	}
}

func (s *CategorizedImage) GetOperation() Operation {
	return s.operation
}

func (s *CategorizedImage) SetOperation(operation Operation) {
	s.operation = operation
}

func (s *CategorizedImage) GetEntry() *Category {
	return s.category
}

func (s *CategorizeCommand) ToLabel() string {
	return s.GetEntry().GetName()
}
