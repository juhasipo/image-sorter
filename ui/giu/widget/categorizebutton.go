package widget

import (
	"fmt"
	"github.com/AllenDang/giu"
	"image"
	"image/color"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/ui/giu/guiapi"
)

type CategoryButtonWidget struct {
	categoryId apitype.CategoryId
	name       string
	shortcut   string
	active     bool
	highlight  bool
	onClick    func(*guiapi.CategoryAction)
	onShowOnly func()
	giu.Widget
}

func CategoryButton(categoryId apitype.CategoryId, name string, shortcut string, active bool, highlight bool, onClick func(action *guiapi.CategoryAction)) *CategoryButtonWidget {
	return &CategoryButtonWidget{
		categoryId: categoryId,
		name:       name,
		shortcut:   shortcut,
		active:     active,
		highlight:  highlight,
		onClick:    onClick,
	}
}

const categoryPrimaryButtonWidth = 100
const categoryArrowButtonWidth = 20
const categoryPrimaryButtonHeight = 20
const categoryIndicatorButtonHeight = 5

var (
	statusActiveColor   = color.RGBA{R: 64, G: 255, B: 64, A: 255}
	statusDisabledColor = color.RGBA{R: 0, G: 0, B: 32, A: 255}
)

func (s *CategoryButtonWidget) Build() {
	operation := apitype.CATEGORIZE
	if s.active {
		operation = apitype.UNCATEGORIZE
	}

	categorizeAction := func(operation apitype.Operation, stayOnImage bool, forceCategory bool) {
		s.onClick(&guiapi.CategoryAction{
			StayOnImage:      stayOnImage,
			ForceCategory:    forceCategory,
			ShowOnlyCategory: false,
		})
	}

	showOnly := func() {
		s.onClick(&guiapi.CategoryAction{
			ShowOnlyCategory: true,
		})
	}

	actionName := "Add"
	if s.active {
		actionName = "Remove"
	}

	showOnlyLabel := "Show only"
	if s.highlight {
		showOnlyLabel = "Show all"
	}
	menuName := fmt.Sprintf("CategoryMenu-%d", s.categoryId)
	const menuButtonWidth = 210
	const menuButtonHeight = 30
	menu := giu.Popup(menuName).
		Layout(
			giu.Button(actionName+" category"+" ("+s.shortcut+")").OnClick(func() {
				categorizeAction(operation, false, false)
				giu.CloseCurrentPopup()
			}).Size(menuButtonWidth, menuButtonHeight),
			giu.Button(actionName+" category and stay"+" (Shift + "+s.shortcut+")").OnClick(func() {
				categorizeAction(operation, true, false)
				giu.CloseCurrentPopup()
			}).Size(menuButtonWidth, menuButtonHeight),
			giu.Button("Force to category"+" (Ctrl + "+s.shortcut+")").OnClick(func() {
				categorizeAction(apitype.CATEGORIZE, false, true)
				giu.CloseCurrentPopup()
			}).Size(menuButtonWidth, menuButtonHeight),
			giu.Button(showOnlyLabel+" (Alt + "+s.shortcut+")").OnClick(func() {
				showOnly()
				giu.CloseCurrentPopup()
			}).Size(menuButtonWidth, menuButtonHeight),
		)

	primaryButton := giu.Button(s.name).
		Size(categoryPrimaryButtonWidth, categoryPrimaryButtonHeight).
		OnClick(func() {
			stayOnImage := giu.IsKeyDown(giu.KeyLeftShift) || giu.IsKeyDown(giu.KeyRightShift)
			forceCategory := giu.IsKeyDown(giu.KeyLeftControl) || giu.IsKeyDown(giu.KeyRightControl)
			categorizeAction(operation, stayOnImage, forceCategory)
		})
	menuButton := giu.ArrowButton("Menu", giu.DirectionDown).OnClick(func() {
		giu.OpenPopup(menuName)
	})

	statusColor := statusActiveColor
	if !s.active {
		statusColor = statusDisabledColor
	}

	style := giu.Style()
	if s.highlight {
		style.SetColor(giu.StyleColorButton, color.RGBA{R: 140, G: 184, B: 255, A: 255})
	}

	statusIndicator := giu.Custom(func() {
		canvas := giu.GetCanvas()
		start := giu.GetCursorScreenPos()
		end := start.Add(image.Pt(categoryPrimaryButtonWidth+categoryArrowButtonWidth, categoryIndicatorButtonHeight))

		canvas.AddRectFilled(start, end, statusColor, 0, 0)
	})

	style.
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
