package component

import (
	"bytes"
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"image"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
)

var zoomLevels = []uint16{
	5, 10, 25, 33, 50, 75, 80, 90, 100, 110, 120, 130, 150, 175, 200, 300, 400, 500, 1000,
}

type CurrentImageView struct {
	scrolledView     *gtk.ScrolledWindow
	view             *gtk.Image
	image            *apitype.ImageFile
	details          *gtk.TextView
	zoomInButton     *gtk.Button
	zoomOutButton    *gtk.Button
	zoomFitButton    *gtk.Button
	zoomLevelLabel   *gtk.Label
	imageInstance    image.Image
	zoomIndex        int
	zoomToFixEnabled bool
	imageChanged     bool
}

func newCurrentImageView(builder *gtk.Builder) *CurrentImageView {
	img, _ := gtk.ImageNew()
	img.SetHExpand(true)
	img.SetVExpand(true)
	img.SetHAlign(gtk.ALIGN_BASELINE)
	img.SetVAlign(gtk.ALIGN_BASELINE)
	img.SetName("current-image")

	view := &CurrentImageView{
		scrolledView:     GetObjectOrPanic(builder, "current-image-window").(*gtk.ScrolledWindow),
		view:             img,
		details:          GetObjectOrPanic(builder, "image-details-view").(*gtk.TextView),
		zoomIndex:        findZoomIndexForValue(100),
		zoomToFixEnabled: true,
		imageChanged:     false,
		zoomInButton:     GetObjectOrPanic(builder, "zoom-in-button").(*gtk.Button),
		zoomOutButton:    GetObjectOrPanic(builder, "zoom-out-button").(*gtk.Button),
		zoomFitButton:    GetObjectOrPanic(builder, "zoom-to-fit-button").(*gtk.Button),
		zoomLevelLabel:   GetObjectOrPanic(builder, "zoom-level-label").(*gtk.Label),
	}

	view.scrolledView.Add(img)
	view.scrolledView.ShowAll()

	return view
}

func findZoomIndexForValue(value uint16) int {
	if value < zoomLevels[0] {
		return 0
	}

	for i := range zoomLevels {
		if zoomLevels[i] > value {
			return i - 1
		}
	}

	return len(zoomLevels) - 1
}

func resolveZoomLevelValueAtIndex(i int) uint16 {
	return zoomLevels[i]
}

func (s *CurrentImageView) formattedZoomLevel() string {
	if s.zoomToFixEnabled {
		return fmt.Sprintf("Fit (%d %%)", s.calculatedZoomLevel())
	} else {
		return fmt.Sprintf("%d %%", resolveZoomLevelValueAtIndex(s.zoomIndex))
	}
}

func (s *CurrentImageView) zoomIn() {
	if s.zoomToFixEnabled {
		s.zoomIndex = findZoomIndexForValue(s.calculatedZoomLevel())
	}

	s.zoomToFixEnabled = false
	s.zoomIndex += 1
	if s.zoomIndex >= len(zoomLevels) {
		s.zoomIndex = len(zoomLevels) - 1
	}
}

func (s *CurrentImageView) updateZoomLevelLabel() {
	s.zoomLevelLabel.SetLabel(s.formattedZoomLevel())
}

func (s *CurrentImageView) zoomOut() {
	if s.zoomToFixEnabled {
		s.zoomIndex = findZoomIndexForValue(s.calculatedZoomLevel())
	}

	s.zoomToFixEnabled = false
	s.zoomIndex -= 1
	if s.zoomIndex < 0 {
		s.zoomIndex = 0
	}
}

func (s *CurrentImageView) zoomToFit() {
	s.zoomIndex = findZoomIndexForValue(100)
	s.zoomToFixEnabled = true
}

func (s *CurrentImageView) currentZoomLevel() float64 {
	return float64(resolveZoomLevelValueAtIndex(s.zoomIndex))
}

func (s *CurrentImageView) UpdateCurrentImage() {
	gtkImage := s.view
	if s.imageInstance != nil {
		fullSize := s.imageInstance.Bounds()
		zoomPercent := s.currentZoomLevel() / 100.0
		var targetSize apitype.Size
		if s.zoomToFixEnabled {
			targetSize = apitype.ZoomedSizeFromWindow(s.scrolledView, zoomPercent)
		} else {
			targetSize = apitype.ZoomedSizeFromRectangle(fullSize, zoomPercent)
		}
		targetFitSize := apitype.RectangleOfScaledToFit(fullSize, targetSize)

		pixBufSize := calculatePixbufSize(gtkImage.GetPixbuf())
		if s.imageChanged ||
			(targetFitSize.Width() != pixBufSize.Width() &&
				targetFitSize.Height() != pixBufSize.Height()) {
			s.imageChanged = false
			scaled, err := asPixbuf(s.imageInstance).ScaleSimple(targetFitSize.Width(), targetFitSize.Height(), gdk.INTERP_TILES)
			if err != nil {
				logger.Error.Print("Could not load Pixbuf", err)
			}
			gtkImage.SetFromPixbuf(scaled)
		}
	} else {
		gtkImage.SetFromPixbuf(nil)
	}
	s.updateZoomLevelLabel()
}

const showExifData = false

func (s *CurrentImageView) SetCurrentImage(imageContainer *apitype.ImageFileAndData) {
	s.imageChanged = true
	imageFile := imageContainer.ImageFile()
	imageData := imageContainer.ImageData()
	s.imageInstance = imageData

	if imageData != nil {
		size := imageData.Bounds()
		buffer, _ := s.details.GetBuffer()
		stringBuffer := bytes.NewBuffer([]byte{})
		stringBuffer.WriteString(fmt.Sprintf("%s\n%.2f MB (%d x %d)", imageFile.Path(), imageFile.ByteSizeInMB(), size.Dx(), size.Dy()))

		/*
			if showExifData {
				for key, value := range metaData.MetaData() {
					stringBuffer.WriteString("\n")
					stringBuffer.WriteString(key)
					stringBuffer.WriteString(": ")
					stringBuffer.WriteString(value)
				}
			}
		*/

		buffer.SetText(stringBuffer.String())
		s.image = imageFile
	} else {
		s.image = nil
		buffer, _ := s.details.GetBuffer()
		buffer.SetText("No image")
	}
}

func (s *CurrentImageView) calculatedZoomLevel() uint16 {
	if s.imageInstance != nil && s.view.GetPixbuf() != nil {
		currentSize := s.imageInstance.Bounds()
		width := s.view.GetPixbuf().GetWidth()

		return uint16(float64(width) / float64(currentSize.Dx()) * 100.0)
	} else {
		return 100
	}
}

func asPixbuf(cachedImage image.Image) *gdk.Pixbuf {
	if img, ok := cachedImage.(*image.NRGBA); ok {

		size := img.Bounds()
		const bitsPerSample = 8
		const hasAlpha = true
		pb, err := PixbufNewFromData(
			img.Pix,
			gdk.COLORSPACE_RGB, hasAlpha,
			bitsPerSample,
			size.Dx(), size.Dy(),
			img.Stride)
		if err != nil {
			return nil
		}
		return pb
	}
	return nil
}

func calculatePixbufSize(pixbuf *gdk.Pixbuf) apitype.Size {
	if pixbuf != nil {
		return apitype.SizeOf(pixbuf.GetWidth(), pixbuf.GetHeight())
	} else {
		return apitype.SizeOf(0, 0)
	}
}
