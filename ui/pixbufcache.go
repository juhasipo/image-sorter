package ui


import "C"
import (
	"github.com/gotk3/gotk3/gdk"
	"sync"
	"vincit.fi/image-sorter/common"
)

type PixbufCache struct {
	imageCache map[common.Handle]*Instance
	mux sync.Mutex
}

func (s *PixbufCache) GetInstance(handle *common.Handle) *Instance {
	if handle.IsValid() {
		s.mux.Lock()
		var instance *Instance
		if val, ok := s.imageCache[*handle]; !ok {
			instance = &Instance{
				handle: handle,
				loader: s,
			}
			s.imageCache[*handle] = instance
		} else {
			instance = val
		}
		s.mux.Unlock()
		return instance
	} else {
		return &EMPTY_INSTANCE
	}
}

func (s *PixbufCache) GetScaled(handle *common.Handle, size Size) *gdk.Pixbuf {
	return s.GetInstance(handle).GetScaled(size)
}
func (s *PixbufCache) GetThumbnail(handle *common.Handle) *gdk.Pixbuf {
	return s.GetInstance(handle).GetThumbnail()
}

func (s*PixbufCache) loadFromHandle(handle *common.Handle) (*gdk.Pixbuf, error) {
	return gdk.PixbufNewFromFile(handle.GetPath())
}
