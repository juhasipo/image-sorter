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
	imageFile *ImageFile
	metaData  *ImageMetaData
}

func (s *ImageFile) IsValid() bool {
	return s != nil && s.path != ""
}

var (
	EmptyHandle          = ImageFile{id: NoImage, path: ""}
	supportedFileEndings = map[string]bool{".jpg": true, ".jpeg": true}
)

func NewHandleWithId(id ImageId, fileDir string, fileName string) *ImageFile {
	return &ImageFile{
		id:        id,
		directory: fileDir,
		filename:  fileName,
		path:      filepath.Join(fileDir, fileName),
	}
}

func NewHandle(fileDir string, fileName string) *ImageFile {
	return NewHandleWithId(NoImage, fileDir, fileName)
}

func GetEmptyHandle() *ImageFile {
	return &EmptyHandle
}

func (s *ImageFile) GetId() ImageId {
	if s != nil {
		return s.id
	} else {
		return NoImage
	}
}

func (s *ImageFile) IsPersisted() bool {
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

func (s *ImageFile) GetPath() string {
	if s != nil {
		return s.path
	} else {
		return ""
	}
}

func (s *ImageFile) GetDir() string {
	if s != nil {
		return s.directory
	} else {
		return ""
	}
}

func (s *ImageFile) GetFile() string {
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

func (s *ImageMetaData) GetMetaData() map[string]string {
	if s != nil {
		return s.metaData
	} else {
		return map[string]string{}
	}
}

func (s *ImageMetaData) SetByteSize(length int64) {
	s.byteSize = length
}

func (s *ImageMetaData) GetByteSize() int64 {
	if s != nil {
		return s.byteSize
	} else {
		return 0
	}
}

func (s *ImageMetaData) GetByteSizeMB() float64 {
	if s != nil {
		return float64(s.byteSize) / (1024.0 * 1024.0)
	} else {
		return 0.0
	}
}

func (s *ImageMetaData) GetRotation() (float64, bool) {
	return s.rotation, s.flipped
}

func NewImageFileAndMetaData(imageFile *ImageFile, metaData *ImageMetaData) *ImageFileWithMetaData {
	return &ImageFileWithMetaData{
		imageFile: imageFile,
		metaData:  metaData,
	}
}

func (s *ImageFileWithMetaData) GetImageFile() *ImageFile {
	return s.imageFile
}

func (s *ImageFileWithMetaData) GetMetaData() *ImageMetaData {
	return s.metaData
}

func (s *ImageFileWithMetaData) GetImageId() ImageId {
	return s.imageFile.GetId()
}

func LoadImageHandles(dir string) []*ImageFile {
	var handles []*ImageFile
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		logger.Error.Fatal(err)
	}

	logger.Debug.Printf("Scanning directory '%s'", dir)
	for _, file := range files {
		extension := filepath.Ext(file.Name())
		if isSupported(extension) {
			handles = append(handles, NewHandle(dir, file.Name()))
		}
	}
	logger.Debug.Printf("Found %d images", len(handles))

	return handles
}

func isSupported(extension string) bool {
	return supportedFileEndings[strings.ToLower(extension)]
}
