package component

import (
	"github.com/gotk3/gotk3/gtk"
	"vincit.fi/image-sorter/api/apitype"
)

type ImageList struct {
	layout    *gtk.ScrolledWindow
	component *gtk.IconView
	model     *gtk.ListStore
	images    []*apitype.ImageContainer
}

func (s *ImageList) addImagesToStore(images []*apitype.ImageContainer) {
	s.model.Clear()
	for _, img := range images {
		iter := s.model.Append()
		if img != nil {
			thumbnail := img.GetImageData()
			s.model.SetValue(iter, 0, asPixbuf(thumbnail))
			s.model.SetValue(iter, 1, img.GetImageFile().GetFile())
		} else {
			s.model.SetValue(iter, 0, nil)
			s.model.SetValue(iter, 1, "")
		}
	}
	s.images = images
}
