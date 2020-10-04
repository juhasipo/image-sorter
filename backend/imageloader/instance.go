package imageloader

import (
	"errors"
	"github.com/disintegration/imaging"
	"image"
	"os"
	"sync"
	"time"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
)

const (
	thumbnailWidth  = 100
	thumbnailHeight = thumbnailWidth
)

var (
	emptyInstance = Instance{}
	thumbnailSize = apitype.SizeOf(thumbnailWidth, thumbnailHeight)
)

type Instance struct {
	handle      *apitype.Handle
	full        image.Image
	thumbnail   image.Image
	scaled      image.Image
	exifData    *apitype.ExifData
	imageLoader api.ImageLoader
	mux         sync.Mutex
}

func NewInstance(handle *apitype.Handle, imageLoader api.ImageLoader) *Instance {
	var instance *Instance
	exifData, err := apitype.LoadExifData(handle)
	if err != nil {
		logger.Warn.Printf("Exif data not properly loaded for '%s'", handle.GetId())
	}

	instance = &Instance{
		handle:      handle,
		exifData:    exifData,
		imageLoader: imageLoader,
	}

	instance.thumbnail, _ = instance.GetThumbnail()
	return instance
}

func (s *Instance) IsValid() bool {
	return s.handle != nil
}

func (s *Instance) GetFull() (image.Image, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.full == nil {
		startTime := time.Now()

		if full, err := s.loadFull(nil); err != nil {
			logger.Error.Println("Could not load full image: " + s.handle.GetPath())
			return nil, err
		} else {
			s.full = full
			endTime := time.Now()
			logger.Trace.Printf("'%s': Full loaded in %s", s.handle.GetPath(), endTime.Sub(startTime).String())
			return s.full, nil
		}
	} else {
		logger.Trace.Print("Use cached full image")
		return s.full, nil
	}
}

func (s *Instance) GetScaled(size apitype.Size) (image.Image, error) {
	if !s.IsValid() {
		return nil, errors.New("invalid handle")
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
		s.scaled = imaging.Resize(full, newSize.GetWidth(), newSize.GetHeight(), imaging.Linear)
	} else {
		size := s.scaled.Bounds()
		if newSize.GetWidth() != size.Dx() && newSize.GetHeight() != size.Dy() {
			s.scaled = imaging.Resize(full, newSize.GetWidth(), newSize.GetHeight(), imaging.Linear)
		} else {
			logger.Trace.Print("Use cached scaled image")
			// Use cached
		}
	}
	endTime := time.Now()
	logger.Trace.Printf("'%s': Scaled loaded in %s", s.handle.GetPath(), endTime.Sub(startTime).String())

	return s.scaled, err
}

func (s *Instance) GetThumbnail() (image.Image, error) {
	if s.handle == nil || !s.handle.IsValid() {
		return nil, errors.New("invalid handle")
	}
	var err error
	startTime := time.Now()
	if s.thumbnail == nil {

		if full, err := s.loadThumbnailFromCache(); err != nil {
			return nil, err
		} else {
			fullSize := full.Bounds()
			newSize := apitype.RectangleOfScaledToFit(fullSize, thumbnailSize)

			s.thumbnail = imaging.Resize(full, newSize.GetWidth(), newSize.GetHeight(), imaging.Linear)
		}
	} else {
		logger.Trace.Print("Use cached thumbnail")
	}
	endTime := time.Now()
	logger.Trace.Printf("'%s': Thumbnail loaded in %s", s.handle.GetPath(), endTime.Sub(startTime).String())
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
		loadedImage, err = s.imageLoader.LoadImageScaled(s.handle, *size)
	} else {
		loadedImage, err = s.imageLoader.LoadImage(s.handle)
	}

	if err != nil {
		logger.Error.Print(err)
		return nil, err
	}

	if fileStat, err := os.Stat(s.handle.GetPath()); err == nil {
		s.handle.SetByteSize(fileStat.Size())
	} else {
		logger.Error.Println("Could not load statistic for " + s.handle.GetPath())
	}

	return apitype.ExifRotateImage(loadedImage, s.exifData)
}

func (s *Instance) loadThumbnailFromCache() (image.Image, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.thumbnail == nil {
		startTime := time.Now()

		if thumbnail, err := s.loadFull(&thumbnailSize); err != nil {
			logger.Error.Println("Could not load thumbnail: "+s.handle.GetPath(), err)
			return nil, err
		} else {
			s.thumbnail = thumbnail
			endTime := time.Now()
			logger.Trace.Printf("'%s': Thumbnail loaded in %s", s.handle.GetPath(), endTime.Sub(startTime).String())
			return s.thumbnail, nil
		}
	} else {
		logger.Trace.Print("Use cached thumbnail image")
		return s.thumbnail, nil
	}
}
