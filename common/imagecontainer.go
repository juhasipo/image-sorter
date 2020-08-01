package common

import "image"

type ImageContainer struct {
	handle *Handle
	img    image.Image
}

func (s *ImageContainer) String() string {
	if s != nil {
		return "ImageContainer{" + s.handle.String() + "}"
	} else {
		return "ImageContainer<nil>"
	}
}

func (s *ImageContainer) GetHandle() *Handle {
	return s.handle
}

func (s *ImageContainer) GetImage() image.Image {
	return s.img
}

func ImageContainerNew(handle *Handle, img image.Image) *ImageContainer {
	return &ImageContainer{
		handle: handle,
		img:    img,
	}
}
