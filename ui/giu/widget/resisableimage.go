package widget

import "github.com/AllenDang/giu"

type ResizableImageWidget struct {
	texturedImage *TexturedImage
	giu.ImageWidget
}

func ResizableImage(image *TexturedImage) *ResizableImageWidget {
	return &ResizableImageWidget{
		texturedImage: image,
		ImageWidget:   *giu.Image(image.Texture),
	}
}

func (s *ResizableImageWidget) Build() {
	paddingW, _ := giu.GetFramePadding()
	maxW, maxH := giu.GetAvailableRegion()
	maxW = maxW - 120 - paddingW*2.0
	newW := maxW
	newH := newW / s.texturedImage.Ratio

	if newH > maxH {
		newW = maxH * s.texturedImage.Ratio
		newH = maxH
	}

	offsetW := (maxW - newW) / 2.0
	offsetH := (maxH - newH) / 2.0

	s.ImageWidget.Size(newW, newH)
	giu.Row(giu.Dummy(offsetW, offsetH), &s.ImageWidget).Build()
}
