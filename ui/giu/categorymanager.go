package gtk

import (
	"github.com/AllenDang/giu"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/ui/giu/guiapi"
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
	callback       func(def *CategoryDef, action *guiapi.CategoryAction)
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

func (s *CategoryKeyManager) HandleCategory(id apitype.CategoryId, action *guiapi.CategoryAction) {
	s.callback(s.categoryIdMap[id], action)
}

func (s *CategoryKeyManager) HandleKeys(action *guiapi.CategoryAction) {
	for key, def := range s.categoryKeyMap {
		if giu.IsKeyPressed(key) {
			s.callback(def, action)
		}
	}
}
