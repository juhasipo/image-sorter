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
func (s *ImageExifRotate) Apply(operationGroup *ImageOperationGroup) (image.Image, *common.ExifData, error) {
	handle := operationGroup.handle
	img := operationGroup.img
	data := operationGroup.exifData
	log.Printf("Exif rotate %s", handle.GetPath())
	rotatedImage, err := imageloader.ExifRotateImage(img, data)
	if err != nil {
		return img, data, err
	}
	data.ResetExifRotate()
	if img != rotatedImage {
		operationGroup.SetModified()
	}
	return rotatedImage, data, err
}
func (s *ImageExifRotate) String() string {
	return "Exif Rotate"
}
