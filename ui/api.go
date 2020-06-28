package ui

import (
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
)

type Gui interface {
	SetImages(handles []*common.Handle, topic event.Topic)
	UpdateCategories(categories []*category.Entry)
	SetImageCategory(command *category.CategorizeCommand)
	Run(args []string)

}
