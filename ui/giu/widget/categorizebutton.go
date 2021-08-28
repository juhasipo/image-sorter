package widget

import (
	"fmt"
	"github.com/AllenDang/giu"
	"golang.org/x/image/colornames"
	"image"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
)

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
		canvas := giu.GetCanvas()
		start := giu.GetCursorScreenPos()
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
