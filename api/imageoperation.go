package api

import (
	"image"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/common/logger"
)

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

func (s *ImageOperationGroup) GetHandle() *common.Handle {
	return s.handle
}

func (s *ImageOperationGroup) GetImage() image.Image {
	return s.img
}

func (s *ImageOperationGroup) GetExifData() *common.ExifData {
	return s.exifData
}

func (s *ImageOperationGroup) GetHasBeenModified() bool {
	return s.hasBeenModified
}

func (s *ImageOperationGroup) GetOperations() []ImageOperation {
	return s.operations
}

func NewImageOperationGroup(handle *common.Handle, img image.Image, exifData *common.ExifData, operations []ImageOperation) *ImageOperationGroup {
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

func (s *ImageOperationGroup) Apply() error {
	for _, operation := range s.operations {
		logger.Debug.Printf("Applying: '%s'", operation)
		var err error
		if s.img, s.exifData, err = operation.Apply(s); err != nil {
			return err
		}
	}
	return nil
}
