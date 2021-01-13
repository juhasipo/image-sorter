package imageloader

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"vincit.fi/image-sorter/api/apitype"
)

func TestDefaultImageStore_Initialize(t *testing.T) {
	a := assert.New(t)

	loader := NewImageLoader()
	cache := NewImageCache(loader)

	a.Equal(uint64(0), cache.GetByteSize())

	cache.Initialize([]*apitype.Handle{
		apitype.NewHandleWithId(1, testAssetsDir, "horizontal.jpg", map[string]string{}),
		apitype.NewHandleWithId(2, testAssetsDir, "vertical.jpg", map[string]string{}),
	})

	a.Equal(uint64(60000), cache.GetByteSize())
	a.InDelta(0.06, cache.GetSizeInMB(), 0.1)
}

func TestDefaultImageStore_Purge(t *testing.T) {
	a := assert.New(t)

	loader := NewImageLoader()
	cache := NewImageCache(loader)

	a.Equal(uint64(0), cache.GetByteSize())

	handle1 := apitype.NewHandleWithId(1, testAssetsDir, "horizontal.jpg", map[string]string{})
	handle2 := apitype.NewHandleWithId(2, testAssetsDir, "vertical.jpg", map[string]string{})
	cache.Initialize([]*apitype.Handle{handle1, handle2})

	a.Equal(uint64(60000), cache.GetByteSize())

	_, _ = cache.GetFull(handle1)
	_, _ = cache.GetFull(handle2)
	size := apitype.SizeOf(100, 100)
	_, _ = cache.GetScaled(handle1, size)
	_, _ = cache.GetScaled(handle2, size)

	a.Equal(uint64(79967424), cache.GetByteSize())
	a.InDelta(76.3, cache.GetSizeInMB(), 0.1)

	cache.Purge()
	a.Equal(uint64(60000), cache.GetByteSize())
}

func TestDefaultImageStore_GetFull(t *testing.T) {
	a := assert.New(t)

	loader := NewImageLoader()
	cache := NewImageCache(loader)

	t.Run("Valid", func(t *testing.T) {
		handle := apitype.NewHandleWithId(1, testAssetsDir, "horizontal.jpg", map[string]string{})
		img, err := cache.GetFull(handle)

		a.Nil(err)
		a.NotNil(img)
	})
	t.Run("No exif", func(t *testing.T) {
		handle := apitype.NewHandleWithId(2, testAssetsDir, "no-exif.jpg", map[string]string{})
		img, err := cache.GetFull(handle)

		a.Nil(err)
		a.NotNil(img)
	})
	t.Run("Invalid", func(t *testing.T) {
		handle := apitype.NewHandle("", "")
		img, err := cache.GetFull(handle)

		a.NotNil(err)
		a.Nil(img)
	})
}

func TestDefaultImageStore_GetScaled(t *testing.T) {
	a := assert.New(t)

	loader := NewImageLoader()
	cache := NewImageCache(loader)

	t.Run("Valid", func(t *testing.T) {
		handle := apitype.NewHandleWithId(1, testAssetsDir, "horizontal.jpg", map[string]string{})
		size := apitype.SizeOf(400, 400)
		img, err := cache.GetScaled(handle, size)

		a.Nil(err)
		a.NotNil(img)
	})
	t.Run("No exif", func(t *testing.T) {
		handle := apitype.NewHandleWithId(2, testAssetsDir, "no-exif.jpg", map[string]string{})
		size := apitype.SizeOf(400, 400)
		img, err := cache.GetScaled(handle, size)

		a.Nil(err)
		a.NotNil(img)
	})
	t.Run("Invalid", func(t *testing.T) {
		handle := apitype.NewHandle("", "")
		size := apitype.SizeOf(400, 400)
		img, err := cache.GetScaled(handle, size)

		a.NotNil(err)
		a.Nil(img)
	})
}

func TestDefaultImageStore_GetThumbnail(t *testing.T) {
	a := assert.New(t)

	loader := NewImageLoader()
	cache := NewImageCache(loader)

	t.Run("Valid", func(t *testing.T) {
		handle := apitype.NewHandleWithId(1, testAssetsDir, "horizontal.jpg", map[string]string{})
		img, err := cache.GetThumbnail(handle)

		a.Nil(err)
		a.NotNil(img)
	})
	t.Run("No exif", func(t *testing.T) {
		handle := apitype.NewHandleWithId(2, testAssetsDir, "no-exif.jpg", map[string]string{})
		img, err := cache.GetThumbnail(handle)

		a.Nil(err)
		a.NotNil(img)
	})
	t.Run("Invalid", func(t *testing.T) {
		handle := apitype.NewHandle("", "")
		img, err := cache.GetThumbnail(handle)

		a.NotNil(err)
		a.Nil(img)
	})
}

func TestDefaultImageStore_GetExifData(t *testing.T) {
	a := assert.New(t)

	loader := NewImageLoader()
	cache := NewImageCache(loader)

	t.Run("Valid", func(t *testing.T) {
		handle := apitype.NewHandleWithId(1, testAssetsDir, "vertical.jpg", map[string]string{})
		exifData := cache.GetExifData(handle)

		a.Equal(gdk.PixbufRotation(270), exifData.GetRotation())
		a.NotNil(exifData)
		if orientationTag := exifData.Get(exif.Orientation); a.NotNil(orientationTag) {
			orientation, _ := orientationTag.Int(0)
			a.Equal(6, orientation)
		}

		if modelTag := exifData.Get(exif.Model); a.NotNil(modelTag) {
			model, _ := modelTag.StringVal()
			a.Equal("XZ-1", strings.TrimSpace(model))
		}
	})
	t.Run("No exif", func(t *testing.T) {
		handle := apitype.NewHandleWithId(2, testAssetsDir, "no-exif.jpg", map[string]string{})
		exifData := cache.GetExifData(handle)

		if a.NotNil(exifData) {
			a.Equal(gdk.PixbufRotation(0), exifData.GetRotation())
		}
	})
	t.Run("Invalid", func(t *testing.T) {
		handle := apitype.NewHandle("", "")
		exifData := cache.GetExifData(handle)

		a.Nil(exifData)
	})
}
