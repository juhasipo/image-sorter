package component

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
)

type ImageView struct {
	currentImage         *CurrentImageView
	nextImages           *ImageList
	prevImages           *ImageList
	imageCache           api.ImageStore
	imagesListImageCount int
}

func NewImageView(builder *gtk.Builder, sender api.Sender, imageCache api.ImageStore) *ImageView {
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

	imageView := &ImageView{
		currentImage:         newCurrentImageView(builder),
		nextImages:           nextImagesList,
		prevImages:           prevImagesList,
		imageCache:           imageCache,
		imagesListImageCount: 5,
	}

	tableNew, _ := gtk.TextTagTableNew()
	bufferNew, _ := gtk.TextBufferNew(tableNew)
	imageView.currentImage.details.SetBuffer(bufferNew)
	imageView.currentImage.scrolledView.Connect("size-allocate", func() {
		imageView.UpdateCurrentImage()
		height := imageView.nextImages.component.GetAllocatedHeight() / 80
		if imageView.imagesListImageCount != height {
			imageView.imagesListImageCount = height
			sender.SendCommandToTopic(api.ImageListSizeChanged, &api.ImageListCommand{
				ImageListSize: height,
			})
		}
	})

	imageView.currentImage.zoomInButton.Connect("clicked", imageView.ZoomIn)
	imageView.currentImage.zoomOutButton.Connect("clicked", imageView.ZoomOut)
	imageView.currentImage.zoomFitButton.Connect("clicked", imageView.ZoomToFit)

	return imageView
}

func initializeStore(imageList *ImageList, layout Layout, sender api.Sender) {
	const requestedSize = 102
	if layout == HORIZONTAL {
		imageList.component.SetSizeRequest(-1, requestedSize)
	} else {
		imageList.component.SetSizeRequest(requestedSize, -1)
	}

	imageList.component.Connect("item-activated", func(view *gtk.IconView, path *gtk.TreePath) {
		index := path.GetIndices()[0]
		imageFile := imageList.images[index].ImageFile()
		sender.SendCommandToTopic(api.ImageRequest, &api.ImageQuery{
			Id: imageFile.Id(),
		})
	})
	imageList.model, _ = gtk.ListStoreNew(PixbufGetType(), glib.TYPE_STRING)
	imageList.component.SetModel(imageList.model)
	imageList.component.SetPixbufColumn(0)
}

func (s *ImageView) UpdateCurrentImage() {
	s.currentImage.UpdateCurrentImage()
}

func (s *ImageView) SetCurrentImage(imageContainer *apitype.ImageFileAndData, metaData *apitype.ImageMetaData) {
	s.currentImage.SetCurrentImage(imageContainer, metaData)
}

func (s *ImageView) AddImagesToNextStore(images []*apitype.ImageFileAndData) {
	s.nextImages.addImagesToStore(images)
}

func (s *ImageView) AddImagesToPrevStore(images []*apitype.ImageFileAndData) {
	s.prevImages.addImagesToStore(images)
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

func (s *ImageView) CurrentImageFile() *apitype.ImageFile {
	return s.currentImage.image
}
