package category

import (
	"bufio"
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
	"vincit.fi/image-sorter/api/apitype"
)

func TestParse(t *testing.T) {
	a := assert.New(t)

	t.Run("2 parts", func(t *testing.T) {
		category, path, shortcut := Parse("Name:P")
		a.Equal("Name", category)
		a.Equal("Name", path)
		a.Equal("P", shortcut)
	})
	t.Run("3 parts", func(t *testing.T) {
		category, path, shortcut := Parse("Name:Path:P")
		a.Equal("Name", category)
		a.Equal("Path", path)
		a.Equal("P", shortcut)
	})
}

func TestFromCategoriesStrings(t *testing.T) {
	a := assert.New(t)

	t.Run("Zero", func(t *testing.T) {
		values := []string{
		}
		categories := fromCategoriesStrings(values)
		a.Equal(0, len(categories))
	})
	t.Run("One", func(t *testing.T) {
		values := []string{
			"Name:Path:P",
		}
		categories := fromCategoriesStrings(values)
		a.Equal(1, len(categories))
		a.Equal("Name", categories[0].GetName())
		a.Equal("Path", categories[0].GetSubPath())
		a.Equal("P", categories[0].GetShortcutAsString())
	})
	t.Run("Multiple", func(t *testing.T) {
		values := []string{
			"Name:Path:P",
			"Another:A",
			"Some:S",
		}
		categories := fromCategoriesStrings(values)
		a.Equal(3, len(categories))
		a.Equal("Name", categories[0].GetName())
		a.Equal("Path", categories[0].GetSubPath())
		a.Equal("P", categories[0].GetShortcutAsString())

		a.Equal("Another", categories[1].GetName())
		a.Equal("Another", categories[1].GetSubPath())
		a.Equal("A", categories[1].GetShortcutAsString())

		a.Equal("Some", categories[2].GetName())
		a.Equal("Some", categories[2].GetSubPath())
		a.Equal("S", categories[2].GetShortcutAsString())
	})
	t.Run("Nil", func(t *testing.T) {
		categories := fromCategoriesStrings(nil)
		a.Equal(0, len(categories))
	})
}

func TestWriteCategoriesToBuffer(t *testing.T) {
	a := assert.New(t)

	t.Run("Zero", func(t *testing.T) {
		var categories []*apitype.Category

		buf := bytes.NewBuffer([]byte{})
		writer := bufio.NewWriter(buf)
		writeCategoriesToBuffer(writer, categories)

		a.Equal("#version:1\n", buf.String())
	})

	t.Run("One", func(t *testing.T) {
		categories := []*apitype.Category{
			apitype.NewCategoryWithId(1, "Name1", "Path1", "S"),
		}

		buf := bytes.NewBuffer([]byte{})
		writer := bufio.NewWriter(buf)
		writeCategoriesToBuffer(writer, categories)

		a.Equal("#version:1\nName1:Path1:S\n", buf.String())
	})

	t.Run("Multiple", func(t *testing.T) {
		categories := []*apitype.Category{
			apitype.NewCategoryWithId(1, "Name1", "Path1", "S"),
			apitype.NewCategoryWithId(2, "Name2", "Path2", "S"),
		}

		buf := bytes.NewBuffer([]byte{})
		writer := bufio.NewWriter(buf)
		writeCategoriesToBuffer(writer, categories)

		a.Equal("#version:1\nName1:Path1:S\nName2:Path2:S\n", buf.String())
	})
}

func TestReadCategoriesFromReader(t *testing.T) {
	a := assert.New(t)

	t.Run("Zero", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{})
		buf.WriteString("#version:1\n")
		reader := bufio.NewReader(buf)
		categories := readCategoriesFromReader(reader)

		a.Equal(0, len(categories))
	})

	t.Run("One", func(t *testing.T) {

		buf := bytes.NewBuffer([]byte{})
		buf.WriteString("#version:1\nName1:Path1:S\n")
		reader := bufio.NewReader(buf)
		categories := readCategoriesFromReader(reader)

		a.Equal(1, len(categories))
		a.Equal("Name1", categories[0].GetName())
		a.Equal("Path1", categories[0].GetSubPath())
		a.Equal("S", categories[0].GetShortcutAsString())
	})

	t.Run("Multiple", func(t *testing.T) {

		buf := bytes.NewBuffer([]byte{})
		buf.WriteString("#version:1\nName1:Path1:S\nName2:Path2:S\n")
		reader := bufio.NewReader(buf)
		categories := readCategoriesFromReader(reader)

		a.Equal(2, len(categories))
		a.Equal("Name1", categories[0].GetName())
		a.Equal("Path1", categories[0].GetSubPath())
		a.Equal("S", categories[0].GetShortcutAsString())

		a.Equal("Name2", categories[1].GetName())
		a.Equal("Path2", categories[1].GetSubPath())
		a.Equal("S", categories[1].GetShortcutAsString())
	})
}
