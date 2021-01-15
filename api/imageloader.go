package api

import (
	"image"
	"vincit.fi/image-sorter/api/apitype"
)

type ImageLoader interface {
	LoadImage(apitype.HandleId) (image.Image, error)
	LoadImageScaled(apitype.HandleId, apitype.Size) (image.Image, error)
	LoadExifData(apitype.HandleId) (*apitype.ExifData, error)
}
