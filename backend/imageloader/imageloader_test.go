package imageloader

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"vincit.fi/image-sorter/api/apitype"
)

const testAssetsDir = "../../testassets"

func TestLibJPEGImageLoader_LoadImage(t *testing.T) {
	a := assert.New(t)

	loader := LibJPEGImageLoader{}
	t.Run("Horizontal", func(t *testing.T) {
		img, err := loader.LoadImage(apitype.NewHandle(testAssetsDir, "horizontal.jpg"))

		a.Nil(err)
		a.NotNil(img)

		a.Equal(3648, img.Bounds().Dx())
		a.Equal(2736, img.Bounds().Dy())
	})
	t.Run("Vertical", func(t *testing.T) {
		img, err := loader.LoadImage(apitype.NewHandle(testAssetsDir, "vertical.jpg"))

		a.Nil(err)
		a.NotNil(img)

		a.Equal(3648, img.Bounds().Dx())
		a.Equal(2736, img.Bounds().Dy())
	})

}

func TestLibJPEGImageLoader_LoadImageScaled(t *testing.T) {
	a := assert.New(t)

	size := apitype.SizeOf(1, 1)

	loader := LibJPEGImageLoader{}
	t.Run("Horizontal is loaded with the smallest image lib JPEG could load image", func(t *testing.T) {
		img, err := loader.LoadImageScaled(apitype.NewHandle(testAssetsDir, "horizontal.jpg"), size)

		a.Nil(err)
		a.NotNil(img)

		a.Equal(456, img.Bounds().Dx())
		a.Equal(342, img.Bounds().Dy())
	})
	t.Run("Vertical is loaded with the smallest image lib JPEG could load image", func(t *testing.T) {
		img, err := loader.LoadImageScaled(apitype.NewHandle(testAssetsDir, "vertical.jpg"), size)

		a.Nil(err)
		a.NotNil(img)

		a.Equal(456, img.Bounds().Dx())
		a.Equal(342, img.Bounds().Dy())
	})

}
