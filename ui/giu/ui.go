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
	currentImageWidget     *widget.ResizableImageWidget
	nextImages             []*widget.TexturedImage
	previousImages         []*widget.TexturedImage
	similarImages          []*widget.TexturedImage
	categoryKeyManager     *CategoryKeyManager
	currentImageCategories map[apitype.CategoryId]bool
	currentCategoryId      apitype.CategoryId
	progressModal          progressModal
	progressBackground     progressModal
	deviceModal            deviceModal
	applyChangesModal      applyChangesModal
	showCategoryEditModal  bool
	categoryEditWidget     *widget.CategoryEditWidget

	nextImagesList     *widget.HorizontalImageListWidget
	previousImagesList *widget.HorizontalImageListWidget
	similarImagesList  *widget.HorizontalImageListWidget
	similarImagesShown bool
	widthInNumOfImage  int
	zoomLevel          int32

	everythingLoaded bool
	api.Gui
}

const defaultZoomLevel = 0

var zoomLevels = []float32{-1, 0.01, 0.05, 0.1, 0.25, 0.5, 0.75, 1, 1.25, 1.50, 1.75, 2, 3, 4, 5, 10, 15}
var zoomLevelItems = []string{"Fit", "1 %", "5 %", "10 %", "25 %", "50 %", "75 %", "100 %", "125 %", "150 %", "175 %", "200 %", "300 %", "400 %", "500 %", "1000 %", "1500 %"}

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

