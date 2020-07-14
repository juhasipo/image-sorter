package imagetools

import (
	"github.com/disintegration/imaging"
	"image"
	"image/color"
	"vincit.fi/image-sorter/common"
)

func LoadImageWithExifCorrection(handle *common.Handle, exifData *ExifData) (image.Image, error) {
	img, err := LoadImage(handle)

	img = imaging.Rotate(img, float64(exifData.GetRotation()), color.Gray{})
	if exifData.IsFlipped() {
		img = imaging.FlipH(img)
	}
	return img, err
}

