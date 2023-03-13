package events

import "sync"

type DomainEvent interface {
	EventName() string
}

type EventStore interface {
	AddEvent(e ...DomainEvent)
	Events() []DomainEvent
}

type eventStore struct {
	events []DomainEvent
	sync.RWMutex
}

func NewEventStore() EventStore {
	return &eventStore{
		events: make([]DomainEvent, 0),
	}
}

func (s *eventStore) AddEvent(e ...DomainEvent) {
	s.Lock()
	s.events = append(s.events, e...)
	s.Unlock()
}

func (s *eventStore) Events() []DomainEvent {
	return s.events
}
