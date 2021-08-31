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
	emptyImageFiles = []*apitype.ImageFileAndData{}
)

type ImageList func(number int) []*apitype.ImageFile

type ImageLibrary struct {
	categoryService    api.CategoryService
	imageCache         api.ImageStore
	imageLoader        api.ImageLoader
	similarityIndex    *database.SimilarityIndex
	imageStore         *database.ImageStore
	imageMetaDataStore *database.ImageMetaDataStore
	hashCalculator     *HashCalculator
	progressReporter   api.ProgressReporter

	api.ImageLibrary
}

func NewImageLibrary(imageCache api.ImageStore, imageLoader api.ImageLoader,
	similarityIndex *database.SimilarityIndex, imageStore *database.ImageStore,
	imageMetaDataStore *database.ImageMetaDataStore,
	progressReporter api.ProgressReporter) *ImageLibrary {
	var service = ImageLibrary{
		imageCache:         imageCache,
		imageLoader:        imageLoader,
		similarityIndex:    similarityIndex,
		imageStore:         imageStore,
		imageMetaDataStore: imageMetaDataStore,
		progressReporter:   progressReporter,
	}
	return &service
}

func (s *ImageLibrary) InitializeFromDirectory(directory string) (time.Time, error) {
	return s.updateImages(directory)
}

func (s *ImageLibrary) GetImages() []*apitype.ImageFile {
	images, _ := s.imageStore.GetAllImages()
	return images
}

func (s *ImageLibrary) GenerateHashes() bool {
	if s.hashCalculator != nil {
		logger.Warn.Print("Already generating hashes")
		return false
	}

	s.hashCalculator = NewHashCalculator(s.similarityIndex, s.imageLoader, s.getThreadCount())

	shouldSendSimilarImages := false
	images, _ := s.imageStore.GetAllImages()
	hashes, err := s.hashCalculator.GenerateHashes(images, func(current int, total int) {
		s.progressReporter.Update("Calculating Hashes...", current, total, true)
	})

	if err == nil {
		err = s.hashCalculator.BuildSimilarityIndex(hashes, func(current int, total int) {
			s.progressReporter.Update("Building Similarity Index...", current, total, false)
		})
	}

	if err != nil {
		s.progressReporter.Error("Error while saving hashes", err)
	}

	// Always send 100% status even if cancelled so that the progress bar is hidden
	s.progressReporter.Update("Done", 0, 0, true)

	// Only send if not cancelled or no error
	if err == nil {
		shouldSendSimilarImages = true
	}

	s.hashCalculator = nil
	return shouldSendSimilarImages
}

func (s *ImageLibrary) StopHashes() {
	if s.hashCalculator != nil {
		s.hashCalculator.StopHashes()
	}
}

func (s *ImageLibrary) GetImagesInCategory(number int, offset int, categoryId apitype.CategoryId) ([]*apitype.ImageFile, error) {
	return s.imageStore.GetImagesInCategory(number, offset, categoryId)
}

func (s *ImageLibrary) AddImageFiles(imageList []*apitype.ImageFile) error {
	latestModifiedImageTimestamp := s.imageStore.GetLatestModifiedImage()
	s.progressReporter.Update("Loading images...", 0, 2, false)
	if err := s.addImagesToDb(imageList); err != nil {
		return err
	}

	s.progressReporter.Update("Loading Meta Data...", 1, 2, false)

	if modifiedImages, err := s.imageStore.GetAllImagesModifiedAfter(latestModifiedImageTimestamp); err != nil {
		logger.Error.Print("cannot read images", err)
		return err
	} else if imagesWithoutMetaData, err := s.imageMetaDataStore.GetAllImagesWithoutMetaData(); err != nil {
		logger.Error.Print("cannot read images", err)
		return err
	} else {
		logger.Debug.Printf("There are %d images modified", len(modifiedImages))
		logger.Debug.Printf("There are %d images without meta data", len(imagesWithoutMetaData))
		filesToAddMetaData := mergeLists(modifiedImages, imagesWithoutMetaData)
		numOfImagesToUpdate := len(filesToAddMetaData)
		logger.Debug.Printf("There are %d images which needs meta data to be fetched", numOfImagesToUpdate)

		if err := s.addImageMetaDataToDb(filesToAddMetaData); err != nil {
			return err
		}
	}

	s.progressReporter.Update("Done", 2, 2, false)

	return nil
}

func mergeLists(modifiedImages []*apitype.ImageFile, imagesWithoutMetaData []*apitype.ImageFile) []*apitype.ImageFile {
	m := map[apitype.ImageId]*apitype.ImageFile{}
	for _, image := range modifiedImages {
		m[image.Id()] = image
	}
	for _, image := range imagesWithoutMetaData {
		m[image.Id()] = image
	}

	var v []*apitype.ImageFile
	for _, file := range m {
		v = append(v, file)
	}
	return v
}

