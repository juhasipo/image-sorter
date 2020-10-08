package api

import "vincit.fi/image-sorter/api/apitype"

type Caster interface {
	StartServer(port int)
	StopServer()
	FindDevices()
	SelectDevice(name string, showBackground bool)
	CastImage(handle *apitype.Handle)
	StopCasting()
	Close()
}
