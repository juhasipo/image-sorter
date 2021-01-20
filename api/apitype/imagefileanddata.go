package apitype

import (
	"image"
)

type ImageFileAndData struct {
	imageFile *ImageFile
	metaData  *ImageMetaData
	imageData image.Image
}

func (s *ImageFileAndData) String() string {
	if s != nil {
		return "ImageFileAndData{" + s.imageFile.String() + "}"
	} else {
		return "ImageFileAndData<nil>"
	}
}

func (s *ImageFileAndData) ImageFile() *ImageFile {
	return s.imageFile
}

func (s *ImageFileAndData) ImageData() image.Image {
	return s.imageData
}

func (s *ImageFileAndData) MetaData() *ImageMetaData {
	return s.metaData
}

func NewImageContainer(imageFile *ImageFileWithMetaData, imageData image.Image) *ImageFileAndData {
	if imageFile != nil {
		return &ImageFileAndData{
			imageFile: &imageFile.ImageFile,
			metaData:  &imageFile.ImageMetaData,
			imageData: imageData,
		}
	} else {
		return &ImageFileAndData{
			imageFile: nil,
			metaData:  nil,
			imageData: imageData,
		}
	}
}

func NewEmptyImageContainer() *ImageFileAndData {
	return &ImageFileAndData{
		imageFile: nil,
		metaData:  nil,
		imageData: nil,
	}
}
