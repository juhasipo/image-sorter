package filter

import (
	"image"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/util"
	"vincit.fi/image-sorter/common/logger"
)

type ImageRemove struct {
	apitype.ImageOperation
}

func NewImageRemove() apitype.ImageOperation {
	return &ImageRemove{}
}
func (s *ImageRemove) Apply(operationGroup *apitype.ImageOperationGroup) (image.Image, *apitype.ExifData, error) {
	handle := operationGroup.GetHandle()
	img := operationGroup.GetImage()
	data := operationGroup.GetExifData()
	logger.Debug.Printf("Remove %s", handle.GetPath())
	return img, data, util.RemoveFile(handle.GetPath())
}
func (s *ImageRemove) String() string {
	return "Remove"
}
