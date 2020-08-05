package ui

import (
	"github.com/gotk3/gotk3/gtk"
	"image"
	"vincit.fi/image-sorter/common"
)

type CurrentImageView struct {
	scrolledView *gtk.ScrolledWindow
	viewport     *gtk.Viewport
	view         *gtk.Image
	image        *common.Handle
	details      *gtk.TextView
	img          image.Image
}
