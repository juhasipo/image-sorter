package library

import (
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/database"
	"vincit.fi/image-sorter/common/logger"
)

var nextImage = &api.ImageAtQuery{Index: 1}
var previousImage = &api.ImageAtQuery{Index: -1}

type Service struct {
	sender  api.Sender
	service *internalService

	api.ImageService
}

func NewImageService(sender api.Sender, imageCache api.ImageStore, imageLoader api.ImageLoader, similarityIndex *database.SimilarityIndex, imageStore *database.ImageStore, imageMetaDataStore *database.ImageMetaDataStore) api.ImageService {
	return &Service{
		sender:  sender,
		service: newImageService(imageCache, imageLoader, similarityIndex, imageStore, imageMetaDataStore),
	}
}

func (s *Service) InitializeFromDirectory(directory string) {
	s.service.InitializeFromDirectory(directory)
}

func (s *Service) GetImageFiles() []*apitype.ImageFile {
	return s.service.GetImages()
}

func (s *Service) ShowOnlyImages(command *api.SelectCategoryCommand) {
	s.service.ShowOnlyImages(command.CategoryId)
	s.RequestImages()
}

func (s *Service) ShowAllImages() {
	s.service.ShowAllImages()
	s.RequestImages()
}

func (s *Service) RequestGenerateHashes() {
	if s.service.GenerateHashes(s.sender) {
		if image, _, _, err := s.service.getCurrentImage(); err != nil {
			s.sender.SendError("Error while generating hashes", err)
		} else {
			s.sendSimilarImages(image.ImageFile().Id())
		}
	}
}

func (s *Service) SetSendSimilarImages(command *api.SimilarImagesCommand) {
	s.service.SetSimilarStatus(command.SendSimilarImages)
}

func (s *Service) RequestStopHashes() {
	s.service.StopHashes()
}

func (s *Service) RequestNextImageWithOffset(query *api.ImageAtQuery) {
	s.service.MoveToNextImageWithOffset(query.Index)
	s.RequestImages()
}

func (s *Service) RequestPreviousImageWithOffset(query *api.ImageAtQuery) {
	s.service.MoveToPreviousImageWithOffset(query.Index)
	s.RequestImages()
}

func (s *Service) RequestNextImage() {
	s.RequestNextImageWithOffset(nextImage)
}

func (s *Service) RequestPreviousImage() {
	s.RequestNextImageWithOffset(previousImage)
}

func (s *Service) RequestImage(query *api.ImageQuery) {
	s.service.MoveToImage(query.Id)
	s.RequestImages()
}

func (s *Service) RequestImageAt(query *api.ImageAtQuery) {
	s.service.MoveToImageAt(query.Index)
	s.RequestImages()
}

func (s *Service) RequestImages() {
	s.sendImages(true)
}

func (s *Service) requestImageLists() {
	s.sendImages(false)
}

func (s *Service) SetImageListSize(command *api.ImageListCommand) {
	if s.service.SetImageListSize(command.ImageListSize) {
		s.requestImageLists()
	}
}

func (s *Service) Close() {
	logger.Info.Print("Shutting down library")
}

func (s *Service) AddImageFiles(imageList []*apitype.ImageFile) {
	if err := s.service.AddImageFiles(imageList); err != nil {
		s.sender.SendError("Error while adding image", err)
	}
}

func (s *Service) GetImageFileById(imageId apitype.ImageId) *apitype.ImageFile {
	return s.service.GetImageFileById(imageId)
}

// Private API

func (s *Service) sendImages(sendCurrentImage bool) {
	if currentImage, metaData, currentIndex, err := s.service.getCurrentImage(); err != nil {
		s.sender.SendError("Error while fetching images", err)
	} else {
		totalImages := s.service.getTotalImages()
		if totalImages == 0 {
			currentIndex = 0
		}

		if sendCurrentImage {
			s.sender.SendCommandToTopic(api.ImageCurrentUpdated, &api.UpdateImageCommand{
				Image:      currentImage,
				MetaData:   metaData,
				Index:      currentIndex,
				Total:      totalImages,
				CategoryId: s.service.getSelectedCategoryId(),
			})
		}

		if nextImages, err := s.service.getNextImages(); err != nil {
			s.sender.SendError("Error while fetching next images", err)
		} else if previousImages, err := s.service.getPreviousImages(); err != nil {
			s.sender.SendError("Error while fetching previous images", err)
		} else {
			s.sender.SendCommandToTopic(api.ImageListUpdated, &api.SetImagesCommand{
				Topic:  api.ImageRequestPrevious,
				Images: previousImages,
			})
			s.sender.SendCommandToTopic(api.ImageListUpdated, &api.SetImagesCommand{
				Topic:  api.ImageRequestNext,
				Images: nextImages,
			})
		}

		if s.service.shouldSendSimilarImages() {
			s.sendSimilarImages(currentImage.ImageFile().Id())
		}
	}
}

func (s *Service) sendSimilarImages(imageId apitype.ImageId) {
	if images, shouldSend, err := s.service.getSimilarImages(imageId); err != nil {
		s.sender.SendError("Error while fetching similar images", err)
	} else if shouldSend {
		s.sender.SendCommandToTopic(api.ImageListUpdated, &api.SetImagesCommand{
			Topic:  api.ImageRequestSimilar,
			Images: images,
		})
	}
}
