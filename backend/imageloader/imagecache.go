package imageloader

import (
	"image"
	"runtime"
	"sync"
	"time"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
)

type DefaultImageStore struct {
	imageCache  map[apitype.ImageId]*Instance
	mux         sync.Mutex
	imageLoader api.ImageLoader

	api.ImageStore
}

func (s *DefaultImageStore) Initialize(imageFiles []*apitype.ImageFile, reporter api.ProgressReporter) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.imageCache = map[apitype.ImageId]*Instance{}
	go func() {
		numOfImages := len(imageFiles)
		logger.Debug.Printf("Start loading %d image instances in cache...", numOfImages)
		startTime := time.Now()
		reporter.Update("Image cache", 0, numOfImages, false, false)
		for i, imageFile := range imageFiles {
			s.loadImageInstance(imageFile)
			reporter.Update("Image cache", i+1, numOfImages, false, false)
		}
		endTime := time.Now()
		totalTime := endTime.Sub(startTime)
		avg := totalTime / time.Duration(numOfImages)
		logger.Debug.Printf("All %d instances loaded in cache in %s (avg. %s)", numOfImages, totalTime.String(), avg.String())
	}()
	runtime.GC()
}

func (s *DefaultImageStore) loadImageInstance(imageFile *apitype.ImageFile) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if _, ok := s.imageCache[imageFile.Id()]; !ok {
		s.imageCache[imageFile.Id()] = NewInstance(imageFile.Id(), s.imageLoader)
	}
}

func NewImageCache(imageLoader api.ImageLoader) api.ImageStore {
	logger.Debug.Printf("Initialize image cache...")
	imageCache := &DefaultImageStore{
		imageCache:  map[apitype.ImageId]*Instance{},
		mux:         sync.Mutex{},
		imageLoader: imageLoader,
	}
	logger.Debug.Printf("Image cache initialized")
	return imageCache
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
