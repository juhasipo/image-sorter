package imageloader

import (
	"github.com/disintegration/imaging"
	"image"
	"image/color"
	"log"
	"os"
	"sync"
	"time"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/imageloader/goimage"
)

const (
	THUMBNAIL_SIZE = 100
)

type Instance struct {
	handle *common.Handle
	full image.Image
	thumbnail image.Image
	scaled image.Image
	exifData *common.ExifData
	mux sync.Mutex
}

func NewInstance(handle *common.Handle) *Instance {
	exifData, _ := common.LoadExifData(handle)
	instance := &Instance{
		handle:      handle,
		exifData: exifData,
	}
	instance.thumbnail = instance.GetThumbnail()

	return instance
}

func (s *Instance) loadFull(size *common.Size) (image.Image, error) {
	return loadImageWithExifCorrection(s.handle, s.exifData, size)
}

var mux = sync.Mutex{}
func loadImageWithExifCorrection(handle *common.Handle, exifData *common.ExifData, size *common.Size) (image.Image, error) {
	//mux.Lock(); defer mux.Unlock()

	var loadedImage image.Image
	var err error
	if size != nil {
		loadedImage, err = goimage.LoadImageScaled(handle, *size)
	} else {
		loadedImage, err = goimage.LoadImage(handle)
	}

	if err != nil {
		log.Print(err)
		return nil, err
	}

	fileStat, _ := os.Stat(handle.GetPath())
	handle.SetByteSize(fileStat.Size())

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
	EMPTY_INSTANCE = Instance {}
)

func (s* Instance) GetScaled(size common.Size) image.Image{
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
			log.Print("Use cached scaled image")
			// Use cached
		}
	}
	endTime := time.Now()
	log.Printf("'%s': Scaled loaded in %s", s.handle.GetPath(), endTime.Sub(startTime).String())

	return s.scaled
}

func (s* Instance) GetThumbnail() image.Image {
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
		log.Print("Use cached thumbnail")
	}
	endTime := time.Now()
	log.Printf("'%s': Thumbnail loaded in %s", s.handle.GetPath(), endTime.Sub(startTime).String())
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
	s.mux.Lock(); defer s.mux.Unlock()
	if s.full == nil {
		startTime := time.Now()

		s.full, _ = s.loadFull(nil)

		endTime := time.Now()
		log.Printf("'%s': Full loaded in %s", s.handle.GetPath(), endTime.Sub(startTime).String())
		return s.full
	} else {
		log.Print("Use cached full image")
		return s.full
	}
}

func (s *Instance) LoadThumbnailFromCache() image.Image {
	s.mux.Lock(); defer s.mux.Unlock()
	if s.thumbnail == nil {
		startTime := time.Now()

		size := common.SizeOf(100, 100)
		s.thumbnail, _ = s.loadFull(&size)

		endTime := time.Now()
		log.Printf("'%s': Thumbnail loaded in %s", s.handle.GetPath(), endTime.Sub(startTime).String())
		return s.thumbnail
	} else {
		log.Print("Use cached thumbnail image")
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
