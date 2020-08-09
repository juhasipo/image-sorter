package gtk

import (
	"errors"
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/imageloader"
	"vincit.fi/image-sorter/logger"
	"vincit.fi/image-sorter/ui"
	"vincit.fi/image-sorter/ui/gtk/component"
)

type Ui struct {
	// General
	win         *gtk.ApplicationWindow
	fullscreen  bool
	application *gtk.Application
	imageCache  imageloader.ImageStore
	sender      event.Sender
	categories  []*common.Category
	rootPath    string

	// UI components
	progressView        *component.ProgressView
	topActionView       *component.TopActionView
	imageView           *component.ImageView
	similarImagesView   *component.SimilarImagesView
	bottomActionView    *component.BottomActionView
	castModal           *component.CastModal
	editCategoriesModal *component.CategoryModal

	ui.Gui
	component.CallbackApi
}

func Init(rootPath string, broker event.Sender, imageCache imageloader.ImageStore) ui.Gui {

	// Create Gtk Application, change appID to your application domain name reversed.
	const appID = "fi.vincit.imagesorter"
	application, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)

	// Check to make sure no errors when creating Gtk Application
	if err != nil {
		logger.Error.Fatal("Could not create application.", err)
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
		logger.Debug.Println("Application activate")

		if cssProvider, err := gtk.CssProviderNew(); err != nil {
			logger.Error.Panic("Error while loading CSS provider", err)
		} else if err = cssProvider.LoadFromPath("style.css"); err != nil {
			logger.Error.Panic("Error while loading CSS ", err)
		} else if screen, err := gdk.ScreenGetDefault(); err != nil {
			logger.Error.Panic("Error while loading screen ", err)
		} else {
			gtk.AddProviderForScreen(screen, cssProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
		}

		builder, err := gtk.BuilderNewFromFile("main-view.glade")
		if err != nil {
			logger.Error.Fatal("Could not load Glade file.", err)
		}

		// Get the object with the id of "main_window".
		s.win = component.GetObjectOrPanic(builder, "window").(*gtk.ApplicationWindow)
		s.win.SetSizeRequest(800, 600)
		s.win.Connect("key_press_event", s.handleKeyPress)

		s.similarImagesView = component.NewSimilarImagesView(builder, s.sender, s.imageCache)
		s.imageView = component.NewImageView(builder, s.sender, s.imageCache)
		s.topActionView = component.NewTopActions(builder, s.sender)
		s.bottomActionView = component.NewBottomActions(builder, s.application.GetActiveWindow(), s, s.sender)
		s.progressView = component.NewProgressView(builder, s.sender)

		s.castModal = component.NewCastModal(builder, s, s.sender)
		s.editCategoriesModal = component.NewCategoryModal(builder, s.sender)

		if directory == "" {
			s.OpenFolderChooser(1)
		} else {
			s.sender.SendToTopicWithData(event.DirectoryChanged, directory)
		}

		// Show the Window and all of its components.
		s.win.Show()
		s.application.AddWindow(s.win)
	})
}

func (s *Ui) handleKeyPress(_ *gtk.ApplicationWindow, e *gdk.Event) bool {
	const bigJumpSize = 10
	const hugeJumpSize = 100

	keyEvent := gdk.EventKeyNewFromEvent(e)
	key := keyEvent.KeyVal()

	modifiers := gtk.AcceleratorGetDefaultModMask()
	state := gdk.ModifierType(keyEvent.State())
	modifierType := state & modifiers
	shiftDown := modifierType&gdk.GDK_CONTROL_MASK > 0
	controlDown := modifierType&gdk.GDK_CONTROL_MASK > 0
	altDown := modifierType&gdk.GDK_MOD1_MASK > 0

	if key == gdk.KEY_F8 {
		s.FindDevices()
	} else if key == gdk.KEY_F10 {
		s.sender.SendToTopic(event.ImageShowAll)
	} else if key == gdk.KEY_Escape {
		s.ExitFullScreen()
	} else if key == gdk.KEY_F11 || (altDown && key == gdk.KEY_Return) {
		noDistractionMode := controlDown
		if noDistractionMode {
			s.EnterFullScreenNoDistraction()
		} else if s.fullscreen {
			s.ExitFullScreen()
		} else {
			s.EnterFullScreen()
		}
	} else if key == gdk.KEY_F12 {
		s.sender.SendToTopic(event.SimilarRequestSearch)
	} else if key == gdk.KEY_Page_Up {
		s.sender.SendToTopicWithData(event.ImageRequestPrevOffset, hugeJumpSize)
	} else if key == gdk.KEY_Page_Down {
		s.sender.SendToTopicWithData(event.ImageRequestNextOffset, hugeJumpSize)
	} else if key == gdk.KEY_Home {
		s.sender.SendToTopicWithData(event.ImageRequestAtIndex, 0)
	} else if key == gdk.KEY_End {
		s.sender.SendToTopicWithData(event.ImageRequestAtIndex, -1)
	} else if key == gdk.KEY_Left {
		if controlDown {
			s.sender.SendToTopicWithData(event.ImageRequestPrevOffset, bigJumpSize)
		} else {
			s.sender.SendToTopic(event.ImageRequestPrev)
		}
	} else if key == gdk.KEY_Right {
		if controlDown {
			s.sender.SendToTopicWithData(event.ImageRequestNextOffset, bigJumpSize)
		} else {
			s.sender.SendToTopic(event.ImageRequestNext)
		}
	} else if command := s.topActionView.FindActionForShortcut(key, s.imageView.GetCurrentHandle()); command != nil {
		switchToCategory := altDown
		if switchToCategory {
			s.sender.SendToTopicWithData(event.CategoriesShowOnly, command.GetEntry())
		} else {
			stayOnSameImage := shiftDown
			forceToCategory := controlDown
			command.SetStayOfSameImage(stayOnSameImage)
			command.SetForceToCategory(forceToCategory)
			s.sender.SendToTopicWithData(event.CategorizeImage, command)
		}
	} else if key == gdk.KEY_plus || key == gdk.KEY_KP_Add {
		s.imageView.ZoomIn()
	} else if key == gdk.KEY_minus || key == gdk.KEY_KP_Subtract {
		s.imageView.ZoomOut()
	} else if key == gdk.KEY_BackSpace {
		s.imageView.ZoomToFit()
	} else {
		return false
	}
	return true
}

