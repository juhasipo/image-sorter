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
	imageCache  map[int64]*Instance
	mux         sync.Mutex
	imageLoader api.ImageLoader

	api.ImageStore
}

func (s *DefaultImageStore) Initialize(handles []*apitype.Handle) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.imageCache = map[int64]*Instance{}
	for _, handle := range handles {
		s.imageCache[handle.GetId()] = NewInstance(handle, s.imageLoader)
	}
	runtime.GC()
}

func NewImageCache(imageLoader api.ImageLoader) api.ImageStore {
	return &DefaultImageStore{
		imageCache:  map[int64]*Instance{},
		mux:         sync.Mutex{},
		imageLoader: imageLoader,
	}
}

func (s *DefaultImageStore) GetFull(handle *apitype.Handle) (image.Image, error) {
	return s.getImage(handle).GetFull()
}
func (s *DefaultImageStore) GetScaled(handle *apitype.Handle, size apitype.Size) (image.Image, error) {
	return s.getImage(handle).GetScaled(size)
}
func (s *DefaultImageStore) GetThumbnail(handle *apitype.Handle) (image.Image, error) {
	return s.getImage(handle).GetThumbnail()
}
func (s *DefaultImageStore) GetExifData(handle *apitype.Handle) *apitype.ExifData {
	return s.getImage(handle).exifData
}

func (s *DefaultImageStore) getImage(handle *apitype.Handle) *Instance {
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
