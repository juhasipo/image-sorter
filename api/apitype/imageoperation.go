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
	imageFile       *ImageFile
	exifData        *ExifData
	img             image.Image
	hasBeenModified bool
	operations      []ImageOperation
}

func (s *ImageOperationGroup) ImageFile() *ImageFile {
	return s.imageFile
}

func (s *ImageOperationGroup) ImageData() image.Image {
	return s.img
}

func (s *ImageOperationGroup) ExifData() *ExifData {
	return s.exifData
}

func (s *ImageOperationGroup) Modified() bool {
	return s.hasBeenModified
}

func (s *ImageOperationGroup) Operations() []ImageOperation {
	return s.operations
}

func NewImageOperationGroup(imageFile *ImageFile, img image.Image, exifData *ExifData, operations []ImageOperation) *ImageOperationGroup {
	return &ImageOperationGroup{
		imageFile:       imageFile,
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
