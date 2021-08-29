package widget

import (
	"github.com/AllenDang/giu"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common"
)

type editableCategory struct {
	id       apitype.CategoryId
	name     string
	subPath  string
	shortcut string
}

func fromCategory(category *apitype.Category) *editableCategory {
	return &editableCategory{
		id:       category.Id(),
		name:     category.Name(),
		subPath:  category.SubPath(),
		shortcut: category.ShortcutAsString(),
	}
}

func (s *editableCategory) toCategory() *apitype.Category {
	path := s.subPath
	if path == "" {
		path = s.name
	}
	return apitype.NewCategoryWithId(
		s.id, s.name, path, s.shortcut,
	)
}

type CategoryEditWidget struct {
	categories             []*editableCategory
	selectedCategoryToEdit int
	onSave                 func(asDefault bool, categories []*apitype.Category)
	onClose                func()
}

func CategoryEdit(onSave func(asDefault bool, categories []*apitype.Category), onClose func()) *CategoryEditWidget {
	return &CategoryEditWidget{
		categories:             []*editableCategory{},
		selectedCategoryToEdit: -1,
		onSave:                 onSave,
		onClose:                onClose,
	}
}

func (s *CategoryEditWidget) SetCategories(categories []*apitype.Category) {
	s.categories = []*editableCategory{}
	for _, category := range categories {
		s.categories = append(s.categories, fromCategory(category))
	}
}

func (s *CategoryEditWidget) HandleKeys() {
	if s.selectedCategoryToEdit >= 0 {
		for i := giu.KeyUnknown; i < giu.KeyLast; i++ {
			if giu.IsKeyPressed(i) {
				s.categories[s.selectedCategoryToEdit].shortcut = common.KeyvalName(uint(i))
				s.selectedCategoryToEdit = -1
			}
		}
	}
}

func (s *CategoryEditWidget) Build() {
	giu.Custom(func() {
		width, height := giu.GetAvailableRegion()
		var categoryRows []*giu.TableRowWidget
		for i, category := range s.categories {
			ci := i
			cat := category
			tableRow := giu.TableRow(
				giu.InputText(&cat.name),
				giu.InputText(&cat.subPath).Hint(cat.name),
				giu.Custom(func() {
					n := cat.shortcut
					if s.selectedCategoryToEdit == ci {
						n = "Press any key..."
					}
					giu.Selectable(n).OnClick(func() {
						// Change to "wait for key" mode
						s.selectedCategoryToEdit = ci
					}).Build()
				}),
				giu.Row(
					giu.Button("Remove").OnClick(func() {
						s.categories = append(s.categories[:ci], s.categories[ci+1:]...)
					}),
				),
			)
			categoryRows = append(categoryRows, tableRow)
		}
		categoryRows = append(categoryRows, giu.TableRow(
			giu.Button("Add new").OnClick(func() {
				s.categories = append(s.categories, &editableCategory{
					name:     "",
					subPath:  "",
					shortcut: "",
				})
			}),
			giu.Row(),
			giu.Row(),
			giu.Row(),
		))

		giu.Column(
			giu.Row(
				giu.Table().
					Columns(
						giu.TableColumn("Category"),
						giu.TableColumn("Path"),
						giu.TableColumn("Shortcut"),
						giu.TableColumn("Actions"),
					).
					Rows(categoryRows...).
					Size(width, height-20),
			),
			giu.Row(
				giu.Button("Save").OnClick(func() {
					var cats []*apitype.Category
					for _, category := range s.categories {
						cats = append(cats, category.toCategory())
					}
					s.onSave(false, cats)
				}),
				giu.Button("Save as Defaults").OnClick(func() {
					var cats []*apitype.Category
					for _, category := range s.categories {
						cats = append(cats, category.toCategory())
					}
					s.onSave(true, cats)
				}),
				giu.Button("Close w/o saving").OnClick(func() {
					s.onClose()
				}),
			)).Build()
	}).Build()
}
