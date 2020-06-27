package image

import (
	"github.com/gotk3/gotk3/gdk"
	"log"
	"os"
	"path/filepath"
)

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
	log.Print("Loading scaled ", s.handle, " (", width, " x ", height, ")...")
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

	s.scaled, _ = s.full.ScaleSimple(newWidth, newHeight, gdk.INTERP_TILES)

	log.Print(" * Scaled to ", newWidth, " x ", newHeight, ")")
	return s.scaled
}

type Loader struct {
}

func (s*Loader) loadFromFile(path string) (*gdk.Pixbuf, error) {
	return gdk.PixbufNewFromFile(path)
}

type Manager struct {
	imageList []Handle
	imageCache map[Handle]Instance
	index int
}

func ManagerForDir(dir string) Manager {
	var manager = Manager{
		imageList: []Handle {},
		imageCache: map[Handle]Instance{},
		index: 0,
	}
	loader := Loader{}
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) != ".jpg" {
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

func (s *Manager) GetScaled(handle *Handle, width int, height int) *gdk.Pixbuf {
	image := s.imageCache[*handle]
	return image.GetScaled(width, height)
}

func (s *Manager) AddImage(h Handle, instance Instance) {
	s.imageList = append(s.imageList, h)
	s.imageCache[h] = instance
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
