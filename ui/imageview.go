package ui

import (
	"fmt"
	"github.com/gotk3/gotk3/gdk"
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
	component *gtk.TreeView
	model     *gtk.ListStore
}

type ImageView struct {
	currentImage *CurrentImageView
	nextImages   *ImageList
	prevImages   *ImageList
	imageCache   *imageloader.ImageCache
}

func ImageViewNew(builder *gtk.Builder, ui *Ui) *ImageView {
	nextImagesList := GetObjectOrPanic(builder, "next-images").(*gtk.TreeView)
	nextImageStore := createImageList(nextImagesList, "Next images", FORWARD, ui.sender)
	prevImagesList := GetObjectOrPanic(builder, "prev-images").(*gtk.TreeView)
	prevImageStore := createImageList(prevImagesList, "Prev images", BACKWARD, ui.sender)
	imageView := &ImageView{
		currentImage: &CurrentImageView{
			scrolledView: GetObjectOrPanic(builder, "current-image-window").(*gtk.ScrolledWindow),
			viewport:     GetObjectOrPanic(builder, "current-image-view").(*gtk.Viewport),
			view:         GetObjectOrPanic(builder, "current-image").(*gtk.Image),
			details:      GetObjectOrPanic(builder, "image-details-view").(*gtk.TextView),
		},
		nextImages: &ImageList{
			component: nextImagesList,
			model:     nextImageStore,
		},
		prevImages: &ImageList{
			component: prevImagesList,
			model:     prevImageStore,
		},
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
	if s.currentImage.image != handle {
		full := img
		s.currentImage.pixbuf = asPixbuf(full)

		size := img.Bounds()
		buffer, _ := s.currentImage.details.GetBuffer()
		details := fmt.Sprintf("%s\n%.2f MB (%d x %d)", handle.GetPath(), handle.GetByteSizeMB(), size.Dx(), size.Dy())
		buffer.SetText(details)
		s.currentImage.image = handle
	}
}

func (s *ImageView) AddImagesToNextStore(images []*common.ImageContainer, imageCache *imageloader.ImageCache) {
	s.addImagesToStore(s.nextImages, images)
}

func (s *ImageView) AddImagesToPrevStore(images []*common.ImageContainer, imageCache *imageloader.ImageCache) {
	s.addImagesToStore(s.prevImages, images)
}

func (s *ImageView) addImagesToStore(list *ImageList, images []*common.ImageContainer) {
	list.model.Clear()
	for _, img := range images {
		iter := list.model.Append()
		thumbnail := img.GetImage()
		list.model.SetValue(iter, 0, asPixbuf(thumbnail))
	}
}

func createImageList(view *gtk.TreeView, title string, direction Direction, sender event.Sender) *gtk.ListStore {
	view.SetSizeRequest(100, -1)
	view.Connect("row-activated", func(view *gtk.TreeView, path *gtk.TreePath, col *gtk.TreeViewColumn) {
		index := path.GetIndices()[0] + 1
		if direction == FORWARD {
			sender.SendToTopicWithData(event.IMAGE_REQUEST_NEXT_OFFSET, index)
		} else {
			sender.SendToTopicWithData(event.IMAGE_REQUEST_PREV_OFFSET, index)
		}
	})
	store, _ := gtk.ListStoreNew(PixbufGetType())
	view.SetModel(store)
	renderer, _ := gtk.CellRendererPixbufNew()
	column, _ := gtk.TreeViewColumnNewWithAttribute(title, renderer, "pixbuf", 0)
	view.AppendColumn(column)
	return store
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
