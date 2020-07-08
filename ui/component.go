package ui

// #cgo pkg-config: gdk-3.0 glib-2.0 gobject-2.0
// #include <gdk/gdk.h>
import "C"
import (
	"fmt"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
)

type Direction int
const (
	FORWARD Direction = iota
	BACKWARD
)

// PixbufGetType is a wrapper around gdk_pixbuf_get_type().
func PixbufGetType() glib.Type {
	return glib.Type(C.gdk_pixbuf_get_type())
}


func GetObjectOrPanic(builder *gtk.Builder, name string) glib.IObject {
	obj, err := builder.GetObject(name)
	if err != nil {
		log.Panic("Could not load object ",name, ": ", err)
	}
	return obj
}

func KeyvalName(keyval uint) string {
	return C.GoString((*C.char)(C.gdk_keyval_name(C.guint(keyval))))
}

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

func TopActionsNew(builder *gtk.Builder, sender event.Sender) *TopActionView {
	topActionView := &TopActionView{
		categoriesView:  GetObjectOrPanic(builder, "categories").(*gtk.Box),
		categoryButtons: map[*category.Entry]*CategoryButton{},
		nextButton:      GetObjectOrPanic(builder, "next-button").(*gtk.Button),
		prevButton:      GetObjectOrPanic(builder, "prev-button").(*gtk.Button),
	}
	topActionView.nextButton.Connect("clicked", func() {
		sender.SendToTopic(event.IMAGE_REQUEST_NEXT)
	})
	topActionView.prevButton.Connect("clicked", func() {
		sender.SendToTopic(event.IMAGE_REQUEST_PREV)
	})

	return topActionView
}

func (v *TopActionView) SetVisible(visible bool) {
	v.categoriesView.SetVisible(visible)
	v.nextButton.SetVisible(visible)
	v.prevButton.SetVisible(visible)
}

