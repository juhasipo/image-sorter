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
	image            *apitype.Handle
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

func getZoomLevelValue(i int) uint16 {
	return zoomLevels[i]
}

func (s *CurrentImageView) getFormattedZoomLevel() string {
	if s.zoomToFixEnabled {
		return fmt.Sprintf("Fit (%d %%)", s.getCalculatedZoomLevel())
	} else {
		return fmt.Sprintf("%d %%", getZoomLevelValue(s.zoomIndex))
	}
}

func (s *CurrentImageView) zoomIn() {
	if s.zoomToFixEnabled {
		s.zoomIndex = findZoomIndexForValue(s.getCalculatedZoomLevel())
	}

	s.zoomToFixEnabled = false
	s.zoomIndex += 1
	if s.zoomIndex >= len(zoomLevels) {
		s.zoomIndex = len(zoomLevels) - 1
	}
}

func (s *CurrentImageView) updateZoomLevelLabel() {
	s.zoomLevelLabel.SetLabel(s.getFormattedZoomLevel())
}

func (s *CurrentImageView) zoomOut() {
	if s.zoomToFixEnabled {
		s.zoomIndex = findZoomIndexForValue(s.getCalculatedZoomLevel())
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

func (s *CurrentImageView) getCurrentZoomLevel() float64 {
	return float64(getZoomLevelValue(s.zoomIndex))
}

func (s *CurrentImageView) UpdateCurrentImage() {
	gtkImage := s.view
	if s.imageInstance != nil {
		fullSize := s.imageInstance.Bounds()
		zoomPercent := s.getCurrentZoomLevel() / 100.0
		var targetSize apitype.Size
		if s.zoomToFixEnabled {
			targetSize = apitype.SizeFromWindow(s.scrolledView, zoomPercent)
		} else {
			targetSize = apitype.SizeFromRectangle(fullSize, zoomPercent)
		}
		targetWidth, targetHeight := apitype.ScaleToFit(
			fullSize.Dx(), fullSize.Dy(),
			targetSize.GetWidth(), targetSize.GetHeight())

		pixBufSize := getPixbufSize(gtkImage.GetPixbuf())
		if s.imageChanged ||
			(targetWidth != pixBufSize.GetWidth() &&
				targetHeight != pixBufSize.GetHeight()) {
			s.imageChanged = false
			scaled, err := asPixbuf(s.imageInstance).ScaleSimple(targetWidth, targetHeight, gdk.INTERP_TILES)
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

func (s *CurrentImageView) SetCurrentImage(imageContainer *apitype.ImageContainer, exifData *apitype.ExifData) {
	s.imageChanged = true
	handle := imageContainer.GetHandle()
	img := imageContainer.GetImage()
	s.imageInstance = img

	if img != nil {
		size := img.Bounds()
		buffer, _ := s.details.GetBuffer()
		stringBuffer := bytes.NewBuffer([]byte{})
		stringBuffer.WriteString(fmt.Sprintf("%s\n%.2f MB (%d x %d)", handle.GetPath(), handle.GetByteSizeMB(), size.Dx(), size.Dy()))

		if showExifData {
			w := &ExifWalker{stringBuffer: stringBuffer}
			exifData.Walk(w)
			stringBuffer.WriteString("\n")
			stringBuffer.WriteString(w.String())
		}

		buffer.SetText(stringBuffer.String())
		s.image = handle
	} else {
		s.image = nil
		buffer, _ := s.details.GetBuffer()
		buffer.SetText("No image")
	}
}

func (s *CurrentImageView) getCalculatedZoomLevel() uint16 {
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

func getPixbufSize(pixbuf *gdk.Pixbuf) apitype.Size {
	if pixbuf != nil {
		return apitype.SizeOf(pixbuf.GetWidth(), pixbuf.GetHeight())
	} else {
		return apitype.SizeOf(0, 0)
	}
}
