package apitype

import (
	"image"
)

type ImageFileAndData struct {
	imageFile *ImageFile
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

func NewImageContainer(imageFile *ImageFile, imageData image.Image) *ImageFileAndData {
	if imageFile != nil {
		return &ImageFileAndData{
			imageFile: imageFile,
			imageData: imageData,
		}
	} else {
		return &ImageFileAndData{
			imageFile: nil,
			imageData: imageData,
		}
	}
}

func NewEmptyImageContainer() *ImageFileAndData {
	return &ImageFileAndData{
		imageFile: nil,
		imageData: nil,
	}
}
