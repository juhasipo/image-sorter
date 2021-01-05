package apitype

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"vincit.fi/image-sorter/common/logger"
)

type HandleId int64

type Handle struct {
	id        HandleId
	directory string
	filename  string
	path      string
	byteSize  int64
}

func (s *Handle) IsValid() bool {
	return s != nil && s.path != ""
}

var (
	EmptyHandle          = Handle{id: -1, path: ""}
	supportedFileEndings = map[string]bool{".jpg": true, ".jpeg": true}
)

func NewPersistedHandle(id HandleId, handle *Handle) *Handle {
	return &Handle{
		id:        id,
		directory: handle.directory,
		filename:  handle.filename,
		path:      handle.path,
		byteSize:  handle.byteSize,
	}
}

func NewHandleWithId(id HandleId, fileDir string, fileName string) *Handle {
	return &Handle{
		id:        id,
		directory: fileDir,
		filename:  fileName,
		path:      filepath.Join(fileDir, fileName),
	}
}

func NewHandle(fileDir string, fileName string) *Handle {
	return NewHandleWithId(-1, fileDir, fileName)
}

func GetEmptyHandle() *Handle {
	return &EmptyHandle
}

func (s *Handle) GetId() HandleId {
	if s != nil {
		return s.id
	} else {
		return -1
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
