package ui

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
)

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
