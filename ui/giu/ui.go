package gtk

import (
	"fmt"
	"github.com/AllenDang/giu"
	"github.com/OpenDiablo2/dialog"
	"image/color"
	"sort"
	"time"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/common/logger"
	"vincit.fi/image-sorter/ui/giu/internal"
	"vincit.fi/image-sorter/ui/giu/internal/guiapi"
	"vincit.fi/image-sorter/ui/giu/internal/widget"
)

type Ui struct {
	// General
	win                    *giu.MasterWindow
	oldWidth, oldHeight    int
	imageCache             api.ImageStore
	sender                 api.Sender
	categories             []*apitype.Category
	rootPath               string
	imageManager           *internal.ImageManager
	currentImageWidget     *widget.ResizableImageWidget
	currentThumbnailWidget *widget.ResizableImageWidget
	nextImages             []*guiapi.TexturedImage
	previousImages         []*guiapi.TexturedImage
	similarImages          []*guiapi.TexturedImage
	categoryKeyManager     *internal.CategoryKeyManager
	currentImageCategories map[apitype.CategoryId]bool
	currentCategoryId      apitype.CategoryId
	progressModal          progressModal
	progressBackground     progressModal
	deviceModal            deviceModal
	applyChangesModal      applyChangesModal
	showCategoryEditModal  bool
	categoryEditWidget     *widget.CategoryEditWidget
	showMetaData           bool

	nextImagesList       *widget.HorizontalImageListWidget
	previousImagesList   *widget.HorizontalImageListWidget
	similarImagesList    *widget.HorizontalImageListWidget
	similarImagesShown   bool
	widthInNumOfImage    int
	zoomStatus           *internal.ZoomStatus
	totalImageCount      int
	currentImagePos      int
	currentImageMetaData []string

	everythingLoaded bool
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
	zoomStep            = 0.1
)

