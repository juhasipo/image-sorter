package common

type Handle struct {
	id string
	path string
}

func (s *Handle) IsValid() bool {
	return s != nil && s.id != ""
}

var (
	EMPTY_HANDLE = Handle {id: "", path: ""}
)

func GetEmptyHandle() *Handle {
	return &EMPTY_HANDLE
}

func (s* Handle) GetPath() string {
	return s.path
}
