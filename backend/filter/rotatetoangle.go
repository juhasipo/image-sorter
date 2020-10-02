package filter

import (
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/common/logger"
)

type ImageRotateToAngle struct {
	rotation float64
	api.ImageOperation
}

func NewImageRotateToAngle(angle int) api.ImageOperation {
	return &ImageRotateToAngle{
		rotation: float64(angle),
	}
}
func (s *ImageRotateToAngle) Apply(operationGroup *api.ImageOperationGroup) (image.Image, *common.ExifData, error) {
	handle := operationGroup.GetHandle()
	img := operationGroup.GetImage()
	data := operationGroup.GetExifData()
	if s.rotation != 0.0 {
		logger.Debug.Printf("Rotate %s to andle %.0f", handle.GetPath(), s.rotation)
		rotatedImage := imaging.Rotate(img, s.rotation, image.Black)
		data.ResetExifRotate()
		operationGroup.SetModified()
		return rotatedImage, data, nil
	} else {
		return img, data, nil
	}

}
func (s *ImageRotateToAngle) String() string {
	return fmt.Sprintf("Rotate to %.2f", s.rotation)
}
