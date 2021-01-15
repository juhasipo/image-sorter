package imageloader

import (
	"image"
	"runtime"
	"sync"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
)

type CacheContainer struct {
	img *image.NRGBA
}

type DefaultImageStore struct {
	imageCache  map[apitype.HandleId]*Instance
	mux         sync.Mutex
	imageLoader api.ImageLoader

	api.ImageStore
}

func (s *DefaultImageStore) Initialize(handles []*apitype.Handle) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.imageCache = map[apitype.HandleId]*Instance{}
	for _, handle := range handles {
		s.imageCache[handle.GetId()] = NewInstance(handle.GetId(), s.imageLoader)
	}
	runtime.GC()
}

func NewImageCache(imageLoader api.ImageLoader) api.ImageStore {
	return &DefaultImageStore{
		imageCache:  map[apitype.HandleId]*Instance{},
		mux:         sync.Mutex{},
		imageLoader: imageLoader,
	}
}

func (s *DefaultImageStore) GetFull(handleId apitype.HandleId) (image.Image, error) {
	return s.getImage(handleId).GetFull()
}
func (s *DefaultImageStore) GetScaled(handleId apitype.HandleId, size apitype.Size) (image.Image, error) {
	return s.getImage(handleId).GetScaled(size)
}
func (s *DefaultImageStore) GetThumbnail(handleId apitype.HandleId) (image.Image, error) {
	return s.getImage(handleId).GetThumbnail()
}
func (s *DefaultImageStore) GetExifData(handleId apitype.HandleId) *apitype.ExifData {
	return nil
}

func (s *DefaultImageStore) getImage(handleId apitype.HandleId) *Instance {
	s.mux.Lock()
	defer s.mux.Unlock()
	if handleId != apitype.NoHandle {
		if existingInstance, ok := s.imageCache[handleId]; !ok {
			instance := NewInstance(handleId, s.imageLoader)
			s.imageCache[handleId] = instance
			return instance
		} else {
			return existingInstance
		}
	} else {
		return &emptyInstance
	}
}

func (s *DefaultImageStore) Purge() {
	for _, instance := range s.imageCache {
		instance.Purge()
	}
}

func (s *DefaultImageStore) GetByteSize() (byteSize uint64) {
	for _, instance := range s.imageCache {
		byteSize += uint64(instance.GetByteLength())
	}
	return
}

func (s *DefaultImageStore) GetSizeInMB() (mbSize float64) {
	return float64(s.GetByteSize()) / (1024 * 1024)
}
