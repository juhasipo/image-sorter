package imageloader

import (
	"image"
	"runtime"
	"sync"
	"vincit.fi/image-sorter/common"
)

type CacheContainer struct {
	img *image.NRGBA
}

type ImageCache struct {
	imageCache map[string]*Instance
	mux        sync.Mutex
}

func (s *ImageCache) Initialize(handles []*common.Handle) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.imageCache = map[string]*Instance{}
	for _, handle := range handles {
		s.imageCache[handle.GetId()] = NewInstance(handle)
	}
	runtime.GC()
}

func ImageCacheNew() *ImageCache {
	return &ImageCache{
		imageCache: map[string]*Instance{},
		mux:        sync.Mutex{},
	}
}

func (s *ImageCache) GetFull(handle *common.Handle) image.Image {
	return s.getImage(handle).LoadFullFromCache()
}
func (s *ImageCache) GetScaled(handle *common.Handle, size common.Size) image.Image {
	return s.getImage(handle).GetScaled(size)
}
func (s *ImageCache) GetThumbnail(handle *common.Handle) image.Image {
	return s.getImage(handle).GetThumbnail()
}

func (s *ImageCache) getImage(handle *common.Handle) *Instance {
	s.mux.Lock()
	defer s.mux.Unlock()
	if handle.IsValid() {
		if existingInstance, ok := s.imageCache[handle.GetId()]; !ok {
			instance := NewInstance(handle)
			s.imageCache[handle.GetId()] = instance
			return instance
		} else {
			return existingInstance
		}
	} else {
		return &EMPTY_INSTANCE
	}
}

func (s *ImageCache) Purge(handle *common.Handle) {
	for _, instance := range s.imageCache {
		instance.Purge()
	}
}
