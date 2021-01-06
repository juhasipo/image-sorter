package database

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"vincit.fi/image-sorter/api/apitype"
)

func initSUT() *CategoryStore {
	return NewCategoryStore(NewInMemoryStore())
}

func TestCategoryStore_AddCategory(t *testing.T) {
	a := assert.New(t)

	sut := initSUT()

	t.Run("One", func(t *testing.T) {
		category, err := sut.AddCategory(apitype.NewCategory("Category 1", "cat1", "C"))

		if a.Nil(err) {
			a.Equal(apitype.CategoryId(1), category.GetId())
			a.Equal("Category 1", category.GetName())
			a.Equal("cat1", category.GetSubPath())
			a.Equal(uint(0x43), category.GetShortcut())
		}
	})

	t.Run("Another one", func(t *testing.T) {
		category, err := sut.AddCategory(apitype.NewCategory("Category 2", "cat2", "D"))

		if a.Nil(err) {
			a.Equal(apitype.CategoryId(2), category.GetId())
			a.Equal("Category 2", category.GetName())
			a.Equal("cat2", category.GetSubPath())
			a.Equal(uint(0x44), category.GetShortcut())
		}
	})

	t.Run("Duplicate", func(t *testing.T) {
		category, err := sut.AddCategory(apitype.NewCategory("Category 1", "cat1", "C"))

		if a.Nil(err) {
			a.Equal(apitype.CategoryId(1), category.GetId())
			a.Equal("Category 1", category.GetName())
			a.Equal("cat1", category.GetSubPath())
			a.Equal(uint(0x43), category.GetShortcut())
		}
	})

	t.Run("Duplicate with different sub path and shortcut", func(t *testing.T) {
		category, err := sut.AddCategory(apitype.NewCategory("Category 1", "cat1_", "E"))

		if a.Nil(err) {
			a.Equal(apitype.CategoryId(1), category.GetId())
			a.Equal("Category 1", category.GetName())
			a.Equal("cat1", category.GetSubPath())
			a.Equal(uint(0x43), category.GetShortcut())
		}
	})

	categories, err := sut.GetCategories()

	if a.Nil(err) {
		a.Equal(2, len(categories))
		a.Equal(apitype.CategoryId(1), categories[0].GetId())
		a.Equal("Category 1", categories[0].GetName())
		a.Equal("cat1", categories[0].GetSubPath())
		a.Equal(uint(0x43), categories[0].GetShortcut())

		a.Equal(apitype.CategoryId(2), categories[1].GetId())
		a.Equal("Category 2", categories[1].GetName())
		a.Equal("cat2", categories[1].GetSubPath())
		a.Equal(uint(0x44), categories[1].GetShortcut())
	}
}

func TestCategoryStore_ResetCategories(t *testing.T) {
	a := assert.New(t)

	sut := initSUT()

	cat1, err := sut.AddCategory(apitype.NewCategory("Category 1", "cat1", "C"))
	a.Nil(err)
	cat2, err := sut.AddCategory(apitype.NewCategory("Category 2", "cat2", "D"))
	a.Nil(err)
	_, err = sut.AddCategory(apitype.NewCategory("Category 3", "cat3", "E"))
	a.Nil(err)

	t.Run("Reset with partially new values", func(t *testing.T) {
		err = sut.ResetCategories([]*apitype.Category{
			cat1,
			cat2,
			apitype.NewCategory("Category 4", "cat4", "F"),
		})

		if a.Nil(err) {
			categories, err := sut.GetCategories()
			if a.Nil(err) {
				a.Equal(3, len(categories))

				a.Equal(apitype.CategoryId(1), categories[0].GetId())
				a.Equal("Category 1", categories[0].GetName())
				a.Equal("cat1", categories[0].GetSubPath())
				a.Equal(uint(0x43), categories[0].GetShortcut())

				a.Equal(apitype.CategoryId(2), categories[1].GetId())
				a.Equal("Category 2", categories[1].GetName())
				a.Equal("cat2", categories[1].GetSubPath())
				a.Equal(uint(0x44), categories[1].GetShortcut())

				a.Equal(apitype.CategoryId(4), categories[2].GetId())
				a.Equal("Category 4", categories[2].GetName())
				a.Equal("cat4", categories[2].GetSubPath())
				a.Equal(uint(0x46), categories[2].GetShortcut())
			}

		}
	})

	t.Run("Reset with new values", func(t *testing.T) {
		err = sut.ResetCategories([]*apitype.Category{
			apitype.NewCategory("Category 5", "cat5", "G"),
			apitype.NewCategory("Category 6", "cat6", "H"),
		})

		if a.Nil(err) {
			categories, err := sut.GetCategories()
			if a.Nil(err) {
				a.Equal(2, len(categories))

				a.Equal(apitype.CategoryId(5), categories[0].GetId())
				a.Equal("Category 5", categories[0].GetName())
				a.Equal("cat5", categories[0].GetSubPath())
				a.Equal(uint(0x47), categories[0].GetShortcut())

				a.Equal(apitype.CategoryId(6), categories[1].GetId())
				a.Equal("Category 6", categories[1].GetName())
				a.Equal("cat6", categories[1].GetSubPath())
				a.Equal(uint(0x48), categories[1].GetShortcut())
			}

		}
	})

	t.Run("Reset with no values", func(t *testing.T) {
		err = sut.ResetCategories([]*apitype.Category{
		})

		if a.Nil(err) {
			categories, err := sut.GetCategories()
			if a.Nil(err) {
				a.Equal(0, len(categories))
			}

		}
	})

}

func TestCategoryStore_GetCategoryById(t *testing.T) {
	a := assert.New(t)

	sut := initSUT()

	cat1, err := sut.AddCategory(apitype.NewCategory("Category 1", "cat1", "C"))
	a.Nil(err)
	cat2, err := sut.AddCategory(apitype.NewCategory("Category 2", "cat2", "D"))
	a.Nil(err)
	_, err = sut.AddCategory(apitype.NewCategory("Category 3", "cat3", "E"))
	a.Nil(err)

	category1 := sut.GetCategoryById(cat1.GetId())
	if a.NotNil(category1) {

		a.Equal(apitype.CategoryId(1), category1.GetId())
		a.Equal("Category 1", category1.GetName())
		a.Equal("cat1", category1.GetSubPath())
		a.Equal(uint(0x43), category1.GetShortcut())
	}

	category2 := sut.GetCategoryById(cat2.GetId())
	if a.NotNil(category2) {

		a.Equal(apitype.CategoryId(2), category2.GetId())
		a.Equal("Category 2", category2.GetName())
		a.Equal("cat2", category2.GetSubPath())
		a.Equal(uint(0x44), category2.GetShortcut())
	}
}
