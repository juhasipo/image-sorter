package library

import (
	"github.com/disintegration/imaging"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"strings"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/pixbuf"
)

var (
	supportedFileEndings = map[string]bool{".jpg": true, ".jpeg": true}
)

func LoadImages(dir string) []*common.Handle {
	var handles []*common.Handle
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Scanning directory '%s'", dir)
	for _, file := range files {
		filePath := path.Join(dir, file.Name())
		extension := filepath.Ext(filePath)
		if IsSupported(extension) {
			filePath := path.Join(dir, file.Name())
			handles = append(handles, common.HandleNew(filePath))
		}
	}
	log.Printf("Found %d images", len(handles))

	return handles
}

func IsSupported(extension string) bool {
	return supportedFileEndings[strings.ToLower(extension)]
}

func LoadImageWithExifCorrection(handle *common.Handle, exifData *pixbuf.ExifData) (*image.Image, error) {
	img, err := LoadImage(handle)

	img = imaging.Rotate(img, float64(exifData.GetRotation()), color.Gray{})
	if exifData.IsFlipped() {
		img = imaging.FlipH(img)
	}
	return &img, err
}
