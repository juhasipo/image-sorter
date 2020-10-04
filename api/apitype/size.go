package apitype

import (
	"github.com/gotk3/gotk3/gtk"
	"image"
)

type Size struct {
	width  int
	height int
}

func (s *Size) GetHeight() int {
	return s.height
}

func (s *Size) GetWidth() int {
	return s.width
}

func SizeOf(width int, height int) Size {
	return Size{width, height}
}

func applyZoom(value int, zoom float64) int {
	return int(float64(value) * zoom)
}

func SizeFromWindow(widget *gtk.ScrolledWindow, zoom float64) Size {
	return Size{
		width:  applyZoom(widget.GetAllocatedWidth(), zoom),
		height: applyZoom(widget.GetAllocatedHeight(), zoom),
	}
}

func SizeFromRectangle(rectangle image.Rectangle, zoom float64) Size {
	return Size{
		width:  applyZoom(rectangle.Dx(), zoom),
		height: applyZoom(rectangle.Dy(), zoom),
	}
}

// TODO: Implement via Size
func ScaleToFit(sourceWidth int, sourceHeight int, targetWidth int, targetHeight int) (int, int) {
	ratio := float32(sourceWidth) / float32(sourceHeight)
	newWidth := int(float32(targetHeight) * ratio)
	newHeight := targetHeight

	if newWidth > targetWidth {
		newWidth = targetWidth
		newHeight = int(float32(targetWidth) / ratio)
	}
	return newWidth, newHeight
}
