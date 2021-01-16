package category

import (
	"bufio"
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"path/filepath"
	"testing"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/backend/database"
	"vincit.fi/image-sorter/common"
)

type MockSender struct {
	api.Sender
	mock.Mock
}

func (s *MockSender) SendToTopic(topic api.Topic) {
	s.Called(topic)
}

func (s *MockSender) SendCommandToTopic(topic api.Topic, command apitype.Command) {
	s.Called(topic, command)
}

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
		a.Equal("Name", categories[0].Name())
		a.Equal("Path", categories[0].SubPath())
		a.Equal("P", categories[0].ShortcutAsString())
	})
	t.Run("Multiple", func(t *testing.T) {
		values := []string{
			"Name:Path:P",
			"Another:A",
			"Some:S",
		}
		categories := fromCategoriesStrings(values)
		a.Equal(3, len(categories))
		a.Equal("Name", categories[0].Name())
		a.Equal("Path", categories[0].SubPath())
		a.Equal("P", categories[0].ShortcutAsString())

		a.Equal("Another", categories[1].Name())
		a.Equal("Another", categories[1].SubPath())
		a.Equal("A", categories[1].ShortcutAsString())

		a.Equal("Some", categories[2].Name())
		a.Equal("Some", categories[2].SubPath())
		a.Equal("S", categories[2].ShortcutAsString())
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
		err := writeCategoriesToBuffer(writer, categories)
		if a.Nil(err) {
			a.Equal("#version:1\n", buf.String())
		}
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
		a.Equal("Name1", categories[0].Name())
		a.Equal("Path1", categories[0].SubPath())
		a.Equal("S", categories[0].ShortcutAsString())
	})

	t.Run("Multiple", func(t *testing.T) {

		buf := bytes.NewBuffer([]byte{})
		buf.WriteString("#version:1\nName1:Path1:S\nName2:Path2:S\n")
		reader := bufio.NewReader(buf)
		categories := readCategoriesFromReader(reader)

		a.Equal(2, len(categories))
		a.Equal("Name1", categories[0].Name())
		a.Equal("Path1", categories[0].SubPath())
		a.Equal("S", categories[0].ShortcutAsString())

		a.Equal("Name2", categories[1].Name())
		a.Equal("Path2", categories[1].SubPath())
		a.Equal("S", categories[1].ShortcutAsString())
	})
}

func TestResetCategories(t *testing.T) {
	a := assert.New(t)

	params := common.NewEmptyParams()
	sender := new(MockSender)
	sender.On("SendCommandToTopic", api.CategoriesUpdated, mock.Anything).Return()

	memoryDatabase := database.NewInMemoryDatabase()
	categoryStore := database.NewCategoryStore(memoryDatabase)
	sut := New(params, sender, categoryStore)

	_, _ = categoryStore.AddCategory(apitype.NewCategory("Cat 1", "C1", "1"))
	cat2, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 2", "C2", "2"))
	_, _ = categoryStore.AddCategory(apitype.NewCategory("Cat 3", "C3", "3"))
	cat4, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 4", "C4", "4"))

	sut.Save(&api.SaveCategoriesCommand{
		Categories: []*apitype.Category{
			apitype.NewCategoryWithId(cat2.Id(), "Cat 2", "C2", "2"),
			apitype.NewCategoryWithId(cat4.Id(), "Cat 4_", "C4", "4"),
			apitype.NewCategory("Cat 5", "C5", "5"),
		},
	})

	categories, _ := categoryStore.GetCategories()

	if a.Equal(3, len(categories)) {
		a.Equal(categories[0].Id(), cat2.Id())
		a.Equal(categories[0].Name(), "Cat 2")
		a.Equal(categories[1].Id(), cat4.Id())
		a.Equal(categories[1].Name(), "Cat 4_")
		a.Equal(categories[2].Name(), "Cat 5")
	}
}

