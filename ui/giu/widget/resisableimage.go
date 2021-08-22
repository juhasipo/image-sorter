package widget

import "github.com/AllenDang/giu"

type ResizableImageWidget struct {
	imageWidth  float32
	imageHeight float32
	imageRatio  float32
	giu.ImageWidget
}

func ResizableImage(texture *giu.Texture, width float32, height float32) *ResizableImageWidget {
	return &ResizableImageWidget{
		imageWidth:  width,
		imageHeight: height,
		imageRatio:  width / height,
		ImageWidget: *giu.Image(texture),
	}
}

func (s *ResizableImageWidget) Build() {
	paddingW, _ := giu.GetFramePadding()
	maxW, maxH := giu.GetAvailableRegion()
	maxW = maxW - 120 - paddingW*2.0
	newW := maxW
	newH := newW / s.imageRatio

	if newH > maxH {
		newW = maxH * s.imageRatio
		newH = maxH
	}

	offsetW := (maxW - newW) / 2.0
	offsetH := (maxH - newH) / 2.0

	s.ImageWidget.Size(newW, newH)
	giu.Row(giu.Dummy(offsetW, offsetH), &s.ImageWidget).Build()
}
