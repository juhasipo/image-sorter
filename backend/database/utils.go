package database

import (
	"os"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/imageloader"
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

func getHandleFileStats(handle *apitype.Handle) (os.FileInfo, error) {
	return os.Stat(handle.GetPath())
}

type ImageHandleConverter interface {
	HandleToImage(handle *apitype.Handle) (*Image, error)
}

type FileSystemImageHandleConverter struct {
	ImageHandleConverter
}

func (s *FileSystemImageHandleConverter) HandleToImage(handle *apitype.Handle) (*Image, error) {
	exifData, err := apitype.LoadExifData(handle)
	if err != nil {
		logger.Warn.Printf("Exif data not properly loaded for '%d'", handle.GetId())
		return nil, err
	}

	fileStat, err := getHandleFileStats(handle)
	if err != nil {
		return nil, err
	}

	image, err := imageloader.NewImageLoader().LoadImage(handle)
	if err != nil {
		logger.Error.Println("Could not load image"+handle.GetPath(), err)
		return nil, err
	}

	return &Image{
		Name:            handle.GetFile(),
		FileName:        handle.GetFile(),
		Directory:       handle.GetDir(),
		ByteSize:        fileStat.Size(),
		ExifOrientation: exifData.GetExifOrientation(),
		ImageAngle:      int(exifData.GetRotation()),
		CreatedTime:     exifData.GetCreatedTime(),
		Width:           uint(image.Bounds().Dx()),
		Height:          uint(image.Bounds().Dy()),
		ModifiedTime:    fileStat.ModTime(),
	}, nil
}