package imageloader

import (
	"github.com/pixiv/go-libjpeg/jpeg"
	"image"
	"os"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
)

var options = &jpeg.DecoderOptions{}

func NewImageLoader() api.ImageLoader {
	return &LibJPEGImageLoader{}
}

type LibJPEGImageLoader struct {
	api.ImageLoader
}

func (s *LibJPEGImageLoader) LoadImage(handle *apitype.Handle) (image.Image, error) {
	file, err := os.Open(handle.GetPath())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	imageFile, err := jpeg.Decode(file, options)
	if err != nil {
		return nil, err
	}
	return imageFile, err
}

func (s *LibJPEGImageLoader) LoadImageScaled(handle *apitype.Handle, size apitype.Size) (image.Image, error) {
	file, err := os.Open(handle.GetPath())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	imageFile, err := jpeg.Decode(file, &jpeg.DecoderOptions{ScaleTarget: image.Rect(0, 0, size.GetWidth(), size.GetHeight())})
	if err != nil {
		return nil, err
	}
	return imageFile, err
}

func (s *LibJPEGImageLoader) LoadExifData(handle *apitype.Handle) (*apitype.ExifData, error) {
	return apitype.LoadExifData(handle)
}
