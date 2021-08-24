package gtk

import (
	"errors"
	"fmt"
	"github.com/AllenDang/giu"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"image"
	"time"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/common/logger"
	"vincit.fi/image-sorter/ui/giu/widget"
	"vincit.fi/image-sorter/ui/gtk/component"
)

type Ui struct {
	// General
	win                    *giu.MasterWindow
	oldWidth, oldHeight    int
	fullscreen             bool
	imageCache             api.ImageStore
	sender                 api.Sender
	categories             []*apitype.Category
	rootPath               string
	currentImageTexture    *texturedImage
	nextImages             []*texturedImage
	previousImages         []*texturedImage
	categoryKeyManager     *CategoryKeyManager
	currentImageCategories map[apitype.CategoryId]bool

	api.Gui
	component.CallbackApi
}

type texturedImage struct {
	width          float32
	height         float32
	texture        *giu.Texture
	oldTexture     *giu.Texture
	imageId        apitype.ImageId
	oldImageId     apitype.ImageId
	lastWidth      int
	lastHeight     int
	newImageLoaded bool
	imageCache     api.ImageStore
}

func (s *texturedImage) changeImage(imageId apitype.ImageId, width float32, height float32) {
	s.oldTexture = s.texture
	s.newImageLoaded = false

	s.oldImageId = s.imageId
	s.imageId = imageId

	s.width = width
	s.height = height

	s.lastWidth = -1
	s.lastHeight = -1

}

func (s *texturedImage) loadImageAsTexture(width float32, height float32) *giu.Texture {
	if s.imageCache == nil {
		return nil
	}

	if s.newImageLoaded {
		if s.texture != nil && int(width) == s.lastWidth && int(height) == s.lastHeight {
			return s.texture
		}
	}

	s.lastWidth = int(width)
	s.lastHeight = int(height)

	scaledImage, _ := s.imageCache.GetScaled(s.imageId, apitype.SizeOf(s.lastWidth, s.lastHeight))
	if scaledImage == nil {
		s.texture = nil
	} else {
		go func() {
			var err error
			s.texture, err = giu.NewTextureFromRgba(scaledImage.(*image.RGBA))
			s.newImageLoaded = true
			if err != nil {
				logger.Error.Print(err)
			}
		}()
	}
	return s.texture
}

func (s *texturedImage) loadImageAsTextureThumbnail() *giu.Texture {
	if s.imageCache == nil {
		return nil
	}

	if s.newImageLoaded {
		if s.texture != nil {
			return s.texture
		}
	}

	scaledImage, _ := s.imageCache.GetThumbnail(s.imageId)
	if scaledImage == nil {
		s.texture = nil
	} else {
		go func() {
			var err error
			s.texture, err = giu.NewTextureFromRgba(scaledImage.(*image.RGBA))
			s.newImageLoaded = true
			if err != nil {
				logger.Error.Print(err)
			}
		}()
	}
	return s.texture
}

const (
	defaultWindowWidth  = 800
	defaultWindowHeight = 600
)

func NewUi(params *common.Params, broker api.Sender, imageCache api.ImageStore) api.Gui {
	gui := Ui{
		win:                 giu.NewMasterWindow("Image Sorter", defaultWindowWidth, defaultWindowHeight, 0),
		imageCache:          imageCache,
		sender:              broker,
		rootPath:            params.RootPath(),
		currentImageTexture: &texturedImage{imageCache: imageCache},
	}

	gui.categoryKeyManager = &CategoryKeyManager{
		callback: func(def *CategoryDef, stayOnImage bool, forceCategory bool) {
			operation := apitype.MOVE
			if _, ok := gui.currentImageCategories[def.categoryId]; ok {
				operation = apitype.NONE
			}

			broker.SendCommandToTopic(api.CategorizeImage, &api.CategorizeCommand{
				ImageId:         gui.currentImageTexture.imageId,
				CategoryId:      def.categoryId,
				Operation:       operation,
				StayOnSameImage: stayOnImage,
				NextImageDelay:  100,
				ForceToCategory: forceCategory,
			})
		},
	}

	gui.Init(params.RootPath())
	return &gui
}

func (s *Ui) Init(directory string) {
}

