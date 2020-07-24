package goimage

import (
	"github.com/pixiv/go-libjpeg/jpeg"
	"image"
	"os"
	"vincit.fi/image-sorter/common"
)

var options = &jpeg.DecoderOptions{}

func ImageLoaderNew() ImageLoader {
	return &ImageLoaderLinux{}
}

type ImageLoaderLinux struct {
	ImageLoader
}

func (s *ImageLoaderLinux) LoadImage(handle *common.Handle) (image.Image, error) {
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

func (s *ImageLoaderLinux) LoadImageScaled(handle *common.Handle, size common.Size) (image.Image, error) {
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

func (s *ImageLoaderLinux) LoadExifData(handle *common.Handle) (*common.ExifData, error) {
	return common.LoadExifData(handle)
}
