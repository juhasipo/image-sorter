package widget

import (
	"github.com/AllenDang/giu"
)

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
	maxW, maxH := giu.GetAvailableRegion()
	newW := maxW
	newH := newW / s.texturedImage.Ratio

	if newH > maxH {
		newW = maxH * s.texturedImage.Ratio
		newH = maxH
	}

	offsetW := (maxW - newW) / 2.0
	offsetH := (maxH - newH) / 2.0

	// dummyV := giu.Button(strconv.FormatFloat(float64(offsetH), 'f', 0, 32)).Size(120, offsetH)
	// dummyH := giu.Button(strconv.FormatFloat(float64(offsetW), 'f', 0, 32)).Size(offsetW, 20)

	dummyV := giu.Dummy(120, offsetH)
	dummyH := giu.Dummy(offsetW, 20)
	s.ImageWidget.Size(newW, newH)

	giu.Column(
		dummyV,
		giu.Row(dummyH, &s.ImageWidget),
	).Build()
}
