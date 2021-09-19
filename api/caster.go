package api

import "vincit.fi/image-sorter/api/apitype"

type SelectDeviceCommand struct {
	Name           string
	ShowBackground bool
	apitype.NotThrottled
}

type Caster interface {
	StartServer(port int)
	StopServer()
	FindDevices()
	SelectDevice(*SelectDeviceCommand)
	CastImage(*ImageCategoryQuery)
	StopCasting()
	Close()
}
