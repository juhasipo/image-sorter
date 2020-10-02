package common

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"vincit.fi/image-sorter/common/logger"
)

// Copies file. Creates destination directories if they don't exist
func CopyFile(srcPath string, srcFile string, dstPath string, dstFile string) error {
	srcFilePath := filepath.Join(srcPath, srcFile)
	dstFilePath := filepath.Join(dstPath, dstFile)

	if err := MakeDirectoriesIfNotExist(srcPath, dstPath); err != nil {
		return err
	}

	return copyInternal(srcFilePath, dstFilePath)
}

// Creates dstPath directories if they do not exist. Uses srdDir file mode
// for the created directories
func MakeDirectoriesIfNotExist(srcPath string, dstPath string) error {
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		if info, err := os.Stat(srcPath); err != nil {
			logger.Error.Println("Could not resolve srdPath: " + srcPath)
		} else if err := os.MkdirAll(dstPath, info.Mode()); err != nil {
			return err
		}
	}
	return nil
}

func copyInternal(src string, dst string) error {
	if sourceFileStat, err := os.Stat(src); err != nil {
		return err
	} else if !sourceFileStat.Mode().IsRegular() {
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

// Removes file
func RemoveFile(src string) error {
	return os.Remove(src)
}

// Returns true if file at the path exists. False if not
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
