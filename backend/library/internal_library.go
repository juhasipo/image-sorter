package library

import (
	"runtime"
	"time"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/database"
	"vincit.fi/image-sorter/common/logger"
)

var (
	emptyHandles = []*apitype.ImageContainer{}
)

const maxSimilarImages = 20

type ImageList func(number int) []*apitype.Handle

type internalManager struct {
	rootDir                     string
	selectedCategoryId          apitype.CategoryId
	index                       int
	shouldSendSimilar           bool
	shouldGenerateSimilarHashed bool
	categoryManager             api.CategoryManager
	imageListSize               int
	imageCache                  api.ImageStore
	imageLoader                 api.ImageLoader
	similarityIndex             *database.SimilarityIndex
	imageStore                  *database.ImageStore
	hashCalculator              *HashCalculator
}

func newLibrary(imageCache api.ImageStore, imageLoader api.ImageLoader,
	similarityIndex *database.SimilarityIndex, imageStore *database.ImageStore) *internalManager {
	var manager = internalManager{
		index:                       0,
		shouldGenerateSimilarHashed: true,
		shouldSendSimilar:           true,
		imageListSize:               0,
		imageCache:                  imageCache,
		imageLoader:                 imageLoader,
		similarityIndex:             similarityIndex,
		imageStore:                  imageStore,
		selectedCategoryId:          apitype.NoCategory,
	}
	return &manager
}

func (s *internalManager) InitializeFromDirectory(directory string) {
	s.rootDir = directory
	s.index = 0
	s.shouldGenerateSimilarHashed = true
	s.updateImages()
}

func (s *internalManager) GetHandles() []*apitype.Handle {
	images, _ := s.imageStore.GetAllImages()
	return images
}

func (s *internalManager) ShowOnlyImages(categoryId apitype.CategoryId) {
	s.selectedCategoryId = categoryId
	s.index = 0
}

func (s *internalManager) ShowAllImages() {
	s.selectedCategoryId = apitype.NoCategory
}

func (s *internalManager) GenerateHashes(sender api.Sender) bool {
	s.hashCalculator = NewHashCalculator(s.similarityIndex, s.imageLoader, s.getThreadCount())

	shouldSendSimilarImages := false
	s.shouldSendSimilar = true
	if s.shouldGenerateSimilarHashed {
		images, _ := s.imageStore.GetAllImages()
		hashes, err := s.hashCalculator.GenerateHashes(images, func(current int, total int) {
			sender.SendToTopicWithData(api.ProcessStatusUpdated, "hash", current, total)
		})

		if err == nil {
			err = s.hashCalculator.BuildSimilarityIndex(hashes, func(current int, total int) {
				sender.SendToTopicWithData(api.ProcessStatusUpdated, "similarity-index", current, total)
			})
		}

		if err != nil {
			sender.SendError("Error while saving hashes", err)
		}

		// Always send 100% status even if cancelled so that the progress bar is hidden
		sender.SendToTopicWithData(api.ProcessStatusUpdated, "hash", 0, 0)

		// Only send if not cancelled or no error
		if err == nil {
			shouldSendSimilarImages = true
		}
		s.shouldGenerateSimilarHashed = false
	} else {
		shouldSendSimilarImages = true
	}

	s.hashCalculator = nil
	return shouldSendSimilarImages
}

func (s *internalManager) SetSimilarStatus(sendSimilarImages bool) {
	s.shouldSendSimilar = sendSimilarImages
}

func (s *internalManager) StopHashes() {
	if s.hashCalculator != nil {
		s.hashCalculator.StopHashes()
	}
}

func (s *internalManager) MoveToImage(handle *apitype.Handle) {
	images, _ := s.imageStore.GetImagesInCategory(-1, 0, s.selectedCategoryId)
	for imageIndex, image := range images {
		if handle.GetId() == image.GetId() {
			s.index = imageIndex
		}
	}
}

