package gtk

import (
	"github.com/AllenDang/giu"
	"time"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/common/logger"
	"vincit.fi/image-sorter/ui/giu/widget"
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
	currentImageTexture    *widget.TexturedImage
	nextImages             []*widget.TexturedImage
	previousImages         []*widget.TexturedImage
	categoryKeyManager     *CategoryKeyManager
	currentImageCategories map[apitype.CategoryId]bool

	api.Gui
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
		currentImageTexture: widget.NewEmptyTexturedImage(imageCache),
	}

	gui.categoryKeyManager = &CategoryKeyManager{
		callback: func(def *CategoryDef, stayOnImage bool, forceCategory bool) {
			operation := apitype.MOVE
			if _, ok := gui.currentImageCategories[def.categoryId]; ok {
				operation = apitype.NONE
			}

			broker.SendCommandToTopic(api.CategorizeImage, &api.CategorizeCommand{
				ImageId:         gui.currentImageTexture.Image.Id(),
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
			go s.currentImageTexture.LoadImageAsTexture(float32(newWidth), float32(newHeight))
			s.oldWidth = newWidth
			s.oldHeight = newHeight
		}

		renderStart := time.Now()

		var categories []giu.Widget
		for _, cat := range s.categories {
			text := cat.Name()

			_, active := s.currentImageCategories[cat.Id()]

			categoryId := cat.Id()
			categorizeButton := widget.CategoryButton(categoryId, text, active, func(command *api.CategorizeCommand) {
				command.NextImageDelay = 100
				command.ImageId = s.currentImageTexture.Image.Id()
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
					widget.ImageList(s.nextImages, false),
					widget.ResizableImage(s.currentImageTexture),
					giu.Dummy(-120, giu.Auto),
					widget.ImageList(s.previousImages, false),
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
		s.nextImages = []*widget.TexturedImage{}
		for _, data := range command.Images {
			ti := widget.NewTexturedImage(
				data.ImageFile(),
				float32(data.ImageData().Bounds().Dx()),
				float32(data.ImageData().Bounds().Dy()),
				s.imageCache,
			)
			s.nextImages = append(s.nextImages, ti)
			ti.LoadImageAsTextureThumbnail()
		}
		//s.imageView.AddImagesToNextStore(command.Images)
	} else if command.Topic == api.ImageRequestPrevious {
		s.previousImages = []*widget.TexturedImage{}
		for _, data := range command.Images {
			ti := widget.NewTexturedImage(
				data.ImageFile(),
				float32(data.ImageData().Bounds().Dx()),
				float32(data.ImageData().Bounds().Dy()),
				s.imageCache,
			)
			s.previousImages = append(s.previousImages, ti)
			ti.LoadImageAsTextureThumbnail()
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

	s.currentImageTexture.ChangeImage(
		command.Image.ImageFile(),
		float32(command.Image.ImageData().Bounds().Dx()),
		float32(command.Image.ImageData().Bounds().Dy()))
	width, height := s.win.GetSize()
	s.currentImageTexture.LoadImageAsTexture(float32(width), float32(height))
	s.sendCurrentImageChangedEvent()

	s.imageCache.Purge()
}

func (s *Ui) sendCurrentImageChangedEvent() {
	s.sender.SendCommandToTopic(api.ImageChanged, &api.ImageCategoryQuery{
		ImageId: s.currentImageTexture.Image.Id(),
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

func (s *Ui) ShowError(command *api.ErrorCommand) {
	logger.Error.Printf("Error: %s", command.Message)
	giu.Msgbox("Error", command.Message)
}
