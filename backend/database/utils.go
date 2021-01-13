package database

import (
	"encoding/json"
	"os"
	"time"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
	"vincit.fi/image-sorter/common/util"
)

func idToCategoryId(id interface{}) apitype.CategoryId {
	return apitype.CategoryId(id.(int64))
}

func toApiHandle(image *Image) (*apitype.Handle, error) {
	var metaData = map[string]string{}
	if err := json.Unmarshal(image.ExifData, &metaData); err != nil {
		return nil, err
	}
	handle := apitype.NewHandleWithId(
		image.Id, image.Directory, image.FileName, metaData,
	)
	handle.SetByteSize(image.ByteSize)
	return handle, nil
}

func toApiHandles(images []Image) []*apitype.Handle {
	handles := make([]*apitype.Handle, len(images))
	for i, image := range images {
		handles[i], _ = toApiHandle(&image)
	}
	return handles
}

func toApiCategorizedImages(categories []CategorizedImage) []*apitype.CategorizedImage {
	apiTypeCategories := make([]*apitype.CategorizedImage, len(categories))
	for i, category := range categories {
		apiTypeCategories[i] = toApiCategorizedImage(&category)
	}
	return apiTypeCategories
}

func toApiCategorizedImage(category *CategorizedImage) *apitype.CategorizedImage {
	return apitype.NewCategorizedImage(
		apitype.NewCategoryWithId(
			category.CategoryId, category.Name, category.SubPath, category.Shortcut),
		apitype.OperationFromId(category.Operation),
	)
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

type ImageHandleConverter interface {
	HandleToImage(handle *apitype.Handle) (*Image, map[string]string, error)
	GetHandleFileStats(handle *apitype.Handle) (os.FileInfo, error)
}

type FileSystemImageHandleConverter struct {
	ImageHandleConverter
}

func (s *FileSystemImageHandleConverter) GetHandleFileStats(handle *apitype.Handle) (os.FileInfo, error) {
	return os.Stat(handle.GetPath())
}

func (s *FileSystemImageHandleConverter) HandleToImage(handle *apitype.Handle) (*Image, map[string]string, error) {

	exifLoadStart := time.Now()
	exifData, err := apitype.LoadExifData(handle)
	if err != nil {
		logger.Warn.Printf("Exif data not properly loaded for '%d'", handle.GetId())
		return nil, nil, err
	}
	exifLoadEnd := time.Now()
	logger.Trace.Printf(" - Loaded exif data in %s", exifLoadEnd.Sub(exifLoadStart))

	fileStatStart := time.Now()
	fileStat, err := s.GetHandleFileStats(handle)
	if err != nil {
		return nil, nil, err
	}
	fileStatEnd := time.Now()
	logger.Trace.Printf(" - Loaded file info in %s", fileStatEnd.Sub(fileStatStart))

	w := util.NewMapExifWalker()
	exifData.Walk(w)

	metaDataJson, err := json.Marshal(w.GetMetaData())
	if err != nil {
		return nil, nil, err
	}

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
		ExifData:        metaDataJson,
	}, w.GetMetaData(), nil
}
