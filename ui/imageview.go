package ui

import (
	"bytes"
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
	"image"
	"sort"
	"strings"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/imageloader"
	"vincit.fi/image-sorter/logger"
)

type CurrentImageView struct {
	scrolledView *gtk.ScrolledWindow
	viewport     *gtk.Viewport
	view         *gtk.Image
	image        *common.Handle
	details      *gtk.TextView
	img          image.Image
}

type ImageList struct {
	layout    *gtk.ScrolledWindow
	component *gtk.IconView
	model     *gtk.ListStore
	images    []*common.ImageContainer
}

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

	imageView := &ImageView{
		currentImage: &CurrentImageView{
			scrolledView: GetObjectOrPanic(builder, "current-image-window").(*gtk.ScrolledWindow),
			viewport:     GetObjectOrPanic(builder, "current-image-view").(*gtk.Viewport),
			view:         GetObjectOrPanic(builder, "current-image").(*gtk.Image),
			details:      GetObjectOrPanic(builder, "image-details-view").(*gtk.TextView),
		},
		nextImages:           nextImagesList,
		prevImages:           prevImagesList,
		imageCache:           ui.imageCache,
		imagesListImageCount: 5,
	}
	tableNew, _ := gtk.TextTagTableNew()
	bufferNew, _ := gtk.TextBufferNew(tableNew)
	imageView.currentImage.details.SetBuffer(bufferNew)
	imageView.currentImage.viewport.Connect("size-allocate", func() {
		ui.UpdateCurrentImage()
		height := ui.imageView.nextImages.component.GetAllocatedHeight() / 80
		if imageView.imagesListImageCount != height {
			imageView.imagesListImageCount = height
			ui.sender.SendToTopicWithData(event.ImageListSizeChanged, height)
		}

	})

	return imageView
}

func (s *ImageView) UpdateCurrentImage() {
	if s.currentImage.img != nil {
		fullSize := s.currentImage.img.Bounds()
		s.currentImage.scrolledView.Remove(s.currentImage.viewport)
		targetSize := common.SizeFromWindow(s.currentImage.scrolledView)
		targetWidth, targetHeight := common.ScaleToFit(
			fullSize.Dx(), fullSize.Dy(),
			targetSize.GetWidth(), targetSize.GetHeight())
		scaled, err := asPixbuf(s.currentImage.img).ScaleSimple(targetWidth, targetHeight, gdk.INTERP_BILINEAR)
		if err != nil {
			logger.Error.Print("Could not load Pixbuf", err)
		}
		s.currentImage.view.SetFromPixbuf(scaled)

		// Hack to prevent image from being center of the scrolled
		// window after minimize. First remove and then add again
		s.currentImage.scrolledView.Add(s.currentImage.viewport)
	} else {
		s.currentImage.view.SetFromPixbuf(nil)
	}
}

type W struct {
	stringBuffer *bytes.Buffer
	values       []string

	exif.Walker
}

func (s *W) Walk(name exif.FieldName, tag *tiff.Tag) error {
	tagValue := strings.Trim(tag.String(), " \t\"")

	if tagValue != "" {
		s.values = append(s.values, fmt.Sprintf("%s: %s", string(name), tagValue))
	}
	return nil
}

func (s *W) String() string {
	sort.Strings(s.values)
	b := bytes.NewBuffer([]byte{})
	for _, value := range s.values {
		b.WriteString(value)
		b.WriteString("\n")
	}
	return b.String()
}

const showExifData = false

func (s *ImageView) SetCurrentImage(imageContainer *common.ImageContainer, exifData *common.ExifData) {
	handle := imageContainer.GetHandle()
	img := imageContainer.GetImage()
	s.currentImage.img = img

	if img != nil {
		size := img.Bounds()
		buffer, _ := s.currentImage.details.GetBuffer()
		stringBuffer := bytes.NewBuffer([]byte{})
		stringBuffer.WriteString(fmt.Sprintf("%s\n%.2f MB (%d x %d)", handle.GetPath(), handle.GetByteSizeMB(), size.Dx(), size.Dy()))

		if showExifData {
			w := &W{stringBuffer: stringBuffer}
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

func (s *ImageList) addImagesToStore(images []*common.ImageContainer) {
	s.model.Clear()
	for _, img := range images {
		iter := s.model.Append()
		if img != nil {
			thumbnail := img.GetImage()
			s.model.SetValue(iter, 0, asPixbuf(thumbnail))
			s.model.SetValue(iter, 1, img.GetHandle().GetId())
		} else {
			s.model.SetValue(iter, 0, nil)
			s.model.SetValue(iter, 1, "")
		}
	}
	s.images = images
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
