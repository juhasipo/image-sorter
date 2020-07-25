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
func (s *ImageRemove) Apply(handle *common.Handle, img image.Image, data *common.ExifData) (image.Image, *common.ExifData, error) {
	log.Printf("Remove %s", handle.GetPath())
	return img, data, common.RemoveFile(handle.GetPath())
}
func (s *ImageRemove) String() string {
	return "Remove"
}
