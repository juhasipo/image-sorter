package filter

import (
	"image"
	"log"
	"vincit.fi/image-sorter/common"
)

type fileOperation struct {
	dstPath string
	dstFile string
}

type ImageOperation interface {
	Apply(handle *common.Handle, img image.Image, data *common.ExifData) (image.Image, *common.ExifData, error)
	String() string
}

type ImageOperationGroup struct {
	handle     *common.Handle
	exifData   *common.ExifData
	img        image.Image
	operations []ImageOperation
}

func ImageOperationGroupNew(handle *common.Handle, img image.Image, exifData *common.ExifData, operations []ImageOperation) *ImageOperationGroup {
	return &ImageOperationGroup{
		handle:     handle,
		img:        img,
		exifData:   exifData,
		operations: operations,
	}
}

func (s *ImageOperationGroup) GetOperations() []ImageOperation {
	return s.operations
}

func (s *ImageOperationGroup) Apply() error {
	for _, operation := range s.operations {
		log.Printf("Applying: '%s'", operation)
		var err error
		if s.img, s.exifData, err = operation.Apply(s.handle, s.img, s.exifData); err != nil {
			return err
		}
	}
	return nil
}
