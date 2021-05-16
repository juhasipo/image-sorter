package caster

import "vincit.fi/image-sorter/api/apitype"

type SettingCategory struct {
	Id   apitype.CategoryId `json:"id"`
	Name string             `json:"name"`
}

type Settings struct {
	Categories []*SettingCategory `json:"categories"`
}

type CurrentImage struct {
	Id                apitype.ImageId      `json:"imageId"`
	CurrentImageIndex int                  `json:"currentImageIndex"`
	TotalImages       int                  `json:"totalImages"`
	Categories        []apitype.CategoryId `json:"categoryIds"`
}

type WebsocketMessage struct {
	Type    string      `json:"type"`
	Message interface{} `json:"data"`
}
