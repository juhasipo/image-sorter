package event

import (
	"github.com/gotk3/gotk3/glib"
	messagebus "github.com/vardius/message-bus"
	"log"
	"reflect"
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

type GuiCallback func (data ...interface{})

func (s *Broker) ConnectToGui(topic Topic, callback interface{}) {
	cb := func(params ...interface{}) {
		glib.IdleAdd(func() {
			args := make([]reflect.Value, 0, len(params))
			for _, param := range params {
				args = append(args, reflect.ValueOf(param))
			}
			log.Printf("Calling topic '%s' with %d arguments", topic, len(args))
			reflect.ValueOf(callback).Call(args)
		})
	}
	err := s.bus.Subscribe(string(topic), cb)
	if err != nil {
		log.Panic("Could not subscribe")
	}
}

func (s *Broker) SendToTopic(topic Topic) {
	log.Printf("Sending to '%s'", topic)
	s.bus.Publish(string(topic))
}
func (s *Broker) SendToTopicWithData(topic Topic, data ...interface{}) {
	log.Printf("Sending to '%s' with %d arguments", topic, len(data))
	s.bus.Publish(string(topic), data...)
}
