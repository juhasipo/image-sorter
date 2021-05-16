package api

type SelectDeviceCommand struct {
	Name           string
	ShowBackground bool
}

type Caster interface {
	StartServer(port int)
	StopServer()
	FindDevices()
	SelectDevice(*SelectDeviceCommand)
	CastImage(*ImageCategoryQuery)
	UpdateCategories(*UpdateCategoriesCommand)
	SetImageCategory(*CategoriesCommand)
	SetCurrentImage(*UpdateImageCommand)
	StopCasting()
	Close()
}
