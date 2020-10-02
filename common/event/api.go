package event

type Command interface{}

type Sender interface {
	SendToTopic(topic Topic)
	SendToTopicWithData(topic Topic, data ...interface{})
}
