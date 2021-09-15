package database

import (
	"os"
	"time"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
	"vincit.fi/image-sorter/common/util"
)

func idToCategoryId(id interface{}) apitype.CategoryId {
	return apitype.CategoryId(id.(int64))
}

func toImageFile(image *Image) (*apitype.ImageFile, error) {
	return apitype.NewImageFileWithIdSizeAndOrientation(
		image.Id, image.Directory, image.FileName, image.ByteSize, float64(image.ImageAngle), image.ImageFlip, int(image.Width), int(image.Height),
	), nil
}

func toImageFiles(images []Image) []*apitype.ImageFile {
	imageFiles := make([]*apitype.ImageFile, len(images))
	for i, image := range images {
		imageFiles[i], _ = toImageFile(&image)
	}
	return imageFiles
}

func toApiCategorizedImages(categories []CategorizedImage) []*api.CategorizedImage {
	apiTypeCategories := make([]*api.CategorizedImage, len(categories))
	for i, category := range categories {
		apiTypeCategories[i] = toApiCategorizedImage(&category)
	}
	return apiTypeCategories
}

func toApiCategorizedImage(category *CategorizedImage) *api.CategorizedImage {
	return &api.CategorizedImage{
		Category: apitype.NewCategoryWithId(
			category.CategoryId, category.Name, category.SubPath, category.Shortcut),
		Operation: apitype.OperationFromId(category.Operation),
	}
}

func toApiCategories(categories []Category) []*apitype.Category {
	apiTypeCategories := make([]*apitype.Category, len(categories))
	for i, category := range categories {
		apiTypeCategories[i] = toApiCategory(category)
	}
	return apiTypeCategories
}

func toApiCategory(category Category) *apitype.Category {
	return apitype.NewCategoryWithId(category.Id, category.Name, category.SubPath, category.Shortcut)
}

type ImageFileConverter interface {
	ImageFileToDbImage(*apitype.ImageFile) (*Image, map[string]string, error)
	GetImageFileStats(*apitype.ImageFile) (os.FileInfo, error)
}

type FileSystemImageFileConverter struct {
	ImageFileConverter
}

func (s *FileSystemImageFileConverter) GetImageFileStats(imageFile *apitype.ImageFile) (os.FileInfo, error) {
	return os.Stat(imageFile.Path())
}

func (s *FileSystemImageFileConverter) ImageFileToDbImage(imageFile *apitype.ImageFile) (*Image, map[string]string, error) {
	exifLoadStart := time.Now()
	exifData, err := util.LoadExifData(imageFile)
	if err != nil {
		logger.Warn.Printf("Exif data not properly loaded for '%d'", imageFile.Id())
		return nil, nil, err
	}
	exifLoadEnd := time.Now()
	logger.Trace.Printf(" - Loaded exif data in %s", exifLoadEnd.Sub(exifLoadStart))

	fileStatStart := time.Now()
	fileStat, err := s.GetImageFileStats(imageFile)
	if err != nil {
		return nil, nil, err
	}
	fileStatEnd := time.Now()
	logger.Trace.Printf(" - Loaded file info in %s", fileStatEnd.Sub(fileStatStart))

	return &Image{
		Name:            imageFile.FileName(),
		FileName:        imageFile.FileName(),
		Directory:       imageFile.Directory(),
		ByteSize:        fileStat.Size(),
		ExifOrientation: exifData.ExifOrientation(),
		ImageAngle:      int(exifData.Rotation()),
		CreatedTime:     exifData.CreatedTime(),
		Width:           exifData.ImageWidth(),
		Height:          exifData.ImageHeight(),
		ModifiedTime:    fileStat.ModTime(),
	}, exifData.Values(), nil
}
