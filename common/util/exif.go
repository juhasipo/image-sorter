package util

import (
	"github.com/rwcarlsen/goexif/exif"
	"os"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
)

func LoadExifData(imageFile *apitype.ImageFile) (*apitype.ExifData, error) {
	fileForExif, err := os.Open(imageFile.Path())
	if fileForExif != nil && err == nil {
		defer fileForExif.Close()

		if decodedExif, err := exif.Decode(fileForExif); err != nil {
			logger.Error.Print("Could not decode Exif data", err)
			return nil, err
		} else {
			return apitype.NewExifData(decodedExif)
		}

	} else {
		return apitype.NewInvalidExifData(), err
	}
}
