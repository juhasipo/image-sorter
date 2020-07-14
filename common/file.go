package common

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

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

func DoesFileExist(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		return false
	}
}

func GetFirstExistingFilePath(filePaths []string) string {
	var filePath string
	for _, path := range filePaths {
		if DoesFileExist(path) {
			filePath = path
			break
		}
	}
	return filePath
}
