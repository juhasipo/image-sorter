package api

import (
	"image"
	"vincit.fi/image-sorter/api/apitype"
)

type ImageStore interface {
	Initialize([]*apitype.Handle)
	GetFull(*apitype.Handle) (image.Image, error)
	GetScaled(*apitype.Handle, apitype.Size) (image.Image, error)
	GetThumbnail(*apitype.Handle) (image.Image, error)
	GetExifData(handle *apitype.Handle) *apitype.ExifData
	GetByteSize() uint64
	GetSizeInMB() float64
	Purge()
}
