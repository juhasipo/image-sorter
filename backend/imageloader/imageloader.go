package imageloader

import (
	"errors"
	"github.com/pixiv/go-libjpeg/jpeg"
	"image"
	"os"
	"time"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/database"
	"vincit.fi/image-sorter/common/logger"
	"vincit.fi/image-sorter/common/util"
)

var options = &jpeg.DecoderOptions{}

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
	start := time.Now()
	n := i.(*image.NRGBA)

	rgba := image.NewRGBA(n.Rect)
	for x := 0; x < n.Rect.Dx(); x++ {
		for y := 0; y < n.Rect.Dy(); y++ {
			// Just pass the pixels as-is without any conversion
			// because all the images are JPGE which means, there is
			// no alpha-channel in use. This should allow us to do this
			nrgbaPixOffset := n.PixOffset(x, y)
			ngrbaStride := n.Pix[nrgbaPixOffset : nrgbaPixOffset+4 : nrgbaPixOffset+4]

			rgbaPixOffset := rgba.PixOffset(x, y)
			rgbaStride := rgba.Pix[rgbaPixOffset : rgbaPixOffset+4 : rgbaPixOffset+4]

			rgbaStride[0] = ngrbaStride[0]
			rgbaStride[1] = ngrbaStride[1]
			rgbaStride[2] = ngrbaStride[2]
			rgbaStride[3] = ngrbaStride[3]
		}
	}
	end := time.Now()

	if logger.IsLogLevel(logger.TRACE) {
		logger.Trace.Printf("Converting from NRGBA to RGBA: %s", end.Sub(start))
	}

	return rgba
}

func (s *LibJPEGImageLoader) LoadExifData(imageFile *apitype.ImageFile) (*apitype.ExifData, error) {
	return util.LoadExifData(imageFile)
}
