package api

import (
	"vincit.fi/image-sorter/api/apitype"
)

type ErrorCommand struct {
	Message string
}

type DeviceFoundCommand struct {
	DeviceName string
}

type UpdateProgressCommand struct {
	Name      string
	Current   int
	Total     int
	CanCancel bool
	Modal     bool
}

type UpdateCategoriesCommand struct {
	Categories []*apitype.Category
}

type UpdateImageCommand struct {
	Image      *apitype.ImageFileAndData
	MetaData   *apitype.ImageMetaData
	Index      int
	Total      int
	CategoryId apitype.CategoryId
}

type SetImagesCommand struct {
	Topic  Topic
	Images []*apitype.ImageFileAndData
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
