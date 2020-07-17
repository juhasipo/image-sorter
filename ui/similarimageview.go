package ui

import (
	"github.com/gotk3/gotk3/gtk"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/imageloader"
)

type SimilarImagesView struct {
	scrollLayout *gtk.ScrolledWindow
	layout       *gtk.FlowBox
	imageCache   *imageloader.ImageCache
}

func (s *SimilarImagesView) SetImages(handles []*common.ImageContainer, sender event.Sender) {
	children := s.layout.GetChildren()
	children.Foreach(func(item interface{}) {
		s.layout.Remove(item.(gtk.IWidget))
	})
	for _, handle := range handles {
		widget := s.createSimilarImage(handle, sender)
		s.layout.Add(widget)
	}
	s.scrollLayout.SetVisible(true)
	s.scrollLayout.ShowAll()
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

func SimilarImagesViewNew(builder *gtk.Builder, imageCache   *imageloader.ImageCache) *SimilarImagesView {
	layout, _ := gtk.FlowBoxNew()
	similarImagesView := &SimilarImagesView{
		scrollLayout: GetObjectOrPanic(builder, "similar-images-view").(*gtk.ScrolledWindow),
		layout:       layout,
		imageCache:   imageCache,
	}

	similarImagesView.layout.SetMaxChildrenPerLine(10)
	similarImagesView.layout.SetRowSpacing(0)
	similarImagesView.layout.SetColumnSpacing(0)
	similarImagesView.layout.SetSizeRequest(-1, 100)
	similarImagesView.scrollLayout.SetVisible(false)
	similarImagesView.scrollLayout.SetSizeRequest(-1, 100)
	similarImagesView.scrollLayout.Add(layout)

	return similarImagesView
}

