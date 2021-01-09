package api

type Sender interface {
	SendToTopic(topic Topic)
	SendToTopicWithData(topic Topic, data ...interface{})
	SendError(message string, err error)
}
