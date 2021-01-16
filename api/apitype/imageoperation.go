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
	imageFile       *ImageFileWithMetaData
	exifData        *ExifData
	img             image.Image
	hasBeenModified bool
	operations      []ImageOperation
}

func (s *ImageOperationGroup) GetImageFile() *ImageFile {
	return s.imageFile.GetImageFile()
}

func (s *ImageOperationGroup) GetMetaData() *ImageMetaData {
	return s.imageFile.GetMetaData()
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

func NewImageOperationGroup(imageFile *ImageFileWithMetaData, img image.Image, exifData *ExifData, operations []ImageOperation) *ImageOperationGroup {
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
