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
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/logger"
)

type ImageCopy struct {
	fileOperation
	quality int

	ImageOperation
}

func NewImageCopy(targetDir string, targetFile string, quality int) ImageOperation {
	return &ImageCopy{
		quality: quality,
		fileOperation: fileOperation{
			dstPath: targetDir,
			dstFile: targetFile,
		},
	}
}
func (s *ImageCopy) Apply(operationGroup *ImageOperationGroup) (image.Image, *common.ExifData, error) {
	handle := operationGroup.handle
	img := operationGroup.img
	exifData := operationGroup.exifData
	logger.Debug.Printf("Copy %s", handle.GetPath())

	if operationGroup.hasBeenModified {
		encodingOptions := &jpeg.Options{
			Quality: s.quality,
		}

		jpegBuffer := bytes.NewBuffer([]byte{})
		dstFilePath := filepath.Join(s.dstPath, s.dstFile)
		if err := jpeg.Encode(jpegBuffer, img, encodingOptions); err != nil {
			logger.Error.Println("Could not encode image", err)
			return img, exifData, err
		} else if err := common.MakeDirectoriesIfNotExist(handle.GetDir(), s.dstPath); err != nil {
			return img, exifData, err
		} else if destination, err := os.Create(dstFilePath); err != nil {
			logger.Error.Println("Could not open file for writing", err)
			return img, exifData, err
		} else {
			defer destination.Close()
			s.writeJpegWithExifData(destination, jpegBuffer, exifData)
			return img, exifData, nil
		}
	} else {
		logger.Debug.Printf("Copy '%s' as is", handle.GetPath())
		return img, exifData, common.CopyFile(handle.GetDir(), handle.GetFile(), s.dstPath, s.dstFile)
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
	return fmt.Sprintf("Copy file '%s' to '%s'", s.dstFile, s.dstPath)
}
