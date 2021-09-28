package widget

import (
	"github.com/AllenDang/giu"
	"image"
	"sync"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
	"vincit.fi/image-sorter/ui/giu/guiapi"
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
	generalMux sync.RWMutex
}

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
	s.generalMux.Lock()
	defer s.generalMux.Unlock()

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
	s.generalMux.RLock()
	defer s.generalMux.RUnlock()

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
	c := s.getCurrent()
	if c != nil {
		return c.Width
	} else {
		return 0
	}
}

func (s *TexturedImage) Height() float32 {
	c := s.getCurrent()
	if c != nil {
		return c.Height
	} else {
		return 0
	}
}

func (s *TexturedImage) Ratio() float32 {
	c := s.getCurrent()
	if c != nil {
		return c.Ratio
	} else {
		return 1
	}
}

func (s *TexturedImage) Texture() *giu.Texture {
	c := s.getCurrent()
	if c != nil {
		return c.Texture
	} else {
		return nil
	}
}

func (s *TexturedImage) Image() *apitype.ImageFile {
	c := s.getCurrent()
	if c != nil {
		return c.Image
	} else {
		return nil
	}
}

func (s *TexturedImage) getCurrent() *textureEntry {
	return s.current
}

func (s *TexturedImage) LoadImageAsTexture(width float32, height float32, zoomFactor float32, zoomMode guiapi.ZoomMode) *giu.Texture {
	s.generalMux.Lock()
	current := s.getCurrent()
	loading := s.loading
	lastWidth, lastHeight := s.lastWidth, s.lastHeight
	s.generalMux.Unlock()

	if s.imageCache == nil || current == nil {
		return nil
	}

	// First check what size is actually required
	var requiredW float32
	var requiredH float32
	if zoomMode == guiapi.ZoomFit {
		// If zoom to fit, then just load what was asked
		requiredW = width
		requiredH = height
	} else if zoomFactor < 1.0 {
		// If zoomed out, load using the image size and zoom factor
		requiredW = current.Width * zoomFactor
		requiredH = current.Height * zoomFactor
	} else {
		// If zoomed in, just load the max size image
		requiredW = current.Width
		requiredH = current.Height
	}

	// Only load new image if the image grows in size
	// No need to optimize image usage in this case
	if current.Texture != nil && int(requiredW) <= lastWidth && int(requiredH) <= lastHeight {
		return current.Texture
	}

	s.generalMux.Lock()
	s.lastWidth = int(requiredW)
	s.lastHeight = int(requiredH)
	lastWidth, lastHeight = s.lastWidth, s.lastHeight
	s.generalMux.Unlock()

	if logger.IsLogLevel(logger.TRACE) {
		logger.Trace.Printf("Load imageId=%d with new size (%d x %d)", current.Image.Id(), lastWidth, lastHeight)
	}

	if loading != nil {
		toLoadId, toLoadWidth, toLoadHeight := loading.Image.Id(), lastWidth, lastHeight
		go func() {
			if toLoadId == apitype.NoImage {
				current.Texture = nil
			} else {
				scaledImage, _ := s.imageCache.GetScaled(toLoadId, apitype.SizeOf(toLoadWidth, toLoadHeight))
				if scaledImage == nil || (toLoadWidth <= 0 || toLoadHeight <= 0) {
					s.generalMux.Lock()
					current.Texture = nil
					s.generalMux.Unlock()
				} else {
					loadedTexture, err := giu.NewTextureFromRgba(scaledImage.(*image.RGBA))
					s.generalMux.Lock()
					if s.loading != nil {
						s.loading.Texture = loadedTexture
						s.current = s.loading
						s.loading = nil
						if err != nil {
							logger.Error.Print(err)
						}
					}
					s.generalMux.Unlock()
				}
			}
			giu.Update()
		}()
	}
	return s.current.Texture
}

func (s *TexturedImage) ClearTexture() {
	s.generalMux.Lock()
	defer s.generalMux.Unlock()

	s.current.Texture = nil
}

func (s *TexturedImage) LoadImageAsTextureThumbnail() *giu.Texture {
	s.generalMux.Lock()
	current := s.getCurrent()
	loading := s.loading
	s.generalMux.Unlock()

	if s.imageCache == nil || current == nil || current.Image == nil {
		return nil
	}

	if loading == nil && current.Texture != nil {
		return current.Texture
	}

	if loading != nil {
		toLoadId := loading.Image.Id()
		go func() {
			scaledImage, _ := s.imageCache.GetThumbnail(toLoadId)
			if scaledImage == nil {
				s.generalMux.Lock()
				current.Texture = nil
				s.generalMux.Unlock()
			} else {
				texture, err := giu.NewTextureFromRgba(scaledImage.(*image.RGBA))
				s.generalMux.Lock()
				s.loading.Texture = texture
				s.current = s.loading
				s.loading = nil
				if err != nil {
					logger.Error.Print(err)
				}
				s.generalMux.Unlock()
			}
			giu.Update()
		}()
	}
	return s.current.Texture
}
