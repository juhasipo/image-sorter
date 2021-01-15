package api

import (
	"image"
	"vincit.fi/image-sorter/api/apitype"
)

type ImageStore interface {
	Initialize([]*apitype.Handle)
	GetFull(apitype.HandleId) (image.Image, error)
	GetScaled(apitype.HandleId, apitype.Size) (image.Image, error)
	GetThumbnail(apitype.HandleId) (image.Image, error)
	GetExifData(apitype.HandleId) *apitype.ExifData
	GetByteSize() uint64
	GetSizeInMB() float64
	Purge()
}
