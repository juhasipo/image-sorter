package event

import (
	"fmt"
	messagebus "github.com/vardius/message-bus"
	"sync"
	"time"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common/logger"
)

const debounceWindowLength = time.Millisecond * 300

type debounceEntry struct {
	topic     api.Topic
	command   apitype.Command
	timer     *time.Timer
	cancelled bool
}

type Broker struct {
	bus           messagebus.MessageBus
	debounceMutex sync.Mutex
	debounced     map[string]*debounceEntry

	api.Sender
}

func InitBus(queueSize int) *Broker {
	broker := &Broker{
		bus:       messagebus.New(queueSize),
		debounced: map[string]*debounceEntry{},
	}

	return broker
}

func (s *Broker) sendOrDebounceMessage(topic api.Topic, command apitype.Command) {
	identifier := string(topic)

	s.debounceMutex.Lock()
	defer s.debounceMutex.Unlock()

	if logger.IsLogLevel(logger.TRACE) {
		logger.Trace.Printf("Processing message to topic %s", identifier)
	}

	if oldEntry, ok := s.debounced[identifier]; !ok {
		s.debounced[identifier] = &debounceEntry{}
	} else {
		if logger.IsLogLevel(logger.TRACE) {
			logger.Trace.Printf("Debouncing message to topic %s. Cancelling old entry...", identifier)
		}
		oldEntry.timer.Stop()
		oldEntry.cancelled = true
	}

	entry := &debounceEntry{}
	s.debounced[identifier] = entry

	entry.topic = topic
	entry.command = command
	entry.timer = s.createDelayedSendFunc(entry, debounceWindowLength)
}

func (s *Broker) createDelayedSendFunc(entry *debounceEntry, duration time.Duration) *time.Timer {
	if logger.IsLogLevel(logger.TRACE) {
		logger.Trace.Printf("Creating new debounce entry for topic %s", entry.topic)
	}
	return time.AfterFunc(duration, func() {
		s.debounceMutex.Lock()
		defer s.debounceMutex.Unlock()

		if !entry.cancelled {
			if logger.IsLogLevel(logger.TRACE) {
				logger.Trace.Printf("Actually sending debounced message to topic %s", entry.topic)
			}
			defer delete(s.debounced, string(entry.topic))

			s.bus.Publish(string(entry.topic), entry.command)
		} else {
			if logger.IsLogLevel(logger.TRACE) {
				logger.Debug.Printf("Skip sending cancelled message to topic %s", entry.topic)
			}
		}
	})
}

func InitDevNullBus() *Broker {
	return &Broker{
		bus: nil,
	}
}

func (s *Broker) Subscribe(topic api.Topic, fn interface{}) {
	if s.bus == nil {
		return
	}

	err := s.bus.Subscribe(string(topic), fn)
	if err != nil {
		logger.Error.Panic("Could not subscribe")
	}
}

type GuiCallback func(data ...interface{})

func (s *Broker) SendToTopic(topic api.Topic) {
	if s.bus == nil {
		return
	}

	logger.Trace.Printf("Sending to '%s'", topic)
	s.bus.Publish(string(topic))
}

func (s *Broker) SendCommandToTopic(topic api.Topic, command apitype.Command) {
	if s.bus == nil {
		return
	}

	logger.Trace.Printf("Sending command to '%s'", topic)
	if command.IsThrottled() {
		s.sendOrDebounceMessage(topic, command)
	} else {
		s.bus.Publish(string(topic), command)
	}
}

func (s *Broker) SendError(message string, err error) {
	if s.bus == nil {
		return
	}

	formattedMessage := ""
	if err != nil {
		formattedMessage = fmt.Sprintf("%s\n%s", message, err.Error())
	} else {
		formattedMessage = message
	}
	logger.Error.Printf("Error: %s", formattedMessage)
	s.SendCommandToTopic(api.ShowError, &api.ErrorCommand{Message: message})
}
