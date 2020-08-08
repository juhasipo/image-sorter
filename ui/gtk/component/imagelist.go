package component

import (
	"github.com/gotk3/gotk3/gtk"
	"vincit.fi/image-sorter/common"
)

type ImageList struct {
	layout    *gtk.ScrolledWindow
	component *gtk.IconView
	model     *gtk.ListStore
	images    []*common.ImageContainer
}

func (s *ImageList) addImagesToStore(images []*common.ImageContainer) {
	s.model.Clear()
	for _, img := range images {
		iter := s.model.Append()
		if img != nil {
			thumbnail := img.GetImage()
			s.model.SetValue(iter, 0, asPixbuf(thumbnail))
			s.model.SetValue(iter, 1, img.GetHandle().GetId())
		} else {
			s.model.SetValue(iter, 0, nil)
			s.model.SetValue(iter, 1, "")
		}
	}
	s.images = images
}
