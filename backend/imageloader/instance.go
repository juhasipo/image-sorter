package imageloader

import (
	"errors"
	"github.com/disintegration/imaging"
	"image"
	"sync"
	"time"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/imagereader"
	"vincit.fi/image-sorter/common/logger"
)

const (
	thumbnailWidth  = 100
	thumbnailHeight = thumbnailWidth
)

var (
	emptyInstance = Instance{imageId: apitype.NoImage}
	thumbnailSize = apitype.SizeOf(thumbnailWidth, thumbnailHeight)
)

type Instance struct {
	imageId     apitype.ImageId
	full        image.Image
	thumbnail   image.Image
	scaled      image.Image
	imageLoader api.ImageLoader
	mux         sync.Mutex
}

func NewInstance(imageId apitype.ImageId, imageLoader api.ImageLoader) *Instance {
	var instance *Instance

	instance = &Instance{
		imageId:     imageId,
		imageLoader: imageLoader,
	}

	instance.thumbnail, _ = instance.GetThumbnail()
	return instance
}

func (s *Instance) IsValid() bool {
	return s.imageId != apitype.NoImage
}

func (s *Instance) GetFull() (image.Image, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.full == nil {
		startTime := time.Now()

		if full, err := s.loadFull(nil); err != nil {
			logger.Error.Printf("Could not load full image: %d", s.imageId)
			return nil, err
		} else {
			s.full = full
			endTime := time.Now()
			logger.Trace.Printf("%d: Full loaded in %s", s.imageId, endTime.Sub(startTime).String())
			return s.full, nil
		}
	} else {
		logger.Trace.Print("Use cached full image")
		return s.full, nil
	}
}

func (s *Instance) GetScaled(size apitype.Size) (image.Image, error) {
	if !s.IsValid() {
		return nil, errors.New("invalid image instance")
	}

	startTime := time.Now()
	var full image.Image
	var err error
	if full, err = s.GetFull(); err != nil {
		return nil, err
	}

	fullSize := full.Bounds()
	newSize := apitype.RectangleOfScaledToFit(fullSize, size)

	if s.scaled == nil {
		s.scaled = imagereader.ConvertNrgbaToRgba(imaging.Resize(full, newSize.Width(), newSize.Height(), imaging.Linear))
	} else {
		size := s.scaled.Bounds()
		if newSize.Width() != size.Dx() && newSize.Height() != size.Dy() {
			s.scaled = imagereader.ConvertNrgbaToRgba(imaging.Resize(full, newSize.Width(), newSize.Height(), imaging.Linear))
		} else {
			logger.Trace.Print("Use cached scaled image")
			// Use cached
		}
	}
	endTime := time.Now()
	logger.Trace.Printf("%d: Scaled loaded in %s", s.imageId, endTime.Sub(startTime).String())

	return s.scaled, err
}

func (s *Instance) GetThumbnail() (image.Image, error) {
	if s.imageId == apitype.NoImage {
		return nil, errors.New("invalid image instance")
	}
	var err error
	startTime := time.Now()
	if s.thumbnail == nil {

		if full, err := s.loadThumbnailFromCache(); err != nil {
			return nil, err
		} else if full == nil {
			return nil, nil
		} else {
			fullSize := full.Bounds()
			newSize := apitype.RectangleOfScaledToFit(fullSize, thumbnailSize)

			s.thumbnail = imagereader.ConvertNrgbaToRgba(imaging.Resize(full, newSize.Width(), newSize.Height(), imaging.Linear))
		}
	} else {
		logger.Trace.Print("Use cached thumbnail")
	}
	endTime := time.Now()
	logger.Trace.Printf("%d: Thumbnail loaded in %s", s.imageId, endTime.Sub(startTime).String())
	return s.thumbnail, err
}

func (s *Instance) Purge() {
	s.full = nil
	s.scaled = nil
}

func (s *Instance) GetByteLength() int {
	var byteLength = 0
	byteLength += GetByteLength(s.full)
	byteLength += GetByteLength(s.scaled)
	byteLength += GetByteLength(s.thumbnail)
	return byteLength
}
func GetByteLength(pixbuf image.Image) int {
	if pixbuf != nil {
		// Approximation using the image size
		const bytesPerPixel = 4
		bounds := pixbuf.Bounds()
		return bounds.Dx() * bounds.Dy() * bytesPerPixel
	} else {
		return 0
	}
}

func (s *Instance) loadFull(size *apitype.Size) (image.Image, error) {
	return s.loadImageWithExifCorrection(size)
}

func (s *Instance) loadImageWithExifCorrection(size *apitype.Size) (image.Image, error) {
	if s.imageLoader == nil {
		return nil, errors.New("no valid loader")
	}

	var loadedImage image.Image
	var err error
	if size != nil {
		loadedImage, err = s.imageLoader.LoadImageScaled(s.imageId, *size)
	} else {
		loadedImage, err = s.imageLoader.LoadImage(s.imageId)
	}

	if err != nil {
		logger.Error.Print(err)
		return nil, err
	}

	return loadedImage, nil
}

func (s *Instance) loadThumbnailFromCache() (image.Image, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.thumbnail == nil {
		startTime := time.Now()

		if thumbnail, err := s.loadFull(&thumbnailSize); err != nil {
			logger.Error.Printf("Could not load thumbnail: %d", s.imageId)
			return nil, err
		} else {
			s.thumbnail = thumbnail
			endTime := time.Now()
			logger.Trace.Printf("%d: Thumbnail loaded in %s", s.imageId, endTime.Sub(startTime).String())
			return s.thumbnail, nil
		}
	} else {
		logger.Trace.Print("Use cached thumbnail image")
		return s.thumbnail, nil
	}
}
