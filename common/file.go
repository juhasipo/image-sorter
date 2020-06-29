package common

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
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

func CopyFile(srcPath string, srcFile string, dstPath string, dstFile string) error {
	srcFilePath := filepath.Join(srcPath, srcFile)
	dstFilePath := filepath.Join(dstPath, dstFile)
	log.Printf("   - Copying '%s' to '%s'", srcFilePath, dstFilePath)

	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		info, _ := os.Stat(srcPath)
		os.MkdirAll(dstPath, info.Mode())
	}

	return CopyInternal(srcFilePath, dstFilePath)
}

func CopyInternal(src string, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
			return err
	}

	if !sourceFileStat.Mode().IsRegular() {
			return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
			return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
			return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

func RemoveFile(src string) error {
	log.Printf("   - Deleting '%s'", src)
	return os.Remove(src)
}
