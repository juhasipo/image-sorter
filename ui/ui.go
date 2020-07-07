package ui

import (
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
	// General
	win         *gtk.ApplicationWindow
	fullscreen  bool
	application *gtk.Application
	pixbufCache *pixbuf.PixbufCache
	sender      event.Sender

	// UI components
	progressView *ProgressView
	topActionView  *TopActionView
	imageView *ImageView
	similarImagesView *SimilarImagesView
	bottomActionView *BottomActionView
	castModal *CastModal

	Gui
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
		sender:      broker,
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
		s.win.Connect("key_press_event", s.handleKeyPress)

		s.InitSimilarImages(builder)
		s.InitImageView(builder)
		s.InitCastModal(builder)
		s.InitTopActions(builder)
		s.InitBottomActions(builder)

		s.InitProgressView(builder)

		s.sender.SendToTopic(event.UI_READY)

		// Show the Window and all of its components.
		s.win.Show()
		s.application.AddWindow(s.win)
	})
}

func (s *Ui) InitProgressView(builder *gtk.Builder) {
	s.progressView = &ProgressView{
		view:        getObjectOrPanic(builder, "progress-view").(*gtk.Box),
		stopButton:  getObjectOrPanic(builder, "stop-progress-button").(*gtk.Button),
		progressbar: getObjectOrPanic(builder, "progress-bar").(*gtk.ProgressBar),
	}
	s.progressView.stopButton.Connect("clicked", func() {
		s.sender.SendToTopic(event.SIMILAR_REQUEST_STOP)
	})
}

func (s *Ui) InitCastModal(builder *gtk.Builder) {
	castModal := getObjectOrPanic(builder, "cast-dialog").(*gtk.Dialog)
	deviceList := getObjectOrPanic(builder, "cast-device-list").(*gtk.TreeView)
	s.castModal = &CastModal{
		modal:          castModal,
		deviceListView: deviceList,
		model:          CreateDeviceList(castModal, deviceList, "Devices", s.sender),
		cancelButton:   getObjectOrPanic(builder, "cast-dialog-cancel-button").(*gtk.Button),
		refreshButton:   getObjectOrPanic(builder, "cast-dialog-refresh-button").(*gtk.Button),
		statusLabel:    getObjectOrPanic(builder, "cast-find-status-label").(*gtk.Label),
	}
	s.castModal.cancelButton.Connect("clicked", func() {
		s.castModal.modal.Hide()
	})
	s.castModal.refreshButton.Connect("clicked", s.findDevices)
}

func (s *Ui) InitImageView(builder *gtk.Builder) {
	nextImagesList := getObjectOrPanic(builder, "next-images").(*gtk.TreeView)
	nextImageStore := CreateImageList(nextImagesList, "Next images", FORWARD, s.sender)
	prevImagesList := getObjectOrPanic(builder, "prev-images").(*gtk.TreeView)
	prevImageStore := CreateImageList(prevImagesList, "Prev images", BACKWARD, s.sender)
	s.imageView = &ImageView{
		currentImage: &CurrentImageView{
			scrolledView: getObjectOrPanic(builder, "current-image-window").(*gtk.ScrolledWindow),
			viewport:     getObjectOrPanic(builder, "current-image-view").(*gtk.Viewport),
			view:         getObjectOrPanic(builder, "current-image").(*gtk.Image),
		},
		nextImages: &ImageList{
			component: nextImagesList,
			model:     nextImageStore,
		},
		prevImages: &ImageList{
			component: prevImagesList,
			model:     prevImageStore,
		},
	}
	s.imageView.currentImage.viewport.Connect("size-allocate", s.UpdateCurrentImage)
}

func (s *Ui) InitTopActions(builder *gtk.Builder) {
	s.topActionView = &TopActionView{
		categoriesView:  getObjectOrPanic(builder, "categories").(*gtk.Box),
		categoryButtons: map[*category.Entry]*CategoryButton{},
		nextButton:      getObjectOrPanic(builder, "next-button").(*gtk.Button),
		prevButton:      getObjectOrPanic(builder, "prev-button").(*gtk.Button),
	}
	s.topActionView.nextButton.Connect("clicked", func() {
		s.sender.SendToTopic(event.IMAGE_REQUEST_NEXT)
	})
	s.topActionView.prevButton.Connect("clicked", func() {
		s.sender.SendToTopic(event.IMAGE_REQUEST_PREV)
	})
}

