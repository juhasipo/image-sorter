package common

import (
	"github.com/gotk3/gotk3/gtk"
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

func SizeFromWindow(widget *gtk.ScrolledWindow) Size {
	return Size{
		width:  widget.GetAllocatedWidth(),
		height: widget.GetAllocatedHeight(),
	}
}

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
