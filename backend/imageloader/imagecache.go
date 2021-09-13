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
	imageCache    map[apitype.ImageId]*Instance
	mux           sync.Mutex
	imageLoader   api.ImageLoader
	stopChannel   chan bool
	outputChannel chan *Instance

	api.ImageStore
}

func (s *DefaultImageStore) Initialize(imageFiles []*apitype.ImageFile, reporter api.ProgressReporter) {
	s.initializeCache()
	numOfImages := len(imageFiles)
	logger.Info.Printf("Start loading %d image instances in cache...", numOfImages)

	threadCount := runtime.NumCPU()
	logger.Info.Printf(" * Using %d threads", threadCount)
	runtime.GOMAXPROCS(threadCount)

	s.stopChannel = make(chan bool)
	inputChannel := make(chan *apitype.ImageFile, numOfImages)
	s.outputChannel = make(chan *Instance)

	// Add images to input queue for goroutines
	for _, imageFile := range imageFiles {
		inputChannel <- imageFile
	}

	// Spin up goroutines which will process the data
	// only same number as CPU cores so that we will only max X hashes are
	// processed at once. Otherwise, the goroutines might start processing
	// all images at once which would use all available RAM
	for i := 0; i < threadCount; i++ {
		go s.loadImageInstance(inputChannel, s.outputChannel, s.stopChannel)
	}

	go func() {
		startTime := time.Now()
		reporter.Update("Image cache", 0, numOfImages, false, false)
		var i = 0
		for instance := range s.outputChannel {
			i++

			s.addImageToCache(instance)
			reporter.Update("Image cache", i, numOfImages, false, false)

			if i == numOfImages {
				if s.stopChannel != nil {
					for i := 0; i < threadCount; i++ {
						s.stopChannel <- true
					}
					close(s.outputChannel)
					close(s.stopChannel)
					s.stopChannel = nil
				}
			}
		}
		close(inputChannel)

		endTime := time.Now()
		totalTime := endTime.Sub(startTime)
		avg := totalTime / time.Duration(numOfImages)
		logger.Info.Printf("All %d instances loaded in cache in %s (avg. %s)", numOfImages, totalTime.String(), avg.String())

		runtime.GC()
	}()
}

func (s *DefaultImageStore) initializeCache() {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.imageCache = map[apitype.ImageId]*Instance{}
}

func (s *DefaultImageStore) addImageToCache(instance *Instance) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.imageCache[instance.imageId] = instance
}

func (s *DefaultImageStore) loadImageInstance(inputChannel chan *apitype.ImageFile, outputChannel chan *Instance, quitChannel chan bool) {
	for {
		select {
		case <-quitChannel:
			return
		case imageFile := <-inputChannel:
			{
				outputChannel <- NewInstance(imageFile.Id(), s.imageLoader)
			}
		}
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
