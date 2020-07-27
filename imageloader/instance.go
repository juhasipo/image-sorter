package imageloader

import (
	"github.com/disintegration/imaging"
	"image"
	"image/color"
	"os"
	"sync"
	"time"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/imageloader/goimage"
	"vincit.fi/image-sorter/logger"
)

const (
	THUMBNAIL_SIZE = 100
)

type Instance struct {
	handle      *common.Handle
	full        image.Image
	thumbnail   image.Image
	scaled      image.Image
	exifData    *common.ExifData
	imageLoader goimage.ImageLoader
	mux         sync.Mutex
}

func NewInstance(handle *common.Handle, imageLoader goimage.ImageLoader) *Instance {
	var instance *Instance
	if exifData, err := common.LoadExifData(handle); err == nil {
		instance = &Instance{
			handle:      handle,
			exifData:    exifData,
			imageLoader: imageLoader,
		}
	} else {
		instance = &Instance{
			handle:      handle,
			exifData:    nil,
			imageLoader: imageLoader,
		}
	}

	instance.thumbnail = instance.GetThumbnail()
	return instance
}

func (s *Instance) loadFull(size *common.Size) (image.Image, error) {
	return s.loadImageWithExifCorrection(size)
}

var mux = sync.Mutex{}

func (s *Instance) loadImageWithExifCorrection(size *common.Size) (image.Image, error) {
	//mux.Lock(); defer mux.Unlock()

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

	return ExifRotateImage(loadedImage, s.exifData)
}

func ExifRotateImage(loadedImage image.Image, exifData *common.ExifData) (image.Image, error) {
	if exifData != nil {
		loadedImage = imaging.Rotate(loadedImage, float64(exifData.GetRotation()), color.Black)
		if exifData.IsFlipped() {
			return imaging.FlipH(loadedImage), nil
		} else {
			return loadedImage, nil
		}
	} else {
		return loadedImage, nil
	}
}

func (s *Instance) IsValid() bool {
	return s.handle != nil
}

var (
	EMPTY_INSTANCE = Instance{}
)

func (s *Instance) GetScaled(size common.Size) image.Image {
	if !s.IsValid() {
		return nil
	}

	startTime := time.Now()
	full := s.LoadFullFromCache()

	fullSize := full.Bounds()
	newWidth, newHeight := common.ScaleToFit(fullSize.Dx(), fullSize.Dy(), size.GetWidth(), size.GetHeight())

	if s.scaled == nil {
		s.scaled = imaging.Resize(full, newWidth, newHeight, imaging.Linear)
	} else {
		size := s.scaled.Bounds()
		if newWidth != size.Dx() && newHeight != size.Dy() {
			s.scaled = imaging.Resize(full, newWidth, newHeight, imaging.Linear)
		} else {
			logger.Trace.Print("Use cached scaled image")
			// Use cached
		}
	}
	endTime := time.Now()
	logger.Trace.Printf("'%s': Scaled loaded in %s", s.handle.GetPath(), endTime.Sub(startTime).String())

	return s.scaled
}

func (s *Instance) GetThumbnail() image.Image {
	if s.handle == nil {
		return nil
	}

	startTime := time.Now()
	if s.thumbnail == nil {

		full := s.LoadThumbnailFromCache()

		fullSize := full.Bounds()
		newWidth, newHeight := common.ScaleToFit(fullSize.Dx(), fullSize.Dy(), THUMBNAIL_SIZE, THUMBNAIL_SIZE)

		s.thumbnail = imaging.Resize(full, newWidth, newHeight, imaging.Linear)
	} else {
		logger.Trace.Print("Use cached thumbnail")
	}
	endTime := time.Now()
	logger.Trace.Printf("'%s': Thumbnail loaded in %s", s.handle.GetPath(), endTime.Sub(startTime).String())
	return s.thumbnail
}

func (s *Instance) Purge() {
	s.full = nil
	s.scaled = nil
}

func (s *Instance) GetByteLength() int {
	var byteLength = 0
	byteLength += GetByteLength(s.scaled)
	byteLength += GetByteLength(s.thumbnail)
	return byteLength
}

func (s *Instance) LoadFullFromCache() image.Image {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.full == nil {
		startTime := time.Now()

		var err error
		if s.full, err = s.loadFull(nil); err != nil {
			logger.Error.Println("Could not load image: "+s.handle.GetPath(), err)
		}

		endTime := time.Now()
		logger.Trace.Printf("'%s': Full loaded in %s", s.handle.GetPath(), endTime.Sub(startTime).String())
		return s.full
	} else {
		logger.Trace.Print("Use cached full image")
		return s.full
	}
}

func (s *Instance) LoadThumbnailFromCache() image.Image {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.thumbnail == nil {
		startTime := time.Now()

		size := common.SizeOf(THUMBNAIL_SIZE, THUMBNAIL_SIZE)
		var err error
		if s.thumbnail, err = s.loadFull(&size); err != nil {
			logger.Error.Println("Could not load thumbnail: "+s.handle.GetPath(), err)
		}

		endTime := time.Now()
		logger.Trace.Printf("'%s': Thumbnail loaded in %s", s.handle.GetPath(), endTime.Sub(startTime).String())
		return s.thumbnail
	} else {
		logger.Trace.Print("Use cached thumbnail image")
		return s.thumbnail
	}
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
