package component

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
)

type ImageGrid struct {
	layout    *gtk.ScrolledWindow
	component *gtk.IconView
	model     *gtk.ListStore
	images    []*common.ImageContainer
}

func (s *ImageGrid) clearImageStore() {
	s.model.Clear()
	s.images = []*common.ImageContainer{}
}

func (s *ImageGrid) addImagesToStore(img *common.ImageContainer) {
	iter := s.model.Append()
	if img != nil {
		s.images = append(s.images, img)
		thumbnail := img.GetImage()
		s.model.SetValue(iter, 0, asPixbuf(thumbnail))
		s.model.SetValue(iter, 1, img.GetHandle().GetId())
	} else {
		s.model.SetValue(iter, 0, nil)
		s.model.SetValue(iter, 1, "")
	}
}

func (s *ImageGrid) initializeStore(sender event.Sender) {
	s.component.Connect("item-activated", func(view *gtk.IconView, path *gtk.TreePath) {
		index := path.GetIndices()[0]
		handle := s.images[index].GetHandle()
		sender.SendToTopicWithData(event.ImageRequest, handle)
	})
	s.model, _ = gtk.ListStoreNew(PixbufGetType(), glib.TYPE_STRING)
	s.component.SetModel(s.model)
	s.component.SetPixbufColumn(0)
}

func (s *ImageGrid) Hide() {
	s.layout.Hide()
}

func (s *ImageGrid) Show() {
	s.layout.Show()
}

func (s *ImageGrid) getSelected() []*common.Handle {
	items := s.component.GetSelectedItems()

	var handles = make([]*common.Handle, items.Length())
	for iter := items.Next(); iter != nil; iter = iter.Next() {
		path := iter.Data().(*gtk.TreePath)
		index := path.GetIndices()[0]
		handle := s.images[index].GetHandle()
		handles[index] = handle
	}
	return handles
}
