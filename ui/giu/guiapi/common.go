package guiapi

type CategoryAction struct {
	StayOnImage      bool
	ForceCategory    bool
	ShowOnlyCategory bool
}

type ZoomMode uint8

const (
	ZoomFixed ZoomMode = iota
	ZoomFit
	ZoomCustom
)
