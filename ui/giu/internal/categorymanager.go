package internal

import (
	"github.com/AllenDang/giu"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/ui/giu/internal/guiapi"
)

type CategoryDef struct {
	Name       string
	Key        giu.Key
	CategoryId apitype.CategoryId
}

type CategoryKeyManager struct {
	Categories     map[string]*CategoryDef
	CategoryKeyMap map[giu.Key]*CategoryDef
	CategoryIdMap  map[apitype.CategoryId]*CategoryDef
	Callback       func(def *CategoryDef, action *guiapi.CategoryAction)
}

func (s *CategoryKeyManager) Reset(categories []*apitype.Category) {
	s.Categories = map[string]*CategoryDef{}
	s.CategoryKeyMap = map[giu.Key]*CategoryDef{}
	s.CategoryIdMap = map[apitype.CategoryId]*CategoryDef{}

	for _, category := range categories {
		key := giu.Key(category.Shortcut())
		def := CategoryDef{
			CategoryId: category.Id(),
			Name:       category.Name(),
			Key:        key,
		}
		s.CategoryKeyMap[key] = &def
		s.CategoryIdMap[category.Id()] = &def
	}
}

func (s *CategoryKeyManager) HandleCategory(id apitype.CategoryId, action *guiapi.CategoryAction) {
	s.Callback(s.CategoryIdMap[id], action)
}

func (s *CategoryKeyManager) HandleKeys(action *guiapi.CategoryAction) {
	for key, def := range s.CategoryKeyMap {
		if giu.IsKeyPressed(key) {
			s.Callback(def, action)
		}
	}
}
