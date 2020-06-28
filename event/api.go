package event

type Command interface {}

type GuiCall func(message Message)
type Sender interface {
	Send(message Message)
	SendToTopic(topic Topic)
	SendToTopicWithData(topic Topic, data Command)
	SendToSubTopicWithData(topic Topic, subTopic Topic, data Command)
}
