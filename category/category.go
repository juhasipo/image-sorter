package category

type Operation int
const(
	NONE Operation = 0
	COPY Operation = 1
	MOVE Operation = 2
)

type Entry struct {
	name string
	subPath string
}

type CategorizedImage struct {
	category *Entry
	operation Operation
}

func CategorizedImageNew(entry *Entry, operation Operation) *CategorizedImage {
	return &CategorizedImage {
		category: entry,
		operation: operation,
	}
}

func (s* CategorizedImage) GetOperation() Operation {
	return s.operation
}

func (s* CategorizedImage) SetOperation(operation Operation) {
	s.operation = operation
}

func (s* CategorizedImage) GetEntry() *Entry {
	return s.category
}

type Manager struct {
	categories []*Entry
}

func FromCategories(categories []string) []*Entry {
	var categoryEntries []*Entry
	for i := range categories {
		categoryName := categories[i]
		categoryEntries = append(categoryEntries, &Entry {
			name: categoryName,
			subPath: categoryName,
		})
	}
	return categoryEntries
}

func (s *Manager) AddCategory(name string, subPath string) *Entry {
	category := Entry {name: name, subPath: subPath}
	s.categories = append(s.categories, &category)
	return &category
}

func (s *Manager) GetCategories() []*Entry {
	return s.categories
}
