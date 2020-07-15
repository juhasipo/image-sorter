package common

import (
	"fmt"
	"path/filepath"
	"strings"
	duplo "vincit.fi/image-sorter/duplo"
)

type Handle struct {
	id        string
	path      string
	hash      *duplo.Hash
	imageType string
	width     int
	height    int
	byteSize  int64
}

func (s *Handle) IsValid() bool {
	return s != nil && s.id != ""
}

var (
	EMPTY_HANDLE = Handle {id: "", path: ""}
)

func HandleNew(fileDir string, fileName string, imageType string, width int, height int) *Handle {
	return &Handle{
		id:   fileName,
		path: filepath.Join(fileDir, fileName),
		hash: nil,
		imageType: imageType,
		width: width,
		height: height,
	}
}

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

func (s* Handle) GetImageType() string {
	return s.imageType
}

func (s* Handle) GetWidth() int {
	return s.width
}

func (s* Handle) GetHeight() int {
	return s.height
}

func (s *Handle) SetHash(hash *duplo.Hash) {
	s.hash = hash
}
func (s *Handle) GetHash() *duplo.Hash {
	return s.hash
}

func (s *Handle) SetSize(width int, height int) {
	s.width = width
	s.height = height
}

func (s *Handle) SetByteSize(length int64) {
	s.byteSize = length
}

func (s *Handle) GetByteSize() int64 {
	return s.byteSize
}

func (s *Handle) GetByteSizeMB() float64 {
	return float64(s.byteSize)/(1024.0*1024.0)
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
