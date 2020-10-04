package api

import (
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/event"
)

type Gui interface {
	SetCurrentImage(*
	apitype.ImageContainer, int, int, string, *apitype.ExifData)
	SetImages(event.Topic, []*apitype.ImageContainer)
	UpdateCategories(categories *apitype.CategoriesCommand)
	SetImageCategory(command []*apitype.CategorizeCommand)
	Run()

	UpdateProgress(name string, status int, total int)

	DeviceFound(name string)
	CastReady()
	CastFindDone()
}