func (s *Ui) Run() {
	s.sender.SendCommandToTopic(api.DirectoryChanged, s.rootPath)
	s.win.Run(func() {
		newWidth, newHeight := s.win.GetSize()

		if newWidth != s.oldWidth || newHeight != s.oldHeight {
			if logger.IsLogLevel(logger.TRACE) {
				logger.Trace.Printf("Window size changed from (%d x %d) to (%d x %d)",
					s.oldWidth, s.oldHeight, newWidth, newHeight)
			}
			go s.currentImageTexture.loadImageAsTexture(float32(newWidth), float32(newHeight))
			s.oldWidth = newWidth
			s.oldHeight = newHeight
		}

		renderStart := time.Now()

		var nextImages []giu.Widget
		for _, nextImage := range s.nextImages {
			maxWidth := float32(120.0)
			height := nextImage.height / nextImage.width * maxWidth
			nextImages = append(nextImages, giu.Image(nextImage.texture).Size(maxWidth, height))
		}
		var previousImages []giu.Widget
		for _, previousImage := range s.previousImages {
			maxWidth := float32(120.0)
			height := previousImage.height / previousImage.width * maxWidth
			previousImages = append(previousImages, giu.Image(previousImage.texture).Size(maxWidth, height))
		}

		var categories []giu.Widget
		for _, cat := range s.categories {
			text := cat.Name()

			_, active := s.currentImageCategories[cat.Id()]

			categoryId := cat.Id()
			categorizeButton := widget.CategoryButton(categoryId, text, active, func(command *api.CategorizeCommand) {
				command.NextImageDelay = 100
				command.ImageId = s.currentImageTexture.imageId
				s.sender.SendCommandToTopic(api.CategorizeImage, command)
			})
			categories = append(categories, categorizeButton)
		}

		previousButton := giu.Button("Previous").OnClick(func() {
			s.sender.SendToTopic(api.ImageRequestPrevious)
		}).Size(120, 30)
		nextButton := giu.Button("Next").OnClick(func() {
			s.sender.SendToTopic(api.ImageRequestNext)
		}).Size(120, 30)

		giu.SingleWindow().
			Layout(
				giu.Row(
					previousButton,
					giu.Row(categories...),
					giu.Dummy(-120, 30),
					nextButton),
				giu.Separator(),
				giu.Row(
					giu.Column(nextImages...),
					giu.Custom(func() {
						widget.
							ResizableImage(
								s.currentImageTexture.texture,
								s.currentImageTexture.width,
								s.currentImageTexture.height,
							).Build()
					}),
					giu.Dummy(-120, giu.Auto),
					giu.Column(previousImages...),
				),
				giu.PrepareMsgbox(),
			)
		renderStop := time.Now()

		renderTime := renderStop.Sub(renderStart)
		if renderTime >= time.Millisecond && logger.IsLogLevel(logger.TRACE) {
			logger.Trace.Printf("Rendered UI in %s", renderTime)
		} else if renderTime >= 10*time.Millisecond {
			logger.Debug.Printf("Rendered UI in %s", renderTime)
		}
		s.handleKeyPress()
	})
}

