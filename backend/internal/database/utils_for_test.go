package database

import (
	"os"
	"time"
	"vincit.fi/image-sorter/api/apitype"
)

type StubImageFileConverter struct {
	ImageFileConverter

	incrementModTimeRequest bool
	useNamedStubs           bool
	currentTime             time.Time
	stubs                   map[string]StubFileInfo
}

func (s *StubImageFileConverter) ImageFileToDbImage(imageFile *apitype.ImageFile) (*Image, map[string]string, error) {
	fileStat, _ := s.GetImageFileStats(imageFile)
	metaData := map[string]string{}
	return &Image{
		Id:              0,
		Name:            imageFile.FileName(),
		FileName:        imageFile.FileName(),
		ByteSize:        1234,
		ExifOrientation: 1,
		ImageAngle:      90,
		ImageFlip:       true,
		CreatedTime:     fileStat.ModTime(),
		Width:           1024,
		Height:          2048,
		ModifiedTime:    fileStat.ModTime(),
	}, metaData, nil
}

func (s *StubImageFileConverter) AddStubFile(name string, modTime time.Time) {
	s.stubs[name] = StubFileInfo{
		modTime: modTime,
	}
}

func (s *StubImageFileConverter) GetImageFileStats(imageFile *apitype.ImageFile) (os.FileInfo, error) {
	if s.incrementModTimeRequest {
		s.currentTime = s.currentTime.Add(time.Second)
	}

	if !s.useNamedStubs {
		return &StubFileInfo{
			modTime: s.currentTime,
		}, nil
	} else {
		info := s.stubs[imageFile.FileName()]
		return &info, nil
	}
}

func (s *StubImageFileConverter) SetIncrementModTimeRequest(value bool) {
	s.incrementModTimeRequest = value
	s.stubs = map[string]StubFileInfo{}
	s.useNamedStubs = false
}

func (s *StubImageFileConverter) SetNamedStubs(value bool) {
	s.useNamedStubs = value
	s.stubs = map[string]StubFileInfo{}
	s.incrementModTimeRequest = false
}

type StubFileInfo struct {
	os.FileInfo

	modTime time.Time
}

func (s *StubFileInfo) ModTime() time.Time {
	return s.modTime
}
