package ui

import (
	"github.com/gotk3/gotk3/gtk"
	"vincit.fi/image-sorter/common"
)

type ImageList struct {
	component *gtk.TreeView
	model *gtk.ListStore
}

type CurrentImage struct {
	view *gtk.Image
	image *common.Handle
}
