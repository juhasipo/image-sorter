package common

import (
	"image"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"vincit.fi/image-sorter/duplo"
)

type ImageContainer struct {
	handle *Handle
	img    image.Image
}

func (s *ImageContainer) String() string {
	return "ImageContainer{" + s.handle.GetId() + "}"
}

func (s *ImageContainer) GetHandle() *Handle {
	return s.handle
}

func (s *ImageContainer) GetImage() image.Image {
	return s.img
}

func ImageContainerNew(handle *Handle, img image.Image) *ImageContainer {
	return &ImageContainer{
		handle: handle,
		img:    img,
	}
}

type Handle struct {
	id        string
	directory string
	filename  string
	path      string
	hash      *duplo.Hash
	byteSize  int64
}

func (s *Handle) IsValid() bool {
	return s != nil && s.id != ""
}

var (
	EMPTY_HANDLE         = Handle{id: "", path: ""}
	supportedFileEndings = map[string]bool{".jpg": true, ".jpeg": true}
)

func HandleNew(fileDir string, fileName string) *Handle {
	return &Handle{
		id:        fileName,
		directory: fileDir,
		filename:  fileName,
		path:      filepath.Join(fileDir, fileName),
		hash:      nil,
	}
}

func GetEmptyHandle() *Handle {
	return &EMPTY_HANDLE
}

func (s *Handle) GetId() string {
	return s.id
}

func (s *Handle) String() string {
	return s.id
}

func (s *Handle) GetPath() string {
	return s.path
}

func (s *Handle) GetDir() string {
	return s.directory
}

func (s *Handle) GetFile() string {
	return s.filename
}

func (s *Handle) SetHash(hash *duplo.Hash) {
	s.hash = hash
}
func (s *Handle) GetHash() *duplo.Hash {
	return s.hash
}

func (s *Handle) SetByteSize(length int64) {
	s.byteSize = length
}

func (s *Handle) GetByteSize() int64 {
	return s.byteSize
}

func (s *Handle) GetByteSizeMB() float64 {
	return float64(s.byteSize) / (1024.0 * 1024.0)
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
			handles = append(handles, HandleNew(dir, file.Name()))
		}
	}
	log.Printf("Found %d images", len(handles))

	return handles
}

func isSupported(extension string) bool {
	return supportedFileEndings[strings.ToLower(extension)]
}
