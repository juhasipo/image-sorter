package ui

import (
	"github.com/gotk3/gotk3/gtk"
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
)

type ImageList struct {
	component *gtk.TreeView
	model *gtk.ListStore
}

type TopActionView struct {
	categoriesView *gtk.Box
	categoryButtons map[*category.Entry]*CategoryButton
	nextButton      *gtk.Button
	prevButton      *gtk.Button
}

type CurrentImageView struct {
	scrolledView *gtk.ScrolledWindow
	viewport   *gtk.Viewport
	view *gtk.Image
	image *common.Handle
}

type ImageView struct {
	currentImage       *CurrentImageView
	nextImages         *ImageList
	prevImages         *ImageList
}

type SimilarImagesView struct {
	scrollLayout *gtk.ScrolledWindow
	layout     *gtk.FlowBox
}

type BottomActionView struct {
	persistButton     *gtk.Button
	findSimilarButton *gtk.Button
	findDevicesButton *gtk.Button
}

type CategoryButton struct {
	button    *gtk.Button
	entry     *category.Entry
	operation category.Operation
}

type CastModal struct {
	modal           *gtk.Dialog
	deviceList      *gtk.TreeView
	deviceListStore *gtk.ListStore
	cancelButton    *gtk.Button
}
