package imageloader

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"vincit.fi/image-sorter/api/apitype"
)

func TestInstance_GetFull(t *testing.T) {
	a := assert.New(t)

	loader := NewImageLoader()

	t.Run("Horizontal", func(t *testing.T) {
		handle := apitype.NewHandleWithId(1, testAssetsDir, "horizontal.jpg", map[string]string{})
		instance := NewInstance(handle, loader)

		scaled, err := instance.GetFull()

		a.Nil(err)
		a.NotNil(scaled)
		a.Equal(3648, scaled.Bounds().Dx())
		a.Equal(2736, scaled.Bounds().Dy())

		a.Equal(39953712, instance.GetByteLength())
	})
	t.Run("Vertical", func(t *testing.T) {
		handle := apitype.NewHandleWithId(2, testAssetsDir, "vertical.jpg", map[string]string{})
		instance := NewInstance(handle, loader)

		scaled, err := instance.GetFull()

		a.Nil(err)
		a.NotNil(scaled)
		a.Equal(2736, scaled.Bounds().Dx())
		a.Equal(3648, scaled.Bounds().Dy())

		a.Equal(39953712, instance.GetByteLength())
	})
	t.Run("Cached", func(t *testing.T) {
		handle := apitype.NewHandleWithId(1, testAssetsDir, "horizontal.jpg", map[string]string{})
		instance := NewInstance(handle, loader)

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
		handle := apitype.NewHandleWithId(3, testAssetsDir, "no_image.jpg", map[string]string{})
		instance := NewInstance(handle, loader)

		scaled, err := instance.GetFull()
		a.NotNil(err)
		a.Nil(scaled)
		a.True(handle.IsValid())
	})
	t.Run("Invalid", func(t *testing.T) {
		handle := apitype.NewHandle("", "")
		instance := NewInstance(handle, loader)

		scaled, err := instance.GetFull()
		a.NotNil(err)
		a.Nil(scaled)
		a.False(handle.IsValid())
	})
	t.Run("Nil", func(t *testing.T) {
		instance := NewInstance(nil, loader)

		scaled, err := instance.GetFull()

		a.NotNil(err)
		a.Nil(scaled)
	})
}

func TestInstance_GetScaled(t *testing.T) {
	a := assert.New(t)

	loader := NewImageLoader()

	t.Run("Horizontal", func(t *testing.T) {
		handle := apitype.NewHandleWithId(1, testAssetsDir, "horizontal.jpg", map[string]string{})
		instance := NewInstance(handle, loader)

		size := apitype.SizeOf(400, 400)
		scaled, err := instance.GetScaled(size)

		a.Nil(err)
		a.NotNil(scaled)
		a.Equal(400, scaled.Bounds().Dx())
		a.Equal(300, scaled.Bounds().Dy())

		a.Equal(40433712, instance.GetByteLength())
	})
	t.Run("Vertical", func(t *testing.T) {
		handle := apitype.NewHandleWithId(2, testAssetsDir, "vertical.jpg", map[string]string{})
		instance := NewInstance(handle, loader)

		size := apitype.SizeOf(400, 400)
		scaled, err := instance.GetScaled(size)

		a.Nil(err)
		a.NotNil(scaled)
		a.Equal(300, scaled.Bounds().Dx())
		a.Equal(400, scaled.Bounds().Dy())

		a.Equal(40433712, instance.GetByteLength())
	})
	t.Run("Cached", func(t *testing.T) {
		handle := apitype.NewHandleWithId(1, testAssetsDir, "horizontal.jpg", map[string]string{})
		instance := NewInstance(handle, loader)

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
		handle := apitype.NewHandleWithId(1, testAssetsDir, "horizontal.jpg", map[string]string{})
		instance := NewInstance(handle, loader)

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
		handle := apitype.NewHandleWithId(3, testAssetsDir, "no_image.jpg", map[string]string{})
		instance := NewInstance(handle, loader)

		size := apitype.SizeOf(400, 400)
		scaled, err := instance.GetScaled(size)

		a.NotNil(err)
		a.Nil(scaled)
		a.True(handle.IsValid())
	})
	t.Run("Invalid", func(t *testing.T) {
		handle := apitype.NewHandle("", "")
		instance := NewInstance(handle, loader)

		size := apitype.SizeOf(400, 400)
		scaled, err := instance.GetScaled(size)

		a.NotNil(err)
		a.Nil(scaled)
		a.False(handle.IsValid())
	})
	t.Run("Nil", func(t *testing.T) {
		instance := NewInstance(nil, loader)

		size := apitype.SizeOf(400, 400)
		scaled, err := instance.GetScaled(size)

		a.NotNil(err)
		a.Nil(scaled)
	})
}

