package pixbuf

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/rwcarlsen/goexif/exif"
	"log"
	"os"
	"vincit.fi/image-sorter/common"
)

const (
	PLACEHOLDER_SIZE = 4
	THUMBNAIL_SIZE = 100
)

type ExifData struct {
	rotation gdk.PixbufRotation
	flipped  bool
}

func (s *ExifData) GetRotation() gdk.PixbufRotation {
	return s.rotation
}

func (s *ExifData) IsFlipped() bool {
	return s.flipped
}

type Instance struct {
	handle *common.Handle
	full *gdk.Pixbuf
	thumbnail *gdk.Pixbuf
	scaled *gdk.Pixbuf
	exifData *ExifData
}

func NewInstance(handle *common.Handle) *Instance {
	exifData, _ := LoadExifData(handle)
	instance := &Instance{
		handle:      handle,
		exifData: exifData,
	}
	instance.thumbnail = instance.GetThumbnail()

	return instance
}

func LoadExifData(handle *common.Handle) (*ExifData, error) {
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
		return &ExifData{
			rotation: angle,
			flipped: flip,
		}, nil
	} else {
		return &ExifData{0, false}, err
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

func (s *Instance) LoadFull() (*gdk.Pixbuf, error) {
	return LoadFullWithExifCorrection(s.handle, s.exifData)
}

func LoadFullWithExifCorrection(handle *common.Handle, exifData *ExifData) (*gdk.Pixbuf, error) {
	pixbuf, err := gdk.PixbufNewFromFile(handle.GetPath())

	pixbuf, err = pixbuf.RotateSimple(exifData.rotation)
	if exifData.flipped {
		pixbuf, err = pixbuf.Flip(true)
	}
	return pixbuf, err
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

	ratio := float32(full.GetWidth()) / float32(full.GetHeight())
	newWidth := int(float32(size.height) * ratio)
	newHeight := size.height

	if newWidth > size.width {
		newWidth = size.width
		newHeight = int(float32(size.width) / ratio)
	}

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

		width, height := THUMBNAIL_SIZE, THUMBNAIL_SIZE
		ratio := float32(full.GetWidth()) / float32(full.GetHeight())
		newWidth := int(float32(height) * ratio)
		newHeight := height

		if newWidth > width {
			newWidth = width
			newHeight = int(float32(width) / ratio)
		}

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
		s.full, _ = s.LoadFull()
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
