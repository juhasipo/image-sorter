package filter

import (
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"vincit.fi/image-sorter/common"
)

const defaultQuality = 100

type fileOperation struct {
	dstPath string
	dstFile string
}

type ImageOperation interface {
	Apply(handle *common.Handle, img image.Image, data *common.ExifData) (image.Image, *common.ExifData, error)
	String() string
}

type ImageOperationGroup struct {
	handle     *common.Handle
	exifData   *common.ExifData
	img        image.Image
	operations []ImageOperation
}

func ImageOperationGroupNew(handle *common.Handle, img image.Image, exifData *common.ExifData, operations []ImageOperation) *ImageOperationGroup {
	return &ImageOperationGroup{
		handle:     handle,
		img:        img,
		exifData:   exifData,
		operations: operations,
	}
}

func (s *ImageOperationGroup) GetOperations() []ImageOperation {
	return s.operations
}

func (s *ImageOperationGroup) Apply() error {
	for _, operation := range s.operations {
		log.Printf("Applying: '%s'", operation)
		var err error
		if s.img, s.exifData, err = operation.Apply(s.handle, s.img, s.exifData); err != nil {
			return err
		}
	}
	return nil
}

// Copy

type ImageCopy struct {
	fileOperation
	quick bool

	ImageOperation
}

func ImageCopyNew(targetDir string, targetFile string, quick bool) ImageOperation {
	return &ImageCopy{
		quick: quick,
		fileOperation: fileOperation{
			dstPath: targetDir,
			dstFile: targetFile,
		},
	}
}
func (s *ImageCopy) Apply(handle *common.Handle, img image.Image, data *common.ExifData) (image.Image, *common.ExifData, error) {
	log.Printf("Copy %s", handle.GetPath())

	if s.quick {
		log.Printf("Copy '%s' as is", handle.GetPath())
		return img, data, common.CopyFile(handle.GetDir(), handle.GetFile(), s.dstPath, s.dstFile)
	} else {
		encodingOptions := &jpeg.Options{
			Quality: defaultQuality,
		}

		dstFilePath := filepath.Join(s.dstPath, s.dstFile)
		if destination, err := os.Create(dstFilePath); err != nil {
			log.Println("Could not open file for writing", err)
			return img, data, err
		} else if err := jpeg.Encode(destination, img, encodingOptions); err != nil {
			log.Println("Could not encode image", err)
			return img, data, err
			// TODO: Write Exif data
		} else {
			return img, data, nil
		}
	}
}
func (s *ImageCopy) String() string {
	return fmt.Sprintf("Copy file '%s' to '%s'", s.dstFile, s.dstPath)
}

// Move

type ImageMove struct {
	fileOperation

	ImageOperation
}

func (s *ImageMove) Apply(handle *common.Handle, img image.Image, data *common.ExifData) (image.Image, *common.ExifData, error) {
	log.Printf("Move %s", handle.GetPath())
	return img, data, nil
}
func (s *ImageMove) String() string {
	return "Move to " + s.dstPath + " " + s.dstFile
}

// Remove

type ImageRemove struct {
	ImageOperation
}

func ImageRemoveNew() ImageOperation {
	return &ImageRemove{}
}
func (s *ImageRemove) Apply(handle *common.Handle, img image.Image, data *common.ExifData) (image.Image, *common.ExifData, error) {
	log.Printf("Remove %s", handle.GetPath())
	return img, data, common.RemoveFile(handle.GetPath())
}
func (s *ImageRemove) String() string {
	return "Remove"
}