func (s *Ui) InitBottomActions(builder *gtk.Builder) {
	s.bottomActionView = &BottomActionView{
		layout:            getObjectOrPanic(builder, "bottom-actions-view").(*gtk.Box),
		persistButton:     getObjectOrPanic(builder, "persist-button").(*gtk.Button),
		findSimilarButton: getObjectOrPanic(builder, "find-similar-button").(*gtk.Button),
		findDevicesButton: getObjectOrPanic(builder, "find-devices-button").(*gtk.Button),
	}
	s.bottomActionView.persistButton.Connect("clicked", func() {
		s.sender.SendToTopic(event.CATEGORY_PERSIST_ALL)
	})

	s.bottomActionView.findSimilarButton.Connect("clicked", func() {
		s.sender.SendToTopic(event.SIMILAR_REQUEST_SEARCH)
	})
	s.bottomActionView.findDevicesButton.Connect("clicked", s.findDevices)
}

func (s *Ui) findDevices() {
	s.castModal.deviceListView.SetVisible(false)
	s.castModal.refreshButton.SetSensitive(false)
	s.castModal.statusLabel.SetVisible(true)
	s.castModal.statusLabel.SetText("Searching for devices...")
	s.castModal.model.Clear()
	s.castModal.modal.SetTransientFor(s.application.GetActiveWindow())
	s.castModal.modal.Show()
	s.castModal.devices = []string{}
	s.sender.SendToTopic(event.CAST_DEVICE_SEARCH)
}

func (s *Ui) InitSimilarImages(builder *gtk.Builder) {
	layout, _ := gtk.FlowBoxNew()
	s.similarImagesView = &SimilarImagesView{
		scrollLayout: getObjectOrPanic(builder, "similar-images-view").(*gtk.ScrolledWindow),
		layout:       layout,
	}

	s.similarImagesView.layout.SetMaxChildrenPerLine(10)
	s.similarImagesView.layout.SetRowSpacing(0)
	s.similarImagesView.layout.SetColumnSpacing(0)
	s.similarImagesView.layout.SetSizeRequest(-1, 100)
	s.similarImagesView.scrollLayout.SetVisible(false)
	s.similarImagesView.scrollLayout.SetSizeRequest(-1, 100)
	s.similarImagesView.scrollLayout.Add(layout)
}

func (s *Ui) handleKeyPress(windows *gtk.ApplicationWindow, e *gdk.Event) bool {
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
		s.sender.SendToTopic(event.IMAGE_REQUEST_PREV)
		return true
	} else if key == gdk.KEY_Right {
		s.sender.SendToTopic(event.IMAGE_REQUEST_NEXT)
		return true
	} else {
		for entry, button := range s.topActionView.categoryButtons {
			if entry.HasShortcut(key) {
				keyName := KeyvalName(key)
				log.Printf("Key pressed: '%s': '%s'", keyName, entry.GetName())
				s.sender.SendToTopicWithData(
					event.CATEGORIZE_IMAGE,
					category.CategorizeCommandNew(s.imageView.currentImage.image, button.entry, button.operation.NextOperation()))
				return true
			}
		}
	}
	return false
}

func (s *Ui) UpdateCategories(categories *category.CategoriesCommand) {
	for _, entry := range categories.GetCategories() {
		button, _ := gtk.ButtonNewWithLabel(entry.GetName())

		categoryButton := &CategoryButton{
			button:    button,
			entry:     entry,
			operation: category.NONE,
		}
		s.topActionView.categoryButtons[entry] = categoryButton

		send := s.CreateSendFuncForEntry(categoryButton)
		button.Connect("clicked", func(button *gtk.Button) {
			send()
		})
		s.topActionView.categoriesView.Add(button)
	}
	s.topActionView.categoriesView.ShowAll()
}

func (s *Ui) CreateSendFuncForEntry(categoryButton *CategoryButton) func() {
	return func() {
		log.Printf("Cat '%s': %d", categoryButton.entry.GetName(), categoryButton.operation)
		s.sender.SendToTopicWithData(
			event.CATEGORIZE_IMAGE,
			category.CategorizeCommandNew(s.imageView.currentImage.image, categoryButton.entry, categoryButton.operation.NextOperation()))
	}
}

