package ui

import "github.com/gotk3/gotk3/gtk"

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
