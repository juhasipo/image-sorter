package event

import (
	"fmt"
	"github.com/gotk3/gotk3/glib"
	messagebus "github.com/vardius/message-bus"
	"reflect"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
)

type Broker struct {
	bus messagebus.MessageBus

	api.Sender
}

func InitBus(queueSize int) *Broker {
	return &Broker{
		bus: messagebus.New(queueSize),
	}
}

func (s *Broker) Subscribe(topic api.Topic, fn interface{}) {
	err := s.bus.Subscribe(string(topic), fn)
	if err != nil {
		logger.Error.Panic("Could not subscribe")
	}
}

type GuiCallback func(data ...interface{})

func (s *Broker) ConnectToGui(topic api.Topic, callback interface{}) {
	cb := func(params ...interface{}) {
		sendFn := func() {
			args := make([]reflect.Value, 0, len(params))
			for _, param := range params {
				args = append(args, reflect.ValueOf(param))
			}
			logger.Trace.Printf("Calling topic '%s' with: %s", topic, params)
			reflect.ValueOf(callback).Call(args)
		}

		if _, err := glib.IdleAdd(sendFn); err != nil {
			s.SendError("Error sending internal message", err)
		}
	}
	err := s.bus.Subscribe(string(topic), cb)
	if err != nil {
		logger.Error.Panic("Could not subscribe")
	}
}

func (s *Broker) SendToTopic(topic api.Topic) {
	logger.Trace.Printf("Sending to '%s'", topic)
	s.bus.Publish(string(topic))
}

func (s *Broker) SendCommandToTopic(topic api.Topic, command apitype.Command) {
	logger.Trace.Printf("Sending command to '%s'", topic)
	s.bus.Publish(string(topic), command)
}

func (s *Broker) SendError(message string, err error) {
	formattedMessage := ""
	if err != nil {
		formattedMessage = fmt.Sprintf("%s\n%s", message, err.Error())
	} else {
		formattedMessage = message
	}
	logger.Error.Printf("Error: %s", formattedMessage)
	s.SendCommandToTopic(api.ShowError, &api.ErrorCommand{Message: message})
}
