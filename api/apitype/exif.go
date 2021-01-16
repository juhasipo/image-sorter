package apitype

import (
	"bytes"
	"errors"
	"github.com/disintegration/imaging"
	"github.com/gotk3/gotk3/gdk"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
	"image"
	"image/color"
	"os"
	"time"
	"vincit.fi/image-sorter/common/logger"
)

type ExifData struct {
	orientation uint8
	rotation    gdk.PixbufRotation
	flipped     bool
	raw         *exif.Exif
	created     time.Time
	width       uint32
	height      uint32
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

func (s *ExifData) GetExifOrientation() uint8 {
	return s.orientation
}

func (s *ExifData) GetRotation() gdk.PixbufRotation {
	return s.rotation
}

func (s *ExifData) IsFlipped() bool {
	return s.flipped
}

func (s *ExifData) GetCreatedTime() time.Time {
	return s.created
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

func (s *ExifData) GetWidth() uint32 {
	return s.width
}

func (s *ExifData) GetHeight() uint32 {
	return s.height
}

func GetInt(decodedExif *exif.Exif, tagName exif.FieldName) (int, error) {
	if tag, err := decodedExif.Get(tagName); err != nil {
		return 0, err
	} else {
		return tag.Int(0)
	}
}

func GetUInt32(decodedExif *exif.Exif, tagName exif.FieldName) (uint32, error) {
	if tag, err := decodedExif.Get(tagName); err != nil {
		return 0, err
	} else if intVal, err := tag.Int(0); err != nil {
		return 0, err
	} else {
		return uint32(intVal), err
	}
}

func GetString(decodedExif *exif.Exif, tagName exif.FieldName) (string, error) {
	if tag, err := decodedExif.Get(tagName); err != nil {
		return "", err
	} else {
		return tag.StringVal()
	}
}

func GetTime(decodedExif *exif.Exif, tagName exif.FieldName) (time.Time, error) {
	if stringVal, err := GetString(decodedExif, tagName); err != nil {
		return time.Unix(0, 0), err
	} else {
		return time.Parse("2006:01:02 15:04:05", stringVal)
	}
}

func LoadExifData(imageFile *ImageFile) (*ExifData, error) {
	fileForExif, err := os.Open(imageFile.GetPath())
	if fileForExif != nil && err == nil {
		defer fileForExif.Close()

		if decodedExif, err := exif.Decode(fileForExif); err != nil {
			logger.Error.Print("Could not decode Exif data", err)
			return nil, err
		} else if orientation, err := GetInt(decodedExif, exif.Orientation); err != nil {
			logger.Warn.Print("Could not resolve orientation", err)
			return getInvalidExifData(decodedExif), err
		} else if timeValue, err := GetTime(decodedExif, exif.DateTimeOriginal); err != nil {
			logger.Warn.Print("Could not resolve created value", err)
			return getInvalidExifData(decodedExif), err
		} else if width, err := GetUInt32(decodedExif, exif.PixelXDimension); err != nil {
			logger.Warn.Print("Could not resolve created value", err)
			return getInvalidExifData(decodedExif), err
		} else if height, err := GetUInt32(decodedExif, exif.PixelYDimension); err != nil {
			logger.Warn.Print("Could not resolve created value", err)
			return getInvalidExifData(decodedExif), err
		} else {
			angle, flip := ExifOrientationToAngleAndFlip(orientation)
			return &ExifData{
				orientation: uint8(orientation),
				rotation:    angle,
				flipped:     flip,
				raw:         decodedExif,
				created:     timeValue,
				width:       width,
				height:      height,
			}, nil
		}
	} else {
		return &ExifData{1, 0, false, nil, time.Unix(0, 0), 0, 0}, err
	}
}

func getInvalidExifData(decodedExif *exif.Exif) *ExifData {
	return &ExifData{1, 0, false, decodedExif, time.Unix(0, 0), 0, 0}
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

func ExifRotateImage(loadedImage image.Image, rotation float64, flipped bool) (image.Image, error) {
	loadedImage = imaging.Rotate(loadedImage, rotation, color.Black)
	if flipped {
		return imaging.FlipH(loadedImage), nil
	} else {
		return loadedImage, nil
	}
}
