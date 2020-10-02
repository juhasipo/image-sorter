package api

import (
	"image"
	"vincit.fi/image-sorter/common"
)

type ImageStore interface {
	Initialize([]*common.Handle)
	GetFull(*common.Handle) (image.Image, error)
	GetScaled(*common.Handle, common.Size) (image.Image, error)
	GetThumbnail(*common.Handle) (image.Image, error)
	GetExifData(handle *common.Handle) *common.ExifData
	GetByteSize() uint64
	GetSizeInMB() float64
	Purge()
}
