package ui

import (
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/pixbuf"
)

type Ui struct {
	win               *gtk.ApplicationWindow
	fullscreen        bool
	application       *gtk.Application
	pixbufCache       *pixbuf.PixbufCache
	currentImage      *CurrentImage
	nextImages        *ImageList
	prevImages        *ImageList
	nextButton        *gtk.Button
	prevButton        *gtk.Button
	currentImageWindow  *gtk.ScrolledWindow
	currentImageView  *gtk.Viewport
	categoriesView    *gtk.Box
	categoryButtons   map[*category.Entry]*CategoryButton
	persistButton     *gtk.Button
	broker            event.Sender
	findSimilarButton *gtk.Button
	findDevicesButton *gtk.Button
	progressBar       *gtk.ProgressBar
	mainButtonsView   *gtk.Box

	Gui
	similarImagesView *gtk.ScrolledWindow
	similarImages     *gtk.FlowBox

	castModal         *CastModal
}

type CastModal struct {
	modal *gtk.Dialog
	deviceList *gtk.TreeView
	deviceListStore *gtk.ListStore
	cancelButton *gtk.Button
}

func Init(broker event.Sender, pixbufCache *pixbuf.PixbufCache) Gui {

	// Create Gtk Application, change appID to your application domain name reversed.
	const appID = "org.gtk.example"
	application, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)

	// Check to make sure no errors when creating Gtk Application
	if err != nil {
		log.Fatal("Could not create application.", err)
	}

	ui := Ui{
		application: application,
		pixbufCache: pixbufCache,
		broker: broker,
	}

	ui.Init()
	return &ui
}