func (s *Ui) UpdateCategories(categories *category.CategoriesCommand) {
	s.categories = make([]*common.Category, len(categories.GetCategories()))
	copy(s.categories, categories.GetCategories())

	s.topActionView.UpdateCategories(categories, s.imageView.GetCurrentHandle())
	s.sender.SendToTopic(event.ImageRequestCurrent)
}

func (s *Ui) UpdateCurrentImage() {
	s.imageView.UpdateCurrentImage()
}

func (s *Ui) SetImages(imageTarget event.Topic, handles []*common.ImageContainer) {
	if imageTarget == event.ImageRequestNext {
		s.imageView.AddImagesToNextStore(handles)
	} else if imageTarget == event.ImageRequestPrev {
		s.imageView.AddImagesToPrevStore(handles)
	} else if imageTarget == event.ImageRequestSimilar {
		s.similarImagesView.SetImages(handles, s.sender)
	}
}

func (s *Ui) SetCurrentImage(image *common.ImageContainer, index int, total int, title string, exifData *common.ExifData) {
	s.topActionView.SetCurrentStatus(index, total, title)
	s.bottomActionView.SetShowOnlyCategory(title != "")
	s.imageView.SetCurrentImage(image, exifData)
	s.UpdateCurrentImage()
	s.sendCurrentImageChangedEvent()

	s.imageCache.Purge()
}

func (s *Ui) sendCurrentImageChangedEvent() {
	s.sender.SendToTopicWithData(event.ImageChanged, s.imageView.GetCurrentHandle())
}

func (s *Ui) Run() {
	s.application.Run([]string{})
}

func (s *Ui) SetImageCategory(commands []*category.CategorizeCommand) {
	for _, button := range s.topActionView.GetCategoryButtons() {
		button.SetStatus(common.NONE)
	}

	for _, command := range commands {
		logger.Debug.Printf("Marked image category: '%s':%d", command.GetEntry().GetName(), command.GetOperation())

		if button, ok := s.topActionView.GetCategoryButton(command.GetEntry().GetId()); ok {
			button.SetStatus(command.GetOperation())
		}
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
	if len(s.castModal.GetDevices()) == 0 {
		s.castModal.SetNoDevices()
	}
	s.castModal.SearchDone()
}

func runAndProcessFolderChooser(folderChooser *gtk.FileChooserDialog, sender event.Sender) {
	response := folderChooser.Run()
	if response == gtk.RESPONSE_ACCEPT {
		folder := folderChooser.GetFilename()
		sender.SendToTopicWithData(event.DirectoryChanged, folder)
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

func (s *Ui) ShowEditCategoriesModal() {
	categories := make([]*common.Category, len(s.categories))
	copy(categories, s.categories)
	s.editCategoriesModal.Show(s.application.GetActiveWindow(), categories)
}

func (s *Ui) OpenFolderChooser(numOfButtons int) {
	folderChooser, err := createFileChooser(numOfButtons, s.application.GetActiveWindow())
	if err != nil {
		logger.Error.Panic("Can't open file chooser")
	}
	defer folderChooser.Destroy()

	runAndProcessFolderChooser(folderChooser, s.sender)
}

func (s *Ui) EnterFullScreenNoDistraction() {
	s.win.Fullscreen()
	s.fullscreen = true
	s.imageView.SetNoDistractionMode(true)
	s.topActionView.SetNoDistractionMode(true)
	s.bottomActionView.SetNoDistractionMode(true)
}

func (s *Ui) EnterFullScreen() {
	s.win.Fullscreen()
	s.fullscreen = true
	s.imageView.SetNoDistractionMode(false)
	s.topActionView.SetNoDistractionMode(false)
	s.bottomActionView.SetNoDistractionMode(false)
	s.bottomActionView.SetShowFullscreenButton(false)
}

func (s *Ui) ExitFullScreen() {
	s.win.Unfullscreen()
	s.fullscreen = false
	s.imageView.SetNoDistractionMode(false)
	s.topActionView.SetNoDistractionMode(false)
	s.bottomActionView.SetNoDistractionMode(false)
	s.bottomActionView.SetShowFullscreenButton(true)
}

func (s *Ui) FindDevices() {
	s.castModal.StartSearch(s.application.GetActiveWindow())
	s.sender.SendToTopic(event.CastDeviceSearch)
}
