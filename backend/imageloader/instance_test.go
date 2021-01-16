package imageloader

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/database"
)

func TestInstance_GetFull(t *testing.T) {
	a := assert.New(t)

	db := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(db, &database.FileSystemImageFileConverter{})

	horizontal, _ := imageStore.AddImage(apitype.NewImageFile(testAssetsDir, "horizontal.jpg"))
	vertical, _ := imageStore.AddImage(apitype.NewImageFile(testAssetsDir, "vertical.jpg"))

	loader := NewImageLoader(imageStore)

	t.Run("Horizontal", func(t *testing.T) {
		instance := NewInstance(horizontal.GetImageId(), loader)

		scaled, err := instance.GetFull()

		a.Nil(err)
		a.NotNil(scaled)
		a.Equal(3648, scaled.Bounds().Dx())
		a.Equal(2736, scaled.Bounds().Dy())

		a.Equal(39953712, instance.GetByteLength())
	})
	t.Run("Vertical", func(t *testing.T) {
		instance := NewInstance(vertical.GetImageId(), loader)

		scaled, err := instance.GetFull()

		a.Nil(err)
		a.NotNil(scaled)
		a.Equal(2736, scaled.Bounds().Dx())
		a.Equal(3648, scaled.Bounds().Dy())

		a.Equal(39953712, instance.GetByteLength())
	})
	t.Run("Cached", func(t *testing.T) {
		instance := NewInstance(horizontal.GetImageId(), loader)

		scaled, err := instance.GetFull()
		scaled, err = instance.GetFull()
		scaled, err = instance.GetFull()

		a.Nil(err)
		a.NotNil(scaled)
		a.Equal(3648, scaled.Bounds().Dx())
		a.Equal(2736, scaled.Bounds().Dy())

		a.Equal(39953712, instance.GetByteLength())
	})
	t.Run("No image", func(t *testing.T) {
		instance := NewInstance(3, loader)

		scaled, err := instance.GetFull()
		a.NotNil(err)
		a.Nil(scaled)
	})
	t.Run("Invalid", func(t *testing.T) {
		instance := NewInstance(apitype.NoImage, loader)

		scaled, err := instance.GetFull()
		a.NotNil(err)
		a.Nil(scaled)
	})
}

func TestInstance_GetScaled(t *testing.T) {
	a := assert.New(t)

	db := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(db, &database.FileSystemImageFileConverter{})

	horizontal, _ := imageStore.AddImage(apitype.NewImageFile(testAssetsDir, "horizontal.jpg"))
	vertical, _ := imageStore.AddImage(apitype.NewImageFile(testAssetsDir, "vertical.jpg"))

	loader := NewImageLoader(imageStore)

	t.Run("Horizontal", func(t *testing.T) {
		instance := NewInstance(horizontal.GetImageId(), loader)

		size := apitype.SizeOf(400, 400)
		scaled, err := instance.GetScaled(size)

		a.Nil(err)
		a.NotNil(scaled)
		a.Equal(400, scaled.Bounds().Dx())
		a.Equal(300, scaled.Bounds().Dy())

		a.Equal(40433712, instance.GetByteLength())
	})
	t.Run("Vertical", func(t *testing.T) {
		instance := NewInstance(vertical.GetImageId(), loader)

		size := apitype.SizeOf(400, 400)
		scaled, err := instance.GetScaled(size)

		a.Nil(err)
		a.NotNil(scaled)
		a.Equal(300, scaled.Bounds().Dx())
		a.Equal(400, scaled.Bounds().Dy())

		a.Equal(40433712, instance.GetByteLength())
	})
	t.Run("Cached", func(t *testing.T) {
		instance := NewInstance(horizontal.GetImageId(), loader)

		size := apitype.SizeOf(400, 400)
		scaled, err := instance.GetScaled(size)
		scaled, err = instance.GetScaled(size)

		a.Nil(err)
		a.NotNil(scaled)
		a.Equal(400, scaled.Bounds().Dx())
		a.Equal(300, scaled.Bounds().Dy())

		a.Equal(40433712, instance.GetByteLength())
	})
	t.Run("Rescaled", func(t *testing.T) {
		instance := NewInstance(horizontal.GetImageId(), loader)

		size1 := apitype.SizeOf(400, 400)
		scaled, err := instance.GetScaled(size1)

		a.Nil(err)
		a.NotNil(scaled)
		a.Equal(400, scaled.Bounds().Dx())
		a.Equal(300, scaled.Bounds().Dy())

		a.Equal(40433712, instance.GetByteLength())

		size2 := apitype.SizeOf(800, 800)
		scaled, err = instance.GetScaled(size2)

		a.Nil(err)
		a.NotNil(scaled)
		a.Equal(800, scaled.Bounds().Dx())
		a.Equal(600, scaled.Bounds().Dy())

		a.Equal(41873712, instance.GetByteLength())
	})
	t.Run("Not found", func(t *testing.T) {
		instance := NewInstance(3, loader)

		size := apitype.SizeOf(400, 400)
		scaled, err := instance.GetScaled(size)

		a.NotNil(err)
		a.Nil(scaled)
	})
	t.Run("Invalid", func(t *testing.T) {
		instance := NewInstance(apitype.NoImage, loader)

		size := apitype.SizeOf(400, 400)
		scaled, err := instance.GetScaled(size)

		a.NotNil(err)
		a.Nil(scaled)
	})
}