func (v *TopActionView) FindActionForShortcut(key uint, handle *common.Handle) *category.CategorizeCommand {
	for entry, button := range v.categoryButtons {
		if entry.HasShortcut(key) {
			keyName := KeyvalName(key)
			log.Printf("Key pressed: '%s': '%s'", keyName, entry.GetName())
			return category.CategorizeCommandNew(handle, button.entry, button.operation.NextOperation())
		}
	}
	return nil
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

func ImageViewNew(builder *gtk.Builder, ui *Ui) *ImageView {
	nextImagesList := GetObjectOrPanic(builder, "next-images").(*gtk.TreeView)
	nextImageStore := createImageList(nextImagesList, "Next images", FORWARD, ui.sender)
	prevImagesList := GetObjectOrPanic(builder, "prev-images").(*gtk.TreeView)
	prevImageStore := createImageList(prevImagesList, "Prev images", BACKWARD, ui.sender)
	imageView := &ImageView{
		currentImage: &CurrentImageView{
			scrolledView: GetObjectOrPanic(builder, "current-image-window").(*gtk.ScrolledWindow),
			viewport:     GetObjectOrPanic(builder, "current-image-view").(*gtk.Viewport),
			view:         GetObjectOrPanic(builder, "current-image").(*gtk.Image),
		},
		nextImages: &ImageList{
			component: nextImagesList,
			model:     nextImageStore,
		},
		prevImages: &ImageList{
			component: prevImagesList,
			model:     prevImageStore,
		},
	}
	imageView.currentImage.viewport.Connect("size-allocate", ui.UpdateCurrentImage)

	return imageView
}

type SimilarImagesView struct {
	scrollLayout *gtk.ScrolledWindow
	layout       *gtk.FlowBox
}

func SimilarImagesViewNew(builder *gtk.Builder) *SimilarImagesView{
	layout, _ := gtk.FlowBoxNew()
	similarImagesView := &SimilarImagesView{
		scrollLayout: GetObjectOrPanic(builder, "similar-images-view").(*gtk.ScrolledWindow),
		layout:       layout,
	}

	similarImagesView.layout.SetMaxChildrenPerLine(10)
	similarImagesView.layout.SetRowSpacing(0)
	similarImagesView.layout.SetColumnSpacing(0)
	similarImagesView.layout.SetSizeRequest(-1, 100)
	similarImagesView.scrollLayout.SetVisible(false)
	similarImagesView.scrollLayout.SetSizeRequest(-1, 100)
	similarImagesView.scrollLayout.Add(layout)

	return similarImagesView
}

type BottomActionView struct {
	layout            *gtk.Box
	persistButton     *gtk.Button
	findSimilarButton *gtk.Button
	findDevicesButton *gtk.Button
	editCategoriesButton *gtk.Button
}

func BottomActionsNew(builder *gtk.Builder, ui *Ui, sender event.Sender) *BottomActionView {
	bottomActionView := &BottomActionView{
		layout:            GetObjectOrPanic(builder, "bottom-actions-view").(*gtk.Box),
		persistButton:     GetObjectOrPanic(builder, "persist-button").(*gtk.Button),
		findSimilarButton: GetObjectOrPanic(builder, "find-similar-button").(*gtk.Button),
		findDevicesButton: GetObjectOrPanic(builder, "find-devices-button").(*gtk.Button),
		editCategoriesButton: GetObjectOrPanic(builder, "edit-categories-button").(*gtk.Button),
	}
	bottomActionView.persistButton.Connect("clicked", func() {
		sender.SendToTopic(event.CATEGORY_PERSIST_ALL)
	})

	bottomActionView.findSimilarButton.Connect("clicked", func() {
		sender.SendToTopic(event.SIMILAR_REQUEST_SEARCH)
	})
	bottomActionView.findDevicesButton.Connect("clicked", ui.findDevices)

	bottomActionView.editCategoriesButton.Connect("clicked", ui.showEditCategoriesModal)

	return bottomActionView
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

func ProgressViewNew(builder *gtk.Builder, sender event.Sender) *ProgressView{
	progressView := &ProgressView{
		view:        GetObjectOrPanic(builder, "progress-view").(*gtk.Box),
		stopButton:  GetObjectOrPanic(builder, "stop-progress-button").(*gtk.Button),
		progressbar: GetObjectOrPanic(builder, "progress-bar").(*gtk.ProgressBar),
	}
	progressView.stopButton.Connect("clicked", func() {
		sender.SendToTopic(event.SIMILAR_REQUEST_STOP)
	})

	return progressView
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

func CastModalNew(builder *gtk.Builder, ui *Ui, sender event.Sender) *CastModal {
	modalDialog := GetObjectOrPanic(builder, "cast-dialog").(*gtk.Dialog)
	deviceList := GetObjectOrPanic(builder, "cast-device-list").(*gtk.TreeView)

	cancelButton := GetObjectOrPanic(builder, "cast-dialog-cancel-button").(*gtk.Button)
	cancelButton.Connect("clicked", func() {
		modalDialog.Hide()
	})

	refreshButton := GetObjectOrPanic(builder, "cast-dialog-refresh-button").(*gtk.Button)
	refreshButton.Connect("clicked", ui.findDevices)

	return &CastModal{
		modal:          modalDialog,
		deviceListView: deviceList,
		model:          createDeviceList(modalDialog, deviceList, "Devices", sender),
		cancelButton:   cancelButton,
		refreshButton:  refreshButton,
		statusLabel:    GetObjectOrPanic(builder, "cast-find-status-label").(*gtk.Label),
	}
}

func (s *CastModal) AddDevice(device string) {
	s.deviceListView.SetVisible(true)

	s.statusLabel.SetText("")
	s.statusLabel.SetVisible(false)

	iter := s.model.Append()
	s.model.SetValue(iter, 0, device)
	s.devices = append(s.devices, device)
}

func (s *CastModal) SetNoDevices() {
	s.statusLabel.SetText("No devices found")
	s.statusLabel.SetVisible(true)
}

func (s *CastModal) StartSearch(parent gtk.IWindow) {
	s.devices = []string{}
	s.deviceListView.SetVisible(false)

	s.refreshButton.SetSensitive(false)

	s.statusLabel.SetVisible(true)
	s.statusLabel.SetText("Searching for devices...")

	s.model.Clear()
	s.modal.SetTransientFor(parent)

	s.modal.Show()
}

func (s *CastModal) SearchDone() {
	s.refreshButton.SetSensitive(true)
}




func createImageList(view *gtk.TreeView, title string, direction Direction, sender event.Sender) *gtk.ListStore {
	view.SetSizeRequest(100, -1)
	view.Connect("row-activated", func(view *gtk.TreeView, path *gtk.TreePath, col *gtk.TreeViewColumn) {
		index := path.GetIndices()[0] + 1
		if direction == FORWARD {
			sender.SendToTopicWithData(event.IMAGE_REQUEST_NEXT_OFFSET, index)
		} else {
			sender.SendToTopicWithData(event.IMAGE_REQUEST_PREV_OFFSET, index)
		}
	})
	store, _ := gtk.ListStoreNew(PixbufGetType())
	view.SetModel(store)
	renderer, _ := gtk.CellRendererPixbufNew()
	column, _ := gtk.TreeViewColumnNewWithAttribute(title, renderer, "pixbuf", 0)
	view.AppendColumn(column)
	return store
}

func createDeviceList(modal *gtk.Dialog, view *gtk.TreeView, title string, sender event.Sender) *gtk.ListStore {
	store, _ := gtk.ListStoreNew(glib.TYPE_STRING)
	view.SetSizeRequest(100, -1)
	view.Connect("row-activated", func(view *gtk.TreeView, path *gtk.TreePath, col *gtk.TreeViewColumn) {
		iter, _ := store.GetIter(path)
		value, _ := store.GetValue(iter, 0)
		stringValue, _ := value.GetString()
		sender.SendToTopicWithData(event.CAST_DEVICE_SELECT, stringValue)
		modal.Hide()
	})
	view.SetModel(store)
	renderer, _ := gtk.CellRendererTextNew()
	column, _ := gtk.TreeViewColumnNewWithAttribute(title, renderer, "text", 0)
	view.AppendColumn(column)
	return store
}

type CategoryModal struct {
	modal *gtk.Dialog
	list  *gtk.TreeView
	model *gtk.ListStore

	sender event.Sender

	saveButton *gtk.Button
	saveDefaultButton *gtk.Button
	cancelButton *gtk.Button

	addButton *gtk.Button
	removeButton *gtk.Button
	editButton *gtk.Button
}

func CategoryModalNew(builder *gtk.Builder, ui *Ui, sender event.Sender) *CategoryModal {
	modalDialog := GetObjectOrPanic(builder, "category-dialog").(*gtk.Dialog)
	deviceList := GetObjectOrPanic(builder, "category-list").(*gtk.TreeView)
	model := createCategoryList(modalDialog, deviceList, "Categories", sender)

	saveButton := GetObjectOrPanic(builder, "category-save-button").(*gtk.Button)
	saveDefaultButton := GetObjectOrPanic(builder, "category-save-default-button").(*gtk.Button)

	cancelButton := GetObjectOrPanic(builder, "category-cancel-button").(*gtk.Button)
	cancelButton.Connect("clicked", func() {
		modalDialog.Hide()
	})

	addButton := GetObjectOrPanic(builder, "category-add-button").(*gtk.Button)
	removeButton := GetObjectOrPanic(builder, "category-remove-button").(*gtk.Button)
	editButton := GetObjectOrPanic(builder, "category-edit-button").(*gtk.Button)

	categoryModal := CategoryModal{
		modal:             modalDialog,
		list:              deviceList,
		model:             model,
		sender:            sender,
		saveButton:        saveButton,
		saveDefaultButton: saveDefaultButton,
		cancelButton:      cancelButton,
		addButton:         addButton,
		removeButton:      removeButton,
		editButton:        editButton,
	}

	saveButton.Connect("clicked", categoryModal.save)
	removeButton.Connect("clicked", categoryModal.remove)

	return &categoryModal
}

func (s *CategoryModal) Show(parent gtk.IWindow, categories []*category.Entry) {
	s.model.Clear()
	for _, entry := range categories {
		iter := s.model.Append()
		s.model.SetValue(iter, 0, entry.GetName())
		s.model.SetValue(iter, 1, entry.GetSubPath())
		s.model.SetValue(iter, 2, KeyvalName(entry.GetShortcuts()[0]))
	}

	s.modal.SetTransientFor(parent)

	s.modal.Show()
}

func (s *CategoryModal) save() {
	s.sender.SendToTopicWithData(event.CATEGORIES_SAVE, s.getCategoriesFromList())
	s.modal.Hide()
}
func (s *CategoryModal) saveDefault() {
	s.sender.SendToTopicWithData(event.CATEGORIES_SAVE_DEFAULT, s.getCategoriesFromList())
	s.modal.Hide()
}

func (s *CategoryModal) remove() {
	selection, _ := s.list.GetSelection()
	_, iter, ok := selection.GetSelected()

	if ok {
		s.model.Remove(iter)
	}
}

func (s *CategoryModal) getCategoriesFromList() []*category.Entry {
	var categories []*category.Entry
	for iter, _ := s.model.GetIterFirst(); s.model.IterIsValid(iter); s.model.IterNext(iter) {
		nameValue, _ := s.model.GetValue(iter, 0)
		pathValue, _ := s.model.GetValue(iter, 1)
		keyValue, _ := s.model.GetValue(iter, 2)

		name, _ := nameValue.GetString()
		path, _ := pathValue.GetString()
		key, _ := keyValue.GetString()
		entry := category.CategoryEntryNew(name, path, key)

		categories = append(categories, entry)
	}

	return categories
}

func createCategoryList(modal *gtk.Dialog, view *gtk.TreeView, title string, sender event.Sender) *gtk.ListStore {
	// Name, folder, shortcut key
	store, _ := gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING)

	view.SetModel(store)
	renderer, _ := gtk.CellRendererTextNew()
	nameColumn, _ := gtk.TreeViewColumnNewWithAttribute(title, renderer, "text", 0)
	nameColumn.SetTitle("Name")
	folderColumn, _ := gtk.TreeViewColumnNewWithAttribute(title, renderer, "text", 1)
	folderColumn.SetTitle("Path")
	shortcutColumn, _ := gtk.TreeViewColumnNewWithAttribute(title, renderer, "text", 2)
	shortcutColumn.SetTitle("Shortcut")

	view.AppendColumn(nameColumn)
	view.AppendColumn(folderColumn)
	view.AppendColumn(shortcutColumn)
	return store
}
