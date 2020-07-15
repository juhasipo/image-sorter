package goimage

import (
	"github.com/disintegration/imaging"
	"image"
	"image/color"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/imagetools"
)

func LoadGoImageWithExifCorrection(handle *common.Handle, exifData *imagetools.ExifData) (image.Image, error) {
	img, err := LoadImage(handle)

	img = imaging.Rotate(img, float64(exifData.GetRotation()), color.Gray{})
	if exifData.IsFlipped() {
		img = imaging.FlipH(img)
	}
	return img, err
}
