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

func NewTexturedImage(image *apitype.ImageFile, imageCache api.ImageStore) *TexturedImage {
	width := float32(0)
	height := float32(0)
	if image != nil {
		width = float32(image.Width())
		height = float32(image.Height())
	}

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
	return NewTexturedImage(nil, imageCache)
}

func (s *TexturedImage) ChangeImage(image *apitype.ImageFile) {
	s.oldTexture = s.Texture
	s.newImageLoaded = false

	s.oldImage = s.Image
	s.Image = image

	width := float32(s.Image.Width())
	height := float32(s.Image.Height())
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

func (s *TexturedImage) LoadImageAsTexture(width float32, height float32, zoomFactor float32) *giu.Texture {
	if s.imageCache == nil || s.Image == nil {
		return nil
	}

	// First check what size is actually required
	var requiredW float32
	var requiredH float32
	if zoomFactor < 0 {
		// If zoom to fit, then just load what was asked
		requiredW = width
		requiredH = height
	} else if zoomFactor < 1.0 {
		// If zoomed out, load using the image size and zoom factor
		requiredW = s.Width * zoomFactor
		requiredH = s.Height * zoomFactor
	} else {
		// If zoomed in, just load the max size image
		requiredW = s.Width
		requiredH = s.Height
	}

	if s.newImageLoaded {
		// Only load new image if the image grows in size
		// No need to optimize image usage in this case
		if s.Texture != nil && int(requiredW) <= s.lastWidth && int(requiredH) <= s.lastHeight {
			return s.Texture
		}
	}

	s.lastWidth = int(requiredW)
	s.lastHeight = int(requiredH)

	if logger.IsLogLevel(logger.TRACE) {
		logger.Trace.Printf("Load imageId=%d with new size (%d x %d)", s.Image.Id(), s.lastWidth, s.lastHeight)
	}

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
