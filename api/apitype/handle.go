package apitype

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"vincit.fi/image-sorter/common/logger"
)

type ImageId int64

const NoImage = ImageId(-1)

type Handle struct {
	id        ImageId
	directory string
	filename  string
	path      string
	byteSize  int64
	rotation  float64
	flipped   bool
	metaData  map[string]string
}

func (s *Handle) IsValid() bool {
	return s != nil && s.path != ""
}

var (
	EmptyHandle          = Handle{id: NoImage, path: ""}
	supportedFileEndings = map[string]bool{".jpg": true, ".jpeg": true}
)

func NewPersistedHandle(id ImageId, handle *Handle, metaData map[string]string) *Handle {
	return &Handle{
		id:        id,
		directory: handle.directory,
		filename:  handle.filename,
		path:      handle.path,
		byteSize:  handle.byteSize,
		metaData:  metaData,
	}
}

func NewHandleWithId(id ImageId, fileDir string, fileName string, rotation float64, flipped bool, metaData map[string]string) *Handle {
	return &Handle{
		id:        id,
		directory: fileDir,
		filename:  fileName,
		path:      filepath.Join(fileDir, fileName),
		metaData:  metaData,
		rotation:  rotation,
		flipped:   flipped,
	}
}

func NewHandle(fileDir string, fileName string) *Handle {
	return NewHandleWithId(NoImage, fileDir, fileName, 0, false, map[string]string{})
}

func NewHandleWithMetaData(fileDir string, fileName string, metaData map[string]string) *Handle {
	return NewHandleWithId(NoImage, fileDir, fileName, 0, false, metaData)
}

func GetEmptyHandle() *Handle {
	return &EmptyHandle
}

func (s *Handle) GetId() ImageId {
	if s != nil {
		return s.id
	} else {
		return NoImage
	}
}

func (s *Handle) IsPersisted() bool {
	if s != nil {
		return s.id > 0
	} else {
		return false
	}
}

func (s *Handle) String() string {
	if s != nil {
		if s.IsValid() {
			return "Handle{" + s.filename + "}"
		} else {
			return "Handle<invalid>"
		}
	} else {
		return "Handle<nil>"
	}
}

func (s *Handle) GetPath() string {
	if s != nil {
		return s.path
	} else {
		return ""
	}
}

func (s *Handle) GetDir() string {
	if s != nil {
		return s.directory
	} else {
		return ""
	}
}

func (s *Handle) GetFile() string {
	if s != nil {
		return s.filename
	} else {
		return ""
	}
}

func (s *Handle) GetMetaData() map[string]string {
	if s != nil {
		return s.metaData
	} else {
		return map[string]string{}
	}
}

func (s *Handle) SetByteSize(length int64) {
	s.byteSize = length
}

func (s *Handle) GetByteSize() int64 {
	if s != nil {
		return s.byteSize
	} else {
		return 0
	}
}

func (s *Handle) GetByteSizeMB() float64 {
	if s != nil {
		return float64(s.byteSize) / (1024.0 * 1024.0)
	} else {
		return 0.0
	}
}

func (s *Handle) GetRotation() (float64, bool) {
	return s.rotation, s.flipped
}

func LoadImageHandles(dir string) []*Handle {
	var handles []*Handle
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
