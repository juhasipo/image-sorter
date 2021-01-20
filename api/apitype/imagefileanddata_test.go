package apitype

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestImageContainer_String(t *testing.T) {
	a := assert.New(t)

	t.Run("Valid", func(t *testing.T) {
		imageFile := NewImageFileWithId(1, "foo", "bar")
		imageMetaData := NewImageMetaData(1024, 90, true, map[string]string{})

		container := NewImageContainer(&ImageFileWithMetaData{
			ImageFile:     *imageFile,
			ImageMetaData: *imageMetaData,
		}, nil)
		a.Equal("ImageFileAndData{ImageFile{bar}}", container.String())
	})
	t.Run("Nil ImageFile", func(t *testing.T) {
		container := NewImageContainer(nil, nil)
		a.Equal("ImageFileAndData{ImageFile<nil>}", container.String())
	})
	t.Run("Nil", func(t *testing.T) {
		var container *ImageFileAndData
		a.Equal("ImageFileAndData<nil>", container.String())
	})

}
