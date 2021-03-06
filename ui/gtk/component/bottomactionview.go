package component

import (
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/common"
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
	aboutButton          *gtk.Button
}

func NewBottomActions(builder *gtk.Builder, appWindow *gtk.Window, callbacks CallbackApi, sender api.Sender) *BottomActionView {
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
		aboutButton:          GetObjectOrPanic(builder, "about-button").(*gtk.Button),
	}
	bottomActionView.exitFullscreenButton.SetVisible(false)

	bottomActionView.persistButton.Connect("clicked", func() {
		confirm := GetObjectOrPanic(builder, "confirm-categorization-dialog").(*gtk.MessageDialog)
		confirm.SetTransientFor(appWindow)
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
			sender.SendCommandToTopic(api.CategoryPersistAll, &api.PersistCategorizationCommand{
				KeepOriginals:  keepOriginalsCB.GetActive(),
				FixOrientation: exifCorrect.GetActive(),
				Quality:        int(value),
			})
		}
	})

	bottomActionView.findSimilarButton.Connect("clicked", func() {
		sender.SendToTopic(api.SimilarRequestSearch)
	})
	bottomActionView.findDevicesButton.Connect("clicked", callbacks.FindDevices)

	bottomActionView.editCategoriesButton.Connect("clicked", callbacks.ShowEditCategoriesModal)

	bottomActionView.openFolderButton.Connect("clicked", func() {
		callbacks.OpenFolderChooser(2)
	})
	bottomActionView.fullscreenButton.Connect("button-release-event", func(button *gtk.Button, e *gdk.Event) bool {
		keyEvent := gdk.EventButtonNewFromEvent(e)

		modifiers := gtk.AcceleratorGetDefaultModMask()
		state := gdk.ModifierType(keyEvent.State())

		controlDown := state&modifiers&gdk.GDK_CONTROL_MASK > 0
		if controlDown {
			callbacks.EnterFullScreenNoDistraction()
		} else {
			callbacks.EnterFullScreen()
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
				callbacks.EnterFullScreenNoDistraction()
			} else {
				callbacks.EnterFullScreen()
			}
			return true
		}
		return false
	})
	bottomActionView.noDistractionsButton.Connect("clicked", func() {
		callbacks.EnterFullScreenNoDistraction()
	})
	bottomActionView.exitFullscreenButton.Connect("clicked", func() {
		callbacks.ExitFullScreen()
	})

	bottomActionView.showAllImagesButton.Connect("clicked", func() {
		sender.SendToTopic(api.ImageShowAll)
	})

	bottomActionView.aboutButton.Connect("clicked", func() {
		logo, _ := gdk.PixbufNewFromFile("assets/icon-128x128.png")

		aboutDialog, _ := gtk.AboutDialogNew()
		aboutDialog.SetLogo(logo)
		_ = aboutDialog.SetIconFromFile("assets/icon-32x32.png")
		aboutDialog.SetAuthors([]string{common.Author})
		aboutDialog.SetName(common.AppName)
		aboutDialog.SetProgramName(common.AppName)
		aboutDialog.SetWebsite(common.WebSiteUrl)
		aboutDialog.SetVersion(common.Version)
		aboutDialog.SetCopyright(fmt.Sprintf("Copyright ©️ %s %s", common.Author, common.CopyrightYear))

		aboutDialog.Run()
		aboutDialog.Destroy()
	})

	return bottomActionView
}

func (s *BottomActionView) SetVisible(visible bool) {
	s.layout.SetVisible(visible)
}

func (s *BottomActionView) SetNoDistractionMode(value bool) {
	value = !value
	s.layout.SetVisible(value)
}

func (s *BottomActionView) SetShowFullscreenButton(visible bool) {
	s.fullscreenButton.SetVisible(visible)
	s.exitFullscreenButton.SetVisible(!visible)
}

func (s *BottomActionView) SetShowOnlyCategory(status bool) {
	s.showAllImagesButton.SetVisible(status)
}
