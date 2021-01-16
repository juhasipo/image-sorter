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
	imageCache  map[apitype.ImageId]*Instance
	mux         sync.Mutex
	imageLoader api.ImageLoader

	api.ImageStore
}

func (s *DefaultImageStore) Initialize(imageFiles []*apitype.ImageFileWithMetaData) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.imageCache = map[apitype.ImageId]*Instance{}
	for _, imageFile := range imageFiles {
		s.imageCache[imageFile.GetImageId()] = NewInstance(imageFile.GetImageId(), s.imageLoader)
	}
	runtime.GC()
}

func NewImageCache(imageLoader api.ImageLoader) api.ImageStore {
	return &DefaultImageStore{
		imageCache:  map[apitype.ImageId]*Instance{},
		mux:         sync.Mutex{},
		imageLoader: imageLoader,
	}
}

func (s *DefaultImageStore) GetFull(imageId apitype.ImageId) (image.Image, error) {
	return s.getImage(imageId).GetFull()
}
func (s *DefaultImageStore) GetScaled(imageId apitype.ImageId, size apitype.Size) (image.Image, error) {
	return s.getImage(imageId).GetScaled(size)
}
func (s *DefaultImageStore) GetThumbnail(imageId apitype.ImageId) (image.Image, error) {
	return s.getImage(imageId).GetThumbnail()
}
func (s *DefaultImageStore) GetExifData(imageId apitype.ImageId) *apitype.ExifData {
	return nil
}

func (s *DefaultImageStore) getImage(imageId apitype.ImageId) *Instance {
	s.mux.Lock()
	defer s.mux.Unlock()
	if imageId != apitype.NoImage {
		if existingInstance, ok := s.imageCache[imageId]; !ok {
			instance := NewInstance(imageId, s.imageLoader)
			s.imageCache[imageId] = instance
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
