package widget

import (
	"github.com/AllenDang/giu"
	"image"
	"image/color"
	"sync"
	"vincit.fi/image-sorter/api/apitype"
)

var imageHoverOverlayColor = color.RGBA{R: 255, G: 255, B: 255, A: 64}

type HorizontalImageListWidget struct {
	images           []*TexturedImage
	showLabel        bool
	width            float32
	reverse          bool
	shrink           bool
	height           float32
	onClick          func(*apitype.ImageFile)
	highlightedImage *apitype.ImageFile
	mux              sync.Mutex
}

func HorizontalImageList(onClick func(*apitype.ImageFile), showLabel bool, reverse bool, shrink bool) *HorizontalImageListWidget {
	return &HorizontalImageListWidget{
		images:    []*TexturedImage{},
		showLabel: showLabel,
		width:     0,
		height:    0,
		reverse:   reverse,
		shrink:    shrink,
		onClick:   onClick,
	}
}

func (s *HorizontalImageListWidget) Size(width float32, height float32) *HorizontalImageListWidget {
	s.width = width
	s.height = height
	return s
}

func (s *HorizontalImageListWidget) SetImages(images []*TexturedImage) *HorizontalImageListWidget {
	s.mux.Lock()
	defer s.mux.Unlock()

	var newHorizontalImageList []*TexturedImage
	var changedImages []*TexturedImage

	// Very naive algorithm to find out which images
	// need to be reloaded and which can be re-used
	for _, image := range images {
		var found *TexturedImage = nil
		for _, texturedImage := range s.images {
			if image.IsSame(texturedImage) {
				found = texturedImage
			}
		}

		if found == nil {
			changedImages = append(changedImages, image)
			newHorizontalImageList = append(newHorizontalImageList, image)
		} else {
			newHorizontalImageList = append(newHorizontalImageList, found)
		}
	}

	s.images = newHorizontalImageList

	for _, texturedImage := range changedImages {
		texturedImage.LoadImageAsTextureThumbnail()
	}

	return s
}

func (s *HorizontalImageListWidget) HighlightedImage() *apitype.ImageFile {
	return s.highlightedImage
}

func (s *HorizontalImageListWidget) Build() {
	giu.Child().
		Layout(giu.Custom(func() {
			regionWidth, _ := giu.GetAvailableRegion()

			marginX := float32(8)

			pos := giu.GetCursorScreenPos()
			canvas := giu.GetCanvas()
			startX := float32(0)

			s.highlightedImage = nil
			for i, img := range s.images {
				factor := float32(1)
				if s.shrink {
					factor = 1 - (float32(i+1) * float32(0.05))
				}

				imageRatio := img.Width() / img.Height()
				targetHeight := s.height * factor

				width := imageRatio * targetHeight
				offsetHeight := (s.height - targetHeight) / 2

				start := image.Point{
					X: int(startX),
					Y: int(offsetHeight),
				}
				startX += width
				end := image.Point{
					X: int(startX),
					Y: int(targetHeight + offsetHeight),
				}
				startX += marginX

				if s.reverse {
					tmp := int(regionWidth - float32(start.X))
					start.X = int(regionWidth - float32(end.X))
					end.X = tmp
				}

				if img.Texture() != nil {
					start.X += pos.X
					start.Y += pos.Y
					end.X += pos.X
					end.Y += pos.Y
					mousePos := giu.GetMousePos()

					imgArea := image.Rectangle{
						Min: start,
						Max: end,
					}
					canvas.AddImage(img.Texture(), start, end)
					if mousePos.In(imgArea) {
						s.highlightedImage = s.images[i].Image()
						giu.SetMouseCursor(giu.MouseCursorHand)
						if giu.IsMouseClicked(giu.MouseButtonLeft) {
							s.onClick(s.images[i].Image())
						}
						canvas.AddRectFilled(start, end, imageHoverOverlayColor, 0, giu.DrawFlagsNone)
					}
				}
			}
		})).
		Border(false).
		Size(s.width, s.height).
		Flags(giu.WindowFlagsNoScrollbar | giu.WindowFlagsNoScrollWithMouse).
		Build()
}
