package ui

// #cgo pkg-config: gdk-3.0 glib-2.0 gobject-2.0
// #include <gdk/gdk.h>
import "C"
import (
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"strings"
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
		log.Panic("Could not load object ", name, ": ", err)
	}
	return obj
}

type ImageList struct {
	component *gtk.TreeView
	model     *gtk.ListStore
}

type TopActionView struct {
	categoriesView  *gtk.Box
	categoryButtons map[string]*CategoryButton
	nextButton      *gtk.Button
	prevButton      *gtk.Button
}

func TopActionsNew(builder *gtk.Builder, sender event.Sender) *TopActionView {
	topActionView := &TopActionView{
		categoriesView:  GetObjectOrPanic(builder, "categories").(*gtk.Box),
		categoryButtons: map[string]*CategoryButton{},
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
	for _, button := range v.categoryButtons {
		entry := button.entry
		if entry.HasShortcut(key) {
			keyName := common.KeyvalName(key)
			log.Printf("Key pressed: '%s': '%s'", keyName, entry.GetName())
			stayOnSameImage := gdk.KeyvalIsUpper(key)
			return category.CategorizeCommandNewWithStayAttr(
				handle, button.entry, button.operation.NextOperation(), stayOnSameImage)
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
	imageView.currentImage.viewport.Connect("size-allocate", func() {
		ui.UpdateCurrentImage()
		height := ui.imageView.nextImages.component.GetAllocatedHeight() / 80
		ui.sender.SendToTopicWithData(event.IMAGE_LIST_SIZE_CHANGED, height)
	})

	return imageView
}

type SimilarImagesView struct {
	scrollLayout *gtk.ScrolledWindow
	layout       *gtk.FlowBox
}

func SimilarImagesViewNew(builder *gtk.Builder) *SimilarImagesView {
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
	layout               *gtk.Box
	persistButton        *gtk.Button
	findSimilarButton    *gtk.Button
	findDevicesButton    *gtk.Button
	editCategoriesButton *gtk.Button
	openFolderButton     *gtk.Button
}

func BottomActionsNew(builder *gtk.Builder, ui *Ui, sender event.Sender) *BottomActionView {
	bottomActionView := &BottomActionView{
		layout:               GetObjectOrPanic(builder, "bottom-actions-view").(*gtk.Box),
		persistButton:        GetObjectOrPanic(builder, "persist-button").(*gtk.Button),
		findSimilarButton:    GetObjectOrPanic(builder, "find-similar-button").(*gtk.Button),
		findDevicesButton:    GetObjectOrPanic(builder, "find-devices-button").(*gtk.Button),
		editCategoriesButton: GetObjectOrPanic(builder, "edit-categories-button").(*gtk.Button),
		openFolderButton: GetObjectOrPanic(builder, "open-folder-button").(*gtk.Button),
	}
	bottomActionView.persistButton.Connect("clicked", func() {
		confirm := gtk.MessageDialogNew(ui.application.GetActiveWindow(), gtk.DIALOG_MODAL, gtk.MESSAGE_QUESTION, gtk.BUTTONS_YES_NO, "Do you really want to move images to cateogry folders?")
		defer confirm.Destroy()
		response := confirm.Run()

		if response == gtk.RESPONSE_YES {
			sender.SendToTopic(event.CATEGORY_PERSIST_ALL)
		}
	})

	bottomActionView.findSimilarButton.Connect("clicked", func() {
		sender.SendToTopic(event.SIMILAR_REQUEST_SEARCH)
	})
	bottomActionView.findDevicesButton.Connect("clicked", ui.findDevices)

	bottomActionView.editCategoriesButton.Connect("clicked", ui.showEditCategoriesModal)

	bottomActionView.openFolderButton.Connect("clicked", func() {
		ui.openFolderChooser(2)
	})

	return bottomActionView
}

func (v *BottomActionView) SetVisible(visible bool) {
	v.layout.SetVisible(visible)
}

type CategoryButton struct {
	button          *gtk.Button
	entry           *common.Category
	operation       common.Operation
	categorizedIcon *gtk.Image
}

func (s *CategoryButton) SetStatus(operation common.Operation) {
	if operation == common.MOVE {
		s.button.SetImage(s.categorizedIcon)
	} else {
		icon, _ := gtk.ImageNew()
		s.button.SetImage(icon)
	}
}

type ProgressView struct {
	view        *gtk.Box
	progressbar *gtk.ProgressBar
	stopButton  *gtk.Button
}

func ProgressViewNew(builder *gtk.Builder, sender event.Sender) *ProgressView {
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
	refreshButton  *gtk.Button
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

	saveButton        *gtk.Button
	saveMenuButton    *gtk.MenuButton
	saveDefaultButton *gtk.Button
	cancelButton      *gtk.Button

	categoryListActions *gtk.Box
	addButton           *gtk.Button
	removeButton        *gtk.Button
	editButton          *gtk.Button

	addEntryView       *gtk.Grid
	nameEntry          *gtk.Entry
	pathEntry          *gtk.Entry
	shortcutComboBox   *gtk.ComboBoxText
	addAddButton       *gtk.Button
	addEditButton      *gtk.Button
	editedCategoryIter *gtk.TreeIter
	addCancelButton    *gtk.Button
}

func CategoryModalNew(builder *gtk.Builder, ui *Ui, sender event.Sender) *CategoryModal {
	modalDialog := GetObjectOrPanic(builder, "category-dialog").(*gtk.Dialog)
	modalDialog.SetSizeRequest(400, 300)
	deviceList := GetObjectOrPanic(builder, "category-list").(*gtk.TreeView)
	model := createCategoryList(modalDialog, deviceList, "Categories", sender)

	saveButton := GetObjectOrPanic(builder, "category-save-button").(*gtk.Button)
	saveMenuButton := GetObjectOrPanic(builder, "category-save-menu-button").(*gtk.MenuButton)
	saveDefaultButton := GetObjectOrPanic(builder, "category-save-default-button").(*gtk.Button)

	cancelButton := GetObjectOrPanic(builder, "category-cancel-button").(*gtk.Button)
	cancelButton.Connect("clicked", func() {
		modalDialog.Hide()
	})

	categoryListActions := GetObjectOrPanic(builder, "category-list-actions").(*gtk.Box)
	addButton := GetObjectOrPanic(builder, "category-add-button").(*gtk.Button)
	removeButton := GetObjectOrPanic(builder, "category-remove-button").(*gtk.Button)
	editButton := GetObjectOrPanic(builder, "category-edit-button").(*gtk.Button)

	addEntryView := GetObjectOrPanic(builder, "category-add-view").(*gtk.Grid)
	nameEntry := GetObjectOrPanic(builder, "category-add-name").(*gtk.Entry)
	pathEntry := GetObjectOrPanic(builder, "category-add-path").(*gtk.Entry)
	shortcutComboBox := GetObjectOrPanic(builder, "category-add-shortcut").(*gtk.ComboBoxText)
	addAddButton := GetObjectOrPanic(builder, "category-add-add-button").(*gtk.Button)
	addEditButton := GetObjectOrPanic(builder, "category-add-edit-button").(*gtk.Button)
	addCancelButton := GetObjectOrPanic(builder, "category-add-cancel-button").(*gtk.Button)

	categoryModal := CategoryModal{
		modal:               modalDialog,
		list:                deviceList,
		model:               model,
		sender:              sender,
		saveButton:          saveButton,
		saveMenuButton:      saveMenuButton,
		saveDefaultButton:   saveDefaultButton,
		cancelButton:        cancelButton,
		categoryListActions: categoryListActions,
		addButton:           addButton,
		removeButton:        removeButton,
		editButton:          editButton,
		addEntryView:        addEntryView,
		nameEntry:           nameEntry,
		pathEntry:           pathEntry,
		shortcutComboBox:    shortcutComboBox,
		addAddButton:        addAddButton,
		addEditButton:       addEditButton,
		addCancelButton:     addCancelButton,
	}

	nameEntry.Connect("changed", func(entry *gtk.Entry) {
		value, _ := entry.GetText()
		if strings.TrimSpace(value) == "" {
			categoryModal.addAddButton.SetSensitive(false)
			categoryModal.addEditButton.SetSensitive(false)
		} else {
			categoryModal.addAddButton.SetSensitive(true)
			categoryModal.addEditButton.SetSensitive(true)
		}
	})
	saveButton.Connect("clicked", categoryModal.save)
	saveDefaultButton.Connect("clicked", categoryModal.saveDefault)
	removeButton.Connect("clicked", categoryModal.remove)
	addButton.Connect("clicked", categoryModal.startAdd)
	editButton.Connect("clicked", categoryModal.startEdit)

	addAddButton.Connect("clicked", categoryModal.addNewCategory)
	addEditButton.Connect("clicked", categoryModal.editCategory)
	addCancelButton.Connect("clicked", categoryModal.endAddOrEdit)

	return &categoryModal
}

func (s *CategoryModal) startEdit() {
	s.addAddButton.Hide()
	s.addEditButton.Show()

	selection, _ := s.list.GetSelection()
	_, iter, ok := selection.GetSelected()
	s.editedCategoryIter = iter

	if ok {
		name, path, key := extractValuesFromModel(s.model, s.editedCategoryIter)

		s.initAddEditView(name, path, key)
	}
}

func (s *CategoryModal) initAddEditView(name string, path string, key string) {
	s.nameEntry.SetText(name)
	if name != path {
		s.pathEntry.SetText(path)
	}
	s.addKeySelections()

	if key == "" {
		s.shortcutComboBox.SetActive(0)
	} else {
		model, _ := s.shortcutComboBox.GetModel()
		keyIndex := findKeyIndex(key, model)
		s.shortcutComboBox.SetActive(keyIndex)
	}

	s.saveButton.Hide()
	s.cancelButton.Hide()
	s.saveMenuButton.Hide()
	s.categoryListActions.Hide()
	s.list.Hide()
	s.addEntryView.Show()
	s.addEntryView.Show()
	s.addButton.SetSensitive(false)
	s.removeButton.SetSensitive(false)
	s.editButton.SetSensitive(false)
}

func findKeyIndex(key string, model *gtk.TreeModel) int {
	upperKey := strings.ToUpper(key)

	i := 0
	iter, _ := model.GetIterFirst()
	for {
		value, _ := model.GetValue(iter, 0)

		iterValueString, _ := value.GetString()
		if strings.ToUpper(iterValueString) == upperKey {
			return i
		}
		if model.IterNext(iter) {
			i += 1
		} else {
			return 0
		}
	}
}

func (s *CategoryModal) startAdd() {
	s.addAddButton.Show()
	s.addEditButton.Hide()

	s.initAddEditView("", "", "")
}

func (s *CategoryModal) addKeySelections() {
	store, _ := gtk.ListStoreNew(glib.TYPE_STRING)
	s.shortcutComboBox.SetModel(store)
	for _, key := range KEYS {
		isInUse := false
		if !isInUse {
			s.shortcutComboBox.AppendText(common.KeyvalName(key))
		}
	}
}
func (s *CategoryModal) addNewCategory() {
	iter := s.model.Append()

	s.applyToIter(iter)

	s.endAddOrEdit()
}

func (s *CategoryModal) editCategory() {
	s.applyToIter(s.editedCategoryIter)
	s.endAddOrEdit()
}

func (s *CategoryModal) applyToIter(iter *gtk.TreeIter) {
	name, _ := s.nameEntry.GetText()
	path, _ := s.pathEntry.GetText()
	if path == "" {
		path = name
	}
	shortcut := s.shortcutComboBox.GetActiveText()

	s.model.SetValue(iter, 0, name)
	s.model.SetValue(iter, 1, path)
	s.model.SetValue(iter, 2, shortcut)
}

func (s *CategoryModal) endAddOrEdit() {
	s.saveButton.Show()
	s.cancelButton.Show()
	s.saveMenuButton.Show()
	s.categoryListActions.Show()
	s.list.Show()
	s.addEntryView.Hide()

	s.addButton.SetSensitive(true)
	s.removeButton.SetSensitive(true)
	s.editButton.SetSensitive(true)
}

func (s *CategoryModal) Show(parent gtk.IWindow, categories []*common.Category) {
	s.list.Show()
	s.addEntryView.Hide()

	s.model.Clear()
	for _, entry := range categories {
		iter := s.model.Append()
		s.model.SetValue(iter, 0, entry.GetName())
		s.model.SetValue(iter, 1, entry.GetSubPath())
		s.model.SetValue(iter, 2, common.KeyvalName(entry.GetShortcuts()[0]))
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

func (s *CategoryModal) getCategoriesFromList() []*common.Category {
	var categories []*common.Category
	for iter, _ := s.model.GetIterFirst(); s.model.IterIsValid(iter); s.model.IterNext(iter) {
		name, path, key := extractValuesFromModel(s.model, iter)
		entry := common.CategoryEntryNew(name, path, key)

		categories = append(categories, entry)
	}

	return categories
}

func extractValuesFromModel(store *gtk.ListStore, iter *gtk.TreeIter) (string, string, string) {
	nameValue, _ := store.GetValue(iter, 0)
	pathValue, _ := store.GetValue(iter, 1)
	keyValue, _ := store.GetValue(iter, 2)
	name, _ := nameValue.GetString()
	path, _ := pathValue.GetString()
	shortcut, _ := keyValue.GetString()
	return name, path, shortcut
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
