package api

import (
	"image"
	"vincit.fi/image-sorter/api/apitype"
)

type ImageStore interface {
	Initialize([]*apitype.Handle)
	GetFull(apitype.ImageId) (image.Image, error)
	GetScaled(apitype.ImageId, apitype.Size) (image.Image, error)
	GetThumbnail(apitype.ImageId) (image.Image, error)
	GetExifData(apitype.ImageId) *apitype.ExifData
	GetByteSize() uint64
	GetSizeInMB() float64
	Purge()
}
