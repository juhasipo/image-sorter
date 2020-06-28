package ui

// #cgo pkg-config: gdk-3.0 glib-2.0 gobject-2.0
// #include <gdk/gdk.h>
import "C"
import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"sync"
	"vincit.fi/image-sorter/common"
)

type Size struct {
	width int
	height int
}

func SizeFromViewport(widget *gtk.Viewport) Size {
	return Size {
		width: widget.GetAllocatedWidth(),
		height: widget.GetAllocatedHeight(),
	}
}
func SizeFromWidget(widget *gtk.Widget) Size {
	return Size {
		width: widget.GetAllocatedWidth(),
		height: widget.GetAllocatedHeight(),
	}
}

func SizeFromInt(width int, height int) Size {
	return Size {
		width: width,
		height: height,
	}
}

// PixbufGetType is a wrapper around gdk_pixbuf_get_type().
func PixbufGetType() glib.Type {
	return glib.Type(C.gdk_pixbuf_get_type())
}

type Instance struct {
	handle *common.Handle
	full *gdk.Pixbuf
	thumbnail *gdk.Pixbuf
	scaled *gdk.Pixbuf
	loader *PixbufCache
}

func (s *Instance) IsValid() bool {
	return s.handle != nil
}

var (
	EMPTY_INSTANCE = Instance {}
)

func (s* Instance) GetScaled(size Size) *gdk.Pixbuf {
	if !s.IsValid() {
		log.Print("Empty instance")
		return nil
	}

	if s.full == nil {
		log.Print(" * Loading full image...")
		s.full, _ = s.loader.loadFromHandle(s.handle)
	}

	ratio := float32(s.full.GetWidth()) / float32(s.full.GetHeight())
	newWidth := int(float32(size.height) * ratio)
	newHeight := size.height

	if newWidth > size.width {
		newWidth = size.width
		newHeight = int(float32(size.width) / ratio)
	}

	if s.scaled == nil {
		log.Print(" * Loading new scaled ", s.handle, " (", newWidth, " x ", newHeight, ")...")
		s.scaled, _ = s.full.ScaleSimple(newWidth, newHeight, gdk.INTERP_TILES)
	} else {
		if newWidth != s.scaled.GetWidth() && newHeight != s.scaled.GetHeight() {
			log.Print(" * Loading re-scaled ", s.handle,
				" (", s.scaled.GetWidth(), " x ", s.scaled.GetHeight(), ") -> ",
				" (", newWidth, " x ", newHeight, ")...")
			s.scaled, _ = s.full.ScaleSimple(newWidth, newHeight, gdk.INTERP_TILES)
		} else {
			log.Print(" * Use cached")
		}
	}

	return s.scaled
}

func (s* Instance) GetThumbnail() *gdk.Pixbuf {
	if s.handle == nil {
		log.Print("Nil handle")
		return nil
	}
	if s.full == nil {
		//log.Print(" * Loading full image...")
		s.full, _ = s.loader.loadFromHandle(s.handle)
	}
	if s.thumbnail == nil {
		width, height := 100, 100
		ratio := float32(s.full.GetWidth()) / float32(s.full.GetHeight())
		newWidth := int(float32(height) * ratio)
		newHeight := height

		if newWidth > width {
			newWidth = width
			newHeight = int(float32(width) / ratio)
		}

		s.thumbnail, _ = s.full.ScaleSimple(newWidth, newHeight, gdk.INTERP_TILES)
	}
	return s.thumbnail
}

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
