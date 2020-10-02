package filter

import (
	"image"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/common/logger"
)

type ImageRemove struct {
	api.ImageOperation
}

func NewImageRemove() api.ImageOperation {
	return &ImageRemove{}
}
func (s *ImageRemove) Apply(operationGroup *api.ImageOperationGroup) (image.Image, *common.ExifData, error) {
	handle := operationGroup.GetHandle()
	img := operationGroup.GetImage()
	data := operationGroup.GetExifData()
	logger.Debug.Printf("Remove %s", handle.GetPath())
	return img, data, common.RemoveFile(handle.GetPath())
}
func (s *ImageRemove) String() string {
	return "Remove"
}
