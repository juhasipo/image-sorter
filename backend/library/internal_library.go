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
	emptyImageFiles = []*apitype.ImageContainer{}
)

const maxSimilarImages = 20

type ImageList func(number int) []*apitype.ImageFile

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

func (s *internalManager) GetImages() []*apitype.ImageFileWithMetaData {
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
			sender.SendCommandToTopic(api.ProcessStatusUpdated, &api.UpdateProgressCommand{
				Name:    "hash",
				Current: current,
				Total:   total,
			})
		})

		if err == nil {
			err = s.hashCalculator.BuildSimilarityIndex(hashes, func(current int, total int) {
				sender.SendCommandToTopic(api.ProcessStatusUpdated, &api.UpdateProgressCommand{
					Name:    "similarity-index",
					Current: current,
					Total:   total,
				})
			})
		}

		if err != nil {
			sender.SendError("Error while saving hashes", err)
		}

		// Always send 100% status even if cancelled so that the progress bar is hidden
		sender.SendCommandToTopic(api.ProcessStatusUpdated, &api.UpdateProgressCommand{
			Name:    "hash",
			Current: 0,
			Total:   0,
		})

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

func (s *internalManager) MoveToImage(imageId apitype.ImageId) {
	images, _ := s.imageStore.GetImagesInCategory(-1, 0, s.selectedCategoryId)
	for imageIndex, image := range images {
		if imageId == image.Id() {
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

func (s *internalManager) AddImageFiles(imageList []*apitype.ImageFile) error {
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

func (s *internalManager) GetImageFileById(imageId apitype.ImageId) *apitype.ImageFileWithMetaData {
	return s.imageStore.GetImageById(imageId)
}

// Private API

func (s *internalManager) getCurrentImage() (*apitype.ImageContainer, int, error) {
	images, _ := s.imageStore.GetImagesInCategory(-1, 0, s.selectedCategoryId)
	if s.index < len(images) {
		imageFile := images[s.index]
		if full, err := s.imageCache.GetFull(imageFile.Id()); err != nil {
			logger.Error.Print("Error while loading full image", err)
			return apitype.NewEmptyImageContainer(), 0, err
		} else {
			return apitype.NewImageContainer(imageFile, full), s.index, nil
		}
	} else {
		return apitype.NewEmptyImageContainer(), 0, nil
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
		return emptyImageFiles, err
	} else {
		return s.toImageContainers(images)
	}
}

func (s *internalManager) getPrevImages() ([]*apitype.ImageContainer, error) {
	if slice, err := s.imageStore.GetPreviousImagesInCategory(s.imageListSize, s.index, s.selectedCategoryId); err != nil {
		return emptyImageFiles, err
	} else if images, err := s.toImageContainers(slice); err != nil {
		return emptyImageFiles, err
	} else {
		return images, nil
	}
}

func (s *internalManager) toImageContainers(nextImageFiles []*apitype.ImageFileWithMetaData) ([]*apitype.ImageContainer, error) {
	images := make([]*apitype.ImageContainer, len(nextImageFiles))
	for i, imageFile := range nextImageFiles {
		if thumbnail, err := s.imageCache.GetThumbnail(imageFile.Id()); err != nil {
			logger.Error.Print("Error while loading thumbnail", err)
			return emptyImageFiles, err
		} else {
			images[i] = apitype.NewImageContainer(imageFile, thumbnail)
		}
	}

	return images, nil
}

func (s *internalManager) updateImages() error {
	imageFiles := apitype.LoadImageFiles(s.rootDir)
	if err := s.AddImageFiles(imageFiles); err != nil {
		return err
	} else if err := s.removeMissingImages(imageFiles); err != nil {
		return err
	} else {
		return nil
	}
}

func (s *internalManager) getThreadCount() int {
	cpuCores := runtime.NumCPU()
	return cpuCores
}

func (s *internalManager) getSimilarImages(imageId apitype.ImageId) ([]*apitype.ImageContainer, bool, error) {
	similarImages := s.similarityIndex.GetSimilarImages(imageId)
	if len(similarImages) > 0 {
		containers := make([]*apitype.ImageContainer, len(similarImages))
		i := 0
		for _, similar := range similarImages {
			if thumbnail, err := s.imageCache.GetThumbnail(similar.Id()); err != nil {
				logger.Error.Print("Error while loading thumbnail", err)
				return emptyImageFiles, false, err
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

func (s *internalManager) removeMissingImages(imageFiles []*apitype.ImageFile) error {
	if images, err := s.imageStore.GetAllImages(); err != nil {
		logger.Error.Print("Error while loading images", err)
		return err
	} else {
		var existing = map[string]int{}
		var toRemove = map[apitype.ImageId]*apitype.ImageFile{}

		for _, imageFile := range imageFiles {
			existing[imageFile.FileName()] = 1
		}

		for _, image := range images {
			if _, ok := existing[image.FileName()]; !ok {
				toRemove[image.Id()] = &image.ImageFile
			}
		}
		if len(toRemove) > 0 {
			logger.Debug.Printf("Found %d images that don't exist anymore", len(toRemove))

			for imageId, image := range toRemove {
				logger.Trace.Printf("Removing image %s because it doesn't exist", image.String())
				if err := s.imageStore.RemoveImage(imageId); err != nil {
					logger.Error.Print("Can't remove", err)
					return err
				}
			}
		} else {
			logger.Trace.Print("No missing images to remove")
		}
		return nil
	}
}
