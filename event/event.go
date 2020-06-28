package event

import (
	"github.com/gotk3/gotk3/glib"
	messagebus "github.com/vardius/message-bus"
	"log"
)

type Topic string
const(
	NEXT_IMAGE = "event-next-image"
	PREV_IMAGE = "event-prev-image"
	CURRENT_IMAGE = "event-current-image"

	IMAGES_UPDATED = "event-images-updated"
)

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


func New(topic Topic) Message {
	return Message {topic: topic}
}
func NewWithData(topic Topic, data interface{}) Message {
	return Message {topic: topic, data: data}
}
func NewWithSubAndData(topic Topic, subTopic Topic, data interface{}) Message {
	return Message {topic: topic, subTopic: subTopic, data: data}
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

type GuiCall func(message Message)
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
