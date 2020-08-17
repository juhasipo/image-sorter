package component

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/imageloader"
)

type ViewMode int

const (
	ListView ViewMode = iota
	GridView
)

type ImageView struct {
	currentImage         *CurrentImageView
	nextImages           *ImageList
	prevImages           *ImageList
	imageCache           imageloader.ImageStore
	imageGrid            *ImageGrid
	imagesListImageCount int
	viewMode             ViewMode
}

func NewImageView(builder *gtk.Builder, sender event.Sender, imageCache imageloader.ImageStore) *ImageView {
	nextImagesList := &ImageList{
		layout:    GetObjectOrPanic(builder, "next-images-scrolled-view").(*gtk.ScrolledWindow),
		component: GetObjectOrPanic(builder, "next-images").(*gtk.IconView),
	}
	initializeStore(nextImagesList, VERTICAL, sender)
	prevImagesList := &ImageList{
		layout:    GetObjectOrPanic(builder, "prev-images-scrolled-view").(*gtk.ScrolledWindow),
		component: GetObjectOrPanic(builder, "prev-images").(*gtk.IconView),
	}
	initializeStore(prevImagesList, VERTICAL, sender)

	grid := &ImageGrid{
		layout:    GetObjectOrPanic(builder, "image-grid-scrolled-view").(*gtk.ScrolledWindow),
		component: GetObjectOrPanic(builder, "image-grid-view").(*gtk.IconView),
	}
	grid.initializeStore(sender)

	imageView := &ImageView{
		currentImage:         newCurrentImageView(builder),
		nextImages:           nextImagesList,
		prevImages:           prevImagesList,
		imageCache:           imageCache,
		imageGrid:            grid,
		imagesListImageCount: 5,
		viewMode:             ListView,
	}

	tableNew, _ := gtk.TextTagTableNew()
	bufferNew, _ := gtk.TextBufferNew(tableNew)
	imageView.currentImage.details.SetBuffer(bufferNew)
	imageView.currentImage.scrolledView.Connect("size-allocate", func() {
		imageView.UpdateCurrentImage()
		height := imageView.nextImages.component.GetAllocatedHeight() / 80
		if imageView.imagesListImageCount != height {
			imageView.imagesListImageCount = height
			sender.SendToTopicWithData(event.ImageListSizeChanged, height)
		}
	})

	imageView.currentImage.zoomInButton.Connect("clicked", imageView.ZoomIn)
	imageView.currentImage.zoomOutButton.Connect("clicked", imageView.ZoomOut)
	imageView.currentImage.zoomFitButton.Connect("clicked", imageView.ZoomToFit)

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

func (s *ImageView) AddImageToGrid(images *common.ImageContainer, index int, total int) {
	if index == 0 {
		s.imageGrid.clearImageStore()
	}

	s.imageGrid.addImagesToStore(images)
}

func (s *ImageView) SetNoDistractionMode(value bool) {
	value = !value
	s.nextImages.layout.SetVisible(value)
	s.prevImages.layout.SetVisible(value)
	s.currentImage.details.SetVisible(value)
}

func (s *ImageView) ZoomIn() {
	s.currentImage.zoomIn()
	s.UpdateCurrentImage()
}

func (s *ImageView) ZoomOut() {
	s.currentImage.zoomOut()
	s.UpdateCurrentImage()
}

func (s *ImageView) ZoomToFit() {
	s.currentImage.zoomToFit()
	s.UpdateCurrentImage()
}

func (s *ImageView) GetCurrentHandles() []*common.Handle {
	if s.viewMode == ListView {
		return []*common.Handle{s.currentImage.image}
	} else {
		return s.imageGrid.getSelected()
	}
}

func (s *ImageView) ShowGridView() {
	s.currentImage.Hide()
	s.nextImages.Hide()
	s.prevImages.Hide()
	s.imageGrid.Show()
	s.viewMode = GridView
}

func (s *ImageView) ShowListView() {
	s.currentImage.Show()
	s.nextImages.Show()
	s.prevImages.Show()
	s.imageGrid.Hide()
	s.viewMode = ListView
}

func (s *ImageView) GetViewMode() ViewMode {
	return s.viewMode
}
