package common

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadExifData(t *testing.T) {
	a := assert.New(t)

	t.Run("Horizontal image", func(t *testing.T) {
		data, err := LoadExifData(NewHandle("../testassets", "horizontal.jpg"))
		a.Nil(err)
		a.Equal(uint8(1), data.orientation)
		a.Equal(gdk.PixbufRotation(0), data.rotation)
		a.Equal(false, data.flipped)
	})

	t.Run("Vertical image", func(t *testing.T) {
		data, err := LoadExifData(NewHandle("../testassets", "vertical.jpg"))
		a.Nil(err)
		a.Equal(uint8(6), data.orientation)
		a.Equal(gdk.PixbufRotation(270), data.rotation)
		a.Equal(false, data.flipped)
	})
}