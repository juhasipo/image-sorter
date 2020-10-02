package api

import (
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/common/event"
)

type Gui interface {
	SetCurrentImage(*common.ImageContainer, int, int, string, *common.ExifData)
	SetImages(event.Topic, []*common.ImageContainer)
	UpdateCategories(categories *CategoriesCommand)
	SetImageCategory(command []*CategorizeCommand)
	Run()

	UpdateProgress(name string, status int, total int)

	DeviceFound(name string)
	CastReady()
	CastFindDone()
}
