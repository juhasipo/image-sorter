package filter

import (
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	"log"
	"vincit.fi/image-sorter/common"
)

type ImageRotateToAngle struct {
	rotation float64
	ImageOperation
}

func ImageRotateToAngleNew(angle int) ImageOperation {
	return &ImageRotateToAngle{
		rotation: float64(angle),
	}
}
func (s *ImageRotateToAngle) Apply(handle *common.Handle, img image.Image, data *common.ExifData) (image.Image, *common.ExifData, error) {
	log.Printf("Exif rotate %s", handle.GetPath())
	rotatedImage := imaging.Rotate(img, s.rotation, image.Black)
	data.ResetExifRotate()
	return rotatedImage, data, nil
}
func (s *ImageRotateToAngle) String() string {
	return fmt.Sprintf("Rotate to %.2f", s.rotation)
}
