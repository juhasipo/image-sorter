package ui

import (
	"errors"
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"os"
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
	imageCache  *imageloader.ImageCache
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

func Init(rootPath string, broker event.Sender, imageCache *imageloader.ImageCache) Gui {

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
	os.Setenv("GTK_THEME", "Adwaita:dark")

	// Application signals available
	// startup -> sets up the application when it first starts
	// activate -> shows the default first window of the application (like a new document). This corresponds to the application being launched by the desktop environment.
	// open -> opens files and shows them in a new window. This corresponds to someone trying to open a document (or documents) using the application from the file browser, or similar.
	// shutdown ->  performs shutdown tasks
	// Setup activate signal with a closure function.
	s.application.Connect("activate", func() {
		log.Println("Application activate")

		builder, err := gtk.BuilderNewFromFile("ui/main-view.glade")
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
	if key == gdk.KEY_F12 {
		s.sender.SendToTopic(event.SIMILAR_REQUEST_SEARCH)
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
	s.categories = make([]*common.Category, len(categories.GetCategories()))
	copy(s.categories, categories.GetCategories())

	s.topActionView.categoryButtons = map[string]*CategoryButton{}
	children := s.topActionView.categoriesView.GetChildren()
	children.Foreach(func(item interface{}) {
		s.topActionView.categoriesView.Remove(item.(gtk.IWidget))
	})

	for _, entry := range categories.GetCategories() {
		s.topActionView.addCategoryButton(entry, func(entry *common.Category, operation common.Operation, stayOnSameImage bool) {
			s.sender.SendToTopicWithData(
				event.CATEGORIZE_IMAGE,
				category.CategorizeCommandNewWithStayAttr(s.imageView.currentImage.image, entry, operation, stayOnSameImage))
		})
	}

	s.topActionView.categoriesView.ShowAll()
	s.sender.SendToTopic(event.IMAGE_REQUEST_CURRENT)
}

func (s *Ui) UpdateCurrentImage() {
	s.imageView.UpdateCurrentImage()
}

func (s *Ui) SetImages(imageTarget event.Topic, handles []*common.ImageContainer) {
	if imageTarget == event.IMAGE_REQUEST_NEXT {
		s.imageView.AddImagesToNextStore(handles)
	} else if imageTarget == event.IMAGE_REQUEST_PREV {
		s.imageView.AddImagesToPrevStore(handles)
	} else if imageTarget == event.IMAGE_REQUEST_SIMILAR {
		s.similarImagesView.SetImages(handles, s.sender)
	} else {
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
