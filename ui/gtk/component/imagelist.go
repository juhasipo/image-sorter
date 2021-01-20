package component

import (
	"github.com/gotk3/gotk3/gtk"
	"vincit.fi/image-sorter/api/apitype"
)

type ImageList struct {
	layout    *gtk.ScrolledWindow
	component *gtk.IconView
	model     *gtk.ListStore
	images    []*apitype.ImageFileAndData
}

func (s *ImageList) addImagesToStore(images []*apitype.ImageFileAndData) {
	s.model.Clear()
	for _, img := range images {
		iter := s.model.Append()
		if img != nil {
			thumbnail := img.ImageData()
			s.model.SetValue(iter, 0, asPixbuf(thumbnail))
			s.model.SetValue(iter, 1, img.ImageFile().FileName())
		} else {
			s.model.SetValue(iter, 0, nil)
			s.model.SetValue(iter, 1, "")
		}
	}
	s.images = images
}
