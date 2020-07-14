package library

import (
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"strings"
	"vincit.fi/image-sorter/common"
)

var (
	supportedFileEndings = map[string]bool{".jpg": true, ".jpeg": true}
)

func LoadImages(dir string) []*common.Handle {
	var handles []*common.Handle
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
			handles = append(handles, common.HandleNew(filePath))
		}
	}
	log.Printf("Found %d images", len(handles))

	return handles
}

func IsSupported(extension string) bool {
	return supportedFileEndings[strings.ToLower(extension)]
}
