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
	handle := operationGroup.GetHandle()
	metaData := operationGroup.GetMetaData()
	img := operationGroup.GetImage()
	data := operationGroup.GetExifData()
	logger.Debug.Printf("Exif rotate %s", handle.GetPath())
	rotation, flipped := metaData.GetRotation()
	rotatedImage, err := apitype.ExifRotateImage(img, rotation, flipped)
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
