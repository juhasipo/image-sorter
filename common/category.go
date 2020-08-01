package common

import (
	"fmt"
	"strings"
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
		shortcut: KeyToUint(shortcut),
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
	return KeyvalName(s.shortcut)
}

func (s *Category) HasShortcut(val uint) bool {
	return s.shortcut == val
}

func (s *Category) Serialize() string {
	shortcut := strings.ToUpper(KeyvalName(s.shortcut))
	return fmt.Sprintf("%s:%s:%s", s.name, s.subPath, shortcut)
}
