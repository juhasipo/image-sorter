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
	categoriesView  *gtk.Box
	categoryButtons map[string]*CategoryButton
	nextButton      *gtk.Button
	prevButton      *gtk.Button
	sender          event.Sender
}

func TopActionsNew(builder *gtk.Builder, sender event.Sender) *TopActionView {
	topActionView := &TopActionView{
		categoriesView:  GetObjectOrPanic(builder, "categories").(*gtk.Box),
		categoryButtons: map[string]*CategoryButton{},
		nextButton:      GetObjectOrPanic(builder, "next-button").(*gtk.Button),
		prevButton:      GetObjectOrPanic(builder, "prev-button").(*gtk.Button),
		sender:          sender,
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

		stayOnSameImage := state&modifiers == gdk.GDK_SHIFT_MASK
		send(stayOnSameImage)
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
			stayOnSameImage := state&modifiers == gdk.GDK_SHIFT_MASK
			send(stayOnSameImage)
			return true
		}
		return false
	})
	s.categoriesView.Add(layout)
}

type CategorizeCallback func(*common.Category, common.Operation, bool)

func (s *TopActionView) createSendFuncForEntry(categoryButton *CategoryButton, categoizeCB CategorizeCallback) func(bool) {
	return func(stayOnSameImage bool) {
		log.Printf("Cat '%s': %d", categoryButton.entry.GetName(), categoryButton.operation)
		categoizeCB(categoryButton.entry, categoryButton.operation.NextOperation(), stayOnSameImage)
	}
}
