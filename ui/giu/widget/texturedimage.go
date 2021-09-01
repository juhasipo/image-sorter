package widget

import (
	"github.com/AllenDang/giu"
	"image"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
)

type TexturedImage struct {
	Width          float32
	Height         float32
	Ratio          float32
	Texture        *giu.Texture
	oldTexture     *giu.Texture
	Image          *apitype.ImageFile
	oldImage       *apitype.ImageFile
	lastWidth      int
	lastHeight     int
	newImageLoaded bool
	imageCache     api.ImageStore
}

var cache = map[apitype.ImageId]*giu.Texture{}

func NewTexturedImage(image *apitype.ImageFile, width float32, height float32, imageCache api.ImageStore) *TexturedImage {
	return &TexturedImage{
		Width:          width,
		Height:         height,
		Ratio:          width / height,
		Texture:        nil,
		oldTexture:     nil,
		Image:          image,
		oldImage:       nil,
		lastWidth:      -1,
		lastHeight:     -1,
		newImageLoaded: false,
		imageCache:     imageCache,
	}
}

func NewEmptyTexturedImage(imageCache api.ImageStore) *TexturedImage {
	return NewTexturedImage(nil, 0, 0, imageCache)
}

func (s *TexturedImage) ChangeImage(image *apitype.ImageFile, width float32, height float32) {
	s.oldTexture = s.Texture
	s.newImageLoaded = false

	s.oldImage = s.Image
	s.Image = image

	s.Width = width
	s.Height = height
	s.Ratio = width / height

	s.lastWidth = -1
	s.lastHeight = -1

}

func (s *TexturedImage) IsSame(other *TexturedImage) bool {
	if s == nil {
		return false
	}
	if other == nil {
		return false
	}
	return s.Image.Id() == other.Image.Id()
}

func (s *TexturedImage) LoadImageAsTexture(width float32, height float32) *giu.Texture {
	if s.imageCache == nil || s.Image == nil {
		return nil
	}

	if s.newImageLoaded {
		if s.Texture != nil && int(width) == s.lastWidth && int(height) == s.lastHeight {
			return s.Texture
		}
	}

	s.lastWidth = int(width)
	s.lastHeight = int(height)

	scaledImage, _ := s.imageCache.GetScaled(s.Image.Id(), apitype.SizeOf(s.lastWidth, s.lastHeight))
	if scaledImage == nil {
		s.Texture = nil
	} else {
		go func() {
			var err error
			s.Texture, err = giu.NewTextureFromRgba(scaledImage.(*image.RGBA))
			s.newImageLoaded = true
			if err != nil {
				logger.Error.Print(err)
			}
		}()
	}
	return s.Texture
}

func (s *TexturedImage) LoadImageAsTextureThumbnail() *giu.Texture {
	if s.imageCache == nil || s.Image == nil {
		return nil
	}

	if s.newImageLoaded {
		if s.Texture != nil {
			return s.Texture
		}
	}

	scaledImage, _ := s.imageCache.GetThumbnail(s.Image.Id())
	if scaledImage == nil {
		s.Texture = nil
	} else {
		go func() {
			var err error
			s.Texture, err = giu.NewTextureFromRgba(scaledImage.(*image.RGBA))
			s.newImageLoaded = true
			if err != nil {
				logger.Error.Print(err)
			}
			giu.Update()
		}()
	}
	return s.Texture
}
