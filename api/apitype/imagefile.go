package apitype

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"vincit.fi/image-sorter/common/logger"
)

type ImageId int64

const NoImage = ImageId(-1)

type ImageMetaData struct {
	data map[string]string
}

type ImageFile struct {
	id        ImageId
	directory string
	filename  string
	path      string
	byteSize  int64
	rotation  float64
	flipped   bool
}

func (s *ImageFile) IsValid() bool {
	return s != nil && s.path != ""
}

var (
	EmptyImageFile       = ImageFile{id: NoImage, path: ""}
	supportedFileEndings = map[string]bool{".jpg": true, ".jpeg": true}
)

func NewImageFileWithId(id ImageId, fileDir string, fileName string) *ImageFile {
	return &ImageFile{
		id:        id,
		directory: fileDir,
		filename:  fileName,
		path:      filepath.Join(fileDir, fileName),
		byteSize:  0,
		rotation:  0.0,
		flipped:   false,
	}
}

func NewImageFileWithIdSizeAndOrientation(id ImageId, fileDir string, fileName string, byteSize int64, rotation float64, flipped bool) *ImageFile {
	return &ImageFile{
		id:        id,
		directory: fileDir,
		filename:  fileName,
		path:      filepath.Join(fileDir, fileName),
		byteSize:  byteSize,
		rotation:  rotation,
		flipped:   flipped,
	}
}

func NewImageFile(fileDir string, fileName string) *ImageFile {
	return NewImageFileWithIdSizeAndOrientation(NoImage, fileDir, fileName, 0, 0.0, false)
}

func GetEmptyImageFile() *ImageFile {
	return &EmptyImageFile
}

func (s *ImageFile) Id() ImageId {
	if s != nil {
		return s.id
	} else {
		return NoImage
	}
}

func (s *ImageFile) Persisted() bool {
	if s != nil {
		return s.id > 0
	} else {
		return false
	}
}

func (s *ImageFile) String() string {
	if s != nil {
		if s.IsValid() {
			return "ImageFile{" + s.filename + "}"
		} else {
			return "ImageFile<invalid>"
		}
	} else {
		return "ImageFile<nil>"
	}
}

func (s *ImageFile) Path() string {
	if s != nil {
		return s.path
	} else {
		return ""
	}
}

func (s *ImageFile) Directory() string {
	if s != nil {
		return s.directory
	} else {
		return ""
	}
}

func (s *ImageFile) FileName() string {
	if s != nil {
		return s.filename
	} else {
		return ""
	}
}

func (s *ImageFile) SetMetaData(byteSize int64, rotation float64, flipped bool) {
	s.byteSize = byteSize
	s.rotation = rotation
	s.flipped = flipped
}

func (s *ImageFile) ByteSize() int64 {
	if s != nil {
		return s.byteSize
	} else {
		return 0
	}
}

func (s *ImageFile) ByteSizeInMB() float64 {
	if s != nil {
		return float64(s.byteSize) / (1024.0 * 1024.0)
	} else {
		return 0.0
	}
}

func (s *ImageFile) Rotation() (float64, bool) {
	return s.rotation, s.flipped
}

func NewImageMetaData(data map[string]string) *ImageMetaData {
	return &ImageMetaData{
		data: data,
	}
}

func NewInvalidImageMetaData() *ImageMetaData {
	return &ImageMetaData{
		data: map[string]string{},
	}
}

func (s *ImageMetaData) MetaData() map[string]string {
	if s != nil {
		return s.data
	} else {
		return map[string]string{}
	}
}

func LoadImageFiles(dir string) []*ImageFile {
	var imageFiles []*ImageFile
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		logger.Error.Fatal(err)
	}

	logger.Debug.Printf("Scanning directory '%s'", dir)
	for _, file := range files {
		extension := filepath.Ext(file.Name())
		if isSupported(extension) {
			imageFiles = append(imageFiles, NewImageFile(dir, file.Name()))
		}
	}
	logger.Debug.Printf("Found %d images", len(imageFiles))

	return imageFiles
}

func isSupported(extension string) bool {
	return supportedFileEndings[strings.ToLower(extension)]
}