func (s *Ui) UpdateCurrentImage() {
	size := pixbuf.SizeFromWindow(s.imageView.currentImage.scrolledView)
	scaled := s.pixbufCache.GetScaled(
		s.imageView.currentImage.image,
		size,
	)
	s.imageView.currentImage.view.SetFromPixbuf(scaled)
	// Hack to prevent image from being center of the scrolled
	// window after minimize
	s.imageView.currentImage.scrolledView.Remove(s.imageView.currentImage.viewport)
	s.imageView.currentImage.scrolledView.Add(s.imageView.currentImage.viewport)
}

func (s *Ui) SetImages(imageTarget event.Topic, handles []*common.Handle) {
	if imageTarget == event.IMAGE_REQUEST_NEXT {
		s.AddImagesToStore(s.imageView.nextImages, handles)
	} else if imageTarget == event.IMAGE_REQUEST_PREV {
		s.AddImagesToStore(s.imageView.prevImages, handles)
	} else if imageTarget == event.IMAGE_REQUEST_SIMILAR {
		children := s.similarImagesView.layout.GetChildren()
		children.Foreach(func(item interface{}) {
			s.similarImagesView.layout.Remove(item.(gtk.IWidget))
		})
		for _, handle := range handles {
			widget := s.createSimilarImage(handle)
			s.similarImagesView.layout.Add(widget)
		}
		s.similarImagesView.scrollLayout.SetVisible(true)
		s.similarImagesView.scrollLayout.ShowAll()
	} else {
		s.SetCurrentImage(handles[0])
		s.pixbufCache.Purge(s.imageView.currentImage.image)
	}
}

func (s *Ui) createSimilarImage(handle *common.Handle) *gtk.EventBox {
	eventBox, _ := gtk.EventBoxNew()
	imageWidget, _ := gtk.ImageNewFromPixbuf(s.pixbufCache.GetThumbnail(handle))
	eventBox.Add(imageWidget)
	eventBox.Connect("button-press-event", func() {
		s.sender.SendToTopicWithData(event.IMAGE_REQUEST, handle)
	})
	return eventBox
}

func (s *Ui) SetCurrentImage(handle *common.Handle) {
	s.imageView.currentImage.image = handle
	s.UpdateCurrentImage()
	s.sendCurrentImageChangedEvent()
}

func (s *Ui) sendCurrentImageChangedEvent() {
	s.sender.SendToTopicWithData(event.IMAGE_CHANGED, s.imageView.currentImage.image)
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
	for _, button := range s.topActionView.categoryButtons {
		button.button.SetLabel(button.entry.GetName())
		button.operation = category.NONE
	}

	for _, command := range commands {
		log.Printf("Marked image category: '%s':%d", command.GetEntry().GetName(), command.GetOperation())
		button := s.topActionView.categoryButtons[command.GetEntry()]
		button.operation = command.GetOperation()
		button.button.SetLabel(command.ToLabel())
	}
}

func (s *Ui) UpdateProgress(name string, status int, total int) {
	if status == 0 {
		s.progressView.SetVisible(true)
		s.topActionView.SetVisible(false)
		s.bottomActionView.SetVisible(false)
	}

	if status == total {
		s.progressView.SetVisible(false)
		s.topActionView.SetVisible(true)
		s.bottomActionView.SetVisible(true)
	} else {
		s.progressView.SetStatus(status, total)
	}
}

func (s *Ui) DeviceFound(name string) {
	s.castModal.deviceListView.SetVisible(true)
	s.castModal.statusLabel.SetText("")
	s.castModal.statusLabel.SetVisible(false)
	iter := s.castModal.model.Append()
	s.castModal.model.SetValue(iter, 0, name)
	s.castModal.devices = append(s.castModal.devices, name)
}

func (s *Ui) CastReady() {
	s.sendCurrentImageChangedEvent()
}
func (s *Ui) CastFindDone() {
	if len(s.castModal.devices) == 0 {
		s.castModal.statusLabel.SetText("No devices found")
		s.castModal.statusLabel.SetVisible(true)
	}
	s.castModal.refreshButton.SetSensitive(true)
}
