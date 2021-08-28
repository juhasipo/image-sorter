package apitype

import (
	"image"
)

type Size struct {
	width  int
	height int
}

func (s *Size) Height() int {
	return s.height
}

func (s *Size) Width() int {
	return s.width
}

func SizeOf(width int, height int) Size {
	return Size{width, height}
}

func applyZoom(value int, zoom float64) int {
	return int(float64(value) * zoom)
}

func ZoomedSizeFromRectangle(rectangle image.Rectangle, zoom float64) Size {
	return Size{
		width:  applyZoom(rectangle.Dx(), zoom),
		height: applyZoom(rectangle.Dy(), zoom),
	}
}

func PointOfScaledToFit(source image.Point, target Size) Size {
	return SizeOf(intSizeOfScaledToFit(source.X, source.Y, target.width, target.height))
}

func RectangleOfScaledToFit(source image.Rectangle, target Size) Size {
	return SizeOf(intSizeOfScaledToFit(source.Dx(), source.Dy(), target.width, target.height))
}

func intSizeOfScaledToFit(sourceWidth int, sourceHeight int, targetWidth int, targetHeight int) (int, int) {
	ratio := float32(sourceWidth) / float32(sourceHeight)
	newWidth := int(float32(targetHeight) * ratio)
	newHeight := targetHeight

	if newWidth > targetWidth {
		newWidth = targetWidth
		newHeight = int(float32(targetWidth) / ratio)
	}
	return newWidth, newHeight
}
