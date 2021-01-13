package api

import (
	"vincit.fi/image-sorter/api/apitype"
)

type Gui interface {
	SetCurrentImage(*apitype.ImageContainer, int, int, apitype.CategoryId)
	SetImages(Topic, []*apitype.ImageContainer)
	UpdateCategories(categories *apitype.CategoriesCommand)
	SetImageCategory(command []*apitype.CategorizeCommand)
	ShowError(message string)
	Run()

	UpdateProgress(name string, status int, total int)

	DeviceFound(name string)
	CastReady()
	CastFindDone()
}
