package filter

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"unsafe"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/util"
	"vincit.fi/image-sorter/common/logger"
)

type fileOperation struct {
	dstPath string
	dstFile string
}

type ImageCopy struct {
	fileOperation
	quality int

	apitype.ImageOperation
}

func NewImageCopy(targetDir string, targetFile string, quality int) apitype.ImageOperation {
	return &ImageCopy{
		quality: quality,
		fileOperation: fileOperation{
			dstPath: targetDir,
			dstFile: targetFile,
		},
	}
}
func (s *ImageCopy) Apply(operationGroup *apitype.ImageOperationGroup) (image.Image, *apitype.ExifData, error) {
	imageFile := operationGroup.ImageFile()
	logger.Debug.Printf("Copy %s", imageFile.Path())

	if operationGroup.Modified() {
		logger.Debug.Printf("Image %s has been modifier. Re-encoding the image...", imageFile.Path())
		imageData := operationGroup.ImageData()
		exifData := operationGroup.ExifData()

		encodingOptions := &jpeg.Options{
			Quality: s.quality,
		}

		jpegBuffer := bytes.NewBuffer([]byte{})
		dstFilePath := filepath.Join(s.dstPath, s.dstFile)
		if err := jpeg.Encode(jpegBuffer, imageData, encodingOptions); err != nil {
			logger.Error.Println("Could not encode image", err)
			return imageData, exifData, err
		} else if err := util.MakeDirectoriesIfNotExist(imageFile.Directory(), s.dstPath); err != nil {
			return imageData, exifData, err
		} else if destination, err := os.Create(dstFilePath); err != nil {
			logger.Error.Println("Could not open file for writing", err)
			return imageData, exifData, err
		} else {
			defer destination.Close()
			s.writeJpegWithExifData(destination, jpegBuffer, exifData)
			return imageData, exifData, nil
		}
	} else {
		logger.Debug.Printf("Copy '%s' as is", imageFile.Path())
		return nil, nil, util.CopyFile(imageFile.Directory(), imageFile.FileName(), s.dstPath, s.dstFile)
	}
}

func (s *ImageCopy) writeJpegWithExifData(destination *os.File, buffer *bytes.Buffer, exifData *apitype.ExifData) {
	writer := bufio.NewWriter(destination)
	// 0xFF 0xD8: Start of JPEG
	writer.Write(buffer.Next(2))

	s.writeJfifBlock(writer, buffer)
	s.writeExifBlock(exifData, writer)

	// Write rest of file
	writer.Write(buffer.Bytes())
	writer.Flush()
}

func (s *ImageCopy) writeExifBlock(data *apitype.ExifData, writer *bufio.Writer) {
	const lengthBytes = 2
	const exifHeader = "Exif\x00\x00"
	headerLength := len(exifHeader)
	// Length includes the length bytes, so we need to add them when writing
	dataLength := data.RawExifDataLength() + uint16(headerLength) + lengthBytes
	dataLengthBytes := (*[2]byte)(unsafe.Pointer(&dataLength))[:]
	writer.Write([]byte{0xFF, 0xE1})
	writer.WriteByte(dataLengthBytes[1])
	writer.WriteByte(dataLengthBytes[0])
	writer.WriteString(exifHeader)
	writer.Write(data.RawExifData())
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
	return fmt.Sprintf("Copy file '%s' to '%s'", s.dstFile, s.dstPath)
}
