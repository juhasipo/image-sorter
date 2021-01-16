package gtk

import (
	"errors"
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/common/logger"
	"vincit.fi/image-sorter/ui/gtk/component"
)

type Ui struct {
	// General
	win         *gtk.ApplicationWindow
	fullscreen  bool
	application *gtk.Application
	imageCache  api.ImageStore
	sender      api.Sender
	categories  []*apitype.Category
	rootPath    string

	// UI components
	progressView        *component.ProgressView
	topActionView       *component.TopActionView
	imageView           *component.ImageView
	similarImagesView   *component.SimilarImagesView
	bottomActionView    *component.BottomActionView
	castModal           *component.CastModal
	editCategoriesModal *component.CategoryModal

	api.Gui
	component.CallbackApi
}

const (
	defaultWindowWidth  = 800
	defaultWindowHeight = 600
)

func NewUi(params *common.Params, broker api.Sender, imageCache api.ImageStore) api.Gui {

	// Create Gtk Application, change appID to your application domain name reversed.
	const appID = "fi.vincit.imagesorter"
	application, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)

	// Check to make sure no errors when creating Gtk Application
	if err != nil {
		logger.Error.Fatal("Could not create application.", err)
	}

	gui := Ui{
		application: application,
		imageCache:  imageCache,
		sender:      broker,
		rootPath:    params.GetRootPath(),
	}

	gui.Init(params.GetRootPath())
	return &gui
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
		s.win.SetSizeRequest(defaultWindowWidth, defaultWindowHeight)
		s.win.Connect("key_press_event", s.handleKeyPress)

		s.similarImagesView = component.NewSimilarImagesView(builder, s.sender, s.imageCache)
		s.imageView = component.NewImageView(builder, s.sender, s.imageCache)
		s.topActionView = component.NewTopActions(builder, s.imageView, s.sender)
		s.bottomActionView = component.NewBottomActions(builder, s.application.GetActiveWindow(), s, s.sender)
		s.progressView = component.NewProgressView(builder, s.sender)

		s.castModal = component.NewCastModal(builder, s, s.sender)
		s.editCategoriesModal = component.NewCategoryModal(builder, s.sender)

		if directory == "" {
			s.OpenFolderChooser(1)
		} else {
			s.sender.SendCommandToTopic(api.DirectoryChanged, directory)
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

	shiftDown, controlDown, altDown := resolveModifierStatuses(keyEvent)

	if key == gdk.KEY_F8 {
		s.FindDevices()
	} else if key == gdk.KEY_F10 {
		s.sender.SendToTopic(api.ImageShowAll)
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
		s.sender.SendToTopic(api.SimilarRequestSearch)
	} else if key == gdk.KEY_Page_Up {
		s.sender.SendCommandToTopic(api.ImageRequestPrevOffset, &api.ImageAtQuery{Index: hugeJumpSize})
	} else if key == gdk.KEY_Page_Down {
		s.sender.SendCommandToTopic(api.ImageRequestNextOffset, &api.ImageAtQuery{Index: hugeJumpSize})
	} else if key == gdk.KEY_Home {
		s.sender.SendCommandToTopic(api.ImageRequestAtIndex, &api.ImageAtQuery{Index: 0})
	} else if key == gdk.KEY_End {
		s.sender.SendCommandToTopic(api.ImageRequestAtIndex, &api.ImageAtQuery{Index: -1})
	} else if key == gdk.KEY_Left {
		if controlDown {
			s.sender.SendCommandToTopic(api.ImageRequestPrevOffset, &api.ImageAtQuery{Index: bigJumpSize})
		} else {
			s.sender.SendToTopic(api.ImageRequestPrev)
		}
	} else if key == gdk.KEY_Right {
		if controlDown {
			s.sender.SendCommandToTopic(api.ImageRequestNextOffset, &api.ImageAtQuery{Index: bigJumpSize})
		} else {
			s.sender.SendToTopic(api.ImageRequestNext)
		}
	} else if command := s.topActionView.NewCommandForShortcut(key, s.imageView.GetCurrentImageFile()); command != nil {
		switchToCategory := altDown
		if switchToCategory {
			s.sender.SendCommandToTopic(api.CategoriesShowOnly, &api.SelectCategoryCommand{CategoryId: command.CategoryId})
		} else {
			command.StayOnSameImage = shiftDown
			command.ForceToCategory = controlDown
			s.sender.SendCommandToTopic(api.CategorizeImage, command)
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

func (s *Ui) UpdateCategories(categories *api.UpdateCategoriesCommand) {
	s.categories = make([]*apitype.Category, len(categories.Categories))
	copy(s.categories, categories.Categories)

	s.topActionView.UpdateCategories(categories)
	s.sender.SendToTopic(api.ImageRequestCurrent)
}

func (s *Ui) UpdateCurrentImage() {
	s.imageView.UpdateCurrentImage()
}

func (s *Ui) SetImages(command *api.SetImagesCommand) {
	if command.Topic == api.ImageRequestNext {
		s.imageView.AddImagesToNextStore(command.Images)
	} else if command.Topic == api.ImageRequestPrev {
		s.imageView.AddImagesToPrevStore(command.Images)
	} else if command.Topic == api.ImageRequestSimilar {
		s.similarImagesView.SetImages(command.Images)
	}
}

func (s *Ui) SetCurrentImage(command *api.UpdateImageCommand) {
	s.topActionView.SetCurrentStatus(command.Index, command.Total, command.CategoryId)
	s.bottomActionView.SetShowOnlyCategory(command.CategoryId != -1)
	s.imageView.SetCurrentImage(command.Image)
	s.UpdateCurrentImage()
	s.sendCurrentImageChangedEvent()

	s.imageCache.Purge()
}

func (s *Ui) sendCurrentImageChangedEvent() {
	s.sender.SendCommandToTopic(api.ImageChanged, &api.ImageCategoryQuery{
		ImageId: s.imageView.GetCurrentImageFile().GetId(),
	})
}

func (s *Ui) Run() {
	s.application.Run([]string{})
}

func (s *Ui) SetImageCategory(command *api.CategoriesCommand) {
	for _, button := range s.topActionView.GetCategoryButtons() {
		button.SetStatus(apitype.NONE)
	}

	for _, category := range command.Categories {
		logger.Debug.Printf("Marked image category: '%s'", category.GetName())

		if button, ok := s.topActionView.GetCategoryButton(category.GetId()); ok {
			button.SetStatus(apitype.MOVE)
		}
	}
}

func (s *Ui) UpdateProgress(command *api.UpdateProgressCommand) {
	if command.Current == 0 {
		s.progressView.SetVisible(true)
		s.topActionView.SetVisible(false)
		s.bottomActionView.SetVisible(false)
	}

	if command.Current == command.Total {
		s.progressView.SetVisible(false)
		s.topActionView.SetVisible(true)
		s.bottomActionView.SetVisible(true)
	} else {
		s.progressView.SetStatus(command.Current, command.Total)
	}
}

func (s *Ui) DeviceFound(command *api.DeviceFoundCommand) {
	s.castModal.AddDevice(command.DeviceName)
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

func runAndProcessFolderChooser(folderChooser *gtk.FileChooserDialog, sender api.Sender) {
	response := folderChooser.Run()
	if response == gtk.RESPONSE_ACCEPT {
		folder := folderChooser.GetFilename()
		sender.SendCommandToTopic(api.DirectoryChanged, folder)
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
	categories := make([]*apitype.Category, len(s.categories))
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
	s.sender.SendToTopic(api.CastDeviceSearch)
}

func resolveModifierStatuses(keyEvent *gdk.EventKey) (shiftDown bool, controlDown bool, altDown bool) {
	modifiers := gtk.AcceleratorGetDefaultModMask()
	state := gdk.ModifierType(keyEvent.State())
	modifierType := state & modifiers

	shiftDown = modifierType&gdk.GDK_SHIFT_MASK > 0
	controlDown = modifierType&gdk.GDK_CONTROL_MASK > 0
	altDown = modifierType&gdk.GDK_MOD1_MASK > 0

	logger.Trace.Printf("Modifiers: Shift = %t, CTRL = %t, ALT = %t", shiftDown, controlDown, altDown)
	return
}

func (s *Ui) ShowError(command *api.ErrorCommand) {
	logger.Error.Printf("Error: %s", command.Message)
	errorDialog := gtk.MessageDialogNew(s.win, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, command.Message)
	errorDialog.SetTitle("Error")
	errorDialog.Run()
	errorDialog.Destroy()
}
