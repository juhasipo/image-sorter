package imageloader

import (
	"image"
	"runtime"
	"sync"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/imageloader/goimage"
)

type CacheContainer struct {
	img *image.NRGBA
}

type ImageCache interface {
	Initialize([]*common.Handle)
	GetFull(*common.Handle) image.Image
	GetScaled(*common.Handle, common.Size) image.Image
	GetThumbnail(*common.Handle) image.Image
	Purge(*common.Handle)
}

type DefaultImageCache struct {
	imageCache  map[string]*Instance
	mux         sync.Mutex
	imageLoader goimage.ImageLoader

	ImageCache
}

func (s *DefaultImageCache) Initialize(handles []*common.Handle) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.imageCache = map[string]*Instance{}
	for _, handle := range handles {
		s.imageCache[handle.GetId()] = NewInstance(handle, s.imageLoader)
	}
	runtime.GC()
}

func ImageCacheNew(imageLoader goimage.ImageLoader) ImageCache {
	return &DefaultImageCache{
		imageCache:  map[string]*Instance{},
		mux:         sync.Mutex{},
		imageLoader: imageLoader,
	}
}

func (s *DefaultImageCache) GetFull(handle *common.Handle) image.Image {
	return s.getImage(handle).LoadFullFromCache()
}
func (s *DefaultImageCache) GetScaled(handle *common.Handle, size common.Size) image.Image {
	return s.getImage(handle).GetScaled(size)
}
func (s *DefaultImageCache) GetThumbnail(handle *common.Handle) image.Image {
	return s.getImage(handle).GetThumbnail()
}

func (s *DefaultImageCache) getImage(handle *common.Handle) *Instance {
	s.mux.Lock()
	defer s.mux.Unlock()
	if handle.IsValid() {
		if existingInstance, ok := s.imageCache[handle.GetId()]; !ok {
			instance := NewInstance(handle, s.imageLoader)
			s.imageCache[handle.GetId()] = instance
			return instance
		} else {
			return existingInstance
		}
	} else {
		return &EMPTY_INSTANCE
	}
}

func (s *DefaultImageCache) Purge(handle *common.Handle) {
	for _, instance := range s.imageCache {
		instance.Purge()
	}
}
