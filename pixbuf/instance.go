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

type Instance struct {
	handle *common.Handle
	full *gdk.Pixbuf
	thumbnail *gdk.Pixbuf
	placeholder *gdk.Pixbuf
	scaled *gdk.Pixbuf
}

func NewInstance(handle *common.Handle) *Instance {
	fullPixbuf, _ := LoadFromHandle(handle)
	placeholderPixbuf, _ := fullPixbuf.ScaleSimple(PLACEHOLDER_SIZE, PLACEHOLDER_SIZE, gdk.INTERP_TILES)
	return &Instance{
		handle:      handle,
		full:        fullPixbuf,
		placeholder: placeholderPixbuf,
	}
}

func LoadExifData(handle *common.Handle) (gdk.PixbufRotation, bool, error) {
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

func LoadFromHandle(handle *common.Handle) (*gdk.Pixbuf, error) {
	angle, flip, err := LoadExifData(handle)

	pixbuf, err := gdk.PixbufNewFromFile(handle.GetPath())

	pixbuf, err = pixbuf.RotateSimple(angle)
	if flip {
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
		log.Print("Empty instance")
		return nil
	}

	if s.full == nil {
		log.Print(" * Loading full image...")
		s.full, _ = LoadFromHandle(s.handle)
	}

	ratio := float32(s.full.GetWidth()) / float32(s.full.GetHeight())
	newWidth := int(float32(size.height) * ratio)
	newHeight := size.height

	if newWidth > size.width {
		newWidth = size.width
		newHeight = int(float32(size.width) / ratio)
	}

	if s.scaled == nil {
		log.Print(" * Loading new scaled ", s.handle, " (", newWidth, " x ", newHeight, ")...")
		s.scaled, _ = s.full.ScaleSimple(newWidth, newHeight, gdk.INTERP_TILES)
	} else {
		if newWidth != s.scaled.GetWidth() && newHeight != s.scaled.GetHeight() {
			log.Print(" * Loading re-scaled ", s.handle,
				" (", s.scaled.GetWidth(), " x ", s.scaled.GetHeight(), ") -> ",
				" (", newWidth, " x ", newHeight, ")...")
			s.scaled, _ = s.full.ScaleSimple(newWidth, newHeight, gdk.INTERP_TILES)
		} else {
			log.Print(" * Use cached")
		}
	}

	return s.scaled
}

func (s* Instance) GetThumbnail() *gdk.Pixbuf {
	if s.handle == nil {
		log.Print("Nil handle")
		return nil
	}
	if s.full == nil {
		//log.Print(" * Loading full image...")
		s.full, _ = LoadFromHandle(s.handle)
	}
	if s.thumbnail == nil {
		width, height := THUMBNAIL_SIZE, THUMBNAIL_SIZE
		ratio := float32(s.full.GetWidth()) / float32(s.full.GetHeight())
		newWidth := int(float32(height) * ratio)
		newHeight := height

		if newWidth > width {
			newWidth = width
			newHeight = int(float32(width) / ratio)
		}

		s.thumbnail, _ = s.full.ScaleSimple(newWidth, newHeight, gdk.INTERP_TILES)
	}
	return s.thumbnail
}
