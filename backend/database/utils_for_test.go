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

func (s *StubImageHandleConverter) HandleToImage(handle *apitype.Handle) (*Image, map[string]string, error) {
	fileStat, _ := s.GetHandleFileStats(handle)
	if jsonData, err := json.Marshal(handle.GetMetaData()); err != nil {
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
		}, map[string]string{}, nil
	}
}

func (s *StubImageHandleConverter) GetHandleFileStats(handle *apitype.Handle) (os.FileInfo, error) {
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
