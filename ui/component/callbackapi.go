package component

type CallbackApi interface {
	ExitFullScreen()
	EnterFullScreenNoDistraction()
	EnterFullScreen()
	OpenFolderChooser(i int)
	ShowEditCategoriesModal()
	FindDevices()
}
