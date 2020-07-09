package common

import (
	"fmt"
	"strings"
	duplo "vincit.fi/image-sorter/duplo"
)

type Handle struct {
	id string
	path string
	hash *duplo.Hash
}

func (s *Handle) IsValid() bool {
	return s != nil && s.id != ""
}

var (
	EMPTY_HANDLE = Handle {id: "", path: ""}
)

func GetEmptyHandle() *Handle {
	return &EMPTY_HANDLE
}

func (s* Handle) GetId() string {
	return s.id
}

func (s* Handle) String() string {
	return s.id
}

func (s* Handle) GetPath() string {
	return s.path
}

func (s *Handle) SetHash(hash *duplo.Hash) {
	s.hash = hash
}
func (s *Handle) GetHash() *duplo.Hash {
	return s.hash
}



type Operation int
const(
	NONE Operation = 0
	MOVE Operation = 1
)

func (s Operation) NextOperation() Operation {
	return (s + 1) % 2
}


type Category struct {
	id string
	name string
	subPath string
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

func (s*Category) GetShortcuts() []uint {
	return s.shortcuts
}

func (s*Category) HasShortcut(val uint) bool {
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
