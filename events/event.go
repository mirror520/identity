package events

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/mirror520/identity/pubsub"
)

var instance pubsub.PubSub

func ReplaceGlobals(pb pubsub.PubSub) {
	instance = pb
}

type DomainEvent interface {
	EventName() string
	Topic() string
}

type EventStore interface {
	AddEvent(e ...DomainEvent)
	Notify() error
	Events() []DomainEvent // debug only
}

type eventStore struct {
	pubsub pubsub.PubSub
	events []DomainEvent
	sync.Mutex
}

func NewEventStore() EventStore {
	return &eventStore{
		pubsub: instance,
		events: make([]DomainEvent, 0),
	}
}

func (s *eventStore) AddEvent(e ...DomainEvent) {
	s.Lock()
	s.events = append(s.events, e...)
	s.Unlock()
}

func (s *eventStore) Notify() error {
	if s.pubsub == nil {
		return errors.New("pubsub not found")
	}

	s.Lock()
	defer s.Unlock()

	for _, e := range s.events {
		data, err := json.Marshal(&e)
		if err != nil {
			return err
		}

		if err := s.pubsub.Publish(e.Topic(), data); err != nil {
			return err
		}
	}

	s.events = make([]DomainEvent, 0)
	return nil
}

func (s *eventStore) Events() []DomainEvent {
	return s.events
}
