package ui

import (
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
)

type CategoryButton struct {
	layout    *gtk.Box
	toggle    *gtk.LevelBar
	button    *gtk.Button
	entry     *common.Category
	operation common.Operation
}

func (s *CategoryButton) SetStatus(operation common.Operation) {
	if operation == common.MOVE {
		s.toggle.SetValue(1.0)
	} else {
		s.toggle.SetValue(0.0)
	}
}

type TopActionView struct {
	categoriesView           *gtk.Box
	categoryButtons          map[string]*CategoryButton
	nextButton               *gtk.Button
	prevButton               *gtk.Button
	currentImagesStatusLabel *gtk.Label
	sender                   event.Sender
}

func TopActionsNew(builder *gtk.Builder, sender event.Sender) *TopActionView {
	topActionView := &TopActionView{
		categoriesView:           GetObjectOrPanic(builder, "categories").(*gtk.Box),
		categoryButtons:          map[string]*CategoryButton{},
		nextButton:               GetObjectOrPanic(builder, "next-button").(*gtk.Button),
		prevButton:               GetObjectOrPanic(builder, "prev-button").(*gtk.Button),
		currentImagesStatusLabel: GetObjectOrPanic(builder, "current-images-status-label").(*gtk.Label),
		sender:                   sender,
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
	v.currentImagesStatusLabel.SetVisible(visible)
}

func (v *TopActionView) FindActionForShortcut(key uint, handle *common.Handle) *category.CategorizeCommand {
	for _, button := range v.categoryButtons {
		entry := button.entry
		keyUpper := gdk.KeyvalToUpper(key)
		if entry.HasShortcut(keyUpper) {
			keyName := common.KeyvalName(key)
			log.Printf("Key pressed: '%s': '%s'", keyName, entry.GetName())
			return category.CategorizeCommandNew(
				handle, button.entry, button.operation.NextOperation())
		}
	}
	return nil
}

func (s *TopActionView) addCategoryButton(entry *common.Category, categorizeCallback CategorizeCallback) {
	layout, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	toggle, _ := gtk.LevelBarNew()
	toggle.SetSensitive(false)
	toggle.SetSizeRequest(-1, 5)
	button, _ := gtk.ButtonNewWithLabel(entry.GetName())
	button.SetHExpand(true)
	layout.Add(button)
	layout.Add(toggle)
	layout.SetHExpand(true)

	categoryButton := &CategoryButton{
		layout:    layout,
		button:    button,
		toggle:    toggle,
		entry:     entry,
		operation: common.NONE,
	}
	s.categoryButtons[entry.GetId()] = categoryButton

	send := s.createSendFuncForEntry(categoryButton, categorizeCallback)
	// Catches mouse click and can also check for keyboard for Shift key status
	button.Connect("button-release-event", func(button *gtk.Button, e *gdk.Event) bool {
		keyEvent := gdk.EventButtonNewFromEvent(e)

		modifiers := gtk.AcceleratorGetDefaultModMask()
		state := gdk.ModifierType(keyEvent.State())

		stayOnSameImage := state&modifiers&gdk.GDK_SHIFT_MASK > 0
		forceToCategory := state&modifiers&gdk.GDK_CONTROL_MASK > 0
		send(stayOnSameImage, forceToCategory)
		return true
	})
	// Since clicked handler is not used, Enter and Space need to be checked manually
	// also check Shift status
	button.Connect("key-press-event", func(button *gtk.Button, e *gdk.Event) bool {
		keyEvent := gdk.EventKeyNewFromEvent(e)
		key := keyEvent.KeyVal()

		if key == gdk.KEY_KP_Enter || key == gdk.KEY_Return || key == gdk.KEY_KP_Space || key == gdk.KEY_space {
			modifiers := gtk.AcceleratorGetDefaultModMask()
			state := gdk.ModifierType(keyEvent.State())
			stayOnSameImage := state&modifiers&gdk.GDK_SHIFT_MASK > 0
			forceToCategory := state&modifiers&gdk.GDK_CONTROL_MASK > 0
			send(stayOnSameImage, forceToCategory)
			return true
		}
		return false
	})
	s.categoriesView.Add(layout)
}

type CategorizeCallback func(*common.Category, common.Operation, bool, bool)

func (s *TopActionView) createSendFuncForEntry(categoryButton *CategoryButton, categoizeCB CategorizeCallback) func(bool, bool) {
	return func(stayOnSameImage bool, forceToCategory bool) {
		log.Printf("Cat '%s': %d", categoryButton.entry.GetName(), categoryButton.operation)
		if forceToCategory {
			categoizeCB(categoryButton.entry, common.MOVE, stayOnSameImage, forceToCategory)
		} else {
			categoizeCB(categoryButton.entry, categoryButton.operation.NextOperation(), stayOnSameImage, forceToCategory)
		}
	}
}

func (s *TopActionView) SetCurrentStatus(index int, total int, title string) {
	if title != "" {
		s.currentImagesStatusLabel.SetText(fmt.Sprintf("%s pictures (%d/%d)", title, index, total))
	} else {
		s.currentImagesStatusLabel.SetText(fmt.Sprintf("All pictures (%d/%d)", index, total))
	}
}
