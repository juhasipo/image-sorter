package common

import "github.com/rivo/duplo"

type Handle struct {
	id string
	path string
	hash *duplo.Hash
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

func (s *Handle) SetHash(hash *duplo.Hash) {
	s.hash = hash
}
func (s *Handle) GetHash() *duplo.Hash {
	return s.hash
}
