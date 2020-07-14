package imagetools

import (
	"image"
	"image/jpeg"
	"os"
	"vincit.fi/image-sorter/common"
)

func LoadImage(handle *common.Handle) (image.Image, error) {
	imageFile, err := os.Open(handle.GetPath())
	if err != nil {
		return nil, err
	}
	defer imageFile.Close()
	if err != nil {
		return nil, err
	}
	return jpeg.Decode(imageFile)
}
