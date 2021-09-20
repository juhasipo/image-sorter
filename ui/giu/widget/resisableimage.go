package widget

import (
	"github.com/AllenDang/giu"
	"image/color"
	"time"
	"vincit.fi/image-sorter/ui/giu/guiapi"
)

type ResizableImageWidget struct {
	texturedImage           *TexturedImage
	maxHeight, maxWidth     float32
	zoomFactor              float32
	zoomMode                guiapi.ZoomMode
	imageWidth, imageHeight float32
	currentActualZoom       float32
	tintScale               float32
	delta                   time.Time
	onZoomIn                func()
	onZoomOut               func()
	giu.ImageWidget
}

func ResizableImage(image *TexturedImage) *ResizableImageWidget {
	return &ResizableImageWidget{
		texturedImage: image,
		maxHeight:     0,
		maxWidth:      0,
		zoomFactor:    1,
		zoomMode:      guiapi.ZoomFit,
		ImageWidget:   *giu.Image(image.Texture()),
		tintScale:     1,
		delta:         time.Now(),
	}
}

func (s *ResizableImageWidget) Size(width float32, height float32) *ResizableImageWidget {
	s.ImageWidget.Size(width, height)
	s.maxWidth = width
	s.maxHeight = height
	giu.Update()
	return s
}

func (s *ResizableImageWidget) ImageSize(width float32, height float32) *ResizableImageWidget {
	s.imageWidth = width
	s.imageHeight = height
	giu.Update()
	return s
}

func (s *ResizableImageWidget) ZoomFactor(zoomFactor float32, zoomMode guiapi.ZoomMode) *ResizableImageWidget {
	s.zoomFactor = zoomFactor
	s.zoomMode = zoomMode
	return s
}

func (s *ResizableImageWidget) SetZoomHandlers(onZoomIn func(), onZoomOut func()) {
	s.onZoomIn = onZoomIn
	s.onZoomOut = onZoomOut
}

func (s *ResizableImageWidget) CurrentActualZoom() float32 {
	if s.imageWidth > 0 {
		return s.currentActualZoom
	} else {
		return 1
	}
}

func (s *ResizableImageWidget) Build() {
	maxW, maxH := giu.GetAvailableRegion()

	var newW float32
	var newH float32
	if s.zoomMode == guiapi.ZoomFit {
		// Check if area size is limited
		// If yes, then those should be used for offset calculation
		// and the image size calculation
		if s.maxWidth != 0 && s.maxHeight != 0 {
			maxW = s.maxWidth
			maxH = s.maxHeight
		}

		newW = maxW
		newH = newW / s.texturedImage.Ratio()

		if newH > maxH {
			newW = maxH * s.texturedImage.Ratio()
			newH = maxH
		}
		s.currentActualZoom = newW / s.imageWidth
	} else { // Show zoomed image => Display area size doesn't affect the image size
		// Image size should be provided for the zoom to work
		newW = s.imageWidth * s.zoomFactor
		newH = newW / s.texturedImage.Ratio()
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

	tintValue := s.resolveTintValue()
	s.ImageWidget.TintColor(color.RGBA{
		R: tintValue,
		G: tintValue,
		B: tintValue,
		A: tintValue,
	})

	if giu.IsKeyDown(giu.KeyLeftControl) || giu.IsKeyDown(giu.KeyRightControl) {
		// TODO: Somehow prevent image scroll when zooming
		delta := giu.Context.IO().GetMouseWheelDelta()
		if delta > 0 {
			if s.onZoomIn != nil {
				s.onZoomIn()
			}
		} else if delta < 0 {
			if s.onZoomOut != nil {
				s.onZoomOut()
			}
		}
	}

	giu.Column(
		dummyV,
		giu.Row(dummyH, &s.ImageWidget, dummyH),
		dummyV,
	).Build()

}

func (s *ResizableImageWidget) resolveTintValue() uint8 {
	if s.texturedImage.NewImageLoaded() {
		if s.tintScale < 1 {
			s.tintScale += 0.1
			giu.Update()
		}
		if s.tintScale > 1 {
			s.tintScale = 1
			giu.Update()
		}
	} else {
		if s.tintScale > 0 {
			s.tintScale -= 0.1
			giu.Update()
		}
		if s.tintScale < 0 {
			s.tintScale = 0
			giu.Update()
		}
	}

	tintValue := uint8(255 * s.tintScale)
	return tintValue
}

func (s *ResizableImageWidget) UpdateImage(texture *TexturedImage) {
	if s != nil {
		s.texturedImage = texture
		s.ImageWidget = *giu.Image(texture.Texture())
		giu.Update()
	}
}
