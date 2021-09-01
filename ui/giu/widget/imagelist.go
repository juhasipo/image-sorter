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

func (s *ImageListWidget) SetHeight(height float32) {
	s.height = height
}

func (s *ImageListWidget) SetImages(images []*TexturedImage) {
	var newImageList []*TexturedImage
	var changedImages []*TexturedImage

	// Very naive algorithm to find out which images
	// need to be reloaded and which can be re-used
	for _, image := range images {
		var found *TexturedImage = nil
		for _, texturedImage := range s.images {
			if image.IsSame(texturedImage) {
				found = texturedImage
			}
		}

		if found == nil {
			changedImages = append(changedImages, image)
			newImageList = append(newImageList, image)
		} else {
			newImageList = append(newImageList, found)
		}
	}

	s.images = newImageList

	for _, texturedImage := range changedImages {
		texturedImage.LoadImageAsTextureThumbnail()
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
