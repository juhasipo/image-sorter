package library

import (
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/database"
	"vincit.fi/image-sorter/common/logger"
)

var nextImage = &api.ImageAtQuery{Index: 1}
var previousImage = &api.ImageAtQuery{Index: -1}

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

func (s *Manager) GetImageFiles() []*apitype.ImageFileWithMetaData {
	return s.manager.GetImages()
}

func (s *Manager) ShowOnlyImages(command *api.SelectCategoryCommand) {
	s.manager.ShowOnlyImages(command.CategoryId)
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
			s.sendSimilarImages(image.GetImageFile().GetId())
		}
	}
}

func (s *Manager) SetSendSimilarImages(command *api.SimilarImagesCommand) {
	s.manager.SetSimilarStatus(command.SendSimilarImages)
}

func (s *Manager) RequestStopHashes() {
	s.manager.StopHashes()
}

func (s *Manager) RequestNextImageWithOffset(query *api.ImageAtQuery) {
	s.manager.MoveToNextImageWithOffset(query.Index)
	s.RequestImages()
}

func (s *Manager) RequestPrevImageWithOffset(query *api.ImageAtQuery) {
	s.manager.MoveToPrevImageWithOffset(query.Index)
	s.RequestImages()
}

func (s *Manager) RequestNextImage() {

	s.RequestNextImageWithOffset(nextImage)
}

func (s *Manager) RequestPrevImage() {

	s.RequestNextImageWithOffset(previousImage)
}

func (s *Manager) RequestImage(query *api.ImageQuery) {
	s.manager.MoveToImage(query.Id)
	s.RequestImages()
}

func (s *Manager) RequestImageAt(query *api.ImageAtQuery) {
	s.manager.MoveToImageAt(query.Index)
	s.RequestImages()
}

func (s *Manager) RequestImages() {
	s.sendImages(true)
}

func (s *Manager) requestImageLists() {
	s.sendImages(false)
}

func (s *Manager) SetImageListSize(command *api.ImageListCommand) {
	if s.manager.SetImageListSize(command.ImageListSize) {
		s.requestImageLists()
	}
}

func (s *Manager) Close() {
	logger.Info.Print("Shutting down library")
}

func (s *Manager) AddImageFiles(imageList []*apitype.ImageFile) {
	if err := s.manager.AddImageFiles(imageList); err != nil {
		s.sender.SendError("Error while adding image", err)
	}
}

func (s *Manager) GetImageFileById(imageId apitype.ImageId) *apitype.ImageFileWithMetaData {
	return s.manager.GetImageFileById(imageId)
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
			s.sender.SendCommandToTopic(api.ImageCurrentUpdated, &api.UpdateImageCommand{
				Image:      currentImage,
				Index:      currentIndex,
				Total:      totalImages,
				CategoryId: s.manager.getSelectedCategoryId(),
			})
		}

		if nextImages, err := s.manager.getNextImages(); err != nil {
			s.sender.SendError("Error while fetching next images", err)
		} else if prevImages, err := s.manager.getPrevImages(); err != nil {
			s.sender.SendError("Error while fetching previous images", err)
		} else {
			s.sender.SendCommandToTopic(api.ImageListUpdated, &api.SetImagesCommand{
				Topic:  api.ImageRequestPrev,
				Images: prevImages,
			})
			s.sender.SendCommandToTopic(api.ImageListUpdated, &api.SetImagesCommand{
				Topic:  api.ImageRequestNext,
				Images: nextImages,
			})
		}

		if s.manager.shouldSendSimilarImages() {
			s.sendSimilarImages(currentImage.GetImageFile().GetId())
		}
	}
}

func (s *Manager) sendSimilarImages(imageId apitype.ImageId) {
	if images, shouldSend, err := s.manager.getSimilarImages(imageId); err != nil {
		s.sender.SendError("Error while fetching similar images", err)
	} else if shouldSend {
		s.sender.SendCommandToTopic(api.ImageListUpdated, &api.SetImagesCommand{
			Topic:  api.ImageRequestSimilar,
			Images: images,
		})
	}
}

func (s *Manager) loadImagesFromRootDir() {
	s.manager.updateImages()
}
