package common

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"vincit.fi/image-sorter/duplo"
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
	supportedFileEndings = map[string]bool{".jpg": true, ".jpeg": true}
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

func LoadImageHandles(dir string) []*Handle {
	var handles []*Handle
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Scanning directory '%s'", dir)
	for _, file := range files {
		extension := filepath.Ext(file.Name())
		if isSupported(extension) {
			handles = append(handles, HandleNew(dir, file.Name(), extension, 0, 0))
		}
	}
	log.Printf("Found %d images", len(handles))

	return handles
}

func isSupported(extension string) bool {
	return supportedFileEndings[strings.ToLower(extension)]
}

