package common

import (
	"os"
	"path/filepath"
	"strings"
)

type Handle struct {
	id string
	path string
}

func (s *Handle) IsValid() bool {
	return s.id != ""
}

var (
	EMPTY_HANDLE = Handle {id: "", path: ""}
)

func GetEmptyHandle() *Handle {
	return &EMPTY_HANDLE
}

func (s* Handle) GetPath() string {
	return s.path
}

func LoadImages(dir string) []*Handle {
	var handles []*Handle
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(strings.ToLower(path)) != ".jpg" {
			return nil
		}
		handles = append(handles, &Handle {id: path, path: path})
		return nil
	})
	return handles
}
