package library

import (
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/database"
	"vincit.fi/image-sorter/common/logger"
)

type Manager struct {
	sender  api.Sender
	manager *internalManager

	api.Library
}

func NewLibrary(sender api.Sender, imageCache api.ImageStore, imageLoader api.ImageLoader, store *database.Store) api.Library {
	return &Manager{
		sender:  sender,
		manager: newLibrary(imageCache, imageLoader, store),
	}
}

func (s *Manager) InitializeFromDirectory(directory string) {
	s.manager.InitializeFromDirectory(directory)
}

func (s *Manager) GetHandles() []*apitype.Handle {
	return s.manager.GetHandles()
}

func (s *Manager) ShowOnlyImages(title string, handles []*apitype.Handle) {
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

func (s *Manager) RequestImage(handle *apitype.Handle) {
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

func (s *Manager) AddHandles(imageList []*apitype.Handle) {
	s.manager.AddHandles(imageList)
}

func (s *Manager) GetHandleById(handleId int64) *apitype.Handle {
	return s.manager.GetHandleById(handleId)
}

func (s *Manager) GetMetaData(handle *apitype.Handle) *apitype.ExifData {
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
		s.sender.SendToTopicWithData(api.ImageCurrentUpdated,
			currentImage, currentIndex, totalImages,
			s.manager.getCurrentCategoryName(),
			s.manager.GetMetaData(currentImage.GetHandle()))
	}

	s.sender.SendToTopicWithData(api.ImageListUpdated, api.ImageRequestNext, s.manager.getNextImages())
	s.sender.SendToTopicWithData(api.ImageListUpdated, api.ImageRequestPrev, s.manager.getPrevImages())

	if s.manager.shouldSendSimilarImages() {
		s.sendSimilarImages(currentImage.GetHandle())
	}
}

func (s *Manager) sendSimilarImages(handle *apitype.Handle) {
	images, shouldSend := s.manager.getSimilarImages(handle)
	if shouldSend {
		s.sender.SendToTopicWithData(api.ImageListUpdated, api.ImageRequestSimilar, images)
	}
}

func (s *Manager) loadImagesFromRootDir() {
	s.manager.loadImagesFromRootDir()
}
