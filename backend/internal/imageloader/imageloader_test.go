package imageloader

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/internal/database"
)

const testAssetsDir = "../../../testassets"

func TestLibJPEGImageLoader_LoadImage(t *testing.T) {
	a := assert.New(t)

	db := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(db, &database.FileSystemImageFileConverter{})

	horizontal, _ := imageStore.AddImage(apitype.NewImageFile(testAssetsDir, "horizontal.jpg"))
	vertical, _ := imageStore.AddImage(apitype.NewImageFile(testAssetsDir, "vertical.jpg"))

	loader := NewImageLoader(imageStore)
	t.Run("Horizontal", func(t *testing.T) {
		img, err := loader.LoadImage(horizontal.Id())

		a.Nil(err)
		a.NotNil(img)

		a.Equal(3648, img.Bounds().Dx())
		a.Equal(2736, img.Bounds().Dy())
	})
	t.Run("Vertical", func(t *testing.T) {
		img, err := loader.LoadImage(vertical.Id())

		a.Nil(err)
		a.NotNil(img)

		a.Equal(2736, img.Bounds().Dx())
		a.Equal(3648, img.Bounds().Dy())
	})

}

func TestLibJPEGImageLoader_LoadImageScaled(t *testing.T) {
	a := assert.New(t)

	db := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(db, &database.FileSystemImageFileConverter{})

	horizontal, _ := imageStore.AddImage(apitype.NewImageFile(testAssetsDir, "horizontal.jpg"))
	vertical, _ := imageStore.AddImage(apitype.NewImageFile(testAssetsDir, "vertical.jpg"))

	size := apitype.SizeOf(1, 1)

	loader := NewImageLoader(imageStore)
	t.Run("Horizontal is loaded with the smallest image lib JPEG could load image", func(t *testing.T) {
		img, err := loader.LoadImageScaled(horizontal.Id(), size)

		a.Nil(err)
		a.NotNil(img)

		a.Equal(456, img.Bounds().Dx())
		a.Equal(342, img.Bounds().Dy())
	})
	t.Run("Vertical is loaded with the smallest image lib JPEG could load image", func(t *testing.T) {
		img, err := loader.LoadImageScaled(vertical.Id(), size)

		a.Nil(err)
		a.NotNil(img)

		a.Equal(342, img.Bounds().Dx())
		a.Equal(456, img.Bounds().Dy())
	})

}
