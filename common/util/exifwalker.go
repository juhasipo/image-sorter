package util

import (
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
	"strings"
)

type MapExifWalker struct {
	values map[string]string

	exif.Walker
}

func NewMapExifWalker() *MapExifWalker {
	return &MapExifWalker{
		values: map[string]string{},
	}
}

func (s *MapExifWalker) Walk(name exif.FieldName, tag *tiff.Tag) error {
	if tagValue := strings.Trim(tag.String(), " \t\""); tagValue != "" {
		key := string(name)
		s.values[key] = tagValue
	}
	return nil
}

func (s *MapExifWalker) GetMetaData() map[string]string {
	return s.values
}
