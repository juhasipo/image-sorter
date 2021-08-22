package gtk

import (
	"errors"
	"fmt"
	"github.com/AllenDang/giu"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"golang.org/x/image/colornames"
	"image"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/common/logger"
	"vincit.fi/image-sorter/ui/gtk/component"
)

type Ui struct {
	// General
	win                 *giu.MasterWindow
	fullscreen          bool
	imageCache          api.ImageStore
	sender              api.Sender
	categories          []*apitype.Category
	rootPath            string
	currentImageTexture *texturedImage
	nextImages          []*texturedImage
	previousImages      []*texturedImage

	api.Gui
	component.CallbackApi
	currentImageCategories map[apitype.CategoryId]bool
}

type texturedImage struct {
	width   float32
	height  float32
	texture *giu.Texture
	imageId apitype.ImageId
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
		currentImageTexture: &texturedImage{},
	}

	gui.Init(params.RootPath())
	return &gui
}

func (s *Ui) Init(directory string) {
}

type ResizableImageWidget struct {
	imageWidth  float32
	imageHeight float32
	imageRatio  float32
	giu.ImageWidget
}

func ResizableImage(texture *giu.Texture, width float32, height float32) *ResizableImageWidget {
	return &ResizableImageWidget{
		imageWidth:  width,
		imageHeight: height,
		imageRatio:  width / height,
		ImageWidget: *giu.Image(texture),
	}
}

type CategoryButtonWidget struct {
	categoryId apitype.CategoryId
	name       string
	active     bool
	onClick    func(*api.CategorizeCommand)
	giu.ColumnWidget
}

func CategoryButton(categoryId apitype.CategoryId, name string, active bool, onClick func(command *api.CategorizeCommand)) *CategoryButtonWidget {
	return &CategoryButtonWidget{
		categoryId: categoryId,
		name:       name,
		active:     active,
		onClick:    onClick,
	}
}

const categoryPrimaryButtonWidth = 100
const categoryArrowButtonWidth = 20
const categoryPrimaryButtonHeight = 20
const categoryIndicatorButtonHeight = 5

func (s *CategoryButtonWidget) Build() {
	statusColor := colornames.Green
	if !s.active {
		statusColor = colornames.Navy
	}

	operation := apitype.MOVE
	if s.active {
		operation = apitype.NONE
	}

	primaryAction := func(operation apitype.Operation, stayOnImage bool, forceCategory bool) {
		s.onClick(&api.CategorizeCommand{
			CategoryId:      s.categoryId,
			Operation:       operation,
			StayOnSameImage: stayOnImage,
			ForceToCategory: forceCategory,
		})
	}

	actionName := "Add"
	if s.active {
		actionName = "Remove"
	}

	menuName := fmt.Sprintf("CategoryMenu-%d", s.categoryId)
	menu := giu.Popup(menuName).
		Layout(
			giu.Button(actionName+" category").OnClick(func() {
				primaryAction(operation, false, false)
				giu.CloseCurrentPopup()
			}).Size(180, 30),
			giu.Button(actionName+" category and stay").OnClick(func() {
				primaryAction(operation, true, false)
				giu.CloseCurrentPopup()
			}).Size(180, 30),
			giu.Button("Force category").OnClick(func() {
				primaryAction(apitype.MOVE, false, true)
				giu.CloseCurrentPopup()
			}).Size(180, 30),
		)

	primaryButton := giu.Button(s.name).
		Size(categoryPrimaryButtonWidth, categoryPrimaryButtonHeight).
		OnClick(func() {
			stayOnImage := giu.IsKeyDown(giu.KeyLeftShift) || giu.IsKeyDown(giu.KeyRightShift)
			forceCategory := giu.IsKeyDown(giu.KeyLeftControl) || giu.IsKeyDown(giu.KeyRightControl)
			primaryAction(operation, stayOnImage, forceCategory)
		})
	menuButton := giu.ArrowButton("Menu", giu.DirectionDown).OnClick(func() {
		giu.OpenPopup(menuName)
	})
	statusIndicator := giu.Custom(func() {
		// w, _ := giu.GetAvailableRegion()
		canvas := giu.GetCanvas()
		start := giu.GetCursorPos()
		end := start.Add(image.Pt(categoryPrimaryButtonWidth+categoryArrowButtonWidth, categoryIndicatorButtonHeight))

		canvas.AddRectFilled(start, end, statusColor, 0, 0)
	})

	giu.Style().
		SetStyle(giu.StyleVarItemSpacing, 0, 0).
		SetStyle(giu.StyleVarItemInnerSpacing, 0, 0).
		To(giu.Column(
			giu.Row(
				primaryButton,
				menuButton,
			),
			menu,
			statusIndicator,
		)).Build()
}

