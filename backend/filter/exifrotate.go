package filter

import (
	"image"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/common/logger"
)

type ImageExifRotate struct {
	api.ImageOperation
}

func NewImageExifRotate() api.ImageOperation {
	return &ImageExifRotate{}
}
func (s *ImageExifRotate) Apply(operationGroup *api.ImageOperationGroup) (image.Image, *common.ExifData, error) {
	handle := operationGroup.GetHandle()
	img := operationGroup.GetImage()
	data := operationGroup.GetExifData()
	logger.Debug.Printf("Exif rotate %s", handle.GetPath())
	rotatedImage, err := common.ExifRotateImage(img, data)
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
