package database

import (
	"os"
	"time"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
)

func idToCategoryId(id interface{}) apitype.CategoryId {
	return apitype.CategoryId(id.(int64))
}

func toApiHandle(image *Image) *apitype.Handle {
	handle := apitype.NewHandleWithId(
		image.Id, image.Directory, image.FileName,
	)
	handle.SetByteSize(image.ByteSize)
	return handle
}

func toApiHandles(images []Image) []*apitype.Handle {
	handles := make([]*apitype.Handle, len(images))
	for i, image := range images {
		handles[i] = toApiHandle(&image)
	}
	return handles
}

func toApiCategorizedImages(categories []CategorizedImage) []*apitype.CategorizedImage {
	cats := make([]*apitype.CategorizedImage, len(categories))
	for i, category := range categories {
		cats[i] = toApiCategorizedImage(&category)
	}
	return cats
}

func toApiCategorizedImage(category *CategorizedImage) *apitype.CategorizedImage {
	return apitype.NewCategorizedImage(
		apitype.NewCategoryWithId(
			category.CategoryId, category.Name, category.SubPath, category.Shortcut),
		apitype.OperationFromId(category.Operation),
	)
}

func toApiCategories(categories []Category) []*apitype.Category {
	cats := make([]*apitype.Category, len(categories))
	for i, category := range categories {
		cats[i] = toApiCategory(category)
	}
	return cats
}

func toApiCategory(category Category) *apitype.Category {
	return apitype.NewCategoryWithId(category.Id, category.Name, category.SubPath, category.Shortcut)
}

type ImageHandleConverter interface {
	HandleToImage(handle *apitype.Handle) (*Image, error)
	GetHandleFileStats(handle *apitype.Handle) (os.FileInfo, error)
}

type FileSystemImageHandleConverter struct {
	ImageHandleConverter
}

func (s *FileSystemImageHandleConverter) GetHandleFileStats(handle *apitype.Handle) (os.FileInfo, error) {
	return os.Stat(handle.GetPath())
}

func (s *FileSystemImageHandleConverter) HandleToImage(handle *apitype.Handle) (*Image, error) {

	exifLoadStart := time.Now()
	exifData, err := apitype.LoadExifData(handle)
	if err != nil {
		logger.Warn.Printf("Exif data not properly loaded for '%d'", handle.GetId())
		return nil, err
	}
	exifLoadEnd := time.Now()
	logger.Trace.Printf(" - Loaded exif data in %s", exifLoadEnd.Sub(exifLoadStart))

	fileStatStart := time.Now()
	fileStat, err := s.GetHandleFileStats(handle)
	if err != nil {
		return nil, err
	}
	fileStatEnd := time.Now()
	logger.Trace.Printf(" - Loaded file info in %s", fileStatEnd.Sub(fileStatStart))

	return &Image{
		Name:            handle.GetFile(),
		FileName:        handle.GetFile(),
		Directory:       handle.GetDir(),
		ByteSize:        fileStat.Size(),
		ExifOrientation: exifData.GetExifOrientation(),
		ImageAngle:      int(exifData.GetRotation()),
		CreatedTime:     exifData.GetCreatedTime(),
		Width:           exifData.GetWidth(),
		Height:          exifData.GetHeight(),
		ModifiedTime:    fileStat.ModTime(),
	}, nil
}
