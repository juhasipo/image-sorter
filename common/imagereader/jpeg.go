package imagereader

import (
	"github.com/pixiv/go-libjpeg/jpeg"
	"image"
	"os"
	"time"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
)

var options = &jpeg.DecoderOptions{}

func LoadImage(path string, rotation float64, flipped bool) (image.Image, error) {
	return loadImage(path, rotation, flipped, options)
}

func LoadScaledImage(path string, rotation float64, flipped bool, width int, height int) (image.Image, error) {
	return loadImage(path, rotation, flipped, &jpeg.DecoderOptions{ScaleTarget: image.Rect(0, 0, width, height)})
}

func loadImage(path string, rotation float64, flipped bool, options *jpeg.DecoderOptions) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	imageFile, err := jpeg.Decode(file, options)
	if err != nil {
		return nil, err
	}
	rotated, err := apitype.ExifRotateImage(imageFile, rotation, flipped)

	return ConvertNrgbaToRgba(rotated), nil
}

func ConvertNrgbaToRgba(i image.Image) image.Image {
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
