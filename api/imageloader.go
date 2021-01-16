package api

import (
	"image"
	"vincit.fi/image-sorter/api/apitype"
)

type ImageLoader interface {
	LoadImage(apitype.ImageId) (image.Image, error)
	LoadImageScaled(apitype.ImageId, apitype.Size) (image.Image, error)
	LoadExifData(apitype.ImageId) (*apitype.ExifData, error)
}
