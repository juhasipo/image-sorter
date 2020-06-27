package event

import (
	messagebus "github.com/vardius/message-bus"
	"log"
)

type Topic string
const(
	NEXT_IMAGE = "event-next-image"
	PREV_IMAGE = "event-prev-image"
)

type Message struct {
	topic Topic
	data  interface{}
}

func New(topic Topic) Message {
	return Message {topic: topic}
}
func NewWithData(topic Topic, data interface{}) Message {
	return Message {topic: topic, data: data}
}

type Sender interface {
	Send(message Message)
}

type Broker struct {
	bus messagebus.MessageBus

	Sender
}

func InitBus(queueSize int) *Broker{
	return &Broker {
		bus: messagebus.New(queueSize),
	}
}

func (s* Broker) Subscribe(topic Topic, fn interface{}) {
	err := s.bus.Subscribe(string(topic), fn)
	if err != nil {
		log.Panic("Could not subscribe")
	}
}

func (s *Broker) Send(message Message) {
	s.bus.Publish(string(message.topic), message)
}
