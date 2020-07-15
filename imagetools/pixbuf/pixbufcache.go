package pixbuf


import "C"
import (
	"github.com/gotk3/gotk3/gdk"
	"log"
	"runtime"
	"sync"
	"vincit.fi/image-sorter/common"
)

type PixbufCache struct {
	imageCache map[*common.Handle]*Instance
	mux sync.Mutex
}

func NewPixbufCache() *PixbufCache {
	return &PixbufCache {
		imageCache: map[*common.Handle]*Instance{},
	}
}

func (s *PixbufCache) Initialize(initialThumbnails []*common.Handle) {
	s.imageCache = map[*common.Handle]*Instance{}
	for _, handle := range initialThumbnails {
		s.imageCache[handle] = NewInstance(handle)
	}
	runtime.GC()
}

func (s *PixbufCache) GetInstance(handle *common.Handle) *Instance {
	if handle.IsValid() {
		s.mux.Lock()
		defer s.mux.Unlock()
		if existingInstance, ok := s.imageCache[handle]; !ok {
			instance := NewInstance(handle)
			s.imageCache[handle] = instance
			return instance
		} else {
			return existingInstance
		}
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

func (s *PixbufCache) Purge(ignored *common.Handle) {
	byteLength := s.GetByteLength()
	log.Printf("Before purge cache has %d B (%.2f MB)", byteLength, float64(byteLength)/float64(1024*1024))

	for handle, instance := range s.imageCache {
		if handle != ignored {
			instance.Purge()
		}
	}
	byteLength = s.GetByteLength()
	log.Printf("After purge cache has %d B (%.2f MB)", byteLength, float64(byteLength)/float64(1024*1024))
	runtime.GC()
}

func (s *PixbufCache) GetByteLength() int {
	var byteLength = 0
	for _, instance := range s.imageCache {
		byteLength += instance.GetByteLength()
	}
	return byteLength
}


