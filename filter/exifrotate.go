package filter

import (
	"image"
	"log"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/imageloader"
)

type ImageExifRotate struct {
	ImageOperation
}

func ImageExifRotateNew() ImageOperation {
	return &ImageExifRotate{}
}
func (s *ImageExifRotate) Apply(handle *common.Handle, img image.Image, data *common.ExifData) (image.Image, *common.ExifData, error) {
	log.Printf("Exif rotate %s", handle.GetPath())
	rotatedImage, err := imageloader.ExifRotateImage(img, data)
	if err != nil {
		return img, data, err
	}
	data.ResetExifRotate()
	return rotatedImage, data, err
}
func (s *ImageExifRotate) String() string {
	return "Exif Rotate"
}
