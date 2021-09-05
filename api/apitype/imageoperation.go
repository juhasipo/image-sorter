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
	imageLoader     func()
	imageData       image.Image
	hasBeenModified bool
	operations      []ImageOperation
	loadImage       func(ImageId) (image.Image, error)
	loadExifData    func(*ImageFile) (*ExifData, error)
}

func (s *ImageOperationGroup) ImageFile() *ImageFile {
	return s.imageFile
}

func (s *ImageOperationGroup) ImageData() image.Image {
	if s.imageData == nil {
		s.imageData, _ = s.loadImage(s.imageFile.Id())
	}
	return s.imageData
}

func (s *ImageOperationGroup) ExifData() *ExifData {
	if s.exifData == nil {
		s.exifData, _ = s.loadExifData(s.imageFile)
	}
	return s.exifData
}

func (s *ImageOperationGroup) Modified() bool {
	return s.hasBeenModified
}

func (s *ImageOperationGroup) Operations() []ImageOperation {
	return s.operations
}

func NewImageOperationGroup(imageFile *ImageFile, loadImage func(ImageId) (image.Image, error), data func(*ImageFile) (*ExifData, error), operations []ImageOperation) *ImageOperationGroup {
	return &ImageOperationGroup{
		imageFile:       imageFile,
		loadImage:       loadImage,
		loadExifData:    data,
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
		imgData, exifData, err := operation.Apply(s)
		if err != nil {
			return err
		}

		if imgData != nil {
			s.imageData = imgData
			s.SetModified()
		}

		if exifData != nil {
			s.exifData = exifData
			s.SetModified()
		}
	}
	s.imageData = nil
	s.exifData = nil
	return nil
}
