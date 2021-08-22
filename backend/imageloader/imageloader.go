package imageloader

import (
	"errors"
	"github.com/pixiv/go-libjpeg/jpeg"
	"image"
	"image/color"
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
		if storedImageFile := s.imageStore.GetImageById(imageId); storedImageFile != nil {
			file, err := os.Open(storedImageFile.Path())
			if err != nil {
				return nil, err
			}
			defer file.Close()

			imageFile, err := jpeg.Decode(file, options)
			if err != nil {
				return nil, err
			}

			rotation, flipped := storedImageFile.Rotation()
			rotated, err := apitype.ExifRotateImage(imageFile, rotation, flipped)

			rotated = convertNrgbaToRgba(rotated)

			return rotated, err
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
			rotated, err := apitype.ExifRotateImage(imageFile, rotation, flipped)

			rotated = convertNrgbaToRgba(rotated)

			return rotated, err
		} else {
			return nil, errors.New("image not found in DB")
		}
	} else {
		return nil, errors.New("invalid image ID")
	}
}

func convertNrgbaToRgba(i image.Image) image.Image {
	n := i.(*image.NRGBA)

	rgba := image.NewRGBA(n.Rect)
	for x := 0; x < n.Rect.Dx(); x++ {
		for y := 0; y < n.Rect.Dy(); y++ {
			pix := n.NRGBAAt(x, y)
			r, g, b, a := pix.RGBA()
			c := color.RGBA{
				R: uint8(r / 256),
				G: uint8(g / 256),
				B: uint8(b / 256),
				A: uint8(a / 256),
			}
			rgba.SetRGBA(x, y, c)
		}
	}
	return rgba
}

func (s *LibJPEGImageLoader) LoadExifData(imageFile *apitype.ImageFile) (*apitype.ExifData, error) {
	return util.LoadExifData(imageFile)
}
