package internal

import "vincit.fi/image-sorter/ui/giu/internal/guiapi"

const defaultZoomIndex = 0

type ZoomLevel struct {
	level float32
	label string
}

type ZoomStatus struct {
	zoomMode         guiapi.ZoomMode
	currentZoomIndex int32
	zoomFactor       float32
}

func NewZoomStatus() *ZoomStatus {
	return &ZoomStatus{
		currentZoomIndex: defaultZoomIndex,
		zoomMode:         guiapi.ZoomFit,
		zoomFactor:       1.0,
	}
}

func (s *ZoomStatus) SelectedZoom() int32 {
	return s.currentZoomIndex
}

func (s *ZoomStatus) ZoomMode() guiapi.ZoomMode {
	return s.zoomMode
}

func (s *ZoomStatus) ZoomLevel() float32 {
	if s.zoomMode == guiapi.ZoomFixed {
		return s.fixedZoomLevel()
	} else {
		return s.zoomFactor
	}
}

func (s *ZoomStatus) SetZoomFit() {
	s.zoomMode = guiapi.ZoomFit
	s.currentZoomIndex = 1
}

func (s *ZoomStatus) SetZoomLevel(currentZoom int32) {
	s.currentZoomIndex = currentZoom
	s.zoomFactor = s.fixedZoomLevel()
	// Must be set last, otherwise
	s.zoomMode = guiapi.ZoomFixed
}

func (s *ZoomStatus) fixedZoomLevel() float32 {
	return zoomLevels[s.currentZoomIndex].level
}

func (s *ZoomStatus) ZoomIn(currentActualZoom float32, delta float32) {
	currentZoom := s.ZoomLevel()
	if s.zoomMode == guiapi.ZoomFit {
		currentZoom = currentActualZoom
	}
	currentZoom += delta

	if currentZoom > 15 {
		currentZoom = 15
	}
	s.zoomMode = guiapi.ZoomCustom
	s.zoomFactor = currentZoom
}

func (s *ZoomStatus) ZoomOut(currentActualZoom float32, delta float32) {
	currentZoom := s.ZoomLevel()
	if s.zoomMode == guiapi.ZoomFit {
		currentZoom = currentActualZoom
	}
	currentZoom -= delta

	if currentZoom < 0.1 {
		currentZoom = 0.1
	}
	s.zoomMode = guiapi.ZoomCustom
	s.zoomFactor = currentZoom
}

func (s *ZoomStatus) ResetZoom() {
	s.currentZoomIndex = defaultZoomIndex
	s.zoomMode = guiapi.ZoomFit
}

var zoomLevels = []ZoomLevel{
	{1, "Fit"},
	{0.01, "1 %"},
	{0.05, "5 %"},
	{0.1, "10 %"},
	{0.25, "25 %"},
	{0.5, "50 %"},
	{0.75, "75 %"},
	{1, "100 %"},
	{1.25, "125 %"},
	{1.50, "150 %"},
	{1.75, "175 %"},
	{2, "200 %"},
	{3, "300 %"},
	{4, "400 %"},
	{5, "500 %"},
	{10, "1000 %"},
	{15, "1500 %"},
}

var zoomLevelLabels []string

func init() {
	for _, zoomLevel := range zoomLevels {
		zoomLevelLabels = append(zoomLevelLabels, zoomLevel.label)
	}
}

func ZoomLabels() []string {
	return zoomLevelLabels
}
