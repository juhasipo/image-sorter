package category

import (
	"fmt"
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

func (s *CategoriesCommand) String() string {
	return fmt.Sprintf("CategoriesCommand{%s}",
		s.categories)
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

func (s *CategorizeCommand) String() string {
	return fmt.Sprintf("CategorizeCommand{%s:%s:%d}",
		s.handle.GetId(), s.entry.GetName(), s.operation)
}
