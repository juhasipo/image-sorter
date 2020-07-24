package goimage

import (
	"image"
	"vincit.fi/image-sorter/common"
)

type ImageLoader interface {
	LoadImage(*common.Handle) (image.Image, error)
	LoadImageScaled(*common.Handle, common.Size) (image.Image, error)
	LoadExifData(*common.Handle) (*common.ExifData, error)
}
