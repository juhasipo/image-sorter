package apitype

import (
	"bytes"
	"errors"
	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
	"image"
	"image/color"
	"strings"
	"time"
	"vincit.fi/image-sorter/common/logger"
)

type MapExifWalker struct {
	values map[string]string

	exif.Walker
}

func NewMapExifWalker() *MapExifWalker {
	return &MapExifWalker{
		values: map[string]string{},
	}
}

func (s *MapExifWalker) Walk(name exif.FieldName, tag *tiff.Tag) error {
	if tagValue := strings.Trim(tag.String(), " \t\""); tagValue != "" {
		key := string(name)
		s.values[key] = tagValue
	}
	return nil
}

func (s *MapExifWalker) MetaData() map[string]string {
	return s.values
}

const (
	exifUnchangedOrientation  = 1
	exifValueMarker           = 0xFF
	orientationIntelOffset    = 8 // Offset for the value from the tag
	orientationMotorolaOffset = 9 // Offset for the value from the tag

	noRotate  = 0
	rotate180 = 180
	left90    = 90
	right90   = 270

	noHorizontalFlip = false
	horizontalFlip   = true
)

// Tag (2 bytes), type (2 bytes), count (4 bytes), value (2 bytes): 0xFF is the marker for value
// Intel byte order
var orientationIntelPattern = []byte{0x12, 0x01, 0x03, 0x00, 0x01, 0x00, 0x00, 0x00, exifValueMarker, 0x00}

// Motorola byte order
var orientationMotorolaPattern = []byte{0x01, 0x12, 0x00, 0x03, 0x00, 0x00, 0x00, 0x01, 0x00, exifValueMarker}

type ExifData struct {
	width       uint32
	height      uint32
	orientation uint8
	rotation    int16
	flipped     bool
	created     time.Time
	values      map[string]string
	raw         *exif.Exif
}

func NewExifData(decodedExif *exif.Exif) (*ExifData, error) {
	walker := NewMapExifWalker()
	if orientation, err := getInt(decodedExif, exif.Orientation); err != nil {
		logger.Warn.Print("Could not resolve orientation", err)
		return NewInvalidExifData(), err
	} else if timeValue, err := getTime(decodedExif, exif.DateTimeOriginal); err != nil {
		logger.Warn.Print("Could not resolve created value", err)
		return NewInvalidExifData(), err
	} else if width, err := getUInt32(decodedExif, exif.PixelXDimension); err != nil {
		logger.Warn.Print("Could not resolve created value", err)
		return NewInvalidExifData(), err
	} else if height, err := getUInt32(decodedExif, exif.PixelYDimension); err != nil {
		logger.Warn.Print("Could not resolve created value", err)
		return NewInvalidExifData(), err
	} else if err := decodedExif.Walk(walker); err != nil {
		logger.Warn.Print("Could not resolve meta data ", err)
		return NewInvalidExifData(), err
	} else {
		angle, flip := exifOrientationToAngleAndFlip(orientation)
		return &ExifData{
			orientation: uint8(orientation),
			rotation:    angle,
			flipped:     flip,
			raw:         decodedExif,
			created:     timeValue,
			width:       width,
			height:      height,
			values:      walker.MetaData(),
		}, nil
	}
}

func NewInvalidExifData() *ExifData {
	return &ExifData{
		orientation: 1,
		created:     time.Unix(0, 0),
	}
}

func NewExifDataFromMap(values map[string]string) *ExifData {
	return &ExifData{
		orientation: 1,
		created:     time.Unix(0, 0),
		values:      values,
	}
}

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

func (s *ExifData) ExifOrientation() uint8 {
	return s.orientation
}

func (s *ExifData) Rotation() int16 {
	return s.rotation
}

func (s *ExifData) Flipped() bool {
	return s.flipped
}

func (s *ExifData) CreatedTime() time.Time {
	return s.created
}

func (s *ExifData) RawExifData() []byte {
	return s.raw.Raw
}

func (s *ExifData) Get(name exif.FieldName) *tiff.Tag {
	if tag, err := s.raw.Get(name); err != nil {
		return nil
	} else {
		return tag
	}
}

func (s *ExifData) Values() map[string]string {
	return s.values
}

func (s *ExifData) RawExifDataLength() uint16 {
	return uint16(len(s.raw.Raw))
}

func (s *ExifData) ImageWidth() uint32 {
	return s.width
}

func (s *ExifData) ImageHeight() uint32 {
	return s.height
}

func ExifRotateImage(loadedImage image.Image, rotation float64, flipped bool) (image.Image, error) {
	loadedImage = imaging.Rotate(loadedImage, rotation, color.Black)
	if flipped {
		return imaging.FlipH(loadedImage), nil
	} else {
		return loadedImage, nil
	}
}

func getInt(decodedExif *exif.Exif, tagName exif.FieldName) (int, error) {
	if tag, err := decodedExif.Get(tagName); err != nil {
		return 0, err
	} else {
		return tag.Int(0)
	}
}

func getUInt32(decodedExif *exif.Exif, tagName exif.FieldName) (uint32, error) {
	if tag, err := decodedExif.Get(tagName); err != nil {
		return 0, err
	} else if intVal, err := tag.Int(0); err != nil {
		return 0, err
	} else {
		return uint32(intVal), err
	}
}

func getString(decodedExif *exif.Exif, tagName exif.FieldName) (string, error) {
	if tag, err := decodedExif.Get(tagName); err != nil {
		return "", err
	} else {
		return tag.StringVal()
	}
}

func getTime(decodedExif *exif.Exif, tagName exif.FieldName) (time.Time, error) {
	if stringVal, err := getString(decodedExif, tagName); err != nil {
		return time.Unix(0, 0), err
	} else {
		return time.Parse("2006:01:02 15:04:05", stringVal)
	}
}

func exifOrientationToAngleAndFlip(orientation int) (int16, bool) {
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
