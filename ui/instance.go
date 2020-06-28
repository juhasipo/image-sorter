package ui

import (
	"github.com/gotk3/gotk3/gdk"
	"log"
	"vincit.fi/image-sorter/common"
)

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
