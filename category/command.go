package category

import (
	"fmt"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
)

type CategorizeCommand struct {
	handle     *common.Handle
	entry      *common.Category
	operation  common.Operation
	stayOnSameImage bool

	event.Command
}

type CategoriesCommand struct {
	categories []*common.Category

	event.Command
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

func CategorizeCommandNew(handle *common.Handle, entry *common.Category, operation common.Operation) *CategorizeCommand {
	return &CategorizeCommand{
		handle:    handle,
		entry:     entry,
		operation: operation,
	}
}

func CategorizeCommandNewWithStayAttr(handle *common.Handle, entry *common.Category, operation common.Operation, stayOnSameImage bool) *CategorizeCommand {
	return &CategorizeCommand{
		handle:    handle,
		entry:     entry,
		operation: operation,
		stayOnSameImage: stayOnSameImage,
	}
}

func (s *CategorizeCommand) String() string {
	return fmt.Sprintf("CategorizeCommand{%s:%s:%d}",
		s.handle.GetId(), s.entry.GetName(), s.operation)
}
