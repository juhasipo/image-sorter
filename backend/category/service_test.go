package category

import (
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

func TestResetCategories(t *testing.T) {
	a := assert.New(t)

	params := common.NewEmptyParams()
	sender := new(MockSender)
	sender.On("SendCommandToTopic", api.CategoriesUpdated, mock.Anything).Return()

	memoryDatabase := database.NewInMemoryDatabase()
	categoryStore := database.NewCategoryStore(memoryDatabase)
	sut := NewCategoryService(params, sender, categoryStore)

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