func (s *Ui) handleKeyPress() bool {
	const bigJumpSize = 10
	const hugeJumpSize = 100

	shiftDown := giu.IsKeyDown(giu.KeyLeftShift) || giu.IsKeyDown(giu.KeyRightShift)
	altDown := giu.IsKeyDown(giu.KeyLeftAlt) || giu.IsKeyDown(giu.KeyRightAlt)
	controlDown := giu.IsKeyDown(giu.KeyLeftControl) || giu.IsKeyDown(giu.KeyRightControl)

	if giu.IsKeyPressed(giu.KeyF8) {
		logger.Debug.Printf("Find devices")
		s.FindDevices()
	}
	if giu.IsKeyPressed(giu.KeyF10) {
		logger.Debug.Printf("Show all images")
		s.sender.SendToTopic(api.ImageShowAll)
	}
	if giu.IsKeyPressed(giu.KeyEscape) {
		logger.Debug.Printf("Exit full screen")
		s.ExitFullScreen()
	}

	if giu.IsKeyPressed(giu.KeyF11) || (giu.IsKeyPressed(giu.KeyEnter) && altDown) {
		logger.Debug.Printf("Enter full screen")
		noDistractionMode := controlDown
		if noDistractionMode {
			s.EnterFullScreenNoDistraction()
		} else if s.fullscreen {
			s.ExitFullScreen()
		} else {
			s.EnterFullScreen()
		}
	}

	if giu.IsKeyPressed(giu.KeyF12) {
		logger.Debug.Printf("Find similar images")
		s.sender.SendToTopic(api.SimilarRequestSearch)
	}

	// Navigation

	if giu.IsKeyPressed(giu.KeyPageUp) {
		s.sender.SendCommandToTopic(api.ImageRequestPreviousOffset, &api.ImageAtQuery{Index: hugeJumpSize})
	}
	if giu.IsKeyPressed(giu.KeyPageUp) {
		s.sender.SendCommandToTopic(api.ImageRequestNextOffset, &api.ImageAtQuery{Index: hugeJumpSize})
	}

	if giu.IsKeyPressed(giu.KeyHome) {
		s.sender.SendCommandToTopic(api.ImageRequestAtIndex, &api.ImageAtQuery{Index: 0})
	}
	if giu.IsKeyPressed(giu.KeyEnd) {
		s.sender.SendCommandToTopic(api.ImageRequestAtIndex, &api.ImageAtQuery{Index: -1})
	}

	if giu.IsKeyPressed(giu.KeyLeft) {
		logger.Debug.Printf("Previous")
		if controlDown {
			s.sender.SendCommandToTopic(api.ImageRequestPreviousOffset, &api.ImageAtQuery{Index: bigJumpSize})
		} else {
			s.sender.SendToTopic(api.ImageRequestPrevious)
		}
	}
	if giu.IsKeyPressed(giu.KeyRight) {
		logger.Debug.Printf("Next")
		if controlDown {
			s.sender.SendCommandToTopic(api.ImageRequestNextOffset, &api.ImageAtQuery{Index: bigJumpSize})
		} else {
			s.sender.SendToTopic(api.ImageRequestNext)
		}
	}

	s.categoryKeyManager.HandleKeys(shiftDown, controlDown)
	/*else if command := s.topActionView.NewCommandForShortcut(key, s.imageView.CurrentImageFile()); command != nil {
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
	}*/
	return true
}

func (s *Ui) UpdateCategories(categories *api.UpdateCategoriesCommand) {
	s.categories = make([]*apitype.Category, len(categories.Categories))
	copy(s.categories, categories.Categories)

	s.categoryKeyManager.Reset(categories.Categories)
	//s.topActionView.UpdateCategories(categories)
	s.sender.SendToTopic(api.ImageRequestCurrent)
}

func (s *Ui) UpdateCurrentImage() {
	//s.imageView.UpdateCurrentImage()
}

func (s *Ui) SetImages(command *api.SetImagesCommand) {
	if command.Topic == api.ImageRequestNext {
		s.nextImages = []*texturedImage{}
		for _, data := range command.Images {
			ti := &texturedImage{
				imageId:    data.ImageFile().Id(),
				width:      float32(data.ImageData().Bounds().Dx()),
				height:     float32(data.ImageData().Bounds().Dy()),
				imageCache: s.imageCache,
			}
			s.nextImages = append(s.nextImages, ti)
			ti.loadImageAsTextureThumbnail()
		}
		//s.imageView.AddImagesToNextStore(command.Images)
	} else if command.Topic == api.ImageRequestPrevious {
		s.previousImages = []*texturedImage{}
		for _, data := range command.Images {
			ti := &texturedImage{
				imageId:    data.ImageFile().Id(),
				width:      float32(data.ImageData().Bounds().Dx()),
				height:     float32(data.ImageData().Bounds().Dy()),
				imageCache: s.imageCache,
			}
			s.previousImages = append(s.previousImages, ti)
			ti.loadImageAsTextureThumbnail()
		}
		//s.imageView.AddImagesToPrevStore(command.Images)
	} else if command.Topic == api.ImageRequestSimilar {
		//s.similarImagesView.SetImages(command.Images)
	}
}

func (s *Ui) SetCurrentImage(command *api.UpdateImageCommand) {
	//s.topActionView.SetCurrentStatus(command.Index, command.Total, command.CategoryId)
	//s.bottomActionView.SetShowOnlyCategory(command.CategoryId != -1)
	//s.imageView.SetCurrentImage(command.Image, command.MetaData)
	s.UpdateCurrentImage()

	imageId := command.Image.ImageFile().Id()
	s.currentImageTexture.changeImage(
		imageId,
		float32(command.Image.ImageData().Bounds().Dx()),
		float32(command.Image.ImageData().Bounds().Dy()))
	width, height := s.win.GetSize()
	s.currentImageTexture.loadImageAsTexture(float32(width), float32(height))
	s.sendCurrentImageChangedEvent()

	s.imageCache.Purge()
}