func TestSaveCategoriesToFile_AndLoadCategoriesFromFile(t *testing.T) {
	a := assert.New(t)

	params := common.NewEmptyParams()
	sender := new(MockSender)
	sender.On("SendToTopicWithData", api.CategoriesUpdated, mock.Anything).Return()

	memoryDatabase := database.NewInMemoryDatabase()
	categoryStore := database.NewCategoryStore(memoryDatabase)
	sut := newManager(params, sender, categoryStore)

	_, _ = categoryStore.AddCategory(apitype.NewCategory("Cat 1", "C1", "1"))
	_, _ = categoryStore.AddCategory(apitype.NewCategory("Cat 2", "C2", "2"))
	_, _ = categoryStore.AddCategory(apitype.NewCategory("Cat 3", "C3", "3"))
	_, _ = categoryStore.AddCategory(apitype.NewCategory("Cat 4", "C4", "4"))

	dir, err := ioutil.TempDir("", "test_dir")
	file := ".categories"
	err = sut.saveCategoriesToFile(dir, file, sut.GetCategories())

	if a.Nil(err) {
		categories := sut.loadCategoriesFromFile(filepath.Join(dir, file))
		a.Equal(4, len(categories))
		a.Equal("Cat 1", categories[0].Name())
		a.Equal("C1", categories[0].SubPath())
		a.Equal(uint(0x31), categories[0].Shortcut())

		a.Equal("Cat 2", categories[1].Name())
		a.Equal("C2", categories[1].SubPath())
		a.Equal(uint(0x32), categories[1].Shortcut())

		a.Equal("Cat 3", categories[2].Name())
		a.Equal("C3", categories[2].SubPath())
		a.Equal(uint(0x33), categories[2].Shortcut())

		a.Equal("Cat 4", categories[3].Name())
		a.Equal("C4", categories[3].SubPath())
		a.Equal(uint(0x34), categories[3].Shortcut())
	}
}

func TestSaveCategoriesToFile_AndLoadCategoriesFromFiles(t *testing.T) {
	a := assert.New(t)

	params := common.NewEmptyParams()
	sender := new(MockSender)
	sender.On("SendToTopicWithData", api.CategoriesUpdated, mock.Anything).Return()

	memoryDatabase := database.NewInMemoryDatabase()
	categoryStore := database.NewCategoryStore(memoryDatabase)
	sut := newManager(params, sender, categoryStore)

	t.Run("File doesn'r exist", func(t *testing.T) {
		categories := sut.loadCategoriesFromFirstExistingLocation([]string{
			filepath.Join("", "not-file-1"),
			filepath.Join("", "not-file-2"),
			filepath.Join("", "not-file-3"),
			filepath.Join("", "not-file-4"),
		})
		a.Equal(0, len(categories))
	})

	t.Run("One file exists", func(t *testing.T) {

		_, _ = categoryStore.AddCategory(apitype.NewCategory("Cat 1", "C1", "1"))
		_, _ = categoryStore.AddCategory(apitype.NewCategory("Cat 2", "C2", "2"))
		_, _ = categoryStore.AddCategory(apitype.NewCategory("Cat 3", "C3", "3"))
		_, _ = categoryStore.AddCategory(apitype.NewCategory("Cat 4", "C4", "4"))

		dir, err := ioutil.TempDir("", "test_dir")
		file := ".categories"
		err = sut.saveCategoriesToFile(dir, file, sut.GetCategories())

		if a.Nil(err) {
			categories := sut.loadCategoriesFromFirstExistingLocation([]string{
				filepath.Join(dir, "not-file-1"),
				filepath.Join(dir, "not-file-2"),
				filepath.Join(dir, file),
				filepath.Join(dir, "not-file-3"),
			})
			a.Equal(4, len(categories))
			a.Equal("Cat 1", categories[0].Name())
			a.Equal("C1", categories[0].SubPath())
			a.Equal(uint(0x31), categories[0].Shortcut())

			a.Equal("Cat 2", categories[1].Name())
			a.Equal("C2", categories[1].SubPath())
			a.Equal(uint(0x32), categories[1].Shortcut())

			a.Equal("Cat 3", categories[2].Name())
			a.Equal("C3", categories[2].SubPath())
			a.Equal(uint(0x33), categories[2].Shortcut())

			a.Equal("Cat 4", categories[3].Name())
			a.Equal("C4", categories[3].SubPath())
			a.Equal(uint(0x34), categories[3].Shortcut())
		}
	})
}
