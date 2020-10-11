package component

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"strings"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common"
)

type CategoryModal struct {
	modal *gtk.Dialog
	list  *gtk.TreeView
	model *gtk.ListStore

	sender api.Sender

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

func NewCategoryModal(builder *gtk.Builder, sender api.Sender) *CategoryModal {
	modalDialog := GetObjectOrPanic(builder, "category-dialog").(*gtk.Dialog)
	modalDialog.SetSizeRequest(400, 300)
	deviceList := GetObjectOrPanic(builder, "category-list").(*gtk.TreeView)
	model := createCategoryList(deviceList, "Categories")

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
	} else {
		s.pathEntry.SetText("")
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

func (s *CategoryModal) Show(parent gtk.IWindow, categories []*apitype.Category) {
	s.list.Show()
	s.addEntryView.Hide()

	s.model.Clear()
	for _, entry := range categories {
		iter := s.model.Append()
		s.model.SetValue(iter, 0, entry.GetName())
		s.model.SetValue(iter, 1, entry.GetSubPath())
		s.model.SetValue(iter, 2, entry.GetShortcutAsString())
	}

	s.modal.SetTransientFor(parent)

	s.modal.Show()
}

func (s *CategoryModal) save() {
	s.sender.SendToTopicWithData(api.CategoriesSave, s.getCategoriesFromList())
	s.modal.Hide()
}
func (s *CategoryModal) saveDefault() {
	s.sender.SendToTopicWithData(api.CategoriesSaveDefault, s.getCategoriesFromList())
	s.modal.Hide()
}

func (s *CategoryModal) remove() {
	selection, _ := s.list.GetSelection()
	_, iter, ok := selection.GetSelected()

	if ok {
		s.model.Remove(iter)
	}
}

func (s *CategoryModal) getCategoriesFromList() []*apitype.Category {
	var categories []*apitype.Category
	for iter, _ := s.model.GetIterFirst(); s.model.IterIsValid(iter); s.model.IterNext(iter) {
		name, path, key := extractValuesFromModel(s.model, iter)
		entry := apitype.NewCategory(name, path, key)

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

func createCategoryList(view *gtk.TreeView, title string) *gtk.ListStore {
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
