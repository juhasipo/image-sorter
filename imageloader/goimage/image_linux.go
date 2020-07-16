package goimage

import (
	"github.com/disintegration/imaging"
	"image"
	"vincit.fi/image-sorter/common"
)

func LoadImage(handle *common.Handle) (image.Image, error) {
	imageFile, err := imaging.Open(handle.GetPath())
	if err != nil {
		return nil, err
	}
	return imageFile, err
}
