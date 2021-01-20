package util

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/stretchr/testify/assert"
	"testing"
	"vincit.fi/image-sorter/api/apitype"
)

const testAssetsDir = "../../testassets"

func TestLoadExifData(t *testing.T) {
	a := assert.New(t)

	t.Run("Horizontal image", func(t *testing.T) {
		data, err := LoadExifData(apitype.NewImageFileWithId(1, testAssetsDir, "horizontal.jpg"))
		a.Nil(err)
		a.Equal(uint8(1), data.ExifOrientation())
		a.Equal(gdk.PixbufRotation(0), data.Rotation())
		a.Equal(false, data.Flipped())
	})

	t.Run("Vertical image", func(t *testing.T) {
		data, err := LoadExifData(apitype.NewImageFileWithId(2, testAssetsDir, "vertical.jpg"))
		a.Nil(err)
		a.Equal(uint8(6), data.ExifOrientation())
		a.Equal(gdk.PixbufRotation(270), data.Rotation())
		a.Equal(false, data.Flipped())
	})
}
