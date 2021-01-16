package database

import (
	"encoding/json"
	"os"
	"time"
	"vincit.fi/image-sorter/api/apitype"
)

type StubImageFileConverter struct {
	ImageFileConverter

	incrementModTimeRequest bool
	currentTime             time.Time
}

func (s *StubImageFileConverter) ImageFileToDbImage(imageFile *apitype.ImageFile) (*Image, map[string]string, error) {
	fileStat, _ := s.GetImageFileStats(imageFile)
	metaData := map[string]string{}
	if jsonData, err := json.Marshal(metaData); err != nil {
		return nil, nil, err
	} else {

		return &Image{
			Id:              0,
			Name:            imageFile.GetFile(),
			FileName:        imageFile.GetFile(),
			Directory:       imageFile.GetDir(),
			ByteSize:        1234,
			ExifOrientation: 1,
			ImageAngle:      90,
			ImageFlip:       true,
			CreatedTime:     fileStat.ModTime(),
			Width:           1024,
			Height:          2048,
			ModifiedTime:    fileStat.ModTime(),

			ExifData: jsonData,
		}, metaData, nil
	}
}

func (s *StubImageFileConverter) GetImageFileStats(imageFile *apitype.ImageFile) (os.FileInfo, error) {
	if s.incrementModTimeRequest {
		s.currentTime = s.currentTime.Add(time.Second)
	}

	return &StubFileInfo{
		modTime: s.currentTime,
	}, nil
}

func (s *StubImageFileConverter) SetIncrementModTimeRequest(value bool) {
	s.incrementModTimeRequest = value
}

type StubFileInfo struct {
	os.FileInfo

	modTime time.Time
}

func (s *StubFileInfo) ModTime() time.Time {
	return s.modTime
}
