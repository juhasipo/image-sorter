package common

import (
	"fmt"
	"strings"
)

type Category struct {
	id        string
	name      string
	subPath   string
	shortcuts []uint
}

func CategoryEntryNew(name string, subPath string, shortcut string) *Category {
	return &Category{
		id:        name,
		name:      name,
		subPath:   subPath,
		shortcuts: KeyToUint(shortcut),
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

func (s *Category) GetShortcuts() []uint {
	return s.shortcuts
}

func (s *Category) HasShortcut(val uint) bool {
	for _, shortcut := range s.shortcuts {
		if shortcut == val {
			return true
		}
	}
	return false
}

func (s *Category) Serialize() string {
	shortcut := strings.ToUpper(KeyvalName(s.shortcuts[0]))
	return fmt.Sprintf("%s:%s:%s", s.name, s.subPath, shortcut)
}
