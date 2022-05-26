package util

type Set[V comparable] struct {
	values map[V]bool
}

func NewSet[V comparable]() *Set[V] {
	return &Set[V]{
		values: map[V]bool{},
	}
}

func (s *Set[V]) Add(value V) {
	s.values[value] = true
}
func (s *Set[V]) Remove(value V) {
	delete(s.values, value)
}
func (s *Set[V]) Contains(value V) bool {
	return s.values[value]
}
