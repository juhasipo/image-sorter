package filter

import (
	"image"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
)

type ImageExifRotate struct {
	apitype.ImageOperation
}

func NewImageExifRotate() apitype.ImageOperation {
	return &ImageExifRotate{}
}
func (s *ImageExifRotate) Apply(operationGroup *apitype.ImageOperationGroup) (image.Image, *apitype.ExifData, error) {
	imageFile := operationGroup.ImageFile()
	metaData := operationGroup.MetaData()
	imageData := operationGroup.ImageData()
	exifData := operationGroup.ExifData()
	logger.Debug.Printf("Exif rotate %s", imageFile.Path())
	rotation, flipped := metaData.Rotation()
	rotatedImage, err := apitype.ExifRotateImage(imageData, rotation, flipped)
	if err != nil {
		return imageData, exifData, err
	}
	exifData.ResetExifRotate()
	if imageData != rotatedImage {
		operationGroup.SetModified()
	}
	return rotatedImage, exifData, err
}
func (s *ImageExifRotate) String() string {
	return "Exif Rotate"
}
