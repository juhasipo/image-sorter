package filter

import (
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
)

type Filter struct {
	id        string
	operation apitype.ImageOperation
}

func (s *Filter) Operation() apitype.ImageOperation {
	return s.operation
}

type FilterService struct {
	filtersToApply map[apitype.ImageId][]*Filter
	filters        map[string]*Filter
}

func NewFilterService() *FilterService {
	return &FilterService{
		filtersToApply: map[apitype.ImageId][]*Filter{},
	}
}

func (s *FilterService) AddFilterForImage(imageFile *apitype.ImageFile, id string) {
	if filter, ok := s.filters[id]; !ok {
		logger.Error.Printf("Could not find filter '%s'", id)
	} else if filterList, ok := s.filtersToApply[imageFile.Id()]; ok {
		s.filtersToApply[imageFile.Id()] = append(filterList, filter)
	}
}

func (s *FilterService) AddFilter(filter *Filter) {
	s.filters[filter.id] = filter
}

func (s *FilterService) GetFilters(imageId apitype.ImageId, options *api.PersistCategorizationCommand) []*Filter {
	filtersToApply := s.getFiltersForImageFile(imageId)

	if options.FixOrientation {
		filtersToApply = append(filtersToApply, &Filter{
			id:        "exifRotate",
			operation: NewImageExifRotate(),
		})
	}
	return filtersToApply
}

func (s *FilterService) getFiltersForImageFile(imageId apitype.ImageId) []*Filter {
	var filtersToApply []*Filter
	if f, ok := s.filtersToApply[imageId]; ok {
		filtersToApply = make([]*Filter, len(f)+1)
		copy(filtersToApply, f)
	} else {
		filtersToApply = []*Filter{}
	}
	return filtersToApply
}
