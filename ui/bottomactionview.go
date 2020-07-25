package ui

import (
	"github.com/gotk3/gotk3/gtk"
	"vincit.fi/image-sorter/common"
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
		openFolderButton:     GetObjectOrPanic(builder, "open-folder-button").(*gtk.Button),
	}
	bottomActionView.persistButton.Connect("clicked", func() {
		confirm := GetObjectOrPanic(builder, "confirm-categorization-dialog").(*gtk.MessageDialog)
		confirm.SetTransientFor(ui.application.GetActiveWindow())
		confirm.SetTitle("Apply categories?")

		confirmChild := GetObjectOrPanic(builder, "confirm-categorization-dialog-content").(*gtk.Box)
		keepOriginalsCB, _ := gtk.CheckButtonNewWithLabel("Keep old images?")
		keepOriginalsCB.SetActive(true)
		exifCorrect, _ := gtk.CheckButtonNewWithLabel("Rotate image to correct orientation?")

		qualityLayout, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
		qualityLayout.SetHExpand(true)
		adjustment, _ := gtk.AdjustmentNew(90, 0, 100, 0, 0, 0)
		qualityScale, _ := gtk.ScaleNew(gtk.ORIENTATION_HORIZONTAL, adjustment)
		qualityScale.SetProperty("value-pos", gtk.POS_LEFT)
		qualityScale.SetProperty("digits", 0)
		qualityScale.SetHExpand(true)

		qualityLabel, _ := gtk.LabelNew("Quality")
		qualityLayout.Add(qualityLabel)
		qualityLayout.Add(qualityScale)

		confirmChild.Add(keepOriginalsCB)
		confirmChild.Add(exifCorrect)
		confirmChild.Add(qualityLayout)
		confirm.ShowAll()

		defer confirm.Hide()
		response := confirm.Run()
		defer keepOriginalsCB.Destroy()
		defer exifCorrect.Destroy()
		defer qualityLayout.Destroy()

		if response == gtk.RESPONSE_YES {
			value := qualityScale.GetValue()
			sender.SendToTopicWithData(event.CATEGORY_PERSIST_ALL, common.PersistCategorizationCommandNew(
				keepOriginalsCB.GetActive(), exifCorrect.GetActive(), int(value)))
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

func (v *BottomActionView) SetNoDistractionMode(value bool) {
	value = !value
	v.layout.SetVisible(value)
}
