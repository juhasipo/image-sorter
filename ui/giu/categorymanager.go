package gtk

import (
	"github.com/AllenDang/giu"
	"vincit.fi/image-sorter/api/apitype"
)

type CategoryDef struct {
	name       string
	key        giu.Key
	categoryId apitype.CategoryId
}

type CategoryKeyManager struct {
	categories     map[string]*CategoryDef
	categoryKeyMap map[giu.Key]*CategoryDef
	categoryIdMap  map[apitype.CategoryId]*CategoryDef
	callback       func(def *CategoryDef, stayOnImage bool, forceCategory bool)
}

func (s *CategoryKeyManager) Reset(categories []*apitype.Category) {
	s.categories = map[string]*CategoryDef{}
	s.categoryKeyMap = map[giu.Key]*CategoryDef{}
	s.categoryIdMap = map[apitype.CategoryId]*CategoryDef{}

	for _, category := range categories {
		key := giu.Key(category.Shortcut())
		def := CategoryDef{
			categoryId: category.Id(),
			name:       category.Name(),
			key:        key,
		}
		s.categoryKeyMap[key] = &def
		s.categoryIdMap[category.Id()] = &def
	}
}

func (s *CategoryKeyManager) HandleCategory(id apitype.CategoryId, stayOnImage bool, forceCategory bool) {
	s.callback(s.categoryIdMap[id], stayOnImage, forceCategory)
}

func (s *CategoryKeyManager) HandleKeys(stayOnImage bool, forceCategory bool) {
	for key, def := range s.categoryKeyMap {
		if giu.IsKeyPressed(key) {
			s.callback(def, stayOnImage, forceCategory)
		}
	}
}
