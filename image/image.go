package image

// #cgo pkg-config: gdk-3.0 glib-2.0 gobject-2.0
// #include <gdk/gdk.h>
import "C"
import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// PixbufGetType is a wrapper around gdk_pixbuf_get_type().
func PixbufGetType() glib.Type {
	return glib.Type(C.gdk_pixbuf_get_type())
}

type Handle struct {
	id string
}

type Instance struct {
	path string
	handle *Handle
	full *gdk.Pixbuf
	thumbnail *gdk.Pixbuf
	scaled *gdk.Pixbuf
	loader *Loader
}

func (s* Instance) GetScaled(width int, height int) *gdk.Pixbuf {
	if s.handle == nil {
		log.Print("Nil handle")
		return nil
	}
	if s.full == nil {
		log.Print(" * Loading full image...")
		s.full, _ = s.loader.loadFromFile(s.path)
	}

	ratio := float32(s.full.GetWidth()) / float32(s.full.GetHeight())
	newWidth := int(float32(height) * ratio)
	newHeight := height

	if newWidth > width {
		newWidth = width
		newHeight = int(float32(width) / ratio)
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
		s.full, _ = s.loader.loadFromFile(s.path)
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

type Loader struct {
}

func (s*Loader) loadFromFile(path string) (*gdk.Pixbuf, error) {
	return gdk.PixbufNewFromFile(path)
}

type Manager struct {
	imageList []Handle
	imageCache map[Handle]*Instance
	index int
}

func ManagerForDir(dir string) Manager {
	var manager = Manager{
		imageList: []Handle {},
		imageCache: map[Handle]*Instance{},
		index: 0,
	}
	loader := Loader{}
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(strings.ToLower(path)) != ".jpg" {
			return nil
		}
		handle := Handle {id: path}
		instance := Instance{
			path:      path,
			handle:    &handle,
			full:      nil,
			thumbnail: nil,
			scaled:    nil,
			loader:    &loader,
		}
		manager.AddImage(handle, instance)
		return nil
	})

	return manager
}

func (s *Manager) NextImage() *Handle {
	s.index++
	if s.index >= len(s.imageList) {
		s.index = len(s.imageList) - 1
	}
	return s.GetCurrentImage()
}

func (s *Manager) PrevImage() *Handle {
	s.index--
	if s.index < 0 {
		s.index = 0
	}
	return s.GetCurrentImage()
}

func (s *Manager) GetCurrentImage() *Handle {
	handle := s.imageList[s.index]
	return &handle
}

var (
	EMPTY_HANDLE []Handle
)

func (s* Manager) GetNextImages(number int) []Handle {
	startIndex := s.index + 1
	endIndex := startIndex + number
	if endIndex > len(s.imageList) {
		endIndex = len(s.imageList)
	}

	if startIndex >= len(s.imageList) - 1 {
		return EMPTY_HANDLE
	}

	return s.imageList[startIndex:endIndex]
}

func (s* Manager) GetPrevImages(number int) []Handle {
        prevIndex := s.index-number
        if prevIndex < 0 {
            prevIndex = 0
        }
        return s.imageList[prevIndex:s.index]
    }

func (s *Manager) GetScaled(handle *Handle, width int, height int) *gdk.Pixbuf {
	image := s.imageCache[*handle]
	return image.GetScaled(width, height)
}
func (s *Manager) GetThumbnail(handle *Handle) *gdk.Pixbuf {
	image := s.imageCache[*handle]
	return image.GetThumbnail()
}

func (s *Manager) AddImage(h Handle, instance Instance) {
	s.imageList = append(s.imageList, h)
	s.imageCache[h] = &instance
}

type CategoryOperation int
const(
	COPY CategoryOperation = 0
	MOVE CategoryOperation = 1
)

type Category struct {
	name string
	subPath string
}

type CategorizedImage struct {
	category Category
	operation CategoryOperation
}

type CategoryManager struct {
	imageCategory map[Handle]CategorizedImage
}
