package ui

import (
	"fmt"
	"github.com/gotk3/gotk3/gtk"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/pixbuf"
)

type CurrentImageView struct {
	scrolledView *gtk.ScrolledWindow
	viewport     *gtk.Viewport
	view         *gtk.Image
	image        *common.Handle
	details      *gtk.TextView
}

type ImageList struct {
	component *gtk.TreeView
	model     *gtk.ListStore
}

type ImageView struct {
	currentImage *CurrentImageView
	nextImages   *ImageList
	prevImages   *ImageList
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

func (s *ImageView) UpdateCurrentImage(pixbufCache *pixbuf.PixbufCache) {
	size := pixbuf.SizeFromWindow(s.currentImage.scrolledView)
	scaled := pixbufCache.GetScaled(
		s.currentImage.image,
		size,
	)
	s.currentImage.view.SetFromPixbuf(scaled)
	// Hack to prevent image from being center of the scrolled
	// window after minimize
	s.currentImage.scrolledView.Remove(s.currentImage.viewport)
	s.currentImage.scrolledView.Add(s.currentImage.viewport)
}

func (s *ImageView) SetCurrentImage(handle *common.Handle) {
	s.currentImage.image = handle

	buffer, _ := s.currentImage.details.GetBuffer()
	details := fmt.Sprintf("%s\n%.2f MB (%d x %d)", handle.GetPath(), handle.GetByteSizeMB(), handle.GetWidth(), handle.GetHeight())
	buffer.SetText(details)
}

func (s *ImageView) AddImagesToNextStore(images []*common.Handle, pixbufCache *pixbuf.PixbufCache) {
	s.addImagesToStore(s.nextImages, images, pixbufCache)
}

func (s *ImageView) AddImagesToPrevStore(images []*common.Handle, pixbufCache *pixbuf.PixbufCache) {
	s.addImagesToStore(s.prevImages, images, pixbufCache)
}

func (s *ImageView) addImagesToStore(list *ImageList, images []*common.Handle, pixbufCache *pixbuf.PixbufCache) {
	list.model.Clear()
	for _, img := range images {
		iter := list.model.Append()
		list.model.SetValue(iter, 0, pixbufCache.GetThumbnail(img))
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
