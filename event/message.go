package event

type Message struct {
	topic Topic
	subTopic Topic
	data  interface{}
}

func (s *Message) GetData() interface{} {
	return s.data
}

func (s *Message) GetSubTopic() Topic {
	return s.subTopic
}

