package imageloader

import (
	"github.com/gotk3/gotk3/gdk"
	"image"
	"log"
	"runtime"
	"sync"
	"time"
	"vincit.fi/image-sorter/common"
)

type CacheContainer struct {
	img *image.NRGBA
}

type ImageCache struct {
	imageCache map[*common.Handle]*Instance
	mux sync.Mutex
}

func (s *ImageCache) Initialize(handles []*common.Handle) {
	s.imageCache = map[*common.Handle]*Instance{}
	for _, handle := range handles {
		s.imageCache[handle] = NewInstance(handle)
	}
	runtime.GC()
}

func ImageCacheNew() *ImageCache{
	return &ImageCache{
		imageCache: map[*common.Handle]*Instance{},
		mux: sync.Mutex{},
	}
}

func (s *ImageCache) GetScaled(handle *common.Handle, size common.Size) image.Image {
	return s.getImage(handle).GetScaled(size)
}
func (s *ImageCache) GetThumbnail(handle *common.Handle) image.Image {
	return s.getImage(handle).GetThumbnail()
}

func (s *ImageCache) GetScaledAsPixbuf(handle *common.Handle, size common.Size) *gdk.Pixbuf {
	s.mux.Lock(); defer s.mux.Unlock()
	startTime := time.Now()
	scaled := s.GetScaled(handle, size)
	cachedImage := scaled
	endTime := time.Now()
	log.Printf("Scaled pixbuf conversion took %s", endTime.Sub(startTime).String())
	return asPixbuf(cachedImage)
}

func (s *ImageCache) GetThumbnailAsPixbuf(handle *common.Handle) *gdk.Pixbuf {
	s.mux.Lock(); defer s.mux.Unlock()
	startTime := time.Now()
	cachedImage := s.GetThumbnail(handle)
	pixbuf := asPixbuf(cachedImage)
	endTime := time.Now()
	log.Printf("Thumbnail pixbuf conversion took %s", endTime.Sub(startTime).String())
	return pixbuf
}

func asPixbuf(cachedImage image.Image) *gdk.Pixbuf {
	if img, ok := cachedImage.(*image.NRGBA); ok {

		size := img.Bounds()
		const bitsPerSample = 8
		const hasAlpha = true
		pb, err := PixbufNewFromData(
			img.Pix,
			gdk.COLORSPACE_RGB, hasAlpha,
			bitsPerSample,
			size.Dx(), size.Dy(),
			img.Stride)
		if err != nil {
			return nil
		}
		return pb
	}
	return nil
}

func (s *ImageCache) getImage(handle *common.Handle) *Instance {
	if handle.IsValid() {
		if existingInstance, ok := s.imageCache[handle]; !ok {
			instance := NewInstance(handle)
			s.imageCache[handle] = instance
			return instance
		} else {
			return existingInstance
		}
	} else {
		return &EMPTY_INSTANCE
	}
}

func (s *ImageCache) Purge(handle *common.Handle) {
	// TODO
}