func (s *Ui) Init() {
	cssProvider, _ := gtk.CssProviderNew()
	if err := cssProvider.LoadFromPath("ui/default.css"); err != nil {
		log.Panic("Error while loading CSS ", err)
	}

	// Application signals available
	// startup -> sets up the application when it first starts
	// activate -> shows the default first window of the application (like a new document). This corresponds to the application being launched by the desktop environment.
	// open -> opens files and shows them in a new window. This corresponds to someone trying to open a document (or documents) using the application from the file browser, or similar.
	// shutdown ->  performs shutdown tasks
	// Setup activate signal with a closure function.
	s.application.Connect("activate", func() {
		log.Println("Application activate")

		screen, err := gdk.ScreenGetDefault()
		if err != nil {
			log.Panic("Error while loading screen ", err)
		}
		gtk.AddProviderForScreen(screen, cssProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

		builder, err := gtk.BuilderNewFromFile("ui/main-view.glade")
		if err != nil {
			log.Fatal("Could not load Glade file.", err)
		}

		// Get the object with the id of "main_window".
		s.win = getObjectOrPanic(builder, "window").(*gtk.ApplicationWindow)
		s.win.SetSizeRequest(800, 600)

		s.win.Connect("key_press_event", func(windows *gtk.ApplicationWindow, e *gdk.Event) bool {
			keyEvent := gdk.EventKeyNewFromEvent(e)
			key := keyEvent.KeyVal()
			if key == gdk.KEY_F11 {
				if s.fullscreen {
					s.win.Unfullscreen()
					s.fullscreen = false
				} else {
					s.win.Fullscreen()
					s.fullscreen = true
				}
				return true
			}
			if key == gdk.KEY_Left {
				s.broker.SendToTopic(event.PREV_IMAGE)
				return true
			} else if key == gdk.KEY_Right {
				s.broker.SendToTopic(event.NEXT_IMAGE)
				return true
			} else {
				for entry, button := range s.categoryButtons {
					if entry.HasShortcut(key) {
						keyName := KeyvalName(key)
						log.Printf("Key pressed: '%s': '%s'", keyName, entry.GetName())
						s.broker.SendToTopicWithData(
							event.CATEGORIZE_IMAGE,
							category.CategorizeCommandNew(s.currentImage.image, button.entry, NextOperation(button.operation)))
						return true
					}
				}
			}
			return false
		})

		nextImagesList := getObjectOrPanic(builder, "next-images").(*gtk.TreeView)
		nextImageStore := CreateImageList(nextImagesList, "Next images", FORWARD, s.broker)
		prevImagesList := getObjectOrPanic(builder, "prev-images").(*gtk.TreeView)
		prevImageStore := CreateImageList(prevImagesList, "Prev images", BACKWARD, s.broker)
		s.similarImagesView = getObjectOrPanic(builder, "similar-images-view").(*gtk.ScrolledWindow)
		similarImages, _ := gtk.FlowBoxNew()
		similarImages.SetMaxChildrenPerLine(10)
		similarImages.SetRowSpacing(0)
		similarImages.SetColumnSpacing(0)
		s.similarImages = similarImages

		s.similarImagesView.SetVisible(false)
		s.similarImagesView.SetSizeRequest(-1, 100)
		s.similarImages.SetSizeRequest(-1, 100)

		s.similarImagesView.Add(similarImages)

		s.nextImages = &ImageList{
			component: nextImagesList,
			model:    nextImageStore,
		}
		s.prevImages = &ImageList{
			component: prevImagesList,
			model:    prevImageStore,
		}
		s.currentImage = &CurrentImage {
			view: getObjectOrPanic(builder, "current-image").(*gtk.Image),
		}
		s.currentImageWindow = getObjectOrPanic(builder, "current-image-window").(*gtk.ScrolledWindow)
		s.currentImageView = getObjectOrPanic(builder, "current-image-view").(*gtk.Viewport)
		s.mainButtonsView = getObjectOrPanic(builder, "main-buttons-view").(*gtk.Box)
		s.nextButton = getObjectOrPanic(builder, "next-button").(*gtk.Button)
		s.prevButton = getObjectOrPanic(builder, "prev-button").(*gtk.Button)
		s.findSimilarButton = getObjectOrPanic(builder, "find-similar-button").(*gtk.Button)
		s.findDevicesButton = getObjectOrPanic(builder, "find-devices-button").(*gtk.Button)
		s.categoriesView = getObjectOrPanic(builder, "categories").(*gtk.Box)
		s.persistButton = getObjectOrPanic(builder, "persist-button").(*gtk.Button)
		s.progressBar = getObjectOrPanic(builder, "progress-bar").(*gtk.ProgressBar)

		castModal := getObjectOrPanic(builder, "cast-dialog").(*gtk.Dialog)
		deviceList := getObjectOrPanic(builder, "cast-device-list").(*gtk.TreeView)
		s.castModal = &CastModal{
			modal:           castModal,
			deviceList:      deviceList,
			deviceListStore: CreateDeviceList(castModal, deviceList, "Devices", s.broker),
			cancelButton:    getObjectOrPanic(builder, "cast-dialog-cancel-button").(*gtk.Button),
		}
		s.castModal.cancelButton.Connect("clicked", func() {
			s.castModal.modal.Hide()
		})

		s.currentImageView.Connect("size-allocate", s.UpdateCurrentImage)

		s.nextButton.Connect("clicked", func() {
			s.broker.SendToTopic(event.NEXT_IMAGE)
		})
		s.prevButton.Connect("clicked", func() {
			s.broker.SendToTopic(event.PREV_IMAGE)
		})
		s.persistButton.Connect("clicked", func() {
			s.broker.SendToTopic(event.PERSIST_CATEGORIES)
		})

		s.findSimilarButton.Connect("clicked", func() {
			s.broker.SendToTopic(event.GENERATE_HASHES)
		})
		s.findDevicesButton.Connect("clicked", func() {
			s.castModal.deviceListStore.Clear()
			s.castModal.modal.Show()
			s.broker.SendToTopic(event.CAST_FIND_DEVICES)
		})

		s.broker.SendToTopic(event.UI_READY)

		s.categoryButtons = map[*category.Entry]*CategoryButton{}

		// Show the Window and all of its components.
		s.win.Show()
		s.application.AddWindow(s.win)
	})
}

type CategoryButton struct {
	button *gtk.Button
	entry *category.Entry
	operation category.Operation
}

func (s *Ui) UpdateCategories(categories *category.CategoriesCommand) {
	children := s.categoriesView.GetChildren()

	for iter := children; iter != nil; iter = children.Next() {
		// TODO: Remove
	}

	for _, entry := range categories.GetCategories() {
		button, _ := gtk.ButtonNewWithLabel(entry.GetName())

		categoryButton := &CategoryButton{
			button:    button,
			entry:     entry,
			operation: category.NONE,
		}
		s.categoryButtons[entry] = categoryButton

		send := s.CreateSendFuncForEntry(categoryButton)
		button.Connect("clicked", func(button *gtk.Button) {
			send()
		})
		s.categoriesView.Add(button)
	}
	s.categoriesView.ShowAll()
}

func (s *Ui) CreateSendFuncForEntry(categoryButton *CategoryButton) func() {
	return func() {
		log.Printf("Cat '%s': %d", categoryButton.entry.GetName(), categoryButton.operation)
		s.broker.SendToTopicWithData(
			event.CATEGORIZE_IMAGE,
			category.CategorizeCommandNew(s.currentImage.image, categoryButton.entry, NextOperation(categoryButton.operation)))
	}
}

func (s *Ui) UpdateCurrentImage() {
	size := pixbuf.SizeFromWindow(s.currentImageWindow)
	scaled := s.pixbufCache.GetScaled(
		s.currentImage.image,
		size,
	)
	s.currentImage.view.SetFromPixbuf(scaled)
	// Hack to prevent image from being center of the scrolled
	// window after minimize
	s.currentImageWindow.Remove(s.currentImageView)
	s.currentImageWindow.Add(s.currentImageView)
}

func (s* Ui) SetImages(imageTarget event.Topic, handles []*common.Handle) {
	if imageTarget == event.NEXT_IMAGE {
		s.AddImagesToStore(s.nextImages, handles)
	} else if imageTarget == event.PREV_IMAGE {
		s.AddImagesToStore(s.prevImages, handles)
	} else if imageTarget == event.SIMILAR_IMAGE {
		children := s.similarImages.GetChildren()
		children.Foreach(func(item interface{}) {
			s.similarImages.Remove(item.(gtk.IWidget))
		})
		for _, handle := range handles {
			widget := s.createSimilarImage(handle)
			s.similarImages.Add(widget)
		}
		s.similarImagesView.SetVisible(true)
		s.similarImages.ShowAll()
	} else {
		s.SetCurrentImage(handles[0])
		s.pixbufCache.Purge(s.currentImage.image)
	}
}

func (s *Ui) createSimilarImage(handle *common.Handle) *gtk.EventBox {
	eventBox, _ := gtk.EventBoxNew()
	imageWidget, _ := gtk.ImageNewFromPixbuf(s.pixbufCache.GetThumbnail(handle))
	eventBox.Add(imageWidget)
	eventBox.Connect("button-press-event", func() {
		s.broker.SendToTopicWithData(event.JUMP_TO_IMAGE, handle)
	})
	return eventBox
}

func (s *Ui) SetCurrentImage(handle *common.Handle) {
	s.currentImage.image = handle
	s.UpdateCurrentImage()
	s.sendCurrentImageChangedEvent()
}

func (s *Ui) sendCurrentImageChangedEvent() {
	s.broker.SendToTopicWithData(event.IMAGE_CHANGED, s.currentImage.image)
}

func (s *Ui) AddImagesToStore(list *ImageList, images []*common.Handle) {
	list.model.Clear()
	for _, img := range images {
		iter := list.model.Append()
		list.model.SetValue(iter, 0, s.pixbufCache.GetThumbnail(img))
	}
}

func (s *Ui) Run() {
	s.application.Run([]string{})
}

func (s *Ui) SetImageCategory(commands []*category.CategorizeCommand) {
	for _, button := range s.categoryButtons {
		button.button.SetLabel(button.entry.GetName())
		button.operation = category.NONE
	}

	for _, command := range commands {
		log.Printf("Marked image category: '%s':%d", command.GetEntry().GetName(), command.GetOperation())
		button := s.categoryButtons[command.GetEntry()]
		button.operation = command.GetOperation()
		button.button.SetLabel(CommandToLabel(command))
	}
}

func CommandToLabel(command *category.CategorizeCommand) string {
	entryName := command.GetEntry().GetName()
	status := ""
	switch command.GetOperation() {
	case category.COPY: status = "C"
	case category.MOVE: status = "M"
	}

	if status != "" {
		return entryName + " (" + status + ")"
	} else {
		return entryName
	}
}

func NextOperation(operation category.Operation) category.Operation {
	return (operation + 1) % 3
}

func (s *Ui) UpdateProgress(name string, status int, total int) {
	if status == 0 {
		s.progressBar.SetVisible(true)
		s.mainButtonsView.SetVisible(false)
	}

	if status == total {
		s.progressBar.SetVisible(false)
		s.mainButtonsView.SetVisible(true)
		return
	}

	statusText := fmt.Sprintf("Processed %d/%d", status, total)
	s.progressBar.SetText(statusText)
	s.progressBar.SetFraction(float64(status) / float64(total))
}

func (s *Ui) DeviceFound(name string) {
	iter := s.castModal.deviceListStore.Append()
	s.castModal.deviceListStore.SetValue(iter, 0, name)
}

func (s *Ui) CastReady() {
	s.sendCurrentImageChangedEvent()
}
