package library

import (
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
)

var nextImage = &api.ImageAtQuery{Index: 1}
var previousImage = &api.ImageAtQuery{Index: -1}

type Service struct {
	sender                            api.Sender
	service                           api.ImageLibrary
	selectedCategoryId                apitype.CategoryId
	index                             int
	imageListSize                     int
	shouldSendSimilar                 bool
	imagesChangedSincePreviousHashing bool

	api.ImageService
}

func NewImageService(sender api.Sender, library api.ImageLibrary) api.ImageService {
	return &Service{
		sender:                            sender,
		service:                           library,
		index:                             0,
		imageListSize:                     0,
		shouldSendSimilar:                 false,
		imagesChangedSincePreviousHashing: false,
		selectedCategoryId:                apitype.NoCategory,
	}
}

func (s *Service) InitializeFromDirectory(directory string) {
	s.index = 0
	s.service.InitializeFromDirectory(directory, s.sender)
	s.imagesChangedSincePreviousHashing = true
}

func (s *Service) GetImageFiles() []*apitype.ImageFile {
	return s.service.GetImages()
}

func (s *Service) ShowOnlyImages(command *api.SelectCategoryCommand) {
	s.index = 0
	s.selectedCategoryId = command.CategoryId
	s.RequestImages()
}

func (s *Service) ShowAllImages() {
	s.selectedCategoryId = apitype.NoCategory
	s.RequestImages()
}

func (s *Service) RequestGenerateHashes() {
	s.shouldSendSimilar = true
	if s.imagesChangedSincePreviousHashing && s.service.GenerateHashes(s.sender) {
		s.imagesChangedSincePreviousHashing = false
	}

	if image, _, _, err := s.getCurrentImage(); err != nil {
		s.sender.SendError("Error while generating hashes", err)
	} else {
		s.sendSimilarImages(image.ImageFile().Id())
	}
}

func (s *Service) getCurrentImage() (*apitype.ImageFileAndData, *apitype.ImageMetaData, int, error) {
	return s.service.GetImageAtIndex(s.index, s.selectedCategoryId)
}

func (s *Service) SetSendSimilarImages(command *api.SimilarImagesCommand) {
	s.shouldSendSimilar = command.SendSimilarImages
}

func (s *Service) RequestStopHashes() {
	s.service.StopHashes()
}

func (s *Service) RequestNextImageWithOffset(query *api.ImageAtQuery) {
	s.moveToNextImageWithOffset(query.Index)
	s.RequestImages()
}

func (s *Service) RequestPreviousImageWithOffset(query *api.ImageAtQuery) {
	s.moveToPreviousImageWithOffset(query.Index)
	s.RequestImages()
}

func (s *Service) RequestNextImage() {
	s.RequestNextImageWithOffset(nextImage)
}

func (s *Service) RequestPreviousImage() {
	s.RequestNextImageWithOffset(previousImage)
}

func (s *Service) RequestImage(query *api.ImageQuery) {
	s.moveToImage(query.Id)

	s.RequestImages()
}

func (s *Service) moveToImage(imageId apitype.ImageId) {
	images, _ := s.service.GetImagesInCategory(-1, 0, s.selectedCategoryId)
	for imageIndex, image := range images {
		if imageId == image.Id() {
			s.index = imageIndex
		}
	}
}

func (s *Service) moveToImageAt(index int) {
	images, _ := s.service.GetImagesInCategory(-1, 0, s.selectedCategoryId)
	if index >= 0 {
		s.index = index
	} else {
		s.index = len(images) + index
	}

	if s.index >= len(images) {
		s.index = len(images) - 1
	}
	if s.index < 0 {
		s.index = 0
	}
}

func (s *Service) RequestImageAt(query *api.ImageAtQuery) {
	s.moveToImageAt(query.Index)
	s.RequestImages()
}

func (s *Service) RequestImages() {
	s.sendImages(true)
}

func (s *Service) moveToNextImageWithOffset(offset int) {
	s.requestImageWithOffset(offset)
}

func (s *Service) moveToPreviousImageWithOffset(offset int) {
	s.requestImageWithOffset(-offset)
}

func (s *Service) requestImageWithOffset(offset int) {
	s.index += offset

	images, _ := s.service.GetImagesInCategory(-1, 0, s.selectedCategoryId)
	if s.index >= len(images) {
		s.index = len(images) - 1
	}
	if s.index < 0 {
		s.index = 0
	}
}

func (s *Service) requestImageLists() {
	s.sendImages(false)
}

func (s *Service) SetImageListSize(command *api.ImageListCommand) {
	if s.imageListSize != command.ImageListSize {
		s.imageListSize = command.ImageListSize
		s.requestImageLists()
	}
}

func (s *Service) Close() {
	logger.Info.Print("Shutting down library")
}

func (s *Service) AddImageFiles(imageList []*apitype.ImageFile) {
	s.index = 0
	if err := s.service.AddImageFiles(imageList, s.sender); err != nil {
		s.sender.SendError("Error while adding image", err)
	}
}

func (s *Service) GetImageFileById(imageId apitype.ImageId) *apitype.ImageFile {
	return s.service.GetImageFileById(imageId)
}

// Private API

func (s *Service) sendImages(sendCurrentImage bool) {
	if currentImage, metaData, currentIndex, err := s.getCurrentImage(); err != nil {
		s.sender.SendError("Error while fetching images", err)
	} else {
		totalImages := s.service.GetTotalImages(s.selectedCategoryId)
		if totalImages == 0 {
			currentIndex = 0
		}

		if sendCurrentImage {
			s.sender.SendCommandToTopic(api.ImageCurrentUpdated, &api.UpdateImageCommand{
				Image:      currentImage,
				MetaData:   metaData,
				Index:      currentIndex,
				Total:      totalImages,
				CategoryId: s.selectedCategoryId,
			})
		}

		if nextImages, err := s.service.GetNextImages(s.index, s.imageListSize, s.selectedCategoryId); err != nil {
			s.sender.SendError("Error while fetching next images", err)
		} else if previousImages, err := s.service.GetPreviousImages(s.index, s.imageListSize, s.selectedCategoryId); err != nil {
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

		if s.shouldSendSimilar {
			s.sendSimilarImages(currentImage.ImageFile().Id())
		}
	}
}

func (s *Service) sendSimilarImages(imageId apitype.ImageId) {
	if images, shouldSend, err := s.service.GetSimilarImages(imageId); err != nil {
		s.sender.SendError("Error while fetching similar images", err)
	} else if shouldSend {
		s.sender.SendCommandToTopic(api.ImageListUpdated, &api.SetImagesCommand{
			Topic:  api.ImageRequestSimilar,
			Images: images,
		})
	}
}
