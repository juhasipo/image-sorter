package ui

import (
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
)

type Gui interface {
	SetImages(topic event.Topic, handles []*common.Handle)
	UpdateCategories(categories *category.CategoriesCommand)
	SetImageCategory(command *category.CategorizeCommand)
	Run(args []string)

}