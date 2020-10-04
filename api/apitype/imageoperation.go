package apitype

import (
	"image"
	"vincit.fi/image-sorter/common/logger"
)

type ImageOperation interface {
	Apply(operationGroup *ImageOperationGroup) (image.Image, *ExifData, error)
	String() string
}

type ImageOperationGroup struct {
	handle          *Handle
	exifData        *ExifData
	img             image.Image
	hasBeenModified bool
	operations      []ImageOperation
}

func (s *ImageOperationGroup) GetHandle() *Handle {
	return s.handle
}

func (s *ImageOperationGroup) GetImage() image.Image {
	return s.img
}

func (s *ImageOperationGroup) GetExifData() *ExifData {
	return s.exifData
}

func (s *ImageOperationGroup) GetHasBeenModified() bool {
	return s.hasBeenModified
}

func (s *ImageOperationGroup) GetOperations() []ImageOperation {
	return s.operations
}

func NewImageOperationGroup(handle *Handle, img image.Image, exifData *ExifData, operations []ImageOperation) *ImageOperationGroup {
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