func (s *ResizableImageWidget) Build() {
	paddingW, _ := giu.GetFramePadding()
	maxW, maxH := giu.GetAvailableRegion()
	maxW = maxW - 120 - paddingW*2.0
	newW := maxW
	newH := newW / s.imageRatio

	if newH > maxH {
		newW = maxH * s.imageRatio
		newH = maxH
	}

	offsetW := (maxW - newW) / 2.0
	offsetH := (maxH - newH) / 2.0

	s.ImageWidget.Size(newW, newH)
	giu.Row(giu.Dummy(offsetW, offsetH), &s.ImageWidget).Build()
}

func (s *Ui) Run() {
	s.sender.SendCommandToTopic(api.DirectoryChanged, s.rootPath)
	s.win.Run(func() {
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
			categorizeButton := CategoryButton(categoryId, text, active, func(command *api.CategorizeCommand) {
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
					giu.Column(ResizableImage(s.currentImageTexture.texture, s.currentImageTexture.width, s.currentImageTexture.height)),
					giu.Dummy(-120, giu.Auto),
					giu.Column(previousImages...),
				),
				giu.PrepareMsgbox(),
			)
	})
}

func (s *Ui) handleKeyPress(_ *gtk.ApplicationWindow, e *gdk.Event) bool {
	const bigJumpSize = 10
	const hugeJumpSize = 100

	keyEvent := gdk.EventKeyNewFromEvent(e)
	key := keyEvent.KeyVal()

	_, controlDown, altDown := resolveModifierStatuses(keyEvent)

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
		s.sender.SendCommandToTopic(api.ImageRequestPreviousOffset, &api.ImageAtQuery{Index: hugeJumpSize})
	} else if key == gdk.KEY_Page_Down {
		s.sender.SendCommandToTopic(api.ImageRequestNextOffset, &api.ImageAtQuery{Index: hugeJumpSize})
	} else if key == gdk.KEY_Home {
		s.sender.SendCommandToTopic(api.ImageRequestAtIndex, &api.ImageAtQuery{Index: 0})
	} else if key == gdk.KEY_End {
		s.sender.SendCommandToTopic(api.ImageRequestAtIndex, &api.ImageAtQuery{Index: -1})
	} else if key == gdk.KEY_Left {
		if controlDown {
			s.sender.SendCommandToTopic(api.ImageRequestPreviousOffset, &api.ImageAtQuery{Index: bigJumpSize})
		} else {
			s.sender.SendToTopic(api.ImageRequestPrevious)
		}
	} else if key == gdk.KEY_Right {
		if controlDown {
			s.sender.SendCommandToTopic(api.ImageRequestNextOffset, &api.ImageAtQuery{Index: bigJumpSize})
		} else {
			s.sender.SendToTopic(api.ImageRequestNext)
		}
	} /*else if command := s.topActionView.NewCommandForShortcut(key, s.imageView.CurrentImageFile()); command != nil {
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

	//s.topActionView.UpdateCategories(categories)
	s.sender.SendToTopic(api.ImageRequestCurrent)
}

func (s *Ui) UpdateCurrentImage() {
	//s.imageView.UpdateCurrentImage()
}

func i2t(i image.Image) (*giu.Texture, error) {
	return giu.NewTextureFromRgba(i.(*image.RGBA))
}

func (s *Ui) SetImages(command *api.SetImagesCommand) {
	if command.Topic == api.ImageRequestNext {
		s.nextImages = []*texturedImage{}
		for _, data := range command.Images {
			i, _ := i2t(data.ImageData())
			ti := &texturedImage{
				width:   float32(data.ImageData().Bounds().Dx()),
				height:  float32(data.ImageData().Bounds().Dy()),
				texture: i,
			}
			s.nextImages = append(s.nextImages, ti)
		}
		//s.imageView.AddImagesToNextStore(command.Images)
	} else if command.Topic == api.ImageRequestPrevious {
		s.previousImages = []*texturedImage{}
		for _, data := range command.Images {
			i, _ := i2t(data.ImageData())
			ti := &texturedImage{
				width:   float32(data.ImageData().Bounds().Dx()),
				height:  float32(data.ImageData().Bounds().Dy()),
				texture: i,
			}
			s.previousImages = append(s.previousImages, ti)
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

	f := command.Image.ImageFile().Id()
	fullImage, _ := s.imageCache.GetFull(f)
	var err error
	t, err := giu.NewTextureFromRgba(fullImage.(*image.RGBA))
	if err != nil {
		logger.Error.Print(err)
	}

	s.currentImageTexture = &texturedImage{
		width:   float32(command.Image.ImageData().Bounds().Dx()),
		height:  float32(command.Image.ImageData().Bounds().Dy()),
		texture: t,
		imageId: command.Image.ImageFile().Id(),
	}
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
