package api

import "vincit.fi/image-sorter/api/apitype"

type Sender interface {
	SendToTopic(topic Topic)
	SendCommandToTopic(topic Topic, command apitype.Command)
	SendError(message string, err error)
}
