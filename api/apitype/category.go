package apitype

import (
	"fmt"
	"strings"
	"vincit.fi/image-sorter/common"
)

type CategoryId int64

const NoCategory = CategoryId(-1)

type Category struct {
	id       CategoryId
	name     string
	subPath  string
	shortcut uint
}

func NewPersistedCategory(id CategoryId, category *Category) *Category {
	return &Category{
		id:       id,
		name:     category.name,
		subPath:  category.subPath,
		shortcut: category.shortcut,
	}
}

func NewCategoryWithId(id CategoryId, name string, subPath string, shortcut string) *Category {
	return &Category{
		id:       id,
		name:     name,
		subPath:  subPath,
		shortcut: common.KeyToUint(shortcut),
	}
}

func NewCategory(name string, subPath string, shortcut string) *Category {
	return NewCategoryWithId(NoCategory, name, subPath, shortcut)
}

func (s *Category) GetId() CategoryId {
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
