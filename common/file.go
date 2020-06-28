package common

import (
	"io/ioutil"
	"log"
	"path"
	"strings"
)

var (
	supportedFileEndings = []string {".jpg", ".jpeg"}
)

func LoadImages(dir string) []*Handle {
	var handles []*Handle
	files, err := ioutil.ReadDir(dir)
	if err != nil {
        log.Fatal(err)
    }

    log.Printf("Scanning directory '%s'", dir)
	for _, file := range files {
		fileNameLower := strings.ToLower(file.Name())
		if IsSupported(fileNameLower) {
			filePath := path.Join(dir, file.Name())
			log.Printf(" - %s", filePath)
			handles = append(handles, &Handle{id: filePath, path: filePath})
		}
    }

	return handles
}

func IsSupported(name string) bool {
	for _, ending := range supportedFileEndings {
		if strings.HasSuffix(name, ending) {
			return true
		}
	}
	return false
}

