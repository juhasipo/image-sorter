package apitype

import (
	"image"
)

type ImageContainer struct {
	imageFile *ImageFile
	metaData  *ImageMetaData
	imageData image.Image
}

func (s *ImageContainer) String() string {
	if s != nil {
		return "ImageContainer{" + s.imageFile.String() + "}"
	} else {
		return "ImageContainer<nil>"
	}
}

func (s *ImageContainer) ImageFile() *ImageFile {
	return s.imageFile
}

func (s *ImageContainer) ImageData() image.Image {
	return s.imageData
}

func (s *ImageContainer) MetaData() *ImageMetaData {
	return s.metaData
}

func NewImageContainer(imageFile *ImageFileWithMetaData, imageData image.Image) *ImageContainer {
	if imageFile != nil {
		return &ImageContainer{
			imageFile: &imageFile.ImageFile,
			metaData:  &imageFile.ImageMetaData,
			imageData: imageData,
		}
	} else {
		return &ImageContainer{
			imageFile: nil,
			metaData:  nil,
			imageData: imageData,
		}
	}
}

func NewEmptyImageContainer() *ImageContainer {
	return &ImageContainer{
		imageFile: nil,
		metaData:  nil,
		imageData: nil,
	}
}
