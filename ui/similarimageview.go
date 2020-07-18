package ui

import (
	"github.com/gotk3/gotk3/gtk"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/imageloader"
)

type SimilarImagesView struct {
	view *gtk.Box
	scrollLayout *gtk.ScrolledWindow
	list         *ImageList
	imageCache   *imageloader.ImageCache
	closeButton  *gtk.Button
	sender       event.Sender
}

func SimilarImagesViewNew(builder *gtk.Builder, sender event.Sender, imageCache *imageloader.ImageCache) *SimilarImagesView {
	imageList := &ImageList{component: GetObjectOrPanic(builder, "similar-images-list").(*gtk.IconView)}
	initializeStore(imageList, HORIZONTAL, sender)

	similarImagesView := &SimilarImagesView{
		view: GetObjectOrPanic(builder, "similar-images-view").(*gtk.Box),
		scrollLayout: GetObjectOrPanic(builder, "similar-images-scrolled-view").(*gtk.ScrolledWindow),
		list: imageList,
		closeButton:  GetObjectOrPanic(builder, "similar-images-close-button").(*gtk.Button),
		imageCache:   imageCache,
		sender:       sender,
	}

	similarImagesView.scrollLayout.SetVisible(false)
	similarImagesView.scrollLayout.SetSizeRequest(-1, 110)

	similarImagesView.closeButton.Connect("clicked", func() {
		similarImagesView.view.SetVisible(false)
		sender.SendToTopicWithData(event.SIMILAR_SET_STATUS, false)
	})

	return similarImagesView
}

func (s *SimilarImagesView) SetImages(handles []*common.ImageContainer, sender event.Sender) {
	s.list.addImagesToStore(handles)
	s.view.SetVisible(true)
	s.view.ShowAll()
}

func (s *SimilarImagesView) createSimilarImage(handle *common.ImageContainer, sender event.Sender) *gtk.EventBox {
	eventBox, _ := gtk.EventBoxNew()
	thumbnail := handle.GetImage()
	imageWidget, _ := gtk.ImageNewFromPixbuf(asPixbuf(thumbnail))
	eventBox.Add(imageWidget)
	eventBox.Connect("button-press-event", func() {
		sender.SendToTopicWithData(event.IMAGE_REQUEST, handle.GetHandle())
	})
	return eventBox
}
