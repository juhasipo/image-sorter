package filter

import (
	"image"
	"log"
	"vincit.fi/image-sorter/common"
)

type ImageRemove struct {
	ImageOperation
}

func ImageRemoveNew() ImageOperation {
	return &ImageRemove{}
}
func (s *ImageRemove) Apply(operationGroup *ImageOperationGroup) (image.Image, *common.ExifData, error) {
	handle := operationGroup.handle
	img := operationGroup.img
	data := operationGroup.exifData
	log.Printf("Remove %s", handle.GetPath())
	return img, data, common.RemoveFile(handle.GetPath())
}
func (s *ImageRemove) String() string {
	return "Remove"
}
