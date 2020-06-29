package pixbuf


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

func NewPixbufCache() *PixbufCache {
	return &PixbufCache {
		imageCache: map[common.Handle]*Instance{},
	}
}

func (s *PixbufCache) GetInstance(handle *common.Handle) *Instance {
	if handle.IsValid() {
		s.mux.Lock()
		defer s.mux.Unlock()
		var instance *Instance
		if val, ok := s.imageCache[*handle]; !ok {
			instance = NewInstance(handle)
			s.imageCache[*handle] = instance
		} else {
			instance = val
		}

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


