package gtk

import (
	"fmt"
	"github.com/AllenDang/giu"
	"github.com/OpenDiablo2/dialog"
	"image/color"
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
	currentProgress        progressModal
	deviceModal            deviceModal
	showCategoryEditModal  bool
	categoryEditWidget     *widget.CategoryEditWidget

	nextImagesList     *widget.HorizontalImageListWidget
	previousImagesList *widget.HorizontalImageListWidget
	widthInNumOfImage  int

	api.Gui
}

type progressModal struct {
	open      bool
	label     string
	position  int
	max       int
	canCancel bool
}

type deviceModal struct {
	open           bool
	label          string
	inProgress     bool
	devices        []string
	selectedIndex  int
	showBackground bool
}

const (
	defaultWindowWidth  = 800
	defaultWindowHeight = 600
)

func NewUi(params *common.Params, broker api.Sender, imageCache api.ImageStore) api.Gui {
	onImageSelected := func(imageFile *apitype.ImageFile) {
		broker.SendCommandToTopic(api.ImageRequest, &api.ImageQuery{
			Id: imageFile.Id(),
		})
	}

	gui := Ui{
		win:                 giu.NewMasterWindow("Image Sorter", defaultWindowWidth, defaultWindowHeight, 0),
		imageCache:          imageCache,
		sender:              broker,
		rootPath:            params.RootPath(),
		currentImageTexture: widget.NewEmptyTexturedImage(imageCache),
		currentProgress: progressModal{
			open:     false,
			label:    "",
			position: 1,
			max:      1,
		},
		deviceModal: deviceModal{
			open:           false,
			label:          "Cast to ChromeCast",
			inProgress:     false,
			devices:        []string{},
			showBackground: true,
		},
		nextImagesList:     widget.HorizontalImageList(onImageSelected, false, true),
		previousImagesList: widget.HorizontalImageList(onImageSelected, false, false),
		widthInNumOfImage:  0,
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
			operation := apitype.CATEGORIZE
			if !action.ForceCategory {
				if _, ok := gui.currentImageCategories[def.categoryId]; ok {
					operation = apitype.UNCATEGORIZE
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

	dialog.Init()

	return &gui
}

func (s *Ui) Init(directory string) {
	if directory == "" {
		logger.Debug.Printf("No root directory specified, open dialog")
		var err error
		if directory, err = dialog.Directory().Title("Choose Image Directory").Browse(); err != nil {
			logger.Error.Fatal("Error while trying to load directory", err)
		}
	}
	logger.Debug.Printf("Opening directory '%s'", directory)

	s.sender.SendCommandToTopic(api.DirectoryChanged, directory)
}

func (s *Ui) Run() {
	s.Init(s.rootPath)
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

			// FIXME: Potential concurrent write to map
			_, active := s.currentImageCategories[cat.Id()]

			highlight := s.currentCategoryId == categoryId
			onClick := func(action *guiapi.CategoryAction) {
				s.categoryKeyManager.HandleCategory(categoryId, action)
			}
			categorizeButton := widget.CategoryButton(categoryId, text, cat.ShortcutAsString(), active, highlight, onClick)
			categories = append(categories, categorizeButton)
		}

		var progressModalWidget giu.Widget
		progressModalWidget = giu.PopupModal("Progress").
			Flags(giu.WindowFlagsAlwaysAutoResize|giu.WindowFlagsNoTitleBar).
			Layout(
				giu.Label(s.currentProgress.label),
				giu.Row(
					giu.ProgressBar(float32(s.currentProgress.position)/float32(s.currentProgress.max)).
						Overlay(fmt.Sprintf("%d/%d", s.currentProgress.position, s.currentProgress.max)),
					giu.Button("Cancel").
						Disabled(!s.currentProgress.canCancel).
						OnClick(func() {
							s.sender.SendToTopic(api.SimilarRequestStop)
						}),
				),
				giu.Custom(func() {
					if !s.currentProgress.open {
						logger.Trace.Printf("Hide progress modal")
						giu.CloseCurrentPopup()
					}
				}))

		var deviceModalWidget giu.Widget
		deviceModalWidget = giu.PopupModal("Device").
			Flags(giu.WindowFlagsAlwaysAutoResize|giu.WindowFlagsNoDecoration).
			Layout(
				giu.Label(s.deviceModal.label),
				giu.Custom(func() {
					if s.deviceModal.inProgress {
						giu.ProgressIndicator("Searching...", 10, 10, 5).Build()
					}
				}),
				giu.ListBox("Cast Devices", s.deviceModal.devices).
					Size(300, 200).
					OnChange(func(selectedIndex int) {
						s.deviceModal.selectedIndex = selectedIndex
					}),
				giu.Checkbox("Show background", &s.deviceModal.showBackground),
				giu.Row(
					giu.Button("OK##SelectDevice").
						Disabled(len(s.deviceModal.devices) == 0).
						OnClick(func() {
							s.sender.SendCommandToTopic(api.CastDeviceSelect, &api.SelectDeviceCommand{
								Name:           s.deviceModal.devices[s.deviceModal.selectedIndex],
								ShowBackground: s.deviceModal.showBackground,
							})
							s.deviceModal.open = false
							s.deviceModal.inProgress = false
						}),
					giu.Button("Cancel##CancelDevice").
						OnClick(func() {
							s.sender.SendToTopic(api.SimilarRequestStop)
							s.deviceModal.open = false
							s.deviceModal.inProgress = false
						}),
				),
				giu.Custom(func() {
					if !s.deviceModal.open {
						logger.Trace.Printf("Hide device list modal")
						giu.CloseCurrentPopup()
					}
				}))

		mainWindow := giu.SingleWindow()
		if s.showCategoryEditModal {
			mainWindow.
				Layout(s.categoryEditWidget)
			s.categoryEditWidget.HandleKeys()
		} else {
			bottomHeight := float32(30.0)
			mainWindow.Layout(
				giu.Custom(func() {
					width, _ := giu.GetAvailableRegion()
					height := float32(60)
					buttonWidth := float32(15)
					centerPieceWidth := float32(120)
					listWidth := (width - buttonWidth*2 - centerPieceWidth) / 2

					widthInNumOfImage := int(listWidth/60) + 1

					if widthInNumOfImage != s.widthInNumOfImage {
						s.sender.SendCommandToTopic(api.ImageListSizeChanged, &api.ImageListCommand{
							ImageListSize: widthInNumOfImage,
						})
					}
					s.widthInNumOfImage = widthInNumOfImage

					giu.PushItemSpacing(0, 0)
					pButton := giu.Button("<").
						OnClick(func() {
							s.sender.SendToTopic(api.ImageRequestPrevious)
						}).
						Size(buttonWidth, height)
					nButton := giu.Button(">").
						OnClick(func() {
							s.sender.SendToTopic(api.ImageRequestNext)
						}).
						Size(buttonWidth, height)
					giu.Row(
						s.nextImagesList.Size(listWidth, height).SetImages(s.nextImages),
						pButton,
						widget.ResizableImage(s.currentImageTexture).Size(120, height),
						nButton,
						s.previousImagesList.Size(listWidth, height).SetImages(s.previousImages),
					).Build()
					giu.PopStyle()
				}),
				giu.Row(
					widget.CategoryButtonView(categories),
				),
				progressModalWidget,
				deviceModalWidget,
				giu.Custom(func() {
					if s.currentProgress.open {
						logger.Trace.Printf("Show progress modal")
						giu.OpenPopup("Progress")
					}
				}),
				giu.Custom(func() {
					width, height := giu.GetAvailableRegion()
					h := height - bottomHeight

					width = width - 30.0*2

					pButton := giu.Button("<").
						OnClick(func() {
							s.sender.SendToTopic(api.ImageRequestPrevious)
						}).
						Size(30, h)
					nButton := giu.Button(">").
						OnClick(func() {
							s.sender.SendToTopic(api.ImageRequestNext)
						}).
						Size(30, h)

					giu.Style().
						SetStyle(giu.StyleVarItemSpacing, 0, 0).
						SetColor(giu.StyleColorBorder, color.RGBA{0, 0, 0, 255}).
						SetColor(giu.StyleColorChildBg, color.RGBA{0, 0, 0, 255}).
						To(
							giu.Row(
								pButton,
								giu.Child().
									Size(width, h).
									Border(true).
									Layout(widget.ResizableImage(s.currentImageTexture)),
								nButton,
							),
						).Build()
				}),
				giu.Row(
					giu.Dummy(0, bottomHeight),
					giu.Button("Edit categories").OnClick(func() {
						s.categoryEditWidget.SetCategories(s.categories)
						s.showCategoryEditModal = true
					}),
					giu.Button("Search similar").OnClick(func() {
						s.sender.SendToTopic(api.SimilarRequestSearch)
					}),
					giu.Button("Cast").OnClick(func() {
						giu.OpenPopup("Device")
						s.FindDevices()
					}),
					giu.Button("Open directory").OnClick(func() {
						s.Init("")
					}),
					giu.Button("Apply Categories").OnClick(func() {
						result := dialog.Message("Do you really want to persist these categories?").YesNo()
						if result {
							s.sender.SendCommandToTopic(api.CategoryPersistAll, &api.PersistCategorizationCommand{
								KeepOriginals:  true,
								FixOrientation: false,
								Quality:        100,
							})
						}
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
		}
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
		}
	} else if command.Topic == api.ImageRequestSimilar {
		//s.similarImagesView.SetImages(command.Images)
	}
	giu.Update()
}

func (s *Ui) SetCurrentImage(command *api.UpdateImageCommand) {
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
	giu.Update()
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
	//	button.SetStatus(apitype.UNCATEGORIZE)
	//}
	//
	//for _, category := range command.Categories {
	//	logger.Debug.Printf("Marked image category: '%s'", category.Name())
	//
	//	if button, ok := s.topActionView.GetCategoryButton(category.Id()); ok {
	//		button.SetStatus(apitype.CATEGORIZE)
	//	}
	//}
}

func (s *Ui) UpdateProgress(command *api.UpdateProgressCommand) {
	if command.Current == command.Total {
		logger.Trace.Printf("Progress '%s' completed", command.Name)
		s.currentProgress.open = false
		s.currentProgress.label = ""
		s.currentProgress.position = 1
		s.currentProgress.max = 1
		s.currentProgress.canCancel = false
	} else {
		logger.Trace.Printf("Update progress '%s' %d/%d", command.Name, command.Current, command.Total)
		s.currentProgress.open = true
		s.currentProgress.label = command.Name
		s.currentProgress.position = command.Current
		s.currentProgress.max = command.Total
		s.currentProgress.canCancel = command.CanCancel
	}
	giu.Update()
}

func (s *Ui) DeviceFound(command *api.DeviceFoundCommand) {
	s.deviceModal.devices = append(s.deviceModal.devices, command.DeviceName)
}

func (s *Ui) CastReady() {
	s.sendCurrentImageChangedEvent()
}

func (s *Ui) CastFindDone() {
	s.deviceModal.inProgress = true
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
	s.deviceModal.devices = []string{}
	s.deviceModal.inProgress = true
	s.deviceModal.open = true
	s.sender.SendToTopic(api.CastDeviceSearch)
}

func (s *Ui) ShowError(command *api.ErrorCommand) {
	logger.Error.Printf("Error: %s", command.Message)
	giu.Msgbox("Error", command.Message)
}
