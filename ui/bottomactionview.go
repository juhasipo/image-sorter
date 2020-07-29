package ui

import (
	"github.com/gotk3/gotk3/gdk"
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
	fullscreenButton     *gtk.Button
	noDistractionsButton *gtk.Button
	exitFullscreenButton *gtk.Button
	showAllImagesButton  *gtk.Button
}

func BottomActionsNew(builder *gtk.Builder, ui *Ui, sender event.Sender) *BottomActionView {
	bottomActionView := &BottomActionView{
		layout:               GetObjectOrPanic(builder, "bottom-actions-view").(*gtk.Box),
		persistButton:        GetObjectOrPanic(builder, "persist-button").(*gtk.Button),
		findSimilarButton:    GetObjectOrPanic(builder, "find-similar-button").(*gtk.Button),
		findDevicesButton:    GetObjectOrPanic(builder, "find-devices-button").(*gtk.Button),
		editCategoriesButton: GetObjectOrPanic(builder, "edit-categories-button").(*gtk.Button),
		openFolderButton:     GetObjectOrPanic(builder, "open-folder-button").(*gtk.Button),
		fullscreenButton:     GetObjectOrPanic(builder, "fullscreen-button").(*gtk.Button),
		noDistractionsButton: GetObjectOrPanic(builder, "no-distractions-button").(*gtk.Button),
		exitFullscreenButton: GetObjectOrPanic(builder, "exit-fullscreen-button").(*gtk.Button),
		showAllImagesButton:  GetObjectOrPanic(builder, "show-all-images-button").(*gtk.Button),
	}
	bottomActionView.exitFullscreenButton.SetVisible(false)

	bottomActionView.persistButton.Connect("clicked", func() {
		confirm := GetObjectOrPanic(builder, "confirm-categorization-dialog").(*gtk.MessageDialog)
		confirm.SetTransientFor(ui.application.GetActiveWindow())
		confirm.SetTitle("Apply categories?")

		confirmChild := GetObjectOrPanic(builder, "confirm-categorization-dialog-content").(*gtk.Box)

		keepOriginalsLayout, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
		keepOriginalsLayout.SetMarginBottom(5)
		keepOriginalsLabel, _ := gtk.LabelNew("Keep old images?")
		keepOriginalsCB, _ := gtk.SwitchNew()
		keepOriginalsCB.SetActive(true)
		keepOriginalsLayout.Add(keepOriginalsCB)
		keepOriginalsLayout.Add(keepOriginalsLabel)

		exifCorrectLayout, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
		exifCorrectLayout.SetMarginBottom(5)
		exifCorrectLabel, _ := gtk.LabelNew("Rotate image to correct orientation?")
		exifCorrect, _ := gtk.SwitchNew()
		exifCorrectLayout.Add(exifCorrect)
		exifCorrectLayout.Add(exifCorrectLabel)

		qualityLayout, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
		qualityLayout.SetHExpand(false)
		adjustment, _ := gtk.AdjustmentNew(90, 0, 100, 0, 0, 0)
		qualityScale, _ := gtk.ScaleNew(gtk.ORIENTATION_HORIZONTAL, adjustment)
		qualityScale.SetProperty("value-pos", gtk.POS_LEFT)
		qualityScale.SetProperty("digits", 0)
		qualityScale.SetHExpand(true)

		qualityLabel, _ := gtk.LabelNew("Quality")
		qualityLayout.Add(qualityLabel)
		qualityLayout.Add(qualityScale)

		confirmChild.Add(keepOriginalsLayout)
		confirmChild.Add(exifCorrectLayout)
		confirmChild.Add(qualityLayout)
		confirm.ShowAll()

		defer confirm.Hide()
		response := confirm.Run()
		defer keepOriginalsLayout.Destroy()
		defer exifCorrectLayout.Destroy()
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
	bottomActionView.fullscreenButton.Connect("button-release-event", func(button *gtk.Button, e *gdk.Event) bool {
		keyEvent := gdk.EventButtonNewFromEvent(e)

		modifiers := gtk.AcceleratorGetDefaultModMask()
		state := gdk.ModifierType(keyEvent.State())

		controlDown := state&modifiers&gdk.GDK_CONTROL_MASK > 0
		if controlDown {
			ui.enterFullScreenNoDistraction()
		} else {
			ui.enterFullScreen()
		}
		return true
	})
	bottomActionView.fullscreenButton.Connect("key-press-event", func(button *gtk.Button, e *gdk.Event) bool {
		keyEvent := gdk.EventKeyNewFromEvent(e)
		key := keyEvent.KeyVal()

		if key == gdk.KEY_KP_Enter || key == gdk.KEY_Return || key == gdk.KEY_KP_Space || key == gdk.KEY_space {
			modifiers := gtk.AcceleratorGetDefaultModMask()
			state := gdk.ModifierType(keyEvent.State())
			controlDown := state&modifiers&gdk.GDK_CONTROL_MASK > 0
			if controlDown {
				ui.enterFullScreenNoDistraction()
			} else {
				ui.enterFullScreen()
			}
			return true
		}
		return false
	})
	bottomActionView.noDistractionsButton.Connect("clicked", func() {
		ui.enterFullScreenNoDistraction()
	})
	bottomActionView.exitFullscreenButton.Connect("clicked", func() {
		ui.exitFullScreen()
	})

	bottomActionView.showAllImagesButton.Connect("clicked", func() {
		sender.SendToTopic(event.IMAGE_SHOW_ALL)
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

func (v *BottomActionView) SetShowFullscreenButton(visible bool) {
	v.fullscreenButton.SetVisible(visible)
	v.exitFullscreenButton.SetVisible(!visible)
}

func (v *BottomActionView) SetShowOnlyCategory(status bool) {
	v.showAllImagesButton.SetVisible(status)
}
