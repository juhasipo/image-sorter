package filter

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"unsafe"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/imageloader"
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
	reEncode bool

	ImageOperation
}

func ImageCopyNew(targetDir string, targetFile string, reEncode bool) ImageOperation {
	return &ImageCopy{
		reEncode: reEncode,
		fileOperation: fileOperation{
			dstPath: targetDir,
			dstFile: targetFile,
		},
	}
}
func (s *ImageCopy) Apply(handle *common.Handle, img image.Image, exifData *common.ExifData) (image.Image, *common.ExifData, error) {
	log.Printf("Copy %s", handle.GetPath())

	if !s.reEncode {
		log.Printf("Copy '%s' as is", handle.GetPath())
		return img, exifData, common.CopyFile(handle.GetDir(), handle.GetFile(), s.dstPath, s.dstFile)
	} else {
		encodingOptions := &jpeg.Options{
			Quality: defaultQuality,
		}

		jpegBuffer := bytes.NewBuffer([]byte{})
		dstFilePath := filepath.Join(s.dstPath, s.dstFile)
		if err := jpeg.Encode(jpegBuffer, img, encodingOptions); err != nil {
			log.Println("Could not encode image", err)
			return img, exifData, err
		} else if err := common.MakeDirectoriesIfNotExist(handle.GetDir(), s.dstPath); err != nil {
			return img, exifData, err
		} else if destination, err := os.Create(dstFilePath); err != nil {
			log.Println("Could not open file for writing", err)
			return img, exifData, err
		} else {
			defer destination.Close()
			s.writeJpegWithExifData(destination, jpegBuffer, exifData)
			return img, exifData, nil
		}
	}
}

func (s *ImageCopy) writeJpegWithExifData(destination *os.File, buffer *bytes.Buffer, exifData *common.ExifData) {
	writer := bufio.NewWriter(destination)
	// 0xFF 0xD8: Start of JPEG
	writer.Write(buffer.Next(2))

	s.writeJfifBlock(writer, buffer)
	s.writeExifBlock(exifData, writer)

	// Write rest of file
	writer.Write(buffer.Bytes())
	writer.Flush()
}

func (s *ImageCopy) writeExifBlock(data *common.ExifData, writer *bufio.Writer) {
	const lengthBytes = 2
	const exifHeader = "Exif\x00\x00"
	headerLength := len(exifHeader)
	// Length includes the length bytes, so we need to add them when writing
	dataLength := data.GetRawLength() + uint16(headerLength) + lengthBytes
	dataLengthBytes := (*[2]byte)(unsafe.Pointer(&dataLength))[:]
	writer.Write([]byte{0xFF, 0xE1})
	writer.WriteByte(dataLengthBytes[1])
	writer.WriteByte(dataLengthBytes[0])
	writer.WriteString(exifHeader)
	writer.Write(data.GetRaw())
}

func (s *ImageCopy) writeJfifBlock(writer *bufio.Writer, bw *bytes.Buffer) {
	// 0xFF 0xE0 length (2 bytes): APP0 block of 0 length
	const lengthBytes = 2
	writer.Write(bw.Next(2))
	e0LengthBytes := bw.Next(2)
	// Length includes the length bytes, so we need to subtract when reading
	e0Length := int(binary.BigEndian.Uint16(e0LengthBytes)) - lengthBytes
	writer.Write(e0LengthBytes)
	writer.Write(bw.Next(e0Length))
}
func (s *ImageCopy) String() string {
	return fmt.Sprintf("Copy file '%s' to '%s', re-encode: %t", s.dstFile, s.dstPath, s.reEncode)
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

// Exif Rotate

type ImageExifRotate struct {
	ImageOperation
}

func ImageExifRotateNew() ImageOperation {
	return &ImageExifRotate{}
}
func (s *ImageExifRotate) Apply(handle *common.Handle, img image.Image, data *common.ExifData) (image.Image, *common.ExifData, error) {
	log.Printf("Exif rotate %s", handle.GetPath())
	rotatedImage, err := imageloader.ExifRotateImage(img, data)
	if err != nil {
		return img, data, err
	}
	data.ResetExifRotate()
	return rotatedImage, data, err
}
func (s *ImageExifRotate) String() string {
	return "Exif Rotate"
}

// Rotate to angle

type ImageRotateToAngle struct {
	rotation float64
	ImageOperation
}

func ImageRotateToAngleNew(angle int) ImageOperation {
	return &ImageRotateToAngle{
		rotation: float64(angle),
	}
}
func (s *ImageRotateToAngle) Apply(handle *common.Handle, img image.Image, data *common.ExifData) (image.Image, *common.ExifData, error) {
	log.Printf("Exif rotate %s", handle.GetPath())
	rotatedImage := imaging.Rotate(img, s.rotation, image.Black)
	data.ResetExifRotate()
	return rotatedImage, data, nil
}
func (s *ImageRotateToAngle) String() string {
	return fmt.Sprintf("Rotate to %.2f", s.rotation)
}
