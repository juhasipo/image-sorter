package guiapi

import (
	"github.com/AllenDang/giu"
	"vincit.fi/image-sorter/api/apitype"
)

type TexturedImage struct {
	Texture   *giu.Texture
	Image     *apitype.ImageFile
	Width     float32
	Height    float32
	Ratio     float32
	IsLoading bool
}

func NewTexturedImage(image *apitype.ImageFile, texture *giu.Texture) *TexturedImage {
	width := float32(0)
	height := float32(0)
	if image != nil {
		width = float32(image.Width())
		height = float32(image.Height())
	}

	return &TexturedImage{
		Texture:   texture,
		Image:     image,
		Width:     width,
		Height:    height,
		Ratio:     width / height,
		IsLoading: false,
	}
}

func NewEmptyTexturedImage() *TexturedImage {
	return &TexturedImage{
		Texture:   nil,
		Image:     nil,
		Width:     0,
		Height:    0,
		Ratio:     1,
		IsLoading: true,
	}
}

func (s *TexturedImage) IsSame(other *TexturedImage) bool {
	if s == nil && other == nil {
		return true
	} else if s != nil || other != nil {
		return false
	} else if s.Image == nil || other.Image == nil {
		return false
	}

	return s.Image.Id() == other.Image.Id()
}

func (s *TexturedImage) SetLoaded(other *TexturedImage) {
	s.Texture = other.Texture
	s.Image = other.Image
	s.Ratio = other.Ratio
	s.Width = other.Width
	s.Height = other.Height
	s.IsLoading = false
}
