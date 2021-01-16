package imageloader

import (
	"github.com/pixiv/go-libjpeg/jpeg"
	"image"
	"os"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/database"
)

var options = &jpeg.DecoderOptions{}

func NewImageLoader(imageStore *database.ImageStore) api.ImageLoader {
	return &LibJPEGImageLoader{
		imageStore: imageStore,
	}
}

type LibJPEGImageLoader struct {
	imageStore *database.ImageStore

	api.ImageLoader
}

func (s *LibJPEGImageLoader) LoadImage(imageId apitype.ImageId) (image.Image, error) {
	handle := s.imageStore.GetImageById(imageId)

	file, err := os.Open(handle.GetPath())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	imageFile, err := jpeg.Decode(file, options)
	if err != nil {
		return nil, err
	}

	rotation, flipped := handle.GetRotation()
	return apitype.ExifRotateImage(imageFile, rotation, flipped)
}

func (s *LibJPEGImageLoader) LoadImageScaled(imageId apitype.ImageId, size apitype.Size) (image.Image, error) {
	handle := s.imageStore.GetImageById(imageId)

	file, err := os.Open(handle.GetPath())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	imageFile, err := jpeg.Decode(file, &jpeg.DecoderOptions{ScaleTarget: image.Rect(0, 0, size.GetWidth(), size.GetHeight())})
	if err != nil {
		return nil, err
	}

	rotation, flipped := handle.GetRotation()
	return apitype.ExifRotateImage(imageFile, rotation, flipped)
}

func (s *LibJPEGImageLoader) LoadExifData(imageId apitype.ImageId) (*apitype.ExifData, error) {
	handle := s.imageStore.GetImageById(imageId)
	return apitype.LoadExifData(handle)
}
