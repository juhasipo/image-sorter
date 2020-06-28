package ui


import "C"
import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/rwcarlsen/goexif/exif"
	"log"
	"os"
	"sync"
	"vincit.fi/image-sorter/common"
)

type PixbufCache struct {
	imageCache map[common.Handle]*Instance
	mux sync.Mutex
}

func NewPixbugCache() *PixbufCache {
	return &PixbufCache{
		imageCache: map[common.Handle]*Instance{},
	}
}

func (s *PixbufCache) GetInstance(handle *common.Handle) *Instance {
	if handle.IsValid() {
		s.mux.Lock()
		defer s.mux.Unlock()
		var instance *Instance
		if val, ok := s.imageCache[*handle]; !ok {
			instance = &Instance{
				handle: handle,
				loader: s,
			}
			s.imageCache[*handle] = instance
		} else {
			instance = val
		}

		return instance
	} else {
		return &EMPTY_INSTANCE
	}
}

func (s *PixbufCache) GetScaled(handle *common.Handle, size Size) *gdk.Pixbuf {
	return s.GetInstance(handle).GetScaled(size)
}
func (s *PixbufCache) GetThumbnail(handle *common.Handle) *gdk.Pixbuf {
	return s.GetInstance(handle).GetThumbnail()
}

func (s*PixbufCache) loadFromHandle(handle *common.Handle) (*gdk.Pixbuf, error) {
	angle, flip, err := s.LoadExifData(handle)

	pixbuf, err := gdk.PixbufNewFromFile(handle.GetPath())

	pixbuf, err = pixbuf.RotateSimple(angle)
	if flip {
		pixbuf, err = pixbuf.Flip(true)
	}
	return pixbuf, err
}

func (s *PixbufCache) LoadExifData(handle *common.Handle) (gdk.PixbufRotation, bool, error) {
	fileForExif, err := os.Open(handle.GetPath())
	if fileForExif != nil && err == nil {
		defer fileForExif.Close()
		decodedExif, err := exif.Decode(fileForExif)
		if err != nil {
			log.Fatal(err)
		}
		orientationTag, _ := decodedExif.Get(exif.Orientation)
		orientation, _ := orientationTag.Int(0)
		angle, flip := ExifOrientationToAngleAndFlip(orientation)
		return angle, flip, err
	} else {
		return 0, false, err
	}
}

const (
	NO_ROTATE = 0
	ROTATE_180 = 180
	LEFT_90 = 90
	RIGHT_90 = 270

	NO_HORIZONTAL_FLIP = false
	HORIZONTAL_FLIP = true
)
func ExifOrientationToAngleAndFlip(orientation int) (gdk.PixbufRotation, bool) {
	switch orientation {
		case 1: return NO_ROTATE, NO_HORIZONTAL_FLIP
		case 2: return NO_ROTATE, HORIZONTAL_FLIP
		case 3: return ROTATE_180, NO_HORIZONTAL_FLIP
		case 4: return ROTATE_180, HORIZONTAL_FLIP
		case 5: return RIGHT_90, HORIZONTAL_FLIP
		case 6: return RIGHT_90, NO_HORIZONTAL_FLIP
		case 7: return LEFT_90, HORIZONTAL_FLIP
		case 8: return LEFT_90, NO_HORIZONTAL_FLIP
		default: return NO_ROTATE, NO_HORIZONTAL_FLIP
	}
}