func NewUi(params *common.Params, broker api.Sender, imageCache api.ImageStore) api.Gui {
	gui := Ui{
		win:          giu.NewMasterWindow("Image Sorter", defaultWindowWidth, defaultWindowHeight, 0),
		imageCache:   imageCache,
		sender:       broker,
		rootPath:     params.RootPath(),
		imageManager: internal.NewImageManager(imageCache),
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
		similarImagesShown: false,
		widthInNumOfImage:  0,
		zoomStatus:         internal.NewZoomStatus(),
	}

	onImageSelected := func(imageFile *apitype.ImageFile) {
		gui.jumpToImageId(imageFile.Id())
	}

	gui.nextImagesList = widget.HorizontalImageList(onImageSelected, false, false, true)
	gui.previousImagesList = widget.HorizontalImageList(onImageSelected, false, true, true)
	gui.similarImagesList = widget.HorizontalImageList(onImageSelected, false, false, false)

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

	gui.categoryKeyManager = &internal.CategoryKeyManager{
		Callback: func(def *internal.CategoryDef, action *guiapi.CategoryAction) {
			operation := apitype.CATEGORIZE
			if !action.ForceCategory {
				if _, ok := gui.currentImageCategories[def.CategoryId]; ok {
					operation = apitype.UNCATEGORIZE
				}
			}

			if action.ShowOnlyCategory {
				if gui.currentCategoryId == apitype.NoCategory {
					broker.SendCommandToTopic(api.CategoriesShowOnly, &api.SelectCategoryCommand{
						CategoryId: def.CategoryId,
					})
				} else {
					broker.SendToTopic(api.ImageShowAll)
				}
			} else {
				broker.SendCommandToTopic(api.CategorizeImage, &api.CategorizeCommand{
					ImageId:         gui.imageManager.LoadedImage().Id(),
					CategoryId:      def.CategoryId,
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
			if err == dialog.ErrCancelled {
				logger.Debug.Printf("User cancelled directory selection")
				return
			}
			logger.Error.Fatal("Error while trying to load directory", err)
		}
	}
	logger.Debug.Printf("Opening directory '%s'", directory)

	s.sender.SendCommandToTopic(api.DirectoryChanged, &api.DirectoryChangedCommand{Directory: directory})
}

func (s *Ui) Run() {
	currentScale := giu.Context.GetPlatform().GetContentScale()
	logger.Info.Print("Scale", currentScale)

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
			s.oldWidth = newWidth
			s.oldHeight = newHeight
			s.imageManager.SetSize(float32(newWidth), float32(newHeight), s.zoomStatus)
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
			actionsHeight := float32(35.0)
			similarImagesHeight := float32(60.0)
			paddings := float32(8.0)
			imageName := ""
			imageInfo := ""
			var highlightedImage *apitype.ImageFile
			if s.nextImagesList.HighlightedImage() != nil {
				highlightedImage = s.nextImagesList.HighlightedImage()
			} else if s.previousImagesList.HighlightedImage() != nil {
				highlightedImage = s.previousImagesList.HighlightedImage()
			} else if s.imageManager.LoadedImage() != nil {
				highlightedImage = s.imageManager.LoadedImage()
			}

			var progressPercent = 0
			if s.totalImageCount > 0 {
				progressPercent = int(float32(s.currentImagePos) / float32(s.totalImageCount) * 100.0)
			}
			progress := fmt.Sprintf("%d/%d (%d %%): ", s.currentImagePos, s.totalImageCount, progressPercent)
			if highlightedImage != nil {
				imageName = highlightedImage.FileName()
				imageInfo = fmt.Sprintf("(%d x %d)",
					highlightedImage.Width(),
					highlightedImage.Height(),
				)
			}
			var categoriesView giu.Widget
			if len(categories) > 0 {
				categoriesView = giu.Row(widget.CategoryButtonView(categories))
			} else {
				categoriesView = giu.Row(giu.Label("No categories defined. Please edit categories."), giu.Button("Edit categories").OnClick(s.openEditCategoriesView))
			}

			showMetaDataButtonLable := ""
			if s.showMetaData {
				showMetaDataButtonLable = "Hide metadata"
			} else {
				showMetaDataButtonLable = "Show metadata"
			}

			mainWindow.Layout(
				s.imagesWidget(),
				giu.Row(
					giu.Button(showMetaDataButtonLable).OnClick(s.toggleShowMetaData),
					giu.Label(progress),
					giu.Label(imageName),
					giu.Condition(imageInfo != "", giu.Layout{giu.Label(imageInfo)}, giu.Layout{giu.Label("")}),
				),
				categoriesView,
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
					s.showMetaData,
					s.currentImageMetaData,
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

const buttonPaddingHorizontal = 20
const narrowButtonPaddingHorizontal = 10
const buttonPaddingVertical = 5

func (s *Ui) actionsWidget(bottomHeight float32) giu.Widget {
	zoomLevel := s.zoomStatus.SelectedZoom()
	return giu.Column(
		giu.Dummy(0, 5),
		giu.Row(
			giu.Dummy(0, bottomHeight),
			giu.Style().SetStyle(giu.StyleVarFramePadding, buttonPaddingHorizontal, buttonPaddingVertical).To(
				giu.Row(
					giu.Button("Edit categories").OnClick(s.openEditCategoriesView),
					giu.Button("Search similar").OnClick(s.searchSimilar),
					giu.Button("Cast").OnClick(s.openCastToDeviceView),
					giu.Button("Open directory").OnClick(s.changeDirectory),
				),
			),
			giu.Custom(func() {
				giu.Style().SetStyle(giu.StyleVarFramePadding, narrowButtonPaddingHorizontal, buttonPaddingVertical).To(
					giu.Row(
						giu.Combo("", s.getZoomPercent(), internal.ZoomLabels(), &zoomLevel).
							Size(100).
							OnChange(func() {
								if zoomLevel == 0 {
									s.zoomStatus.SetZoomFit()
								} else {
									s.zoomStatus.SetZoomLevel(zoomLevel)
								}
							}),
						giu.Button("-").OnClick(s.zoomOut),
						giu.Button("+").OnClick(s.zoomIn),
						giu.Button("Fit").OnClick(s.resetZoom),
					)).Build()
			}),
			giu.Dummy(-160, 0),
			giu.Style().SetStyle(giu.StyleVarFramePadding, buttonPaddingHorizontal, buttonPaddingVertical).To(
				giu.Button("Apply Categories").OnClick(s.applyCategories),
			),
		))
}

func conditionalSize(condition bool, size float32) float32 {
	if condition {
		return size
	} else {
		return 0
	}
}

const metaDataWidth = float32(300)

func (s *Ui) mainImageWidget(showMetaData bool, metaData []string, widgetHeights ...float32) *giu.CustomWidget {
	return giu.Custom(func() {
		availableWidth, availableHeight := giu.GetAvailableRegion()
		height := availableHeight
		for _, widgetHeight := range widgetHeights {
			height -= widgetHeight
		}

		availableWidth = availableWidth - 30.0*2
		if showMetaData {
			availableWidth -= metaDataWidth
		}

		previousButton := giu.Button("<").
			OnClick(func() {
				s.jumpToOffset(-1)
			}).
			Size(30, height)
		nextButton := giu.Button(">").
			OnClick(func() {
				s.jumpToOffset(1)
			}).
			Size(30, height)

		loadedImageTexture := s.imageManager.LoadedImageTexture()
		if s.currentImageWidget == nil {
			s.currentImageWidget = widget.ResizableImage(loadedImageTexture)
			s.currentImageWidget.SetZoomHandlers(s.zoomIn, s.zoomOut)
		} else {
			s.currentImageWidget.UpdateImage(loadedImageTexture)
		}

		giu.Style().
			SetStyle(giu.StyleVarItemSpacing, 0, 0).
			SetColor(giu.StyleColorBorder, color.RGBA{0, 0, 0, 255}).
			SetColor(giu.StyleColorChildBg, color.RGBA{0, 0, 0, 255}).
			To(
				giu.Row(
					giu.Condition(showMetaData,
						giu.Layout{
							giu.Child().
								Size(metaDataWidth, height).
								Layout(giu.ListBox("imageMetaData", metaData))},
						giu.Layout{giu.Dummy(0, 0)},
					),
					previousButton,
					giu.Child().
						Size(availableWidth, height).
						Border(true).
						Flags(giu.WindowFlagsHorizontalScrollbar).
						Layout(s.currentImageWidget.
							ZoomFactor(s.getZoomFactor()).
							ImageSize(loadedImageTexture.Width, loadedImageTexture.Height),
						),
					nextButton,
				),
			).Build()
	})
}

func (s *Ui) imagesWidget() *giu.CustomWidget {
	return giu.Custom(func() {
		availableWidth, _ := giu.GetAvailableRegion()
		height := float32(60)
		buttonWidth := float32(30)
		centerPieceWidth := float32(120)
		listWidth := (availableWidth - buttonWidth*2 - centerPieceWidth) / 2

		widthInNumOfImage := int(listWidth/60) + 1

		if widthInNumOfImage != s.widthInNumOfImage {
			s.sender.SendCommandToTopic(api.ImageListSizeChanged, &api.ImageListCommand{
				ImageListSize: widthInNumOfImage,
			})
		}
		s.widthInNumOfImage = widthInNumOfImage

		if s.currentThumbnailWidget == nil {
			s.currentThumbnailWidget = widget.ResizableImage(s.imageManager.LoadedImageTexture())
		}
		s.currentThumbnailWidget.UpdateImage(s.imageManager.LoadedImageTexture())
		s.currentThumbnailWidget.Size(centerPieceWidth, height)

		giu.PushItemSpacing(0, 0)
		firstImageButton := giu.Button("<<").
			OnClick(func() {
				s.jumpToIndex(0)
			}).
			Size(buttonWidth, height)
		lastImageButton := giu.Button(">>").
			OnClick(func() {
				s.jumpToIndex(-1)
			}).
			Size(buttonWidth, height)
		giu.Row(
			firstImageButton,
			s.previousImagesList.Size(listWidth, height).SetImages(s.previousImages),
			giu.Row(s.currentThumbnailWidget),
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
			giu.SliderInt(&modal.quality, 0, 100).Label("Quality"),
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
	const mediumJumpSize = 10
	const bigJumpSize = 50
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
	if giu.IsKeyPressed(giu.KeyF11) {
		s.toggleShowMetaData()
	}
	if giu.IsKeyPressed(giu.KeyF12) {
		s.searchSimilar()
	}
	if giu.IsKeyPressed(giu.KeyEnter) && controlDown {
		s.applyCategories()
	}

	// Navigation

	if giu.IsKeyPressed(giu.KeyPageDown) {
		s.jumpToOffset(-hugeJumpSize)
	}
	if giu.IsKeyPressed(giu.KeyPageUp) {
		s.jumpToOffset(hugeJumpSize)
	}

	if giu.IsKeyPressed(giu.KeyHome) {
		s.jumpToIndex(0)
	}
	if giu.IsKeyPressed(giu.KeyEnd) {
		s.jumpToIndex(-1)
	}

	if giu.IsKeyPressed(giu.KeyLeft) {
		if shiftDown {
			s.jumpToOffset(-mediumJumpSize)
		} else if controlDown {
			s.jumpToOffset(-bigJumpSize)
		} else {
			s.jumpToOffset(-1)
		}
	}
	if giu.IsKeyPressed(giu.KeyRight) {
		if shiftDown {
			s.jumpToOffset(mediumJumpSize)
		} else if controlDown {
			s.jumpToOffset(bigJumpSize)
		} else {
			s.jumpToOffset(1)
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

func (s *Ui) jumpToOffset(jumpSize int) {
	if jumpSize == 0 {
		return
	}

	//s.imageManager.SetCurrentImage(nil)

	jumpMagnitude := jumpSize
	if jumpMagnitude < 0 {
		jumpMagnitude = 0 - jumpMagnitude
	}

	if jumpSize < 0 {
		if jumpSize == -1 {
			s.sender.SendToTopic(api.ImageRequestPrevious)
		} else {
			s.sender.SendCommandToTopic(api.ImageRequestPreviousOffset, &api.ImageAtQuery{Index: jumpMagnitude})
		}
	} else if jumpSize > 0 {
		if jumpSize == 1 {
			s.sender.SendToTopic(api.ImageRequestNext)
		} else {
			s.sender.SendCommandToTopic(api.ImageRequestNextOffset, &api.ImageAtQuery{Index: jumpMagnitude})
		}
	}
}

func (s *Ui) jumpToIndex(index int) {
	//s.imageManager.SetCurrentImage(nil)
	s.sender.SendCommandToTopic(api.ImageRequestAtIndex, &api.ImageAtQuery{
		Index: index,
	})
}

func (s *Ui) jumpToImageId(imageId apitype.ImageId) {
	//s.imageManager.SetCurrentImage(nil)
	s.sender.SendCommandToTopic(api.ImageRequest, &api.ImageQuery{
		Id: imageId,
	})
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
		s.nextImages = []*guiapi.TexturedImage{}
		for _, data := range command.Images {
			ti := s.imageManager.GetThumbnailTexture(data)
			s.nextImages = append(s.nextImages, ti)
		}
	} else if command.Topic == api.ImageRequestPrevious {
		s.previousImages = []*guiapi.TexturedImage{}
		for _, data := range command.Images {
			ti := s.imageManager.GetThumbnailTexture(data)
			s.previousImages = append(s.previousImages, ti)
		}
	} else if command.Topic == api.ImageRequestSimilar {
		s.similarImages = []*guiapi.TexturedImage{}
		for _, data := range command.Images {
			ti := s.imageManager.GetThumbnailTexture(data)
			s.similarImages = append(s.similarImages, ti)
		}
	}
	giu.Update()
}

func (s *Ui) SetCurrentImage(command *api.UpdateImageCommand) {
	width, height := s.win.GetSize()
	s.imageManager.SetCurrentImage(command.Image, float32(width), float32(height), s.zoomStatus)
	s.currentCategoryId = command.CategoryId
	s.sendCurrentImageChangedEvent()

	s.imageCache.Purge()
	s.totalImageCount = command.Total
	s.currentImagePos = command.Index + 1
	s.currentImageMetaData = []string{}
	for k, v := range command.MetaData.MetaData() {
		s.currentImageMetaData = append(s.currentImageMetaData, fmt.Sprintf("%s: %s", k, v))
	}
	sort.Strings(s.currentImageMetaData)

	giu.Update()
}

func (s *Ui) sendCurrentImageChangedEvent() {
	s.sender.SendCommandToTopic(api.ImageChanged, &api.ImageCategoryQuery{
		ImageId: s.imageManager.ActiveImageId(),
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
	s.zoomStatus.ZoomIn(s.currentImageWidget.CurrentActualZoom(), zoomStep)
}

func (s *Ui) zoomOut() {
	s.zoomStatus.ZoomOut(s.currentImageWidget.CurrentActualZoom(), zoomStep)
}

func (s *Ui) resetZoom() {
	s.zoomStatus.ResetZoom()
}

func (s *Ui) getZoomFactor() (float32, guiapi.ZoomMode) {
	if s.zoomStatus.ZoomMode() == guiapi.ZoomCustom || s.zoomStatus.ZoomMode() == guiapi.ZoomFixed {
		return s.zoomStatus.ZoomLevel(), s.zoomStatus.ZoomMode()
	} else {
		if s.currentImageWidget != nil {
			return s.currentImageWidget.CurrentActualZoom(), s.zoomStatus.ZoomMode()
		} else {
			return 1, s.zoomStatus.ZoomMode()
		}
	}
}

func (s *Ui) getZoomPercent() string {
	zoomFactor, zoomMode := s.getZoomFactor()
	zoomFactorFmt := internal.FormatZoomFactor(zoomFactor)
	switch zoomMode {
	case guiapi.ZoomCustom, guiapi.ZoomFixed:
		return zoomFactorFmt
	case guiapi.ZoomFit:
		return fmt.Sprintf("Fit (%s)", zoomFactorFmt)
	default:
		return ""
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

func (s *Ui) toggleShowMetaData() {
	s.showMetaData = !s.showMetaData
}
