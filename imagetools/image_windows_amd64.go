package imagetools

import (
	"image"
	"image/jpeg"
	"log"
	"os"
	"vincit.fi/image-sorter/common"
)

func loadImage(handle *common.Handle) (image.Image, error) {
	log.Printf("Loading image %s", handle.GetId())
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
