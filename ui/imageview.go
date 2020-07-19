package ui

import (
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"image"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/imageloader"
)

type CurrentImageView struct {
	scrolledView *gtk.ScrolledWindow
	viewport     *gtk.Viewport
	view         *gtk.Image
	image        *common.Handle
	details      *gtk.TextView
	pixbuf       *gdk.Pixbuf
}

type ImageList struct {
	component *gtk.IconView
	model     *gtk.ListStore
	images    []*common.ImageContainer
}

type ImageView struct {
	currentImage *CurrentImageView
	nextImages   *ImageList
	prevImages   *ImageList
	imageCache   *imageloader.ImageCache
}

func ImageViewNew(builder *gtk.Builder, ui *Ui) *ImageView {
	nextImagesList := &ImageList{component: GetObjectOrPanic(builder, "next-images").(*gtk.IconView)}
	initializeStore(nextImagesList, VERTICAL, ui.sender)
	prevImagesList := &ImageList{component: GetObjectOrPanic(builder, "prev-images").(*gtk.IconView)}
	initializeStore(prevImagesList, VERTICAL, ui.sender)

	imageView := &ImageView{
		currentImage: &CurrentImageView{
			scrolledView: GetObjectOrPanic(builder, "current-image-window").(*gtk.ScrolledWindow),
			viewport:     GetObjectOrPanic(builder, "current-image-view").(*gtk.Viewport),
			view:         GetObjectOrPanic(builder, "current-image").(*gtk.Image),
			details:      GetObjectOrPanic(builder, "image-details-view").(*gtk.TextView),
		},
		nextImages: nextImagesList,
		prevImages: prevImagesList,
		imageCache: ui.imageCache,
	}
	tableNew, _ := gtk.TextTagTableNew()
	bufferNew, _ := gtk.TextBufferNew(tableNew)
	imageView.currentImage.details.SetBuffer(bufferNew)
	imageView.currentImage.viewport.Connect("size-allocate", func() {
		ui.UpdateCurrentImage()
		height := ui.imageView.nextImages.component.GetAllocatedHeight() / 80
		ui.sender.SendToTopicWithData(event.IMAGE_LIST_SIZE_CHANGED, height)
	})

	return imageView
}

func (s *ImageView) UpdateCurrentImage() {
	if s.currentImage.pixbuf != nil {
		s.currentImage.scrolledView.Remove(s.currentImage.viewport)
		size := common.SizeFromWindow(s.currentImage.scrolledView)
		w, h := common.ScaleToFit(s.currentImage.pixbuf.GetWidth(), s.currentImage.pixbuf.GetHeight(),
			size.GetWidth(), size.GetHeight())
		scaled, _ := s.currentImage.pixbuf.ScaleSimple(w, h, gdk.INTERP_BILINEAR)
		s.currentImage.view.SetFromPixbuf(scaled)

		// Hack to prevent image from being center of the scrolled
		// window after minimize. First remove and then add again
		s.currentImage.scrolledView.Add(s.currentImage.viewport)
	}
}

func (s *ImageView) SetCurrentImage(imageContainer *common.ImageContainer) {
	handle := imageContainer.GetHandle()
	img := imageContainer.GetImage()
	full := img
	s.currentImage.pixbuf = asPixbuf(full)

	if img != nil {
		size := img.Bounds()
		buffer, _ := s.currentImage.details.GetBuffer()
		details := fmt.Sprintf("%s\n%.2f MB (%d x %d)", handle.GetPath(), handle.GetByteSizeMB(), size.Dx(), size.Dy())
		buffer.SetText(details)
		s.currentImage.image = handle
	} else {
		s.currentImage.image = nil
	}
}

func (s *ImageView) AddImagesToNextStore(images []*common.ImageContainer) {
	s.nextImages.addImagesToStore(images)
}

func (s *ImageView) AddImagesToPrevStore(images []*common.ImageContainer) {
	s.prevImages.addImagesToStore(images)
}

func (s *ImageList) addImagesToStore(images []*common.ImageContainer) {
	s.model.Clear()
	for _, img := range images {
		iter := s.model.Append()
		thumbnail := img.GetImage()
		s.model.SetValue(iter, 0, asPixbuf(thumbnail))
		s.model.SetValue(iter, 1, img.GetHandle().GetId())
	}
	s.images = images
}

func initializeStore(imageList *ImageList, layout Layout, sender event.Sender) {
	const requestedSize = 100
	if layout == HORIZONTAL {
		imageList.component.SetSizeRequest(-1, requestedSize)
	} else {
		imageList.component.SetSizeRequest(requestedSize, -1)
	}

	imageList.component.Connect("item-activated", func(view *gtk.IconView, path *gtk.TreePath) {
		index := path.GetIndices()[0]
		handle := imageList.images[index].GetHandle()
		sender.SendToTopicWithData(event.IMAGE_REQUEST, handle)
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
