package apitype

import (
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestGetEmptyHandle(t *testing.T) {
	a := assert.New(t)

	handle := GetEmptyHandle()

	a.False(handle.IsValid())
}

func TestHandle_String(t *testing.T) {
	a := assert.New(t)

	var nilHandle *ImageFile
	a.Equal("ImageFile<nil>", nilHandle.String())
	a.Equal("ImageFile<invalid>", NewHandle("", "").String())
	a.Equal("ImageFile{file.jpeg}", NewHandleWithId(2, "/some/dir", "file.jpeg").String())
}

func TestValidHandle(t *testing.T) {
	a := assert.New(t)

	handle := NewHandleWithId(1, "some/dir", "file.jpeg")

	t.Run("Validity", func(t *testing.T) {
		a.True(handle.IsValid())
	})
	t.Run("Properties", func(t *testing.T) {
		a.Equal(ImageId(1), handle.GetId())
		a.Equal("file.jpeg", handle.GetFile())
		a.Equal("some/dir", handle.GetDir())
		a.Equal(filepath.Join("some", "dir", "file.jpeg"), handle.GetPath())
	})
}

func TestInvalidHandle(t *testing.T) {
	a := assert.New(t)

	handle := NewHandleWithId(NoImage, "", "")

	t.Run("Validity", func(t *testing.T) {
		a.False(handle.IsValid())
	})
	t.Run("Properties", func(t *testing.T) {
		a.Equal(NoImage, handle.GetId())
		a.Equal("", handle.GetFile())
		a.Equal("", handle.GetDir())
		a.Equal("", handle.GetPath())
	})
}

func TestNilHandle(t *testing.T) {
	a := assert.New(t)

	var handle *ImageFile

	t.Run("Validity", func(t *testing.T) {
		a.False(handle.IsValid())
	})
	t.Run("Properties", func(t *testing.T) {
		a.Equal(NoImage, handle.GetId())
		a.Equal("", handle.GetFile())
		a.Equal("", handle.GetDir())
		a.Equal("", handle.GetPath())
	})
}

func TestIsSupported(t *testing.T) {
	a := assert.New(t)

	t.Run("Valid", func(t *testing.T) {
		validValues := []string{
			"jpeg", "JPEG", "jpg", "JPG",
		}
		for _, value := range validValues {
			a.True(isSupported("." + value))
		}
	})

	t.Run("Invalid", func(t *testing.T) {
		invalidValues := []string{
			"exe", "EXE",
		}
		for _, value := range invalidValues {
			a.False(isSupported("." + value))
		}
	})
}
