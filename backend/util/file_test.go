package util

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestDoesFileExist_FileExists(t *testing.T) {
	require.True(t, DoesFileExist("file_test.go"))
}
func TestDoesFileExist_DirExists(t *testing.T) {
	require.True(t, DoesFileExist("../util"))
}
func TestDoesFileExist_DoesntExist(t *testing.T) {
	require.False(t, DoesFileExist("foobarfile"))
}

func TestMakeDirectoriesIfNotExist(t *testing.T) {
	a := assert.New(t)

	dir, err := ioutil.TempDir("", "test_dir")
	if a.Nil(err) {
		newDir1 := filepath.Join(dir, "test1")
		newDir2 := filepath.Join(newDir1, "test2")
		err = MakeDirectoriesIfNotExist(dir, newDir2)
		a.Nil(err)

		a.True(DoesFileExist(newDir1))
		a.True(DoesFileExist(newDir2))

		err = os.Remove(newDir2)
		a.Nil(err)
		err = os.Remove(newDir1)
		a.Nil(err)
		err = os.Remove(dir)
		a.Nil(err)
	}
}

func TestCopyFile(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)

	dir, err := ioutil.TempDir("", "test_dir")
	if a.Nil(err) {
		file1, err := ioutil.TempFile(dir, "file_1_")
		if a.Nil(err) {
			// Write content to file and copy
			_, err := file1.WriteString("Test string")
			r.Nil(err)
			filename := filepath.Base(file1.Name())
			err = CopyFile(dir, filename, dir, "file2")
			r.Nil(err)
			file1.Close()

			// Assert that file exists and has the correct content
			copiedFile := filepath.Join(dir, "file2")
			a.True(DoesFileExist(copiedFile))
			file2Bytes, err := ioutil.ReadFile(copiedFile)
			if a.Nil(err) {
				file2Content := string(file2Bytes)
				a.Equal("Test string", file2Content)
			}

			// Clean up
			err = os.Remove(copiedFile)
			a.Nil(err)
			err = os.Remove(file1.Name())
			a.Nil(err)
			err = os.Remove(dir)
			a.Nil(err)
		}
	}
}

func TestCopyFile_FileDoesntExist(t *testing.T) {
	a := assert.New(t)

	dir, err := ioutil.TempDir("", "test_dir")
	if a.Nil(err) {
		err = CopyFile(dir, "foobar", dir, "file2")
		a.NotNil(err)

		err = os.Remove(dir)
		a.Nil(err)
	}
}

func TestRemoveFile(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)

	dir, err := ioutil.TempDir("", "test_dir")
	a.Nil(err)
	file1, err := ioutil.TempFile(dir, "file_1_")
	r.Nil(err)
	r.True(DoesFileExist(file1.Name()))
	file1.Close()

	err = RemoveFile(file1.Name())
	a.Nil(err)
	r.False(DoesFileExist(file1.Name()))

	err = os.Remove(dir)
	a.Nil(err)
}

func TestRemoveFile_FileNotExists(t *testing.T) {
	a := assert.New(t)

	dir, err := ioutil.TempDir("", "test_dir")
	a.Nil(err)

	err = RemoveFile(filepath.Join(dir, "not_existing_file"))
	a.NotNil(err)

	err = os.Remove(dir)
	a.Nil(err)
}
