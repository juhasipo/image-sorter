package ui

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/imageloader"
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

	imageView := &ImageView{
		currentImage:         newCurrentImageView(builder),
		nextImages:           nextImagesList,
		prevImages:           prevImagesList,
		imageCache:           ui.imageCache,
		imagesListImageCount: 5,
	}

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

	imageView.currentImage.zoomInButton.Connect("clicked", imageView.zoomIn)
	imageView.currentImage.zoomOutButton.Connect("clicked", imageView.zoomOut)
	imageView.currentImage.zoomFitButton.Connect("clicked", imageView.zoomToFit)

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
	s.currentImage.UpdateCurrentImage()
}

func (s *ImageView) SetCurrentImage(imageContainer *common.ImageContainer, exifData *common.ExifData) {
	s.currentImage.SetCurrentImage(imageContainer, exifData)
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
	s.currentImage.zoomIn()
	s.UpdateCurrentImage()
}

func (s *ImageView) zoomOut() {
	s.currentImage.zoomOut()
	s.UpdateCurrentImage()
}

func (s *ImageView) zoomToFit() {
	s.currentImage.zoomToFit()
	s.UpdateCurrentImage()
}
