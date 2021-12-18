package filter

import (
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
)

type ImageRotateToAngle struct {
	rotation float64
	apitype.ImageOperation
}

func NewImageRotateToAngle(angle int) apitype.ImageOperation {
	return &ImageRotateToAngle{
		rotation: float64(angle),
	}
}
func (s *ImageRotateToAngle) Apply(operationGroup *apitype.ImageOperationGroup) (image.Image, *apitype.ExifData, error) {
	if s.rotation != 0.0 {
		imageFile := operationGroup.ImageFile()
		imageData := operationGroup.ImageData()
		exifData := operationGroup.ExifData()
		logger.Debug.Printf("Rotate %s to andle %.0f", imageFile.Path(), s.rotation)
		rotatedImage := imaging.Rotate(imageData, s.rotation, image.Black)
		exifData.ResetExifRotate()
		operationGroup.SetModified()
		return rotatedImage, exifData, nil
	} else {
		return nil, nil, nil
	}

}
func (s *ImageRotateToAngle) String() string {
	return fmt.Sprintf("Rotate to %.2f", s.rotation)
}
