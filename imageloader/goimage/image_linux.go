package goimage

import (
	"github.com/pixiv/go-libjpeg/jpeg"
	"image"
	"os"
	"vincit.fi/image-sorter/common"
)

var options = &jpeg.DecoderOptions{}

func LoadImage(handle *common.Handle) (image.Image, error) {
	file, err := os.Open(handle.GetPath())
	if err != nil {
		return nil, err
	}

	imageFile, err := jpeg.Decode(file, options)
	if err != nil {
		return nil, err
	}
	return imageFile, err
}

func LoadImageScaled(handle *common.Handle, size common.Size) (image.Image, error) {
	file, err := os.Open(handle.GetPath())
	if err != nil {
		return nil, err
	}

	imageFile, err := jpeg.Decode(file, &jpeg.DecoderOptions{ScaleTarget: image.Rect(0, 0, size.GetWidth(), size.GetHeight())})
	if err != nil {
		return nil, err
	}
	return imageFile, err
}
