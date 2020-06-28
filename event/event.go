package event

import (
	"github.com/gotk3/gotk3/glib"
	messagebus "github.com/vardius/message-bus"
	"log"
)

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

func (s *Broker) SubscribeGuiEvent(topic Topic, guidCall GuiCall) {
	cb := func(event Message) {
		glib.IdleAdd(func() {
			guidCall(event)
		})
	}
	err := s.bus.Subscribe(string(topic), cb)
	if err != nil {
		log.Panic("Could not subscribe")
	}
}

func (s *Broker) Send(message Message) {
	s.bus.Publish(string(message.topic), message)
}
func (s *Broker) SendToTopic(topic Topic) {
	s.Send(Message {topic: topic})
}
func (s *Broker) SendToTopicWithData(topic Topic, data Command) {
	s.Send(Message {topic: topic, data: data})
}
func (s *Broker) SendToSubTopicWithData(topic Topic, subTopic Topic, data Command) {
	s.Send(Message {topic: topic, subTopic: subTopic, data: data})
}
