package imageloader

import (
	"errors"
	"image"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/internal/database"
	"vincit.fi/image-sorter/common/imagereader"
	"vincit.fi/image-sorter/common/logger"
	"vincit.fi/image-sorter/common/util"
)

func NewImageLoader(imageStore *database.ImageStore) api.ImageLoader {
	logger.Debug.Printf("Initializing image loader...")
	jpegLoader := &LibJPEGImageLoader{
		imageStore: imageStore,
	}
	logger.Debug.Printf("Image loader initialized")
	return jpegLoader
}

type LibJPEGImageLoader struct {
	imageStore *database.ImageStore

	api.ImageLoader
}

func (s *LibJPEGImageLoader) LoadImage(imageId apitype.ImageId) (image.Image, error) {
	if imageId != apitype.NoImage {
		if storedImageFile := s.imageStore.GetImageById(imageId); storedImageFile != nil {
			rotation, flipped := storedImageFile.Rotation()
			return imagereader.LoadImage(storedImageFile.Path(), rotation, flipped)
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
			rotation, flipped := imageFileWithMetaData.Rotation()
			return imagereader.LoadScaledImage(imageFileWithMetaData.Path(), rotation, flipped, size.Width(), size.Height())
		} else {
			return nil, errors.New("image not found in DB")
		}
	} else {
		return nil, errors.New("invalid image ID")
	}
}

func (s *LibJPEGImageLoader) LoadExifData(imageFile *apitype.ImageFile) (*apitype.ExifData, error) {
	return util.LoadExifData(imageFile)
}
