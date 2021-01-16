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
	byteSize int64
	rotation float64
	flipped  bool
	metaData map[string]string
}

type ImageFile struct {
	id        ImageId
	directory string
	filename  string
	path      string
}

type ImageFileWithMetaData struct {
	ImageFile
	ImageMetaData
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
	}
}

func NewImageFile(fileDir string, fileName string) *ImageFile {
	return NewImageFileWithId(NoImage, fileDir, fileName)
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

func NewImageMetaData(byteSize int64, rotation float64, flipped bool, metaData map[string]string) *ImageMetaData {
	return &ImageMetaData{
		byteSize: byteSize,
		rotation: rotation,
		flipped:  flipped,
		metaData: metaData,
	}
}

func (s *ImageMetaData) MetaData() map[string]string {
	if s != nil {
		return s.metaData
	} else {
		return map[string]string{}
	}
}

func (s *ImageMetaData) SetByteSize(length int64) {
	s.byteSize = length
}

func (s *ImageMetaData) ByteSize() int64 {
	if s != nil {
		return s.byteSize
	} else {
		return 0
	}
}

func (s *ImageMetaData) ByteSizeInMB() float64 {
	if s != nil {
		return float64(s.byteSize) / (1024.0 * 1024.0)
	} else {
		return 0.0
	}
}

func (s *ImageMetaData) Rotation() (float64, bool) {
	return s.rotation, s.flipped
}

func NewImageFileAndMetaData(imageFile *ImageFile, metaData *ImageMetaData) *ImageFileWithMetaData {
	return &ImageFileWithMetaData{
		ImageFile:     *imageFile,
		ImageMetaData: *metaData,
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
