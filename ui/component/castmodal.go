package component

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"vincit.fi/image-sorter/event"
)

type CastModal struct {
	modal            *gtk.Dialog
	deviceListView   *gtk.TreeView
	model            *gtk.ListStore
	devices          []string
	cancelButton     *gtk.Button
	refreshButton    *gtk.Button
	statusLabel      *gtk.Label
	showBackgroundCB *gtk.CheckButton
}

func NewCastModal(builder *gtk.Builder, ui CallbackApi, sender event.Sender) *CastModal {
	modalDialog := GetObjectOrPanic(builder, "cast-dialog").(*gtk.Dialog)
	deviceList := GetObjectOrPanic(builder, "cast-device-list").(*gtk.TreeView)

	cancelButton := GetObjectOrPanic(builder, "cast-dialog-cancel-button").(*gtk.Button)
	cancelButton.Connect("clicked", func() {
		modalDialog.Hide()
	})

	refreshButton := GetObjectOrPanic(builder, "cast-dialog-refresh-button").(*gtk.Button)
	refreshButton.Connect("clicked", ui.FindDevices)

	castModal := &CastModal{
		modal:            modalDialog,
		deviceListView:   deviceList,
		cancelButton:     cancelButton,
		refreshButton:    refreshButton,
		statusLabel:      GetObjectOrPanic(builder, "cast-find-status-label").(*gtk.Label),
		showBackgroundCB: GetObjectOrPanic(builder, "caster-show-background-checkbox").(*gtk.CheckButton),
	}
	castModal.model = createDeviceList(castModal, modalDialog, deviceList, "Devices", sender)

	return castModal
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

func (s *CastModal) GetDevices() []string {
	return s.devices
}

func createDeviceList(castModal *CastModal, modal *gtk.Dialog, view *gtk.TreeView, title string, sender event.Sender) *gtk.ListStore {
	store, _ := gtk.ListStoreNew(glib.TYPE_STRING)
	view.SetSizeRequest(100, -1)
	view.Connect("row-activated", func(view *gtk.TreeView, path *gtk.TreePath, col *gtk.TreeViewColumn) {
		iter, _ := store.GetIter(path)
		value, _ := store.GetValue(iter, 0)
		stringValue, _ := value.GetString()
		sender.SendToTopicWithData(event.CastDeviceSelect, stringValue, castModal.showBackgroundCB.GetActive())
		modal.Hide()
	})
	view.SetModel(store)
	renderer, _ := gtk.CellRendererTextNew()
	column, _ := gtk.TreeViewColumnNewWithAttribute(title, renderer, "text", 0)
	view.AppendColumn(column)
	return store
}
