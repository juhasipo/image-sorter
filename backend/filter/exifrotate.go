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
	rotation, flipped := operationGroup.ImageFile().Rotation()

	logger.Debug.Printf("Exif rotate %s: angle=%f, flipped=%s", imageFile.Path(), rotation, flipped)

	if rotation != 0.0 || flipped {
		logger.Debug.Printf("Image %s needs to be rotated", imageFile.Path())
		imageData := operationGroup.ImageData()
		exifData := operationGroup.ExifData()
		exifData.ResetExifRotate()
		return imageData, exifData, nil
	} else {
		logger.Debug.Printf("No rotation needed for %s", imageFile.Path())
		return nil, nil, nil
	}
}
func (s *ImageExifRotate) String() string {
	return "Exif Rotate"
}
