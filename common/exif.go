package common

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/rwcarlsen/goexif/exif"
	"log"
	"os"
)

type ExifData struct {
	rotation gdk.PixbufRotation
	flipped  bool
	raw      *exif.Exif
}

func (s *ExifData) GetRotation() gdk.PixbufRotation {
	return s.rotation
}

func (s *ExifData) IsFlipped() bool {
	return s.flipped
}

func (s *ExifData) GetRaw() []byte {
	return s.raw.Raw
}

func (s *ExifData) GetRawLength() uint16 {
	return uint16(len(s.raw.Raw))
}

func LoadExifData(handle *Handle) (*ExifData, error) {
	fileForExif, err := os.Open(handle.GetPath())
	if fileForExif != nil && err == nil {
		defer fileForExif.Close()

		orientation := 0
		if decodedExif, err := exif.Decode(fileForExif); err != nil {
			log.Print("Could not decode Exif data", err)
			return nil, err
		} else if orientationTag, err := decodedExif.Get(exif.Orientation); err != nil {
			log.Print("Could not resolve orientation flag", err)
			return nil, err
		} else if orientation, err = orientationTag.Int(0); err != nil {
			log.Print("Could not resolve orientation value", err)
			return nil, err
		} else {
			angle, flip := ExifOrientationToAngleAndFlip(orientation)
			return &ExifData{
				rotation: angle,
				flipped:  flip,
				raw:      decodedExif,
			}, nil
		}
	} else {
		return &ExifData{0, false, nil}, err
	}
}

const (
	NO_ROTATE  = 0
	ROTATE_180 = 180
	LEFT_90    = 90
	RIGHT_90   = 270

	NO_HORIZONTAL_FLIP = false
	HORIZONTAL_FLIP    = true
)

func ExifOrientationToAngleAndFlip(orientation int) (gdk.PixbufRotation, bool) {
	switch orientation {
	case 1:
		return NO_ROTATE, NO_HORIZONTAL_FLIP
	case 2:
		return NO_ROTATE, HORIZONTAL_FLIP
	case 3:
		return ROTATE_180, NO_HORIZONTAL_FLIP
	case 4:
		return ROTATE_180, HORIZONTAL_FLIP
	case 5:
		return RIGHT_90, HORIZONTAL_FLIP
	case 6:
		return RIGHT_90, NO_HORIZONTAL_FLIP
	case 7:
		return LEFT_90, HORIZONTAL_FLIP
	case 8:
		return LEFT_90, NO_HORIZONTAL_FLIP
	default:
		return NO_ROTATE, NO_HORIZONTAL_FLIP
	}
}
