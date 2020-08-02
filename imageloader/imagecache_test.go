package imageloader

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"vincit.fi/image-sorter/common"
)

func TestDefaultImageStore_Initialize(t *testing.T) {
	a := assert.New(t)

	loader := NewImageLoader()
	cache := NewImageCache(loader)

	a.Equal(uint64(0), cache.GetByteSize())

	cache.Initialize([]*common.Handle{
		common.NewHandle("../testassets", "horizontal.jpg"),
		common.NewHandle("../testassets", "vertical.jpg"),
	})

	a.Equal(uint64(60000), cache.GetByteSize())
	a.InDelta(0.06, cache.GetSizeInMB(), 0.1)
}

func TestDefaultImageStore_Purge(t *testing.T) {
	a := assert.New(t)

	loader := NewImageLoader()
	cache := NewImageCache(loader)

	a.Equal(uint64(0), cache.GetByteSize())

	handle1 := common.NewHandle("../testassets", "horizontal.jpg")
	handle2 := common.NewHandle("../testassets", "vertical.jpg")
	cache.Initialize([]*common.Handle{handle1, handle2})

	a.Equal(uint64(60000), cache.GetByteSize())

	_, _ = cache.GetFull(handle1)
	_, _ = cache.GetFull(handle2)
	size := common.SizeOf(100, 100)
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

	handle := common.NewHandle("../testassets", "horizontal.jpg")
	img, err := cache.GetFull(handle)

	a.Nil(err)
	a.NotNil(img)
}

func TestDefaultImageStore_GetScaled(t *testing.T) {
	a := assert.New(t)

	loader := NewImageLoader()
	cache := NewImageCache(loader)

	handle := common.NewHandle("../testassets", "horizontal.jpg")
	size := common.SizeOf(400, 400)
	img, err := cache.GetScaled(handle, size)

	a.Nil(err)
	a.NotNil(img)
}

func TestDefaultImageStore_GetThumbnail(t *testing.T) {
	a := assert.New(t)

	loader := NewImageLoader()
	cache := NewImageCache(loader)

	handle := common.NewHandle("../testassets", "horizontal.jpg")
	img, err := cache.GetThumbnail(handle)

	a.Nil(err)
	a.NotNil(img)
}

func TestDefaultImageStore_GetExifData(t *testing.T) {
	a := assert.New(t)

	loader := NewImageLoader()
	cache := NewImageCache(loader)

	handle := common.NewHandle("../testassets", "vertical.jpg")
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
}
