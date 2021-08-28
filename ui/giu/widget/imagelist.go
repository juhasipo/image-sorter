package widget

import (
	"github.com/AllenDang/giu"
)

type ImageListWidget struct {
	images    []*TexturedImage
	showLabel bool
}

func ImageList(images []*TexturedImage, showLabel bool) *ImageListWidget {
	return &ImageListWidget{
		images:    images,
		showLabel: showLabel,
	}
}

func (s *ImageListWidget) Build() {
	var w []giu.Widget
	for _, nextImage := range s.images {
		maxWidth := float32(120.0)
		height := nextImage.Height / nextImage.Width * maxWidth
		w = append(w, giu.Image(nextImage.Texture).Size(maxWidth, height))
		if s.showLabel {
			w = append(w, giu.Label(nextImage.Image.FileName()))
		}
	}

	giu.Column(w...).Build()
}
