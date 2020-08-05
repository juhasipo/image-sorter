package ui

import (
	"bytes"
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"image"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/logger"
)

var zoomLevels = []uint16{
	5, 10, 25, 36, 50, 75, 80, 90, 100, 110, 120, 130, 150, 175, 200, 300, 400, 500, 1000,
}

type CurrentImageView struct {
	scrolledView   *gtk.ScrolledWindow
	view           *gtk.Image
	image          *common.Handle
	details        *gtk.TextView
	zoomInButton   *gtk.Button
	zoomOutButton  *gtk.Button
	zoomFitButton  *gtk.Button
	zoomLevelLabel *gtk.Label
	imageInstance  image.Image
	zoomIndex      int
	imageChanged   bool
}

func newCurrentImageView(builder *gtk.Builder) *CurrentImageView {
	img, _ := gtk.ImageNew()
	img.SetHExpand(true)
	img.SetVExpand(true)
	img.SetHAlign(gtk.ALIGN_BASELINE)
	img.SetVAlign(gtk.ALIGN_BASELINE)
	img.SetName("current-image")

	view := &CurrentImageView{
		scrolledView:   GetObjectOrPanic(builder, "current-image-window").(*gtk.ScrolledWindow),
		view:           img,
		details:        GetObjectOrPanic(builder, "image-details-view").(*gtk.TextView),
		zoomIndex:      findZoomIndexForValue(100),
		imageChanged:   false,
		zoomInButton:   GetObjectOrPanic(builder, "zoom-in-button").(*gtk.Button),
		zoomOutButton:  GetObjectOrPanic(builder, "zoom-out-button").(*gtk.Button),
		zoomFitButton:  GetObjectOrPanic(builder, "zoom-to-fit-button").(*gtk.Button),
		zoomLevelLabel: GetObjectOrPanic(builder, "zoom-level-label").(*gtk.Label),
	}

	view.scrolledView.Add(img)
	view.scrolledView.ShowAll()

	return view
}

func findZoomIndexForValue(value uint16) int {
	for i := range zoomLevels {
		if zoomLevels[i] == value {
			return i
		}
	}
	logger.Error.Panic("Invalid initial zoom value")
	return 0
}

func getZoomLevelValue(i int) uint16 {
	return zoomLevels[i]
}

func getFormattedZoomLevel(i int) string {
	return fmt.Sprintf("%d %%", getZoomLevelValue(i))
}

func (s *CurrentImageView) zoomIn() {
	s.zoomIndex += 1
	if s.zoomIndex >= len(zoomLevels) {
		s.zoomIndex = len(zoomLevels) - 1
	}
	s.zoomLevelLabel.SetLabel(getFormattedZoomLevel(s.zoomIndex))
}

func (s *CurrentImageView) zoomOut() {
	s.zoomIndex -= 1
	if s.zoomIndex < 0 {
		s.zoomIndex = 0
	}
	s.zoomLevelLabel.SetLabel(getFormattedZoomLevel(s.zoomIndex))
}

func (s *CurrentImageView) zoomToFit() {
	s.zoomIndex = findZoomIndexForValue(100)
	s.zoomLevelLabel.SetLabel(getFormattedZoomLevel(s.zoomIndex))
}

func (s *CurrentImageView) getCurrentZoomLevel() float64 {
	return float64(getZoomLevelValue(s.zoomIndex))
}

func (s *CurrentImageView) UpdateCurrentImage() {
	gtkImage := s.view
	if s.imageInstance != nil {
		fullSize := s.imageInstance.Bounds()
		zoomPercent := s.getCurrentZoomLevel() / 100.0
		targetSize := common.SizeFromWindow(s.scrolledView, zoomPercent)
		targetWidth, targetHeight := common.ScaleToFit(
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
}

const showExifData = false

func (s *CurrentImageView) SetCurrentImage(imageContainer *common.ImageContainer, exifData *common.ExifData) {
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

func getPixbufSize(pixbuf *gdk.Pixbuf) common.Size {
	if pixbuf != nil {
		return common.SizeOf(pixbuf.GetWidth(), pixbuf.GetHeight())
	} else {
		return common.SizeOf(0, 0)
	}
}
