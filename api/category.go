package api

import (
	"fmt"
	"time"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/common/event"
)

type CategoryManager interface {
	InitializeFromDirectory(categories []string, rootDir string)
	GetCategories() []*common.Category
	RequestCategories()
	Save(categories []*common.Category)
	SaveDefault(categories []*common.Category)
	Close()
	GetCategoryById(id string) *common.Category
}

type CategorizeCommand struct {
	handle          *common.Handle
	entry           *common.Category
	operation       common.Operation
	stayOnSameImage bool
	nextImageDelay  time.Duration
	forceToCategory bool

	event.Command
}

type CategoriesCommand struct {
	categories []*common.Category

	event.Command
}

func NewCategoriesCommand(categories []*common.Category) *CategoriesCommand {
	return &CategoriesCommand{
		categories: categories,
	}
}

func (s *CategoriesCommand) GetCategories() []*common.Category {
	return s.categories
}

func (s *CategoriesCommand) String() string {
	return fmt.Sprintf("CategoriesCommand{%s}",
		s.categories)
}

func (s *CategorizeCommand) GetHandle() *common.Handle {
	return s.handle
}
func (s *CategorizeCommand) GetEntry() *common.Category {
	return s.entry
}
func (s *CategorizeCommand) GetOperation() common.Operation {
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

func NewCategorizeCommand(handle *common.Handle, entry *common.Category, operation common.Operation) *CategorizeCommand {
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
	category  *common.Category
	operation common.Operation
}

func NewCategorizedImage(entry *common.Category, operation common.Operation) *CategorizedImage {
	return &CategorizedImage{
		category:  entry,
		operation: operation,
	}
}

func (s *CategorizedImage) GetOperation() common.Operation {
	return s.operation
}

func (s *CategorizedImage) SetOperation(operation common.Operation) {
	s.operation = operation
}

func (s *CategorizedImage) GetEntry() *common.Category {
	return s.category
}

func (s *CategorizeCommand) ToLabel() string {
	return s.GetEntry().GetName()
}
