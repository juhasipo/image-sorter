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
	Apply(operationGroup *ImageOperationGroup) (image.Image, *common.ExifData, error)
	String() string
}

type ImageOperationGroup struct {
	handle          *common.Handle
	exifData        *common.ExifData
	img             image.Image
	hasBeenModified bool
	operations      []ImageOperation
}

func ImageOperationGroupNew(handle *common.Handle, img image.Image, exifData *common.ExifData, operations []ImageOperation) *ImageOperationGroup {
	return &ImageOperationGroup{
		handle:          handle,
		img:             img,
		exifData:        exifData,
		hasBeenModified: false,
		operations:      operations,
	}
}

func (s *ImageOperationGroup) SetModified() {
	s.hasBeenModified = true
}

func (s *ImageOperationGroup) GetOperations() []ImageOperation {
	return s.operations
}

func (s *ImageOperationGroup) Apply() error {
	for _, operation := range s.operations {
		log.Printf("Applying: '%s'", operation)
		var err error
		if s.img, s.exifData, err = operation.Apply(s); err != nil {
			return err
		}
	}
	return nil
}
