package event

type Message struct {
	topic Topic
	subTopic Topic
	data  interface{}
}

func (s *Message) GetData() interface{} {
	return s.data
}

func (s *Message) GetTopic() Topic {
	return s.topic
}

func (s *Message) HasSubTopic() bool {
	return s.subTopic != ""
}

func (s *Message) GetSubTopic() Topic {
	return s.subTopic
}

func (s *Message) ToString() string {
	if s.HasSubTopic() {
		return string(s.topic) + ":" + string(s.subTopic)
	} else {
		return string(s.topic)
	}
}

