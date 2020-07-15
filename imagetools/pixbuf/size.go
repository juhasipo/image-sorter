package pixbuf

import (
	"github.com/gotk3/gotk3/gtk"
)

type Size struct {
	width int
	height int
}

func (s *Size) GetHeight() int {
	return s.height
}

func (s *Size) GetWidth() int {
	return s.width
}

func SizeFromWindow(widget *gtk.ScrolledWindow) Size {
	return Size {
		width: widget.GetAllocatedWidth(),
		height: widget.GetAllocatedHeight(),
	}
}
