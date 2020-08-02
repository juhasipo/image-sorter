package common

import (
	"bytes"
	"errors"
	"github.com/gotk3/gotk3/gdk"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
	"os"
	"vincit.fi/image-sorter/logger"
)

type ExifData struct {
	orientation uint8
	rotation    gdk.PixbufRotation
	flipped     bool
	raw         *exif.Exif
}

const exifUnchangedOrientation = 1
const exifValueMarker = 0xFF

// Tag (2 bytes), type (2 bytes), count (4 bytes), value (2 bytes): 0xFF is the marker for value
// Intel byte order
var orientationIntelPattern = []byte{0x12, 0x01, 0x03, 0x00, 0x01, 0x00, 0x00, 0x00, exifValueMarker, 0x00}

const orientationIntelOffset = 8 // Offset for the value from the tag

// Motorola byte order
var orientationMotorolaPattern = []byte{0x01, 0x12, 0x00, 0x03, 0x00, 0x00, 0x00, 0x01, 0x00, exifValueMarker}

const orientationMotorolaOffset = 9 // Offset for the value from the tag

func (s *ExifData) ResetExifRotate() {
	orientationByteIndex, err := findOrientationByteIndex(s.raw.Raw, s.orientation)
	if err != nil {
		return
	}
	s.orientation = exifUnchangedOrientation
	s.rotation = 0
	s.flipped = false
	s.raw.Raw[orientationByteIndex] = exifUnchangedOrientation
}

// Finds the index for orientation with the given value
func findOrientationByteIndex(exifData []byte, value uint8) (int, error) {
	buffer := copyAndSetValue(orientationIntelPattern, value)
	if result, err := find(exifData, buffer); err == nil {
		return result + orientationIntelOffset, nil
	} else {
		buffer = copyAndSetValue(orientationMotorolaPattern, value)
		if result, err := find(exifData, buffer); err == nil {
			return result + orientationMotorolaOffset, nil
		}
	}
	return 0, errors.New("not found")
}

func copyAndSetValue(buf []byte, value uint8) []byte {
	byteArray := make([]byte, len(buf))
	copy(byteArray, buf)
	byteArray[bytes.IndexByte(byteArray, exifValueMarker)] = value
	return byteArray
}

func find(exifData []byte, s []byte) (int, error) {
	index := bytes.Index(exifData, s)
	if index < 0 {
		return 0, errors.New("not found")
	} else {
		return index, nil
	}
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

func (s *ExifData) Get(name exif.FieldName) *tiff.Tag {
	if tag, err := s.raw.Get(name); err != nil {
		return nil
	} else {
		return tag
	}
}

func (s *ExifData) Walk(walker exif.Walker) {
	_ = s.raw.Walk(walker)
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
			logger.Error.Print("Could not decode Exif data", err)
			return nil, err
		} else if orientationTag, err := decodedExif.Get(exif.Orientation); err != nil {
			logger.Warn.Print("Could not resolve orientation flag", err)
			return &ExifData{1, 0, false, decodedExif}, err
		} else if orientation, err = orientationTag.Int(0); err != nil {
			logger.Warn.Print("Could not resolve orientation value", err)
			return &ExifData{1, 0, false, decodedExif}, err
		} else {
			angle, flip := ExifOrientationToAngleAndFlip(orientation)
			return &ExifData{
				orientation: uint8(orientation),
				rotation:    angle,
				flipped:     flip,
				raw:         decodedExif,
			}, nil
		}
	} else {
		return &ExifData{1, 0, false, nil}, err
	}
}

const (
	noRotate  = 0
	rotate180 = 180
	left90    = 90
	right90   = 270

	noHorizontalFlip = false
	horizontalFlip   = true
)

func ExifOrientationToAngleAndFlip(orientation int) (gdk.PixbufRotation, bool) {
	switch orientation {
	case 1:
		return noRotate, noHorizontalFlip
	case 2:
		return noRotate, horizontalFlip
	case 3:
		return rotate180, noHorizontalFlip
	case 4:
		return rotate180, horizontalFlip
	case 5:
		return right90, horizontalFlip
	case 6:
		return right90, noHorizontalFlip
	case 7:
		return left90, horizontalFlip
	case 8:
		return left90, noHorizontalFlip
	default:
		return noRotate, noHorizontalFlip
	}
}
