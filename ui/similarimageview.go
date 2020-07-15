package ui

import (
	"github.com/gotk3/gotk3/gtk"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/pixbuf"
)

type SimilarImagesView struct {
	scrollLayout *gtk.ScrolledWindow
	layout       *gtk.FlowBox
}

func (s *SimilarImagesView) SetImages(handles []*common.Handle, pixbufCache *pixbuf.PixbufCache, sender event.Sender) {
	children := s.layout.GetChildren()
	children.Foreach(func(item interface{}) {
		s.layout.Remove(item.(gtk.IWidget))
	})
	for _, handle := range handles {
		widget := s.createSimilarImage(handle, pixbufCache, sender)
		s.layout.Add(widget)
	}
	s.scrollLayout.SetVisible(true)
	s.scrollLayout.ShowAll()
}

func (s *SimilarImagesView) createSimilarImage(handle *common.Handle, pixbufCache *pixbuf.PixbufCache, sender event.Sender) *gtk.EventBox {
	eventBox, _ := gtk.EventBoxNew()
	imageWidget, _ := gtk.ImageNewFromPixbuf(pixbufCache.GetThumbnail(handle))
	eventBox.Add(imageWidget)
	eventBox.Connect("button-press-event", func() {
		sender.SendToTopicWithData(event.IMAGE_REQUEST, handle)
	})
	return eventBox
}

func SimilarImagesViewNew(builder *gtk.Builder) *SimilarImagesView {
	layout, _ := gtk.FlowBoxNew()
	similarImagesView := &SimilarImagesView{
		scrollLayout: GetObjectOrPanic(builder, "similar-images-view").(*gtk.ScrolledWindow),
		layout:       layout,
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

