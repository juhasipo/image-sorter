package apitype

import (
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestGetEmptyImageFile(t *testing.T) {
	a := assert.New(t)

	imageFile := GetEmptyImageFile()

	a.False(imageFile.IsValid())
}

func TestImageFile_String(t *testing.T) {
	a := assert.New(t)

	var nilImageFile *ImageFile
	a.Equal("ImageFile<nil>", nilImageFile.String())
	a.Equal("ImageFile<invalid>", NewImageFile("", "").String())
	a.Equal("ImageFile{file.jpeg}", NewImageFileWithId(2, "/some/dir", "file.jpeg").String())
}

func TestValidImageFile(t *testing.T) {
	a := assert.New(t)

	imageFile := NewImageFileWithId(1, "some/dir", "file.jpeg")

	t.Run("Validity", func(t *testing.T) {
		a.True(imageFile.IsValid())
	})
	t.Run("Properties", func(t *testing.T) {
		a.Equal(ImageId(1), imageFile.GetId())
		a.Equal("file.jpeg", imageFile.GetFile())
		a.Equal("some/dir", imageFile.GetDir())
		a.Equal(filepath.Join("some", "dir", "file.jpeg"), imageFile.GetPath())
	})
}

func TestInvalidImageFile(t *testing.T) {
	a := assert.New(t)

	imageFile := NewImageFileWithId(NoImage, "", "")

	t.Run("Validity", func(t *testing.T) {
		a.False(imageFile.IsValid())
	})
	t.Run("Properties", func(t *testing.T) {
		a.Equal(NoImage, imageFile.GetId())
		a.Equal("", imageFile.GetFile())
		a.Equal("", imageFile.GetDir())
		a.Equal("", imageFile.GetPath())
	})
}

func TestNilImageFile(t *testing.T) {
	a := assert.New(t)

	var imageFile *ImageFile

	t.Run("Validity", func(t *testing.T) {
		a.False(imageFile.IsValid())
	})
	t.Run("Properties", func(t *testing.T) {
		a.Equal(NoImage, imageFile.GetId())
		a.Equal("", imageFile.GetFile())
		a.Equal("", imageFile.GetDir())
		a.Equal("", imageFile.GetPath())
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
