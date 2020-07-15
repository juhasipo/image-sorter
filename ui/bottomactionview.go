package ui

import (
	"github.com/gotk3/gotk3/gtk"
	"vincit.fi/image-sorter/event"
)

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

