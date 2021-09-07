package widget

import (
	"github.com/AllenDang/giu"
)

type ResizableImageWidget struct {
	texturedImage       *TexturedImage
	maxHeight, maxWidth float32
	zoomFactor          float32
	giu.ImageWidget
}

func ResizableImage(image *TexturedImage) *ResizableImageWidget {
	return &ResizableImageWidget{
		texturedImage: image,
		maxHeight:     0,
		maxWidth:      0,
		zoomFactor:    1,
		ImageWidget:   *giu.Image(image.Texture),
	}
}

func (s *ResizableImageWidget) Size(width float32, height float32) *ResizableImageWidget {
	s.ImageWidget.Size(width, height)
	s.maxWidth = width
	s.maxHeight = height
	return s
}

func (s *ResizableImageWidget) ZoomFactor(zoomFactor float32) *ResizableImageWidget {
	s.zoomFactor = zoomFactor
	return s
}

func (s *ResizableImageWidget) Build() {
	maxW, maxH := giu.GetAvailableRegion()

	if s.maxWidth != 0 && s.maxHeight != 0 {
		maxW = s.maxWidth
		maxH = s.maxHeight
	}

	newW := maxW * s.zoomFactor
	newH := newW / s.texturedImage.Ratio

	if newH > maxH {
		newW = maxH * s.texturedImage.Ratio * s.zoomFactor
		newH = maxH * s.zoomFactor
	}

	offsetW := (maxW - newW) / 2.0
	offsetH := (maxH - newH) / 2.0

	if offsetW < 0 {
		offsetW = 0
	}
	if offsetH < 0 {
		offsetH = 0
	}

	dummyV := giu.Dummy(0, offsetH)
	dummyH := giu.Dummy(offsetW, 20)
	s.ImageWidget.Size(newW, newH)

	giu.Column(
		dummyV,
		giu.Row(dummyH, &s.ImageWidget, dummyH),
		dummyV,
	).Build()
}
