package api

import (
	"vincit.fi/image-sorter/api/apitype"
)

type ErrorCommand struct {
	Message string
	apitype.NotThrottled
}

type DeviceFoundCommand struct {
	DeviceName string
	apitype.NotThrottled
}

type UpdateProgressCommand struct {
	Name      string
	Current   int
	Total     int
	CanCancel bool
	Modal     bool
	apitype.NotThrottled
}

type UpdateCategoriesCommand struct {
	Categories []*apitype.Category
	apitype.NotThrottled
}

type UpdateImageCommand struct {
	Image      *apitype.ImageFile
	MetaData   *apitype.ImageMetaData
	Index      int
	Total      int
	CategoryId apitype.CategoryId
	apitype.NotThrottled
}

type SetImagesCommand struct {
	Topic  Topic
	Images []*apitype.ImageFile
	apitype.NotThrottled
}

type Gui interface {
	SetCurrentImage(*UpdateImageCommand)
	SetImages(*SetImagesCommand)
	UpdateCategories(*UpdateCategoriesCommand)
	SetImageCategory(*CategoriesCommand)
	ShowError(*ErrorCommand)
	Run()

	Pause()
	Ready()

	UpdateProgress(*UpdateProgressCommand)

	DeviceFound(*DeviceFoundCommand)
	CastReady()
	CastFindDone()
}
