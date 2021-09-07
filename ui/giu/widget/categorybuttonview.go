package widget

import "github.com/AllenDang/giu"

type CategoryButtonViewWidget struct {
	buttons []giu.Widget
}

func CategoryButtonView(buttons []*CategoryButtonWidget) *CategoryButtonViewWidget {
	var widgets []giu.Widget

	for _, button := range buttons {
		widgets = append(widgets, button)
	}

	return &CategoryButtonViewWidget{
		buttons: widgets,
	}
}

const approximateButtonWidth = categoryPrimaryButtonWidth + categoryArrowButtonWidth + 8

func (s *CategoryButtonViewWidget) Build() {
	giu.Custom(func() {
		regionWidth, _ := giu.GetAvailableRegion()

		approximateButtonsWidth := float32(len(s.buttons) * approximateButtonWidth)

		offsetWidth := (regionWidth - approximateButtonsWidth) / 2.0
		if offsetWidth < 0 {
			offsetWidth = 0
		}

		giu.Row(
			giu.Dummy(offsetWidth, 0),
			giu.Row(s.buttons...),
		).Build()
	}).Build()
}
