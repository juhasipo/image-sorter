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

func NewLibrary(sender api.Sender, imageCache api.ImageStore, imageLoader api.ImageLoader, similarityIndex *database.SimilarityIndex, imageStore *database.ImageStore) api.Library {
	return &Manager{
		sender:  sender,
		manager: newLibrary(imageCache, imageLoader, similarityIndex, imageStore),
	}
}

func (s *Manager) InitializeFromDirectory(directory string) {
	s.manager.InitializeFromDirectory(directory)
}

func (s *Manager) GetHandles() []*apitype.Handle {
	return s.manager.GetHandles()
}

func (s *Manager) ShowOnlyImages(title string) {
	s.manager.ShowOnlyImages(title)
	s.RequestImages()
}

func (s *Manager) ShowAllImages() {
	s.manager.ShowAllImages()
	s.RequestImages()
}

func (s *Manager) RequestGenerateHashes() {
	if s.manager.GenerateHashes(s.sender) {
		if image, _, err := s.manager.getCurrentImage(); err != nil {
			s.sender.SendError("Error while generating hashes", err)
		} else {
			s.sendSimilarImages(image.GetHandle())
		}
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
	if err := s.manager.AddHandles(imageList); err != nil {
		s.sender.SendError("Error while adding image", err)
	}
}

func (s *Manager) GetHandleById(handleId apitype.HandleId) *apitype.Handle {
	return s.manager.GetHandleById(handleId)
}

func (s *Manager) GetMetaData(handle *apitype.Handle) *apitype.ExifData {
	return s.manager.GetMetaData(handle)
}

// Private API

func (s *Manager) sendImages(sendCurrentImage bool) {
	if currentImage, currentIndex, err := s.manager.getCurrentImage(); err != nil {
		s.sender.SendError("Error while fetching images", err)
	} else {
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

		if nextImages, err := s.manager.getNextImages(); err != nil {
			s.sender.SendError("Error while fetching next images", err)
		} else if prevImages, err := s.manager.getPrevImages(); err != nil {
			s.sender.SendError("Error while fetching previous images", err)
		} else {
			s.sender.SendToTopicWithData(api.ImageListUpdated, api.ImageRequestPrev, prevImages)
			s.sender.SendToTopicWithData(api.ImageListUpdated, api.ImageRequestNext, nextImages)
		}

		if s.manager.shouldSendSimilarImages() {
			s.sendSimilarImages(currentImage.GetHandle())
		}
	}
}

func (s *Manager) sendSimilarImages(handle *apitype.Handle) {
	if images, shouldSend, err := s.manager.getSimilarImages(handle); err != nil {
		s.sender.SendError("Error while fetching similar images", err)
	} else if shouldSend {
		s.sender.SendToTopicWithData(api.ImageListUpdated, api.ImageRequestSimilar, images)
	}
}

func (s *Manager) loadImagesFromRootDir() {
	s.manager.updateImages()
}
