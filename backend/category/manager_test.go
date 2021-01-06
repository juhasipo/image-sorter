package category

import (
	"bufio"
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func (s *MockSender) SendToTopicWithData(topic api.Topic, data ...interface{}) {
	s.Called(topic, data)
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

func TestResetCategories(t *testing.T) {
	a := assert.New(t)

	params := common.NewEmptyParams()
	sender := new(MockSender)
	sender.On("SendToTopicWithData", api.CategoriesUpdated, mock.Anything).Return()

	store := database.NewInMemoryStore()
	categoryStore := database.NewCategoryStore(store)
	sut := New(params, sender, categoryStore)

	_, _ = categoryStore.AddCategory(apitype.NewCategory("Cat 1", "C1", "1"))
	cat2, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 2", "C2", "2"))
	_, _ = categoryStore.AddCategory(apitype.NewCategory("Cat 3", "C3", "3"))
	cat4, _ := categoryStore.AddCategory(apitype.NewCategory("Cat 4", "C4", "4"))

	sut.Save([]*apitype.Category{
		apitype.NewCategoryWithId(cat2.GetId(), "Cat 2", "C2", "2"),
		apitype.NewCategoryWithId(cat4.GetId(), "Cat 4_", "C4", "4"),
		apitype.NewCategory("Cat 5", "C5", "5"),
	})

	categories, _ := categoryStore.GetCategories()

	if a.Equal(3, len(categories)) {
		a.Equal(categories[0].GetId(), cat2.GetId())
		a.Equal(categories[0].GetName(), "Cat 2")
		a.Equal(categories[1].GetId(), cat4.GetId())
		a.Equal(categories[1].GetName(), "Cat 4_")
		a.Equal(categories[2].GetName(), "Cat 5")
	}
}
