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

func SizeFromViewport(widget *gtk.Viewport) Size {
	return Size {
		width: widget.GetAllocatedWidth(),
		height: widget.GetAllocatedHeight(),
	}
}
func SizeFromWindow(widget *gtk.ScrolledWindow) Size {
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
