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
}

func (s *ExifData) GetRotation() gdk.PixbufRotation {
	return s.rotation
}

func (s *ExifData) IsFlipped() bool {
	return s.flipped
}

func LoadExifData(handle *Handle) (*ExifData, error) {
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
