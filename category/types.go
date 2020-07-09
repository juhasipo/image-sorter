package category

import (
	"vincit.fi/image-sorter/common"
)

type CategorizedImage struct {
	category *common.Category
	operation common.Operation
}

func CategorizedImageNew(entry *common.Category, operation common.Operation) *CategorizedImage {
	return &CategorizedImage {
		category: entry,
		operation: operation,
	}
}

func (s* CategorizedImage) GetOperation() common.Operation {
	return s.operation
}

func (s* CategorizedImage) SetOperation(operation common.Operation) {
	s.operation = operation
}

func (s* CategorizedImage) GetEntry() *common.Category {
	return s.category
}

func (s *CategorizeCommand) ToLabel() string {
	return s.GetEntry().GetName()
}
