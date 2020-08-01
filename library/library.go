package library

import (
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/imageloader"
	"vincit.fi/image-sorter/logger"
)

type Manager struct {
	sender  event.Sender
	manager *internalManager

	Library
}

func NewLibrary(sender event.Sender, imageCache imageloader.ImageStore, imageLoader imageloader.ImageLoader) Library {
	return &Manager{
		sender:  sender,
		manager: newLibrary(imageCache, imageLoader),
	}
}

func (s *Manager) InitializeFromDirectory(directory string) {
	s.manager.InitializeFromDirectory(directory)
}

func (s *Manager) GetHandles() []*common.Handle {
	return s.manager.GetHandles()
}

func (s *Manager) ShowOnlyImages(title string, handles []*common.Handle) {
	s.manager.ShowOnlyImages(title, handles)
	s.RequestImages()
}

func (s *Manager) ShowAllImages() {
	s.manager.ShowAllImages()
	s.RequestImages()
}

func (s *Manager) RequestGenerateHashes() {
	if s.manager.GenerateHashes(s.sender) {
		image, _ := s.manager.getCurrentImage()
		s.sendSimilarImages(image.GetHandle())
	}
}

func (s *Manager) SetSendSimilarImages(sendSimilarImages bool) {
	s.manager.SetSimilarStatus(sendSimilarImages)
}

func (s *Manager) RequestStopHashes() {
	s.manager.StopHashes()
}

func (s *Manager) RequestNextImageWithOffset(offset int) {
	s.manager.MoveToNextImageWithOffset(offset)
	s.RequestImages()
}

func (s *Manager) RequestPrevImageWithOffset(offset int) {
	s.manager.MoveToPrevImageWithOffset(offset)
	s.RequestImages()
}

func (s *Manager) RequestNextImage() {
	s.RequestNextImageWithOffset(1)
}

func (s *Manager) RequestPrevImage() {
	s.RequestNextImageWithOffset(-1)
}

func (s *Manager) RequestImage(handle *common.Handle) {
	s.manager.MoveToImage(handle)
	s.RequestImages()
}

func (s *Manager) RequestImageAt(index int) {
	s.manager.MoveToImageAt(index)
	s.RequestImages()
}

func (s *Manager) RequestImages() {
	s.sendImages(true)
}

func (s *Manager) requestImageLists() {
	s.sendImages(false)
}

func (s *Manager) SetImageListSize(imageListSize int) {
	if s.manager.SetImageListSize(imageListSize) {
		s.requestImageLists()
	}
}

func (s *Manager) Close() {
	logger.Info.Print("Shutting down library")
}

func (s *Manager) AddHandles(imageList []*common.Handle) {
	s.manager.AddHandles(imageList)
}

func (s *Manager) GetHandleById(handleId string) *common.Handle {
	return s.manager.GetHandleById(handleId)
}

func (s *Manager) GetMetaData(handle *common.Handle) *common.ExifData {
	return s.manager.GetMetaData(handle)
}

// Private API

func (s *Manager) sendImages(sendCurrentImage bool) {
	currentImage, currentIndex := s.manager.getCurrentImage()
	totalImages := s.manager.getTotalImages()
	if totalImages == 0 {
		currentIndex = 0
	}

	if sendCurrentImage {
		s.sender.SendToTopicWithData(event.ImageCurrentUpdated,
			currentImage, currentIndex, totalImages,
			s.manager.getCurrentCategoryName(),
			s.manager.GetMetaData(currentImage.GetHandle()))
	}

	s.sender.SendToTopicWithData(event.ImageListUpdated, event.ImageRequestNext, s.manager.getNextImages())
	s.sender.SendToTopicWithData(event.ImageListUpdated, event.ImageRequestPrev, s.manager.getPrevImages())

	if s.manager.shouldSendSimilarImages() {
		s.sendSimilarImages(currentImage.GetHandle())
	}
}

func (s *Manager) sendSimilarImages(handle *common.Handle) {
	images, shouldSend := s.manager.getSimilarImages(handle)
	if shouldSend {
		s.sender.SendToTopicWithData(event.ImageListUpdated, event.ImageRequestSimilar, images)
	}
}

func (s *Manager) loadImagesFromRootDir() {
	s.manager.loadImagesFromRootDir()
}
