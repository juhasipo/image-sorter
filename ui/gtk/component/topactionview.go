package component

import (
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"time"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/common/event"
	"vincit.fi/image-sorter/common/logger"
)

type CategoryButton struct {
	layout    *gtk.Box
	toggle    *gtk.LevelBar
	button    *gtk.Button
	entry     *apitype.Category
	operation apitype.Operation
}

func (s *CategoryButton) SetStatus(operation apitype.Operation) {
	s.operation = operation
	if operation == apitype.MOVE {
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

func NewTopActions(builder *gtk.Builder, sender event.Sender) *TopActionView {
	topActionView := &TopActionView{
		categoriesView:           GetObjectOrPanic(builder, "categories").(*gtk.Box),
		categoryButtons:          map[string]*CategoryButton{},
		nextButton:               GetObjectOrPanic(builder, "next-button").(*gtk.Button),
		prevButton:               GetObjectOrPanic(builder, "prev-button").(*gtk.Button),
		currentImagesStatusLabel: GetObjectOrPanic(builder, "current-images-status-label").(*gtk.Label),
		sender:                   sender,
	}
	topActionView.nextButton.Connect("clicked", func() {
		sender.SendToTopic(event.ImageRequestNext)
	})
	topActionView.prevButton.Connect("clicked", func() {
		sender.SendToTopic(event.ImageRequestPrev)
	})

	return topActionView
}

func (v *TopActionView) SetVisible(visible bool) {
	v.categoriesView.SetVisible(visible)
	v.nextButton.SetVisible(visible)
	v.prevButton.SetVisible(visible)
	v.currentImagesStatusLabel.SetVisible(visible)
}

func (v *TopActionView) SetNoDistractionMode(value bool) {
	value = !value
	v.nextButton.SetVisible(value)
	v.prevButton.SetVisible(value)
}

func (v *TopActionView) FindActionForShortcut(key uint, handle *apitype.Handle) *apitype.CategorizeCommand {
	for _, button := range v.categoryButtons {
		entry := button.entry
		keyUpper := gdk.KeyvalToUpper(key)
		if entry.HasShortcut(keyUpper) {
			keyName := common.KeyvalName(key)
			logger.Debug.Printf("Key pressed: '%s': '%s'", keyName, entry.GetName())
			return apitype.NewCategorizeCommand(
				handle, button.entry, button.operation.NextOperation())
		}
	}
	return nil
}

func (s *TopActionView) addCategoryButton(entry *apitype.Category, categorizeCallback CategorizeCallback) {
	layout, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	button, _ := gtk.ButtonNewWithLabel(fmt.Sprintf("%s (%s)", entry.GetName(), entry.GetShortcutAsString()))
	toggle, _ := gtk.LevelBarNew()

	categoryButton := &CategoryButton{
		layout:    layout,
		button:    button,
		toggle:    toggle,
		entry:     entry,
		operation: apitype.NONE,
	}
	s.categoryButtons[entry.GetId()] = categoryButton

	send := s.createSendFuncForEntry(categoryButton, categorizeCallback)

	toggle.SetSensitive(false)
	toggle.SetSizeRequest(-1, 5)
	button.SetHExpand(true)

	menuButton, _ := gtk.MenuButtonNew()
	menuPopover, _ := gtk.PopoverNew(menuButton)

	menuButton.SetPopover(menuPopover)

	buttonBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	buttonBox.SetHExpand(true)
	buttonBox.SetChildPacking(button, true, true, 0, gtk.PACK_START)
	buttonBox.SetChildPacking(menuButton, false, true, 0, gtk.PACK_START)

	buttonBox.Add(button)
	buttonBox.Add(menuButton)
	menuBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)

	browseButton, _ := gtk.ButtonNewWithLabel(fmt.Sprintf("Browse '%s' (ALT + %s)", entry.GetName(), entry.GetShortcutAsString()))
	browseButton.SetRelief(gtk.RELIEF_NONE)
	browseButton.Connect("clicked", func() {
		s.sender.SendToTopicWithData(event.CategoriesShowOnly, entry)
	})
	menuBox.Add(browseButton)

	addAndStayButton, _ := gtk.ButtonNewWithLabel(fmt.Sprintf("Add '%s' and Stay (Shift + %s)", entry.GetName(), entry.GetShortcutAsString()))
	addAndStayButton.SetRelief(gtk.RELIEF_NONE)
	addAndStayButton.Connect("clicked", func() {
		send(true, false)
	})
	menuBox.Add(addAndStayButton)

	setAsOnly, _ := gtk.ButtonNewWithLabel(fmt.Sprintf("Set '%s' as only (CTRL + %s)", entry.GetName(), entry.GetShortcutAsString()))
	setAsOnly.SetRelief(gtk.RELIEF_NONE)
	setAsOnly.Connect("clicked", func() {
		send(false, true)
	})
	menuBox.Add(setAsOnly)
	menuBox.ShowAll()

	menuPopover.Add(menuBox)

	layout.Add(buttonBox)
	layout.Add(toggle)
	layout.SetHExpand(true)

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

type CategorizeCallback func(*apitype.Category, apitype.Operation, bool, bool)

func (s *TopActionView) createSendFuncForEntry(categoryButton *CategoryButton, categoizeCB CategorizeCallback) func(bool, bool) {
	return func(stayOnSameImage bool, forceToCategory bool) {
		logger.Debug.Printf("Cat '%s': %d", categoryButton.entry.GetName(), categoryButton.operation)
		if forceToCategory {
			categoizeCB(categoryButton.entry, apitype.MOVE, stayOnSameImage, forceToCategory)
		} else {
			categoizeCB(categoryButton.entry, categoryButton.operation.NextOperation(), stayOnSameImage, forceToCategory)
		}
	}
}

func (s *TopActionView) SetCurrentStatus(index int, total int, title string) {
	progressText := ""
	if total == 0 {
		progressText = "No images"
	} else {
		progressText = fmt.Sprintf("%d/%d", index, total)
	}

	if title != "" {
		s.currentImagesStatusLabel.SetText(fmt.Sprintf("%s pictures (%s)", title, progressText))
	} else {
		s.currentImagesStatusLabel.SetText(fmt.Sprintf("All pictures (%s)", progressText))
	}
}

func (s *TopActionView) UpdateCategories(categories *apitype.CategoriesCommand, currentImageHandle *apitype.Handle) {
	s.categoryButtons = map[string]*CategoryButton{}
	children := s.categoriesView.GetChildren()
	children.Foreach(func(item interface{}) {
		s.categoriesView.Remove(item.(gtk.IWidget))
	})

	for _, entry := range categories.GetCategories() {
		s.addCategoryButton(entry, func(entry *apitype.Category, operation apitype.Operation, stayOnSameImage bool, forceToCategory bool) {
			command := apitype.NewCategorizeCommand(currentImageHandle, entry, operation)
			command.SetForceToCategory(forceToCategory)
			command.SetStayOfSameImage(stayOnSameImage)
			command.SetNextImageDelay(200 * time.Millisecond)
			s.sender.SendToTopicWithData(
				event.CategorizeImage,
				command)
		})
	}

	s.categoriesView.ShowAll()
}

func (s *TopActionView) GetCategoryButtons() map[string]*CategoryButton {
	return s.categoryButtons
}

func (s *TopActionView) GetCategoryButton(id string) (*CategoryButton, bool) {
	button, ok := s.categoryButtons[id]
	return button, ok
}
