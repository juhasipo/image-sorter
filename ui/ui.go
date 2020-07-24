package ui

import (
	"errors"
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/imageloader"
)

type Ui struct {
	// General
	win         *gtk.ApplicationWindow
	fullscreen  bool
	application *gtk.Application
	imageCache  imageloader.ImageCache
	sender      event.Sender
	categories  []*common.Category
	rootPath    string

	// UI components
	progressView        *ProgressView
	topActionView       *TopActionView
	imageView           *ImageView
	similarImagesView   *SimilarImagesView
	bottomActionView    *BottomActionView
	castModal           *CastModal
	editCategoriesModal *CategoryModal

	Gui
}

func Init(rootPath string, broker event.Sender, imageCache imageloader.ImageCache) Gui {

	// Create Gtk Application, change appID to your application domain name reversed.
	const appID = "org.gtk.example"
	application, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)

	// Check to make sure no errors when creating Gtk Application
	if err != nil {
		log.Fatal("Could not create application.", err)
	}

	ui := Ui{
		application: application,
		imageCache:  imageCache,
		sender:      broker,
		rootPath:    rootPath,
	}

	ui.Init(rootPath)
	return &ui
}

func (s *Ui) Init(directory string) {
	// Application signals available
	// startup -> sets up the application when it first starts
	// activate -> shows the default first window of the application (like a new document). This corresponds to the application being launched by the desktop environment.
	// open -> opens files and shows them in a new window. This corresponds to someone trying to open a document (or documents) using the application from the file browser, or similar.
	// shutdown ->  performs shutdown tasks
	// Setup activate signal with a closure function.
	s.application.Connect("activate", func() {
		log.Println("Application activate")

		if cssProvider, err := gtk.CssProviderNew(); err != nil {
			log.Panic("Error while loading CSS provider", err)
		} else if err = cssProvider.LoadFromPath("style.css"); err != nil {
			log.Panic("Error while loading CSS ", err)
		} else if screen, err := gdk.ScreenGetDefault(); err != nil {
			log.Panic("Error while loading screen ", err)
		} else {
			gtk.AddProviderForScreen(screen, cssProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
		}

		builder, err := gtk.BuilderNewFromFile("main-view.glade")
		if err != nil {
			log.Fatal("Could not load Glade file.", err)
		}

		// Get the object with the id of "main_window".
		s.win = GetObjectOrPanic(builder, "window").(*gtk.ApplicationWindow)
		s.win.SetSizeRequest(800, 600)
		s.win.Connect("key_press_event", s.handleKeyPress)

		s.similarImagesView = SimilarImagesViewNew(builder, s.sender, s.imageCache)
		s.imageView = ImageViewNew(builder, s)
		s.topActionView = TopActionsNew(builder, s.sender)
		s.bottomActionView = BottomActionsNew(builder, s, s.sender)
		s.progressView = ProgressViewNew(builder, s.sender)

		s.castModal = CastModalNew(builder, s, s.sender)
		s.editCategoriesModal = CategoryModalNew(builder, s, s.sender)

		if directory == "" {
			s.openFolderChooser(1)
		} else {
			s.sender.SendToTopicWithData(event.DIRECTORY_CHANGED, directory)
		}

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
	const BIG_JUMP_SIZE = 10
	const HUGE_JUMP_SIZE = 100

	keyEvent := gdk.EventKeyNewFromEvent(e)
	key := keyEvent.KeyVal()

	modifiers := gtk.AcceleratorGetDefaultModMask()
	state := gdk.ModifierType(keyEvent.State())
	controlDown := state&modifiers&gdk.GDK_CONTROL_MASK > 0

	if key == gdk.KEY_F8 {
		s.findDevices()
		return true
	} else if key == gdk.KEY_F10 {
		s.sender.SendToTopic(event.IMAGE_SHOW_ALL)
		return true
	} else if key == gdk.KEY_Escape {
		s.exitFullScreen()
		return true
	} else if key == gdk.KEY_F11 {
		noDistractionMode := controlDown
		if noDistractionMode {
			s.enterFullScreenNoDistraction()
		} else if s.fullscreen {
			s.exitFullScreen()
		} else {
			s.enterFullScreen()
		}
		return true
	} else if key == gdk.KEY_F12 {
		s.sender.SendToTopic(event.SIMILAR_REQUEST_SEARCH)
		return true
	}

	if key == gdk.KEY_Page_Up {
		s.sender.SendToTopicWithData(event.IMAGE_REQUEST_PREV_OFFSET, HUGE_JUMP_SIZE)
		return true
	} else if key == gdk.KEY_Page_Down {
		s.sender.SendToTopicWithData(event.IMAGE_REQUEST_NEXT_OFFSET, HUGE_JUMP_SIZE)
		return true
	} else if key == gdk.KEY_Home {
		s.sender.SendToTopicWithData(event.IMAGE_REQUEST_AT_INDEX, 0)
		return true
	} else if key == gdk.KEY_End {
		s.sender.SendToTopicWithData(event.IMAGE_REQUEST_AT_INDEX, -1)
		return true
	}

	if key == gdk.KEY_Left {
		if controlDown {
			s.sender.SendToTopicWithData(event.IMAGE_REQUEST_PREV_OFFSET, BIG_JUMP_SIZE)
		} else {
			s.sender.SendToTopic(event.IMAGE_REQUEST_PREV)
		}
		return true
	} else if key == gdk.KEY_Right {
		if controlDown {
			s.sender.SendToTopicWithData(event.IMAGE_REQUEST_NEXT_OFFSET, BIG_JUMP_SIZE)
		} else {
			s.sender.SendToTopic(event.IMAGE_REQUEST_NEXT)
		}
		return true
	} else if command := s.topActionView.FindActionForShortcut(key, s.imageView.currentImage.image); command != nil {
		switchToCategory := state&modifiers&gdk.GDK_MOD1_MASK > 0
		if switchToCategory {
			s.sender.SendToTopicWithData(event.CATEGORIES_SHOW_ONLY, command.GetEntry())
		} else {
			stayOnSameImage := state&modifiers&gdk.GDK_SHIFT_MASK > 0
			forceToCategory := state&modifiers&gdk.GDK_CONTROL_MASK > 0
			command.SetStayOfSameImage(stayOnSameImage)
			command.SetForceToCategory(forceToCategory)
			s.sender.SendToTopicWithData(event.CATEGORIZE_IMAGE, command)
		}
		return true
	}
	return false
}

func (s *Ui) enterFullScreenNoDistraction() {
	s.win.Fullscreen()
	s.fullscreen = true
	s.imageView.SetNoDistractionMode(true)
	s.topActionView.SetNoDistractionMode(true)
	s.bottomActionView.SetNoDistractionMode(true)
}

func (s *Ui) enterFullScreen() {
	s.win.Fullscreen()
	s.fullscreen = true
	s.imageView.SetNoDistractionMode(false)
	s.topActionView.SetNoDistractionMode(false)
	s.bottomActionView.SetNoDistractionMode(false)
}

func (s *Ui) exitFullScreen() {
	s.win.Unfullscreen()
	s.fullscreen = false
	s.imageView.SetNoDistractionMode(false)
	s.topActionView.SetNoDistractionMode(false)
	s.bottomActionView.SetNoDistractionMode(false)
}

func (s *Ui) UpdateCategories(categories *category.CategoriesCommand) {
	s.categories = make([]*common.Category, len(categories.GetCategories()))
	copy(s.categories, categories.GetCategories())

	s.topActionView.categoryButtons = map[string]*CategoryButton{}
	children := s.topActionView.categoriesView.GetChildren()
	children.Foreach(func(item interface{}) {
		s.topActionView.categoriesView.Remove(item.(gtk.IWidget))
	})

	for _, entry := range categories.GetCategories() {
		s.topActionView.addCategoryButton(entry, func(entry *common.Category, operation common.Operation, stayOnSameImage bool, forceToCategory bool) {
			s.sender.SendToTopicWithData(
				event.CATEGORIZE_IMAGE,
				category.CategorizeCommandNewWithStayAttr(s.imageView.currentImage.image, entry, operation, stayOnSameImage, forceToCategory))
		})
	}

	s.topActionView.categoriesView.ShowAll()
	s.sender.SendToTopic(event.IMAGE_REQUEST_CURRENT)
}

func (s *Ui) UpdateCurrentImage() {
	s.imageView.UpdateCurrentImage()
}

func (s *Ui) SetImages(imageTarget event.Topic, handles []*common.ImageContainer, index int, total int, title string) {
	if imageTarget == event.IMAGE_REQUEST_NEXT {
		s.imageView.AddImagesToNextStore(handles)
	} else if imageTarget == event.IMAGE_REQUEST_PREV {
		s.imageView.AddImagesToPrevStore(handles)
	} else if imageTarget == event.IMAGE_REQUEST_SIMILAR {
		s.similarImagesView.SetImages(handles, s.sender)
	} else {
		s.topActionView.SetCurrentStatus(index, total, title)
		s.SetCurrentImage(handles[0])
		s.imageCache.Purge(s.imageView.currentImage.image)
	}
}

func (s *Ui) SetCurrentImage(handle *common.ImageContainer) {
	s.imageView.SetCurrentImage(handle)

	s.UpdateCurrentImage()
	s.sendCurrentImageChangedEvent()
}

func (s *Ui) sendCurrentImageChangedEvent() {
	s.sender.SendToTopicWithData(event.IMAGE_CHANGED, s.imageView.currentImage.image)
}

func (s *Ui) Run() {
	s.application.Run([]string{})
}

func (s *Ui) SetImageCategory(commands []*category.CategorizeCommand) {
	log.Print("Start setting image category")
	for _, button := range s.topActionView.categoryButtons {
		button.button.SetLabel(button.entry.GetName())
		button.operation = common.NONE
		button.SetStatus(button.operation)
	}

	for _, command := range commands {
		log.Printf("Marked image category: '%s':%d", command.GetEntry().GetName(), command.GetOperation())
		button := s.topActionView.categoryButtons[command.GetEntry().GetId()]
		if button != nil {
			button.operation = command.GetOperation()
			button.button.SetLabel(command.ToLabel())
			button.SetStatus(button.operation)
		}
	}
	log.Print("End setting image category")
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
	categories := make([]*common.Category, len(s.categories))
	copy(categories, s.categories)
	s.editCategoriesModal.Show(s.application.GetActiveWindow(), categories)
}

func (s *Ui) openFolderChooser(numOfButtons int) {
	folderChooser, err := createFileChooser(numOfButtons, s.application.GetActiveWindow())
	if err != nil {
		log.Panic("Can't open file chooser")
	}
	defer folderChooser.Destroy()

	runAndProcessFolderChooser(folderChooser, s.sender)
}

func runAndProcessFolderChooser(folderChooser *gtk.FileChooserDialog, sender event.Sender) {
	response := folderChooser.Run()
	if response == gtk.RESPONSE_ACCEPT {
		folder := folderChooser.GetFilename()
		sender.SendToTopicWithData(event.DIRECTORY_CHANGED, folder)
	}
}

func createFileChooser(numOfButtons int, parent gtk.IWindow) (*gtk.FileChooserDialog, error) {
	var folderChooser *gtk.FileChooserDialog
	var err error

	if numOfButtons == 1 {
		folderChooser, err = gtk.FileChooserDialogNewWith1Button(
			"Select folder", parent,
			gtk.FILE_CHOOSER_ACTION_SELECT_FOLDER,
			"Select", gtk.RESPONSE_ACCEPT)
	} else if numOfButtons == 2 {
		folderChooser, err = gtk.FileChooserDialogNewWith2Buttons(
			"Select folder", parent,
			gtk.FILE_CHOOSER_ACTION_SELECT_FOLDER,
			"Select", gtk.RESPONSE_ACCEPT,
			"Cancel", gtk.RESPONSE_CANCEL)
	} else {
		err = errors.New(fmt.Sprintf("Invalid number of buttons: %d", numOfButtons))
	}
	return folderChooser, err
}