func TestInstance_GetThumbnail(t *testing.T) {
	a := assert.New(t)

	db := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(db, &database.FileSystemImageFileConverter{})

	horizontal, _ := imageStore.AddImage(apitype.NewImageFile(testAssetsDir, "horizontal.jpg"))
	vertical, _ := imageStore.AddImage(apitype.NewImageFile(testAssetsDir, "vertical.jpg"))

	loader := NewImageLoader(imageStore)

	t.Run("Horizontal", func(t *testing.T) {
		instance := NewInstance(horizontal.GetImageId(), loader)

		scaled, err := instance.GetThumbnail()

		a.Nil(err)
		a.NotNil(scaled)
		a.Equal(100, scaled.Bounds().Dx())
		a.Equal(75, scaled.Bounds().Dy())

		a.Equal(30000, instance.GetByteLength())
	})
	t.Run("Vertical", func(t *testing.T) {
		instance := NewInstance(vertical.GetImageId(), loader)

		scaled, err := instance.GetThumbnail()

		a.Nil(err)
		a.NotNil(scaled)
		a.Equal(75, scaled.Bounds().Dx())
		a.Equal(100, scaled.Bounds().Dy())

		a.Equal(30000, instance.GetByteLength())
	})
	t.Run("Cached", func(t *testing.T) {
		instance := NewInstance(horizontal.GetImageId(), loader)

		scaled, err := instance.GetThumbnail()
		scaled, err = instance.GetThumbnail()

		a.Nil(err)
		a.NotNil(scaled)
		a.Equal(100, scaled.Bounds().Dx())
		a.Equal(75, scaled.Bounds().Dy())

		a.Equal(30000, instance.GetByteLength())
	})
	t.Run("Not found", func(t *testing.T) {
		instance := NewInstance(3, loader)

		scaled, err := instance.GetThumbnail()

		a.NotNil(err)
		a.Nil(scaled)
	})
	t.Run("Invalid", func(t *testing.T) {
		instance := NewInstance(apitype.NoImage, loader)

		scaled, err := instance.GetThumbnail()

		a.NotNil(err)
		a.Nil(scaled)
	})
}

func TestInstance_Purge(t *testing.T) {
	a := assert.New(t)

	db := database.NewInMemoryDatabase()
	imageStore := database.NewImageStore(db, &database.FileSystemImageFileConverter{})

	horizontal, _ := imageStore.AddImage(apitype.NewImageFile(testAssetsDir, "horizontal.jpg"))
	_, _ = imageStore.AddImage(apitype.NewImageFile(testAssetsDir, "vertical.jpg"))

	loader := NewImageLoader(imageStore)

	instance := NewInstance(horizontal.GetImageId(), loader)

	scaled, err := instance.GetFull()

	a.Nil(err)
	a.NotNil(scaled)
	a.Equal(3648, scaled.Bounds().Dx())
	a.Equal(2736, scaled.Bounds().Dy())

	a.Equal(39953712, instance.GetByteLength())

	instance.Purge()

	a.Equal(30000, instance.GetByteLength())
}