func (s *internalManager) MoveToImageAt(index int) {
	images, _ := s.imageStore.GetImagesInCategory(-1, 0, s.selectedCategoryId)
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

func (s *internalManager) RequestNextImage() {
	s.MoveToNextImageWithOffset(1)
}

func (s *internalManager) RequestPrevImage() {
	s.MoveToPrevImageWithOffset(1)
}

func (s *internalManager) MoveToNextImageWithOffset(offset int) {
	s.requestImageWithOffset(offset)
}

func (s *internalManager) MoveToPrevImageWithOffset(offset int) {
	s.requestImageWithOffset(-offset)
}

func (s *internalManager) requestImageWithOffset(offset int) {
	s.index += offset

	images, _ := s.imageStore.GetImagesInCategory(-1, 0, s.selectedCategoryId)
	if s.index >= len(images) {
		s.index = len(images) - 1
	}
	if s.index < 0 {
		s.index = 0
	}
}

func (s *internalManager) SetImageListSize(imageListSize int) bool {
	if s.imageListSize != imageListSize {
		s.imageListSize = imageListSize
		return true
	} else {
		return false
	}
}

func (s *internalManager) AddHandles(imageList []*apitype.Handle) error {
	s.index = 0
	start := time.Now()
	if err := s.imageStore.AddImages(imageList); err != nil {
		logger.Error.Print("cannot add images", err)
		return err
	}
	end := time.Now()

	imageCount := len(imageList)
	duration := end.Sub(start)
	avg := duration / time.Duration(imageCount)
	logger.Debug.Printf("Added %d images in %s (avg. %s/image)", imageCount, duration, avg)
	return nil
}

func (s *internalManager) GetHandleById(handleId apitype.HandleId) *apitype.Handle {
	return s.imageStore.GetImageById(handleId)
}

// Private API

func (s *internalManager) getCurrentImage() (*apitype.ImageContainer, int, error) {
	images, _ := s.imageStore.GetImagesInCategory(-1, 0, s.selectedCategoryId)
	if s.index < len(images) {
		handle := images[s.index]
		if full, err := s.imageCache.GetFull(handle); err != nil {
			logger.Error.Print("Error while loading full image", err)
			return apitype.NewImageContainer(apitype.GetEmptyHandle(), nil), 0, err
		} else {
			return apitype.NewImageContainer(handle, full), s.index, nil
		}
	} else {
		return apitype.NewImageContainer(apitype.GetEmptyHandle(), nil), 0, nil
	}
}

func (s *internalManager) getTotalImages() int {
	return s.imageStore.GetImageCount(s.selectedCategoryId)
}
func (s *internalManager) getSelectedCategoryId() apitype.CategoryId {
	return s.selectedCategoryId
}
func (s *internalManager) shouldSendSimilarImages() bool {
	return s.shouldSendSimilar
}
func (s *internalManager) getImageListSize() int {
	return s.imageListSize
}

func (s *internalManager) getNextImages() ([]*apitype.ImageContainer, error) {
	if images, err := s.imageStore.GetNextImagesInCategory(s.imageListSize, s.index, s.selectedCategoryId); err != nil {
		return emptyHandles, err
	} else {
		return s.toImageContainers(images)
	}
}

func (s *internalManager) getPrevImages() ([]*apitype.ImageContainer, error) {
	if slice, err := s.imageStore.GetPreviousImagesInCategory(s.imageListSize, s.index, s.selectedCategoryId); err != nil {
		return emptyHandles, err
	} else if images, err := s.toImageContainers(slice); err != nil {
		return emptyHandles, err
	} else {
		return images, nil
	}
}

func (s *internalManager) toImageContainers(nextImageHandles []*apitype.Handle) ([]*apitype.ImageContainer, error) {
	images := make([]*apitype.ImageContainer, len(nextImageHandles))
	for i, handle := range nextImageHandles {
		if thumbnail, err := s.imageCache.GetThumbnail(handle); err != nil {
			logger.Error.Print("Error while loading thumbnail", err)
			return emptyHandles, err
		} else {
			images[i] = apitype.NewImageContainer(handle, thumbnail)
		}
	}

	return images, nil
}

func (s *internalManager) updateImages() error {
	handles := apitype.LoadImageHandles(s.rootDir)
	if err := s.AddHandles(handles); err != nil {
		return err
	} else if err := s.removeMissingImages(handles); err != nil {
		return err
	} else {
		return nil
	}
}

func (s *internalManager) getThreadCount() int {
	cpuCores := runtime.NumCPU()
	return cpuCores
}

func (s *internalManager) getSimilarImages(handle *apitype.Handle) ([]*apitype.ImageContainer, bool, error) {
	similarImages := s.similarityIndex.GetSimilarImages(handle.GetId())
	if len(similarImages) > 0 {
		containers := make([]*apitype.ImageContainer, len(similarImages))
		i := 0
		for _, similar := range similarImages {
			if thumbnail, err := s.imageCache.GetThumbnail(similar); err != nil {
				logger.Error.Print("Error while loading thumbnail", err)
				return emptyHandles, false, err
			} else {
				containers[i] = apitype.NewImageContainer(similar, thumbnail)
			}
			i++
		}

		return containers, true, nil
	} else {
		return []*apitype.ImageContainer{}, false, nil
	}
}

func (s *internalManager) removeMissingImages(handles []*apitype.Handle) error {
	if images, err := s.imageStore.GetAllImages(); err != nil {
		logger.Error.Print("Error while loading images", err)
		return err
	} else {
		var existing = map[string]int{}
		var toRemove = map[apitype.HandleId]*apitype.Handle{}

		for _, handle := range handles {
			existing[handle.GetFile()] = 1
		}

		for _, image := range images {
			if _, ok := existing[image.GetFile()]; !ok {
				toRemove[image.GetId()] = image
			}
		}
		logger.Debug.Printf("Found %d images that don't exist anymore", len(toRemove))

		for handleId, image := range toRemove {
			logger.Trace.Printf("Removing image %s because it doesn't exist", image.String())
			if err := s.imageStore.RemoveImage(handleId); err != nil {
				logger.Error.Print("Can't remove", err)
				return err
			}
		}
		return nil
	}
}
