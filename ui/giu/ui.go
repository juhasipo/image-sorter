package gtk

import (
	"fmt"
	"github.com/AllenDang/giu"
	"time"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/common/logger"
	"vincit.fi/image-sorter/ui/giu/guiapi"
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
	currentCategoryId      apitype.CategoryId
	currentProgress        progress
	showCategoryEditModal  bool
	categoryEditWidget     *widget.CategoryEditWidget

	api.Gui
}

type progress struct {
	open     bool
	label    string
	position int
	max      int
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
		currentProgress: progress{
			open:     false,
			label:    "",
			position: 1,
			max:      1,
		},
	}

	gui.categoryEditWidget = widget.CategoryEdit(
		func(asDefault bool, categories []*apitype.Category) {
			if asDefault {
				broker.SendCommandToTopic(api.CategoriesSaveDefault, &api.SaveCategoriesCommand{Categories: categories})
			} else {
				broker.SendCommandToTopic(api.CategoriesSave, &api.SaveCategoriesCommand{Categories: categories})
			}
			gui.showCategoryEditModal = false
		},
		func() {
			gui.showCategoryEditModal = false
		},
	)

	gui.categoryKeyManager = &CategoryKeyManager{
		callback: func(def *CategoryDef, action *guiapi.CategoryAction) {
			operation := apitype.MOVE
			if !action.ForceCategory {
				if _, ok := gui.currentImageCategories[def.categoryId]; ok {
					operation = apitype.NONE
				}
			}

			if action.ShowOnlyCategory {
				if gui.currentCategoryId == apitype.NoCategory {
					broker.SendCommandToTopic(api.CategoriesShowOnly, &api.SelectCategoryCommand{
						CategoryId: def.categoryId,
					})
				} else {
					broker.SendToTopic(api.ImageShowAll)
				}
			} else {
				broker.SendCommandToTopic(api.CategorizeImage, &api.CategorizeCommand{
					ImageId:         gui.currentImageTexture.Image.Id(),
					CategoryId:      def.categoryId,
					Operation:       operation,
					StayOnSameImage: action.StayOnImage,
					NextImageDelay:  100,
					ForceToCategory: action.ForceCategory,
				})
			}
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

		var categories []*widget.CategoryButtonWidget
		for _, cat := range s.categories {
			categoryId := cat.Id()
			text := cat.Name()

			_, active := s.currentImageCategories[cat.Id()]

			highlight := s.currentCategoryId == categoryId
			onClick := func(action *guiapi.CategoryAction) {
				s.categoryKeyManager.HandleCategory(categoryId, action)
			}
			categorizeButton := widget.CategoryButton(categoryId, text, active, highlight, onClick)
			categories = append(categories, categorizeButton)
		}

		previousButton := giu.Button("Previous").OnClick(func() {
			s.sender.SendToTopic(api.ImageRequestPrevious)
		}).Size(120, 30)
		nextButton := giu.Button("Next").OnClick(func() {
			s.sender.SendToTopic(api.ImageRequestNext)
		}).Size(120, 30)

		var modal giu.Widget = giu.Row()
		if s.currentProgress.open {
			modal = giu.PopupModal("Progress").
				Flags(giu.WindowFlagsAlwaysAutoResize|giu.WindowFlagsNoTitleBar).
				Layout(
					giu.Label(s.currentProgress.label),
					giu.Row(
						giu.ProgressBar(float32(s.currentProgress.position)/float32(s.currentProgress.max)).
							Overlay(fmt.Sprintf("%d/%d", s.currentProgress.position, s.currentProgress.max)),
						giu.Button("Cancel").
							OnClick(func() {
								s.sender.SendToTopic(api.SimilarRequestStop)
							}),
					),
					giu.Custom(func() {
						if !s.currentProgress.open {
							giu.CloseCurrentPopup()
						}
					}))
		}

		mainWindow := giu.SingleWindow()
		if s.showCategoryEditModal {
			mainWindow.
				Layout(s.categoryEditWidget)
			s.categoryEditWidget.HandleKeys()
		} else {
			mainWindow.Layout(
				giu.Row(
					previousButton,
					widget.CategoryButtonView(categories),
					giu.Dummy(-120, 30),
					nextButton),
				giu.Separator(),
				modal,
				giu.Custom(func() {
					if s.currentProgress.open {
						giu.OpenPopup("Progress")
					}
				}),
				giu.Custom(func() {
					_, height := giu.GetAvailableRegion()
					h := height - 30
					giu.Column(giu.Row(
						widget.ImageList(s.nextImages, false, h),
						widget.ResizableImage(s.currentImageTexture),
						giu.Dummy(-120, h),
						widget.ImageList(s.previousImages, false, h),
					)).Build()
				}),
				giu.Separator(),
				giu.Row(
					giu.Button("Edit categories").OnClick(func() {
						s.categoryEditWidget.SetCategories(s.categories)
						s.showCategoryEditModal = true
					}), giu.Button("Search similar").OnClick(func() {
						s.sender.SendToTopic(api.SimilarRequestSearch)
					}),
				),
				giu.PrepareMsgbox(),
			)

			// Ignore all input when the progress bar is shown
			// This prevents any unexpected changes
			if !s.currentProgress.open {
				s.handleKeyPress()
			}
		}

		renderStop := time.Now()

		renderTime := renderStop.Sub(renderStart)
		if renderTime >= time.Millisecond && logger.IsLogLevel(logger.TRACE) {
			logger.Trace.Printf("Rendered UI in %s", renderTime)
		} else if renderTime >= 10*time.Millisecond {
			logger.Debug.Printf("Rendered UI in %s", renderTime)
		}
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

	s.categoryKeyManager.HandleKeys(&guiapi.CategoryAction{
		StayOnImage:      shiftDown,
		ForceCategory:    controlDown,
		ShowOnlyCategory: altDown,
	})
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

	s.categoryEditWidget.SetCategories(s.categories)
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
	s.currentCategoryId = command.CategoryId
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
	if command.Current == command.Total {
		s.currentProgress.open = false
		s.currentProgress.label = ""
		s.currentProgress.position = 1
		s.currentProgress.max = 1
	} else {
		s.currentProgress.open = true
		s.currentProgress.label = command.Name
		s.currentProgress.position = command.Current
		s.currentProgress.max = command.Total
	}
	giu.Update()
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
