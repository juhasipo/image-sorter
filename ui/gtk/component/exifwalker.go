package component

import (
	"bytes"
	"fmt"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
	"sort"
	"strings"
)

type ExifWalker struct {
	stringBuffer *bytes.Buffer
	values       []string

	exif.Walker
}

func (s *ExifWalker) Walk(name exif.FieldName, tag *tiff.Tag) error {
	tagValue := strings.Trim(tag.String(), " \t\"")

	if tagValue != "" {
		s.values = append(s.values, fmt.Sprintf("%s: %s", string(name), tagValue))
	}
	return nil
}

func (s *ExifWalker) String() string {
	sort.Strings(s.values)
	b := bytes.NewBuffer([]byte{})
	for _, value := range s.values {
		b.WriteString(value)
		b.WriteString("\n")
	}
	return b.String()
}