func (s *ImageLibrary) GetImageFileById(imageId apitype.ImageId) *apitype.ImageFile {
	return s.imageStore.GetImageById(imageId)
}

// Private API

func (s *ImageLibrary) GetImageAtIndex(index int, categoryId apitype.CategoryId) (*apitype.ImageFileAndData, *apitype.ImageMetaData, int, error) {
	imageCount := s.imageStore.GetImageCount(categoryId)
	if index >= 0 && index < imageCount {
		images, _ := s.imageStore.GetImagesInCategory(1, index, categoryId)
		imageFile := images[0]
		if full, err := s.imageCache.GetFull(imageFile.Id()); err != nil {
			logger.Error.Print("Error while loading full image", err)
			return apitype.NewEmptyImageContainer(), apitype.NewInvalidImageMetaData(), 0, err
		} else if metaData, err := s.imageMetaDataStore.GetMetaDataByImageId(imageFile.Id()); err != nil {
			return apitype.NewEmptyImageContainer(), apitype.NewInvalidImageMetaData(), 0, nil
		} else {
			return apitype.NewImageContainer(imageFile, full), metaData, index, nil
		}
	}
	return apitype.NewEmptyImageContainer(), apitype.NewInvalidImageMetaData(), 0, nil
}

func (s *ImageLibrary) GetTotalImages(categoryId apitype.CategoryId) int {
	return s.imageStore.GetImageCount(categoryId)
}

func (s *ImageLibrary) GetNextImages(index int, count int, categoryId apitype.CategoryId) ([]*apitype.ImageFileAndData, error) {
	if images, err := s.imageStore.GetNextImagesInCategory(count, index, categoryId); err != nil {
		return emptyImageFiles, err
	} else {
		return s.toImageContainers(images)
	}
}

func (s *ImageLibrary) GetPreviousImages(index int, count int, categoryId apitype.CategoryId) ([]*apitype.ImageFileAndData, error) {
	if slice, err := s.imageStore.GetPreviousImagesInCategory(count, index, categoryId); err != nil {
		return emptyImageFiles, err
	} else if images, err := s.toImageContainers(slice); err != nil {
		return emptyImageFiles, err
	} else {
		return images, nil
	}
}

func (s *ImageLibrary) toImageContainers(nextImageFiles []*apitype.ImageFile) ([]*apitype.ImageFileAndData, error) {
	images := make([]*apitype.ImageFileAndData, len(nextImageFiles))
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

func (s *ImageLibrary) updateImages(rootDir string) (time.Time, error) {
	imageFiles := apitype.LoadImageFiles(rootDir)
	if err := s.AddImageFiles(imageFiles); err != nil {
		return time.Unix(0, 0), err
	} else if err := s.removeMissingImages(imageFiles); err != nil {
		return time.Unix(0, 0), err
	} else {
		return s.imageStore.GetLatestModifiedImage(), nil
	}
}

func (s *ImageLibrary) getThreadCount() int {
	cpuCores := runtime.NumCPU()
	return cpuCores
}

func (s *ImageLibrary) GetSimilarImages(imageId apitype.ImageId) ([]*apitype.ImageFileAndData, bool, error) {
	similarImages := s.similarityIndex.GetSimilarImages(imageId)
	if len(similarImages) > 0 {
		containers := make([]*apitype.ImageFileAndData, len(similarImages))
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
		return []*apitype.ImageFileAndData{}, false, nil
	}
}

func (s *ImageLibrary) removeMissingImages(imageFiles []*apitype.ImageFile) error {
	if images, err := s.imageStore.GetAllImages(); err != nil {
		logger.Error.Print("Error while loading images", err)
		return err
	} else {
		var existing = map[string]bool{}
		var toRemove = map[apitype.ImageId]*apitype.ImageFile{}

		for _, imageFile := range imageFiles {
			existing[imageFile.FileName()] = true
		}

		for _, image := range images {
			if _, ok := existing[image.FileName()]; !ok {
				toRemove[image.Id()] = image
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
			return nil
		} else {
			logger.Trace.Print("No missing images to remove")
			return nil
		}
	}
}

func (s *ImageLibrary) addImagesToDb(imageList []*apitype.ImageFile) error {
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

func (s *ImageLibrary) addImageMetaDataToDb(images []*apitype.ImageFile) error {
	start := time.Now()
	if err := s.imageMetaDataStore.AddMetaDataForImages(images, s.imageLoader.LoadExifData); err != nil {
		return err
	}

	end := time.Now()

	imageCount := len(images)
	duration := end.Sub(start)
	avg := time.Duration(0)
	if imageCount > 0 {
		avg = duration / time.Duration(imageCount)
	}
	logger.Debug.Printf("Added meta data for %d images in %s (avg. %s/image)", imageCount, duration, avg)
	return nil
}
