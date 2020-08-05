package ui

import (
	"bytes"
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"image"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/imageloader"
	"vincit.fi/image-sorter/logger"
)

type ImageView struct {
	currentImage         *CurrentImageView
	nextImages           *ImageList
	prevImages           *ImageList
	imageCache           imageloader.ImageStore
	imagesListImageCount int
}

func NewImageView(builder *gtk.Builder, ui *Ui) *ImageView {
	nextImagesList := &ImageList{
		layout:    GetObjectOrPanic(builder, "next-images-scrolled-view").(*gtk.ScrolledWindow),
		component: GetObjectOrPanic(builder, "next-images").(*gtk.IconView),
	}
	initializeStore(nextImagesList, VERTICAL, ui.sender)
	prevImagesList := &ImageList{
		layout:    GetObjectOrPanic(builder, "prev-images-scrolled-view").(*gtk.ScrolledWindow),
		component: GetObjectOrPanic(builder, "prev-images").(*gtk.IconView),
	}
	initializeStore(prevImagesList, VERTICAL, ui.sender)

	img, _ := gtk.ImageNew()
	img.SetHExpand(true)
	img.SetVExpand(true)
	img.SetHAlign(gtk.ALIGN_BASELINE)
	img.SetVAlign(gtk.ALIGN_BASELINE)
	img.SetName("current-image")

	imageView := &ImageView{
		currentImage: &CurrentImageView{
			scrolledView: GetObjectOrPanic(builder, "current-image-window").(*gtk.ScrolledWindow),
			view:         img,
			details:      GetObjectOrPanic(builder, "image-details-view").(*gtk.TextView),
			zoomLevel:    100,
			imageChanged: false,
		},
		nextImages:           nextImagesList,
		prevImages:           prevImagesList,
		imageCache:           ui.imageCache,
		imagesListImageCount: 5,
	}
	imageView.currentImage.scrolledView.Add(img)
	imageView.currentImage.scrolledView.ShowAll()
	tableNew, _ := gtk.TextTagTableNew()
	bufferNew, _ := gtk.TextBufferNew(tableNew)
	imageView.currentImage.details.SetBuffer(bufferNew)
	imageView.currentImage.scrolledView.Connect("size-allocate", func() {
		ui.UpdateCurrentImage()
		height := ui.imageView.nextImages.component.GetAllocatedHeight() / 80
		if imageView.imagesListImageCount != height {
			imageView.imagesListImageCount = height
			ui.sender.SendToTopicWithData(event.ImageListSizeChanged, height)
		}
	})

	return imageView
}

func initializeStore(imageList *ImageList, layout Layout, sender event.Sender) {
	const requestedSize = 102
	if layout == HORIZONTAL {
		imageList.component.SetSizeRequest(-1, requestedSize)
	} else {
		imageList.component.SetSizeRequest(requestedSize, -1)
	}

	imageList.component.Connect("item-activated", func(view *gtk.IconView, path *gtk.TreePath) {
		index := path.GetIndices()[0]
		handle := imageList.images[index].GetHandle()
		sender.SendToTopicWithData(event.ImageRequest, handle)
	})
	imageList.model, _ = gtk.ListStoreNew(PixbufGetType(), glib.TYPE_STRING)
	imageList.component.SetModel(imageList.model)
	imageList.component.SetPixbufColumn(0)
}

func (s *ImageView) UpdateCurrentImage() {
	gtkImage := s.currentImage.view
	if s.currentImage.imageInstance != nil {
		fullSize := s.currentImage.imageInstance.Bounds()
		zoomPercent := float64(s.currentImage.zoomLevel) / 100.0
		targetSize := common.SizeFromWindow(s.currentImage.scrolledView, zoomPercent)
		targetWidth, targetHeight := common.ScaleToFit(
			fullSize.Dx(), fullSize.Dy(),
			targetSize.GetWidth(), targetSize.GetHeight())

		pixBufSize := getPixbufSize(gtkImage.GetPixbuf())
		if s.currentImage.imageChanged ||
			(targetWidth != pixBufSize.GetWidth() &&
				targetHeight != pixBufSize.GetHeight()) {
			s.currentImage.imageChanged = false
			//s.currentImage.scrolledView.Remove(gtkImage)
			scaled, err := asPixbuf(s.currentImage.imageInstance).ScaleSimple(targetWidth, targetHeight, gdk.INTERP_TILES)
			if err != nil {
				logger.Error.Print("Could not load Pixbuf", err)
			}

			logger.Info.Print("Render ", targetHeight, "x", targetWidth)
			gtkImage.SetFromPixbuf(scaled)

			// Hack to prevent image from being center of the scrolled
			// window after minimize. First remove and then add again
			//s.currentImage.scrolledView.Add(gtkImage)
			//s.currentImage.scrolledView.GetVAdjustment().SetValue(0.0)
			//s.currentImage.scrolledView.GetHAdjustment().SetValue(0.0)
		}
	} else {
		gtkImage.SetFromPixbuf(nil)
	}
}

func getPixbufSize(pixbuf *gdk.Pixbuf) common.Size {
	if pixbuf != nil {
		return common.SizeOf(pixbuf.GetWidth(), pixbuf.GetHeight())
	} else {
		return common.SizeOf(0, 0)
	}
}

const showExifData = false

func (s *ImageView) SetCurrentImage(imageContainer *common.ImageContainer, exifData *common.ExifData) {
	s.currentImage.imageChanged = true
	handle := imageContainer.GetHandle()
	img := imageContainer.GetImage()
	s.currentImage.imageInstance = img

	if img != nil {
		size := img.Bounds()
		buffer, _ := s.currentImage.details.GetBuffer()
		stringBuffer := bytes.NewBuffer([]byte{})
		stringBuffer.WriteString(fmt.Sprintf("%s\n%.2f MB (%d x %d)", handle.GetPath(), handle.GetByteSizeMB(), size.Dx(), size.Dy()))

		if showExifData {
			w := &ExifWalker{stringBuffer: stringBuffer}
			exifData.Walk(w)
			stringBuffer.WriteString("\n")
			stringBuffer.WriteString(w.String())
		}

		buffer.SetText(stringBuffer.String())
		s.currentImage.image = handle
	} else {
		s.currentImage.image = nil
		buffer, _ := s.currentImage.details.GetBuffer()
		buffer.SetText("No image")
	}
}

func (s *ImageView) AddImagesToNextStore(images []*common.ImageContainer) {
	s.nextImages.addImagesToStore(images)
}

func (s *ImageView) AddImagesToPrevStore(images []*common.ImageContainer) {
	s.prevImages.addImagesToStore(images)
}

func (s *ImageView) SetNoDistractionMode(value bool) {
	value = !value
	s.nextImages.layout.SetVisible(value)
	s.prevImages.layout.SetVisible(value)
	s.currentImage.details.SetVisible(value)
}

func (s *ImageView) zoomIn() {
	s.currentImage.zoomLevel += 10
	if s.currentImage.zoomLevel >= 1000 {
		s.currentImage.zoomLevel = 1000
	}
	s.UpdateCurrentImage()
}

func (s *ImageView) zoomOut() {
	s.currentImage.zoomLevel -= 10
	if s.currentImage.zoomLevel < 10 {
		s.currentImage.zoomLevel = 10
	}
	s.UpdateCurrentImage()
}

func (s *ImageView) zoomToFit() {
	s.currentImage.zoomLevel = 100
	s.UpdateCurrentImage()
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