func (s *Ui) sendCurrentImageChangedEvent() {
	s.sender.SendCommandToTopic(api.ImageChanged, &api.ImageCategoryQuery{
		ImageId: s.currentImageTexture.imageId,
	})
}

func (s *Ui) SetImageCategory(command *api.CategoriesCommand) {
	s.currentImageCategories = map[apitype.CategoryId]bool{}

	for _, category := range command.Categories {
		logger.Debug.Printf("Marked image category: '%s'", category.Name())
		s.currentImageCategories[category.Id()] = true
	}

	//for _, button := range s.topActionView.GetCategoryButtons() {
	//	button.SetStatus(apitype.NONE)
	//}
	//
	//for _, category := range command.Categories {
	//	logger.Debug.Printf("Marked image category: '%s'", category.Name())
	//
	//	if button, ok := s.topActionView.GetCategoryButton(category.Id()); ok {
	//		button.SetStatus(apitype.MOVE)
	//	}
	//}
}

func (s *Ui) UpdateProgress(command *api.UpdateProgressCommand) {
	//if command.Current == 0 {
	//	s.progressView.SetVisible(true)
	//	s.topActionView.SetVisible(false)
	//	s.bottomActionView.SetVisible(false)
	//}
	//
	//if command.Current == command.Total {
	//	s.progressView.SetVisible(false)
	//	s.topActionView.SetVisible(true)
	//	s.bottomActionView.SetVisible(true)
	//} else {
	//	s.progressView.SetStatus(command.Name, command.Current, command.Total)
	//}
}

func (s *Ui) DeviceFound(command *api.DeviceFoundCommand) {
	//s.castModal.AddDevice(command.DeviceName)
}

func (s *Ui) CastReady() {
	s.sendCurrentImageChangedEvent()
}
func (s *Ui) CastFindDone() {
	//if len(s.castModal.Devices()) == 0 {
	//	s.castModal.SetNoDevices()
	//}
	//s.castModal.SearchDone()
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
}

func (s *Ui) OpenFolderChooser(numOfButtons int) {
}

func (s *Ui) EnterFullScreenNoDistraction() {
	s.fullscreen = true
	//s.imageView.SetNoDistractionMode(true)
	//s.topActionView.SetNoDistractionMode(true)
	//s.bottomActionView.SetNoDistractionMode(true)
}

func (s *Ui) EnterFullScreen() {
	s.fullscreen = true
	//s.imageView.SetNoDistractionMode(false)
	//s.topActionView.SetNoDistractionMode(false)
	//s.bottomActionView.SetNoDistractionMode(false)
	//s.bottomActionView.SetShowFullscreenButton(false)
}

func (s *Ui) ExitFullScreen() {
	s.fullscreen = false
	//s.imageView.SetNoDistractionMode(false)
	//s.topActionView.SetNoDistractionMode(false)
	//s.bottomActionView.SetNoDistractionMode(false)
	//s.bottomActionView.SetShowFullscreenButton(true)
}

func (s *Ui) FindDevices() {
	// s.castModal.StartSearch(s.application.GetActiveWindow())
	s.sender.SendToTopic(api.CastDeviceSearch)
}

func resolveModifierStatuses(keyEvent *gdk.EventKey) (shiftDown bool, controlDown bool, altDown bool) {
	modifiers := gtk.AcceleratorGetDefaultModMask()
	state := gdk.ModifierType(keyEvent.State())
	modifierType := state & modifiers

	shiftDown = modifierType&gdk.SHIFT_MASK > 0
	controlDown = modifierType&gdk.CONTROL_MASK > 0
	altDown = modifierType&gdk.MOD1_MASK > 0

	logger.Trace.Printf("Modifiers: Shift = %t, CTRL = %t, ALT = %t", shiftDown, controlDown, altDown)
	return
}

func (s *Ui) ShowError(command *api.ErrorCommand) {
	logger.Error.Printf("Error: %s", command.Message)
	giu.Msgbox("Error", command.Message)
}
