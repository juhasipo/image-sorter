package filter

import (
	"image"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/logger"
)

type ImageRemove struct {
	ImageOperation
}

func NewImageRemove() ImageOperation {
	return &ImageRemove{}
}
func (s *ImageRemove) Apply(operationGroup *ImageOperationGroup) (image.Image, *common.ExifData, error) {
	handle := operationGroup.handle
	img := operationGroup.img
	data := operationGroup.exifData
	logger.Debug.Printf("Remove %s", handle.GetPath())
	return img, data, common.RemoveFile(handle.GetPath())
}
func (s *ImageRemove) String() string {
	return "Remove"
}
