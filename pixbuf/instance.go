package pixbuf

import (
	"github.com/gotk3/gotk3/gdk"
	"log"
	"sync"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/imagetools"
)

const (
	PLACEHOLDER_SIZE = 4
	THUMBNAIL_SIZE = 100
)

type Instance struct {
	handle *common.Handle
	full *gdk.Pixbuf
	thumbnail *gdk.Pixbuf
	scaled *gdk.Pixbuf
	exifData *imagetools.ExifData
}

func NewInstance(handle *common.Handle) *Instance {
	exifData, _ := imagetools.LoadExifData(handle)
	instance := &Instance{
		handle:      handle,
		exifData: exifData,
	}
	instance.thumbnail = instance.GetThumbnail()

	return instance
}


func (s *Instance) loadFull() (*gdk.Pixbuf, error) {
	return loadFullWithExifCorrection(s.handle, s.exifData)
}

var mux = sync.Mutex{}
func loadFullWithExifCorrection(handle *common.Handle, exifData *imagetools.ExifData) (*gdk.Pixbuf, error) {
	mux.Lock(); defer mux.Unlock()
	pixbuf, err := gdk.PixbufNewFromFile(handle.GetPath())

	if err != nil {
		log.Print(err)
		return nil, err
	}

	if exifData != nil {
		pixbuf, err = pixbuf.RotateSimple(exifData.GetRotation())
		if err != nil {
			log.Print(err)
			return nil, err
		}

		if exifData.IsFlipped() {
			return pixbuf.Flip(true)
		} else {
			return pixbuf, nil
		}
	} else {
		return pixbuf, nil
	}
}

func (s *Instance) IsValid() bool {
	return s.handle != nil
}

var (
	EMPTY_INSTANCE = Instance {}
)

func (s* Instance) GetScaled(size Size) *gdk.Pixbuf {
	if !s.IsValid() {
		return nil
	}

	full := s.LoadFullFromCache()

	newWidth, newHeight := imagetools.ScaleToFit(full.GetWidth(), full.GetHeight(), size.width, size.height)

	if s.scaled == nil {
		//log.Print(" * Loading new scaled ", s.handle, " (", newWidth, " x ", newHeight, ")...")
		s.scaled, _ = full.ScaleSimple(newWidth, newHeight, gdk.INTERP_TILES)
	} else {
		if newWidth != s.scaled.GetWidth() && newHeight != s.scaled.GetHeight() {
			/*
				log.Print(" * Loading re-scaled ", s.handle,
					" (", s.scaled.GetWidth(), " x ", s.scaled.GetHeight(), ") -> ",
					" (", newWidth, " x ", newHeight, ")...")
			*/
			s.scaled, _ = full.ScaleSimple(newWidth, newHeight, gdk.INTERP_TILES)
		} else {
			//log.Print(" * Use cached")
		}
	}

	return s.scaled
}

func (s* Instance) GetThumbnail() *gdk.Pixbuf {
	if s.handle == nil {
		return nil
	}

	if s.thumbnail == nil {
		full := s.LoadFullFromCache()

		newWidth, newHeight := imagetools.ScaleToFit(full.GetWidth(), full.GetHeight(), THUMBNAIL_SIZE, THUMBNAIL_SIZE)

		s.thumbnail, _ = full.ScaleSimple(newWidth, newHeight, gdk.INTERP_TILES)
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

func (s *Instance) LoadFullFromCache() *gdk.Pixbuf {
	if s.full == nil {
		s.full, _ = s.loadFull()
		return s.full
	} else {
		return s.full
	}
}

func GetByteLength(pixbuf *gdk.Pixbuf) int {
	if pixbuf != nil {
		return pixbuf.GetByteLength()
	} else {
		return 0
	}
}
