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
	imageFile := operationGroup.ImageFile()
	logger.Debug.Printf("Remove %s", imageFile.Path())
	return nil, nil, util.RemoveFile(imageFile.Path())
}
func (s *ImageRemove) String() string {
	return "Remove"
}
