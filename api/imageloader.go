package api

import (
	"image"
	"vincit.fi/image-sorter/api/apitype"
)

type ImageLoader interface {
	LoadImage(*apitype.Handle) (image.Image, error)
	LoadImageScaled(*apitype.Handle, apitype.Size) (image.Image, error)
	LoadExifData(*apitype.Handle) (*apitype.ExifData, error)
}
