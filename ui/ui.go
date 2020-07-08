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
	categories  []*category.Entry

	// UI components
	progressView *ProgressView
	topActionView  *TopActionView
	imageView *ImageView
	similarImagesView *SimilarImagesView
	bottomActionView *BottomActionView
	castModal *CastModal
	editCategoriesModal *CategoryModal

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
		s.win = GetObjectOrPanic(builder, "window").(*gtk.ApplicationWindow)
		s.win.SetSizeRequest(800, 600)
		s.win.Connect("key_press_event", s.handleKeyPress)

		s.similarImagesView = SimilarImagesViewNew(builder)
		s.imageView = ImageViewNew(builder, s)
		s.topActionView = TopActionsNew(builder, s.sender)
		s.bottomActionView = BottomActionsNew(builder, s, s.sender)
		s.progressView = ProgressViewNew(builder, s.sender)

		s.castModal = CastModalNew(builder, s, s.sender)
		s.editCategoriesModal = CategoryModalNew(builder, s, s.sender)

		s.sender.SendToTopic(event.UI_READY)

		// Show the Window and all of its components.
		s.win.Show()
		s.application.AddWindow(s.win)
	})
}

func (s *Ui) findDevices() {
	s.castModal.StartSearch(s.application.GetActiveWindow())
	s.sender.SendToTopic(event.CAST_DEVICE_SEARCH)
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
		command := s.topActionView.FindActionForShortcut(key, s.imageView.currentImage.image)
		if command != nil {
			s.sender.SendToTopicWithData(event.CATEGORIZE_IMAGE, command)
		}
	}
	return false
}

func (s *Ui) UpdateCategories(categories *category.CategoriesCommand) {
	s.categories = make([]*category.Entry, len(categories.GetCategories()))
	copy(s.categories, categories.GetCategories())

	s.topActionView.categoryButtons = map[*category.Entry]*CategoryButton{}
	children := s.topActionView.categoriesView.GetChildren()
	children.Foreach(func(item interface{}) {
		s.topActionView.categoriesView.Remove(item.(gtk.IWidget))
	})

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
	s.castModal.AddDevice(name)
}

func (s *Ui) CastReady() {
	s.sendCurrentImageChangedEvent()
}
func (s *Ui) CastFindDone() {
	if len(s.castModal.devices) == 0 {
		s.castModal.SetNoDevices()
	}
	s.castModal.SearchDone()
}

func (s *Ui) showEditCategoriesModal() {
	categories := make([]*category.Entry, len(s.categories))
	copy(categories, s.categories)
	s.editCategoriesModal.Show(s.application.GetActiveWindow(), categories)
}