type applyChangesModal struct {
	open           bool
	label          string
	keepOriginals  bool
	fixOrientation bool
	quality        int32
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
		progressModal: progressModal{
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
		applyChangesModal: applyChangesModal{
			open:           false,
			label:          "Apply changes",
			keepOriginals:  true,
			fixOrientation: false,
			quality:        90,
		},
		nextImagesList:     widget.HorizontalImageList(onImageSelected, false, false, true),
		previousImagesList: widget.HorizontalImageList(onImageSelected, false, true, true),
		similarImagesList:  widget.HorizontalImageList(onImageSelected, false, false, false),
		similarImagesShown: false,
		widthInNumOfImage:  0,
		zoomLevel:          defaultZoomLevel,
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

func (s *Ui) Pause() {
	logger.Debug.Printf("UI waiting for backend services...")
	s.everythingLoaded = false
}

func (s *Ui) Ready() {
	logger.Debug.Printf("Backend services ready, UI can be started")
	s.everythingLoaded = true
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
		mainWindow := giu.SingleWindow()
		if !s.everythingLoaded {
			mainWindow.Layout(giu.Label("Loading services..."))
			return
		}
		newWidth, newHeight := s.win.GetSize()

		if newWidth != s.oldWidth || newHeight != s.oldHeight {
			if logger.IsLogLevel(logger.TRACE) {
				logger.Trace.Printf("Window size changed from (%d x %d) to (%d x %d)",
					s.oldWidth, s.oldHeight, newWidth, newHeight)
			}
			go s.currentImageTexture.LoadImageAsTexture(float32(newWidth), float32(newHeight), s.getZoomFactor())
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

		if s.showCategoryEditModal {
			mainWindow.
				Layout(s.categoryEditWidget)
			s.categoryEditWidget.HandleKeys()
		} else {
			progressHeight := float32(20.0)
			actionsHeight := float32(30.0)
			similarImagesHeight := float32(60.0)
			paddings := float32(8.0)
			imageName := ""
			imageInfo := ""
			var highlightedImage *apitype.ImageFileAndData
			if s.nextImagesList.HighlightedImage() != nil {
				highlightedImage = s.nextImagesList.HighlightedImage()
			} else if s.previousImagesList.HighlightedImage() != nil {
				highlightedImage = s.previousImagesList.HighlightedImage()
			} else if s.currentImageTexture.Image != nil {
				highlightedImage = s.currentImageTexture.ImageFileAndData
			}

			if highlightedImage != nil {
				imageName = highlightedImage.ImageFile().FileName()
				imageInfo = fmt.Sprintf("(%d x %d)",
					s.currentImageTexture.ImageFileAndData.ImageData().Bounds().Dx(),
					s.currentImageTexture.ImageFileAndData.ImageData().Bounds().Dy(),
				)
			}
			mainWindow.Layout(
				s.imagesWidget(),
				giu.Row(
					giu.Label(imageName),
					giu.Condition(imageInfo != "", giu.Layout{giu.Label(imageInfo)}, giu.Layout{giu.Label("")}),
				),
				giu.Row(
					widget.CategoryButtonView(categories),
				),
				// Modals
				getProgressModal("ProgressModal", s.sender, &s.progressModal),
				getDeviceModal("DeviceModal", s.sender, &s.deviceModal),
				getApplyChangesModal("ApplyCategoriesModal", s.sender, &s.applyChangesModal),
				giu.Custom(func() {
					// Process modal states
					if s.progressModal.open {
						giu.OpenPopup("ProgressModal")
					}
					if s.deviceModal.open {
						giu.OpenPopup("DeviceModal")
					}
					if s.applyChangesModal.open {
						giu.OpenPopup("ApplyCategoriesModal")
					}
				}),
				s.mainImageWidget(
					paddings,
					actionsHeight,
					conditionalSize(s.similarImagesShown, similarImagesHeight),
					conditionalSize(s.progressBackground.open, progressHeight),
				),
				giu.Condition(s.similarImagesShown, giu.Layout{
					s.similarImagesWidget(similarImagesHeight),
				},
					nil),
				giu.Condition(s.progressBackground.open, giu.Layout{
					giu.Row(
						giu.Label("Caching images..."),
						giu.ProgressBar(float32(s.progressBackground.position)/float32(s.progressBackground.max)).
							Overlay(fmt.Sprintf("%d/%d", s.progressBackground.position, s.progressBackground.max)).
							Size(giu.Auto, progressHeight),
					),
				}, nil),
				s.actionsWidget(actionsHeight),
				giu.PrepareMsgbox(),
			)

			// Ignore all input when the progress bar is shown
			// This prevents any unexpected changes
			if !s.progressModal.open && !s.deviceModal.open && !s.applyChangesModal.open {
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

func (s *Ui) similarImagesWidget(height float32) *giu.RowWidget {
	return giu.Row(
		giu.Button("Hide").
			Size(50, height).
			OnClick(func() {
				s.similarImagesShown = false
			}),
		s.similarImagesList.SetImages(s.similarImages).Size(giu.Auto, height),
	)
}

func (s *Ui) actionsWidget(bottomHeight float32) *giu.RowWidget {
	return giu.Row(
		giu.Dummy(0, bottomHeight),
		giu.Button("Edit categories").OnClick(s.openEditCategoriesView),
		giu.Button("Search similar").OnClick(s.searchSimilar),
		giu.Button("Cast").OnClick(s.openCastToDeviceView),
		giu.Button("Open directory").OnClick(s.changeDirectory),
		giu.Custom(func() {
			giu.Row(
				giu.Combo("", s.getZoomPercent(), zoomLevelItems, &s.zoomLevel).Size(100),
				giu.Button("-").OnClick(s.zoomOut),
				giu.Button("+").OnClick(s.zoomIn),
				giu.Button("Fit").OnClick(s.resetZoom),
			).Build()
		}),
		giu.Dummy(-120, 0),
		giu.Button("Apply Categories").OnClick(s.applyCategories),
	)
}

func conditionalSize(condition bool, size float32) float32 {
	if condition {
		return size
	} else {
		return 0
	}
}

func (s *Ui) mainImageWidget(bottomHeight ...float32) *giu.CustomWidget {
	return giu.Custom(func() {
		width, height := giu.GetAvailableRegion()
		h := height
		for _, hh := range bottomHeight {
			h -= hh
		}

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

		s.currentImageWidget = widget.ResizableImage(s.currentImageTexture)
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
						Flags(giu.WindowFlagsHorizontalScrollbar).
						Layout(s.currentImageWidget.
							ZoomFactor(s.getZoomFactor()).
							ImageSize(s.currentImageTexture.Width, s.currentImageTexture.Height),
						),
					nButton,
				),
			).Build()
	})
}

func (s *Ui) imagesWidget() *giu.CustomWidget {
	return giu.Custom(func() {
		width, _ := giu.GetAvailableRegion()
		height := float32(60)
		buttonWidth := float32(30)
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
		firstImageButton := giu.Button("<<").
			OnClick(func() {
				s.sender.SendCommandToTopic(api.ImageRequestAtIndex, &api.ImageAtQuery{
					Index: 0,
				})
			}).
			Size(buttonWidth, height)
		lastImageButton := giu.Button(">>").
			OnClick(func() {
				s.sender.SendCommandToTopic(api.ImageRequestAtIndex, &api.ImageAtQuery{
					Index: -1,
				})
			}).
			Size(buttonWidth, height)
		giu.Row(
			firstImageButton,
			s.previousImagesList.Size(listWidth, height).SetImages(s.previousImages),
			giu.Row(
				widget.ResizableImage(s.currentImageTexture).
					Size(centerPieceWidth, height),
			),
			s.nextImagesList.Size(listWidth, height).SetImages(s.nextImages),
			lastImageButton,
		).Build()
		giu.PopStyle()
	})
}

func getProgressModal(id string, sender api.Sender, modal *progressModal) giu.Widget {
	return giu.PopupModal(id).
		Flags(giu.WindowFlagsAlwaysAutoResize|giu.WindowFlagsNoTitleBar).
		Layout(
			giu.Label(modal.label),
			giu.Row(
				giu.ProgressBar(float32(modal.position)/float32(modal.max)).
					Overlay(fmt.Sprintf("%d/%d", modal.position, modal.max)),
				giu.Condition(modal.canCancel,
					giu.Layout{
						giu.Button("Cancel").
							OnClick(func() {
								sender.SendToTopic(api.SimilarRequestStop)
							}),
					},
					nil),
			),
			giu.Custom(func() {
				if !modal.open {
					giu.CloseCurrentPopup()
				}
			}))
}

func getDeviceModal(id string, sender api.Sender, modal *deviceModal) giu.Widget {
	return giu.PopupModal(id).
		Flags(giu.WindowFlagsAlwaysAutoResize|giu.WindowFlagsNoDecoration).
		Layout(
			giu.Label(modal.label),
			giu.Custom(func() {
				if modal.inProgress {
					giu.ProgressIndicator("Searching...", 10, 10, 5).Build()
				}
			}),
			giu.ListBox("Found devices", modal.devices).
				Size(300, 200).
				OnChange(func(selectedIndex int) {
					modal.selectedIndex = selectedIndex
				}),
			giu.Checkbox("Show blurred background", &modal.showBackground),
			giu.Row(
				giu.Button("OK##SelectDevice").
					Disabled(len(modal.devices) == 0).
					OnClick(func() {
						sender.SendCommandToTopic(api.CastDeviceSelect, &api.SelectDeviceCommand{
							Name:           modal.devices[modal.selectedIndex],
							ShowBackground: modal.showBackground,
						})
						modal.open = false
						modal.inProgress = false
					}),
				giu.Button("Cancel##CancelDevice").
					OnClick(func() {
						sender.SendToTopic(api.SimilarRequestStop)
						modal.open = false
						modal.inProgress = false
					}),
			),
			giu.Custom(func() {
				if !modal.open {
					giu.CloseCurrentPopup()
				}
			}))
}

func getApplyChangesModal(id string, sender api.Sender, modal *applyChangesModal) giu.Widget {
	return giu.PopupModal(id).
		Flags(giu.WindowFlagsAlwaysAutoResize|giu.WindowFlagsNoDecoration).
		Layout(
			giu.Label(modal.label),
			giu.Checkbox("Keep original images", &modal.keepOriginals),
			giu.Checkbox("Fix orientation", &modal.fixOrientation),
			giu.SliderInt("Quality", &modal.quality, 0, 100),
			giu.Row(
				giu.Button("Apply##ApplyChanges").
					OnClick(func() {
						sender.SendCommandToTopic(api.CategoryPersistAll, &api.PersistCategorizationCommand{
							KeepOriginals:  modal.keepOriginals,
							FixOrientation: modal.fixOrientation,
							Quality:        int(modal.quality),
						})
						modal.open = false
					}),
				giu.Button("Cancel##ApplyChanges").
					OnClick(func() {
						sender.SendToTopic(api.SimilarRequestStop)
						modal.open = false
					}),
			),
			giu.Custom(func() {
				if !modal.open {
					giu.CloseCurrentPopup()
				}
			}))
}

func (s *Ui) handleKeyPress() bool {
	const bigJumpSize = 10
	const hugeJumpSize = 100

	shiftDown := giu.IsKeyDown(giu.KeyLeftShift) || giu.IsKeyDown(giu.KeyRightShift)
	altDown := giu.IsKeyDown(giu.KeyLeftAlt) || giu.IsKeyDown(giu.KeyRightAlt)
	controlDown := giu.IsKeyDown(giu.KeyLeftControl) || giu.IsKeyDown(giu.KeyRightControl)

	if giu.IsKeyPressed(giu.KeyF8) {
		s.openCastToDeviceView()
	}
	if giu.IsKeyPressed(giu.KeyF10) {
		s.sender.SendToTopic(api.ImageShowAll)
	}
	if giu.IsKeyPressed(giu.KeyF12) {
		s.searchSimilar()
	}
	if giu.IsKeyPressed(giu.KeyEnter) && controlDown {
		s.applyCategories()
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
		if controlDown {
			s.sender.SendCommandToTopic(api.ImageRequestPreviousOffset, &api.ImageAtQuery{Index: bigJumpSize})
		} else {
			s.sender.SendToTopic(api.ImageRequestPrevious)
		}
	}
	if giu.IsKeyPressed(giu.KeyRight) {
		if controlDown {
			s.sender.SendCommandToTopic(api.ImageRequestNextOffset, &api.ImageAtQuery{Index: bigJumpSize})
		} else {
			s.sender.SendToTopic(api.ImageRequestNext)
		}
	}

	// Zoom shortcuts are based on US layout at least for now
	if giu.IsKeyPressed(giu.KeyKPAdd) || giu.IsKeyPressed(giu.KeyEqual) {
		s.zoomIn()
	}
	if giu.IsKeyPressed(giu.KeyKPSubtract) || giu.IsKeyPressed(giu.KeyMinus) {
		s.zoomOut()
	}

	s.categoryKeyManager.HandleKeys(&guiapi.CategoryAction{
		StayOnImage:      shiftDown,
		ForceCategory:    controlDown,
		ShowOnlyCategory: altDown,
	})
	return true
}

func (s *Ui) UpdateCategories(categories *api.UpdateCategoriesCommand) {
	s.categories = make([]*apitype.Category, len(categories.Categories))
	copy(s.categories, categories.Categories)

	s.categoryEditWidget.SetCategories(s.categories)
	s.categoryKeyManager.Reset(categories.Categories)
	s.sender.SendToTopic(api.ImageRequestCurrent)
}

func (s *Ui) SetImages(command *api.SetImagesCommand) {
	if command.Topic == api.ImageRequestNext {
		s.nextImages = []*widget.TexturedImage{}
		for _, data := range command.Images {
			ti := widget.NewTexturedImage(
				data,
				s.imageCache,
			)
			s.nextImages = append(s.nextImages, ti)
		}
	} else if command.Topic == api.ImageRequestPrevious {
		s.previousImages = []*widget.TexturedImage{}
		for _, data := range command.Images {
			ti := widget.NewTexturedImage(
				data,
				s.imageCache,
			)
			s.previousImages = append(s.previousImages, ti)
		}
	} else if command.Topic == api.ImageRequestSimilar {
		s.similarImages = []*widget.TexturedImage{}
		for _, data := range command.Images {
			ti := widget.NewTexturedImage(
				data,
				s.imageCache,
			)
			s.similarImages = append(s.similarImages, ti)
		}
	}
	giu.Update()
}

func (s *Ui) SetCurrentImage(command *api.UpdateImageCommand) {
	s.currentImageTexture.ChangeImage(command.Image)
	width, height := s.win.GetSize()
	s.currentCategoryId = command.CategoryId
	s.currentImageTexture.LoadImageAsTexture(float32(width), float32(height), s.getZoomFactor())
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
}

func (s *Ui) UpdateProgress(command *api.UpdateProgressCommand) {
	var progress *progressModal
	if command.Modal {
		progress = &s.progressModal
	} else {
		progress = &s.progressBackground
	}

	if command.Current == command.Total {
		logger.Trace.Printf("Progress '%s' completed", command.Name)
		progress.open = false
		progress.label = ""
		progress.position = 1
		progress.max = 1
		progress.canCancel = false
	} else {
		logger.Trace.Printf("Update progress '%s' %d/%d", command.Name, command.Current, command.Total)
		progress.open = true
		progress.label = command.Name
		progress.position = command.Current
		progress.max = command.Total
		progress.canCancel = command.CanCancel
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

func (s *Ui) FindDevices() {
	s.deviceModal.devices = []string{}
	s.deviceModal.inProgress = true
	s.deviceModal.open = true
	s.sender.SendToTopic(api.CastDeviceSearch)
}

func (s *Ui) ShowError(command *api.ErrorCommand) {
	logger.Error.Printf("Error: %s", command.Message)
	giu.Msgbox("Error", command.Message)
}

func (s *Ui) zoomIn() {
	// Zoom in by finding the next zoom level
	// Using this algorithm so that the zoom in feature works
	// also with the "fit to page" zoom level
	currentZoom := s.getZoomFactor()
	if currentZoom < 0 && s.currentImageWidget != nil {
		currentZoom = s.currentImageWidget.CurrentActualZoom()
	}
	for i, level := range zoomLevels {
		if level > currentZoom {
			s.zoomLevel = int32(i)
			break
		}
	}
}

func (s *Ui) zoomOut() {
	// Zoom out by finding the next smaller zoom level
	// Using this algorithm so that the zoom out feature works
	// also with the "fit to page" zoom level
	currentZoom := s.getZoomFactor()
	if currentZoom < 0 && s.currentImageWidget != nil {
		currentZoom = s.currentImageWidget.CurrentActualZoom()
	}
	for i, level := range zoomLevels {
		if level < currentZoom {
			s.zoomLevel = int32(i)
			// No break here because we want to find the "maximum" smallest
			// before the value goes above the target zoom level
		}
	}
}

func (s *Ui) resetZoom() {
	s.zoomLevel = defaultZoomLevel
}

func (s *Ui) getZoomFactor() float32 {
	return zoomLevels[s.zoomLevel]
}

func (s *Ui) getZoomPercent() string {
	zoomFactor := s.getZoomFactor()
	if zoomFactor < 0.0 {
		fitPercentage := float32(1)
		if s.currentImageWidget != nil {
			fitPercentage = s.currentImageWidget.CurrentActualZoom()
		}
		return fmt.Sprintf("Fit (%d %%)", int(fitPercentage*100))
	} else {
		return fmt.Sprintf("%d %%", int(zoomFactor*100))
	}
}

func (s *Ui) searchSimilar() {
	s.similarImagesShown = true
	s.sender.SendToTopic(api.SimilarRequestSearch)
}

func (s *Ui) openEditCategoriesView() {
	s.categoryEditWidget.SetCategories(s.categories)
	s.showCategoryEditModal = true
}

func (s *Ui) openCastToDeviceView() {
	s.deviceModal.open = true
	s.FindDevices()
}

func (s *Ui) applyCategories() {
	s.applyChangesModal.open = true
}

func (s *Ui) changeDirectory() {
	s.Init("")
}
