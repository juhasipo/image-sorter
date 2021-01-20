package imageloader

import (
	"errors"
	"github.com/pixiv/go-libjpeg/jpeg"
	"image"
	"os"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/database"
	"vincit.fi/image-sorter/common/util"
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
	if imageId != apitype.NoImage {
		if imageFileWithMetaData := s.imageStore.GetImageById(imageId); imageFileWithMetaData != nil {
			file, err := os.Open(imageFileWithMetaData.Path())
			if err != nil {
				return nil, err
			}
			defer file.Close()

			imageFile, err := jpeg.Decode(file, options)
			if err != nil {
				return nil, err
			}

			rotation, flipped := imageFileWithMetaData.Rotation()
			return apitype.ExifRotateImage(imageFile, rotation, flipped)
		} else {
			return nil, errors.New("image not found in DB")
		}
	} else {
		return nil, errors.New("invalid image ID")
	}
}

func (s *LibJPEGImageLoader) LoadImageScaled(imageId apitype.ImageId, size apitype.Size) (image.Image, error) {
	if imageId != apitype.NoImage {
		if imageFileWithMetaData := s.imageStore.GetImageById(imageId); imageFileWithMetaData != nil {
			file, err := os.Open(imageFileWithMetaData.Path())
			if err != nil {
				return nil, err
			}
			defer file.Close()

			imageFile, err := jpeg.Decode(file, &jpeg.DecoderOptions{ScaleTarget: image.Rect(0, 0, size.Width(), size.Height())})
			if err != nil {
				return nil, err
			}

			rotation, flipped := imageFileWithMetaData.Rotation()
			return apitype.ExifRotateImage(imageFile, rotation, flipped)
		} else {
			return nil, errors.New("image not found in DB")
		}
	} else {
		return nil, errors.New("invalid image ID")
	}
}

func (s *LibJPEGImageLoader) LoadExifData(imageId apitype.ImageId) (*apitype.ExifData, error) {
	imageFile := s.imageStore.GetImageById(imageId)
	return util.LoadExifData(&imageFile.ImageFile)
}
