package ui

import (
	"fmt"
	"github.com/gotk3/gotk3/gtk"
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
)

type ImageList struct {
	component *gtk.TreeView
	model     *gtk.ListStore
}

type TopActionView struct {
	categoriesView  *gtk.Box
	categoryButtons map[*category.Entry]*CategoryButton
	nextButton      *gtk.Button
	prevButton      *gtk.Button
}

func (v *TopActionView) SetVisible(visible bool) {
	v.categoriesView.SetVisible(visible)
	v.nextButton.SetVisible(visible)
	v.prevButton.SetVisible(visible)
}

type CurrentImageView struct {
	scrolledView *gtk.ScrolledWindow
	viewport     *gtk.Viewport
	view         *gtk.Image
	image        *common.Handle
}

type ImageView struct {
	currentImage *CurrentImageView
	nextImages   *ImageList
	prevImages   *ImageList
}

type SimilarImagesView struct {
	scrollLayout *gtk.ScrolledWindow
	layout       *gtk.FlowBox
}

type BottomActionView struct {
	layout            *gtk.Box
	persistButton     *gtk.Button
	findSimilarButton *gtk.Button
	findDevicesButton *gtk.Button
}

func (v *BottomActionView) SetVisible(visible bool) {
	v.layout.SetVisible(visible)
}

type CategoryButton struct {
	button    *gtk.Button
	entry     *category.Entry
	operation category.Operation
}

type ProgressView struct {
	view        *gtk.Box
	progressbar *gtk.ProgressBar
	stopButton  *gtk.Button
}

func (v *ProgressView) SetVisible(visible bool) {
	v.view.SetVisible(visible)
}

func (v *ProgressView) SetStatus(status int, total int) {
	statusText := fmt.Sprintf("Processed %d/%d", status, total)
	v.progressbar.SetText(statusText)
	v.progressbar.SetFraction(float64(status) / float64(total))
}

type CastModal struct {
	modal          *gtk.Dialog
	deviceListView *gtk.TreeView
	model          *gtk.ListStore
	devices        []string
	cancelButton   *gtk.Button
	refreshButton   *gtk.Button
	statusLabel    *gtk.Label
}