func TestInstance_GetThumbnail(t *testing.T) {
	a := assert.New(t)

	loader := NewImageLoader()

	t.Run("Horizontal", func(t *testing.T) {
		handle := apitype.NewHandleWithId(1, testAssetsDir, "horizontal.jpg", map[string]string{})
		instance := NewInstance(handle, loader)

		scaled, err := instance.GetThumbnail()

		a.Nil(err)
		a.NotNil(scaled)
		a.Equal(100, scaled.Bounds().Dx())
		a.Equal(75, scaled.Bounds().Dy())

		a.Equal(30000, instance.GetByteLength())
	})
	t.Run("Vertical", func(t *testing.T) {
		handle := apitype.NewHandleWithId(2, testAssetsDir, "vertical.jpg", map[string]string{})
		instance := NewInstance(handle, loader)

		scaled, err := instance.GetThumbnail()

		a.Nil(err)
		a.NotNil(scaled)
		a.Equal(75, scaled.Bounds().Dx())
		a.Equal(100, scaled.Bounds().Dy())

		a.Equal(30000, instance.GetByteLength())
	})
	t.Run("Cached", func(t *testing.T) {
		handle := apitype.NewHandleWithId(1, testAssetsDir, "horizontal.jpg", map[string]string{})
		instance := NewInstance(handle, loader)

		scaled, err := instance.GetThumbnail()
		scaled, err = instance.GetThumbnail()

		a.Nil(err)
		a.NotNil(scaled)
		a.Equal(100, scaled.Bounds().Dx())
		a.Equal(75, scaled.Bounds().Dy())

		a.Equal(30000, instance.GetByteLength())
	})
	t.Run("Not found", func(t *testing.T) {
		handle := apitype.NewHandleWithId(3, testAssetsDir, "no_image.jpg", map[string]string{})
		instance := NewInstance(handle, loader)

		scaled, err := instance.GetThumbnail()

		a.NotNil(err)
		a.Nil(scaled)
		a.True(handle.IsValid())
	})
	t.Run("Invalid", func(t *testing.T) {
		handle := apitype.NewHandle("", "")
		instance := NewInstance(handle, loader)

		scaled, err := instance.GetThumbnail()

		a.NotNil(err)
		a.Nil(scaled)
		a.False(handle.IsValid())
	})
	t.Run("Nil", func(t *testing.T) {
		instance := NewInstance(nil, loader)

		scaled, err := instance.GetThumbnail()

		a.NotNil(err)
		a.Nil(scaled)
	})
}

func TestInstance_Purge(t *testing.T) {
	a := assert.New(t)

	loader := NewImageLoader()

	handle := apitype.NewHandleWithId(1, testAssetsDir, "horizontal.jpg", map[string]string{})
	instance := NewInstance(handle, loader)

	scaled, err := instance.GetFull()

	a.Nil(err)
	a.NotNil(scaled)
	a.Equal(3648, scaled.Bounds().Dx())
	a.Equal(2736, scaled.Bounds().Dy())

	a.Equal(39953712, instance.GetByteLength())

	instance.Purge()

	a.Equal(30000, instance.GetByteLength())
}
