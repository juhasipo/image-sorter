package imageloader

import (
	"github.com/disintegration/imaging"
	"image"
	"image/color"
	"log"
	"os"
	"sync"
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

func (s *Instance) loadFull() (image.Image, error) {
	return loadImageWithExifCorrection(s.handle, s.exifData)
}

var mux = sync.Mutex{}
func loadImageWithExifCorrection(handle *common.Handle, exifData *common.ExifData) (image.Image, error) {
	mux.Lock(); defer mux.Unlock()

	loadedImage, err := goimage.LoadImage(handle)

	if err != nil {
		log.Print(err)
		return nil, err
	}

	size := loadedImage.Bounds()
	handle.SetSize(size.Dx(), size.Dy())
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

	return s.scaled
}

func (s* Instance) GetThumbnail() image.Image {
	if s.handle == nil {
		return nil
	}

	if s.thumbnail == nil {
		full := s.LoadFullFromCache()

		fullSize := full.Bounds()
		newWidth, newHeight := common.ScaleToFit(fullSize.Dx(), fullSize.Dy(), THUMBNAIL_SIZE, THUMBNAIL_SIZE)

		s.thumbnail = imaging.Resize(full, newWidth, newHeight, imaging.Linear)
	} else {
		log.Print("Use cached thumbnail")
	}
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
	if s.full == nil {
		s.full, _ = s.loadFull()
		return s.full
	} else {
		log.Print("Use cached full image")
		return s.full
	}
}

func GetByteLength(pixbuf image.Image) int {
	if pixbuf != nil {
		return 0
	} else {
		return 0
	}
}
