package database

import (
	"encoding/json"
	"os"
	"time"
	"vincit.fi/image-sorter/api/apitype"
)

type StubImageHandleConverter struct {
	ImageHandleConverter

	incrementModTimeRequest bool
	currentTime             time.Time
}

func (s *StubImageHandleConverter) HandleToImage(handle *apitype.ImageFile) (*Image, map[string]string, error) {
	fileStat, _ := s.GetHandleFileStats(handle)
	metaData := map[string]string{}
	if jsonData, err := json.Marshal(metaData); err != nil {
		return nil, nil, err
	} else {

		return &Image{
			Id:              0,
			Name:            handle.GetFile(),
			FileName:        handle.GetFile(),
			Directory:       handle.GetDir(),
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

func (s *StubImageHandleConverter) GetHandleFileStats(handle *apitype.ImageFile) (os.FileInfo, error) {
	if s.incrementModTimeRequest {
		s.currentTime = s.currentTime.Add(time.Second)
	}

	return &StubFileInfo{
		modTime: s.currentTime,
	}, nil
}

func (s *StubImageHandleConverter) SetIncrementModTimeRequest(value bool) {
	s.incrementModTimeRequest = value
}

type StubFileInfo struct {
	os.FileInfo

	modTime time.Time
}

func (s *StubFileInfo) ModTime() time.Time {
	return s.modTime
}
