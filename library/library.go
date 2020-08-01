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

func LibraryNew(sender event.Sender, imageCache imageloader.ImageStore, imageLoader imageloader.ImageLoader) Library {
	return &Manager{
		manager: libraryNew(imageCache, imageLoader),
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
	s.sendStatus(true)
}

func (s *Manager) ShowAllImages() {
	s.ShowAllImages()
	s.sendStatus(true)
}

func (s *Manager) RequestGenerateHashes() {
	if s.manager.RequestGenerateHashes(s.sender) {
		image, _ := s.manager.getCurrentImage()
		s.sendSimilarImages(image.GetHandle())
	}
}

func (s *Manager) SetSimilarStatus(sendSimilarImages bool) {
	s.manager.SetSimilarStatus(sendSimilarImages)
}

func (s *Manager) RequestStopHashes() {
	s.manager.RequestStopHashes()
}

func (s *Manager) RequestNextImage() {
	s.RequestNextImageWithOffset(1)
}

func (s *Manager) RequestNextImageWithOffset(offset int) {
	s.manager.RequestNextImageWithOffset(offset)
	s.sendStatus(true)
}

func (s *Manager) RequestImage(handle *common.Handle) {
	s.manager.RequestImage(handle)
	s.RequestImages()
}

func (s *Manager) RequestImageAt(index int) {
	s.manager.RequestImageAt(index)
	s.RequestImages()
}

func (s *Manager) RequestPrevImage() {
	s.RequestPrevImageWithOffset(1)
}

func (s *Manager) RequestPrevImageWithOffset(offset int) {
	s.manager.RequestPrevImageWithOffset(offset)
	s.sendStatus(true)
}

func (s *Manager) RequestImages() {
	s.sendStatus(true)
}

func (s *Manager) ChangeImageListSize(imageListSize int) {
	if s.manager.ChangeImageListSize(imageListSize) {
		s.sendStatus(false)
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

func (s *Manager) sendStatus(sendCurrentImage bool) {
	currentImage, currentIndex := s.manager.getCurrentImage()
	totalImages := s.manager.getTotalImages()
	if totalImages == 0 {
		currentIndex = 0
	}
	if sendCurrentImage {
		s.sender.SendToTopicWithData(event.IMAGE_CURRENT_UPDATE,
			currentImage, currentIndex, totalImages,
			s.manager.getCurrentCategoryName(),
			s.manager.GetMetaData(currentImage.GetHandle()))
	}

	s.sender.SendToTopicWithData(event.IMAGE_LIST_UPDATE, event.IMAGE_REQUEST_NEXT, s.manager.getNextImages())
	s.sender.SendToTopicWithData(event.IMAGE_LIST_UPDATE, event.IMAGE_REQUEST_PREV, s.manager.getPrevImages())

	if s.manager.shouldSendSimilarImages() {
		s.sendSimilarImages(currentImage.GetHandle())
	}
}

func (s *Manager) sendSimilarImages(handle *common.Handle) {
	images, shouldSend := s.manager.getSimilarImages(handle)
	if shouldSend {
		s.sender.SendToTopicWithData(event.IMAGE_LIST_UPDATE, event.IMAGE_REQUEST_SIMILAR, images, 0, 0, "")
	}
}

func (s *Manager) loadImagesFromRootDir() {
	s.manager.loadImagesFromRootDir()
}
