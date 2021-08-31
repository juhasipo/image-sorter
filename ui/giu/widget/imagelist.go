package widget

import (
	"github.com/AllenDang/giu"
)

type ImageListWidget struct {
	images    []*TexturedImage
	showLabel bool
	height    float32
}

func ImageList(images []*TexturedImage, showLabel bool, height float32) *ImageListWidget {
	return &ImageListWidget{
		images:    images,
		showLabel: showLabel,
		height:    height,
	}
}

func (s *ImageListWidget) Build() {
	maxWidth := float32(120.0)
	var w []giu.Widget
	for _, nextImage := range s.images {
		height := nextImage.Height / nextImage.Width * maxWidth
		w = append(w, giu.Image(nextImage.Texture).Size(maxWidth, height))
		if s.showLabel {
			w = append(w, giu.Label(nextImage.Image.FileName()))
		}
	}

	giu.PushItemSpacing(0, 8)
	giu.Child().
		Layout(w...).
		Border(false).
		Size(maxWidth, s.height).
		Flags(giu.WindowFlagsNoScrollbar).
		Build()
	giu.PopStyle()
}
