package category

import "vincit.fi/image-sorter/image"

type Operation int
const(
	COPY Operation = 0
	MOVE Operation = 1
)

type Entry struct {
	name string
	subPath string
}

type CategorizedImage struct {
	category *Entry
	operation Operation
}

type Manager struct {
	categories []*Entry
}

func (s *Manager) AddCategory(name string, subPath string) *Entry {
	category := Entry {name: name, subPath: subPath}
	s.categories = append(s.categories, &category)
	return &category
}

func (s *Manager) GetCategories() []*Entry {
	return s.categories
}

func (s *Manager) ToggleCategory(image *library.Handle, category *Entry, operation Operation) {
	if val, ok := s.imageCategory[image]; ok {
		if val.operation != operation {
			val.operation = operation
		} else {
			delete(s.imageCategory, image)
		}
	} else {
		s.imageCategory[image] = CategorizedImage{
			category:  category,
			operation: operation,
		}
	}
}
