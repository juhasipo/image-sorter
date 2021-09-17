package widget

import (
	"github.com/AllenDang/giu"
	"image"
	"sync"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
)

type textureEntry struct {
	Texture *giu.Texture
	Image   *apitype.ImageFile
	Width   float32
	Height  float32
	Ratio   float32
}

type TexturedImage struct {
	current    *textureEntry
	loading    *textureEntry
	lastWidth  int
	lastHeight int
	imageCache api.ImageStore
	mux        sync.Mutex
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
		current: &textureEntry{
			Texture: nil,
			Image:   image,
			Width:   width,
			Height:  height,
			Ratio:   width / height,
		},
		loading: &textureEntry{
			Texture: nil,
			Image:   image,
			Width:   width,
			Height:  height,
			Ratio:   width / height,
		},
		lastWidth:  -1,
		lastHeight: -1,
		imageCache: imageCache,
	}
}

func NewEmptyTexturedImage(imageCache api.ImageStore) *TexturedImage {
	return NewTexturedImage(nil, imageCache)
}

func (s *TexturedImage) ChangeImage(image *apitype.ImageFile) {
	s.mux.Lock()
	defer s.mux.Unlock()

	width := float32(image.Width())
	height := float32(image.Height())

	s.loading = &textureEntry{
		Texture: nil,
		Image:   image,
		Width:   width,
		Height:  height,
		Ratio:   width / height,
	}

	if s.current.Texture == nil {
		s.current = s.loading
	}

	s.lastWidth = -1
	s.lastHeight = -1
	giu.Update()
}

func (s *TexturedImage) IsSame(other *TexturedImage) bool {
	if s == nil || s.current == nil {
		return false
	}
	if other == nil || other.current == nil {
		return false
	}
	return s.current.Image.Id() == other.current.Image.Id()
}

func (s *TexturedImage) NewImageLoaded() bool {
	return s.loading == nil
}

func (s *TexturedImage) Width() float32 {
	return s.current.Width
}

func (s *TexturedImage) Height() float32 {
	return s.current.Height
}

func (s *TexturedImage) Ratio() float32 {
	return s.current.Ratio
}

func (s *TexturedImage) Texture() *giu.Texture {
	return s.current.Texture
}

func (s *TexturedImage) Image() *apitype.ImageFile {
	return s.current.Image
}

func (s *TexturedImage) LoadImageAsTexture(width float32, height float32, zoomFactor float32) *giu.Texture {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.imageCache == nil || s.current == nil {
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
		requiredW = s.current.Width * zoomFactor
		requiredH = s.current.Height * zoomFactor
	} else {
		// If zoomed in, just load the max size image
		requiredW = s.current.Width
		requiredH = s.current.Height
	}

	if s.current != nil {
		// Only load new image if the image grows in size
		// No need to optimize image usage in this case
		if s.current.Texture != nil && int(requiredW) <= s.lastWidth && int(requiredH) <= s.lastHeight {
			return s.current.Texture
		}
	}

	s.lastWidth = int(requiredW)
	s.lastHeight = int(requiredH)

	if logger.IsLogLevel(logger.TRACE) {
		logger.Trace.Printf("Load imageId=%d with new size (%d x %d)", s.current.Image.Id(), s.lastWidth, s.lastHeight)
	}

	if s.loading != nil {
		scaledImage, _ := s.imageCache.GetScaled(s.loading.Image.Id(), apitype.SizeOf(s.lastWidth, s.lastHeight))
		if scaledImage == nil {
			s.current.Texture = nil
		} else {
			go func() {
				s.mux.Lock()
				defer s.mux.Unlock()
				loadedTexture, err := giu.NewTextureFromRgba(scaledImage.(*image.RGBA))

				if s.loading != nil {
					s.loading.Texture = loadedTexture
					s.current = s.loading
					s.loading = nil
					if err != nil {
						logger.Error.Print(err)
					}
				}
				giu.Update()
			}()
		}
	}
	return s.current.Texture
}

func (s *TexturedImage) LoadImageAsTextureThumbnail() *giu.Texture {
	if s.imageCache == nil || s.current == nil || s.current.Image == nil {
		return nil
	}

	if s.loading == nil {
		if s.current.Texture != nil {
			return s.current.Texture
		}
	}

	if s.loading != nil {
		scaledImage, _ := s.imageCache.GetThumbnail(s.loading.Image.Id())
		if scaledImage == nil {
			s.current.Texture = nil
		} else {
			go func() {
				var err error
				s.loading.Texture, err = giu.NewTextureFromRgba(scaledImage.(*image.RGBA))
				s.current = s.loading
				s.loading = nil
				if err != nil {
					logger.Error.Print(err)
				}
				giu.Update()
			}()
		}
	}
	return s.current.Texture
}
