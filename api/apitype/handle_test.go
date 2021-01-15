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

	var nilHandle *Handle
	a.Equal("Handle<nil>", nilHandle.String())
	a.Equal("Handle<invalid>", NewHandle("", "").String())
	a.Equal("Handle{file.jpeg}", NewHandleWithId(2, "/some/dir", "file.jpeg", 0, false, map[string]string{}).String())
}

func TestValidHandle(t *testing.T) {
	a := assert.New(t)

	handle := NewHandleWithId(1, "some/dir", "file.jpeg", 0, false, map[string]string{})
	handle.SetByteSize(1.5 * 1024 * 1024)

	t.Run("Validity", func(t *testing.T) {
		a.True(handle.IsValid())
	})
	t.Run("Properties", func(t *testing.T) {
		a.Equal(HandleId(1), handle.GetId())
		a.Equal("file.jpeg", handle.GetFile())
		a.Equal("some/dir", handle.GetDir())
		a.Equal(filepath.Join("some", "dir", "file.jpeg"), handle.GetPath())
		a.Equal(int64(1.5*1024*1024), handle.GetByteSize())
		a.Equal(1.5, handle.GetByteSizeMB())
	})
}

func TestInvalidHandle(t *testing.T) {
	a := assert.New(t)

	handle := NewHandleWithId(NoHandle, "", "", 0, false, map[string]string{})

	t.Run("Validity", func(t *testing.T) {
		a.False(handle.IsValid())
	})
	t.Run("Properties", func(t *testing.T) {
		a.Equal(NoHandle, handle.GetId())
		a.Equal("", handle.GetFile())
		a.Equal("", handle.GetDir())
		a.Equal("", handle.GetPath())
		a.Equal(int64(0), handle.GetByteSize())
		a.Equal(0.0, handle.GetByteSizeMB())
	})
}

func TestNilHandle(t *testing.T) {
	a := assert.New(t)

	var handle *Handle

	t.Run("Validity", func(t *testing.T) {
		a.False(handle.IsValid())
	})
	t.Run("Properties", func(t *testing.T) {
		a.Equal(NoHandle, handle.GetId())
		a.Equal("", handle.GetFile())
		a.Equal("", handle.GetDir())
		a.Equal("", handle.GetPath())
		a.Equal(int64(0), handle.GetByteSize())
		a.Equal(0.0, handle.GetByteSizeMB())
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
