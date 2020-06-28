package library

import (
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
)

type ImageCommand struct {
	handles []*common.Handle
	event.Command
}

func (s *ImageCommand) GetHandles() []*common.Handle {
	return s.handles
}
