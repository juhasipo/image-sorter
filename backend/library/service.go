package library

import (
	"sync"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
)

var nextImage = &api.ImageAtQuery{Index: 1}
var previousImage = &api.ImageAtQuery{Index: -1}

type Service struct {
	sender                            api.Sender
	library                           api.ImageLibrary
	selectedCategoryId                apitype.CategoryId
	index                             int
	imageListSize                     int
	shouldSendSimilar                 bool
	imagesChangedSincePreviousHashing bool
	imageLoadMux                      sync.Mutex

	api.ImageService
}

func NewImageService(sender api.Sender, library api.ImageLibrary) *Service {
	return &Service{
		sender:                            sender,
		library:                           library,
		index:                             0,
		imageListSize:                     0,
		shouldSendSimilar:                 false,
		imagesChangedSincePreviousHashing: false,
		selectedCategoryId:                apitype.NoCategory,
		imageLoadMux:                      sync.Mutex{},
	}
}

func (s *Service) InitializeFromDirectory(directory string) {
	s.imageLoadMux.Lock()
	defer s.imageLoadMux.Unlock()
	s.index = 0
	s.library.InitializeFromDirectory(directory)
	s.imagesChangedSincePreviousHashing = true
}

func (s *Service) GetImageFiles() []*apitype.ImageFile {
	s.imageLoadMux.Lock()
	defer s.imageLoadMux.Unlock()
	return s.library.GetImages()
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
	if s.imagesChangedSincePreviousHashing && s.library.GenerateHashes() {
		s.imagesChangedSincePreviousHashing = false
	}

	if image, _, _, err := s.getCurrentImage(); err != nil {
		s.sender.SendError("Error while generating hashes", err)
	} else {
		s.sendSimilarImages(image.ImageFile().Id())
	}
}

func (s *Service) getCurrentImage() (*apitype.ImageFileAndData, *apitype.ImageMetaData, int, error) {
	return s.library.GetImageAtIndex(s.index, s.selectedCategoryId)
}

func (s *Service) SetSendSimilarImages(command *api.SimilarImagesCommand) {
	s.shouldSendSimilar = command.SendSimilarImages
}

func (s *Service) RequestStopHashes() {
	s.library.StopHashes()
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
	s.index = s.findImageIndex(imageId, s.selectedCategoryId)
}

func (s *Service) findImageIndex(imageId apitype.ImageId, categoryId apitype.CategoryId) int {
	images, _ := s.library.GetImagesInCategory(-1, 0, categoryId)
	for imageIndex, image := range images {
		if imageId == image.Id() {
			return imageIndex
		}
	}
	return 0
}

func (s *Service) moveToImageAt(index int) {
	count := s.library.GetTotalImages(s.selectedCategoryId)
	newIndex := s.calculateNewIndexAndWrapNegative(index, count)

	s.index = newIndex
}

func (s *Service) calculateNewIndexAndWrapNegative(index int, count int) int {
	newIndex := 0
	if index >= 0 {
		newIndex = index
	} else {
		newIndex = count + index
	}

	if newIndex >= count {
		newIndex = count - 1
	}

	if newIndex < 0 {
		newIndex = 0
	}
	return newIndex
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
	count := s.library.GetTotalImages(s.selectedCategoryId)
	s.index = s.calculateIndexOffsetAndClamp(s.index, offset, count)
}

func (s *Service) calculateIndexOffsetAndClamp(oldIndex int, offset int, count int) int {
	newIndex := oldIndex + offset

	if newIndex >= count {
		newIndex = count - 1
	}
	if newIndex < 0 {
		newIndex = 0
	}
	return newIndex
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
	if err := s.library.AddImageFiles(imageList); err != nil {
		s.sender.SendError("Error while adding image", err)
	}
}

func (s *Service) GetImageFileById(imageId apitype.ImageId) *apitype.ImageFile {
	return s.library.GetImageFileById(imageId)
}

// Private API

func (s *Service) sendImages(sendCurrentImage bool) {
	s.imageLoadMux.Lock()
	defer s.imageLoadMux.Unlock()

	if currentImage, metaData, currentIndex, err := s.getCurrentImage(); err != nil {
		s.sender.SendError("Error while fetching images", err)
	} else {
		totalImages := s.library.GetTotalImages(s.selectedCategoryId)
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

		if nextImages, err := s.library.GetNextImages(s.index, s.imageListSize, s.selectedCategoryId); err != nil {
			s.sender.SendError("Error while fetching next images", err)
		} else if previousImages, err := s.library.GetPreviousImages(s.index, s.imageListSize, s.selectedCategoryId); err != nil {
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
	if images, shouldSend, err := s.library.GetSimilarImages(imageId); err != nil {
		s.sender.SendError("Error while fetching similar images", err)
	} else if shouldSend {
		s.sender.SendCommandToTopic(api.ImageListUpdated, &api.SetImagesCommand{
			Topic:  api.ImageRequestSimilar,
			Images: images,
		})
	}
}
