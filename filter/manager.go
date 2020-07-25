package filter

import (
	"log"
	"vincit.fi/image-sorter/common"
)

type Filter struct {
	id        string
	operation ImageOperation
}

func (s *Filter) GetOperation() ImageOperation {
	return s.operation
}

type Manager struct {
	filtersToApply map[string][]*Filter
	filters        map[string]*Filter
}

func FilterManagerNew() *Manager {
	return &Manager{
		filtersToApply: map[string][]*Filter{},
	}
}

func (s *Manager) AddFilterForImage(handle *common.Handle, id string) {
	if filter, ok := s.filters[id]; !ok {
		log.Printf("Could not find filter '%s'", id)
	} else if filterList, ok := s.filtersToApply[handle.GetId()]; ok {
		s.filtersToApply[handle.GetId()] = append(filterList, filter)
	}
}

func (s *Manager) AddFilter(filter *Filter) {
	s.filters[filter.id] = filter
}

func (s *Manager) GetFilters(handle *common.Handle) []*Filter {
	if filters, ok := s.filtersToApply[handle.GetId()]; ok {
		return filters
	} else {
		return []*Filter{}
	}
}
