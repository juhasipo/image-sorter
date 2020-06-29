package common

import (
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"strings"
)

var (
	supportedFileEndings = map[string]bool{".jpg": true, ".jpeg": true}
)

func LoadImages(dir string) []*Handle {
	var handles []*Handle
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Scanning directory '%s'", dir)
	for _, file := range files {
		filePath := path.Join(dir, file.Name())
		extension := filepath.Ext(filePath)
		if IsSupported(extension) {
			filePath := path.Join(dir, file.Name())
			log.Printf(" - %s", filePath)
			handles = append(handles, &Handle{id: filePath, path: filePath})
		}
	}

	return handles
}

func IsSupported(extension string) bool {
	return supportedFileEndings[strings.ToLower(extension)]
}
