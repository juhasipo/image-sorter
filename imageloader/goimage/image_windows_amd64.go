package goimage

import (
	"image"
	"image/jpeg"
	"log"
	"os"
	"vincit.fi/image-sorter/common"
)

func LoadImage(handle *common.Handle) (image.Image, error) {
	log.Printf("Loading image %s", handle.GetId())
	imageFile, err := os.Open(handle.GetPath())
	if err != nil {
		return nil, err
	}
	defer imageFile.Close()
	if err != nil {
		return nil, err
	}

	// TODO: Convert to RGB if necessary
	return jpeg.Decode(imageFile)
}
