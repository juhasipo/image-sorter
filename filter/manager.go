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

func (s *Manager) GetFilters(handle *common.Handle, options common.PersistCategorizationCommand) []*Filter {
	filtersToApply := s.getFiltersForHandle(handle)

	if options.ShouldFixOrientation() {
		filtersToApply = append(filtersToApply, &Filter{
			id:        "exifRotate",
			operation: ImageExifRotateNew(),
		})
	}
	return filtersToApply
}

func (s *Manager) getFiltersForHandle(handle *common.Handle) []*Filter {
	var filtersToApply []*Filter
	if f, ok := s.filtersToApply[handle.GetId()]; ok {
		filtersToApply = make([]*Filter, len(f)+1)
		copy(filtersToApply, f)
	} else {
		filtersToApply = []*Filter{}
	}
	return filtersToApply
}
